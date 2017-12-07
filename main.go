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
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func main() {

	var awsProfile = flag.String("aws-profile", "", "AWS user profile to use")
	var tags = flag.String("tags", "", "List instances having key:value as tag(s)")
	var asgs = flag.Bool("list-asgs", false, "List all ASGs")
	var asgName = flag.String("asg", "", "Descibe a specific ASG")

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

	// If we are not querying via asg name or for asgs
	if len(*asgName) == 0 && !*asgs {
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
	} else {
		var asgNames []*string
		if len(*asgName) != 0 {
			asgNames = append(asgNames, aws.String(*asgName))
		}
		// Default to 100 here, not sure how this works
		// with paging when we have  more than 100 ASGs
		maxSize := int64(100)
		params := &autoscaling.DescribeAutoScalingGroupsInput{
			AutoScalingGroupNames: asgNames,
			MaxRecords:            &maxSize,
		}
		svc := autoscaling.New(sess)
		err := svc.DescribeAutoScalingGroupsPages(params,
			func(result *autoscaling.DescribeAutoScalingGroupsOutput, lastPage bool) bool {
				// When we support multiple ASG names, this will be a way
				// to list all the instances attached to the ASGs
				for _, group := range result.AutoScalingGroups {
					if len(asgNames) != 0 {
						for _, instance := range group.Instances {
							input := &autoscaling.DescribeAutoScalingInstancesInput{
								InstanceIds: []*string{
									aws.String(*instance.InstanceId),
								},
							}
							result, err := svc.DescribeAutoScalingInstances(input)
							if err != nil {
								if aerr, ok := err.(awserr.Error); ok {
									switch aerr.Code() {
									case autoscaling.ErrCodeInvalidNextToken:
										fmt.Println(autoscaling.ErrCodeInvalidNextToken, aerr.Error())
									case autoscaling.ErrCodeResourceContentionFault:
										fmt.Println(autoscaling.ErrCodeResourceContentionFault, aerr.Error())
									default:
										fmt.Println(aerr.Error())
									}
								} else {
									// Print the error, cast err to awserr.Error to get the Code and
									// Message from an error.
									fmt.Println(err.Error())
								}
							}
							for _, instance := range result.AutoScalingInstances {
								fmt.Println(*instance.InstanceId, ":", *instance.AutoScalingGroupName, ":", *instance.ProtectedFromScaleIn)
							}
						}
					} else {
						fmt.Println(*group.AutoScalingGroupName)
					}
				}
				return lastPage
			})
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				switch aerr.Code() {
				case autoscaling.ErrCodeInvalidNextToken:
					fmt.Println(autoscaling.ErrCodeInvalidNextToken, aerr.Error())
				case autoscaling.ErrCodeResourceContentionFault:
					fmt.Println(autoscaling.ErrCodeResourceContentionFault, aerr.Error())
				default:
					fmt.Println(aerr.Error())
				}
			} else {
				// Print the error, cast err to awserr.Error to get the Code and
				// Message from an error.
				fmt.Println(err.Error())
			}
			return
		}
	}
}
