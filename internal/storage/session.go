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
	"strconv"
	"time"
)

const defaultTTLMinutes = 30

// sessionFile is the path to the session token file. Overridable in tests.
var sessionFile = filepath.Join(os.TempDir(), ".svault_session")

type session struct {
	Key    []byte    `json:"key"`
	Expiry time.Time `json:"expiry"`
}

func SessionPath() string {
	return sessionFile
}

// SetSessionFile overrides the session file path. Intended for tests so they
// can isolate session state from the real ~/.svault session.
func SetSessionFile(path string) {
	sessionFile = path
}

func SaveSession(key []byte) error {
	s := session{
		Key:    key,
		Expiry: time.Now().Add(sessionTTL()),
	}
	data, err := json.Marshal(s)
	if err != nil {
		return fmt.Errorf("marshal session: %w", err)
	}
	if err := os.WriteFile(sessionFile, data, 0600); err != nil {
		return fmt.Errorf("write session: %w", err)
	}
	return nil
}

func LoadSession() ([]byte, error) {
	data, err := os.ReadFile(sessionFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("vault is locked, run 'svault unlock'")
		}
		return nil, fmt.Errorf("read session: %w", err)
	}
	var s session
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parse session: %w", err)
	}
	if time.Now().After(s.Expiry) {
		_ = DeleteSession()
		return nil, fmt.Errorf("session expired, run 'svault unlock'")
	}
	return s.Key, nil
}

// SessionRemaining returns the time left on the active session and whether one
// is active. Returns (0, false) if the vault is locked or the session expired.
func SessionRemaining() (time.Duration, bool) {
	data, err := os.ReadFile(sessionFile)
	if err != nil {
		return 0, false
	}
	var s session
	if err := json.Unmarshal(data, &s); err != nil {
		return 0, false
	}
	remaining := time.Until(s.Expiry)
	if remaining <= 0 {
		return 0, false
	}
	return remaining, true
}

func DeleteSession() error {
	err := os.Remove(sessionFile)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

func sessionTTL() time.Duration {
	if v := os.Getenv("SVAULT_SESSION_TTL"); v != "" {
		if mins, err := strconv.Atoi(v); err == nil && mins > 0 {
			return time.Duration(mins) * time.Minute
		}
	}
	return defaultTTLMinutes * time.Minute
}
