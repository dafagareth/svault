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
	"encoding/json"
	"strings"
	"testing"
)

// resetListFlags restores every list flag to its zero value so global state does
// not leak between tests.
func resetListFlags() {
	listAll = false
	listValues = false
	listMask = false
	listLong = false
	listFilter = ""
	listCount = false
	listJSON = false
}

func TestFilteredKeys(t *testing.T) {
	secrets := map[string]string{
		"DB_HOST":     "h",
		"DB_PORT":     "p",
		"JWT_SECRET":  "s",
		"MAIL_SENDER": "m",
	}

	// No filter: all keys, sorted.
	got := filteredKeys(secrets, "")
	want := []string{"DB_HOST", "DB_PORT", "JWT_SECRET", "MAIL_SENDER"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("no filter: got %v, want %v", got, want)
	}

	// Case-insensitive substring filter.
	got = filteredKeys(secrets, "db")
	want = []string{"DB_HOST", "DB_PORT"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("filter db: got %v, want %v", got, want)
	}

	// No match returns empty.
	if got := filteredKeys(secrets, "zzz"); len(got) != 0 {
		t.Errorf("filter zzz: got %v, want empty", got)
	}
}

func TestRunList_Values(t *testing.T) {
	setupVault(t)
	defer resetListFlags()
	seed(t, "default", "TOKEN", "abc123")
	listValues = true

	out := captureStdout(t, func() {
		if err := runList(newCmd(), nil); err != nil {
			t.Fatalf("runList: %v", err)
		}
	})
	if !strings.Contains(out, "TOKEN") || !strings.Contains(out, "abc123") {
		t.Errorf("expected key and value: %q", out)
	}
}

func TestRunList_Long(t *testing.T) {
	setupVault(t)
	defer resetListFlags()
	seed(t, "default", "TOKEN", "abc123") // len 6
	listLong = true

	out := captureStdout(t, func() {
		if err := runList(newCmd(), nil); err != nil {
			t.Fatalf("runList: %v", err)
		}
	})
	// Long format shows key and length, but NOT the value unless -v/-m is added.
	if !strings.Contains(out, "TOKEN") || !strings.Contains(out, "6") {
		t.Errorf("expected key and length: %q", out)
	}
	if strings.Contains(out, "abc123") {
		t.Errorf("long format leaked the value: %q", out)
	}
}

func TestRunList_LongValues(t *testing.T) {
	setupVault(t)
	defer resetListFlags()
	seed(t, "default", "TOKEN", "abc123") // len 6
	listLong = true
	listValues = true

	out := captureStdout(t, func() {
		if err := runList(newCmd(), nil); err != nil {
			t.Fatalf("runList: %v", err)
		}
	})
	// With -l -v the value is revealed alongside the length.
	if !strings.Contains(out, "TOKEN") || !strings.Contains(out, "6") || !strings.Contains(out, "abc123") {
		t.Errorf("expected key, length and value: %q", out)
	}
}

func TestRunList_Mask(t *testing.T) {
	setupVault(t)
	defer resetListFlags()
	seed(t, "default", "TOKEN", "abc123")
	listMask = true

	out := captureStdout(t, func() {
		if err := runList(newCmd(), nil); err != nil {
			t.Fatalf("runList: %v", err)
		}
	})
	if strings.Contains(out, "abc123") {
		t.Errorf("masked output leaked the value: %q", out)
	}
	if !strings.Contains(out, "****") {
		t.Errorf("expected masked value: %q", out)
	}
}

func TestRunList_Filter(t *testing.T) {
	setupVault(t)
	defer resetListFlags()
	seed(t, "default", "DB_HOST", "h")
	seed(t, "default", "JWT_SECRET", "s")
	listFilter = "db"

	out := captureStdout(t, func() {
		if err := runList(newCmd(), nil); err != nil {
			t.Fatalf("runList: %v", err)
		}
	})
	if !strings.Contains(out, "DB_HOST") {
		t.Errorf("expected DB_HOST: %q", out)
	}
	if strings.Contains(out, "JWT_SECRET") {
		t.Errorf("filter should have excluded JWT_SECRET: %q", out)
	}
}

func TestRunList_Count(t *testing.T) {
	setupVault(t)
	defer resetListFlags()
	seed(t, "default", "A", "1")
	seed(t, "default", "B", "2")
	seed(t, "default", "C", "3")
	listCount = true

	out := captureStdout(t, func() {
		if err := runList(newCmd(), nil); err != nil {
			t.Fatalf("runList: %v", err)
		}
	})
	if strings.TrimSpace(out) != "3" {
		t.Errorf("got %q, want 3", strings.TrimSpace(out))
	}
}

func TestRunList_JSONKeys(t *testing.T) {
	setupVault(t)
	defer resetListFlags()
	seed(t, "default", "A", "1")
	seed(t, "default", "B", "2")
	listJSON = true

	out := captureStdout(t, func() {
		if err := runList(newCmd(), nil); err != nil {
			t.Fatalf("runList: %v", err)
		}
	})
	var keys []string
	if err := json.Unmarshal([]byte(out), &keys); err != nil {
		t.Fatalf("output is not a JSON array: %v\n%s", err, out)
	}
	if strings.Join(keys, ",") != "A,B" {
		t.Errorf("got %v, want [A B]", keys)
	}
}

func TestRunList_JSONValues(t *testing.T) {
	setupVault(t)
	defer resetListFlags()
	seed(t, "default", "A", "1")
	listJSON = true
	listValues = true

	out := captureStdout(t, func() {
		if err := runList(newCmd(), nil); err != nil {
			t.Fatalf("runList: %v", err)
		}
	})
	var m map[string]string
	if err := json.Unmarshal([]byte(out), &m); err != nil {
		t.Fatalf("output is not a JSON object: %v\n%s", err, out)
	}
	if m["A"] != "1" {
		t.Errorf("got %v, want {A:1}", m)
	}
}
