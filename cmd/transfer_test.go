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
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunSet_Stdin(t *testing.T) {
	setupVault(t)
	setQuiet = true
	setStdin = true
	setInput = strings.NewReader("secret-from-pipe\n")
	defer func() {
		setQuiet = false
		setStdin = false
		setInput = os.Stdin
	}()

	if err := runSet(newCmd(), []string{"PIPED"}); err != nil {
		t.Fatalf("runSet --stdin: %v", err)
	}
	got, ok := readSecret(t, "default", "PIPED")
	if !ok || got != "secret-from-pipe" {
		t.Errorf("got %q (ok=%v), want secret-from-pipe", got, ok)
	}
}

func TestRunSet_StdinRejectsEqualsForm(t *testing.T) {
	setupVault(t)
	setStdin = true
	setInput = strings.NewReader("x")
	defer func() {
		setStdin = false
		setInput = os.Stdin
	}()

	if err := runSet(newCmd(), []string{"KEY=VALUE"}); err == nil {
		t.Error("expected error when passing KEY=VALUE with --stdin")
	}
}

func TestRunExport(t *testing.T) {
	setupVault(t)
	seed(t, "default", "B_KEY", "2")
	seed(t, "default", "A_KEY", "1")

	out := captureStdout(t, func() {
		if err := runExport(newCmd(), nil); err != nil {
			t.Fatalf("runExport: %v", err)
		}
	})
	// Keys are sorted, so A_KEY comes before B_KEY.
	ai := strings.Index(out, "A_KEY=1")
	bi := strings.Index(out, "B_KEY=2")
	if ai < 0 || bi < 0 {
		t.Fatalf("missing keys in export: %q", out)
	}
	if ai > bi {
		t.Errorf("export not sorted: %q", out)
	}
}

func TestRunImport(t *testing.T) {
	setupVault(t)

	dir := t.TempDir()
	envFile := filepath.Join(dir, ".env")
	content := "# a comment\nDB_URL=postgres://localhost\nAPI_KEY=\"quoted-value\"\n\nEMPTY_LINE_ABOVE=ok\n"
	if err := os.WriteFile(envFile, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	out := captureStdout(t, func() {
		if err := runImport(newCmd(), []string{envFile}); err != nil {
			t.Fatalf("runImport: %v", err)
		}
	})
	if !strings.Contains(out, "Imported 3 keys") {
		t.Errorf("unexpected import summary: %q", out)
	}
	if v, _ := readSecret(t, "default", "DB_URL"); v != "postgres://localhost" {
		t.Errorf("DB_URL = %q", v)
	}
	// Quotes around the value should be stripped.
	if v, _ := readSecret(t, "default", "API_KEY"); v != "quoted-value" {
		t.Errorf("API_KEY = %q, want quoted-value", v)
	}
}

func TestRunImport_SkipExisting(t *testing.T) {
	setupVault(t)
	seed(t, "default", "EXISTING", "original")

	dir := t.TempDir()
	envFile := filepath.Join(dir, ".env")
	if err := os.WriteFile(envFile, []byte("EXISTING=changed\nNEW=value\n"), 0600); err != nil {
		t.Fatal(err)
	}

	importOverwrite = false
	defer func() { importOverwrite = true }()

	out := captureStdout(t, func() {
		if err := runImport(newCmd(), []string{envFile}); err != nil {
			t.Fatalf("runImport: %v", err)
		}
	})
	if !strings.Contains(out, "skipped") {
		t.Errorf("expected skipped note: %q", out)
	}
	if v, _ := readSecret(t, "default", "EXISTING"); v != "original" {
		t.Errorf("EXISTING overwritten: got %q, want original", v)
	}
	if v, _ := readSecret(t, "default", "NEW"); v != "value" {
		t.Errorf("NEW = %q, want value", v)
	}
}

func TestRunImport_EmptyFile(t *testing.T) {
	setupVault(t)
	dir := t.TempDir()
	envFile := filepath.Join(dir, "empty.env")
	if err := os.WriteFile(envFile, []byte("# only comments\n\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := runImport(newCmd(), []string{envFile}); err == nil {
		t.Error("expected error for file with no valid keys")
	}
}

func TestRunBackupAndRestore(t *testing.T) {
	setupVault(t)
	seed(t, "default", "KEEP", "value")

	dir := t.TempDir()
	backupFile := filepath.Join(dir, "vault.bak")

	out := captureStdout(t, func() {
		if err := runBackup(newCmd(), []string{backupFile}); err != nil {
			t.Fatalf("runBackup: %v", err)
		}
	})
	if !strings.Contains(out, "backed up") {
		t.Errorf("unexpected backup output: %q", out)
	}
	if _, err := os.Stat(backupFile); err != nil {
		t.Fatalf("backup file not created: %v", err)
	}

	// Mutate the live vault, then restore from backup.
	seed(t, "default", "TEMP", "should-disappear")

	_ = captureStdout(t, func() {
		if err := runRestore(newCmd(), []string{backupFile}); err != nil {
			t.Fatalf("runRestore: %v", err)
		}
	})

	// Restore clears the session, so unlock again before reading.
	reunlock(t)
	if _, ok := readSecret(t, "default", "TEMP"); ok {
		t.Error("TEMP should be gone after restore")
	}
	if v, _ := readSecret(t, "default", "KEEP"); v != "value" {
		t.Errorf("KEEP = %q after restore, want value", v)
	}
}

func TestRunBackup_AutoName(t *testing.T) {
	setupVault(t)

	out := captureStdout(t, func() {
		if err := runBackup(newCmd(), nil); err != nil {
			t.Fatalf("runBackup auto: %v", err)
		}
	})
	if !strings.Contains(out, "vault.enc.backup-") {
		t.Errorf("expected timestamped backup name: %q", out)
	}
}
