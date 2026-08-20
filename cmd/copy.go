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
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"svault/internal/storage"
)

var copyNamespace string

var copyCmd = &cobra.Command{
	Use:   "copy KEY",
	Short: "Copy a secret value to clipboard (auto-clears after 30s)",
	Args:  cobra.ExactArgs(1),
	RunE:  runCopy,
}

func init() {
	copyCmd.Flags().StringVar(&copyNamespace, "ns", "default", "namespace")
	rootCmd.AddCommand(copyCmd)
}

func runCopy(cmd *cobra.Command, args []string) error {
	key := args[0]
	ns := activeNamespace(cmd, copyNamespace)

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

	val, ok := vd.Namespaces[ns][key]
	if !ok {
		return fmt.Errorf("key not found: %s", key)
	}

	if err := writeClipboard(val); err != nil {
		return fmt.Errorf("clipboard: %w", err)
	}

	scheduleClearClipboard(30)
	fmt.Printf("Copied %s to clipboard. Will clear in 30 seconds.\n", key)
	return nil
}

// writeClipboard writes text to the system clipboard.
func writeClipboard(text string) error {
	var cmd *exec.Cmd
	if os.Getenv("WAYLAND_DISPLAY") != "" {
		cmd = exec.Command("wl-copy")
	} else if os.Getenv("DISPLAY") != "" {
		cmd = exec.Command("xsel", "--clipboard", "--input")
	} else {
		return fmt.Errorf("no display found (WAYLAND_DISPLAY or DISPLAY must be set)")
	}
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

// scheduleClearClipboard spawns a detached background process that clears
// the clipboard after n seconds. The main process exits immediately.
func scheduleClearClipboard(seconds int) {
	var clearCmd string
	if os.Getenv("WAYLAND_DISPLAY") != "" {
		clearCmd = fmt.Sprintf("sleep %d && wl-copy --clear", seconds)
	} else {
		clearCmd = fmt.Sprintf("sleep %d && xsel --clipboard --input < /dev/null", seconds)
	}
	c := exec.Command("sh", "-c", clearCmd)
	c.Stdin = nil
	c.Stdout = nil
	c.Stderr = nil
	_ = c.Start()
}
