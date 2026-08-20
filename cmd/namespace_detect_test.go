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
	"testing"

	"github.com/spf13/cobra"
)

func TestNamespaceSource_EnvWins(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("SVAULT_NS", "from-env")

	ns, src := namespaceSource()
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

	if got := activeNamespace(c, "from-flag"); got != "from-flag" {
		t.Errorf("got %q, want from-flag", got)
	}
}

func TestActiveNamespace_FallsBackToEnv(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("SVAULT_NS", "envns")

	// Bare command: --ns not changed, so it should fall through to SVAULT_NS.
	if got := activeNamespace(&cobra.Command{}, "default"); got != "envns" {
		t.Errorf("got %q, want envns", got)
	}
}
