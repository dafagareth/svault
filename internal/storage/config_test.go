package storage

import (
	"testing"
)

func TestReadConfigDefault(t *testing.T) {
	dir := t.TempDir()

	cfg, err := ReadConfig(dir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ActiveNamespace != "default" {
		t.Errorf("namespace default: got %q, want %q", cfg.ActiveNamespace, "default")
	}
}

func TestWriteReadConfigRoundTrip(t *testing.T) {
	dir := t.TempDir()

	if err := WriteConfig(dir, &Config{ActiveNamespace: "production"}); err != nil {
		t.Fatal("write:", err)
	}
	cfg, err := ReadConfig(dir)
	if err != nil {
		t.Fatal("read:", err)
	}
	if cfg.ActiveNamespace != "production" {
		t.Errorf("got %q, want %q", cfg.ActiveNamespace, "production")
	}
}

func TestReadConfigEmptyNamespaceFallsBackToDefault(t *testing.T) {
	dir := t.TempDir()

	if err := WriteConfig(dir, &Config{ActiveNamespace: ""}); err != nil {
		t.Fatal(err)
	}
	cfg, err := ReadConfig(dir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ActiveNamespace != "default" {
		t.Errorf("got %q, want %q", cfg.ActiveNamespace, "default")
	}
}
