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

	"github.com/spf13/cobra"
	"svault/internal/storage"
)

var useCmd = &cobra.Command{
	Use:   "use NAMESPACE",
	Short: "Set the active namespace",
	Args:  cobra.ExactArgs(1),
	RunE:  runUse,
}

func init() {
	rootCmd.AddCommand(useCmd)
}

func runUse(_ *cobra.Command, args []string) error {
	ns := args[0]

	dir, err := vaultDir()
	if err != nil {
		return err
	}
	cfg, err := storage.ReadConfig(dir)
	if err != nil {
		return err
	}
	cfg.ActiveNamespace = ns
	if err := storage.WriteConfig(dir, cfg); err != nil {
		return err
	}

	fmt.Printf("Switched to namespace: %s\n", ns)
	return nil
}
