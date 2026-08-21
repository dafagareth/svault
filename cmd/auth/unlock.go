package auth

import (
	"fmt"
	"os"
	"svault/cmd/common"

	"github.com/spf13/cobra"
	"golang.org/x/term"
	"svault/internal/audit"
	"svault/internal/crypto"
	"svault/internal/storage"
)

var UnlockCmd = &cobra.Command{
	GroupID: "auth",
	Use:     "unlock",
	Short:   "Unlock the vault and start a session",
	RunE:    runUnlock,
}

func runUnlock(_ *cobra.Command, _ []string) error {
	vpath, err := common.VaultPath()
	if err != nil {
		return err
	}

	salt, err := storage.ReadSalt(vpath)
	if err != nil {
		return fmt.Errorf("vault not initialized, run 'svault init'")
	}

	fmt.Print("Master password: ")
	pw, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		return fmt.Errorf("read password: %w", err)
	}
	defer common.Wipe(pw)

	key := crypto.DeriveKey(pw, salt)

	if _, err := storage.ReadVault(vpath, key); err != nil {
		return fmt.Errorf("wrong password")
	}

	if err := storage.SaveSession(key); err != nil {
		return err
	}

	lpath, _ := common.LogPath()
	_ = audit.Append(lpath, audit.OpUnlock, "-", "default")

	fmt.Println("Vault unlocked. Session valid for 30 minutes.")
	return nil
}
