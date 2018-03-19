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

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/spf13/cobra"
)

// listAsgCmd represents the listAsg command
var listAsgCmd = &cobra.Command{
	Use:   "list-asgs",
	Short: "List Autoscaling Groups",
	Run: func(cmd *cobra.Command, args []string) {
		// Default to 100 here, not sure how this works
		// with paging when we have  more than 100 ASGs
		maxSize := int64(100)
		params := &autoscaling.DescribeAutoScalingGroupsInput{
			MaxRecords: &maxSize,
		}

		sess := createSession()
		svc := autoscaling.New(sess)
		err := svc.DescribeAutoScalingGroupsPages(params,
			func(result *autoscaling.DescribeAutoScalingGroupsOutput, lastPage bool) bool {
				// When we support multiple ASG names, this will be a way
				// to list all the instances attached to the ASGs
				for _, group := range result.AutoScalingGroups {
					fmt.Println(*group.AutoScalingGroupName)
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
			} else {
				// Print the error, cast err to awserr.Error to get the Code and
				// Message from an error.
				fmt.Println(err.Error())
			}
			return
		}

	},
}

func init() {
	asgCmd.AddCommand(listAsgCmd)
}
