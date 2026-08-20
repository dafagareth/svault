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

	"svault/internal/storage"
)

func TestRunStatus_Unlocked(t *testing.T) {
	setupVault(t) // setupVault saves a session, so we are unlocked
	statusShort = false

	out := captureStdout(t, func() {
		if err := runStatus(newCmd(), nil); err != nil {
			t.Fatalf("runStatus: %v", err)
		}
	})
	if !strings.Contains(out, "Unlocked") {
		t.Errorf("expected Unlocked status: %q", out)
	}
}

func TestRunStatus_UnlockedShowsStats(t *testing.T) {
	setupVault(t)
	statusShort = false
	seed(t, "default", "A", "1")
	seed(t, "prod", "B", "2")

	out := captureStdout(t, func() {
		if err := runStatus(newCmd(), nil); err != nil {
			t.Fatalf("runStatus: %v", err)
		}
	})
	for _, want := range []string{"Namespaces:", "Total keys:", "Active ns:", "Vault size:"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in output: %q", want, out)
		}
	}
}

func TestHumanSize(t *testing.T) {
	cases := map[int64]string{
		512:             "512 B",
		2048:            "2.0 KB",
		3 * 1024 * 1024: "3.0 MB",
	}
	for in, want := range cases {
		if got := humanSize(in); got != want {
			t.Errorf("humanSize(%d) = %q, want %q", in, got, want)
		}
	}
}

func TestRunStatus_Locked(t *testing.T) {
	setupVault(t)
	if err := storage.DeleteSession(); err != nil {
		t.Fatal(err)
	}
	statusShort = false

	out := captureStdout(t, func() {
		if err := runStatus(newCmd(), nil); err != nil {
			t.Fatalf("runStatus: %v", err)
		}
	})
	if !strings.Contains(out, "Locked") {
		t.Errorf("expected Locked status: %q", out)
	}
}

func TestRunStatus_ShortLocked(t *testing.T) {
	setupVault(t)
	_ = storage.DeleteSession()
	statusShort = true
	defer func() { statusShort = false }()

	out := captureStdout(t, func() {
		_ = runStatus(newCmd(), nil)
	})
	if !strings.Contains(out, "🔒") {
		t.Errorf("expected lock emoji: %q", out)
	}
}
