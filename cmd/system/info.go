package system

import (
	"fmt"
	"os"
	"strings"
	"svault/cmd/common"

	"github.com/spf13/cobra"
	"svault/internal/storage"
)

// Version is set by Execute() in cmd/root.go from the -ldflags build injection.
// Using the exported Version (from version.go) ensures svault info always prints
// the real build version instead of always printing "dev".

var InfoCmd = &cobra.Command{
	GroupID: "system",
	Use:     "info",
	Short:   "Show vault information",
	RunE:    runInfo,
}

func runInfo(_ *cobra.Command, _ []string) error {
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

	totalSecrets := 0
	for _, ns := range vd.Namespaces {
		totalSecrets += len(ns)
	}

	displayPath := vpath
	if home, err := os.UserHomeDir(); err == nil && strings.HasPrefix(vpath, home) {
		displayPath = "~" + vpath[len(home):]
	}

	ns, src := common.NamespaceSource()

	remaining, sessionActive := storage.SessionRemaining()
	var sessionStr string
	if sessionActive {
		sessionStr = fmt.Sprintf("unlocked, %dm remaining", int(remaining.Minutes()))
	} else {
		sessionStr = "locked"
	}

	fmt.Printf("svault v%s\n", Version)
	fmt.Printf("Vault file : %s\n", displayPath)
	fmt.Printf("Namespace  : %s (from %s, %d total)\n", ns, src, len(vd.Namespaces))
	fmt.Printf("Secrets    : %d total\n", totalSecrets)
	fmt.Printf("Encryption : AES-256-GCM\n")
	fmt.Printf("Session    : %s\n", sessionStr)
	return nil
}
