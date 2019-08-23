package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var eksListCmd = &cobra.Command{
	Use:   "list-clusters",
	Short: "List EKS clusters",
	Long:  "List the current AWS EKS clusters",
	Run: func(cmd *cobra.Command, args []string) {

		result := ListEKSClusters(details)
		fmt.Printf("%#v\n", result)

	},
	Args: cobra.NoArgs,
}

var details bool

func init() {
	eksCmd.AddCommand(eksListCmd)
	eksListCmd.Flags().BoolVarP(&details, "details", "", false, "Show cluster details")
}
