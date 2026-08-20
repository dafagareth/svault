// Copyright 2026 Dafa
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package cmd

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"svault/internal/storage"
)

var openNamespace string

var openCmd = &cobra.Command{
	Use:   "open KEY",
	Short: "Open a URL from vault in browser and copy password to clipboard",
	Long: `Opens the URL stored in KEY_URL (or KEY if it starts with http) in the
default browser, then copies KEY_PASS or KEY_PASSWORD to clipboard.

Examples:
  svault set GITHUB_URL=https://github.com/login
  svault set GITHUB_PASS=mysecret
  svault open GITHUB`,
	Args: cobra.ExactArgs(1),
	RunE: runOpen,
}

func init() {
	openCmd.Flags().StringVar(&openNamespace, "ns", "default", "namespace")
	rootCmd.AddCommand(openCmd)
}

func runOpen(cmd *cobra.Command, args []string) error {
	key := strings.ToUpper(args[0])
	ns := activeNamespace(cmd, openNamespace)

	vpath, err := vaultPath()
	if err != nil {
		return err
	}
	encKey, err := storage.LoadSession()
	if err != nil {
		return err
	}
	vd, err := storage.ReadVault(vpath, encKey)
	if err != nil {
		return err
	}

	nsMap := vd.Namespaces[ns]

	// Resolve URL: KEY_URL → KEY (if starts with http) → error
	url := ""
	for _, candidate := range []string{key + "_URL", key} {
		if v, ok := nsMap[candidate]; ok && (strings.HasPrefix(v, "http://") || strings.HasPrefix(v, "https://")) {
			url = v
			break
		}
	}
	if url == "" {
		return fmt.Errorf("no URL found for %s (set %s_URL or %s with an http URL)", key, key, key)
	}

	// Open browser
	if err := exec.Command("xdg-open", url).Start(); err != nil {
		return fmt.Errorf("open browser: %w", err)
	}
	fmt.Printf("Opening %s\n", url)

	// Resolve password: KEY_PASS → KEY_PASSWORD → not found
	passKey := ""
	for _, candidate := range []string{key + "_PASS", key + "_PASSWORD", key + "_TOKEN"} {
		if _, ok := nsMap[candidate]; ok {
			passKey = candidate
			break
		}
	}

	if passKey != "" {
		if err := writeClipboard(nsMap[passKey]); err == nil {
			scheduleClearClipboard(30)
			fmt.Printf("Copied %s to clipboard. Will clear in 30 seconds.\n", passKey)
		}
	} else {
		fmt.Printf("Tip: set %s_PASS to auto-copy password on open.\n", key)
	}

	return nil
}
