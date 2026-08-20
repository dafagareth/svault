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

	"github.com/spf13/cobra"
	"svault/internal/audit"
	"svault/internal/envfile"
	"svault/internal/storage"
)

var importNamespace string
var importOverwrite bool

var importCmd = &cobra.Command{
	Use:   "import FILE",
	Short: "Import secrets from a .env file",
	Args:  cobra.ExactArgs(1),
	RunE:  runImport,
}

func init() {
	importCmd.Flags().StringVar(&importNamespace, "ns", "default", "target namespace")
	importCmd.Flags().BoolVar(&importOverwrite, "overwrite", true, "overwrite existing keys (use --overwrite=false to skip)")
	rootCmd.AddCommand(importCmd)
}

func runImport(cmd *cobra.Command, args []string) error {
	ns := activeNamespace(cmd, importNamespace)

	parsed, err := envfile.Parse(args[0])
	if err != nil {
		return err
	}

	entries := make(map[string]string, len(parsed))
	for _, e := range parsed {
		if err := validateKey(e.Key); err != nil {
			return fmt.Errorf("invalid key in %s: %w", args[0], err)
		}
		entries[e.Key] = e.Value
	}
	if len(entries) == 0 {
		return fmt.Errorf("no valid keys found in %s", args[0])
	}

	skipped := 0
	err = mutateVault(func(vd *storage.VaultData) error {
		if vd.Namespaces[ns] == nil {
			vd.Namespaces[ns] = make(map[string]string)
		}
		for k, v := range entries {
			if !importOverwrite {
				if _, exists := vd.Namespaces[ns][k]; exists {
					skipped++
					continue
				}
			}
			vd.Namespaces[ns][k] = v
		}
		return nil
	})
	if err != nil {
		return err
	}

	lpath, _ := logPath()
	_ = audit.Append(lpath, audit.OpImport, "-", ns)

	imported := len(entries) - skipped
	fmt.Printf("Imported %d keys into [%s]", imported, ns)
	if skipped > 0 {
		fmt.Printf(", %d skipped (already exist)", skipped)
	}
	fmt.Println()
	return nil
}
