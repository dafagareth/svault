package utils

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"svault/cmd/common"

	"github.com/spf13/cobra"
	"svault/internal/storage"
)

var openNamespace string

var OpenCmd = &cobra.Command{
	GroupID: "utils",
	Use:     "open KEY",
	Short:   "Open a URL from vault in browser and copy password to clipboard",
	Long: `Opens the URL stored in KEY_URL (or KEY if it starts with http) in the
default browser, then copies KEY_PASS or KEY_PASSWORD to clipboard.

Examples:
  svault set GITHUB_URL=https://github.com/login
  svault set GITHUB_PASS=mysecret
  svault open GITHUB`,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return common.CompleteActiveKeys(cmd)
	},
	Args: cobra.ExactArgs(1),
	RunE: runOpen,
}

func init() {
	OpenCmd.Flags().StringVar(&openNamespace, "ns", "default", "namespace")
}

func runOpen(cmd *cobra.Command, args []string) error {
	key := strings.ToUpper(args[0])
	ns := common.ActiveNamespace(cmd, openNamespace)

	vpath, err := common.VaultPath()
	if err != nil {
		return err
	}
	encKey, err := storage.LoadSession()
	if err != nil {
		return err
	}
	vd, err := storage.ReadVault(vpath, encKey)
	if err != nil {
		return err
	}

	nsMap := vd.Namespaces[ns]

	// Resolve URL: KEY_URL → KEY (if starts with http) → error
	url := ""
	for _, candidate := range []string{key + "_URL", key} {
		if v, ok := nsMap[candidate]; ok && (strings.HasPrefix(v, "http://") || strings.HasPrefix(v, "https://")) {
			url = v
			break
		}
	}
	if url == "" {
		return fmt.Errorf("no URL found for %s (set %s_URL or %s with an http URL)", key, key, key)
	}

	if err := openBrowser(url); err != nil {
		return fmt.Errorf("open browser: %w", err)
	}
	fmt.Printf("Opening %s\n", url)

	// Resolve password: KEY_PASS → KEY_PASSWORD → KEY_TOKEN → not found
	passKey := ""
	for _, candidate := range []string{key + "_PASS", key + "_PASSWORD", key + "_TOKEN"} {
		if _, ok := nsMap[candidate]; ok {
			passKey = candidate
			break
		}
	}

	if passKey != "" {
		if err := common.CopyToClipboard(nsMap[passKey]); err == nil {
			common.ScheduleClipboardClear(30)
			fmt.Printf("Copied %s to clipboard. Will clear in 30 seconds.\n", passKey)
		}
	} else {
		fmt.Printf("Tip: set %s_PASS to auto-copy password on open.\n", key)
	}

	return nil
}

// openBrowser opens the given URL in the default browser.
// Supports Linux (xdg-open), macOS (open), and Windows (start).
func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}
