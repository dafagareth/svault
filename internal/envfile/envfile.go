package envfile

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
)

type Entry struct {
	Key   string
	Value string
}

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
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		}
		idx := strings.IndexByte(line, '=')
		if idx < 1 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		if key == "" {
			continue
		}
		valueRaw := strings.TrimSpace(line[idx+1:])
		value := parseValue(valueRaw)
		entries = append(entries, Entry{Key: key, Value: value})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}
	return entries, nil
}

func parseValue(v string) string {
	if len(v) == 0 {
		return ""
	}
	if v[0] == '"' || v[0] == '\'' {
		quote := v[0]
		endIdx := -1
		for i := 1; i < len(v); i++ {
			if v[i] == quote && v[i-1] != '\\' {
				endIdx = i
				break
			}
		}
		if endIdx > 0 {
			res := v[1:endIdx]
			res = strings.ReplaceAll(res, "\\\"", "\"")
			res = strings.ReplaceAll(res, "\\'", "'")
			return res
		}
	}
	if hashIdx := strings.IndexByte(v, '#'); hashIdx >= 0 {
		v = strings.TrimSpace(v[:hashIdx])
	}
	return v
}

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
