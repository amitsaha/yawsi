package cmd

import "github.com/spf13/cobra"
import "os"

// completionCmd represents the completion command
var bashCompletionCmd = &cobra.Command{
	Use:   "generate-bash-completion",
	Short: "Generates bash completion scripts",
	Long: `To load completion run

. <(yawsi generate-bash-completion)

To configure your bash shell to load completions for each session add to your bashrc

# ~/.bashrc or ~/.profile
. <(yawsi generate-bash-completion)
`,
	Run: func(cmd *cobra.Command, args []string) {
		RootCmd.GenBashCompletion(os.Stdout)
	},
}

func init() {
    RootCmd.AddCommand(bashCompletionCmd)
}
