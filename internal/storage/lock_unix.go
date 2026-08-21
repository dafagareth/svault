package storage

import (
	"os"
	"syscall"
)

// lockFile takes an exclusive advisory lock on f, blocking until it is held.
// The lock is released automatically when the file descriptor is closed or the
// process exits, so a crash never leaves a stale lock behind.
func lockFile(f *os.File) error {
	return syscall.Flock(int(f.Fd()), syscall.LOCK_EX)
}

func unlockFile(f *os.File) error {
	return syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
}
