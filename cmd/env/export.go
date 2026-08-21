package env

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"svault/cmd/common"

	"github.com/spf13/cobra"
	"svault/internal/audit"
	"svault/internal/storage"
)

var exportNamespace string
var exportOutput string

var ExportCmd = &cobra.Command{
	GroupID: "env",
	Use:     "export",
	Short:   "Export secrets to .env format",
	RunE:    runExport,
}

func init() {
	ExportCmd.Flags().StringVar(&exportNamespace, "ns", "default", "source namespace")
	ExportCmd.Flags().StringVarP(&exportOutput, "output", "o", "", "write to file instead of stdout")
}

func runExport(cmd *cobra.Command, _ []string) error {
	ns := common.ActiveNamespace(cmd, exportNamespace)

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

	secrets := vd.Namespaces[ns]
	keys := make([]string, 0, len(secrets))
	for k := range secrets {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build output in memory first so nothing is written on error.
	var buf bytes.Buffer
	for _, k := range keys {
		fmt.Fprintf(&buf, "%s=%s\n", k, secrets[k])
	}

	if exportOutput != "" {
		// Atomic write: build in a temp file in the same directory, then rename.
		dir := filepath.Dir(exportOutput)
		if dir == "" {
			dir = "."
		}
		tmpFile, err := os.CreateTemp(dir, ".export-*.tmp")
		if err != nil {
			return fmt.Errorf("create temp export: %w", err)
		}
		tmpName := tmpFile.Name()
		defer os.Remove(tmpName)

		if _, err := tmpFile.Write(buf.Bytes()); err != nil {
			tmpFile.Close()
			return fmt.Errorf("write temp export: %w", err)
		}
		if err := tmpFile.Sync(); err != nil {
			tmpFile.Close()
			return fmt.Errorf("sync temp export: %w", err)
		}
		if err := tmpFile.Close(); err != nil {
			return fmt.Errorf("close temp export: %w", err)
		}
		if err := os.Rename(tmpName, exportOutput); err != nil {
			return fmt.Errorf("atomic rename export: %w", err)
		}
	} else {
		fmt.Print(buf.String())
	}

	lpath, _ := common.LogPath()
	_ = audit.Append(lpath, audit.OpExport, "-", ns)

	if exportOutput != "" {
		fmt.Printf("Exported %d keys from [%s] to %s\n", len(keys), ns, exportOutput)
	} else {
		fmt.Fprintf(os.Stderr, "Exported %d keys from [%s]\n", len(keys), ns)
	}
	return nil
}
