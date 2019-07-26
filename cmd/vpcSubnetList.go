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

	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/spf13/cobra"

	"text/tabwriter"
)

func getSubnetType(subnetID *string) string {
	routes := getRoutes(*subnetID)
	for _, route := range routes {
		for _, r := range route.Routes {
			if r.DestinationCidrBlock != nil && *r.DestinationCidrBlock == "0.0.0.0/0" {
				if r.GatewayId != nil && len(*r.GatewayId) != 0 {
					return "public"
				}

				if r.NetworkInterfaceId != nil && len(*r.NetworkInterfaceId) > 0 && strings.HasPrefix(*r.NetworkInterfaceId, "eni-") {
					// TODO: would be nice to return the NAT instance ID?
					return "private"
				}
			}
		}
	}
	// default to private considering the absence of a route to 0.0.0.0/0 CIDR
	return "private"
}

func getSubnetName(tags []*ec2.Tag) string {
	for _, tag := range tags {
		if *tag.Key == "Name" {
			return *tag.Value
		}
	}
	return "Subnet Name"
}

func displaySubnetDetails(subnets []*ec2.Subnet) {
	w := new(tabwriter.Writer)

	// Format in tab-separated columns with a tab stop of 8.
	//w.Init(os.Stdout, 0, 40, 0, '\t', tabwriter.AlignRight)
	w.Init(os.Stdout, 35, 0, 1, ' ', 0)
	fmt.Fprintln(w, "Name                  \tSubnetID  \tCIDRBlock\tSubnetType\tTags\t")
	fmt.Fprintln(w, "---------------------\t----------\t--------\t---------\t------\t")
	for _, subnet := range subnets {
		fmt.Fprintf(w, "%s\t", getSubnetName(subnet.Tags))
		fmt.Fprintf(w, "%s\t", *subnet.SubnetId)
		fmt.Fprintf(w, "%s\t", *subnet.CidrBlock)
		fmt.Fprintf(w, "%s\t", getSubnetType(subnet.SubnetId))
		fmt.Fprintf(w, "%s\t\n", getTagsAsString(subnet.Tags, " "))
	}
	fmt.Fprintln(w)
	w.Flush()
}

// listAsgCmd represents the listAsg command
var listSubnetsCmd = &cobra.Command{
	Use:   "list-subnets",
	Short: "List Subnets in a VPC",
	Long: `List the subnets in a specific vpc
	
	To list the subnets in a VPC selected interactively:

		$ yawsi vpc list-subnets
		
	To list subnets in a specific VPC:

	    $ yawsi vpc list-subnets --vpc-id <vpc-id>	
	`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(vpcId) == 0 {
			vpcId = selectVPCInteractive()
		}
		input := &ec2.DescribeSubnetsInput{
			Filters: []*ec2.Filter{
				{
					Name: aws.String("vpc-id"),
					Values: []*string{
						aws.String(vpcId),
					},
				},
			},
		}

		subnets := getSubnets(input)
		displaySubnetDetails(subnets)
	},
}

var vpcId string

func init() {
	vpcCmd.AddCommand(listSubnetsCmd)
	listSubnetsCmd.Flags().StringVarP(&vpcId, "vpc-id", "", "", "List subnets in a specific VPC")
}
