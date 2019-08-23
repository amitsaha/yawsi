package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/ktr0731/go-fuzzyfinder"
)

var eksWorkonCmd = &cobra.Command{
	Use:   "workon",
	Short: "Set current kubernetes context",
	Long:  "Set current kubernetes context",
	Run: func(cmd *cobra.Command, args []string) {

		var kubeContextName string
		if len(args) != 1 {
			contexts := GetKubeConfigContexts()
			idx, err := fuzzyfinder.Find(
				contexts,
				func(i int) string {
					return contexts[i].Name
				},
			)
			if err != nil {
				log.Fatal(err)
			}
			kubeContextName = contexts[idx].Name
		} else {
			kubeContextName = args[0]
		}

		err := SetKubeConfigCurrentContext(kubeContextName)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Context set to %s\n", kubeContextName)

	},
}

func init() {
	eksCmd.AddCommand(eksWorkonCmd)
}
