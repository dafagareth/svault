package secrets

import (
	"fmt"
	"svault/cmd/common"

	"github.com/spf13/cobra"
	"svault/internal/audit"
	"svault/internal/storage"
)

var (
	moveFromNS string
	moveToNS   string
)

var MoveCmd = &cobra.Command{
	GroupID: "secrets",
	Use:     "move KEY",
	Short:   "Move a key from one namespace to another",
	Args:    cobra.ExactArgs(1),
	RunE:    runMove,
}

func init() {
	MoveCmd.Flags().StringVar(&moveFromNS, "from", "default", "source namespace")
	MoveCmd.Flags().StringVar(&moveToNS, "to", "", "destination namespace (required)")
	_ = MoveCmd.MarkFlagRequired("to")
}

func runMove(cmd *cobra.Command, args []string) error {
	key := args[0]
	from := common.ActiveNamespace(cmd, moveFromNS)
	to := moveToNS

	if from == to {
		return fmt.Errorf("source and destination namespaces are the same: %s", from)
	}

	err := common.MutateVault(func(vd *storage.VaultData) error {
		val, ok := vd.Namespaces[from][key]
		if !ok {
			return fmt.Errorf("key not found in [%s]: %s", from, key)
		}
		if vd.Namespaces[to] == nil {
			vd.Namespaces[to] = make(map[string]string)
		}
		if _, exists := vd.Namespaces[to][key]; exists {
			return fmt.Errorf("key already exists in [%s]: %s", to, key)
		}
		vd.Namespaces[to][key] = val
		delete(vd.Namespaces[from], key)
		return nil
	})
	if err != nil {
		return err
	}

	lpath, _ := common.LogPath()
	_ = audit.Append(lpath, audit.OpDelete, key, from)
	_ = audit.Append(lpath, audit.OpSet, key, to)

	fmt.Printf("Moved %s from [%s] to [%s]\n", key, from, to)
	return nil
}
