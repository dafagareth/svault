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
	"svault/internal/audit"
	"svault/internal/storage"
)

var getNamespace string
var getClip bool
var getTimeout int

var getCmd = &cobra.Command{
	Use:   "get KEY [KEY...]",
	Short: "Get one or more secret values",
	Long: `Print one or more secret values from the active namespace.

A single key prints just the value. Several keys print KEY=VALUE per line.
With --clip the value is copied to the clipboard instead of printed, and the
clipboard is cleared after --timeout seconds.

Examples:
  svault get DB_PASSWORD              # print the value
  svault get DB_HOST DB_PORT          # print KEY=VALUE for each
  svault get DB_PASSWORD --clip       # copy to clipboard, clears in 20s
  svault get DB_PASSWORD -c --timeout 0   # copy, do not auto-clear`,
	Args: cobra.MinimumNArgs(1),
	RunE: runGet,
}

func init() {
	getCmd.Flags().StringVar(&getNamespace, "ns", "default", "source namespace")
	getCmd.Flags().BoolVarP(&getClip, "clip", "c", false, "copy value to clipboard instead of printing")
	getCmd.Flags().IntVar(&getTimeout, "timeout", 20, "seconds before clearing the clipboard (0 = never)")
	rootCmd.AddCommand(getCmd)
}

func runGet(cmd *cobra.Command, args []string) error {
	ns := activeNamespace(cmd, getNamespace)

	if getClip && len(args) > 1 {
		return fmt.Errorf("--clip works with a single key only")
	}

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

	lpath, _ := logPath()
	multi := len(args) > 1

	for _, key := range args {
		value, ok := vd.Namespaces[ns][key]
		if !ok {
			return fmt.Errorf("key not found: %s", key)
		}
		_ = audit.Append(lpath, audit.OpGet, key, ns)

		if getClip {
			if err := copyToClipboard(value); err != nil {
				return err
			}
			fmt.Printf("Copied %s to clipboard", key)
			if getTimeout > 0 {
				scheduleClipboardClear(getTimeout)
				fmt.Printf(" (clears in %ds)", getTimeout)
			}
			fmt.Println()
			return nil
		}

		if multi {
			fmt.Printf("%s=%s\n", key, value)
		} else {
			fmt.Println(value)
		}
	}
	return nil
}

// scheduleClipboardClear spawns a detached process that wipes the clipboard
// after the given number of seconds. The secret is never passed as an argument.
func scheduleClipboardClear(seconds int) {
	backend, ok := availableClipboardBackend()
	if !ok {
		return
	}
	script := fmt.Sprintf("sleep %d; printf '' | %s", seconds, strings.Join(backend.copyCmd, " "))
	_ = exec.Command("sh", "-c", script).Start()
}
