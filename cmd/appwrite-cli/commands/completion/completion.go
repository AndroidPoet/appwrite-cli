package completion

import (
	"os"

	"github.com/spf13/cobra"
)

// CompletionCmd generates shell completions
var CompletionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion scripts",
	Long: `Generate shell completion scripts for appwrite-cli.

Examples:
  # Bash
  aw completion bash > /usr/local/etc/bash_completion.d/aw

  # Zsh
  aw completion zsh > "${fpath[1]}/_aw"

  # Fish
  aw completion fish > ~/.config/fish/completions/aw.fish

  # PowerShell
  aw completion powershell > aw.ps1`,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	DisableFlagsInUseLine: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			return cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			return cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
		}
		return nil
	},
}
