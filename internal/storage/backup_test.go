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
	"path/filepath"
	"testing"

	"svault/internal/crypto"
)

func TestBackupRotation_KeepsAtMostRetention(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "vault.enc")
	password := []byte("password")

	if err := InitVault(path, password); err != nil {
		t.Fatal(err)
	}
	salt, _ := ReadSalt(path)
	key := crypto.DeriveKey(password, salt)
	vd, _ := ReadVault(path, key)

	// Write more times than the retention limit.
	for i := 0; i < backupRetention+4; i++ {
		if err := WriteVault(path, key, vd); err != nil {
			t.Fatal(err)
		}
	}

	matches, err := filepath.Glob(path + ".bak-*")
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) > backupRetention {
		t.Errorf("kept %d timestamped backups, want at most %d", len(matches), backupRetention)
	}
	if len(matches) == 0 {
		t.Error("expected at least one timestamped backup")
	}

	// The legacy single .bak should also exist.
	if _, err := ReadSalt(path + ".bak"); err != nil {
		t.Errorf("legacy .bak missing: %v", err)
	}
}
