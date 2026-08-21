package env

import (
	"fmt"
	"sort"
	"svault/cmd/common"

	"github.com/spf13/cobra"
	"svault/internal/envfile"
	"svault/internal/storage"
)

var checkNamespace string
var checkExtra bool
var checkStrict bool

var CheckCmd = &cobra.Command{
	GroupID: "env",
	Use:     "check [FILE]",
	Short:   "Check which .env.example keys are missing from the vault",
	Long: `Compare the keys in a .env file against the active namespace.

Reports each key as OK (present) or MISS (missing). With --extra it also lists
EXTRA keys that exist in the vault but not in the file. Use --strict in a
pre-flight step so a missing secret fails the pipeline before the app starts.

If no FILE is given, .env.example is used.

Examples:
  svault check                        # check against .env.example
  svault check .env.production        # check against another file
  svault check --extra                # also list unused vault keys
  svault check --strict               # exit non-zero if any key is missing`,
	Args: cobra.MaximumNArgs(1),
	RunE: runCheck,
}

func init() {
	CheckCmd.Flags().StringVar(&checkNamespace, "ns", "default", "namespace to check")
	CheckCmd.Flags().BoolVar(&checkExtra, "extra", false, "also list vault keys not present in the file")
	CheckCmd.Flags().BoolVar(&checkStrict, "strict", false, "exit non-zero if any key is missing")
}

func runCheck(cmd *cobra.Command, args []string) error {
	file := ".env.example"
	if len(args) == 1 {
		file = args[0]
	}

	keys, err := envfile.Keys(file)
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		return fmt.Errorf("no keys found in %s", file)
	}

	ns := common.ActiveNamespace(cmd, checkNamespace)
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

	vaultNs := vd.Namespaces[ns]
	found, missing := 0, 0
	inFile := make(map[string]bool, len(keys))
	for _, k := range keys {
		inFile[k] = true
		if _, ok := vaultNs[k]; ok {
			fmt.Printf("OK    %s\n", k)
			found++
		} else {
			fmt.Printf("MISS  %s\n", k)
			missing++
		}
	}

	if checkExtra {
		extra := make([]string, 0)
		for k := range vaultNs {
			if !inFile[k] {
				extra = append(extra, k)
			}
		}
		sort.Strings(extra)
		for _, k := range extra {
			fmt.Printf("EXTRA %s\n", k)
		}
	}

	fmt.Printf("\n%d/%d keys present in namespace [%s]", found, len(keys), ns)
	if missing > 0 {
		fmt.Printf(", %d missing", missing)
	}
	fmt.Println()

	if checkStrict && missing > 0 {
		return fmt.Errorf("%d key(s) missing from namespace [%s]", missing, ns)
	}
	return nil
}
