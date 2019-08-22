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
	"os"
	"reflect"
	"strings"
	"text/template"
	"time"

	fuzzyfinder "github.com/ktr0731/go-fuzzyfinder"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/spf13/cobra"
)

type listInstanceData struct {
	Uptime time.Duration
	Name   string
	instanceState
}

func displayFixedInstanceDetails(instancesData ...*instanceState) {

	tmpl := template.New("fixedEC2InstanceDetails")

	tmpl, err := tmpl.Parse(listInstancesFormat)
	if err != nil {
		log.Fatal("Error Parsing template: ", err)
		return
	}

	for _, instance := range instancesData {

		d := listInstanceData{}
		now := time.Now()
		d.Uptime = now.Sub(*instance.LaunchTime)

		for _, tag := range instance.Tags {
			if *tag.Key == "Name" {
				d.Name = *tag.Value
			}
		}

		d.instanceState = *instance

		err1 := tmpl.Execute(os.Stdout, d)
		if err1 != nil {
			log.Fatal("Error executing template: ", err1)

		}
		fmt.Println()
	}
}

// listInstancesCmd represents the listInstances command
var describeInstancesCmd = &cobra.Command{
	Use:   "describe-instances",
	Short: "Describe EC2 instances",
	Long:  `Describe EC2 instances. Filter by tags (tag1:value1, tag2:value2), auto scaling group, interactive selection and more`,
	Run: func(cmd *cobra.Command, args []string) {
		var ec2Filters []*ec2.Filter
		var inputInstanceIds []*string
		var inputInstanceIdsMap = make(map[string]bool)

		if listInstancesFormatHelp {
			fmt.Print("Available format fields:\n\n")
			instanceData := listInstanceData{}
			instanceState := instanceState{}

			s := reflect.ValueOf(&instanceData).Elem()
			typeOfT := s.Type()
			for i := 0; i < s.NumField(); i++ {
				if typeOfT.Field(i).Name != "instanceState" {
					fmt.Printf("%s \n", typeOfT.Field(i).Name)
				}
			}

			s = reflect.ValueOf(&instanceState).Elem()
			typeOfT = s.Type()
			for i := 0; i < s.NumField(); i++ {
				fmt.Printf("%s \n", typeOfT.Field(i).Name)
			}

			fmt.Println()
			os.Exit(0)
		}

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

		if instanceAsgFilter && len(asgName) != 0 {
			cmd.Usage()
			log.Fatal("Only one of --instance-asg-filter and --asg must be specified")
		}

		if instanceAsgFilter {
			maxSize := int64(100)
			params := &autoscaling.DescribeAutoScalingGroupsInput{
				MaxRecords: &maxSize,
			}
			autoScalingGroups := getAutoScalingGroups(params)
			idx, _ := fuzzyfinder.Find(autoScalingGroups, func(i int) string {
				return fmt.Sprintf("%s", *autoScalingGroups[i].AutoScalingGroupName)
			})
			asgName = *autoScalingGroups[idx].AutoScalingGroupName
		}

		// Not filtering by ASG name
		var instanceIDs []*string
		if len(asgName) == 0 {
			if listInstances {
				instancesData := getEC2InstanceData(ec2Filters, inputInstanceIds...)
				displayFixedInstanceDetails(instancesData...)
			} else {
				getEC2InstanceIDs(ec2Filters, &instanceIDs)
				fmt.Println(len(instanceIDs))
				displayEC2Interactive(instanceIDs)
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

var listInstancesFormat string
var listInstancesFormatHelp bool
var listInstances bool
var tags string
var asgName string
var listTags bool
var instanceAsgFilter bool

func init() {
	ec2Cmd.AddCommand(describeInstancesCmd)

	describeInstancesCmd.Flags().BoolVarP(&listInstances, "list", "", false, "List instances")
	describeInstancesCmd.Flags().StringVarP(&listInstancesFormat, "list-format", "", "{{.Name}} {{.Uptime}} {{.PrivateIPAddresses}}", "List instances format string")
	describeInstancesCmd.Flags().BoolVarP(&listInstancesFormatHelp, "list-format-help", "", false, "List all valid format fields")
	describeInstancesCmd.Flags().StringVarP(&instanceIds, "instance-id", "i", "", "Show details of the specified instance(s) (Example: i-a121aas, i=1212aa)")
	describeInstancesCmd.Flags().StringVarP(&tags, "tags", "t", "", "Tags to filter by (tag1:value1, tag2:value2)")
	describeInstancesCmd.Flags().StringVarP(&asgName, "asg", "a", "", "List instances attached to this ASG")
	describeInstancesCmd.Flags().BoolVarP(&instanceAsgFilter, "filter-by-asg", "", false, "Select instances attached to an Auto Scaling Group")
}
