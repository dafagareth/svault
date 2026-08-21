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

var sessionFile = defaultSessionPath()

func defaultSessionPath() string {
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		return filepath.Join(home, ".svault", ".session")
	}
	return filepath.Join(os.TempDir(), ".svault_session")
}

type session struct {
	Key    []byte    `json:"key"`
	Expiry time.Time `json:"expiry"`
}

func SessionPath() string {
	return sessionFile
}

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
	dir := filepath.Dir(sessionFile)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create session dir: %w", err)
	}
	tmpFile, err := os.CreateTemp(dir, ".session-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp session: %w", err)
	}
	tmpName := tmpFile.Name()
	defer os.Remove(tmpName)

	if err := tmpFile.Chmod(0600); err != nil {
		tmpFile.Close()
		return fmt.Errorf("chmod temp session: %w", err)
	}
	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		return fmt.Errorf("write temp session: %w", err)
	}
	if err := tmpFile.Sync(); err != nil {
		tmpFile.Close()
		return fmt.Errorf("sync temp session: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("close temp session: %w", err)
	}
	if err := os.Rename(tmpName, sessionFile); err != nil {
		return fmt.Errorf("atomic rename session: %w", err)
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
