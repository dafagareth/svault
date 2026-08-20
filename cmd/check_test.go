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

// writeEnvFile writes content to a temp .env file and returns its path.
func writeEnvFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), ".env.example")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestRunCheck_AllPresent(t *testing.T) {
	setupVault(t)
	seed(t, "default", "DB_HOST", "h")
	file := writeEnvFile(t, "DB_HOST=\n")

	out := captureStdout(t, func() {
		if err := runCheck(newCmd(), []string{file}); err != nil {
			t.Fatalf("runCheck: %v", err)
		}
	})
	if !strings.Contains(out, "OK    DB_HOST") {
		t.Errorf("expected OK for DB_HOST: %q", out)
	}
	if strings.Contains(out, "missing") {
		t.Errorf("did not expect missing: %q", out)
	}
}

func TestRunCheck_Strict(t *testing.T) {
	setupVault(t)
	defer func() { checkStrict = false }()
	seed(t, "default", "DB_HOST", "h")
	file := writeEnvFile(t, "DB_HOST=\nSTRIPE_KEY=\n")
	checkStrict = true

	err := captureStdoutErr(t, func() error {
		return runCheck(newCmd(), []string{file})
	})
	if err == nil {
		t.Error("expected non-zero error with --strict when a key is missing")
	}
}

func TestRunCheck_Extra(t *testing.T) {
	setupVault(t)
	defer func() { checkExtra = false }()
	seed(t, "default", "DB_HOST", "h")
	seed(t, "default", "LEGACY_TOKEN", "x") // present in vault, not in file
	file := writeEnvFile(t, "DB_HOST=\n")
	checkExtra = true

	out := captureStdout(t, func() {
		if err := runCheck(newCmd(), []string{file}); err != nil {
			t.Fatalf("runCheck: %v", err)
		}
	})
	if !strings.Contains(out, "EXTRA LEGACY_TOKEN") {
		t.Errorf("expected EXTRA LEGACY_TOKEN: %q", out)
	}
}

// captureStdoutErr is like captureStdout but also returns the error from fn,
// for commands whose exit code (not just output) matters.
func captureStdoutErr(t *testing.T, fn func() error) error {
	t.Helper()
	var err error
	captureStdout(t, func() { err = fn() })
	return err
}
