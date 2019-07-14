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
	"text/template"
	"time"

	"github.com/aws/aws-sdk-go/service/autoscaling"
	fuzzyfinder "github.com/ktr0731/go-fuzzyfinder"
	"github.com/spf13/cobra"
)

type listASGData struct {
	Created time.Time
	Name    string
}

func displayFixedASGDetails(asgsData ...*autoscaling.Group) {

	tmpl := template.New("fixedEC2ASGDetails")

	tmpl, err := tmpl.Parse(listASGsFormat)
	if err != nil {
		log.Fatal("Error Parsing template: ", err)
		return
	}

	for _, asg := range asgsData {

		d := listASGData{}
		d.Created = *asg.CreatedTime
		d.Name = *asg.AutoScalingGroupName

		err1 := tmpl.Execute(os.Stdout, d)
		if err1 != nil {
			log.Fatal("Error executing template: ", err1)

		}
		fmt.Println()
	}
}

// listInstancesCmd represents the listInstances command
var describeASGsCmd = &cobra.Command{
	Use:   "describe-asgs",
	Short: "Describe ASGs",
	Long:  `Describe auto scaling groups. Filter by tags (tag1:value1, tag2:value2), auto scaling group, interactive selection and more`,
	Run: func(cmd *cobra.Command, args []string) {

		if listASGsFormatHelp {
			fmt.Print("Available format fields:\n\nTODO")
			fmt.Println()
			os.Exit(0)
		}

		// Not filtering by ASG name
		if len(asgName) == 0 {
			// Default to 100 here, not sure how this works
			// with paging when we have  more than 100 ASGs
			maxSize := int64(100)
			params := &autoscaling.DescribeAutoScalingGroupsInput{
				MaxRecords: &maxSize,
			}
			asgsData := getAutoScalingGroups(params)
			//log.Printf("%v\n", autoScalingGroups)

			displayEC2InteractiveASG(asgsData)
		}
	},
}

func displayEC2InteractiveASG(asgsData []*autoscaling.Group) {
	selectedData := selectEC2ASGInteractive(asgsData)
	displayFixedASGDetails(selectedData)
}

func selectEC2ASGInteractive(asgData []*autoscaling.Group) *autoscaling.Group {

	idx, _ := fuzzyfinder.Find(asgData,
		func(i int) string {
			return fmt.Sprintf("[%s]", *asgData[i].AutoScalingGroupName)
		},
		fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if i == -1 {
				return ""
			}

			tags := "Tags:\n"
			for _, tag := range asgData[i].Tags {
				tags = tags + fmt.Sprintf("  %s: %s\n", *tag.Key, *tag.Value)
			}

			instances := "Instances:\n"
			for _, instance := range asgData[i].Instances {
				instances = instances + fmt.Sprintf("  %s\n", *instance.InstanceId)
			}

			return fmt.Sprintf("Name: %s\nARN: %s\nVPCZoneIdentifiers: %s\nCreated: %s\n\nDesired Capacity: %d\nMaxSize: %d\nMinSize: %d\n\n%s\n%s",
				*asgData[i].AutoScalingGroupName,
				*asgData[i].AutoScalingGroupARN,
				*asgData[i].VPCZoneIdentifier,
				*asgData[i].CreatedTime,
				*asgData[i].DesiredCapacity,
				*asgData[i].MaxSize,
				*asgData[i].MinSize,
				tags,
				instances,
			)
		}))
	return asgData[idx]
}

var listASGsFormat string
var listASGsFormatHelp bool

func init() {
	ec2Cmd.AddCommand(describeASGsCmd)
	describeASGsCmd.Flags().StringVarP(&listASGsFormat, "list-format", "", "[{{.Name}}] : {{.Created}}", "List ASGs format string")
}
