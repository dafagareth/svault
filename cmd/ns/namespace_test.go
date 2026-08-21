package ns

import (
	"strings"
	"testing"

	"svault/cmd/common"
	"svault/internal/storage"
)

func TestRunNs_List(t *testing.T) {
	common.SetupVault(t)
	common.Seed(t, "default", "A", "1")
	common.Seed(t, "prod", "B", "2")

	out := common.CaptureStdout(t, func() {
		if err := runNsList(common.NewCmd(), nil); err != nil {
			t.Fatalf("runNsList: %v", err)
		}
	})
	if !strings.Contains(out, "default") || !strings.Contains(out, "prod") {
		t.Errorf("expected both namespaces listed: %q", out)
	}
}

func TestRunUse(t *testing.T) {
	common.SetupVault(t)
	common.Seed(t, "staging", "TEMP", "1")
	if err := runUse(common.NewCmd(), []string{"staging"}); err != nil {
		t.Fatalf("runUse: %v", err)
	}

	dir, err := common.VaultDir()
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := storage.ReadConfig(dir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ActiveNamespace != "staging" {
		t.Errorf("active namespace = %q, want staging", cfg.ActiveNamespace)
	}
}

func TestRunDiff(t *testing.T) {
	common.SetupVault(t)
	common.Seed(t, "staging", "SAME", "1")
	common.Seed(t, "staging", "ONLY_STAGING", "2")
	common.Seed(t, "staging", "DIFF", "val1")

	common.Seed(t, "prod", "SAME", "1")
	common.Seed(t, "prod", "ONLY_PROD", "3")
	common.Seed(t, "prod", "DIFF", "val2")

	out := common.CaptureStdout(t, func() {
		if err := runDiff(common.NewCmd(), []string{"staging", "prod"}); err != nil {
			t.Fatalf("runDiff: %v", err)
		}
	})
	if !strings.Contains(out, "= SAME") {
		t.Errorf("missing SAME line: %q", out)
	}
	if !strings.Contains(out, "< ONLY_STAGING") {
		t.Errorf("missing ONLY_STAGING line: %q", out)
	}
	if !strings.Contains(out, "> ONLY_PROD") {
		t.Errorf("missing ONLY_PROD line: %q", out)
	}
	if !strings.Contains(out, "~ DIFF") {
		t.Errorf("missing DIFF line: %q", out)
	}
}
