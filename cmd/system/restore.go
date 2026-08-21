package system

import (
	"fmt"
	"os"
	"path/filepath"
	"svault/cmd/common"

	"github.com/spf13/cobra"
	"svault/internal/crypto"
	"svault/internal/storage"
)

var RestoreCmd = &cobra.Command{
	GroupID: "system",
	Use:     "restore FILE",
	Short:   "Restore the vault from a backup file",
	Args:    cobra.ExactArgs(1),
	RunE:    runRestore,
}

func runRestore(_ *cobra.Command, args []string) error {
	backupPath := args[0]

	vpath, err := common.VaultPath()
	if err != nil {
		return err
	}
	dir, err := common.VaultDir()
	if err != nil {
		return err
	}

	// Read the backup before locking so a bad path fails fast.
	src, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("read backup: %w", err)
	}

	// Validate that the file is a structurally valid svault vault
	// (at minimum: long enough to contain a salt header).
	if len(src) <= crypto.SaltSize {
		return fmt.Errorf("invalid backup: file is too small to be a vault")
	}

	err = storage.WithVaultLock(dir, func() error {
		// Preserve the current vault as a pre-restore safety copy.
		if current, err := os.ReadFile(vpath); err == nil {
			_ = os.WriteFile(vpath+".pre-restore", current, 0600)
		}

		// Atomic write: write to a temp file then rename so that a crash
		// mid-restore does not leave a partially written vault.enc.
		tmpFile, err := os.CreateTemp(dir, ".vault-restore-*.tmp")
		if err != nil {
			return fmt.Errorf("create temp restore: %w", err)
		}
		tmpName := tmpFile.Name()
		defer os.Remove(tmpName)

		if err := tmpFile.Chmod(0600); err != nil {
			tmpFile.Close()
			return fmt.Errorf("chmod temp restore: %w", err)
		}
		if _, err := tmpFile.Write(src); err != nil {
			tmpFile.Close()
			return fmt.Errorf("write temp restore: %w", err)
		}
		if err := tmpFile.Sync(); err != nil {
			tmpFile.Close()
			return fmt.Errorf("sync temp restore: %w", err)
		}
		if err := tmpFile.Close(); err != nil {
			return fmt.Errorf("close temp restore: %w", err)
		}
		if err := os.Rename(tmpName, vpath); err != nil {
			return fmt.Errorf("atomic rename restore: %w", err)
		}
		return nil
	})
	if err != nil {
		return err
	}

	_ = storage.DeleteSession()

	fmt.Printf("Vault restored from %s\n", filepath.Base(backupPath))
	fmt.Println("Session cleared. Run 'svault unlock' to start a new session.")
	return nil
}
