package ns

import (
	"fmt"
	"sort"
	"svault/cmd/common"

	"github.com/spf13/cobra"
	"svault/internal/storage"
)

var NsCmd = &cobra.Command{
	GroupID: "ns",
	Use:     "ns",
	Short:   "Manage namespaces",
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
	NsCmd.AddCommand(nsListCmd, nsDeleteCmd, nsRenameCmd)
}

func loadVault() (string, []byte, *storage.VaultData, error) {
	vpath, err := common.VaultPath()
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
	_, _, vd, err := loadVault()
	if err != nil {
		return err
	}
	names := make([]string, 0, len(vd.Namespaces))
	for n := range vd.Namespaces {
		names = append(names, n)
	}
	sort.Strings(names)

	active, _ := common.NamespaceSource()
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
	err := common.MutateVault(func(vd *storage.VaultData) error {
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

	err := common.MutateVault(func(vd *storage.VaultData) error {
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
