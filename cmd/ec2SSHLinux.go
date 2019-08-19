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
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/spf13/cobra"
)

var ec2SSHLinuxCmd = &cobra.Command{
	Use:   "ssh-linux",
	Short: "Start a SSH session into a Linux instance",
	Long: `Start a SSH session into a Linux instance:

	    yawsi ec2 ssh-linux i-0121212
	`,
	Run: func(cmd *cobra.Command, args []string) {
		var instanceID string
		var ec2Filters []*ec2.Filter
		var instanceDetails []*instanceState

		if len(args) == 1 {
			instanceID = args[0]
			instanceDetails = getEC2InstanceData(ec2Filters, &instanceID)
		} else {
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

			if len(tagKeys) != 0 {
				for _, tag := range strings.Split(tagKeys, ",") {
					tag = strings.TrimSpace(tag)
					ec2Filters = append(ec2Filters, &ec2.Filter{
						Name: aws.String("tag-key"),
						Values: []*string{
							aws.String(tag),
						},
					})
				}
			}

			var instanceIDs []*string
			go getEC2InstanceIDs(ec2Filters, &instanceIDs)
			selectedInstanceDetails := selectEC2InstanceInteractive(&instanceIDs)
			instanceDetails = append(instanceDetails, selectedInstanceDetails)
		}

		startSSHSessionLinux(instanceDetails, PrivateIP, PublicIP, KeyPath, sshUsername)
	},
}

var sshUsername string

func init() {
	ec2Cmd.AddCommand(ec2SSHLinuxCmd)
	ec2SSHLinuxCmd.Flags().StringVarP(&tags, "tags", "t", "", "Tags to filter by (tag1:value1, tag2:value2)")
	ec2SSHLinuxCmd.Flags().BoolVarP(&PrivateIP, "use-private-ip", "", true, "Use Private IP address")
	ec2SSHLinuxCmd.Flags().BoolVarP(&PublicIP, "use-public-ip", "", false, "Use Public IP address")
	ec2SSHLinuxCmd.Flags().StringVarP(&KeyPath, "key-path", "k", "", "Private Key to decrypt the password")
	ec2SSHLinuxCmd.Flags().StringVarP(&sshUsername, "username", "u", "", "Username to SSH in as")
}
