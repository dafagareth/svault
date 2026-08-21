package secrets

import (
	"fmt"
	"io"
	"os"
	"strings"
	"svault/cmd/common"

	"github.com/spf13/cobra"
	"svault/internal/audit"
	"svault/internal/storage"
)

var setNamespace string
var setQuiet bool
var setStdin bool
var setClip bool

// setInput is the reader used for --stdin. Overridable in tests.
var setInput io.Reader = os.Stdin

var SetCmd = &cobra.Command{
	GroupID: "secrets",
	Use:     "set KEY [VALUE]",
	Short:   "Save a secret to the vault",
	Long: `Save a secret into the active namespace.

The value can come from an argument, stdin, or the clipboard. Reading from stdin
or the clipboard keeps the value out of your shell history.

Examples:
  svault set DB_PASSWORD supersecret        # KEY VALUE form
  svault set API_KEY=sk-123                  # KEY=VALUE form
  echo "supersecret" | svault set DB_PASSWORD --stdin
  svault set API_KEY --clip                  # read from the clipboard`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runSet,
}

func init() {
	SetCmd.Flags().StringVar(&setNamespace, "ns", "default", "target namespace (overrides auto-detect)")
	SetCmd.Flags().BoolVarP(&setQuiet, "quiet", "q", false, "suppress output")
	SetCmd.Flags().BoolVar(&setStdin, "stdin", false, "read value from stdin (keeps it out of shell history)")
	SetCmd.Flags().BoolVar(&setClip, "clip", false, "read value from the clipboard")
}

func runSet(cmd *cobra.Command, args []string) error {
	if setStdin && setClip {
		return fmt.Errorf("use only one of --stdin or --clip")
	}

	var key, value string
	if setClip {
		key = args[0]
		if strings.ContainsRune(key, '=') {
			return fmt.Errorf("with --clip, pass only the KEY, not KEY=VALUE")
		}
		clip, err := common.ReadClipboard()
		if err != nil {
			return err
		}
		value = clip
	} else if setStdin {
		key = args[0]
		if strings.ContainsRune(key, '=') {
			return fmt.Errorf("with --stdin, pass only the KEY, not KEY=VALUE")
		}
		data, err := io.ReadAll(setInput)
		if err != nil {
			return fmt.Errorf("read stdin: %w", err)
		}
		// Strip a single trailing newline so 'echo x | svault set K --stdin' is clean.
		value = strings.TrimSuffix(string(data), "\n")
	} else if len(args) == 1 {
		idx := strings.IndexByte(args[0], '=')
		if idx < 1 {
			return fmt.Errorf("invalid format: use KEY=VALUE or KEY VALUE")
		}
		key, value = args[0][:idx], args[0][idx+1:]
	} else {
		key, value = args[0], args[1]
	}
	if err := common.ValidateKey(key); err != nil {
		return err
	}
	ns := common.ActiveNamespace(cmd, setNamespace)

	err := common.MutateVault(func(vd *storage.VaultData) error {
		if vd.Namespaces[ns] == nil {
			vd.Namespaces[ns] = make(map[string]string)
		}
		vd.Namespaces[ns][key] = value
		return nil
	})
	if err != nil {
		return err
	}

	lpath, _ := common.LogPath()
	_ = audit.Append(lpath, audit.OpSet, key, ns)

	if !setQuiet {
		if !setStdin && !setClip {
			fmt.Fprintln(os.Stderr, "Notice: Secret values passed via CLI arguments may be visible in process listings. Consider --stdin or --clip.")
		}
		fmt.Printf("OK  %s saved\n", key)
	}
	return nil
}
