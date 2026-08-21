package crypto

import (
	"bytes"
	"testing"
)

func TestEncryptDecryptRoundTrip(t *testing.T) {
	key := make([]byte, KeySize)
	for i := range key {
		key[i] = byte(i)
	}
	plaintext := []byte("hello, vault!")

	encrypted, err := Encrypt(key, plaintext)
	if err != nil {
		t.Fatal(err)
	}
	decrypted, err := Decrypt(key, encrypted)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(plaintext, decrypted) {
		t.Errorf("round-trip mismatch: got %q, want %q", decrypted, plaintext)
	}
}

func TestEncryptProducesUniqueCiphertexts(t *testing.T) {
	key := make([]byte, KeySize)
	plaintext := []byte("same input")

	c1, err := Encrypt(key, plaintext)
	if err != nil {
		t.Fatal(err)
	}
	c2, err := Encrypt(key, plaintext)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(c1, c2) {
		t.Error("two encryptions of the same plaintext produced identical ciphertexts")
	}
}

func TestDecryptWrongKeyFails(t *testing.T) {
	key := make([]byte, KeySize)
	wrongKey := make([]byte, KeySize)
	wrongKey[0] = 0xFF

	encrypted, err := Encrypt(key, []byte("secret"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Decrypt(wrongKey, encrypted); err == nil {
		t.Error("expected error decrypting with wrong key")
	}
}

func TestDecryptTamperedCiphertextFails(t *testing.T) {
	key := make([]byte, KeySize)

	encrypted, err := Encrypt(key, []byte("secret"))
	if err != nil {
		t.Fatal(err)
	}
	encrypted[len(encrypted)-1] ^= 0xFF
	if _, err := Decrypt(key, encrypted); err == nil {
		t.Error("expected error decrypting tampered ciphertext")
	}
}
