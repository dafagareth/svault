package utils

import (
	"fmt"
	"svault/cmd/common"

	"github.com/spf13/cobra"
	"svault/internal/storage"
)

var copyNamespace string

var CopyCmd = &cobra.Command{
	GroupID: "utils",
	Use:     "copy KEY",
	Short:   "Copy a secret value to clipboard (auto-clears after 30s)",
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return common.CompleteActiveKeys(cmd)
	},
	Args: cobra.ExactArgs(1),
	RunE: runCopy,
}

func init() {
	CopyCmd.Flags().StringVar(&copyNamespace, "ns", "default", "namespace")
}

func runCopy(cmd *cobra.Command, args []string) error {
	key := args[0]
	ns := common.ActiveNamespace(cmd, copyNamespace)

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

	val, ok := vd.Namespaces[ns][key]
	if !ok {
		return fmt.Errorf("key not found: %s", key)
	}

	if err := common.CopyToClipboard(val); err != nil {
		return fmt.Errorf("clipboard: %w", err)
	}

	common.ScheduleClipboardClear(30)
	fmt.Printf("Copied %s to clipboard. Will clear in 30 seconds.\n", key)
	return nil
}
