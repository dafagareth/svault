package auth

import (
	"strings"
	"svault/cmd/common"
	"testing"

	"svault/internal/storage"
)

func TestRunStatus_Unlocked(t *testing.T) {
	common.SetupVault(t) // setupVault saves a session, so we are unlocked
	statusShort = false

	out := common.CaptureStdout(t, func() {
		if err := runStatus(common.NewCmd(), nil); err != nil {
			t.Fatalf("runStatus: %v", err)
		}
	})
	if !strings.Contains(out, "Unlocked") {
		t.Errorf("expected Unlocked status: %q", out)
	}
}

func TestRunStatus_UnlockedShowsStats(t *testing.T) {
	common.SetupVault(t)
	statusShort = false
	common.Seed(t, "default", "A", "1")
	common.Seed(t, "prod", "B", "2")

	out := common.CaptureStdout(t, func() {
		if err := runStatus(common.NewCmd(), nil); err != nil {
			t.Fatalf("runStatus: %v", err)
		}
	})
	for _, want := range []string{"Namespaces:", "Total keys:", "Active ns:", "Vault size:"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in output: %q", want, out)
		}
	}
}

func TestHumanSize(t *testing.T) {
	cases := map[int64]string{
		512:             "512 B",
		2048:            "2.0 KB",
		3 * 1024 * 1024: "3.0 MB",
	}
	for in, want := range cases {
		if got := humanSize(in); got != want {
			t.Errorf("humanSize(%d) = %q, want %q", in, got, want)
		}
	}
}

func TestRunStatus_Locked(t *testing.T) {
	common.SetupVault(t)
	if err := storage.DeleteSession(); err != nil {
		t.Fatal(err)
	}
	statusShort = false

	out := common.CaptureStdout(t, func() {
		if err := runStatus(common.NewCmd(), nil); err != nil {
			t.Fatalf("runStatus: %v", err)
		}
	})
	if !strings.Contains(out, "Locked") {
		t.Errorf("expected Locked status: %q", out)
	}
}

func TestRunStatus_ShortLocked(t *testing.T) {
	common.SetupVault(t)
	_ = storage.DeleteSession()
	statusShort = true
	defer func() { statusShort = false }()

	out := common.CaptureStdout(t, func() {
		_ = runStatus(common.NewCmd(), nil)
	})
	if !strings.Contains(out, "🔒") {
		t.Errorf("expected lock emoji: %q", out)
	}
}
