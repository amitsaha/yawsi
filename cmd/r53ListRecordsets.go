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

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/spf13/cobra"
)

// listVpcsCmd represents the list vpc command
var listR53RecordsCmd = &cobra.Command{
	Use:   "list-records",
	Short: "List Route53 records for a zone",
	Long: `List the Route53 records
	
	To list the records:

		$ yawsi r53 list-records --zone-id example.com
	
	`,
	Run: func(cmd *cobra.Command, args []string) {

		if len(r53zoneId) == 0 {
			log.Fatal("Specify zone-name")
		}
		svc := route53.New(session.New())
		r := ListR53RecordSets(svc, r53zoneId)
		fmt.Println(r)
	},
}

var r53zoneId string

func init() {
	r53Cmd.AddCommand(listR53RecordsCmd)
	listR53RecordsCmd.Flags().StringVarP(&r53zoneId, "zone-id", "", "", "Zone ID to list records for")
}
