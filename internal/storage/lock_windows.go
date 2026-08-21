package storage

import (
	"os"

	"golang.org/x/sys/windows"
)

// lockFile takes an exclusive lock on f using LockFileEx. The lock is released
// when the handle is closed or the process exits.
func lockFile(f *os.File) error {
	overlapped := new(windows.Overlapped)
	return windows.LockFileEx(
		windows.Handle(f.Fd()),
		windows.LOCKFILE_EXCLUSIVE_LOCK,
		0, 1, 0, overlapped,
	)
}

func unlockFile(f *os.File) error {
	overlapped := new(windows.Overlapped)
	return windows.UnlockFileEx(windows.Handle(f.Fd()), 0, 1, 0, overlapped)
}
