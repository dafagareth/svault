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
	"os"
	"strings"

	"github.com/spf13/cobra"
	"svault/internal/storage"
)

// version di-inject saat build via: -ldflags "-X 'svault/cmd.version=x.y.z'"
var version = "dev"

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show vault information",
	RunE:  runInfo,
}

func init() {
	rootCmd.AddCommand(infoCmd)
}

func runInfo(_ *cobra.Command, _ []string) error {
	vpath, err := vaultPath()
	if err != nil {
		return err
	}
	dir, err := vaultDir()
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

	cfg, err := storage.ReadConfig(dir)
	if err != nil {
		return err
	}

	totalSecrets := 0
	for _, ns := range vd.Namespaces {
		totalSecrets += len(ns)
	}

	displayPath := vpath
	if home, err := os.UserHomeDir(); err == nil && strings.HasPrefix(vpath, home) {
		displayPath = "~" + vpath[len(home):]
	}

	ns, src := namespaceSource()
	_ = cfg

	remaining, sessionActive := storage.SessionRemaining()
	var sessionStr string
	if sessionActive {
		sessionStr = fmt.Sprintf("unlocked, %dm remaining", int(remaining.Minutes()))
	} else {
		sessionStr = "locked"
	}

	fmt.Printf("svault v%s\n", version)
	fmt.Printf("Vault file : %s\n", displayPath)
	fmt.Printf("Namespace  : %s (from %s, %d total)\n", ns, src, len(vd.Namespaces))
	fmt.Printf("Secrets    : %d total\n", totalSecrets)
	fmt.Printf("Encryption : AES-256-GCM\n")
	fmt.Printf("Session    : %s\n", sessionStr)
	return nil
}
