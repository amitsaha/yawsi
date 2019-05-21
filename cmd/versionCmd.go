package cmd

import (
  "fmt"

  "github.com/spf13/cobra"
)

func init() {
  RootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
  Use:   "version",
  Short: "Print the version number of yawsi",
  Long:  `All software has versions. This is Hugo's`,
  Run: func(cmd *cobra.Command, args []string) {
    fmt.Println("Yet Another AWS CLI v0.1.1")
  },
}
