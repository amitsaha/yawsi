// Copyright © 2018 Amit Saha<amitsaha.in@gmail.com>
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
	"encoding/base64"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/spf13/cobra"
)

// launchMoreLikeCmd represents the launchMoreLike command
var launchMoreLikeCmd = &cobra.Command{
	Use:   "launch-more-like",
	Short: "Launch more AWS EC2 instance like another instance",
	Long:  `launch-more-like creates another AWS instance given another instance id`,
	Run: func(cmd *cobra.Command, args []string) {
		cloneInstanceID := args[0]
		sess := createSession()
		svc := ec2.New(sess)
		var ec2Filters []*ec2.Filter
		ec2Filters = append(ec2Filters, &ec2.Filter{
			Name: aws.String("instance-id"),
			Values: []*string{
				aws.String(cloneInstanceID),
			},
		})
		params := &ec2.DescribeInstancesInput{
			DryRun:  aws.Bool(false),
			Filters: ec2Filters,
		}

		instancesOutput, err := svc.DescribeInstances(params)
		if err != nil {
			log.Fatal(err)
		}
		if len(instancesOutput.Reservations) == 0 {
			log.Fatalf("No instance found with ID: %s", cloneInstanceID)
		}
		instance := instancesOutput.Reservations[0].Instances[0]

		var iamInstanceProfileSpec *ec2.IamInstanceProfileSpecification

		if instance.IamInstanceProfile != nil {
			iamInstanceProfileSpec.Arn = instance.IamInstanceProfile.Arn
		}

		var securityGroupIds []*string
		for _, sg := range instance.SecurityGroups {
			securityGroupIds = append(securityGroupIds, sg.GroupId)
		}
		input := &ec2.DescribeInstanceAttributeInput{
			Attribute:  aws.String("userData"),
			InstanceId: aws.String(*instance.InstanceId),
		}

		result, err := svc.DescribeInstanceAttribute(input)
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				switch aerr.Code() {
				default:
					fmt.Println(aerr.Error())
				}
			} else {
				fmt.Println(err.Error())
			}
			return
		}

		var instanceTags []*ec2.Tag
		for _, tag := range instance.Tags {
			if !strings.HasPrefix(*tag.Key, "aws:") {
				instanceTags = append(instanceTags, &ec2.Tag{
					Key:   tag.Key,
					Value: tag.Value,
				})
			}

		}

		// Add/update additional user specified tags
		// Since the tags will be added to the instance in order
		// they are created, any user specified tag value will
		// override the existing tag value automatically
		if len(updateTags) != 0 {
			for _, f := range strings.Split(updateTags, ",") {
				kv := strings.Split(f, ":")
				key := strings.TrimSpace(kv[0])
				value := strings.TrimSpace(kv[1])

				instanceTags = append(instanceTags, &ec2.Tag{
					Key:   &key,
					Value: &value,
				})
			}
		}

		var userData *string
		if editUserData {
			currentUserData, err := base64.StdEncoding.DecodeString(*result.UserData.Value)
			if err != nil {
				fmt.Println("error:", err)
				return
			}
			userData, err = modifyUserData(string(currentUserData))
			if err != nil {
				log.Fatalf("Error editing user data: %s", err)
			}
			userDataEncoded := base64.StdEncoding.EncodeToString([]byte(*userData))
			userData = &userDataEncoded
		} else {
			userData = result.UserData.Value
		}

		var imageID string
		if len(updatedAMI) > 0 {
			imageID = updatedAMI
		} else {
			imageID = *instance.ImageId
		}

		launchParams := &ec2.RunInstancesInput{
			ImageId:            aws.String(imageID),
			InstanceType:       aws.String(*instance.InstanceType),
			MinCount:           aws.Int64(1),
			MaxCount:           aws.Int64(1),
			IamInstanceProfile: iamInstanceProfileSpec,
			SecurityGroupIds:   securityGroupIds,
			UserData:           userData,
		}

		if len(*instance.SubnetId) > 0 {
			launchParams.SubnetId = instance.SubnetId
		}

		log.Printf("Launching instance with %#v\n", launchParams)
		runResult, err := svc.RunInstances(launchParams)
		if err != nil {
			log.Fatal("Could not create instance", err)
		}

		log.Println("Created instance", *runResult.Instances[0].InstanceId)

		// Add tags to the instance
		_, err = svc.CreateTags(&ec2.CreateTagsInput{
			Resources: []*string{runResult.Instances[0].InstanceId},
			Tags:      instanceTags,
		})
		if err != nil {
			log.Fatal("Could not create tags for instance", runResult.Instances[0].InstanceId, err)
		}

		log.Println("Successfully tagged instance")
	},
	Args: cobra.ExactArgs(1),
}

var editUserData bool
var updateTags string
var updatedAMI string

func init() {
	ec2Cmd.AddCommand(launchMoreLikeCmd)
	launchMoreLikeCmd.Flags().BoolVarP(&editUserData, "edit-user-data", "", false, "Edit User Data")
	launchMoreLikeCmd.Flags().StringVarP(&updateTags, "update-tags", "", "", "Add/update tags")
	launchMoreLikeCmd.Flags().StringVarP(&updatedAMI, "ami-id", "", "", "Use a different AMI")

}
