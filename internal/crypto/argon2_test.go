package crypto

import (
	"bytes"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	ArgonMemory = 8 * 1024 // lower memory for fast tests
	ArgonTime = 1
	os.Exit(m.Run())
}

func TestDeriveKeyDeterministic(t *testing.T) {
	password := []byte("mypassword")
	salt := make([]byte, SaltSize)

	k1 := DeriveKey(password, salt)
	k2 := DeriveKey(password, salt)

	if !bytes.Equal(k1, k2) {
		t.Error("same password+salt should produce same key")
	}
	if len(k1) != KeySize {
		t.Errorf("key size: got %d, want %d", len(k1), KeySize)
	}
}

func TestDeriveKeyDifferentSalt(t *testing.T) {
	password := []byte("mypassword")
	salt1 := make([]byte, SaltSize)
	salt2 := make([]byte, SaltSize)
	salt2[0] = 1

	k1 := DeriveKey(password, salt1)
	k2 := DeriveKey(password, salt2)

	if bytes.Equal(k1, k2) {
		t.Error("different salts should produce different keys")
	}
}

func TestDeriveKeyDifferentPassword(t *testing.T) {
	salt := make([]byte, SaltSize)

	k1 := DeriveKey([]byte("password1"), salt)
	k2 := DeriveKey([]byte("password2"), salt)

	if bytes.Equal(k1, k2) {
		t.Error("different passwords should produce different keys")
	}
}

func TestNewSaltLength(t *testing.T) {
	salt, err := NewSalt()
	if err != nil {
		t.Fatal(err)
	}
	if len(salt) != SaltSize {
		t.Errorf("salt size: got %d, want %d", len(salt), SaltSize)
	}
}

func TestNewSaltUnique(t *testing.T) {
	s1, err := NewSalt()
	if err != nil {
		t.Fatal(err)
	}
	s2, err := NewSalt()
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(s1, s2) {
		t.Error("two salts should not be identical")
	}
}
