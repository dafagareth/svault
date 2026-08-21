package secrets

import (
	"fmt"
	"svault/cmd/common"

	"github.com/spf13/cobra"
	"svault/internal/audit"
	"svault/internal/storage"
)

var renameNamespace string

var RenameCmd = &cobra.Command{
	GroupID: "secrets",
	Use:     "rename OLD NEW",
	Short:   "Rename a key, keeping its value",
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return common.CompleteActiveKeys(cmd)
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	Args: cobra.ExactArgs(2),
	RunE: runRename,
}

func init() {
	RenameCmd.Flags().StringVar(&renameNamespace, "ns", "default", "namespace")
}

func runRename(cmd *cobra.Command, args []string) error {
	oldKey, newKey := args[0], args[1]
	if err := common.ValidateKey(newKey); err != nil {
		return err
	}
	ns := common.ActiveNamespace(cmd, renameNamespace)

	err := common.MutateVault(func(vd *storage.VaultData) error {
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

	lpath, _ := common.LogPath()
	_ = audit.Append(lpath, audit.OpDelete, oldKey, ns)
	_ = audit.Append(lpath, audit.OpSet, newKey, ns)

	fmt.Printf("Renamed %s to %s in [%s]\n", oldKey, newKey, ns)
	return nil
}
