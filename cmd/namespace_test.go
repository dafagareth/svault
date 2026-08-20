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
	"strings"
	"testing"
)

func TestNsList(t *testing.T) {
	setupVault(t)
	seed(t, "default", "A", "1")
	seed(t, "work", "B", "2")
	seed(t, "work", "C", "3")

	out := captureStdout(t, func() {
		if err := runNsList(newCmd(), nil); err != nil {
			t.Fatalf("runNsList: %v", err)
		}
	})
	if !strings.Contains(out, "default") || !strings.Contains(out, "work") {
		t.Errorf("expected both namespaces in output: %q", out)
	}
	if !strings.Contains(out, "2 keys") {
		t.Errorf("expected 'work' to show 2 keys: %q", out)
	}
	// Active namespace marker.
	if !strings.Contains(out, "* default") {
		t.Errorf("expected active marker on default: %q", out)
	}
}

func TestNsDelete(t *testing.T) {
	setupVault(t)
	seed(t, "temp", "X", "1")

	if err := runNsDelete(newCmd(), []string{"temp"}); err != nil {
		t.Fatalf("runNsDelete: %v", err)
	}
	if _, ok := readSecret(t, "temp", "X"); ok {
		t.Error("namespace not deleted")
	}
}

func TestNsDelete_DefaultProtected(t *testing.T) {
	setupVault(t)
	if err := runNsDelete(newCmd(), []string{"default"}); err == nil {
		t.Error("expected error deleting default namespace")
	}
}

func TestNsDelete_NotFound(t *testing.T) {
	setupVault(t)
	if err := runNsDelete(newCmd(), []string{"ghost"}); err == nil {
		t.Error("expected error for missing namespace")
	}
}

func TestNsRename(t *testing.T) {
	setupVault(t)
	seed(t, "old", "K", "v")

	if err := runNsRename(newCmd(), []string{"old", "new"}); err != nil {
		t.Fatalf("runNsRename: %v", err)
	}
	got, ok := readSecret(t, "new", "K")
	if !ok || got != "v" {
		t.Errorf("new/K = %q (ok=%v), want v", got, ok)
	}
	if _, ok := readSecret(t, "old", "K"); ok {
		t.Error("old namespace still present")
	}
}

func TestNsRename_DefaultProtected(t *testing.T) {
	setupVault(t)
	if err := runNsRename(newCmd(), []string{"default", "x"}); err == nil {
		t.Error("expected error renaming default namespace")
	}
}

func TestNsRename_TargetExists(t *testing.T) {
	setupVault(t)
	seed(t, "a", "1", "x")
	seed(t, "b", "1", "y")
	if err := runNsRename(newCmd(), []string{"a", "b"}); err == nil {
		t.Error("expected error when target namespace exists")
	}
}

func TestRunSearch(t *testing.T) {
	setupVault(t)
	seed(t, "default", "DB_URL", "1")
	seed(t, "default", "DB_PASSWORD", "2")
	seed(t, "default", "API_KEY", "3")

	out := captureStdout(t, func() {
		if err := runSearch(newCmd(), []string{"db"}); err != nil {
			t.Fatalf("runSearch: %v", err)
		}
	})
	if !strings.Contains(out, "DB_URL") || !strings.Contains(out, "DB_PASSWORD") {
		t.Errorf("expected DB_* keys: %q", out)
	}
	if strings.Contains(out, "API_KEY") {
		t.Errorf("API_KEY should not match 'db': %q", out)
	}
}

func TestRunSearch_NoMatch(t *testing.T) {
	setupVault(t)
	seed(t, "default", "FOO", "1")

	out := captureStdout(t, func() {
		if err := runSearch(newCmd(), []string{"zzz"}); err != nil {
			t.Fatalf("runSearch: %v", err)
		}
	})
	if !strings.Contains(out, "No keys match") {
		t.Errorf("expected no-match message: %q", out)
	}
}

func TestRunEnv(t *testing.T) {
	setupVault(t)
	seed(t, "default", "HOST", "localhost")
	seed(t, "default", "PORT", "8080")

	out := captureStdout(t, func() {
		if err := runEnv(newCmd(), nil); err != nil {
			t.Fatalf("runEnv: %v", err)
		}
	})
	if !strings.Contains(out, "export HOST='localhost'") {
		t.Errorf("missing HOST export: %q", out)
	}
	if !strings.Contains(out, "export PORT='8080'") {
		t.Errorf("missing PORT export: %q", out)
	}
}

func TestShellQuote(t *testing.T) {
	cases := map[string]string{
		"simple":      "'simple'",
		"with space":  "'with space'",
		"it's quoted": `'it'\''s quoted'`,
	}
	for in, want := range cases {
		if got := shellQuote(in); got != want {
			t.Errorf("shellQuote(%q) = %q, want %q", in, got, want)
		}
	}
}
