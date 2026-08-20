// Copyright 2026 Dafa
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

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
