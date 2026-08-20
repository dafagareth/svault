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
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

var backupCmd = &cobra.Command{
	Use:   "backup [FILE]",
	Short: "Back up the vault to a file",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runBackup,
}

func init() {
	rootCmd.AddCommand(backupCmd)
}

func runBackup(_ *cobra.Command, args []string) error {
	vpath, err := vaultPath()
	if err != nil {
		return err
	}

	dest := ""
	if len(args) == 1 {
		dest = args[0]
	} else {
		dir, err := vaultDir()
		if err != nil {
			return err
		}
		ts := time.Now().Format("20060102-150405")
		dest = filepath.Join(dir, "vault.enc.backup-"+ts)
	}

	src, err := os.ReadFile(vpath)
	if err != nil {
		return fmt.Errorf("read vault: %w", err)
	}
	if err := os.WriteFile(dest, src, 0600); err != nil {
		return fmt.Errorf("write backup: %w", err)
	}

	fmt.Printf("Vault backed up to %s\n", dest)
	return nil
}
