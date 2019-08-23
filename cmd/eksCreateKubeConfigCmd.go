package cmd

import (
	"fmt"
	"log"
	"net/url"

	"github.com/spf13/cobra"
)

var eksCreateKubeConfigCmd = &cobra.Command{
	Use:   "create-kube-config",
	Short: "Create/update kubectl configuration",
	Long: `Create/update kubectl configuration:

	For direct AWS users:

	    yawsi eks create-kube-config --cluster-name k8s-cluster-non-production
	
	For project teams:

	    yawsi eks create-kube-config --cluster-name k8s-cluster-non-production --project projectA --environment qa
	
	`,
	Run: func(cmd *cobra.Command, args []string) {

		if len(clusterName) == 0 {
			log.Fatal("Please specify cluster name")
		}
		clusterData := DescribeEKSCluster(&clusterName)
		if clusterData == nil {
			log.Fatal("Couldn't find cluster data")
		}

		err := WriteKubeConfigToFile(clusterData, projectName, projectEnvironment)
		if err != nil {
			log.Fatal(err)
		}

		if showHostsFileEntry {
			fmt.Printf("\n\n--------------------------/etc/hosts/ file entry ---------------------\n\n")
			u, err := url.Parse(*clusterData.Cluster.Endpoint)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("%s %s\n\n", *GetPrivateMasterIP(&clusterName), u.Hostname())
		}
	},
	Args: cobra.NoArgs,
}

var showHostsFileEntry bool
var clusterName string
var projectName string
var projectEnvironment string

func init() {
	eksCmd.AddCommand(eksCreateKubeConfigCmd)
	eksCreateKubeConfigCmd.Flags().StringVarP(&clusterName, "cluster-name", "", "", "Cluster name to create context for")
	eksCreateKubeConfigCmd.Flags().StringVarP(&projectName, "project", "", "", "Project name to create context for")
	eksCreateKubeConfigCmd.Flags().StringVarP(&projectEnvironment, "environment", "", "", "Project environment to create context for (qa, staging, production)")
	eksCreateKubeConfigCmd.Flags().BoolVarP(&showHostsFileEntry, "show-hosts-file-entry", "", true, "Show /etc/hosts file entry for private clusters")
}
