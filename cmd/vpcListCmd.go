// Copyright © 2018 Amit Saha <amitsaha.in@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/spf13/cobra"
)

func listVpcDetails(vpcs *ec2.DescribeVpcsOutput) {
	var vpcName string

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 35, 0, 1, ' ', 0)

	fmt.Fprintln(w, "Name  \tVPCID  \tCIDRBlock\tDefault?\tTags\t")
	fmt.Fprintln(w, "-----\t----------\t--------\t------\t---------\t")
	for _, v := range vpcs.Vpcs {
		for _, tag := range v.Tags {
			if *tag.Key == "Name" {
				vpcName = *tag.Value
			}
		}

		fmt.Fprintf(w, "%s\t", vpcName)
		fmt.Fprintf(w, "%s\t", *v.VpcId)
		fmt.Fprintf(w, "%s\t", *v.CidrBlock)
		fmt.Fprintf(w, "%v\t", *v.IsDefault)
		fmt.Fprintf(w, "%s\t\n", getTagsAsString(v.Tags, " "))
	}
	fmt.Fprintln(w)
	w.Flush()
}

// listVpcsCmd represents the list vpc command
var listVpcsCmd = &cobra.Command{
	Use:   "list",
	Short: "List VPCs",
	Long: `List the VPCs and other details
	
	To list the VPCs alongwith some key information:

		$ yawsi vpc  list
		
	To show further details in an interactive window:

	    $ yawsi vpc list --details
	
	`,
	Run: func(cmd *cobra.Command, args []string) {

		if !vpcDetails {
			vpcs := getVpcs()
			listVpcDetails(vpcs)
		} else {
			displayVPCDetails()
		}
	},
}

var vpcDetails bool

func init() {
	vpcCmd.AddCommand(listVpcsCmd)
	listVpcsCmd.Flags().BoolVarP(&vpcDetails, "details", "", false, "Show VPC details")
}
