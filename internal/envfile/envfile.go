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

// Package envfile parses .env-style files: KEY=VALUE pairs, one per line, with
// support for comments and blank lines. It is shared by the import and check
// commands so the two stay in sync on what a valid line looks like.
package envfile

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
)

// Entry is a single parsed KEY=VALUE pair, in file order.
type Entry struct {
	Key   string
	Value string
}

// Parse reads an .env-style file and returns its entries in file order. It
// skips blank lines and lines beginning with '#', splits each remaining line on
// the first '=', trims surrounding whitespace from both sides, and strips a
// single layer of matching surrounding quotes from the value. Lines without a
// '=' or with an empty key are skipped.
func Parse(path string) ([]Entry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	var entries []Entry
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		idx := strings.IndexByte(line, '=')
		if idx < 1 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		if key == "" {
			continue
		}
		value := strings.TrimSpace(line[idx+1:])
		value = stripQuotes(value)
		entries = append(entries, Entry{Key: key, Value: value})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}
	return entries, nil
}

// Keys returns the sorted, de-duplicated key names from an .env-style file.
func Keys(path string) ([]string, error) {
	entries, err := Parse(path)
	if err != nil {
		return nil, err
	}
	seen := make(map[string]struct{}, len(entries))
	keys := make([]string, 0, len(entries))
	for _, e := range entries {
		if _, ok := seen[e.Key]; ok {
			continue
		}
		seen[e.Key] = struct{}{}
		keys = append(keys, e.Key)
	}
	sort.Strings(keys)
	return keys, nil
}

// stripQuotes removes a single pair of matching surrounding quotes, if present.
func stripQuotes(v string) string {
	if len(v) >= 2 {
		first, last := v[0], v[len(v)-1]
		if (first == '"' && last == '"') || (first == '\'' && last == '\'') {
			return v[1 : len(v)-1]
		}
	}
	return v
}
