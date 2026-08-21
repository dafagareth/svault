package system

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"svault/cmd/common"
)

func TestRunBackupAndRestore(t *testing.T) {
	common.SetupVault(t)
	common.Seed(t, "default", "KEEP", "value")

	dir := t.TempDir()
	backupFile := filepath.Join(dir, "vault.bak")

	out := common.CaptureStdout(t, func() {
		if err := runBackup(common.NewCmd(), []string{backupFile}); err != nil {
			t.Fatalf("runBackup: %v", err)
		}
	})
	if !strings.Contains(out, "backed up") {
		t.Errorf("unexpected backup output: %q", out)
	}
	if _, err := os.Stat(backupFile); err != nil {
		t.Fatalf("backup file not created: %v", err)
	}

	common.Seed(t, "default", "TEMP", "should-disappear")

	_ = common.CaptureStdout(t, func() {
		if err := runRestore(common.NewCmd(), []string{backupFile}); err != nil {
			t.Fatalf("runRestore: %v", err)
		}
	})

	common.Reunlock(t)
	if _, ok := common.ReadSecret(t, "default", "TEMP"); ok {
		t.Error("TEMP should be gone after restore")
	}
	if v, _ := common.ReadSecret(t, "default", "KEEP"); v != "value" {
		t.Errorf("KEEP = %q after restore, want value", v)
	}
}

func TestRunBackup_AutoName(t *testing.T) {
	common.SetupVault(t)

	out := common.CaptureStdout(t, func() {
		if err := runBackup(common.NewCmd(), nil); err != nil {
			t.Fatalf("runBackup auto: %v", err)
		}
	})
	if !strings.Contains(out, "vault.enc.backup-") {
		t.Errorf("expected timestamped backup name: %q", out)
	}
}

func TestRunDoctor(t *testing.T) {
	common.SetupVault(t)

	out := common.CaptureStdout(t, func() {
		if err := runDoctor(common.NewCmd(), nil); err != nil {
			t.Fatalf("runDoctor: %v", err)
		}
	})
	if !strings.Contains(out, "vault directory") || !strings.Contains(out, "vault file") {
		t.Errorf("unexpected doctor output: %q", out)
	}
}

func TestRunInfo(t *testing.T) {
	common.SetupVault(t)
	common.Seed(t, "default", "SECRET1", "val1")

	out := common.CaptureStdout(t, func() {
		if err := runInfo(common.NewCmd(), nil); err != nil {
			t.Fatalf("runInfo: %v", err)
		}
	})
	if !strings.Contains(out, "Vault file") || !strings.Contains(out, "Namespace") {
		t.Errorf("unexpected info output: %q", out)
	}
}

func TestRunVersion(t *testing.T) {
	Version = "1.0.0-test"

	out := common.CaptureStdout(t, func() {
		VersionCmd.Run(common.NewCmd(), nil)
	})
	if !strings.Contains(out, "1.0.0-test") {
		t.Errorf("expected version 1.0.0-test in output: %q", out)
	}
}

func TestRunLog(t *testing.T) {
	common.SetupVault(t)
	logTail = 10
	defer func() { logTail = 0 }()

	out := common.CaptureStdout(t, func() {
		if err := runLog(common.NewCmd(), nil); err != nil {
			t.Fatalf("runLog: %v", err)
		}
	})
	if !strings.Contains(out, "SET") && !strings.Contains(out, "empty") {
		t.Errorf("unexpected log output: %q", out)
	}
}
