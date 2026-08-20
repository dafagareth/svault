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

import "testing"

func TestValidateKey_Valid(t *testing.T) {
	valid := []string{"DB_URL", "_private", "API_KEY_2", "a", "X9", "lower_case"}
	for _, k := range valid {
		if err := validateKey(k); err != nil {
			t.Errorf("validateKey(%q) = %v, want nil", k, err)
		}
	}
}

func TestValidateKey_Invalid(t *testing.T) {
	invalid := []string{
		"",          // empty
		"FOO BAR",   // space
		"FOO=BAR",   // equals
		"9LEADING",  // starts with digit
		"has-dash",  // dash
		"new\nline", // newline
		"dot.key",   // dot
	}
	for _, k := range invalid {
		if err := validateKey(k); err == nil {
			t.Errorf("validateKey(%q) = nil, want error", k)
		}
	}
}

func TestRunSet_RejectsInvalidKey(t *testing.T) {
	setupVault(t)
	setQuiet = true
	defer func() { setQuiet = false }()

	if err := runSet(newCmd(), []string{"BAD KEY", "value"}); err == nil {
		t.Error("expected error for key with a space")
	}
	if _, ok := readSecret(t, "default", "BAD KEY"); ok {
		t.Error("invalid key should not have been stored")
	}
}
