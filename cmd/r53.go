package cmd

import "github.com/spf13/cobra"

var r53Cmd = &cobra.Command{
	Use:   "r53",
	Short: "Commands for working with AWS Route53",
}

func init() {
	RootCmd.AddCommand(r53Cmd)
}
