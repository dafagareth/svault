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

package audit

import (
	"fmt"
	"os"
	"time"
)

type Op string

const (
	OpSet    Op = "SET"
	OpGet    Op = "GET"
	OpDelete Op = "DELETE"
	OpInit   Op = "INIT"
	OpUnlock Op = "UNLOCK"
	OpLock   Op = "LOCK"
	OpExport Op = "EXPORT"
	OpImport Op = "IMPORT"
	OpRotate Op = "ROTATE"
)

func Append(logPath string, op Op, key, namespace string) error {
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("open audit log: %w", err)
	}
	defer f.Close()
	ts := time.Now().Format("2006-01-02 15:04:05")
	if _, err := fmt.Fprintf(f, "%s  %-6s  %-20s  [%s]\n", ts, op, key, namespace); err != nil {
		return fmt.Errorf("write audit log: %w", err)
	}
	return nil
}
