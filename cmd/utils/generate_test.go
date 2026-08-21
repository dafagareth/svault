package utils

import (
	"strings"
	"svault/cmd/common"
	"testing"
)

// resetGenerateFlags restores generate's package-level flags to safe defaults
// so tests do not leak state into each other or touch the clipboard.
func resetGenerateFlags() {
	genLength = 24
	genNoSymbols = false
	genSaveKey = ""
	genNamespace = "default"
	genNoCopy = true // never touch the real clipboard in tests
}

func TestRunGenerate_LengthAndCharset(t *testing.T) {
	common.SetupVault(t)
	resetGenerateFlags()
	genLength = 32
	genNoSymbols = true

	out := common.CaptureStdout(t, func() {
		if err := runGenerate(common.NewCmd(), nil); err != nil {
			t.Fatalf("runGenerate: %v", err)
		}
	})
	password := strings.TrimSpace(out)
	if len(password) != 32 {
		t.Errorf("length = %d, want 32", len(password))
	}
	const allowed = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	for _, c := range password {
		if !strings.ContainsRune(allowed, c) {
			t.Errorf("unexpected symbol %q in --no-symbols output", c)
		}
	}
}

func TestRunGenerate_Save(t *testing.T) {
	common.SetupVault(t)
	resetGenerateFlags()
	genSaveKey = "NEW_PASS"

	out := common.CaptureStdout(t, func() {
		if err := runGenerate(common.NewCmd(), nil); err != nil {
			t.Fatalf("runGenerate: %v", err)
		}
	})
	password := strings.SplitN(strings.TrimSpace(out), "\n", 2)[0]

	stored, ok := common.ReadSecret(t, "default", "NEW_PASS")
	if !ok {
		t.Fatal("generated password not saved")
	}
	if stored != password {
		t.Errorf("stored %q != printed %q", stored, password)
	}
}

func TestRunGenerate_TooShort(t *testing.T) {
	common.SetupVault(t)
	resetGenerateFlags()
	genLength = 4
	if err := runGenerate(common.NewCmd(), nil); err == nil {
		t.Error("expected error for length < 8")
	}
}

func TestRunGenerate_Randomness(t *testing.T) {
	common.SetupVault(t)
	resetGenerateFlags()

	first := common.CaptureStdout(t, func() { _ = runGenerate(common.NewCmd(), nil) })
	second := common.CaptureStdout(t, func() { _ = runGenerate(common.NewCmd(), nil) })
	if strings.TrimSpace(first) == strings.TrimSpace(second) {
		t.Error("two generated passwords were identical")
	}
}
