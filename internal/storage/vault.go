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

package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"svault/internal/crypto"
)

type VaultData struct {
	Version    string                       `json:"version"`
	Namespaces map[string]map[string]string `json:"namespaces"`
}

func NewVaultData() *VaultData {
	return &VaultData{
		Version:    "1",
		Namespaces: map[string]map[string]string{"default": {}},
	}
}

// InitVault creates a new encrypted vault at path using password. Generates a fresh salt.
func InitVault(path string, password []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("create vault dir: %w", err)
	}
	salt, err := crypto.NewSalt()
	if err != nil {
		return err
	}
	key := crypto.DeriveKey(password, salt)
	return writeVaultFile(path, salt, key, NewVaultData())
}

// ReadSalt reads the 16-byte salt from the vault file header.
func ReadSalt(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open vault: %w", err)
	}
	defer f.Close()
	salt := make([]byte, crypto.SaltSize)
	if _, err := f.Read(salt); err != nil {
		return nil, fmt.Errorf("read salt: %w", err)
	}
	return salt, nil
}

// ReadVault reads and decrypts the vault file using the provided derived key.
func ReadVault(path string, key []byte) (*VaultData, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read vault: %w", err)
	}
	if len(raw) < crypto.SaltSize {
		return nil, fmt.Errorf("vault file corrupted")
	}
	plaintext, err := crypto.Decrypt(key, raw[crypto.SaltSize:])
	if err != nil {
		return nil, fmt.Errorf("decrypt vault: %w", err)
	}
	var vd VaultData
	if err := json.Unmarshal(plaintext, &vd); err != nil {
		return nil, fmt.Errorf("parse vault: %w", err)
	}
	return &vd, nil
}

// WriteVault encrypts and writes vault data, preserving the existing salt.
// Creates a .bak backup before every write.
func WriteVault(path string, key []byte, data *VaultData) error {
	salt, err := ReadSalt(path)
	if err != nil {
		return fmt.Errorf("read salt for write: %w", err)
	}
	if err := backupVault(path); err != nil {
		return fmt.Errorf("backup vault: %w", err)
	}
	return writeVaultFile(path, salt, key, data)
}

// RotateVault re-encrypts the vault with a new password and returns the new derived key.
func RotateVault(path string, currentKey, newPassword []byte) ([]byte, error) {
	vd, err := ReadVault(path, currentKey)
	if err != nil {
		return nil, fmt.Errorf("read vault for rotate: %w", err)
	}
	newSalt, err := crypto.NewSalt()
	if err != nil {
		return nil, err
	}
	newKey := crypto.DeriveKey(newPassword, newSalt)
	if err := backupVault(path); err != nil {
		return nil, fmt.Errorf("backup vault: %w", err)
	}
	if err := writeVaultFile(path, newSalt, newKey, vd); err != nil {
		return nil, err
	}
	return newKey, nil
}

func writeVaultFile(path string, salt, key []byte, data *VaultData) error {
	plaintext, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal vault: %w", err)
	}
	encrypted, err := crypto.Encrypt(key, plaintext)
	if err != nil {
		return fmt.Errorf("encrypt vault: %w", err)
	}
	blob := append(append([]byte(nil), salt...), encrypted...)
	if err := os.WriteFile(path, blob, 0600); err != nil {
		return fmt.Errorf("write vault: %w", err)
	}
	return nil
}

// backupRetention is the number of timestamped rollback backups kept.
const backupRetention = 5

// backupVault copies the current vault to a timestamped rollback file and
// prunes old ones so only the most recent backupRetention remain. Keeping
// several backups means one bad write cannot destroy a good earlier copy,
// unlike a single overwritten .bak file.
func backupVault(path string) error {
	src, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read vault for backup: %w", err)
	}

	// Keep the legacy single .bak for backward compatibility.
	if err := os.WriteFile(path+".bak", src, 0600); err != nil {
		return err
	}

	// Timestamped rollback copy (nanosecond precision avoids collisions on
	// rapid successive writes).
	stamp := time.Now().Format("20060102-150405.000000000")
	if err := os.WriteFile(path+".bak-"+stamp, src, 0600); err != nil {
		return err
	}

	return pruneBackups(path)
}

// pruneBackups removes the oldest timestamped backups beyond backupRetention.
func pruneBackups(path string) error {
	matches, err := filepath.Glob(path + ".bak-*")
	if err != nil {
		return nil // globbing failure is non-fatal for a backup prune
	}
	if len(matches) <= backupRetention {
		return nil
	}
	sort.Strings(matches) // timestamp format sorts chronologically
	for _, old := range matches[:len(matches)-backupRetention] {
		_ = os.Remove(old)
	}
	return nil
}
