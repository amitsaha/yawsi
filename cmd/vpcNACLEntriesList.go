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
	"os"

	"html/template"
	"log"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/spf13/cobra"
)

func getPortFrom(portRange *ec2.PortRange) string {
	if portRange == nil {
		return strconv.FormatInt(0, 10)
	} else {
		return strconv.FormatInt(*portRange.From, 10)
	}
}

func getPortTo(portRange *ec2.PortRange) string {
	if portRange == nil {
		return strconv.FormatInt(0, 10)
	} else {
		return strconv.FormatInt(*portRange.To, 10)
	}
}

func getIcmpCode(icmpTypeCode *ec2.IcmpTypeCode) string {
	return strconv.FormatInt(*icmpTypeCode.Code, 10)
}

func getIcmpType(icmpTypeCode *ec2.IcmpTypeCode) string {
	return strconv.FormatInt(*icmpTypeCode.Type, 10)
}

func outputTerraformInline(entries []*ec2.NetworkAclEntry) {

	egressRule := `
egress {
  protocol   = "{{.Protocol }}"
  rule_no    = "{{ .RuleNumber }}"
  action     = "{{ .RuleAction }}"
  cidr_block = "{{ .CidrBlock }}"
  from_port  = "{{ .PortRange | getPortFrom }}"
  to_port    = "{{ .PortRange | getPortTo }}"
}
`

	ingressRule := `
ingress {
  protocol   = "{{.Protocol }}"
  rule_no    = "{{ .RuleNumber }}"
  action     = "{{ .RuleAction }}"
  cidr_block = "{{ .CidrBlock }}"
  from_port  = "{{ .PortRange | getPortFrom }}"
  to_port    = "{{ .PortRange | getPortTo }}"
}
`

	funcMap := template.FuncMap{
		"getPortFrom": getPortFrom,
		"getPortTo":   getPortTo,
	}

	egressTmpl := template.New("egress").Funcs(funcMap)
	egressTmpl, err := egressTmpl.Parse(egressRule)
	if err != nil {
		log.Fatal("Error Parsing template: ", err)
		return
	}

	ingressTmpl := template.New("ingress").Funcs(funcMap)
	ingressTmpl, err = ingressTmpl.Parse(ingressRule)
	if err != nil {
		log.Fatal("Error Parsing template: ", err)
		return
	}

	for _, entry := range entries {
		// Rule nos from 32767+ are reserved by AWS and hence no point in
		// generating terraform output for it
		if *entry.RuleNumber >= 32767 {
			continue
		}
		if *entry.Egress {
			err1 := egressTmpl.Execute(os.Stdout, entry)
			if err1 != nil {
				log.Fatal("Error executing template: ", err1)

			}
		} else {

			err1 := ingressTmpl.Execute(os.Stdout, entry)
			if err1 != nil {
				log.Fatal("Error executing template: ", err1)

			}
		}
	}
}

func outputTerraformResource(aclId string, entries []*ec2.NetworkAclEntry) {

	type Rule struct {
		NetworkACLId string
		*ec2.NetworkAclEntry
	}

	resource := `
resource "aws_network_acl_rule" "rule_{{ .RuleNumber }}" {
  network_acl_id = "{{ .NetworkACLId }}
  egress         = {{ .Egress }}
  protocol   = "{{.Protocol }}"
  rule_number    = "{{ .RuleNumber }}"
  rule_action     = "{{ .RuleAction }}"{{if .CidrBlock}}
  cidr_block = "{{ .CidrBlock }}"
  {{ end }}{{if .Ipv6CidrBlock}}
  ipv6_cidr_block = "{{ .Ipv6CidrBlock }}"
  {{ end }}{{if .IcmpTypeCode}}
  icmp_type  = "{{ .IcmpTypeCode | getIcmpType }}"
  icmp_code  = "{{ .IcmpTypeCode | getIcmpCode }}"
  {{ end }}
}
`

	funcMap := template.FuncMap{
		"getPortFrom": getPortFrom,
		"getPortTo":   getPortTo,
		"getIcmpType": getIcmpType,
		"getIcmpCode": getIcmpCode,
	}

	tmpl := template.New("resource").Funcs(funcMap)
	tmpl, err := tmpl.Parse(resource)
	if err != nil {
		log.Fatal("Error Parsing template: ", err)
		return
	}

	for _, entry := range entries {
		// Rule nos from 32767+ are reserved by AWS and hence no point in
		// generating terraform output for it
		if *entry.RuleNumber >= 32767 {
			continue
		}
		rule := Rule{
			NetworkACLId:    aclId,
			NetworkAclEntry: entry,
		}

		err1 := tmpl.Execute(os.Stdout, rule)
		if err1 != nil {
			log.Fatal("Error executing template: ", err1)
		}

	}
}

var listNACLEntriesCmd = &cobra.Command{
	Use:   "list-nacl-entries",
	Short: "List nacl entries in a Network ACL",
	Long: `List NACL entries attached

	.\yawsi.exe vpc list-nacl-entries --nacl-id acl-a7f118c1 --output-format tf_resource

	resource "aws_network_acl_rule" "rule_10" {
		network_acl_id = "acl-a7f118c1"
		egress         = false
		protocol       = "-1"
		rule_number    = "100"
		rule_action     = "allow"
		cidr_block = "0.0.0.0/0"
	}


	`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(naclID) == 0 {
			cmd.Usage()
			os.Exit(1)
		}
		input := &ec2.DescribeNetworkAclsInput{
			NetworkAclIds: []*string{
				aws.String(naclID),
			},
		}

		sess := createSession()
		svc := ec2.New(sess)
		result, err := svc.DescribeNetworkAcls(input)
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				switch aerr.Code() {
				default:
					fmt.Println(aerr.Error())
				}
			} else {
				// Print the error, cast err to awserr.Error to get the Code and
				// Message from an error.
				fmt.Println(err.Error())
			}
		} else {
			if len(result.NetworkAcls) == 1 {
				if len(outputFormat) == 0 {
					log.Printf("%v\n", result.NetworkAcls[0].Entries)
				} else {
					if outputFormat == "tf_inline" {
						outputTerraformInline(result.NetworkAcls[0].Entries)
					} else if outputFormat == "tf_resource" {
						outputTerraformResource(naclID, result.NetworkAcls[0].Entries)
					} else {
						log.Fatal("Invalid output format specified")
					}
				}
			}
		}
	},
}

var naclID string
var outputFormat string

func init() {
	vpcCmd.AddCommand(listNACLEntriesCmd)
	listNACLEntriesCmd.Flags().StringVarP(&naclID, "nacl-id", "", "", "List NACL entries in a specific Network ACL")
	listNACLEntriesCmd.Flags().StringVarP(&outputFormat, "output-format", "", "", "Specify output format - tf_inline, tf_resource")
}
