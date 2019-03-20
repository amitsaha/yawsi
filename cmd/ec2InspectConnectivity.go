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
	"github.com/aws/aws-sdk-go/aws" //"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

func getSubnetCIDR(svc *ec2.EC2, subnetIDs ...string) map[string]string {

	var subnetCIDR = make(map[string]string)

	for _, subnetID := range subnetIDs {
		input := &ec2.DescribeSubnetsInput{
			SubnetIds: []*string{
				aws.String(subnetID),
			},
		}
		result, err := svc.DescribeSubnets(input)
		if err != nil {
			log.Fatal(err)
		}
		subnetCIDR[subnetID] = *result.Subnets[0].CidrBlock
	}
	return subnetCIDR
}

func getNetworkAcls(svc *ec2.EC2, subnetIDs ...string) map[string]*ec2.NetworkAcl {

	var networkACLs = make(map[string]*ec2.NetworkAcl)

	for _, subnetID := range subnetIDs {
		input := &ec2.DescribeNetworkAclsInput{
			Filters: []*ec2.Filter{
				{
					Name: aws.String("association.subnet-id"),
					Values: []*string{
						aws.String(subnetID),
					},
				},
			},
		}

		result, err := svc.DescribeNetworkAcls(input)
		if err != nil {
			log.Fatal(err)
		}
		if len(result.NetworkAcls) == 1 {
			networkACLs[subnetID] = result.NetworkAcls[0]
		} else {
			log.Fatal("Expected 1 network acl entry")
		}
	}
	return networkACLs
}

func getSecurityGroupRules(svc *ec2.EC2, securityGroups []*ec2.GroupIdentifier) []*SecurityGroupRule {

	var securityGroupIds []*string
	for _, group := range securityGroups {
		securityGroupIds = append(securityGroupIds, aws.String(*group.GroupId))
	}
	input := &ec2.DescribeSecurityGroupsInput{
		GroupIds: securityGroupIds,
	}

	result, err := svc.DescribeSecurityGroups(input)
	if err != nil {
		log.Fatal(err.Error())
	}
	var rules []*SecurityGroupRule
	for _, group := range result.SecurityGroups {
		for _, ingressPermission := range group.IpPermissions {
			rule := SecurityGroupRule{egress: false, permission: ingressPermission}
			rules = append(rules, &rule)
		}
		for _, ingressPermission := range group.IpPermissionsEgress {
			rule := SecurityGroupRule{egress: true, permission: ingressPermission}
			rules = append(rules, &rule)
		}
	}
	return rules
}

// check if source can send outgoing traffic to destination over specified port and protocol
func checkNACLEgressAllow(source *instanceState, destination *instanceState, protocol string, destPortRange *ec2.PortRange) []*checkResult {

	var result []*checkResult

	var destIP string

	if len(destPrivateIPAddress) != 0 {
		destIP = destPrivateIPAddress
	} else {

		if usingPublicIP {
			destIP = destination.PublicIP
		}
	}

	for subnetID, nacl := range source.NetworkAcls {
		r := newCheckResult()
		r.DisplayText = "Egress ACL from Subnet " + subnetID

		for _, entry := range nacl.Entries {
			if *entry.Egress && (protocolMapping[*entry.Protocol] == protocol || protocolMapping[*entry.Protocol] == "all") {
				_, allowedIpv4Net, err := net.ParseCIDR(*entry.CidrBlock)
				if err != nil {
					log.Fatal(err)
				}
				if allowedIpv4Net.Contains(net.ParseIP(destIP)) {
					// This condition here will also match the "default" rule with *, but it's rule
					// number is 32767, so it will be overridden by a lower numberd matching rule here
					if protocolMapping[*entry.Protocol] == "all" || (*destPortRange.From >= *entry.PortRange.From && *destPortRange.To <= *entry.PortRange.To) {
						if r.Metadata["MatchedACL"] != nil {
							acl := r.Metadata["MatchedACL"].(ec2.NetworkAclEntry)
							if *entry.RuleNumber < *acl.RuleNumber {
								r.Metadata["MatchedACL"] = *entry
							}
						} else {
							r.Metadata["MatchedACL"] = *entry
						}
						r.Result = true
					}
				}
			}
		}
		result = append(result, &r)
	}
	return result
}

// check if destination allows incoming traffic from source over specified port and protocol
func checkNACLIngressAllow(source *instanceState, destination *instanceState, protocol string, destPortRange *ec2.PortRange) []*checkResult {

	var result []*checkResult

	var sourceIPAddresses []string

	if len(destPrivateIPAddress) != 0 {
		sourceIPAddresses = source.PrivateIPAddresses
	} else {
		if usingPublicIP {
			// The source must either have a public IP or a NAT instance IP
			if len(source.PublicIP) != 0 {
				sourceIPAddresses = append(sourceIPAddresses, source.PublicIP)
			} else {
				// NAT IP

			}
		}
	}

	for _, sourceIP := range sourceIPAddresses {

		r := newCheckResult()
		r.DisplayText = "Ingress ACL at Destination from " + sourceIP

		for subnetId, acl := range destination.NetworkAcls {
			// check the ACLs for the relevant subnet for the destination
			_, destinationSubnetCIDR, _ := net.ParseCIDR(destination.SubnetCIDRs[subnetId])
			if destinationSubnetCIDR.Contains(net.ParseIP(destPrivateIPAddress)) {
				for _, entry := range acl.Entries {
					//log.Printf("%v\n", entry)
					if !*entry.Egress && (protocolMapping[*entry.Protocol] == protocol || protocolMapping[*entry.Protocol] == "all") {
						_, allowedIpv4Net, err := net.ParseCIDR(*entry.CidrBlock)
						if err != nil {
							log.Fatal(err)
						}

						if allowedIpv4Net.Contains(net.ParseIP(sourceIP)) {
							// This condition here will also match the "default" rule with *, but it's rule
							// number is 32767, so it will be overridden by a lower numberd matching rule here
							if protocolMapping[*entry.Protocol] == "all" || (*destPortRange.From >= *entry.PortRange.From && *destPortRange.To <= *entry.PortRange.To) {
								if r.Metadata["MatchedACL"] != nil {
									acl := r.Metadata["MatchedACL"].(ec2.NetworkAclEntry)
									if *entry.RuleNumber < *acl.RuleNumber {
										r.Metadata["MatchedACL"] = *entry
									}
								} else {
									r.Metadata["MatchedACL"] = *entry
								}
								if *entry.RuleAction == "allow" {
									r.Result = true
								}
							}
						}
					}
				}
				result = append(result, &r)
			}
		}

	}
	return result
}

func checkInstanceEgressAllow(source *instanceState, dest *instanceState, protocol string, destPortRange *ec2.PortRange) []*checkResult {

	var result []*checkResult

	r := newCheckResult()
	r.DisplayText = "Security Group at Source allows Egress Traffic"

	for _, rule := range source.SecurityGroupRules {
		if rule.egress && (protocolMapping[*rule.permission.IpProtocol] == protocol || protocolMapping[*rule.permission.IpProtocol] == "all") {

			for _, ipRange := range rule.permission.IpRanges {
				_, allowedIpv4Net, err := net.ParseCIDR(*ipRange.CidrIp)
				if err != nil {
					log.Fatal(err)
				}
				if allowedIpv4Net.Contains(net.ParseIP(destPrivateIPAddress)) {
					if protocolMapping[*rule.permission.IpProtocol] == "all" {
						r.Metadata["MatchedSecurityGroupRule"] = rule
						r.Result = true
						result = append(result, &r)
						return result
					}

					if *destPortRange.From >= *rule.permission.FromPort && *destPortRange.To <= *rule.permission.ToPort {
						r.Metadata["MatchedSecurityGroupRule"] = rule
						r.Result = true
						result = append(result, &r)
						return result
					}

				}
			}

			for _, userIDGroupPair := range rule.permission.UserIdGroupPairs {
				for _, sg := range dest.SecurityGroups {
					if userIDGroupPair.GroupId == sg.GroupId {
						r.Metadata["MatchedSecurityGroupRule"] = rule
						r.Result = true
						result = append(result, &r)
						return result
					}
				}
			}

		}
	}
	return result
}

func checkInstanceIngressAllow(source *instanceState, dest *instanceState, protocol string, destPortRange *ec2.PortRange) []*checkResult {
	var result []*checkResult

	for _, sourceIP := range source.PrivateIPAddresses {
		r := newCheckResult()
		r.DisplayText = "Security Group at Destination allows Ingress traffic from " + sourceIP

		for _, rule := range dest.SecurityGroupRules {
			if !rule.egress && (protocolMapping[*rule.permission.IpProtocol] == protocol || protocolMapping[*rule.permission.IpProtocol] == "all") {

				// Check for source IP ranges
				for _, ipRange := range rule.permission.IpRanges {

					_, allowedIpv4Net, err := net.ParseCIDR(*ipRange.CidrIp)
					if err != nil {
						log.Fatal(err)
					}

					if allowedIpv4Net.Contains(net.ParseIP(sourceIP)) {
						if protocolMapping[*rule.permission.IpProtocol] == "all" {
							r.Metadata["MatchedSecurityGroupRule"] = rule
							r.Result = true
						}

						if *destPortRange.From >= *rule.permission.FromPort && *destPortRange.To <= *rule.permission.ToPort {
							r.Metadata["MatchedSecurityGroupRule"] = rule
							r.Result = true
						}

					}
					if r.Result {
						result = append(result, &r)
						return result
					}
				}
			}
		}
	}

	for _, sg := range source.SecurityGroups {

		r := newCheckResult()
		r.DisplayText = "Security Group at Destination allows Ingress traffic from security group " + *sg.GroupId

		for _, rule := range dest.SecurityGroupRules {
			if !rule.egress && (protocolMapping[*rule.permission.IpProtocol] == protocol || protocolMapping[*rule.permission.IpProtocol] == "all") {

				// Check for source security groups
				for _, userIDGroupPair := range rule.permission.UserIdGroupPairs {

					if *destPortRange.From >= *rule.permission.FromPort && *destPortRange.To <= *rule.permission.ToPort {
						// While using public IP address to connect to the destination instance in a VPC
						// allowing source security group doesn't allow access
						if *userIDGroupPair.GroupId == *sg.GroupId {

							if len(dest.SubnetIds) != 0 && usingPublicIP {
								continue
							}
							r.Metadata["MatchedSecurityGroupRule"] = rule
							r.Result = true
						}
					}
				}
			}
		}
		result = append(result, &r)
		// If we have a matching rule, we return early
		if r.Result {
			return result
		}
	}
	return result
}

func checkHasRoute(source *instanceState, dest *instanceState, displayText string) []*checkResult {

	// TODO: dest is not currently being used

	var result []*checkResult

	r := newCheckResult()
	r.DisplayText = displayText

	var routes []*ec2.Route
	for _, routeTable := range source.Routes {

		for _, route := range routeTable.Routes {
			if route.DestinationCidrBlock != nil {
				_, destIpv4Net, err := net.ParseCIDR(*route.DestinationCidrBlock)
				if err != nil {
					log.Fatal(err)
				}
				if destIpv4Net.Contains(net.ParseIP(destPrivateIPAddress)) {
					routes = append(routes, route)
				}
			}
		}
	}
	if len(routes) > 0 {
		r.Metadata["MatchedRoutes"] = routes

		r.Result = true
	}
	result = append(result, &r)
	return result
}

var inspectConnectivityCmd = &cobra.Command{
	Use:   "connectivity",
	Short: "Check connectivity between two EC2 instances on a port",
	Long: `This sub-command allows for running more fine grained connectivity checks. Examples follow:

Can we connect to instance i-03fb71646161e8626 from i-06d80024e0df241da on TCP port 5985 using the
destination's private IP address:



	$ yawsi ec2  inspect connectivity i-06d80024e0df241da --to i-03fb71646161e8626 --dport 5985 --protocol tcp --using-private-ip
	true


The --verbose flag gives us more information about the checks that are run:


	yawsi ec2 inspect connectivity i-0a80024e0df241da --to i-03fb71646161e8626 --dport 5985 --protocol tcp --destination-private-ip 172.31.13.182 --verbose
	✔ Egress ACL from Subnet subnet-ecd74e89
	✔ Route exists from Source to Destination
	✔ Ingress ACL at Destination from 172.31.41.185
	✔ Egress ACL from Subnet subnet-157b9470
	✔ Egress ACL from Subnet subnet-0e57366b
	✔ Route exists from Destination to Source
	✔ Security Group at Source allows Egress Traffic
	✔ Security Group at Destination allows Ingress traffic from 172.31.41.185
	true
	true


This command also has logic around non-obvious issues. For example, if by mistake, we are trying to
use the Public IP address of an instance in a VPC to connect from another instance and we are relying on
ingress security group rules to allow access, it will not work - we will have to use the private IP address.


Since AWS Network ACLs are stateless and your network setup may be setup to explicitly allow a certain range
of ephermal ports for incoming connections, you can specify a custom ephermal port range. By default, it is
32768-61000. To specify a custom ephermal port range, use --override-ephermal-port-range


	yawsi ec2 inspect connectivity i-03fb71646161e8626 --to i-d3ed150c --dport 20014 --protocol TCP \
		--destination-private-ip 172.31.13.182 --override-ephermal-port-range 49152,65535 --verbose
	`,
	Run: func(cmd *cobra.Command, args []string) {
		var ec2Filters []*ec2.Filter
		var inputInstanceIds []*string
		var ephermalPortRange ec2.PortRange

		// Array of check results
		var (
			checkResults []*checkResult
			result       []*checkResult
		)

		sourceInstanceState := instanceState{}
		destInstanceState := instanceState{}

		if len(toDest) == 0 {
			cmd.Usage()
			os.Exit(1)
		}

		if !(usingPublicIP || len(destPrivateIPAddress) != 0) || (usingPublicIP && len(destPrivateIPAddress) != 0) {
			log.Printf("Must specify --using-private-ip or --using-public-ip")
			cmd.Usage()
			os.Exit(1)
		}

		sess := createSession()
		svc := ec2.New(sess)

		if len(toDest) > 0 {
			if destPort == -1 || len(protocol) == 0 {
				cmd.Usage()
				os.Exit(1)
			}
			protocol = strings.ToLower(protocol)

			if len(customEphermalPortRange) == 0 {
				ephermalPortRange = defaultEphermalPortRange
			} else {
				portRange := strings.Split(customEphermalPortRange, ",")
				lower, err := strconv.ParseInt(portRange[0], 10, 64)
				if err != nil {
					log.Fatal("Invalid port range specified", err)
				}
				higher, err := strconv.ParseInt(portRange[1], 10, 64)
				if err != nil {
					log.Fatal("Invalid port range specified", err)
				}

				ephermalPortRange = ec2.PortRange{From: &lower, To: &higher}
			}
			fromSource := args[0]
			// EC2 instance -> EC2 instance
			if strings.HasPrefix(fromSource, "i-") && strings.HasPrefix(toDest, "i-") {
				inputInstanceIds = append(inputInstanceIds, &fromSource)
				inputInstanceIds = append(inputInstanceIds, &toDest)

				params := &ec2.DescribeInstancesInput{
					DryRun:      aws.Bool(false),
					Filters:     ec2Filters,
					InstanceIds: inputInstanceIds,
				}
				err := svc.DescribeInstancesPages(params,
					func(result *ec2.DescribeInstancesOutput, lastPage bool) bool {
						for _, r := range result.Reservations {
							for _, instance := range r.Instances {

								if *instance.InstanceId == args[0] {
									instanceData := getEC2InstanceData(nil, instance.InstanceId)
									if len(instanceData) != 1 {
										panic("Expected size of instanceData to be 1")
									}
									sourceInstanceState = *instanceData[0]
								}
								if *instance.InstanceId == toDest {
									instanceData := getEC2InstanceData(nil, instance.InstanceId)
									if len(instanceData) != 1 {
										panic("Expected size of instanceData to be 1")
									}
									destInstanceState = *instanceData[0]
								}
							}
						}
						return lastPage
					})
				if err != nil {
					log.Fatal(err)
				}

				// If using public IP, the source must have a public IP address or have a NAT IP

				// If using public IP, the destination must have a public IP address
				if len(destInstanceState.PublicIP) == 0 {

				}

				// If using private IP, the destination must have a private IP address
				// (EC2 classic accounts wouldn't have private IP address)

				// Get Subnet CIDR for each subnet
				// Not currently used
				if len(sourceInstanceState.SubnetIds) != 0 {
					sourceInstanceState.SubnetCIDRs = getSubnetCIDR(svc, sourceInstanceState.SubnetIds...)
					sourceInstanceState.NetworkAcls = getNetworkAcls(svc, sourceInstanceState.SubnetIds...)
				}
				if len(destInstanceState.SubnetIds) != 0 {
					destInstanceState.SubnetCIDRs = getSubnetCIDR(svc, destInstanceState.SubnetIds...)
					destInstanceState.NetworkAcls = getNetworkAcls(svc, destInstanceState.SubnetIds...)
				}

				// Get SG rules for each EC2 instance
				sourceInstanceState.SecurityGroupRules = getSecurityGroupRules(svc, sourceInstanceState.SecurityGroups)
				destInstanceState.SecurityGroupRules = getSecurityGroupRules(svc, destInstanceState.SecurityGroups)

				destPortRange := ec2.PortRange{From: &destPort, To: &destPort}

				// 1. Check egress acl for source subnet
				if len(sourceInstanceState.SubnetCIDRs) != 0 {
					result = checkNACLEgressAllow(&sourceInstanceState, &destInstanceState, protocol, &destPortRange)
					if verboseOutput || debugOutput {
						displayResult(result...)
					}

					for _, r := range result {
						checkResults = append(checkResults, r)
					}

					//log.Printf("%v", *result1)
				}

				// 2. Check if we have a route to the destination
				if len(sourceInstanceState.SubnetIds) != 0 {
					sourceInstanceState.Routes = getRoutes(sourceInstanceState.SubnetIds...)
					result = checkHasRoute(&sourceInstanceState, &destInstanceState, "Route exists from Source to Destination")
					if verboseOutput || debugOutput {
						displayResult(result...)
					}
					for _, r := range result {
						checkResults = append(checkResults, r)
					}
				}

				// 3. Check ingress acl for destination subnet
				if len(destInstanceState.SubnetCIDRs) != 0 {
					result = checkNACLIngressAllow(&sourceInstanceState, &destInstanceState, protocol, &destPortRange)
					//log.Printf("%v", *result2)
					displayResult(result...)

					for _, r := range result {
						checkResults = append(checkResults, r)
					}

					// 4. Check egress acl for destination subnet
					result = checkNACLEgressAllow(&destInstanceState, &sourceInstanceState, protocol, &ephermalPortRange)
					//log.Printf("%v", *result3)

					displayResult(result...)

					for _, r := range result {
						checkResults = append(checkResults, r)
					}

					// 5. Check if destination has a route to the source
					destInstanceState.Routes = getRoutes(destInstanceState.SubnetIds...)
					result = checkHasRoute(&destInstanceState, &sourceInstanceState, "Route exists from Destination to Source")

					displayResult(result...)

					for _, r := range result {
						checkResults = append(checkResults, r)
					}
				}

				if len(sourceInstanceState.SubnetCIDRs) != 0 {
					// 6. Check ingress acl for source subnet
					result = checkNACLIngressAllow(&destInstanceState, &sourceInstanceState, protocol, &ephermalPortRange)
					//log.Printf("%v", *result4)
					displayResult(result...)

					for _, r := range result {
						checkResults = append(checkResults, r)
					}
				}

				// Security group rules are state preserving, so need to check:
				// 1. If egress rules on source allows traffic out
				// 2. If ingress rules on dest allows traffic in
				result = checkInstanceEgressAllow(&sourceInstanceState, &destInstanceState, protocol, &destPortRange)

				displayResult(result...)

				for _, r := range result {
					checkResults = append(checkResults, r)
				}

				result = checkInstanceIngressAllow(&sourceInstanceState, &destInstanceState, protocol, &destPortRange)

				displayResult(result...)

				for _, r := range result {
					checkResults = append(checkResults, r)
				}

				if summarizeResults(checkResults...) {
					color.Green("true\n")
				} else {
					color.Red("false\n")
				}
			} else if sourceIP := net.ParseIP(fromSource); sourceIP != nil && strings.HasPrefix(toDest, "i-") {
				// Source is an IP address and destination is an EC2 instance
				sourceInstanceState.PublicIP = sourceIP.String()
				sourceInstanceState.PrivateIPAddresses = append(sourceInstanceState.PrivateIPAddresses, sourceIP.String())
				instanceData := getEC2InstanceData(nil, &toDest)
				if len(instanceData) != 1 {
					panic("Expected size of instanceData to be 1")
				}
				destInstanceState = *instanceData[0]

				destPortRange := ec2.PortRange{From: &destPort, To: &destPort}

				// 1. Check ingress acl for destination subnet
				if len(destInstanceState.SubnetCIDRs) != 0 {
					result = checkNACLIngressAllow(&sourceInstanceState, &destInstanceState, protocol, &destPortRange)
					//log.Printf("%v", *result2)

					displayResult(result...)

					for _, r := range result {
						checkResults = append(checkResults, r)
					}

					// 2. Check egress acl for destination subnet
					result = checkNACLEgressAllow(&destInstanceState, &sourceInstanceState, protocol, &ephermalPortRange)
					//log.Printf("%v", *result3)

					displayResult(result...)

					for _, r := range result {
						checkResults = append(checkResults, r)
					}

					// 3. Check if destination has a route to the public internet
					destInstanceState.Routes = getRoutes(destInstanceState.SubnetIds...)
					result = checkHasRoute(&destInstanceState, &sourceInstanceState, "Route exists from Destination to Source")

					displayResult(result...)

					for _, r := range result {
						checkResults = append(checkResults, r)
					}
				}

				// Security group rules are state preserving, so need to check:
				// 2. If ingress rules on dest allows traffic in

				result = checkInstanceIngressAllow(&sourceInstanceState, &destInstanceState, protocol, &destPortRange)

				displayResult(result...)

				for _, r := range result {
					checkResults = append(checkResults, r)
				}

				if summarizeResults(checkResults...) {
					color.Green("true\n")
				} else {
					color.Red("false\n")
				}
			} else {
				// TODO: ec2 instance to IP address
				//       lambda function to RDS instance
				//       ec2 instance to RDS instance
				log.Fatal("Unrecognized source specification")
			}
		}
	},
	Args: cobra.ExactArgs(1),
}

var toDest string
var destPort int64
var protocol string
var customEphermalPortRange string
var usingPublicIP bool
var destPrivateIPAddress string

func init() {
	inspectInstancesCmd.AddCommand(inspectConnectivityCmd)
	inspectConnectivityCmd.Flags().StringVarP(&toDest, "to", "", "", "Connectivity Destination - EC2 instance Id/IP address/public")
	inspectConnectivityCmd.Flags().Int64VarP(&destPort, "dport", "", -1, "Destination port")
	inspectConnectivityCmd.Flags().StringVarP(&protocol, "protocol", "", "", "Network protocol (TCP/UDP)")
	inspectConnectivityCmd.Flags().StringVarP(&customEphermalPortRange, "override-ephermal-port-range", "", "", "Override ephermal port range")
	inspectConnectivityCmd.Flags().BoolVarP(&verboseOutput, "verbose", "v", false, "Display more information about the result")
	inspectConnectivityCmd.Flags().BoolVarP(&debugOutput, "debug", "", false, "Display more information about the result")
	inspectConnectivityCmd.Flags().BoolVarP(&usingPublicIP, "using-public-ip", "", false, "Using public IP?")
	inspectConnectivityCmd.Flags().StringVarP(&destPrivateIPAddress, "destination-private-ip", "", "", "Specify private IP address of destination")

	// Examples:
	// yawsi ec2 inspect connectivity instance1 --to instance2 --dport 21000 --protocol TCP
	// yawsi ec2 inspect connectivity instance1 --to public --dport 21000 --protocol UDP
	// /yawsi ec2 inspect connectivity i-048df976bb9da254b --to i-01ed3693328373825 --dport 5985 --protocol TCP --verboseOutput --ephermal-port-range 49152,65535
}
