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

var rdpWindowsCmd = &cobra.Command{
	Use:   "rdp-windows",
	Short: "RDP into Windows",
	Long: `RDP into a EC2 instance running Windows

	yawsi ec2 rdp-windows i-0121212
	`,
	Run: func(cmd *cobra.Command, args []string) {
		var instanceID string
		var ec2Filters []*ec2.Filter

		if len(args) == 1 {
			instanceID = args[0]
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

			instanceData := getEC2InstanceData(ec2Filters)
			selectedInstance := selectEC2InstanceInteractive(instanceData)
			instanceID = selectedInstance.InstanceId
		}
		rdpWindowsHelper(instanceID, PrivateIP, PublicIP, ShowCommand, KeyPath, rdpPassword)
	},
	//Args: cobra.ExactArgs(1),
}

var PrivateIP bool
var PublicIP bool
var ShowCommand bool
var tagKeys string
var rdpPassword string

func init() {
	ec2Cmd.AddCommand(rdpWindowsCmd)
	rdpWindowsCmd.Flags().StringVarP(&tags, "tags", "t", "", "Tags to filter by (tag1:value1, tag2:value2)")
	rdpWindowsCmd.Flags().StringVarP(&tagKeys, "tag-keys", "", "", "Tag keys to filter by (tag1, tags)")
	rdpWindowsCmd.Flags().BoolVarP(&PrivateIP, "use-private-ip", "", true, "Use Private IP address")
	rdpWindowsCmd.Flags().BoolVarP(&PublicIP, "use-public-ip", "", false, "Use Public IP address")
	rdpWindowsCmd.Flags().StringVarP(&KeyPath, "key-path", "k", "", "Private Key to decrypt the password")
	rdpWindowsCmd.Flags().StringVarP(&rdpPassword, "rdp-password", "", "", "RDP password")
	rdpWindowsCmd.Flags().BoolVarP(&ShowCommand, "show-command", "", false, "Only display the OS command to execute")

}
