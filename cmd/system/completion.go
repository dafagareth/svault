package system

import (
	"os"

	"github.com/spf13/cobra"
)

var CompletionCmd = &cobra.Command{
	GroupID: "system",
	Use:     "completion [bash|zsh|fish|powershell]",
	Short:   "Generate shell completion script",
	Long: `Generate a shell completion script for svault.

Bash:
  svault completion bash | sudo tee /etc/bash_completion.d/svault > /dev/null

Zsh:
  svault completion zsh > "${fpath[1]}/_svault"

Fish:
  svault completion fish > ~/.config/fish/completions/svault.fish

PowerShell:
  svault completion powershell | Out-String | Invoke-Expression`,
	Args:      cobra.ExactArgs(1),
	ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
	RunE: func(cmd *cobra.Command, args []string) error {
		root := cmd.Root()
		switch args[0] {
		case "bash":
			return root.GenBashCompletionV2(os.Stdout, true)
		case "zsh":
			return root.GenZshCompletion(os.Stdout)
		case "fish":
			return root.GenFishCompletion(os.Stdout, true)
		case "powershell":
			return root.GenPowerShellCompletionWithDesc(os.Stdout)
		}
		return nil
	},
}
