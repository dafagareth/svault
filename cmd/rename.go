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
	"svault/internal/storage"
)

var renameNamespace string

var renameCmd = &cobra.Command{
	Use:   "rename OLD NEW",
	Short: "Rename a key, keeping its value",
	Args:  cobra.ExactArgs(2),
	RunE:  runRename,
}

func init() {
	renameCmd.Flags().StringVar(&renameNamespace, "ns", "default", "namespace")
	rootCmd.AddCommand(renameCmd)
}

func runRename(cmd *cobra.Command, args []string) error {
	oldKey, newKey := args[0], args[1]
	if err := validateKey(newKey); err != nil {
		return err
	}
	ns := activeNamespace(cmd, renameNamespace)

	err := mutateVault(func(vd *storage.VaultData) error {
		nsMap := vd.Namespaces[ns]
		val, ok := nsMap[oldKey]
		if !ok {
			return fmt.Errorf("key not found: %s", oldKey)
		}
		if _, exists := nsMap[newKey]; exists {
			return fmt.Errorf("key already exists: %s", newKey)
		}
		nsMap[newKey] = val
		delete(nsMap, oldKey)
		return nil
	})
	if err != nil {
		return err
	}

	lpath, _ := logPath()
	_ = audit.Append(lpath, audit.OpDelete, oldKey, ns)
	_ = audit.Append(lpath, audit.OpSet, newKey, ns)

	fmt.Printf("Renamed %s to %s in [%s]\n", oldKey, newKey, ns)
	return nil
}
