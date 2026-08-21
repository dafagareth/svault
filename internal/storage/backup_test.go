package storage

import (
	"path/filepath"
	"testing"

	"svault/internal/crypto"
)

func TestBackupRotation_KeepsAtMostRetention(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "vault.enc")
	password := []byte("password")

	if err := InitVault(path, password); err != nil {
		t.Fatal(err)
	}
	salt, _ := ReadSalt(path)
	key := crypto.DeriveKey(password, salt)
	vd, _ := ReadVault(path, key)

	// Write more times than the retention limit.
	for i := 0; i < backupRetention+4; i++ {
		if err := WriteVault(path, key, vd); err != nil {
			t.Fatal(err)
		}
	}

	matches, err := filepath.Glob(path + ".bak-*")
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) > backupRetention {
		t.Errorf("kept %d timestamped backups, want at most %d", len(matches), backupRetention)
	}
	if len(matches) == 0 {
		t.Error("expected at least one timestamped backup")
	}

	// The legacy single .bak should also exist.
	if _, err := ReadSalt(path + ".bak"); err != nil {
		t.Errorf("legacy .bak missing: %v", err)
	}
}
