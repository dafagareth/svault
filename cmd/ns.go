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
	"sort"

	"github.com/spf13/cobra"
	"svault/internal/storage"
)

var nsCmd = &cobra.Command{
	Use:   "ns",
	Short: "Manage namespaces",
}

var nsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all namespaces with their key counts",
	RunE:  runNsList,
}

var nsDeleteCmd = &cobra.Command{
	Use:   "delete NAMESPACE",
	Short: "Delete a namespace and all its secrets",
	Args:  cobra.ExactArgs(1),
	RunE:  runNsDelete,
}

var nsRenameCmd = &cobra.Command{
	Use:   "rename OLD NEW",
	Short: "Rename a namespace",
	Args:  cobra.ExactArgs(2),
	RunE:  runNsRename,
}

func init() {
	nsCmd.AddCommand(nsListCmd, nsDeleteCmd, nsRenameCmd)
	rootCmd.AddCommand(nsCmd)
}

func loadVaultRW() (string, []byte, *storage.VaultData, error) {
	vpath, err := vaultPath()
	if err != nil {
		return "", nil, nil, err
	}
	encKey, err := storage.LoadSession()
	if err != nil {
		return "", nil, nil, err
	}
	vd, err := storage.ReadVault(vpath, encKey)
	if err != nil {
		return "", nil, nil, err
	}
	return vpath, encKey, vd, nil
}

func runNsList(_ *cobra.Command, _ []string) error {
	_, _, vd, err := loadVaultRW()
	if err != nil {
		return err
	}
	names := make([]string, 0, len(vd.Namespaces))
	for n := range vd.Namespaces {
		names = append(names, n)
	}
	sort.Strings(names)

	active, _ := namespaceSource()
	for _, n := range names {
		marker := "  "
		if n == active {
			marker = "* "
		}
		fmt.Printf("%s%-20s %d keys\n", marker, n, len(vd.Namespaces[n]))
	}
	return nil
}

func runNsDelete(_ *cobra.Command, args []string) error {
	name := args[0]
	if name == "default" {
		return fmt.Errorf("cannot delete the default namespace")
	}

	count := 0
	err := mutateVault(func(vd *storage.VaultData) error {
		if _, ok := vd.Namespaces[name]; !ok {
			return fmt.Errorf("namespace not found: %s", name)
		}
		count = len(vd.Namespaces[name])
		delete(vd.Namespaces, name)
		return nil
	})
	if err != nil {
		return err
	}
	fmt.Printf("Deleted namespace [%s] (%d keys removed)\n", name, count)
	return nil
}

func runNsRename(_ *cobra.Command, args []string) error {
	oldName, newName := args[0], args[1]
	if oldName == "default" {
		return fmt.Errorf("cannot rename the default namespace")
	}

	err := mutateVault(func(vd *storage.VaultData) error {
		if _, ok := vd.Namespaces[oldName]; !ok {
			return fmt.Errorf("namespace not found: %s", oldName)
		}
		if _, exists := vd.Namespaces[newName]; exists {
			return fmt.Errorf("namespace already exists: %s", newName)
		}
		vd.Namespaces[newName] = vd.Namespaces[oldName]
		delete(vd.Namespaces, oldName)
		return nil
	})
	if err != nil {
		return err
	}
	fmt.Printf("Renamed namespace [%s] to [%s]\n", oldName, newName)
	return nil
}
