package cmd

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/service/cloudtrail"
	"github.com/spf13/cobra"
)

var eksWhoisCmd = &cobra.Command{
	Use:   "whois",
	Short: "Find out the AWS username who performed an operation on EKS cluster",
	Run: func(cmd *cobra.Command, args []string) {
		eksUserIdParts := strings.Split(eksUserId, ":")
		if len(eksUserIdParts) != 3 {
			log.Fatal("Invalid EKS audit username.")
		}
		eksUsernameParts := strings.Split(eksUsername, "-")
		if len(eksUsernameParts) != 3 && eksUserIdParts[0] != "heptio-authenticator-aws" {
			log.Fatal("Invalid EKS audit user id. Expected heptio-authenticator-aws:<AWS-ACCOUNT-ID>:<string>")
		}
		assumedRoleResourceName := eksUserIdParts[2] + ":" + eksUsernameParts[2]

		mySession := createSession()
		client := cloudtrail.New(mySession)

		attributeKey := "EventName"
		attributeValue := "AssumeRole"
		a := cloudtrail.LookupAttribute{
			AttributeKey:   &attributeKey,
			AttributeValue: &attributeValue,
		}

		attrs := []*cloudtrail.LookupAttribute{&a}

		timeNow := time.Now()
		timeTo := timeNow.UTC()
		timeFrom := timeNow.Add(time.Duration(-lookbackDuration) * time.Hour)

		eventsInput := cloudtrail.LookupEventsInput{
			LookupAttributes: attrs,
			StartTime:        &timeFrom,
			EndTime:          &timeTo,
		}

		err := client.LookupEventsPages(&eventsInput,
			func(page *cloudtrail.LookupEventsOutput, lastPage bool) bool {
				for _, e := range page.Events {
					for _, r := range e.Resources {
						if *r.ResourceType == "AWS::STS::AssumedRole" && *r.ResourceName == assumedRoleResourceName {
							fmt.Println(e)
						}
					}
				}
				return !lastPage
			})
		if err != nil {
			log.Fatal(err)
		}

	},
}

var eksUserId string
var eksUsername string
var lookbackDuration int

func init() {
	eksWhoisCmd.Flags().StringVarP(&eksUserId, "uid", "", "", "EKS event user ID")
	eksWhoisCmd.Flags().StringVarP(&eksUsername, "username", "", "", "EKS event username")
	eksWhoisCmd.Flags().IntVarP(&lookbackDuration, "lookback", "", 24, "Lookback duration in hours")
	eksCmd.AddCommand(eksWhoisCmd)
}
