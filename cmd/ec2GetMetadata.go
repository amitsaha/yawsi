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
	"reflect"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/spf13/cobra"
)

type ec2Metadata struct {
	AWSRegion    string `env:"AWS_REGION"`
	AWSAz        string `env:"AWS_AZ"`
	InstanceID   string `env:"INSTANCE_ID"`
	InstanceType string `env:"INSTANCE_TYPE"`
	Name         string `env:"INSTANCE_NAME"`
	PrivateIP    string `env:"PRIVATE_IP"`
	AMI          string `env:"AMI"`
}

// FIXME: have a refresh loop

// getMetadataCmd represents the ec2 get metadata commond
var ec2GetMetadataCmd = &cobra.Command{
	Use:   "get-metadata",
	Short: "Get EC2 Metadata of the current instance",
	Long:  `Get EC2 metadata`,
	Run: func(cmd *cobra.Command, args []string) {
		sess := createSession()
		ec2MetadataSvc := ec2metadata.New(sess)

		if !ec2MetadataSvc.Available() {
			log.Fatal("Could not access EC2 metadata service")
		}

		md := ec2Metadata{}

		idDoc, err := ec2MetadataSvc.GetInstanceIdentityDocument()
		if err != nil {
			log.Fatal("Error retrieving instance Identity document", err.Error())
		}

		md.AWSRegion = idDoc.Region
		md.AWSAz = idDoc.AvailabilityZone
		md.InstanceID = idDoc.InstanceID
		md.InstanceType = idDoc.InstanceType
		md.PrivateIP = idDoc.PrivateIP
		md.AMI = idDoc.ImageID

		// To get the name tag we have to use the AWS API
		ec2Svc := ec2.New(createSession())

		input := &ec2.DescribeInstancesInput{
			InstanceIds: []*string{
				aws.String(md.InstanceID),
			},
		}

		result, err := ec2Svc.DescribeInstances(input)
		if err != nil {
			log.Fatal("Error getting instance details", err.Error())
		}
		for _, r := range result.Reservations {
			for _, instance := range r.Instances {
				for _, tag := range instance.Tags {
					if *tag.Key == "Name" {
						md.Name = *tag.Value
					}
				}
			}
		}

		if exportEnvironmentVariablesBash {
			val := reflect.ValueOf(&md).Elem()
			for i := 0; i < val.NumField(); i++ {
				valueField := val.Field(i)
				typeField := val.Type().Field(i)
				tag := typeField.Tag
				envVar := tag.Get("env")
				fmt.Printf("export %s=%s\n", envVar, valueField.Interface())
			}
		} else {
			fmt.Printf("%#v\n", md)
		}
	},
}

var exportEnvironmentVariablesBash bool

func init() {
	ec2Cmd.AddCommand(ec2GetMetadataCmd)
	ec2GetMetadataCmd.Flags().BoolVar(&exportEnvironmentVariablesBash, "export-env-vars-bash", false, "Print Bash statements for exporting the metadata")
}
