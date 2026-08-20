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

func TestRunSet_KeyValueForm(t *testing.T) {
	setupVault(t)
	setQuiet = true
	defer func() { setQuiet = false }()

	if err := runSet(newCmd(), []string{"DB_URL", "postgres://localhost"}); err != nil {
		t.Fatalf("runSet: %v", err)
	}
	got, ok := readSecret(t, "default", "DB_URL")
	if !ok || got != "postgres://localhost" {
		t.Errorf("got %q (ok=%v), want postgres://localhost", got, ok)
	}
}

func TestRunSet_EqualsForm(t *testing.T) {
	setupVault(t)
	setQuiet = true
	defer func() { setQuiet = false }()

	if err := runSet(newCmd(), []string{"API_KEY=sk-123"}); err != nil {
		t.Fatalf("runSet: %v", err)
	}
	got, ok := readSecret(t, "default", "API_KEY")
	if !ok || got != "sk-123" {
		t.Errorf("got %q (ok=%v), want sk-123", got, ok)
	}
}

func TestRunSet_EqualsFormKeepsLaterEquals(t *testing.T) {
	setupVault(t)
	setQuiet = true
	defer func() { setQuiet = false }()

	// Value contains '=' (e.g. a URL with query params) — only first '=' splits.
	if err := runSet(newCmd(), []string{"URL=https://x.com?a=1&b=2"}); err != nil {
		t.Fatalf("runSet: %v", err)
	}
	got, _ := readSecret(t, "default", "URL")
	if got != "https://x.com?a=1&b=2" {
		t.Errorf("got %q", got)
	}
}

func TestRunSet_InvalidEqualsForm(t *testing.T) {
	setupVault(t)
	if err := runSet(newCmd(), []string{"=novalue"}); err == nil {
		t.Error("expected error for '=novalue'")
	}
	if err := runSet(newCmd(), []string{"nokeyequals"}); err == nil {
		t.Error("expected error for missing '='")
	}
}

func TestRunGet(t *testing.T) {
	setupVault(t)
	seed(t, "default", "TOKEN", "abc123")

	out := captureStdout(t, func() {
		if err := runGet(newCmd(), []string{"TOKEN"}); err != nil {
			t.Fatalf("runGet: %v", err)
		}
	})
	if strings.TrimSpace(out) != "abc123" {
		t.Errorf("got %q, want abc123", strings.TrimSpace(out))
	}
}

func TestRunGet_NotFound(t *testing.T) {
	setupVault(t)
	if err := runGet(newCmd(), []string{"MISSING"}); err == nil {
		t.Error("expected error for missing key")
	}
}

func TestRunGet_MultipleKeys(t *testing.T) {
	setupVault(t)
	seed(t, "default", "HOST", "localhost")
	seed(t, "default", "PORT", "5432")

	out := captureStdout(t, func() {
		if err := runGet(newCmd(), []string{"HOST", "PORT"}); err != nil {
			t.Fatalf("runGet: %v", err)
		}
	})
	// With more than one key, output is KEY=VALUE per line.
	if !strings.Contains(out, "HOST=localhost") || !strings.Contains(out, "PORT=5432") {
		t.Errorf("expected KEY=VALUE lines: %q", out)
	}
}

func TestRunGet_ClipRejectsMultipleKeys(t *testing.T) {
	setupVault(t)
	seed(t, "default", "HOST", "localhost")
	seed(t, "default", "PORT", "5432")
	getClip = true
	defer func() { getClip = false }()

	// --clip with several keys must error before touching the clipboard.
	if err := runGet(newCmd(), []string{"HOST", "PORT"}); err == nil {
		t.Error("expected error for --clip with multiple keys")
	}
}

func TestRunGet_Clip(t *testing.T) {
	setupVault(t)
	seed(t, "default", "TOKEN", "secret-value")
	getClip = true
	getTimeout = 0 // no auto-clear, avoid spawning a background process in tests
	var copied string
	orig := copyToClipboard
	copyToClipboard = func(s string) error { copied = s; return nil }
	defer func() {
		getClip = false
		getTimeout = 20
		copyToClipboard = orig
	}()

	out := captureStdout(t, func() {
		if err := runGet(newCmd(), []string{"TOKEN"}); err != nil {
			t.Fatalf("runGet: %v", err)
		}
	})
	if copied != "secret-value" {
		t.Errorf("clipboard got %q, want secret-value", copied)
	}
	// The value itself must never be printed to stdout in clip mode.
	if strings.Contains(out, "secret-value") {
		t.Errorf("value leaked to stdout: %q", out)
	}
}

func TestRunSet_Clip(t *testing.T) {
	setupVault(t)
	setQuiet = true
	setClip = true
	orig := readClipboard
	readClipboard = func() (string, error) { return "from-clipboard", nil }
	defer func() {
		setQuiet = false
		setClip = false
		readClipboard = orig
	}()

	if err := runSet(newCmd(), []string{"API_KEY"}); err != nil {
		t.Fatalf("runSet: %v", err)
	}
	got, ok := readSecret(t, "default", "API_KEY")
	if !ok || got != "from-clipboard" {
		t.Errorf("got %q (ok=%v), want from-clipboard", got, ok)
	}
}

func TestRunSet_ClipRejectsEqualsForm(t *testing.T) {
	setupVault(t)
	setClip = true
	orig := readClipboard
	readClipboard = func() (string, error) { return "x", nil }
	defer func() {
		setClip = false
		readClipboard = orig
	}()

	if err := runSet(newCmd(), []string{"KEY=VALUE"}); err == nil {
		t.Error("expected error: --clip must receive only the KEY")
	}
}

func TestRunSet_StdinClipMutuallyExclusive(t *testing.T) {
	setupVault(t)
	setStdin = true
	setClip = true
	defer func() {
		setStdin = false
		setClip = false
	}()

	if err := runSet(newCmd(), []string{"KEY"}); err == nil {
		t.Error("expected error when both --stdin and --clip are set")
	}
}

func TestRunDelete(t *testing.T) {
	setupVault(t)
	seed(t, "default", "TEMP", "x")

	out := captureStdout(t, func() {
		if err := runDelete(newCmd(), []string{"TEMP"}); err != nil {
			t.Fatalf("runDelete: %v", err)
		}
	})
	if !strings.Contains(out, "Deleted TEMP") {
		t.Errorf("unexpected output: %q", out)
	}
	if _, ok := readSecret(t, "default", "TEMP"); ok {
		t.Error("key still present after delete")
	}
}

func TestRunDelete_NotFound(t *testing.T) {
	setupVault(t)
	if err := runDelete(newCmd(), []string{"NOPE"}); err == nil {
		t.Error("expected error for missing key")
	}
}

func TestRunRename(t *testing.T) {
	setupVault(t)
	seed(t, "default", "OLD", "val")

	if err := runRename(newCmd(), []string{"OLD", "NEW"}); err != nil {
		t.Fatalf("runRename: %v", err)
	}
	if _, ok := readSecret(t, "default", "OLD"); ok {
		t.Error("OLD still present")
	}
	got, ok := readSecret(t, "default", "NEW")
	if !ok || got != "val" {
		t.Errorf("NEW = %q (ok=%v), want val", got, ok)
	}
}

func TestRunRename_SourceMissing(t *testing.T) {
	setupVault(t)
	if err := runRename(newCmd(), []string{"GHOST", "NEW"}); err == nil {
		t.Error("expected error when source key missing")
	}
}

func TestRunRename_TargetExists(t *testing.T) {
	setupVault(t)
	seed(t, "default", "A", "1")
	seed(t, "default", "B", "2")
	if err := runRename(newCmd(), []string{"A", "B"}); err == nil {
		t.Error("expected error when target key already exists")
	}
}

func TestRunMove(t *testing.T) {
	setupVault(t)
	seed(t, "default", "SHARED", "v")
	moveToNS = "prod"
	defer func() { moveToNS = "" }()

	if err := runMove(newCmd(), []string{"SHARED"}); err != nil {
		t.Fatalf("runMove: %v", err)
	}
	if _, ok := readSecret(t, "default", "SHARED"); ok {
		t.Error("key still in source namespace")
	}
	got, ok := readSecret(t, "prod", "SHARED")
	if !ok || got != "v" {
		t.Errorf("prod/SHARED = %q (ok=%v), want v", got, ok)
	}
}

func TestRunMove_SourceMissing(t *testing.T) {
	setupVault(t)
	moveToNS = "prod"
	defer func() { moveToNS = "" }()
	if err := runMove(newCmd(), []string{"GHOST"}); err == nil {
		t.Error("expected error when source key missing")
	}
}

func TestRunMove_SameNamespace(t *testing.T) {
	setupVault(t)
	seed(t, "default", "K", "v")
	moveToNS = "default"
	defer func() { moveToNS = "" }()
	if err := runMove(newCmd(), []string{"K"}); err == nil {
		t.Error("expected error when source and dest namespaces match")
	}
}
