package env

import (
	"encoding/json"
	"strings"
	"svault/cmd/common"
	"testing"
)

func TestRunEnv_Dotenv(t *testing.T) {
	common.SetupVault(t)
	defer func() { envFormat = "shell" }()
	common.Seed(t, "default", "HOST", "localhost")
	envFormat = "dotenv"

	out := common.CaptureStdout(t, func() {
		if err := runEnv(common.NewCmd(), nil); err != nil {
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
	common.SetupVault(t)
	defer func() { envFormat = "shell" }()
	common.Seed(t, "default", "HOST", "localhost")
	common.Seed(t, "default", "PORT", "8080")
	envFormat = "json"

	out := common.CaptureStdout(t, func() {
		if err := runEnv(common.NewCmd(), nil); err != nil {
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
	common.SetupVault(t)
	defer func() { envFormat = "shell" }()
	common.Seed(t, "default", "HOST", "localhost")
	envFormat = "yaml"

	out := common.CaptureStdout(t, func() {
		if err := runEnv(common.NewCmd(), nil); err != nil {
			t.Fatalf("runEnv: %v", err)
		}
	})
	if !strings.Contains(out, `HOST: "localhost"`) {
		t.Errorf("expected yaml line: %q", out)
	}
}

func TestRunEnv_UnknownFormat(t *testing.T) {
	common.SetupVault(t)
	defer func() { envFormat = "shell" }()
	common.Seed(t, "default", "HOST", "localhost")
	envFormat = "toml"

	if err := runEnv(common.NewCmd(), nil); err == nil {
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
