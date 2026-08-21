package audit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAppendCreatesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "vault.log")

	if err := Append(path, OpSet, "MY_KEY", "default"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal("log file not created:", err)
	}
}

func TestAppendContents(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "vault.log")

	if err := Append(path, OpSet, "MY_KEY", "default"); err != nil {
		t.Fatal(err)
	}
	if err := Append(path, OpGet, "OTHER_KEY", "production"); err != nil {
		t.Fatal(err)
	}
	if err := Append(path, OpDelete, "OLD_KEY", "default"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	for _, want := range []string{"SET", MaskKey("MY_KEY"), "[default]", "GET", MaskKey("OTHER_KEY"), "[production]", "DELETE", MaskKey("OLD_KEY")} {
		if !strings.Contains(content, want) {
			t.Errorf("expected %q in audit log", want)
		}
	}
}

func TestAppendIsAppendOnly(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "vault.log")

	if err := Append(path, OpSet, "KEY1", "default"); err != nil {
		t.Fatal(err)
	}
	if err := Append(path, OpSet, "KEY2", "default"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 log lines, got %d", len(lines))
	}
}

func TestAppendNoSecretValues(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "vault.log")

	if err := Append(path, OpSet, "MY_SECRET", "default"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "supersecret") {
		t.Error("audit log must not contain secret values")
	}
}
