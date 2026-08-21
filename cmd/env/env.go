package env

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"svault/cmd/common"

	"github.com/spf13/cobra"
	"svault/internal/storage"
)

var envNamespace string
var envFormat string

var EnvCmd = &cobra.Command{
	GroupID: "env",
	Use:     "env",
	Short:   "Print secrets in shell, dotenv, json, or yaml format",
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
	EnvCmd.Flags().StringVar(&envNamespace, "ns", "default", "source namespace")
	EnvCmd.Flags().StringVarP(&envFormat, "format", "f", "shell", "output format: shell, dotenv, json, yaml")
}

func runEnv(cmd *cobra.Command, _ []string) error {
	ns := common.ActiveNamespace(cmd, envNamespace)

	vpath, err := common.VaultPath()
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
