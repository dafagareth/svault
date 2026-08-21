package secrets

import (
	"fmt"
	"svault/cmd/common"

	"github.com/spf13/cobra"
	"svault/internal/audit"
	"svault/internal/storage"
)

var getNamespace string
var getClip bool
var getTimeout int

var GetCmd = &cobra.Command{
	GroupID: "secrets",
	Use:     "get KEY [KEY...]",
	Short:   "Get one or more secret values",
	Long: `Print one or more secret values from the active namespace.

A single key prints just the value. Several keys print KEY=VALUE per line.
With --clip the value is copied to the clipboard instead of printed, and the
clipboard is cleared after --timeout seconds.

Examples:
  svault get DB_PASSWORD              # print the value
  svault get DB_HOST DB_PORT          # print KEY=VALUE for each
  svault get DB_PASSWORD --clip       # copy to clipboard, clears in 20s
  svault get DB_PASSWORD -c --timeout 0   # copy, do not auto-clear`,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return common.CompleteActiveKeys(cmd)
	},
	Args: cobra.MinimumNArgs(1),
	RunE: runGet,
}

func init() {
	GetCmd.Flags().StringVar(&getNamespace, "ns", "default", "source namespace")
	GetCmd.Flags().BoolVarP(&getClip, "clip", "c", false, "copy value to clipboard instead of printing")
	GetCmd.Flags().IntVar(&getTimeout, "timeout", 20, "seconds before clearing the clipboard (0 = never)")
}

func runGet(cmd *cobra.Command, args []string) error {
	ns := common.ActiveNamespace(cmd, getNamespace)

	if getClip && len(args) > 1 {
		return fmt.Errorf("--clip works with a single key only")
	}

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

	lpath, _ := common.LogPath()
	multi := len(args) > 1

	for _, key := range args {
		value, ok := vd.Namespaces[ns][key]
		if !ok {
			return fmt.Errorf("key not found: %s", key)
		}
		_ = audit.Append(lpath, audit.OpGet, key, ns)

		if getClip {
			if err := common.CopyToClipboard(value); err != nil {
				return err
			}
			fmt.Printf("Copied %s to clipboard", key)
			if getTimeout > 0 {
				common.ScheduleClipboardClear(getTimeout)
				fmt.Printf(" (clears in %ds)", getTimeout)
			}
			fmt.Println()
			return nil
		}

		if multi {
			fmt.Printf("%s=%s\n", key, value)
		} else {
			fmt.Println(value)
		}
	}
	return nil
}
