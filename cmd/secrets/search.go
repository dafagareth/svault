package secrets

import (
	"fmt"
	"sort"
	"strings"
	"svault/cmd/common"

	"github.com/spf13/cobra"
	"svault/internal/storage"
)

var searchAllNamespaces bool

var SearchCmd = &cobra.Command{
	GroupID: "secrets",
	Use:     "search PATTERN",
	Short:   "Search keys by name (case-insensitive substring match)",
	Args:    cobra.ExactArgs(1),
	RunE:    runSearch,
}

func init() {
	SearchCmd.Flags().BoolVar(&searchAllNamespaces, "all", false, "search across all namespaces")
}

func runSearch(cmd *cobra.Command, args []string) error {
	pattern := strings.ToLower(args[0])

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

	nsNames := []string{common.ActiveNamespace(cmd, "default")}
	if searchAllNamespaces {
		nsNames = nsNames[:0]
		for n := range vd.Namespaces {
			nsNames = append(nsNames, n)
		}
		sort.Strings(nsNames)
	}

	matches := 0
	for _, n := range nsNames {
		keys := make([]string, 0, len(vd.Namespaces[n]))
		for k := range vd.Namespaces[n] {
			if strings.Contains(strings.ToLower(k), pattern) {
				keys = append(keys, k)
			}
		}
		if len(keys) == 0 {
			continue
		}
		sort.Strings(keys)
		if searchAllNamespaces {
			fmt.Printf("[%s]\n", n)
			for _, k := range keys {
				fmt.Printf("  %s\n", k)
			}
		} else {
			for _, k := range keys {
				fmt.Println(k)
			}
		}
		matches += len(keys)
	}

	if matches == 0 {
		fmt.Printf("No keys match %q\n", args[0])
	}
	return nil
}
