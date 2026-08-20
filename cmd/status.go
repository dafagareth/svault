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

	"github.com/spf13/cobra"
	"svault/internal/storage"
)

var statusShort bool

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show whether the vault is locked or unlocked",
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
	statusCmd.Flags().BoolVar(&statusShort, "short", false, "compact output for shell prompts")
	rootCmd.AddCommand(statusCmd)
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
		if vpath, err := vaultPath(); err == nil {
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
	vpath, err := vaultPath()
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

	nsName, _ := namespaceSource()
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
