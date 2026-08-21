package common

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"svault/internal/storage"
)

func Wipe(b []byte) {
	for i := range b {
		b[i] = 0
	}
}

func VaultDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}
	return filepath.Join(home, ".svault"), nil
}

func VaultPath() (string, error) {
	dir, err := VaultDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "vault.enc"), nil
}

func LogPath() (string, error) {
	dir, err := VaultDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "vault.log"), nil
}

func GitRepoName() string {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return ""
	}
	return filepath.Base(strings.TrimSpace(string(out)))
}

func NamespaceSource() (string, string) {
	if ns := os.Getenv("SVAULT_NS"); ns != "" {
		return ns, "env"
	}
	if ns := GitRepoName(); ns != "" {
		return ns, "git"
	}
	dir, err := VaultDir()
	if err != nil {
		return "default", "default"
	}
	cfg, err := storage.ReadConfig(dir)
	if err != nil {
		return "default", "default"
	}
	if cfg.ActiveNamespace != "" && cfg.ActiveNamespace != "default" {
		return cfg.ActiveNamespace, "config"
	}
	return "default", "default"
}

func ActiveNamespace(cmd *cobra.Command, flagVal string) string {
	if cmd != nil && cmd.Flags().Changed("ns") {
		return flagVal
	}
	ns, _ := NamespaceSource()
	return ns
}

func ValidateKey(key string) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}
	for i, r := range key {
		isLetter := (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '_'
		isDigit := r >= '0' && r <= '9'
		if i == 0 && !isLetter {
			return fmt.Errorf("invalid key %q: must start with a letter or underscore", key)
		}
		if !isLetter && !isDigit {
			return fmt.Errorf("invalid key %q: only letters, digits, and underscore allowed", key)
		}
	}
	return nil
}

func MutateVault(fn func(vd *storage.VaultData) error) error {
	vpath, err := VaultPath()
	if err != nil {
		return err
	}
	dir, err := VaultDir()
	if err != nil {
		return err
	}
	encKey, err := storage.LoadSession()
	if err != nil {
		return err
	}
	return storage.WithVaultLock(dir, func() error {
		vd, err := storage.ReadVault(vpath, encKey)
		if err != nil {
			return err
		}
		if err := fn(vd); err != nil {
			return err
		}
		return storage.WriteVault(vpath, encKey, vd)
	})
}

func CompleteActiveKeys(cmd *cobra.Command) ([]string, cobra.ShellCompDirective) {
	vpath, err := VaultPath()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	encKey, err := storage.LoadSession()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	vd, err := storage.ReadVault(vpath, encKey)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	ns := ActiveNamespace(cmd, "")
	keysMap, ok := vd.Namespaces[ns]
	if !ok {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	var keys []string
	for k := range keysMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys, cobra.ShellCompDirectiveNoFileComp
}

func CompleteActiveNamespaces(cmd *cobra.Command) ([]string, cobra.ShellCompDirective) {
	vpath, err := VaultPath()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	encKey, err := storage.LoadSession()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	vd, err := storage.ReadVault(vpath, encKey)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var nss []string
	for ns := range vd.Namespaces {
		nss = append(nss, ns)
	}
	sort.Strings(nss)
	return nss, cobra.ShellCompDirectiveNoFileComp
}

type ClipboardBackend struct {
	CopyCmd  []string
	PasteCmd []string
}

var clipboardBackends = []ClipboardBackend{
	{CopyCmd: []string{"wl-copy"}, PasteCmd: []string{"wl-paste", "--no-newline"}},
	{CopyCmd: []string{"xclip", "-selection", "clipboard"}, PasteCmd: []string{"xclip", "-selection", "clipboard", "-o"}},
	{CopyCmd: []string{"xsel", "--clipboard", "--input"}, PasteCmd: []string{"xsel", "--clipboard", "--output"}},
	{CopyCmd: []string{"pbcopy"}, PasteCmd: []string{"pbpaste"}},
}

func AvailableClipboardBackend() (ClipboardBackend, bool) {
	for _, t := range clipboardBackends {
		if _, err := exec.LookPath(t.CopyCmd[0]); err == nil {
			return t, true
		}
	}
	return ClipboardBackend{}, false
}

func ClipboardAvailable() bool {
	_, ok := AvailableClipboardBackend()
	return ok
}

// ScheduleClipboardClear spawns a detached process that overwrites the clipboard
// with an empty string after the given number of seconds. The secret is never
// passed as a process argument.
func ScheduleClipboardClear(seconds int) {
	backend, ok := AvailableClipboardBackend()
	if !ok {
		return
	}
	script := fmt.Sprintf("sleep %d; printf '' | %s", seconds, strings.Join(backend.CopyCmd, " "))
	_ = exec.Command("sh", "-c", script).Start()
}

var CopyToClipboard = RealCopyToClipboard
var ReadClipboard = RealReadClipboard

func RealCopyToClipboard(s string) error {
	t, ok := AvailableClipboardBackend()
	if !ok {
		return fmt.Errorf("no clipboard tool found (install wl-clipboard, xclip, or xsel)")
	}
	cmd := exec.Command(t.CopyCmd[0], t.CopyCmd[1:]...)
	cmd.Stdin = strings.NewReader(s)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("clipboard copy failed: %w", err)
	}
	return nil
}

func RealReadClipboard() (string, error) {
	t, ok := AvailableClipboardBackend()
	if !ok {
		return "", fmt.Errorf("no clipboard tool found (install wl-clipboard, xclip, or xsel)")
	}
	out, err := exec.Command(t.PasteCmd[0], t.PasteCmd[1:]...).Output()
	if err != nil {
		return "", fmt.Errorf("clipboard read failed: %w", err)
	}
	return strings.TrimRight(string(out), "\n"), nil
}
