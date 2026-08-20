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
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/spf13/cobra"
	"svault/internal/audit"
	"svault/internal/storage"
)

var (
	genLength    int
	genNoSymbols bool
	genSaveKey   string
	genNamespace string
	genNoCopy    bool
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a secure random password",
	RunE:  runGenerate,
}

func init() {
	generateCmd.Flags().IntVarP(&genLength, "length", "l", 24, "password length")
	generateCmd.Flags().BoolVar(&genNoSymbols, "no-symbols", false, "alphanumeric only")
	generateCmd.Flags().StringVar(&genSaveKey, "save", "", "save as KEY in vault")
	generateCmd.Flags().StringVar(&genNamespace, "ns", "default", "namespace for --save")
	generateCmd.Flags().BoolVar(&genNoCopy, "no-copy", false, "do not copy to clipboard")
	rootCmd.AddCommand(generateCmd)
}

func runGenerate(cmd *cobra.Command, _ []string) error {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const symbols = "!@#$%^&*()-_=+[]{}|;:,.<>?"
	charset := letters
	if !genNoSymbols {
		charset += symbols
	}

	if genLength < 8 {
		return fmt.Errorf("length must be at least 8")
	}

	buf := make([]byte, genLength)
	for i := range buf {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return fmt.Errorf("random: %w", err)
		}
		buf[i] = charset[n.Int64()]
	}
	password := string(buf)

	fmt.Println(password)

	if genSaveKey != "" {
		if err := validateKey(genSaveKey); err != nil {
			return err
		}
		ns := activeNamespace(cmd, genNamespace)
		err := mutateVault(func(vd *storage.VaultData) error {
			if vd.Namespaces[ns] == nil {
				vd.Namespaces[ns] = make(map[string]string)
			}
			vd.Namespaces[ns][genSaveKey] = password
			return nil
		})
		if err != nil {
			return err
		}
		lpath, _ := logPath()
		_ = audit.Append(lpath, audit.OpSet, genSaveKey, ns)
		fmt.Printf("Saved as %s in [%s]\n", genSaveKey, ns)
	}

	if !genNoCopy {
		if err := writeClipboard(password); err == nil {
			scheduleClearClipboard(30)
			fmt.Println("Copied to clipboard. Will clear in 30 seconds.")
		}
	}

	return nil
}
