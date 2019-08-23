package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var eksListContextsCmd = &cobra.Command{
	Use:   "list-kube-contexts",
	Short: "List Kubernetes contexts",
	Long:  "Lists the currently available kubernetes contexts",
	Run: func(cmd *cobra.Command, args []string) {

		contexts := GetKubeConfigContexts()
		for _, c := range contexts {
			fmt.Printf("%s\n", c.Name)
		}

	},
	Args: cobra.NoArgs,
}

func init() {
	eksCmd.AddCommand(eksListContextsCmd)
}
