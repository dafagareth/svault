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
	"bytes"
	"encoding/json"
	"os"
	"testing"
	"time"
)

func init() {
	// Redirect session file to a temp location for all tests in this package.
	// TestMain (in vault_test.go) already handles crypto params.
	dir, err := os.MkdirTemp("", "svault-session-*")
	if err != nil {
		panic(err)
	}
	sessionFile = dir + "/.session"
}

func TestSaveLoadSession(t *testing.T) {
	defer os.Remove(sessionFile)

	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i + 1)
	}

	if err := SaveSession(key); err != nil {
		t.Fatal("save:", err)
	}
	loaded, err := LoadSession()
	if err != nil {
		t.Fatal("load:", err)
	}
	if !bytes.Equal(key, loaded) {
		t.Errorf("key mismatch: got %x, want %x", loaded, key)
	}
}

func TestLoadSessionNotExist(t *testing.T) {
	os.Remove(sessionFile)
	_, err := LoadSession()
	if err == nil {
		t.Error("expected error when no session file")
	}
}

func TestDeleteSession(t *testing.T) {
	key := make([]byte, 32)
	if err := SaveSession(key); err != nil {
		t.Fatal(err)
	}
	if err := DeleteSession(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(sessionFile); !os.IsNotExist(err) {
		t.Error("expected session file to be deleted")
	}
}

func TestDeleteSessionIdempotent(t *testing.T) {
	os.Remove(sessionFile)
	if err := DeleteSession(); err != nil {
		t.Error("delete of non-existent session should not error:", err)
	}
}

func TestLoadSessionExpired(t *testing.T) {
	// Write a session that expired in the past.
	s := session{
		Key:    make([]byte, 32),
		Expiry: time.Now().Add(-1 * time.Minute),
	}
	data, _ := json.Marshal(s)
	os.WriteFile(sessionFile, data, 0600)

	_, err := LoadSession()
	if err == nil {
		t.Error("expected error for expired session")
	}
	if _, statErr := os.Stat(sessionFile); !os.IsNotExist(statErr) {
		t.Error("expired session file should be deleted")
	}
}

func TestSessionPath(t *testing.T) {
	if SessionPath() == "" {
		t.Error("SessionPath should not be empty")
	}
}
