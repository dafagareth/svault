package auth

import (
	"fmt"
	"svault/cmd/common"

	"github.com/spf13/cobra"
	"svault/internal/audit"
	"svault/internal/storage"
)

var LockCmd = &cobra.Command{
	GroupID: "auth",
	Use:     "lock",
	Short:   "Lock the vault and delete the session",
	RunE:    runLock,
}

func runLock(_ *cobra.Command, _ []string) error {
	if err := storage.DeleteSession(); err != nil {
		return err
	}

	lpath, _ := common.LogPath()
	_ = audit.Append(lpath, audit.OpLock, "-", "-")

	fmt.Println("Vault locked.")
	return nil
}
