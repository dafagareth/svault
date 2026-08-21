package storage

import (
	"os"
	"path/filepath"
	"testing"

	"svault/internal/crypto"
)

func TestMain(m *testing.M) {
	crypto.ArgonMemory = 8 * 1024
	crypto.ArgonTime = 1
	os.Exit(m.Run())
}

func TestInitVault(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "vault.enc")

	if err := InitVault(path, []byte("password")); err != nil {
		t.Fatal("init vault:", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal("vault file not created:", err)
	}
}

func TestReadSalt(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "vault.enc")

	if err := InitVault(path, []byte("password")); err != nil {
		t.Fatal(err)
	}
	salt, err := ReadSalt(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(salt) != crypto.SaltSize {
		t.Errorf("salt size: got %d, want %d", len(salt), crypto.SaltSize)
	}
}

func TestReadWriteVaultRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "vault.enc")
	password := []byte("testpassword")

	if err := InitVault(path, password); err != nil {
		t.Fatal("init:", err)
	}

	salt, err := ReadSalt(path)
	if err != nil {
		t.Fatal("read salt:", err)
	}
	key := crypto.DeriveKey(password, salt)

	vd, err := ReadVault(path, key)
	if err != nil {
		t.Fatal("read:", err)
	}
	if vd.Version != "1" {
		t.Errorf("version: got %q, want %q", vd.Version, "1")
	}
	if _, ok := vd.Namespaces["default"]; !ok {
		t.Error("expected default namespace")
	}

	vd.Namespaces["default"]["SECRET_KEY"] = "hunter2"
	if err := WriteVault(path, key, vd); err != nil {
		t.Fatal("write:", err)
	}

	vd2, err := ReadVault(path, key)
	if err != nil {
		t.Fatal("read after write:", err)
	}
	if vd2.Namespaces["default"]["SECRET_KEY"] != "hunter2" {
		t.Errorf("secret: got %q, want %q", vd2.Namespaces["default"]["SECRET_KEY"], "hunter2")
	}
}

func TestWriteVaultCreatesBackup(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "vault.enc")
	password := []byte("password")

	if err := InitVault(path, password); err != nil {
		t.Fatal(err)
	}
	salt, _ := ReadSalt(path)
	key := crypto.DeriveKey(password, salt)
	vd, _ := ReadVault(path, key)

	if err := WriteVault(path, key, vd); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path + ".bak"); err != nil {
		t.Error("expected backup file:", err)
	}
}

func TestRotateVault(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "vault.enc")
	oldPassword := []byte("oldpassword")
	newPassword := []byte("newpassword")

	if err := InitVault(path, oldPassword); err != nil {
		t.Fatal("init:", err)
	}

	// Simpan secret sebelum rotate
	salt, _ := ReadSalt(path)
	oldKey := crypto.DeriveKey(oldPassword, salt)
	vd, _ := ReadVault(path, oldKey)
	vd.Namespaces["default"]["SECRET"] = "value123"
	_ = WriteVault(path, oldKey, vd)

	// Rotate password
	newKey, err := RotateVault(path, oldKey, newPassword)
	if err != nil {
		t.Fatal("rotate:", err)
	}

	// Key lama tidak boleh bisa buka vault
	if _, err := ReadVault(path, oldKey); err == nil {
		t.Error("key lama masih bisa decrypt setelah rotate")
	}

	// Key baru harus bisa buka dan data tetap ada
	vd2, err := ReadVault(path, newKey)
	if err != nil {
		t.Fatal("baca dengan key baru:", err)
	}
	if vd2.Namespaces["default"]["SECRET"] != "value123" {
		t.Errorf("data hilang setelah rotate: got %q", vd2.Namespaces["default"]["SECRET"])
	}
}

func TestReadVaultWrongKeyFails(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "vault.enc")

	if err := InitVault(path, []byte("correct")); err != nil {
		t.Fatal(err)
	}
	wrongKey := make([]byte, crypto.KeySize)
	if _, err := ReadVault(path, wrongKey); err == nil {
		t.Error("expected error with wrong key")
	}
}
