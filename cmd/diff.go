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

	"github.com/spf13/cobra"
	"svault/internal/storage"
)

var diffCmd = &cobra.Command{
	Use:   "diff NS1 NS2",
	Short: "Compare the contents of two namespaces",
	Args:  cobra.ExactArgs(2),
	RunE:  runDiff,
}

func init() {
	rootCmd.AddCommand(diffCmd)
}

func runDiff(_ *cobra.Command, args []string) error {
	ns1Name, ns2Name := args[0], args[1]

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

	ns1 := vd.Namespaces[ns1Name]
	ns2 := vd.Namespaces[ns2Name]

	allKeys := make(map[string]struct{})
	for k := range ns1 {
		allKeys[k] = struct{}{}
	}
	for k := range ns2 {
		allKeys[k] = struct{}{}
	}

	if len(allKeys) == 0 {
		fmt.Printf("Both namespaces [%s] and [%s] are empty\n", ns1Name, ns2Name)
		return nil
	}

	keys := make([]string, 0, len(allKeys))
	for k := range allKeys {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	same, diff := 0, 0
	for _, k := range keys {
		v1, in1 := ns1[k]
		v2, in2 := ns2[k]
		switch {
		case in1 && !in2:
			fmt.Printf("< %-30s  only in [%s]\n", k, ns1Name)
			diff++
		case !in1 && in2:
			fmt.Printf("> %-30s  only in [%s]\n", k, ns2Name)
			diff++
		case v1 != v2:
			fmt.Printf("~ %-30s  value differs\n", k)
			diff++
		default:
			fmt.Printf("= %-30s  same\n", k)
			same++
		}
	}

	fmt.Printf("\n[%s] vs [%s]: %d same, %d differ\n", ns1Name, ns2Name, same, diff)
	return nil
}
