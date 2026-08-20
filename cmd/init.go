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
	"svault/internal/storage"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new vault with a master password",
	RunE:  runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(_ *cobra.Command, _ []string) error {
	vpath, err := vaultPath()
	if err != nil {
		return err
	}
	if _, err := os.Stat(vpath); err == nil {
		return fmt.Errorf("vault already exists at %s", vpath)
	}

	fmt.Print("Master password: ")
	pw1, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		return fmt.Errorf("read password: %w", err)
	}
	defer wipe(pw1)

	if len(pw1) == 0 {
		return fmt.Errorf("password cannot be empty")
	}

	fmt.Print("Confirm password: ")
	pw2, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		return fmt.Errorf("read confirmation: %w", err)
	}
	defer wipe(pw2)

	if string(pw1) != string(pw2) {
		return fmt.Errorf("passwords do not match")
	}

	if err := storage.InitVault(vpath, pw1); err != nil {
		return err
	}

	lpath, _ := logPath()
	_ = audit.Append(lpath, audit.OpInit, "-", "default")

	fmt.Printf("Vault initialized at %s\n", vpath)
	return nil
}

func wipe(b []byte) {
	for i := range b {
		b[i] = 0
	}
}
