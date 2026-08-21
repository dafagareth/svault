package env

import (
	"os"
	"path/filepath"
	"strings"
	"svault/cmd/common"
	"testing"
)

// writeEnvFile writes content to a temp .env file and returns its path.
func writeEnvFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), ".env.example")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestRunCheck_AllPresent(t *testing.T) {
	common.SetupVault(t)
	common.Seed(t, "default", "DB_HOST", "h")
	file := writeEnvFile(t, "DB_HOST=\n")

	out := common.CaptureStdout(t, func() {
		if err := runCheck(common.NewCmd(), []string{file}); err != nil {
			t.Fatalf("runCheck: %v", err)
		}
	})
	if !strings.Contains(out, "OK    DB_HOST") {
		t.Errorf("expected OK for DB_HOST: %q", out)
	}
	if strings.Contains(out, "missing") {
		t.Errorf("did not expect missing: %q", out)
	}
}

func TestRunCheck_Strict(t *testing.T) {
	common.SetupVault(t)
	defer func() { checkStrict = false }()
	common.Seed(t, "default", "DB_HOST", "h")
	file := writeEnvFile(t, "DB_HOST=\nSTRIPE_KEY=\n")
	checkStrict = true

	err := captureStdoutErr(t, func() error {
		return runCheck(common.NewCmd(), []string{file})
	})
	if err == nil {
		t.Error("expected non-zero error with --strict when a key is missing")
	}
}

func TestRunCheck_Extra(t *testing.T) {
	common.SetupVault(t)
	defer func() { checkExtra = false }()
	common.Seed(t, "default", "DB_HOST", "h")
	common.Seed(t, "default", "LEGACY_TOKEN", "x") // present in vault, not in file
	file := writeEnvFile(t, "DB_HOST=\n")
	checkExtra = true

	out := common.CaptureStdout(t, func() {
		if err := runCheck(common.NewCmd(), []string{file}); err != nil {
			t.Fatalf("runCheck: %v", err)
		}
	})
	if !strings.Contains(out, "EXTRA LEGACY_TOKEN") {
		t.Errorf("expected EXTRA LEGACY_TOKEN: %q", out)
	}
}

// captureStdoutErr is like captureStdout but also returns the error from fn,
// for commands whose exit code (not just output) matters.
func captureStdoutErr(t *testing.T, fn func() error) error {
	t.Helper()
	var err error
	common.CaptureStdout(t, func() { err = fn() })
	return err
}
