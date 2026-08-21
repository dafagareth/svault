package system

import (
	"fmt"
	"os"
	"svault/cmd/common"

	"github.com/spf13/cobra"
	"golang.org/x/term"
	"svault/internal/audit"
	"svault/internal/storage"
)

var RotateCmd = &cobra.Command{
	GroupID: "system",
	Use:     "rotate",
	Short:   "Change the master password and re-encrypt the vault",
	RunE:    runRotate,
}

func runRotate(_ *cobra.Command, _ []string) error {
	vpath, err := common.VaultPath()
	if err != nil {
		return err
	}

	currentKey, err := storage.LoadSession()
	if err != nil {
		return err
	}

	fmt.Print("New password: ")
	pw1, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		return fmt.Errorf("read password: %w", err)
	}
	defer common.Wipe(pw1)

	if len(pw1) == 0 {
		return fmt.Errorf("password cannot be empty")
	}

	fmt.Print("Confirm new password: ")
	pw2, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		return fmt.Errorf("read confirmation: %w", err)
	}
	defer common.Wipe(pw2)

	if string(pw1) != string(pw2) {
		return fmt.Errorf("passwords do not match")
	}

	dir, err := common.VaultDir()
	if err != nil {
		return err
	}
	var newKey []byte
	err = storage.WithVaultLock(dir, func() error {
		k, err := storage.RotateVault(vpath, currentKey, pw1)
		if err != nil {
			return err
		}
		newKey = k
		return nil
	})
	if err != nil {
		return err
	}

	if err := storage.SaveSession(newKey); err != nil {
		return err
	}

	lpath, _ := common.LogPath()
	_ = audit.Append(lpath, audit.OpRotate, "-", "-")

	fmt.Println("Master password changed. Session updated.")
	return nil
}
