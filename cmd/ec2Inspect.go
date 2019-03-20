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
	//"fmt"

	// "github.com/aws/aws-sdk-go/aws/awserr"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

func checkPublicIngress(state *instanceState) *checkResult {
	result := newCheckResult()
	result.DisplayText = "Checking if the instance has a Public IP address"
	result.Result = len(state.PublicIP) != 0
	result.Metadata["PublicIP"] = state.PublicIP

	return &result
}

func checkPublicEgress(state *instanceState) *checkResult {
	result := newCheckResult()
	result.DisplayText = "Checking for a route to 0.0.0.0/0"
	metadata := make(map[string]interface{})

	for _, routeContainer := range state.Routes {
		for _, route := range routeContainer.Routes {
			if route.DestinationCidrBlock != nil && *route.DestinationCidrBlock == "0.0.0.0/0" {
				if route.GatewayId != nil && len(*route.GatewayId) != 0 {
					metadata["GatewayId"] = *route.GatewayId
				}
				if route.NatGatewayId != nil && len(*route.NatGatewayId) != 0 {
					metadata["NatGatewayId"] = *route.NatGatewayId
				}
				if route.InstanceId != nil && len(*route.InstanceId) != 0 {
					metadata["InstanceId"] = *route.InstanceId
				}
				if route.NetworkInterfaceId != nil && len(*route.NetworkInterfaceId) != 0 {
					metadata["NetworkInterfaceId"] = *route.NetworkInterfaceId
				}
				if route.VpcPeeringConnectionId != nil && len(*route.VpcPeeringConnectionId) != 0 {
					metadata["VpcPeeringConnectionId"] = *route.VpcPeeringConnectionId
				}
			}
		}
	}

	if len(metadata) != 0 {
		result.Result = true
		result.Metadata = metadata
	}
	return &result
}

func checkPublic(state *instanceState) *checkResult {

	ingressResult := checkPublicIngress(state)
	egressResult := checkPublicEgress(state)

	result := newCheckResult()
	result.Result = true

	if ingressResult.Result && egressResult.Result {
		result.DisplayText = "Instance can initiate connection with the outside world and vice-versa"
		result.Result = true
	} else {
		if !ingressResult.Result {
			result.DisplayText = "Outside world cannot initiate connection with the instance."
		}
		if !egressResult.Result {
			result.DisplayText = result.DisplayText + "Instance cannot initiate connection with the outside world."
		}
		result.Result = false
	}

	result.Metadata = ingressResult.Metadata
	for k, v := range egressResult.Metadata {
		if result.Metadata[k] != nil && len(result.Metadata[k].(string)) != 0 {
			log.Fatal("Duplicate key, Fix it!")
		}
		result.Metadata[k] = v
	}

	return &result
}

var inspectInstancesCmd = &cobra.Command{
	Use:   "inspect",
	Short: "Perform various checks",
	Long: `Perform various checks on EC2 instances.

Check if an instance can send traffic to the outside world:

	yawsi ec2  inspect i-06d80024e0df241da --public-egress --verbose
	✔ Checking for a route to 0.0.0.0/0
	map[InstanceId:i-0685cbd9 NetworkInterfaceId:eni-ed43078a]
	true

	yawsi ec2 inspect i-06d80024e0df241da --public-egress
	true

Check if the outside world can initiate communication with the EC2 instance:

	yawsi ec2  inspect i-06d80024e0df241da --public-ingress
	false

	yawsi ec2  inspect i-06d80024e0df241da --public-ingress --verbose
	✖ Checking if the instance has a Public IP address
	false

Check if the outside world can initiate communication with a EC2 instance and vice-versa:

	yawsi ec2  inspect i-06d80024e0df241da --public
	false

	yawsi.exe ec2  inspect i-06d80024e0df241da --public --verbose
	✖ Outside world cannot initiate connection with the instance.
	false
	`,
	Run: func(cmd *cobra.Command, args []string) {
		var inputInstanceIds []*string

		if !(publicIngress || publicEgress || public) {
			cmd.Help()
			os.Exit(1)
		}

		inputInstanceIds = append(inputInstanceIds, &args[0])
		instanceData := getEC2InstanceData(nil, inputInstanceIds...)

		if len(instanceData) != 1 {
			panic("Couldn't retrieve instance data")
		}
		instanceData[0].Routes = getRoutes(instanceData[0].SubnetIds...)

		var result *checkResult

		if publicIngress {
			result = checkPublicIngress(instanceData[0])
		}

		if publicEgress {
			result = checkPublicEgress(instanceData[0])
		}

		if public {
			result = checkPublic(instanceData[0])
		}

		displayResult(result)
		fmt.Printf("%v\n", summarizeResults(result))
	},
	Args: cobra.ExactArgs(1),
}

var publicIngress bool
var publicEgress bool

var public bool

func init() {
	ec2Cmd.AddCommand(inspectInstancesCmd)
	inspectInstancesCmd.Flags().BoolVarP(&publicIngress, "public-ingress", "", false, "Am I visible to the outside world (do I have a public IP)?")
	inspectInstancesCmd.Flags().BoolVarP(&publicEgress, "public-egress", "", false, "Can I see the outside world?")
	inspectInstancesCmd.Flags().BoolVarP(&public, "public", "", false, "Can the outside world see me and vice-versa?")
	inspectInstancesCmd.Flags().BoolVarP(&verboseOutput, "verbose", "v", false, "Display more information about the result")
	inspectInstancesCmd.Flags().BoolVarP(&debugOutput, "debug", "", false, "Display debugging information about the result")
}
