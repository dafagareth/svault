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
	"sort"

	"github.com/spf13/cobra"
	"svault/internal/audit"
	"svault/internal/storage"
)

var exportNamespace string
var exportOutput string

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export secrets to .env format",
	RunE:  runExport,
}

func init() {
	exportCmd.Flags().StringVar(&exportNamespace, "ns", "default", "source namespace")
	exportCmd.Flags().StringVarP(&exportOutput, "output", "o", "", "write to file instead of stdout")
	rootCmd.AddCommand(exportCmd)
}

func runExport(cmd *cobra.Command, _ []string) error {
	ns := activeNamespace(cmd, exportNamespace)

	vpath, err := vaultPath()
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

	out := os.Stdout
	if exportOutput != "" {
		f, err := os.Create(exportOutput)
		if err != nil {
			return fmt.Errorf("create file: %w", err)
		}
		defer f.Close()
		out = f
	}

	for _, k := range keys {
		fmt.Fprintf(out, "%s=%s\n", k, secrets[k])
	}

	lpath, _ := logPath()
	_ = audit.Append(lpath, audit.OpExport, "-", ns)

	if exportOutput != "" {
		fmt.Printf("Exported %d keys from [%s] to %s\n", len(keys), ns, exportOutput)
	} else {
		fmt.Fprintf(os.Stderr, "Exported %d keys from [%s]\n", len(keys), ns)
	}
	return nil
}
