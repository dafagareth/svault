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
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"svault/internal/storage"
)

var envNamespace string
var envFormat string

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Print secrets in shell, dotenv, json, or yaml format",
	Long: `Print all secrets in the active namespace in the chosen format.

Formats:
  shell   export KEY=VALUE   (default, for: eval $(svault env))
  dotenv  KEY=VALUE          (for writing a .env file)
  json    {"KEY":"VALUE"}
  yaml    KEY: "VALUE"

Examples:
  eval $(svault env)
  svault env -f dotenv > .env
  svault env -f json --ns production`,
	RunE: runEnv,
}

func init() {
	envCmd.Flags().StringVar(&envNamespace, "ns", "default", "source namespace")
	envCmd.Flags().StringVarP(&envFormat, "format", "f", "shell", "output format: shell, dotenv, json, yaml")
	rootCmd.AddCommand(envCmd)
}

func runEnv(cmd *cobra.Command, _ []string) error {
	ns := activeNamespace(cmd, envNamespace)

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

	secrets := vd.Namespaces[ns]
	keys := make([]string, 0, len(secrets))
	for k := range secrets {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	switch envFormat {
	case "shell":
		for _, k := range keys {
			fmt.Printf("export %s=%s\n", k, shellQuote(secrets[k]))
		}
	case "dotenv":
		for _, k := range keys {
			fmt.Printf("%s=%s\n", k, secrets[k])
		}
	case "json":
		m := make(map[string]string, len(keys))
		for _, k := range keys {
			m[k] = secrets[k]
		}
		out, _ := json.MarshalIndent(m, "", "  ")
		fmt.Println(string(out))
	case "yaml":
		for _, k := range keys {
			fmt.Printf("%s: %s\n", k, yamlQuote(secrets[k]))
		}
	default:
		return fmt.Errorf("unknown format %q (use: shell, dotenv, json, yaml)", envFormat)
	}
	return nil
}

// shellQuote wraps a value in single quotes, escaping any embedded single quotes,
// so it is safe to eval in a POSIX shell.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// yamlQuote wraps a value in double quotes, escaping backslashes and double
// quotes, producing a safe YAML scalar.
func yamlQuote(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return `"` + s + `"`
}
