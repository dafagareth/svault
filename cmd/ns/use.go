package ns

import (
	"fmt"
	"svault/cmd/common"

	"github.com/spf13/cobra"
	"svault/internal/storage"
)

var UseCmd = &cobra.Command{
	GroupID: "ns",
	Use:     "use NAMESPACE",
	Short:   "Set the active namespace",
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return common.CompleteActiveNamespaces(cmd)
	},
	Args: cobra.ExactArgs(1),
	RunE: runUse,
}

func runUse(_ *cobra.Command, args []string) error {
	ns := args[0]

	// Validate that the namespace exists in the vault before persisting it.
	vpath, err := common.VaultPath()
	if err != nil {
		return err
	}
	encKey, err := storage.LoadSession()
	if err != nil {
		return err
	}
	vd, err := storage.ReadVault(vpath, encKey)
	if err != nil {
		return err
	}
	if _, ok := vd.Namespaces[ns]; !ok {
		return fmt.Errorf("namespace not found: %s", ns)
	}

	dir, err := common.VaultDir()
	if err != nil {
		return err
	}
	cfg, err := storage.ReadConfig(dir)
	if err != nil {
		return err
	}
	cfg.ActiveNamespace = ns
	if err := storage.WriteConfig(dir, cfg); err != nil {
		return err
	}

	fmt.Printf("Switched to namespace: %s\n", ns)
	return nil
}
