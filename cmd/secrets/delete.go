package secrets

import (
	"fmt"
	"svault/cmd/common"

	"github.com/spf13/cobra"
	"svault/internal/audit"
	"svault/internal/storage"
)

var deleteNamespace string

var DeleteCmd = &cobra.Command{
	GroupID: "secrets",
	Use:     "delete KEY",
	Short:   "Delete a secret from the vault",
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return common.CompleteActiveKeys(cmd)
	},
	Args: cobra.ExactArgs(1),
	RunE: runDelete,
}

func init() {
	DeleteCmd.Flags().StringVar(&deleteNamespace, "ns", "default", "target namespace")
}

func runDelete(cmd *cobra.Command, args []string) error {
	key := args[0]
	ns := common.ActiveNamespace(cmd, deleteNamespace)

	err := common.MutateVault(func(vd *storage.VaultData) error {
		if _, ok := vd.Namespaces[ns][key]; !ok {
			return fmt.Errorf("key not found: %s", key)
		}
		delete(vd.Namespaces[ns], key)
		return nil
	})
	if err != nil {
		return err
	}

	lpath, _ := common.LogPath()
	_ = audit.Append(lpath, audit.OpDelete, key, ns)

	fmt.Printf("Deleted %s\n", key)
	return nil
}
