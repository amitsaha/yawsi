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
	"github.com/aws/aws-sdk-go/aws" //"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/spf13/cobra"
	"log"
	"os"
	"strings"
	"text/tabwriter"
)

func displayRoutingTables(routes []*RouteContainer) {

	sess := createSession()
	svc := ec2.New(sess)

	w := new(tabwriter.Writer)

	// Format in tab-separated columns with a tab stop of 8.
	w.Init(os.Stdout, 0, 40, 0, '\t', tabwriter.AlignRight)
	fmt.Fprintln(w, "RoutTableID\tMain\tDestination\tTarget\t")
	fmt.Fprintln(w, "-----------\t----\t------------\t------\t")

	for _, routeTable := range routes {

		for _, route := range routeTable.Routes {
			fmt.Fprintf(w, "%v\t", routeTable.RouteTableId)
			fmt.Fprintf(w, "%v\t", routeTable.Main)

			if route.DestinationCidrBlock != nil {
				fmt.Fprintf(w, "%s\t", *route.DestinationCidrBlock)
			}

			if route.DestinationPrefixListId != nil {
				fmt.Fprintf(w, "%s", *route.DestinationPrefixListId)
				input := &ec2.DescribePrefixListsInput{
					PrefixListIds: []*string{route.DestinationPrefixListId},
				}
				result, err := svc.DescribePrefixLists(input)
				if err != nil {
					log.Fatal(err)
				}

				if len(result.PrefixLists) != 1 {
					panic("Expected the prefix list query to return a result")
				}

				prefixListServices := map[string]string{
					"s3":                   "S3",
					"dynamodb":             "DynamoDB",
					"ec2":                  "EC2",
					"ec2messages":          "EC2 Messages",
					"elasticloadbalancing": "ELB API",
					"kinesis":              "Kinesis",
					"ssm":                  "SSM",
				}
				for k, v := range prefixListServices {
					if strings.Contains(*result.PrefixLists[0].PrefixListName, k) {
						fmt.Fprintf(w, "(%s)\t", v)
					}
				}

			}

			if route.GatewayId != nil && len(*route.GatewayId) != 0 {
				fmt.Fprintf(w, "%s\t\n", *route.GatewayId)
				// TODO: Add more details on the gateway as required
			}
			if route.NatGatewayId != nil && len(*route.NatGatewayId) != 0 {
				fmt.Fprintf(w, "%s\t\n", *route.NatGatewayId)
			}
			// NAT instance - this means we likely have both the instance ID
			// and network interface ID set corresponding to the ENI interface
			if route.InstanceId != nil && len(*route.InstanceId) != 0 {
				fmt.Fprintf(w, "%s - ", *route.InstanceId)
				if route.NetworkInterfaceId != nil && len(*route.NetworkInterfaceId) != 0 {
					fmt.Fprintf(w, "%s\t\n", *route.NetworkInterfaceId)
				} else {
					fmt.Fprintf(w, " (No ENI)\t\n")
				}
			}
			if route.VpcPeeringConnectionId != nil && len(*route.VpcPeeringConnectionId) != 0 {

				input := &ec2.DescribeVpcPeeringConnectionsInput{
					VpcPeeringConnectionIds: []*string{route.VpcPeeringConnectionId},
				}

				result, err := svc.DescribeVpcPeeringConnections(input)
				if err != nil {
					if aerr, ok := err.(awserr.Error); ok {
						switch aerr.Code() {
						default:
							fmt.Println(aerr.Error())
						}
					} else {
						// Print the error, cast err to awserr.Error to get the Code and
						// Message from an error.
						fmt.Println(err.Error())
					}
				}

				if result == nil {
					fmt.Println("Could not retrieve VPC peering connection details")

				}
				vpcId := *result.VpcPeeringConnections[0].AccepterVpcInfo.VpcId
				input1 := &ec2.DescribeVpcsInput{
					VpcIds: []*string{
						aws.String(vpcId),
					},
				}

				result1, err := svc.DescribeVpcs(input1)
				if err != nil {
					if aerr, ok := err.(awserr.Error); ok {
						switch aerr.Code() {
						default:
							fmt.Println(aerr.Error())
						}
					} else {
						// Print the error, cast err to awserr.Error to get the Code and
						// Message from an error.
						fmt.Println(err.Error())
					}
					return
				}

				vpcName := ""
				for _, tag := range result1.Vpcs[0].Tags {
					if *tag.Key == "Name" {
						vpcName = *tag.Value
					}
				}
				fmt.Fprintf(w, "%s (%s - %s) \t\n", *route.VpcPeeringConnectionId, vpcId, vpcName)
			}
		}
	}

	fmt.Fprintln(w)
	w.Flush()

}

var inspectRoutingTablesInstancesCmd = &cobra.Command{
	Use:   "routing-tables",
	Short: "Routing table entries associated with an instance",
	Long: `Show the routing table associated with an EC2 instance:

	$ yawsi ec2  inspect routing-tables i-06d80024e0df241da

	RoutTableID     Main    Destination     Target
	-----------     ----    ------------    ------
	rtb-d1df42b5    false   172.31.0.0/16   local
	rtb-d24342b5    false   pl-6ca54005(S3) vpce-6b2ecf02
	rtb-942315f1    true    172.31.0.0/16   pcx-cd9541a4 (vpc-20988a4 - VPCA)
	rtb-63caa9f1    true    0.0.0.0/0       igw-121234
	`,
	Run: func(cmd *cobra.Command, args []string) {
		//var ec2Filters []*ec2.Filter
		var inputInstanceIds []*string

		inputInstanceIds = append(inputInstanceIds, &args[0])
		instanceData := getEC2InstanceData(nil, inputInstanceIds...)

		if len(instanceData) != 1 {
			panic("Expected ")
		}
		instanceState := instanceData[0]

		routes := getRoutes(instanceState.SubnetIds...)
		displayRoutingTables(routes)
	},
	Args: cobra.ExactArgs(1),
}

//var publicIngress bool
//var publicEgress bool
//var public bool

func init() {
	inspectInstancesCmd.AddCommand(inspectRoutingTablesInstancesCmd)
}
