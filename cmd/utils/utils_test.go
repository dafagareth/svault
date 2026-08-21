package utils

import (
	"strings"
	"svault/cmd/common"
	"testing"
)

func TestRunGenerate_Default(t *testing.T) {
	common.SetupVault(t)
	genLength = 16
	genNoSymbols = false
	genSaveKey = ""
	genNoCopy = true
	defer func() {
		genLength = 24
		genNoCopy = false
	}()

	out := common.CaptureStdout(t, func() {
		if err := runGenerate(common.NewCmd(), nil); err != nil {
			t.Fatalf("runGenerate: %v", err)
		}
	})

	pw := strings.TrimSpace(out)
	if len(pw) != 16 {
		t.Errorf("expected generated password length 16, got %d (%q)", len(pw), pw)
	}
}

func TestRunGenerate_SaveToVault(t *testing.T) {
	common.SetupVault(t)
	genLength = 12
	genNoSymbols = true
	genSaveKey = "GEN_PASS"
	genNamespace = "default"
	genNoCopy = true
	defer func() {
		genLength = 24
		genSaveKey = ""
		genNoCopy = false
	}()

	out := common.CaptureStdout(t, func() {
		if err := runGenerate(common.NewCmd(), nil); err != nil {
			t.Fatalf("runGenerate: %v", err)
		}
	})

	if !strings.Contains(out, "Saved as GEN_PASS") {
		t.Errorf("expected 'Saved as GEN_PASS' in output: %q", out)
	}

	val, ok := common.ReadSecret(t, "default", "GEN_PASS")
	if !ok {
		t.Fatal("expected GEN_PASS to be saved in vault")
	}
	if len(val) != 12 {
		t.Errorf("expected saved password length 12, got %d (%q)", len(val), val)
	}
}

func TestRunGenerate_TooShortFails(t *testing.T) {
	common.SetupVault(t)
	genLength = 4
	defer func() { genLength = 24 }()

	err := runGenerate(common.NewCmd(), nil)
	if err == nil {
		t.Fatal("expected error for length < 8, got nil")
	}
}

func TestRunCopy_KeyNotFound(t *testing.T) {
	common.SetupVault(t)
	copyNamespace = "default"

	err := runCopy(common.NewCmd(), []string{"NONEXISTENT_KEY"})
	if err == nil {
		t.Fatal("expected error for non-existent key, got nil")
	}
}
