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
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"svault/internal/storage"
)

var listNamespace string
var listAll bool
var listValues bool
var listMask bool
var listFilter string
var listCount bool
var listJSON bool
var listLong bool

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all keys in the active namespace",
	Long: `List the keys in the active namespace.

Boolean flags combine like 'ls', for example 'svault list -lm' or 'svault list -ac'.

Examples:
  svault list                # key names only
  svault list -v             # keys with values
  svault list -l             # long format: key and value length (no values)
  svault list -l -v          # long format including values
  svault list -m             # masked values (****), safe for screen sharing
  svault list -F DB          # filter keys containing "DB" (case-insensitive)
  svault list -c             # count of keys only
  svault list -a             # every namespace
  svault list --json         # machine-readable output (add -v for KEY:VALUE)`,
	RunE: runList,
}

func init() {
	listCmd.Flags().StringVar(&listNamespace, "ns", "default", "namespace to list")
	listCmd.Flags().BoolVarP(&listAll, "all", "a", false, "list keys from all namespaces")
	listCmd.Flags().BoolVarP(&listValues, "values", "v", false, "show values alongside keys")
	listCmd.Flags().BoolVarP(&listLong, "long", "l", false, "long format: key and value length (add -v to reveal values)")
	listCmd.Flags().BoolVarP(&listMask, "mask", "m", false, "show masked values (****)")
	listCmd.Flags().StringVarP(&listFilter, "filter", "F", "", "filter keys by substring (case-insensitive)")
	listCmd.Flags().BoolVarP(&listCount, "count", "c", false, "show only the count of keys")
	listCmd.Flags().BoolVar(&listJSON, "json", false, "output as JSON")
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, _ []string) error {
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

	if listAll {
		nsNames := make([]string, 0, len(vd.Namespaces))
		for n := range vd.Namespaces {
			nsNames = append(nsNames, n)
		}
		sort.Strings(nsNames)

		if listJSON {
			return printAllJSON(vd.Namespaces, nsNames)
		}

		for _, n := range nsNames {
			keys := filteredKeys(vd.Namespaces[n], listFilter)
			if listCount {
				fmt.Printf("[%s]  (%d keys)\n", n, len(keys))
				continue
			}
			fmt.Printf("[%s]\n", n)
			printKeyList(keys, vd.Namespaces[n], "  ")
		}
		return nil
	}

	nsName := activeNamespace(cmd, listNamespace)
	secrets := vd.Namespaces[nsName]
	keys := filteredKeys(secrets, listFilter)

	if listJSON {
		return printNsJSON(secrets, keys)
	}

	if listCount {
		fmt.Println(len(keys))
		return nil
	}

	if len(keys) == 0 {
		fmt.Println("(empty)")
		return nil
	}

	printKeyList(keys, secrets, "")
	return nil
}

func filteredKeys(secrets map[string]string, filter string) []string {
	f := strings.ToLower(filter)
	keys := make([]string, 0, len(secrets))
	for k := range secrets {
		if f == "" || strings.Contains(strings.ToLower(k), f) {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	return keys
}

func printKeyList(keys []string, secrets map[string]string, indent string) {
	if !listValues && !listMask && !listLong {
		for _, k := range keys {
			fmt.Printf("%s%s\n", indent, k)
		}
		return
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	for _, k := range keys {
		val := secrets[k]
		if listLong {
			// Long format shows the value length, never the value itself,
			// unless the value is explicitly requested (-v) or masked (-m).
			if listValues || listMask {
				displayed := val
				if listMask {
					displayed = "****"
				}
				fmt.Fprintf(w, "%s%s\t%d\t%s\n", indent, k, len(val), displayed)
			} else {
				fmt.Fprintf(w, "%s%s\t%d\n", indent, k, len(val))
			}
			continue
		}
		if listMask {
			val = "****"
		}
		fmt.Fprintf(w, "%s%s\t%s\n", indent, k, val)
	}
	w.Flush()
}

func printNsJSON(secrets map[string]string, keys []string) error {
	if listValues || listMask {
		m := make(map[string]string, len(keys))
		for _, k := range keys {
			if listMask {
				m[k] = "****"
			} else {
				m[k] = secrets[k]
			}
		}
		out, _ := json.MarshalIndent(m, "", "  ")
		fmt.Println(string(out))
		return nil
	}
	out, _ := json.MarshalIndent(keys, "", "  ")
	fmt.Println(string(out))
	return nil
}

func printAllJSON(namespaces map[string]map[string]string, nsNames []string) error {
	if listValues || listMask {
		result := make(map[string]map[string]string, len(nsNames))
		for _, n := range nsNames {
			keys := filteredKeys(namespaces[n], listFilter)
			m := make(map[string]string, len(keys))
			for _, k := range keys {
				if listMask {
					m[k] = "****"
				} else {
					m[k] = namespaces[n][k]
				}
			}
			result[n] = m
		}
		out, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(out))
		return nil
	}
	result := make(map[string][]string, len(nsNames))
	for _, n := range nsNames {
		result[n] = filteredKeys(namespaces[n], listFilter)
	}
	out, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(out))
	return nil
}
