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
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"svault/internal/storage"
)

var searchAllNamespaces bool

var searchCmd = &cobra.Command{
	Use:   "search PATTERN",
	Short: "Search keys by name (case-insensitive substring match)",
	Args:  cobra.ExactArgs(1),
	RunE:  runSearch,
}

func init() {
	searchCmd.Flags().BoolVar(&searchAllNamespaces, "all", false, "search across all namespaces")
	rootCmd.AddCommand(searchCmd)
}

func runSearch(cmd *cobra.Command, args []string) error {
	pattern := strings.ToLower(args[0])

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

	nsNames := []string{activeNamespace(cmd, "default")}
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
