package cmd

import (
	"testing"
)

func TestRootCmdConfigured(t *testing.T) {
	if rootCmd.Use != "svault" {
		t.Errorf("rootCmd Use = %q, want 'svault'", rootCmd.Use)
	}
	if !rootCmd.CompletionOptions.DisableDefaultCmd {
		t.Error("expected DisableDefaultCmd to be true")
	}
}
