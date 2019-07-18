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
	"text/template"

	"github.com/aws/aws-sdk-go/service/autoscaling"
	fuzzyfinder "github.com/ktr0731/go-fuzzyfinder"
	"github.com/spf13/cobra"
)

// listInstancesCmd represents the listInstances command
var describeASGsCmd = &cobra.Command{
	Use:   "describe-asgs",
	Short: "Describe ASGs",
	Long:  `Describe auto scaling groups with interactive selection.`,
	Run: func(cmd *cobra.Command, args []string) {

		if listASGsFormatHelp {
			fmt.Print("Available format fields:\n\nTODO")
			asg := autoscaling.Group{}

			s := reflect.ValueOf(&asg).Elem()
			typeOfT := s.Type()
			for i := 0; i < s.NumField(); i++ {
				fmt.Printf("%s \n", typeOfT.Field(i).Name)
			}

			fmt.Println()
			os.Exit(0)
		}

		// Default to 100 here, not sure how this works
		// with paging when we have  more than 100 ASGs
		maxSize := int64(5)
		params := &autoscaling.DescribeAutoScalingGroupsInput{
			MaxRecords: &maxSize,
		}
		var asgs autoscalingGroups = getAutoScalingGroups(params)
		selectMulti(asgs)
	},
}

type autoscalingGroups []*autoscaling.Group

func (asgs autoscalingGroups) Summary(i int) string {
	return fmt.Sprintf("[%s]", *asgs[i].AutoScalingGroupName)
}

func (asgs autoscalingGroups) Output(idxs ...int) {
	tmpl := template.New("fixedEC2ASGDetails")
	tmpl, err := tmpl.Parse(listASGsFormat)
	if err != nil {
		log.Fatal("Error Parsing template: ", err)
		return
	}

	for _, i := range idxs {
		err1 := tmpl.Execute(os.Stdout, *asgs[i])
		if err1 != nil {
			log.Fatal("Error executing template: ", err1)
		}
		fmt.Println()
	}
}

func (asgs autoscalingGroups) Details(i int) string {
	tags := "Tags:\n"
	d := asgs[i]
	for _, tag := range d.Tags {
		tags = tags + fmt.Sprintf("  %s: %s\n", *tag.Key, *tag.Value)
	}

	instances := "Instances:\n"
	for _, instance := range d.Instances {
		instances = instances + fmt.Sprintf("  %s\n", *instance.InstanceId)
	}

	return fmt.Sprintf("Name: %s\nARN: %s\nVPCZoneIdentifiers: %s\nCreated: %s\n\nDesired Capacity: %d\nMaxSize: %d\nMinSize: %d\n\n%s\n%s",
		*d.AutoScalingGroupName,
		*d.AutoScalingGroupARN,
		*d.VPCZoneIdentifier,
		*d.CreatedTime,
		*d.DesiredCapacity,
		*d.MaxSize,
		*d.MinSize,
		tags,
		instances,
	)
}

type Selectable interface {
	// Summary to display for a selectable item
	Summary(int) string
	// Details/Preview to display for a selectable item
	Details(int) string
	// Results to return when an item a selected
	Output(...int)
}

func selectMulti(data Selectable) {
	indices, err := fuzzyfinder.FindMulti(
		data,
		data.Summary,
		fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if i == -1 {
				return ""
			}
			return data.Details(i)
		}))
	if err != nil {
		panic(err)
	}
	data.Output(indices...)
}

var listASGsFormat string
var listASGsFormatHelp bool

func init() {
	ec2Cmd.AddCommand(describeASGsCmd)
	describeASGsCmd.Flags().StringVarP(&listASGsFormat, "list-format", "", "[{{.AutoScalingGroupName}}] : {{.CreatedTime}}", "List ASGs format string")
	describeASGsCmd.Flags().BoolVarP(&listASGsFormatHelp, "list-format-help", "", false, "List all valid format fields")
}
