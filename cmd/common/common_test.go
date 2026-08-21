package common

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestWipe(t *testing.T) {
	b := []byte("secret_password_123")
	Wipe(b)
	for i, v := range b {
		if v != 0 {
			t.Fatalf("byte at index %d was not wiped: %d", i, v)
		}
	}
}

func TestValidateKey(t *testing.T) {
	validKeys := []string{"DB_PASS", "api_key", "_SECRET", "VAR123", "a"}
	for _, k := range validKeys {
		if err := ValidateKey(k); err != nil {
			t.Errorf("expected valid key %q, got error: %v", k, err)
		}
	}

	invalidKeys := []string{"", "123VAR", "DB-PASS", "KEY@NAME", "MY KEY"}
	for _, k := range invalidKeys {
		if err := ValidateKey(k); err == nil {
			t.Errorf("expected error for invalid key %q, got nil", k)
		}
	}
}

func TestVaultDirAndPath(t *testing.T) {
	dir, err := VaultDir()
	if err != nil {
		t.Fatalf("VaultDir() error: %v", err)
	}
	if dir == "" {
		t.Fatal("VaultDir() returned empty string")
	}

	vpath, err := VaultPath()
	if err != nil {
		t.Fatalf("VaultPath() error: %v", err)
	}
	if filepath.Base(vpath) != "vault.enc" {
		t.Fatalf("expected vault.enc, got %s", filepath.Base(vpath))
	}

	lpath, err := LogPath()
	if err != nil {
		t.Fatalf("LogPath() error: %v", err)
	}
	if filepath.Base(lpath) != "vault.log" {
		t.Fatalf("expected vault.log, got %s", filepath.Base(lpath))
	}
}

func TestGitRepoName(t *testing.T) {
	name := GitRepoName()
	if name != "" && !strings.Contains(name, "svault") {
		t.Logf("GitRepoName() returned: %s", name)
	}
}

func TestNamespaceSource(t *testing.T) {
	t.Setenv("SVAULT_NS", "test-env-ns")
	ns, src := NamespaceSource()
	if ns != "test-env-ns" || src != "env" {
		t.Errorf("NamespaceSource() = (%q, %q), want ('test-env-ns', 'env')", ns, src)
	}
}
