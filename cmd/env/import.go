package env

import (
	"fmt"
	"svault/cmd/common"

	"github.com/spf13/cobra"
	"svault/internal/audit"
	"svault/internal/envfile"
	"svault/internal/storage"
)

var importNamespace string
var importOverwrite bool

var ImportCmd = &cobra.Command{
	GroupID: "env",
	Use:     "import FILE",
	Short:   "Import secrets from a .env file",
	Args:    cobra.ExactArgs(1),
	RunE:    runImport,
}

func init() {
	ImportCmd.Flags().StringVar(&importNamespace, "ns", "default", "target namespace")
	ImportCmd.Flags().BoolVar(&importOverwrite, "overwrite", true, "overwrite existing keys (use --overwrite=false to skip)")
}

func runImport(cmd *cobra.Command, args []string) error {
	ns := common.ActiveNamespace(cmd, importNamespace)

	parsed, err := envfile.Parse(args[0])
	if err != nil {
		return err
	}

	entries := make(map[string]string, len(parsed))
	for _, e := range parsed {
		if err := common.ValidateKey(e.Key); err != nil {
			return fmt.Errorf("invalid key in %s: %w", args[0], err)
		}
		entries[e.Key] = e.Value
	}
	if len(entries) == 0 {
		return fmt.Errorf("no valid keys found in %s", args[0])
	}

	skipped := 0
	err = common.MutateVault(func(vd *storage.VaultData) error {
		if vd.Namespaces[ns] == nil {
			vd.Namespaces[ns] = make(map[string]string)
		}
		for k, v := range entries {
			if !importOverwrite {
				if _, exists := vd.Namespaces[ns][k]; exists {
					skipped++
					continue
				}
			}
			vd.Namespaces[ns][k] = v
		}
		return nil
	})
	if err != nil {
		return err
	}

	lpath, _ := common.LogPath()
	_ = audit.Append(lpath, audit.OpImport, "-", ns)

	imported := len(entries) - skipped
	fmt.Printf("Imported %d keys into [%s]", imported, ns)
	if skipped > 0 {
		fmt.Printf(", %d skipped (already exist)", skipped)
	}
	fmt.Println()
	return nil
}
