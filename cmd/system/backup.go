package system

import (
	"fmt"
	"os"
	"path/filepath"
	"svault/cmd/common"
	"time"

	"github.com/spf13/cobra"
)

var BackupCmd = &cobra.Command{
	GroupID: "system",
	Use:     "backup [FILE]",
	Short:   "Back up the vault to a file",
	Args:    cobra.MaximumNArgs(1),
	RunE:    runBackup,
}

func runBackup(_ *cobra.Command, args []string) error {
	vpath, err := common.VaultPath()
	if err != nil {
		return err
	}

	dest := ""
	if len(args) == 1 {
		dest = args[0]
	} else {
		dir, err := common.VaultDir()
		if err != nil {
			return err
		}
		ts := time.Now().Format("20060102-150405")
		dest = filepath.Join(dir, "vault.enc.backup-"+ts)
	}

	src, err := os.ReadFile(vpath)
	if err != nil {
		return fmt.Errorf("read vault: %w", err)
	}

	// Atomic write: write to a temp file in the same directory, then rename,
	// so a crash mid-backup does not produce a partially written backup file.
	dir := filepath.Dir(dest)
	if dir == "" {
		dir = "."
	}
	tmpFile, err := os.CreateTemp(dir, ".backup-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp backup: %w", err)
	}
	tmpName := tmpFile.Name()
	defer os.Remove(tmpName)

	if err := tmpFile.Chmod(0600); err != nil {
		tmpFile.Close()
		return fmt.Errorf("chmod temp backup: %w", err)
	}
	if _, err := tmpFile.Write(src); err != nil {
		tmpFile.Close()
		return fmt.Errorf("write temp backup: %w", err)
	}
	if err := tmpFile.Sync(); err != nil {
		tmpFile.Close()
		return fmt.Errorf("sync temp backup: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("close temp backup: %w", err)
	}
	if err := os.Rename(tmpName, dest); err != nil {
		return fmt.Errorf("atomic rename backup: %w", err)
	}

	fmt.Printf("Vault backed up to %s\n", dest)
	return nil
}
