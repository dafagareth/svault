package auth

import (
	"fmt"
	"os"
	"svault/cmd/common"

	"github.com/spf13/cobra"
	"golang.org/x/term"
	"svault/internal/audit"
	"svault/internal/storage"
)

var InitCmd = &cobra.Command{
	GroupID: "auth",
	Use:     "init",
	Short:   "Initialize a new vault with a master password",
	RunE:    runInit,
}

func runInit(_ *cobra.Command, _ []string) error {
	vpath, err := common.VaultPath()
	if err != nil {
		return err
	}
	if _, err := os.Stat(vpath); err == nil {
		return fmt.Errorf("vault already exists at %s", vpath)
	}

	fmt.Print("Master password: ")
	pw1, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		return fmt.Errorf("read password: %w", err)
	}
	defer common.Wipe(pw1)

	if len(pw1) == 0 {
		return fmt.Errorf("password cannot be empty")
	}

	fmt.Print("Confirm password: ")
	pw2, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		return fmt.Errorf("read confirmation: %w", err)
	}
	defer common.Wipe(pw2)

	if string(pw1) != string(pw2) {
		return fmt.Errorf("passwords do not match")
	}

	if err := storage.InitVault(vpath, pw1); err != nil {
		return err
	}

	lpath, _ := common.LogPath()
	_ = audit.Append(lpath, audit.OpInit, "-", "default")

	fmt.Printf("Vault initialized at %s\n", vpath)
	return nil
}
