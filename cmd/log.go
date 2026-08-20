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
	"strings"

	"github.com/spf13/cobra"
)

var logTail int

var logCmd = &cobra.Command{
	Use:   "log",
	Short: "Show the vault audit log",
	RunE:  runLog,
}

func init() {
	logCmd.Flags().IntVarP(&logTail, "tail", "n", 0, "show last N entries (0 = all)")
	rootCmd.AddCommand(logCmd)
}

func runLog(_ *cobra.Command, _ []string) error {
	lpath, err := logPath()
	if err != nil {
		return err
	}
	data, err := os.ReadFile(lpath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("(empty)")
			return nil
		}
		return fmt.Errorf("read log: %w", err)
	}
	if len(data) == 0 {
		fmt.Println("(empty)")
		return nil
	}
	if logTail > 0 {
		lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
		if logTail < len(lines) {
			lines = lines[len(lines)-logTail:]
		}
		fmt.Println(strings.Join(lines, "\n"))
		return nil
	}
	fmt.Print(string(data))
	return nil
}
