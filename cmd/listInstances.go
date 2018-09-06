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
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/spf13/cobra"
)

// listInstancesCmd represents the listInstances command
var listInstancesCmd = &cobra.Command{
	Use:   "list-instances",
	Short: "List EC2 instances",
	Long:  `List EC2 instances. Filter by tags (tag1:value1, tag2:value2)`,
	Run: func(cmd *cobra.Command, args []string) {
		var ec2Filters []*ec2.Filter
		var inputInstanceIds []*string
		var inputInstanceIdsMap = make(map[string]bool)

		if len(instanceIds) != 0 {
			instances := strings.Split(instanceIds, ",")
			for idx := range instances {
				inputInstanceIds = append(inputInstanceIds, &instances[idx])
				inputInstanceIdsMap[instances[idx]] = true
			}

		}

		if len(tags) != 0 {
			for _, tag := range strings.Split(tags, ",") {
				tag = strings.TrimSpace(tag)
				key := tag[0:strings.LastIndex(tag, ":")]
				value := tag[strings.LastIndex(tag, ":")+1 : len(tag)]

				ec2Filters = append(ec2Filters, &ec2.Filter{
					Name: aws.String("tag:" + key),
					Values: []*string{
						aws.String(value),
					},
				})
			}
		}
		// Not filtering by ASG name
		if len(asgName) == 0 {
			params := &ec2.DescribeInstancesInput{
				DryRun:      aws.Bool(false),
				Filters:     ec2Filters,
				InstanceIds: inputInstanceIds,
			}
			sess := createSession()
			svc := ec2.New(sess)
			err := svc.DescribeInstancesPages(params,
				func(result *ec2.DescribeInstancesOutput, lastPage bool) bool {
					for _, r := range result.Reservations {
						for _, instance := range r.Instances {
							now := time.Now()
							uptime := now.Sub(*instance.LaunchTime)
							fmt.Println(*instance.InstanceId, ":", *instance.State.Name, ":", uptime, ":", *instance.PublicDnsName, ":", *instance.PrivateDnsName)

							if listTags == true {
								for _, tag := range instance.Tags {
									fmt.Printf("%s:%s\n", *tag.Key, *tag.Value)
								}
							}
						}
					}
					return lastPage
				})
			if err != nil {
				log.Fatal(err)
			}
		}

		if len(asgName) != 0 {
			var asgNames []*string
			asgNames = append(asgNames, aws.String(asgName))
			// Default to 100 here, not sure how this works
			// with paging when we have  more than 100 ASGs
			maxSize := int64(100)
			params := &autoscaling.DescribeAutoScalingGroupsInput{
				AutoScalingGroupNames: asgNames,
				MaxRecords:            &maxSize,
			}
			sess := createSession()
			svc := autoscaling.New(sess)
			err := svc.DescribeAutoScalingGroupsPages(params,
				func(result *autoscaling.DescribeAutoScalingGroupsOutput, lastPage bool) bool {
					// When we support multiple ASG names, this will be a way
					// to list all the instances attached to the ASGs
					for _, group := range result.AutoScalingGroups {
						if len(asgNames) != 0 {
							for _, instance := range group.Instances {
								input := &autoscaling.DescribeAutoScalingInstancesInput{
									InstanceIds: []*string{
										aws.String(*instance.InstanceId),
									},
								}
								result, err := svc.DescribeAutoScalingInstances(input)
								if err != nil {
									if aerr, ok := err.(awserr.Error); ok {
										switch aerr.Code() {
										case autoscaling.ErrCodeInvalidNextToken:
											fmt.Println(autoscaling.ErrCodeInvalidNextToken, aerr.Error())
										case autoscaling.ErrCodeResourceContentionFault:
											fmt.Println(autoscaling.ErrCodeResourceContentionFault, aerr.Error())
										default:
											fmt.Println(aerr.Error())
										}
									} else {
										// Print the error, cast err to awserr.Error to get the Code and
										// Message from an error.
										fmt.Println(err.Error())
									}
								}
								for _, currentInstance := range result.AutoScalingInstances {
									// If we are filtering by instance IDs, only show the details for the specified
									// instance IDs
									if len(inputInstanceIds) != 0 {
										if _, ok := inputInstanceIdsMap[*currentInstance.InstanceId]; ok {
											fmt.Println(*currentInstance.InstanceId, ":", *currentInstance.AutoScalingGroupName, ":", *currentInstance.AvailabilityZone, ":", *currentInstance.ProtectedFromScaleIn)
										}
									} else {
										fmt.Println(*currentInstance.InstanceId, ":", *currentInstance.AutoScalingGroupName, ":", *currentInstance.AvailabilityZone, ":", *currentInstance.ProtectedFromScaleIn)
									}
								}
							}
						} else {
							fmt.Println(*group.AutoScalingGroupName)
						}
					}
					return lastPage
				})
			if err != nil {
				if aerr, ok := err.(awserr.Error); ok {
					switch aerr.Code() {
					case autoscaling.ErrCodeInvalidNextToken:
						fmt.Println(autoscaling.ErrCodeInvalidNextToken, aerr.Error())
					case autoscaling.ErrCodeResourceContentionFault:
						fmt.Println(autoscaling.ErrCodeResourceContentionFault, aerr.Error())
					default:
						fmt.Println(aerr.Error())
					}
				}
			}
		}
	},
}

var tags string
var asgName string
var instanceIds string
var listTags bool

func init() {
	ec2Cmd.AddCommand(listInstancesCmd)
	listInstancesCmd.Flags().StringVarP(&instanceIds, "instance-id", "i", "", "Show details of the specified instance(s) (Example: i-a121aas, i=1212aa)")
	listInstancesCmd.Flags().StringVarP(&tags, "tags", "t", "", "Tags to filter by (tag1:value1, tag2:value2)")
	listInstancesCmd.Flags().StringVarP(&asgName, "asg", "a", "", "List instances attached to this ASG")
	listInstancesCmd.Flags().BoolVarP(&listTags, "show-tags", "s", false, "Show instance tags")

}
