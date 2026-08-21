package storage

import (
	"fmt"
	"os"
	"path/filepath"
)

const lockFileName = "vault.lock"

// WithVaultLock runs fn while holding an exclusive lock on the vault directory.
// Concurrent svault processes serialize on this lock, which prevents a
// read-modify-write race where two writers could clobber each other's changes.
//
// The lock is held on a dedicated vault.lock file, not on vault.enc itself, so
// the lock is never affected by the vault being rewritten or replaced.
func WithVaultLock(dir string, fn func() error) error {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create vault dir: %w", err)
	}
	lockPath := filepath.Join(dir, lockFileName)
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return fmt.Errorf("open lock file: %w", err)
	}
	defer f.Close()

	if err := lockFile(f); err != nil {
		return fmt.Errorf("acquire vault lock: %w", err)
	}
	defer func() { _ = unlockFile(f) }()

	return fn()
}
