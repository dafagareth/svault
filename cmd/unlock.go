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
	"golang.org/x/term"
	"svault/internal/audit"
	"svault/internal/crypto"
	"svault/internal/storage"
)

var unlockCmd = &cobra.Command{
	Use:   "unlock",
	Short: "Unlock the vault and start a session",
	RunE:  runUnlock,
}

func init() {
	rootCmd.AddCommand(unlockCmd)
}

func runUnlock(_ *cobra.Command, _ []string) error {
	vpath, err := vaultPath()
	if err != nil {
		return err
	}

	salt, err := storage.ReadSalt(vpath)
	if err != nil {
		return fmt.Errorf("vault not initialized, run 'svault init'")
	}

	fmt.Print("Master password: ")
	pw, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		return fmt.Errorf("read password: %w", err)
	}
	defer wipe(pw)

	key := crypto.DeriveKey(pw, salt)

	if _, err := storage.ReadVault(vpath, key); err != nil {
		return fmt.Errorf("wrong password")
	}

	if err := storage.SaveSession(key); err != nil {
		return err
	}

	lpath, _ := logPath()
	_ = audit.Append(lpath, audit.OpUnlock, "-", "default")

	fmt.Println("Vault unlocked. Session valid for 30 minutes.")
	return nil
}
