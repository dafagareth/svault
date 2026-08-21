package secrets

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"svault/cmd/common"

	"github.com/spf13/cobra"
	"svault/internal/audit"
	"svault/internal/storage"
)

var editNamespace string

var EditCmd = &cobra.Command{
	GroupID: "secrets",
	Use:     "edit KEY",
	Short:   "Edit a secret value in $EDITOR (good for multiline secrets)",
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return common.CompleteActiveKeys(cmd)
	},
	Args: cobra.ExactArgs(1),
	RunE: runEdit,
}

func init() {
	EditCmd.Flags().StringVar(&editNamespace, "ns", "default", "namespace")
}

func runEdit(cmd *cobra.Command, args []string) error {
	key := args[0]
	if err := common.ValidateKey(key); err != nil {
		return err
	}
	ns := common.ActiveNamespace(cmd, editNamespace)

	vpath, err := common.VaultPath()
	if err != nil {
		return err
	}
	dir, err := common.VaultDir()
	if err != nil {
		return err
	}
	encKey, err := storage.LoadSession()
	if err != nil {
		return err
	}

	// Read the current value just to seed the editor. We do not hold the vault
	// lock during editing, since the editor session can be open arbitrarily
	// long; the actual write below re-reads under lock.
	vd, err := storage.ReadVault(vpath, encKey)
	if err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, "edit-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	_ = os.Chmod(tmpPath, 0600)

	if cur, ok := vd.Namespaces[ns][key]; ok {
		if _, err := tmp.WriteString(cur); err != nil {
			tmp.Close()
			return fmt.Errorf("write temp file: %w", err)
		}
	}
	tmp.Close()

	if err := openEditor(tmpPath); err != nil {
		return err
	}

	edited, err := os.ReadFile(tmpPath)
	if err != nil {
		return fmt.Errorf("read edited value: %w", err)
	}
	// Strip a single trailing newline added by most editors.
	value := strings.TrimSuffix(string(edited), "\n")

	err = common.MutateVault(func(vd *storage.VaultData) error {
		if vd.Namespaces[ns] == nil {
			vd.Namespaces[ns] = make(map[string]string)
		}
		vd.Namespaces[ns][key] = value
		return nil
	})
	if err != nil {
		return err
	}

	lpath, _ := common.LogPath()
	_ = audit.Append(lpath, audit.OpSet, key, ns)

	fmt.Printf("OK  %s saved\n", key)
	return nil
}

// openEditor launches $EDITOR (or a sensible default) on path and waits.
func openEditor(path string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		if _, err := exec.LookPath("nano"); err == nil {
			editor = "nano"
		} else {
			editor = "vi"
		}
	}
	// Support editors invoked with args, e.g. EDITOR="code --wait".
	parts := strings.Fields(editor)
	bin := parts[0]
	cmdArgs := append(parts[1:], filepath.Clean(path))

	c := exec.Command(bin, cmdArgs...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		return fmt.Errorf("editor exited with error: %w", err)
	}
	return nil
}
