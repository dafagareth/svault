package audit

import (
	"crypto/sha256"
	"fmt"
	"os"
	"time"
)

type Op string

const (
	OpSet    Op = "SET"
	OpGet    Op = "GET"
	OpDelete Op = "DELETE"
	OpInit   Op = "INIT"
	OpUnlock Op = "UNLOCK"
	OpLock   Op = "LOCK"
	OpExport Op = "EXPORT"
	OpImport Op = "IMPORT"
	OpRotate Op = "ROTATE"
)

func MaskKey(key string) string {
	if key == "" || key == "-" {
		return "-"
	}
	h := sha256.Sum256([]byte(key))
	return fmt.Sprintf("id:%x", h[:4])
}

func Append(logPath string, op Op, key, namespace string) error {
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("open audit log: %w", err)
	}
	defer f.Close()
	ts := time.Now().Format("2006-01-02 15:04:05")
	maskedKey := MaskKey(key)
	if _, err := fmt.Fprintf(f, "%s  %-6s  %-20s  [%s]\n", ts, op, maskedKey, namespace); err != nil {
		return fmt.Errorf("write audit log: %w", err)
	}
	return nil
}
