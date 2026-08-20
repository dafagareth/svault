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
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"svault/internal/storage"
)

var execNamespace string

var execCmd = &cobra.Command{
	Use:   "exec -- COMMAND [ARGS...]",
	Short: "Run a command with secrets injected as environment variables",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runExecCmd,
}

func init() {
	execCmd.Flags().StringVar(&execNamespace, "ns", "default", "source namespace")
	rootCmd.AddCommand(execCmd)
}

func runExecCmd(cmd *cobra.Command, args []string) error {
	ns := activeNamespace(cmd, execNamespace)

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

	extra := make([]string, 0, len(vd.Namespaces[ns]))
	for k, v := range vd.Namespaces[ns] {
		extra = append(extra, k+"="+v)
	}

	c := exec.Command(args[0], args[1:]...)
	c.Env = append(os.Environ(), extra...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	if err := c.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.ExitCode())
		}
		return fmt.Errorf("exec: %w", err)
	}
	return nil
}
