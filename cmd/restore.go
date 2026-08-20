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

	"github.com/spf13/cobra"
	"svault/internal/storage"
)

var restoreCmd = &cobra.Command{
	Use:   "restore FILE",
	Short: "Restore the vault from a backup file",
	Args:  cobra.ExactArgs(1),
	RunE:  runRestore,
}

func init() {
	rootCmd.AddCommand(restoreCmd)
}

func runRestore(_ *cobra.Command, args []string) error {
	backupPath := args[0]

	vpath, err := vaultPath()
	if err != nil {
		return err
	}
	dir, err := vaultDir()
	if err != nil {
		return err
	}

	// Read the backup before locking so a bad path fails fast.
	src, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("read backup: %w", err)
	}

	err = storage.WithVaultLock(dir, func() error {
		if current, err := os.ReadFile(vpath); err == nil {
			_ = os.WriteFile(vpath+".pre-restore", current, 0600)
		}
		if err := os.WriteFile(vpath, src, 0600); err != nil {
			return fmt.Errorf("write vault: %w", err)
		}
		return nil
	})
	if err != nil {
		return err
	}

	_ = storage.DeleteSession()

	fmt.Printf("Vault restored from %s\n", backupPath)
	fmt.Println("Session cleared. Run 'svault unlock' to start a new session.")
	return nil
}
