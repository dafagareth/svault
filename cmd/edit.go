// Copyright 2026 Dafa
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"svault/internal/audit"
	"svault/internal/storage"
)

var editNamespace string

var editCmd = &cobra.Command{
	Use:   "edit KEY",
	Short: "Edit a secret value in $EDITOR (good for multiline secrets)",
	Args:  cobra.ExactArgs(1),
	RunE:  runEdit,
}

func init() {
	editCmd.Flags().StringVar(&editNamespace, "ns", "default", "namespace")
	rootCmd.AddCommand(editCmd)
}

func runEdit(cmd *cobra.Command, args []string) error {
	key := args[0]
	if err := validateKey(key); err != nil {
		return err
	}
	ns := activeNamespace(cmd, editNamespace)

	vpath, err := vaultPath()
	if err != nil {
		return err
	}
	dir, err := vaultDir()
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

	err = mutateVault(func(vd *storage.VaultData) error {
		if vd.Namespaces[ns] == nil {
			vd.Namespaces[ns] = make(map[string]string)
		}
		vd.Namespaces[ns][key] = value
		return nil
	})
	if err != nil {
		return err
	}

	lpath, _ := logPath()
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
