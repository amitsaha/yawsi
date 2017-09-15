package main

// Steps to use this:
// 1. Create ~/.aws/credentials of the form:
// [profile_name]
// aws_access_key_id=
// aws_secret_access_key=
// ..

// 2. Create ~/.aws/config of the form:
// [profile_name]
// region=ap-southeast-2/us-east-1

// cloudi -role=<myrole> -aws-profile=profile_name

import (
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func main() {

	var awsProfile = flag.String("aws-profile", "", "AWS user profile to use")
	var tags = flag.String("tags", "", "List instances having key:value as tag(s)")

	flag.Parse()

	if len(*awsProfile) == 0 {
		log.Fatal("Must specify --aws-profile")
	}

	// We force loading of shared configuration (a.k.a. ~/.aws/config)
	// so that we don't have to specify another environment variable
	// (AWS_SDK_LOAD_CONFIG)
	// We also set the AWS profile to use instead of having to set AWS_PROFILE
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Profile:           *awsProfile,
	}))
	svc := ec2.New(sess)

	// Build up the array of filters
	var ec2Filters []*ec2.Filter
	if len(*tags) != 0 {
		for _, f := range strings.Split(*tags, ",") {
			kv := strings.Split(f, ":")
			ec2Filters = append(ec2Filters, &ec2.Filter{
				Name: aws.String("tag:" + strings.TrimSpace(kv[0])),
				Values: []*string{
					aws.String(strings.TrimSpace(kv[1])),
				},
			})
		}
	}
	params := &ec2.DescribeInstancesInput{
		DryRun:  aws.Bool(false),
		Filters: ec2Filters,
	}
	err := svc.DescribeInstancesPages(params,
		func(result *ec2.DescribeInstancesOutput, lastPage bool) bool {
			for _, r := range result.Reservations {
				for _, instance := range r.Instances {
					now := time.Now()
					uptime := now.Sub(*instance.LaunchTime)
					fmt.Println(*instance.InstanceId, ":", *instance.State.Name, ":", uptime, ":", *instance.PublicDnsName, ":", *instance.PrivateDnsName)
				}
			}
			return lastPage
		})
	if err != nil {
		log.Fatal(err)
	}
}
