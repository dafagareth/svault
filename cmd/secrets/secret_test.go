package secrets

import (
	"strings"
	"svault/cmd/common"
	"testing"
)

func TestRunSet_KeyValueForm(t *testing.T) {
	common.SetupVault(t)
	setQuiet = true
	defer func() { setQuiet = false }()

	if err := runSet(common.NewCmd(), []string{"DB_URL", "postgres://localhost"}); err != nil {
		t.Fatalf("runSet: %v", err)
	}
	got, ok := common.ReadSecret(t, "default", "DB_URL")
	if !ok || got != "postgres://localhost" {
		t.Errorf("got %q (ok=%v), want postgres://localhost", got, ok)
	}
}

func TestRunSet_EqualsForm(t *testing.T) {
	common.SetupVault(t)
	setQuiet = true
	defer func() { setQuiet = false }()

	if err := runSet(common.NewCmd(), []string{"API_KEY=sk-123"}); err != nil {
		t.Fatalf("runSet: %v", err)
	}
	got, ok := common.ReadSecret(t, "default", "API_KEY")
	if !ok || got != "sk-123" {
		t.Errorf("got %q (ok=%v), want sk-123", got, ok)
	}
}

func TestRunSet_EqualsFormKeepsLaterEquals(t *testing.T) {
	common.SetupVault(t)
	setQuiet = true
	defer func() { setQuiet = false }()

	// Value contains '=' (e.g. a URL with query params) — only first '=' splits.
	if err := runSet(common.NewCmd(), []string{"URL=https://x.com?a=1&b=2"}); err != nil {
		t.Fatalf("runSet: %v", err)
	}
	got, _ := common.ReadSecret(t, "default", "URL")
	if got != "https://x.com?a=1&b=2" {
		t.Errorf("got %q", got)
	}
}

func TestRunSet_InvalidEqualsForm(t *testing.T) {
	common.SetupVault(t)
	if err := runSet(common.NewCmd(), []string{"=novalue"}); err == nil {
		t.Error("expected error for '=novalue'")
	}
	if err := runSet(common.NewCmd(), []string{"nokeyequals"}); err == nil {
		t.Error("expected error for missing '='")
	}
}

func TestRunGet(t *testing.T) {
	common.SetupVault(t)
	common.Seed(t, "default", "TOKEN", "abc123")

	out := common.CaptureStdout(t, func() {
		if err := runGet(common.NewCmd(), []string{"TOKEN"}); err != nil {
			t.Fatalf("runGet: %v", err)
		}
	})
	if strings.TrimSpace(out) != "abc123" {
		t.Errorf("got %q, want abc123", strings.TrimSpace(out))
	}
}

func TestRunGet_NotFound(t *testing.T) {
	common.SetupVault(t)
	if err := runGet(common.NewCmd(), []string{"MISSING"}); err == nil {
		t.Error("expected error for missing key")
	}
}

func TestRunGet_MultipleKeys(t *testing.T) {
	common.SetupVault(t)
	common.Seed(t, "default", "HOST", "localhost")
	common.Seed(t, "default", "PORT", "5432")

	out := common.CaptureStdout(t, func() {
		if err := runGet(common.NewCmd(), []string{"HOST", "PORT"}); err != nil {
			t.Fatalf("runGet: %v", err)
		}
	})
	// With more than one key, output is KEY=VALUE per line.
	if !strings.Contains(out, "HOST=localhost") || !strings.Contains(out, "PORT=5432") {
		t.Errorf("expected KEY=VALUE lines: %q", out)
	}
}

func TestRunGet_ClipRejectsMultipleKeys(t *testing.T) {
	common.SetupVault(t)
	common.Seed(t, "default", "HOST", "localhost")
	common.Seed(t, "default", "PORT", "5432")
	getClip = true
	defer func() { getClip = false }()

	// --clip with several keys must error before touching the clipboard.
	if err := runGet(common.NewCmd(), []string{"HOST", "PORT"}); err == nil {
		t.Error("expected error for --clip with multiple keys")
	}
}

func TestRunGet_Clip(t *testing.T) {
	common.SetupVault(t)
	common.Seed(t, "default", "TOKEN", "secret-value")
	getClip = true
	getTimeout = 0 // no auto-clear, avoid spawning a background process in tests
	var copied string
	orig := common.CopyToClipboard
	common.CopyToClipboard = func(s string) error { copied = s; return nil }
	defer func() {
		getClip = false
		getTimeout = 20
		common.CopyToClipboard = orig
	}()

	out := common.CaptureStdout(t, func() {
		if err := runGet(common.NewCmd(), []string{"TOKEN"}); err != nil {
			t.Fatalf("runGet: %v", err)
		}
	})
	if copied != "secret-value" {
		t.Errorf("clipboard got %q, want secret-value", copied)
	}
	// The value itself must never be printed to stdout in clip mode.
	if strings.Contains(out, "secret-value") {
		t.Errorf("value leaked to stdout: %q", out)
	}
}

func TestRunSet_Clip(t *testing.T) {
	common.SetupVault(t)
	setQuiet = true
	setClip = true
	orig := common.ReadClipboard
	common.ReadClipboard = func() (string, error) { return "from-clipboard", nil }
	defer func() {
		setQuiet = false
		setClip = false
		common.ReadClipboard = orig
	}()

	if err := runSet(common.NewCmd(), []string{"API_KEY"}); err != nil {
		t.Fatalf("runSet: %v", err)
	}
	got, ok := common.ReadSecret(t, "default", "API_KEY")
	if !ok || got != "from-clipboard" {
		t.Errorf("got %q (ok=%v), want from-clipboard", got, ok)
	}
}

func TestRunSet_ClipRejectsEqualsForm(t *testing.T) {
	common.SetupVault(t)
	setClip = true
	orig := common.ReadClipboard
	common.ReadClipboard = func() (string, error) { return "x", nil }
	defer func() {
		setClip = false
		common.ReadClipboard = orig
	}()

	if err := runSet(common.NewCmd(), []string{"KEY=VALUE"}); err == nil {
		t.Error("expected error: --clip must receive only the KEY")
	}
}

func TestRunSet_StdinClipMutuallyExclusive(t *testing.T) {
	common.SetupVault(t)
	setStdin = true
	setClip = true
	defer func() {
		setStdin = false
		setClip = false
	}()

	if err := runSet(common.NewCmd(), []string{"KEY"}); err == nil {
		t.Error("expected error when both --stdin and --clip are set")
	}
}

func TestRunDelete(t *testing.T) {
	common.SetupVault(t)
	common.Seed(t, "default", "TEMP", "x")

	out := common.CaptureStdout(t, func() {
		if err := runDelete(common.NewCmd(), []string{"TEMP"}); err != nil {
			t.Fatalf("runDelete: %v", err)
		}
	})
	if !strings.Contains(out, "Deleted TEMP") {
		t.Errorf("unexpected output: %q", out)
	}
	if _, ok := common.ReadSecret(t, "default", "TEMP"); ok {
		t.Error("key still present after delete")
	}
}

func TestRunDelete_NotFound(t *testing.T) {
	common.SetupVault(t)
	if err := runDelete(common.NewCmd(), []string{"NOPE"}); err == nil {
		t.Error("expected error for missing key")
	}
}

func TestRunRename(t *testing.T) {
	common.SetupVault(t)
	common.Seed(t, "default", "OLD", "val")

	if err := runRename(common.NewCmd(), []string{"OLD", "NEW"}); err != nil {
		t.Fatalf("runRename: %v", err)
	}
	if _, ok := common.ReadSecret(t, "default", "OLD"); ok {
		t.Error("OLD still present")
	}
	got, ok := common.ReadSecret(t, "default", "NEW")
	if !ok || got != "val" {
		t.Errorf("NEW = %q (ok=%v), want val", got, ok)
	}
}

func TestRunRename_SourceMissing(t *testing.T) {
	common.SetupVault(t)
	if err := runRename(common.NewCmd(), []string{"GHOST", "NEW"}); err == nil {
		t.Error("expected error when source key missing")
	}
}

func TestRunRename_TargetExists(t *testing.T) {
	common.SetupVault(t)
	common.Seed(t, "default", "A", "1")
	common.Seed(t, "default", "B", "2")
	if err := runRename(common.NewCmd(), []string{"A", "B"}); err == nil {
		t.Error("expected error when target key already exists")
	}
}

func TestRunMove(t *testing.T) {
	common.SetupVault(t)
	common.Seed(t, "default", "SHARED", "v")
	moveToNS = "prod"
	defer func() { moveToNS = "" }()

	if err := runMove(common.NewCmd(), []string{"SHARED"}); err != nil {
		t.Fatalf("runMove: %v", err)
	}
	if _, ok := common.ReadSecret(t, "default", "SHARED"); ok {
		t.Error("key still in source namespace")
	}
	got, ok := common.ReadSecret(t, "prod", "SHARED")
	if !ok || got != "v" {
		t.Errorf("prod/SHARED = %q (ok=%v), want v", got, ok)
	}
}

func TestRunMove_SourceMissing(t *testing.T) {
	common.SetupVault(t)
	moveToNS = "prod"
	defer func() { moveToNS = "" }()
	if err := runMove(common.NewCmd(), []string{"GHOST"}); err == nil {
		t.Error("expected error when source key missing")
	}
}

func TestRunMove_SameNamespace(t *testing.T) {
	common.SetupVault(t)
	common.Seed(t, "default", "K", "v")
	moveToNS = "default"
	defer func() { moveToNS = "" }()
	if err := runMove(common.NewCmd(), []string{"K"}); err == nil {
		t.Error("expected error when source and dest namespaces match")
	}
}
