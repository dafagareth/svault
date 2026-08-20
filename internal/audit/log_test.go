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

	for _, want := range []string{"SET", "MY_KEY", "[default]", "GET", "OTHER_KEY", "[production]", "DELETE", "OLD_KEY"} {
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
