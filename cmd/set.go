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
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"svault/internal/audit"
	"svault/internal/storage"
)

var setNamespace string
var setQuiet bool
var setStdin bool
var setClip bool

// setInput is the reader used for --stdin. Overridable in tests.
var setInput io.Reader = os.Stdin

var setCmd = &cobra.Command{
	Use:   "set KEY VALUE  |  set KEY=VALUE  |  set KEY --stdin",
	Short: "Save a secret to the vault",
	Long: `Save a secret into the active namespace.

The value can come from an argument, stdin, or the clipboard. Reading from stdin
or the clipboard keeps the value out of your shell history.

Examples:
  svault set DB_PASSWORD supersecret        # KEY VALUE form
  svault set API_KEY=sk-123                  # KEY=VALUE form
  echo "supersecret" | svault set DB_PASSWORD --stdin
  svault set API_KEY --clip                  # read from the clipboard`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runSet,
}

func init() {
	setCmd.Flags().StringVar(&setNamespace, "ns", "default", "target namespace (overrides auto-detect)")
	setCmd.Flags().BoolVarP(&setQuiet, "quiet", "q", false, "suppress output")
	setCmd.Flags().BoolVar(&setStdin, "stdin", false, "read value from stdin (keeps it out of shell history)")
	setCmd.Flags().BoolVar(&setClip, "clip", false, "read value from the clipboard")
	rootCmd.AddCommand(setCmd)
}

func runSet(cmd *cobra.Command, args []string) error {
	if setStdin && setClip {
		return fmt.Errorf("use only one of --stdin or --clip")
	}

	var key, value string
	if setClip {
		key = args[0]
		if strings.ContainsRune(key, '=') {
			return fmt.Errorf("with --clip, pass only the KEY, not KEY=VALUE")
		}
		clip, err := readClipboard()
		if err != nil {
			return err
		}
		value = clip
	} else if setStdin {
		key = args[0]
		if strings.ContainsRune(key, '=') {
			return fmt.Errorf("with --stdin, pass only the KEY, not KEY=VALUE")
		}
		data, err := io.ReadAll(setInput)
		if err != nil {
			return fmt.Errorf("read stdin: %w", err)
		}
		// Strip a single trailing newline so 'echo x | svault set K --stdin' is clean.
		value = strings.TrimSuffix(string(data), "\n")
	} else if len(args) == 1 {
		idx := strings.IndexByte(args[0], '=')
		if idx < 1 {
			return fmt.Errorf("invalid format: use KEY=VALUE or KEY VALUE")
		}
		key, value = args[0][:idx], args[0][idx+1:]
	} else {
		key, value = args[0], args[1]
	}
	if err := validateKey(key); err != nil {
		return err
	}
	ns := activeNamespace(cmd, setNamespace)

	err := mutateVault(func(vd *storage.VaultData) error {
		if vd.Namespaces[ns] == nil {
			vd.Namespaces[ns] = make(map[string]string)
		}
		vd.Namespaces[ns][key] = value
		return nil
	})
	if err != nil {
		return err
	}

	lpath, _ := logPath()
	_ = audit.Append(lpath, audit.OpSet, key, ns)

	if !setQuiet {
		fmt.Printf("OK  %s saved\n", key)
	}
	return nil
}
