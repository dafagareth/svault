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
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"svault/internal/storage"
)

var rootCmd = &cobra.Command{
	Use:   "svault",
	Short: "Local encrypted secret vault",
	// Cobra prints the error itself (with its "Error:" prefix and any
	// "Did you mean" suggestion); we just exit non-zero in Execute. Silencing
	// usage keeps a plain runtime error from dumping the full help text.
	SilenceUsage: true,
}

func Execute() {
	rootCmd.Version = version
	rootCmd.SetVersionTemplate("svault v{{.Version}}\n")
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		exempt := map[string]bool{
			"init":       true,
			"help":       true,
			"completion": true,
			"version":    true,
			"doctor":     true,
		}
		if exempt[cmd.Name()] {
			return nil
		}
		vpath, err := vaultPath()
		if err != nil {
			return err
		}
		if _, err := os.Stat(vpath); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "svault: no vault found.\n\nGet started:\n  svault init        create a new encrypted vault\n  svault --help      show all commands\n\nVault will be stored at ~/.svault/vault.enc\n")
			os.Exit(1)
		}
		return nil
	}
	// Route a leading flag (for example `svault -l`) to the `list` command, so
	// list flags work without typing `list`. A bare `svault` shows help.
	if shouldDefaultToList(os.Args[1:]) {
		os.Args = append([]string{os.Args[0], "list"}, os.Args[1:]...)
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// shouldDefaultToList reports whether an invocation should be routed to `list`.
// True only when the first arg is a flag other than the root-level -h/--help or
// -v/--version. A bare `svault` (no args) returns false so cobra shows help, and
// a leading subcommand name (no dash) is left untouched so `svault get X`,
// `svault completion`, etc. work normally.
func shouldDefaultToList(args []string) bool {
	if len(args) == 0 {
		return false
	}
	first := args[0]
	if !strings.HasPrefix(first, "-") {
		return false // a subcommand (or typo) — let cobra handle it
	}
	switch first {
	case "-h", "--help", "-v", "--version":
		return false
	}
	return true
}

func vaultDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}
	return filepath.Join(home, ".svault"), nil
}

func vaultPath() (string, error) {
	dir, err := vaultDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "vault.enc"), nil
}

func logPath() (string, error) {
	dir, err := vaultDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "vault.log"), nil
}

// gitRepoName returns the basename of the current git repo root, or "".
func gitRepoName() string {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return ""
	}
	return filepath.Base(strings.TrimSpace(string(out)))
}

// namespaceSource returns (namespace, source) where source is "git", "env", "config", or "default".
func namespaceSource() (string, string) {
	if ns := os.Getenv("SVAULT_NS"); ns != "" {
		return ns, "env"
	}
	if ns := gitRepoName(); ns != "" {
		return ns, "git"
	}
	dir, err := vaultDir()
	if err != nil {
		return "default", "default"
	}
	cfg, err := storage.ReadConfig(dir)
	if err != nil {
		return "default", "default"
	}
	if cfg.ActiveNamespace != "" && cfg.ActiveNamespace != "default" {
		return cfg.ActiveNamespace, "config"
	}
	return "default", "default"
}

// activeNamespace returns the namespace to use for a command.
// Priority: --ns flag > SVAULT_NS env > git repo name > config > "default".
func activeNamespace(cmd *cobra.Command, flagVal string) string {
	if cmd.Flags().Changed("ns") {
		return flagVal
	}
	ns, _ := namespaceSource()
	return ns
}
