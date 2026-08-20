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
	"svault/internal/audit"
	"svault/internal/storage"
)

var (
	moveFromNS string
	moveToNS   string
)

var moveCmd = &cobra.Command{
	Use:   "move KEY --to NAMESPACE",
	Short: "Move a key from one namespace to another",
	Args:  cobra.ExactArgs(1),
	RunE:  runMove,
}

func init() {
	moveCmd.Flags().StringVar(&moveFromNS, "from", "default", "source namespace")
	moveCmd.Flags().StringVar(&moveToNS, "to", "", "destination namespace (required)")
	_ = moveCmd.MarkFlagRequired("to")
	rootCmd.AddCommand(moveCmd)
}

func runMove(cmd *cobra.Command, args []string) error {
	key := args[0]
	from := activeNamespace(cmd, moveFromNS)
	to := moveToNS

	if from == to {
		return fmt.Errorf("source and destination namespaces are the same: %s", from)
	}

	err := mutateVault(func(vd *storage.VaultData) error {
		val, ok := vd.Namespaces[from][key]
		if !ok {
			return fmt.Errorf("key not found in [%s]: %s", from, key)
		}
		if vd.Namespaces[to] == nil {
			vd.Namespaces[to] = make(map[string]string)
		}
		if _, exists := vd.Namespaces[to][key]; exists {
			return fmt.Errorf("key already exists in [%s]: %s", to, key)
		}
		vd.Namespaces[to][key] = val
		delete(vd.Namespaces[from], key)
		return nil
	})
	if err != nil {
		return err
	}

	lpath, _ := logPath()
	_ = audit.Append(lpath, audit.OpDelete, key, from)
	_ = audit.Append(lpath, audit.OpSet, key, to)

	fmt.Printf("Moved %s from [%s] to [%s]\n", key, from, to)
	return nil
}
