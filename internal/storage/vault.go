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

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create vault dir: %w", err)
	}
	tmpFile, err := os.CreateTemp(dir, ".vault-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp vault: %w", err)
	}
	tmpName := tmpFile.Name()
	defer os.Remove(tmpName)

	if err := tmpFile.Chmod(0600); err != nil {
		tmpFile.Close()
		return fmt.Errorf("chmod temp vault: %w", err)
	}
	if _, err := tmpFile.Write(blob); err != nil {
		tmpFile.Close()
		return fmt.Errorf("write temp vault: %w", err)
	}
	if err := tmpFile.Sync(); err != nil {
		tmpFile.Close()
		return fmt.Errorf("sync temp vault: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("close temp vault: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("atomic rename vault: %w", err)
	}
	return nil
}

const backupRetention = 5

func backupVault(path string) error {
	src, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read vault for backup: %w", err)
	}

	if err := os.WriteFile(path+".bak", src, 0600); err != nil {
		return err
	}

	stamp := time.Now().Format("20060102-150405.000000000")
	if err := os.WriteFile(path+".bak-"+stamp, src, 0600); err != nil {
		return err
	}

	return pruneBackups(path)
}

func pruneBackups(path string) error {
	matches, err := filepath.Glob(path + ".bak-*")
	if err != nil {
		return nil
	}
	if len(matches) <= backupRetention {
		return nil
	}
	sort.Strings(matches)
	for _, old := range matches[:len(matches)-backupRetention] {
		_ = os.Remove(old)
	}
	return nil
}
