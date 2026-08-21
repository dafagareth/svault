package system

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"svault/cmd/common"

	"github.com/spf13/cobra"
	"svault/internal/storage"
)

var DoctorCmd = &cobra.Command{
	GroupID: "system",
	Use:     "doctor",
	Short:   "Check svault installation and vault health",
	RunE:    runDoctor,
}

func runDoctor(_ *cobra.Command, _ []string) error {
	fails := 0
	warns := 0

	print := func(status, label, detail string) {
		switch status {
		case "FAIL":
			fails++
		case "WARN":
			warns++
		}
		fmt.Printf("[%s] %-28s %s\n", status, label, detail)
	}

	vdir, err := common.VaultDir()
	if err != nil {
		return err
	}
	vpath, _ := common.VaultPath()
	cfgPath := filepath.Join(vdir, "config.json")
	logfile, _ := common.LogPath()

	// Vault directory
	if info, err := os.Stat(vdir); err != nil || !info.IsDir() {
		print("FAIL", "vault directory", vdir+" not found — run 'svault init'")
	} else {
		print("OK  ", "vault directory", vdir)
	}

	// vault.enc
	if info, err := os.Stat(vpath); err != nil {
		print("FAIL", "vault file", "not found — run 'svault init'")
	} else {
		perm := info.Mode().Perm()
		if perm&0o077 != 0 {
			print("WARN", "vault file", fmt.Sprintf("permissions %04o (should be 0600)", perm))
		} else {
			print("OK  ", "vault file", fmt.Sprintf("%d bytes, mode 0600", info.Size()))
		}
	}

	// config.json
	if _, err := os.Stat(cfgPath); err != nil {
		print("WARN", "config file", "not found — will use defaults")
	} else {
		print("OK  ", "config file", cfgPath)
	}

	// vault.log
	if _, err := os.Stat(logfile); err != nil {
		print("WARN", "audit log", "no log yet (created on first write)")
	} else {
		print("OK  ", "audit log", logfile)
	}

	// Session
	remaining, active := storage.SessionRemaining()
	if active {
		print("OK  ", "session", fmt.Sprintf("unlocked, %dm remaining", int(remaining.Minutes())))
	} else {
		print("WARN", "session", "locked — run 'svault unlock'")
	}

	// git
	if _, err := exec.LookPath("git"); err != nil {
		print("WARN", "git (auto-namespace)", "not found — namespace detection disabled")
	} else {
		if ns := common.GitRepoName(); ns != "" {
			print("OK  ", "git (auto-namespace)", "current namespace: "+ns)
		} else {
			print("OK  ", "git (auto-namespace)", "found (not inside a git repo)")
		}
	}

	// Clipboard
	if backend, ok := common.AvailableClipboardBackend(); !ok {
		print("WARN", "clipboard", "no tool found (wl-copy / xsel / xclip / pbcopy / clip.exe)")
	} else {
		print("OK  ", "clipboard", backend.CopyCmd[0])
	}

	fmt.Println()
	switch {
	case fails > 0:
		fmt.Printf("%d critical issue(s) found. Fix FAIL items above.\n", fails)
	case warns > 0:
		fmt.Printf("All good, %d warning(s). See WARN items above.\n", warns)
	default:
		fmt.Println("All checks passed.")
	}
	return nil
}
