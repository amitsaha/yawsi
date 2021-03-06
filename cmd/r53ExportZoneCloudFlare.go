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
	"log"
	"os"

	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/cloudflare/cloudflare-go"
	"github.com/spf13/cobra"
)

func AddRecordsToCloudflare(cfClient *cloudflare.API, zoneName string, recordSet *route53.ListResourceRecordSetsOutput) []*cloudflare.DNSRecordResponse {

	zoneID, err := cfClient.ZoneIDByName(zoneName)
	if err != nil {
		log.Fatal("Cloudflare: " + err.Error())
	}

	var cfDNSRecords []*cloudflare.DNSRecordResponse

	for _, r53record := range recordSet.ResourceRecordSets {
		if *r53record.Type == "SOA" {
			log.Printf("Skipping %v\n", r53record)
			continue
		}
		if r53record.ResourceRecords != nil {
			cFlareRecord := cloudflare.DNSRecord{}
			if strings.HasSuffix(*r53record.Name, ".") {
				log.Printf("Trailing . in route53 record name: %v\n", *r53record.Name)
				cFlareRecord.Name = strings.TrimSuffix(*r53record.Name, ".")
			} else {
				cFlareRecord.Name = *r53record.Name
			}
			cFlareRecord.Type = *r53record.Type

			records, err := cfClient.DNSRecords(zoneID, cFlareRecord)
			if err != nil {
				log.Fatal(err)
			}
			if len(records) != 0 {
				log.Printf("Found existing record(s) for %s\n", cFlareRecord.Name)
				for _, r := range records {
					log.Println(r)
				}
			} else {

				for _, r53r := range r53record.ResourceRecords {
					cFlareRecord.Content = *r53r.Value

					if r53record.AliasTarget != nil {
						cFlareRecord.Content = *r53record.AliasTarget.DNSName
						cFlareRecord.Type = "CNAME"
					}

					log.Printf("Creating DNS record in cloudflare: %v\n", cFlareRecord)
					r, err := cfClient.CreateDNSRecord(zoneID, cFlareRecord)
					if err != nil {
						log.Fatal(err)
					}
					cfDNSRecords = append(cfDNSRecords, r)
				}
			}
		}
	}

	return cfDNSRecords
}

var exportR53ZoneCloudflareCmd = &cobra.Command{
	Use:   "export-zone-cloudflare",
	Short: "Export zone and record sets to cloudflare",
	Long: `Create zone and copy DNS records to CloudFlare DNS
		
	`,
	Run: func(cmd *cobra.Command, args []string) {

		if len(r53ZoneName) == 0 {
			log.Fatal("Specify zone name")
		}

		r53ZoneId, err := GetRoute53ZoneID(r53ZoneName)
		if err != nil {
			log.Fatal(err)
		}
		svc := route53.New(session.New())
		cfClient, err := cloudflare.New(os.Getenv("CF_API_KEY"), os.Getenv("CF_API_EMAIL"))
		if err != nil {
			log.Fatal(err)
		}
		res := AddRecordsToCloudflare(cfClient, r53ZoneName, ListR53RecordSets(svc, r53ZoneId))
		fmt.Println(res)

	},
}

var r53ZoneName string

func init() {
	r53Cmd.AddCommand(exportR53ZoneCloudflareCmd)
	exportR53ZoneCloudflareCmd.Flags().StringVarP(&r53ZoneName, "zone-name", "", "", "Zone name to export records for")
}
