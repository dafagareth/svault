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

func TestRunEnv_Dotenv(t *testing.T) {
	setupVault(t)
	defer func() { envFormat = "shell" }()
	seed(t, "default", "HOST", "localhost")
	envFormat = "dotenv"

	out := captureStdout(t, func() {
		if err := runEnv(newCmd(), nil); err != nil {
			t.Fatalf("runEnv: %v", err)
		}
	})
	if !strings.Contains(out, "HOST=localhost") {
		t.Errorf("expected dotenv line: %q", out)
	}
	if strings.Contains(out, "export ") {
		t.Errorf("dotenv format should not contain 'export': %q", out)
	}
}

func TestRunEnv_JSON(t *testing.T) {
	setupVault(t)
	defer func() { envFormat = "shell" }()
	seed(t, "default", "HOST", "localhost")
	seed(t, "default", "PORT", "8080")
	envFormat = "json"

	out := captureStdout(t, func() {
		if err := runEnv(newCmd(), nil); err != nil {
			t.Fatalf("runEnv: %v", err)
		}
	})
	var m map[string]string
	if err := json.Unmarshal([]byte(out), &m); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, out)
	}
	if m["HOST"] != "localhost" || m["PORT"] != "8080" {
		t.Errorf("unexpected JSON: %v", m)
	}
}

func TestRunEnv_YAML(t *testing.T) {
	setupVault(t)
	defer func() { envFormat = "shell" }()
	seed(t, "default", "HOST", "localhost")
	envFormat = "yaml"

	out := captureStdout(t, func() {
		if err := runEnv(newCmd(), nil); err != nil {
			t.Fatalf("runEnv: %v", err)
		}
	})
	if !strings.Contains(out, `HOST: "localhost"`) {
		t.Errorf("expected yaml line: %q", out)
	}
}

func TestRunEnv_UnknownFormat(t *testing.T) {
	setupVault(t)
	defer func() { envFormat = "shell" }()
	seed(t, "default", "HOST", "localhost")
	envFormat = "toml"

	if err := runEnv(newCmd(), nil); err == nil {
		t.Error("expected error for unknown format")
	}
}

func TestYamlQuote(t *testing.T) {
	cases := map[string]string{
		"simple":        `"simple"`,
		`with "quotes"`: `"with \"quotes\""`,
		`back\slash`:    `"back\\slash"`,
	}
	for in, want := range cases {
		if got := yamlQuote(in); got != want {
			t.Errorf("yamlQuote(%q) = %q, want %q", in, got, want)
		}
	}
}
