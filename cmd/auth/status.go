package auth

import (
	"fmt"
	"os"
	"svault/cmd/common"

	"github.com/spf13/cobra"
	"svault/internal/storage"
)

var statusShort bool

var StatusCmd = &cobra.Command{
	GroupID: "auth",
	Use:     "status",
	Short:   "Show whether the vault is locked or unlocked",
	Long: `Show whether the vault is locked or unlocked.

When unlocked, it also reports the namespace count, total keys, the active
namespace, and the vault file size. Use --short for a compact lock icon or
remaining minutes, suitable for shell prompts.

Examples:
  svault status              # full status with vault stats
  svault status --short      # compact form for a shell prompt`,
	RunE: runStatus,
}

func init() {
	StatusCmd.Flags().BoolVar(&statusShort, "short", false, "compact output for shell prompts")
}

func runStatus(_ *cobra.Command, _ []string) error {
	remaining, active := storage.SessionRemaining()

	if statusShort {
		if active {
			fmt.Printf("🔓 %dm\n", int(remaining.Minutes()))
		} else {
			fmt.Println("🔒")
		}
		return nil
	}

	if active {
		mins := int(remaining.Minutes())
		secs := int(remaining.Seconds()) % 60
		fmt.Printf("Unlocked, session expires in %dm %ds\n", mins, secs)
		printVaultStats()
	} else {
		fmt.Println("Locked. Run 'svault unlock' to start a session.")
		if vpath, err := common.VaultPath(); err == nil {
			if info, err := os.Stat(vpath); err == nil {
				fmt.Printf("Vault size:  %s\n", humanSize(info.Size()))
			}
		}
	}
	return nil
}

// printVaultStats reads the vault and prints namespace/key counts and file size.
// Best-effort: any error is silently skipped so 'status' never fails on stats.
func printVaultStats() {
	vpath, err := common.VaultPath()
	if err != nil {
		return
	}
	encKey, err := storage.LoadSession()
	if err != nil {
		return
	}
	vd, err := storage.ReadVault(vpath, encKey)
	if err != nil {
		return
	}

	totalKeys := 0
	for _, secrets := range vd.Namespaces {
		totalKeys += len(secrets)
	}

	nsName, _ := common.NamespaceSource()
	fmt.Printf("Namespaces:  %d\n", len(vd.Namespaces))
	fmt.Printf("Total keys:  %d\n", totalKeys)
	fmt.Printf("Active ns:   %s (%d keys)\n", nsName, len(vd.Namespaces[nsName]))
	if info, err := os.Stat(vpath); err == nil {
		fmt.Printf("Vault size:  %s\n", humanSize(info.Size()))
	}
}

// humanSize formats a byte count as B, KB, or MB.
func humanSize(n int64) string {
	switch {
	case n < 1024:
		return fmt.Sprintf("%d B", n)
	case n < 1024*1024:
		return fmt.Sprintf("%.1f KB", float64(n)/1024)
	default:
		return fmt.Sprintf("%.1f MB", float64(n)/(1024*1024))
	}
}
