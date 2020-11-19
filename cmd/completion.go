package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(completionCmd)
}

// TODO add fish, powershell, zsh support?
var completionCmd = &cobra.Command{
	Use:   "completion [bash]",
	Short: "Generate completion script",
	Long: `To load completions:

Bash:

$ source <(historian completion bash)

# To load completions for each session, execute once:
Linux:
  $ historian completion bash > /etc/bash_completion.d/historian
MacOS:
  $ historian completion bash > /usr/local/etc/bash_completion.d/historian
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.ExactValidArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			cmd.Root().GenPowerShellCompletion(os.Stdout)
		}
	},
}
