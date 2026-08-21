package storage

import (
	"sync"
	"sync/atomic"
	"testing"
)

func TestWithVaultLock_Serializes(t *testing.T) {
	dir := t.TempDir()

	var inside atomic.Int32
	var maxConcurrent atomic.Int32
	var wg sync.WaitGroup

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = WithVaultLock(dir, func() error {
				n := inside.Add(1)
				// Record the highest concurrency observed inside the lock.
				for {
					m := maxConcurrent.Load()
					if n <= m || maxConcurrent.CompareAndSwap(m, n) {
						break
					}
				}
				// Give other goroutines a chance to violate the lock if broken.
				for j := 0; j < 1000; j++ {
					_ = j
				}
				inside.Add(-1)
				return nil
			})
		}()
	}
	wg.Wait()

	if maxConcurrent.Load() != 1 {
		t.Errorf("lock allowed %d concurrent holders, want 1", maxConcurrent.Load())
	}
}

func TestWithVaultLock_ReturnsFnError(t *testing.T) {
	dir := t.TempDir()
	sentinel := errTest
	err := WithVaultLock(dir, func() error { return sentinel })
	if err != sentinel {
		t.Errorf("got %v, want sentinel error", err)
	}
}

var errTest = &testError{"boom"}

type testError struct{ msg string }

func (e *testError) Error() string { return e.msg }
