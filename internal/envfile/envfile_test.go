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

package envfile

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, ".env")
	if err := os.WriteFile(p, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestParse(t *testing.T) {
	content := `# a comment

DB_URL=postgres://localhost
API_KEY="quoted"
TOKEN='single'
SPACED  =  value with spaces
NO_EQUALS_LINE
=novalue
`
	entries, err := Parse(writeTemp(t, content))
	if err != nil {
		t.Fatal(err)
	}

	want := map[string]string{
		"DB_URL":  "postgres://localhost",
		"API_KEY": "quoted",
		"TOKEN":   "single",
		"SPACED":  "value with spaces",
	}
	if len(entries) != len(want) {
		t.Fatalf("got %d entries, want %d: %+v", len(entries), len(want), entries)
	}
	for _, e := range entries {
		w, ok := want[e.Key]
		if !ok {
			t.Errorf("unexpected key %q", e.Key)
			continue
		}
		if e.Value != w {
			t.Errorf("%s = %q, want %q", e.Key, e.Value, w)
		}
	}
}

func TestParse_PreservesOrder(t *testing.T) {
	entries, err := Parse(writeTemp(t, "C=3\nA=1\nB=2\n"))
	if err != nil {
		t.Fatal(err)
	}
	order := []string{entries[0].Key, entries[1].Key, entries[2].Key}
	want := []string{"C", "A", "B"}
	for i := range want {
		if order[i] != want[i] {
			t.Errorf("order = %v, want %v", order, want)
			break
		}
	}
}

func TestParse_ValueWithEquals(t *testing.T) {
	entries, err := Parse(writeTemp(t, "URL=https://x.com?a=1&b=2\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Value != "https://x.com?a=1&b=2" {
		t.Errorf("got %+v", entries)
	}
}

func TestParse_FileNotFound(t *testing.T) {
	if _, err := Parse("/nonexistent/path/.env"); err == nil {
		t.Error("expected error for missing file")
	}
}

func TestKeys_SortedAndDeduped(t *testing.T) {
	keys, err := Keys(writeTemp(t, "B=1\nA=2\nB=3\nC=4\n"))
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"A", "B", "C"}
	if len(keys) != len(want) {
		t.Fatalf("got %v, want %v", keys, want)
	}
	for i := range want {
		if keys[i] != want[i] {
			t.Errorf("keys = %v, want %v", keys, want)
			break
		}
	}
}
