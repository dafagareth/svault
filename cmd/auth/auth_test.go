package auth

import (
	"strings"
	"testing"

	"svault/cmd/common"
	"svault/internal/storage"
)

func TestRunLock(t *testing.T) {
	common.SetupVault(t)
	out := common.CaptureStdout(t, func() {
		if err := runLock(common.NewCmd(), nil); err != nil {
			t.Fatalf("runLock: %v", err)
		}
	})
	if !strings.Contains(out, "Vault locked") {
		t.Errorf("expected 'Vault locked' output, got: %q", out)
	}

	_, err := storage.LoadSession()
	if err == nil {
		t.Fatal("expected error loading session after lock, got nil")
	}
}

func TestRunStatus_ShortUnlocked(t *testing.T) {
	common.SetupVault(t)
	statusShort = true
	defer func() { statusShort = false }()

	out := common.CaptureStdout(t, func() {
		_ = runStatus(common.NewCmd(), nil)
	})
	if !strings.Contains(out, "🔓") {
		t.Errorf("expected unlock icon: %q", out)
	}
}
