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

package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"svault/internal/crypto"
	"svault/internal/storage"
)

func TestMain(m *testing.M) {
	// Speed up Argon2id derivation for tests.
	crypto.ArgonMemory = 8 * 1024
	crypto.ArgonTime = 1
	os.Exit(m.Run())
}

// setupVault creates an isolated vault + unlocked session in a temp HOME and
// returns nothing; all package path helpers will resolve into the temp dir.
// It registers cleanup automatically via t.TempDir and t.Setenv.
func setupVault(t *testing.T) {
	t.Helper()

	home := t.TempDir()
	t.Setenv("HOME", home)
	// Force namespace detection to be deterministic ("default") instead of
	// picking up the surrounding git repo or user config.
	t.Setenv("SVAULT_NS", "default")

	// Isolate the session file from the real ~/.svault session.
	storage.SetSessionFile(filepath.Join(home, ".session"))

	vpath := filepath.Join(home, ".svault", "vault.enc")
	password := []byte("testpassword")
	if err := storage.InitVault(vpath, password); err != nil {
		t.Fatalf("init vault: %v", err)
	}
	salt, err := storage.ReadSalt(vpath)
	if err != nil {
		t.Fatalf("read salt: %v", err)
	}
	key := crypto.DeriveKey(password, salt)
	if err := storage.SaveSession(key); err != nil {
		t.Fatalf("save session: %v", err)
	}
}

// reunlock re-derives the key from the test password and saves a fresh session.
// Used after operations like restore that intentionally clear the session.
func reunlock(t *testing.T) {
	t.Helper()
	vpath, err := vaultPath()
	if err != nil {
		t.Fatal(err)
	}
	salt, err := storage.ReadSalt(vpath)
	if err != nil {
		t.Fatal(err)
	}
	key := crypto.DeriveKey([]byte("testpassword"), salt)
	if err := storage.SaveSession(key); err != nil {
		t.Fatal(err)
	}
}

// seed writes key=value into the given namespace of the test vault.
func seed(t *testing.T, ns, key, value string) {
	t.Helper()
	vpath, err := vaultPath()
	if err != nil {
		t.Fatal(err)
	}
	encKey, err := storage.LoadSession()
	if err != nil {
		t.Fatal(err)
	}
	vd, err := storage.ReadVault(vpath, encKey)
	if err != nil {
		t.Fatal(err)
	}
	if vd.Namespaces[ns] == nil {
		vd.Namespaces[ns] = map[string]string{}
	}
	vd.Namespaces[ns][key] = value
	if err := storage.WriteVault(vpath, encKey, vd); err != nil {
		t.Fatal(err)
	}
}

// readSecret returns the stored value and whether it exists.
func readSecret(t *testing.T, ns, key string) (string, bool) {
	t.Helper()
	vpath, err := vaultPath()
	if err != nil {
		t.Fatal(err)
	}
	encKey, err := storage.LoadSession()
	if err != nil {
		t.Fatal(err)
	}
	vd, err := storage.ReadVault(vpath, encKey)
	if err != nil {
		t.Fatal(err)
	}
	v, ok := vd.Namespaces[ns][key]
	return v, ok
}

// captureStdout runs fn and returns everything it printed to os.Stdout.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	done := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		done <- buf.String()
	}()
	fn()
	_ = w.Close()
	os.Stdout = old
	return <-done
}

// newCmd returns a bare cobra command for invoking run* functions in tests.
func newCmd() *cobra.Command {
	return &cobra.Command{}
}
