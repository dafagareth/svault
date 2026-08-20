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

// resetGenerateFlags restores generate's package-level flags to safe defaults
// so tests do not leak state into each other or touch the clipboard.
func resetGenerateFlags() {
	genLength = 24
	genNoSymbols = false
	genSaveKey = ""
	genNamespace = "default"
	genNoCopy = true // never touch the real clipboard in tests
}

func TestRunGenerate_LengthAndCharset(t *testing.T) {
	setupVault(t)
	resetGenerateFlags()
	genLength = 32
	genNoSymbols = true

	out := captureStdout(t, func() {
		if err := runGenerate(newCmd(), nil); err != nil {
			t.Fatalf("runGenerate: %v", err)
		}
	})
	password := strings.TrimSpace(out)
	if len(password) != 32 {
		t.Errorf("length = %d, want 32", len(password))
	}
	const allowed = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	for _, c := range password {
		if !strings.ContainsRune(allowed, c) {
			t.Errorf("unexpected symbol %q in --no-symbols output", c)
		}
	}
}

func TestRunGenerate_Save(t *testing.T) {
	setupVault(t)
	resetGenerateFlags()
	genSaveKey = "NEW_PASS"

	out := captureStdout(t, func() {
		if err := runGenerate(newCmd(), nil); err != nil {
			t.Fatalf("runGenerate: %v", err)
		}
	})
	password := strings.SplitN(strings.TrimSpace(out), "\n", 2)[0]

	stored, ok := readSecret(t, "default", "NEW_PASS")
	if !ok {
		t.Fatal("generated password not saved")
	}
	if stored != password {
		t.Errorf("stored %q != printed %q", stored, password)
	}
}

func TestRunGenerate_TooShort(t *testing.T) {
	setupVault(t)
	resetGenerateFlags()
	genLength = 4
	if err := runGenerate(newCmd(), nil); err == nil {
		t.Error("expected error for length < 8")
	}
}

func TestRunGenerate_Randomness(t *testing.T) {
	setupVault(t)
	resetGenerateFlags()

	first := captureStdout(t, func() { _ = runGenerate(newCmd(), nil) })
	second := captureStdout(t, func() { _ = runGenerate(newCmd(), nil) })
	if strings.TrimSpace(first) == strings.TrimSpace(second) {
		t.Error("two generated passwords were identical")
	}
}
