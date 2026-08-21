package ns

import (
	"svault/cmd/common"
	"testing"

	"github.com/spf13/cobra"
)

func TestNamespaceSource_EnvWins(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("SVAULT_NS", "from-env")

	ns, src := common.NamespaceSource()
	if ns != "from-env" || src != "env" {
		t.Errorf("got (%q, %q), want (from-env, env)", ns, src)
	}
}

func TestActiveNamespace_FlagOverridesEnv(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("SVAULT_NS", "from-env")

	// Build a command with the --ns flag registered and explicitly set.
	c := &cobra.Command{}
	var nsFlag string
	c.Flags().StringVar(&nsFlag, "ns", "default", "")
	if err := c.Flags().Set("ns", "from-flag"); err != nil {
		t.Fatal(err)
	}

	if got := common.ActiveNamespace(c, "from-flag"); got != "from-flag" {
		t.Errorf("got %q, want from-flag", got)
	}
}

func TestActiveNamespace_FallsBackToEnv(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("SVAULT_NS", "envns")

	// Bare command: --ns not changed, so it should fall through to SVAULT_NS.
	if got := common.ActiveNamespace(&cobra.Command{}, "default"); got != "envns" {
		t.Errorf("got %q, want envns", got)
	}
}
