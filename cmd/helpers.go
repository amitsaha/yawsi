package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"

	"runtime"
	"time"

	"github.com/aws/aws-sdk-go/service/databasemigrationservice"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/fatih/color"

	fuzzyfinder "github.com/ktr0731/go-fuzzyfinder"
)

var protocolMapping = map[string]string{
	"6":    "tcp",
	"17":   "udp",
	"-1":   "all",
	"tcp":  "tcp",
	"udp":  "udp",
	"icmp": "icmp",
}

// https://en.wikipedia.org/wiki/Ephemeral_port
var (
	startingEphermalPort int64 = 32768
	endingEphermalPort   int64 = 61000
)

var defaultEphermalPortRange = ec2.PortRange{
	From: &startingEphermalPort,
	To:   &endingEphermalPort,
}

func createSession(region ...string) *session.Session {
	var sess *session.Session
	var err error

	profile := os.Getenv("AWS_PROFILE")
	if len(profile) != 0 {
		sess = session.Must(session.NewSessionWithOptions(session.Options{
			SharedConfigState: session.SharedConfigEnable,
			Profile:           profile,
		}))
	} else {
		var awsRegion string
		if len(region) == 1 {
			awsRegion = region[0]
		} else if os.Getenv("AWS_REGION") != "" {
			awsRegion = os.Getenv("AWS_REGION")
		} else {
			awsRegion = "us-east-1"
		}
		sess, err = session.NewSession(&aws.Config{
			Region: aws.String(awsRegion),
		})
		if err != nil {
			log.Fatal("Couldn't create a session to talk to AWS", err.Error())
		}
	}
	return sess
}

func modifyUserData(userData string) (*string, error) {
	// TODO: support this better:
	// https://bbengfort.github.io/snippets/2018/01/06/cli-editor-app.html
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}
	tmpDir := os.TempDir()
	tmpFile, tmpFileErr := ioutil.TempFile(tmpDir, "yawsiTmp")
	if tmpFileErr != nil {
		return nil, tmpFileErr
	}

	err := ioutil.WriteFile(tmpFile.Name(), []byte(userData), 0644)
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpFile.Name())

	path, err := exec.LookPath(editor)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(path, tmpFile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	if err != nil {
		return nil, err
	}
	err = cmd.Wait()
	if err != nil {
		return nil, err
	}

	// Read the contents back
	editedFileContents, err := ioutil.ReadFile(tmpFile.Name())
	editedFileContentsStr := string(editedFileContents)
	return &editedFileContentsStr, err
}

func describeSubnetAttachedRouteTables(svc *ec2.EC2, subnetID *string) []*RouteContainer {
	input := &ec2.DescribeRouteTablesInput{

		Filters: []*ec2.Filter{
			{
				Name: aws.String("association.subnet-id"),
				Values: []*string{
					aws.String(*subnetID),
				},
			},
		},
	}

	var routes []*RouteContainer

	result, err := svc.DescribeRouteTables(input)
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
		return nil
	}

	for _, routeTable := range result.RouteTables {
		route := RouteContainer{
			RouteTableId: *routeTable.RouteTableId,
			Main:         false,
			Routes:       routeTable.Routes,
		}
		routes = append(routes, &route)
	}
	return routes

}

func describeVpcMainRouteTables(svc *ec2.EC2, vpcID *string) []*RouteContainer {
	input := &ec2.DescribeRouteTablesInput{

		Filters: []*ec2.Filter{
			{
				Name: aws.String("vpc-id"),
				Values: []*string{
					aws.String(*vpcID),
				},
			},
		},
	}

	var routes []*RouteContainer

	result, err := svc.DescribeRouteTables(input)
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
		return nil
	}

	for _, routeTable := range result.RouteTables {
		// we can have multiple associations here and we want those route tables
		// which are not associated with any subnet and hence are "main" tables
		for _, association := range routeTable.Associations {
			if *association.Main && len(routeTable.Associations) > 1 {
				log.Fatal("Unexpected associations")
			}
			if *association.Main {
				route := RouteContainer{
					Main:         true,
					RouteTableId: *routeTable.RouteTableId,
					Routes:       routeTable.Routes,
				}
				routes = append(routes, &route)
			}
		}

	}
	return routes
}

func getRoutes(subnetIDs ...string) []*RouteContainer {

	sess := createSession()
	svc := ec2.New(sess)

	var routes []*RouteContainer

	var subnetIDsInput []*string

	for _, subnetID := range subnetIDs {
		subnetIDsInput = append(subnetIDsInput, &subnetID)
	}

	input := &ec2.DescribeSubnetsInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("subnet-id"),
				Values: subnetIDsInput,
			},
		},
	}

	result, err := svc.DescribeSubnets(input)
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
		return nil
	}

	if result == nil {
		fmt.Println("Could not retrieve subnet details")
		return nil
	}

	var r []*RouteContainer
	for _, subnetID := range subnetIDs {
		r = describeSubnetAttachedRouteTables(svc, &subnetID)
		if r == nil {
			r = describeVpcMainRouteTables(svc, result.Subnets[0].VpcId)
		}
		for _, route := range r {
			routes = append(routes, route)
		}
	}

	return routes
}

func displayResult(result ...*checkResult) {

	for _, r := range result {
		if verboseOutput || debugOutput {
			if r.Result {
				color.Green("\u2714 %s\n", r.DisplayText)
			} else {
				color.Red("\u2716 %s", r.DisplayText)
			}
		}
		if debugOutput {
			fmt.Printf("%v\n", r.Metadata)
		}
	}
}

func summarizeResults(results ...*checkResult) bool {
	for _, result := range results {
		if !result.Result {
			return false
		}
	}
	return true
}

func getEC2InstanceData(ec2Filters []*ec2.Filter, instanceIds ...*string) []*instanceState {
	params := &ec2.DescribeInstancesInput{
		DryRun:      aws.Bool(false),
		InstanceIds: instanceIds,
		Filters:     ec2Filters,
	}
	sess := createSession()
	svc := ec2.New(sess)

	var instanceStates []*instanceState

	err := svc.DescribeInstancesPages(params,
		func(result *ec2.DescribeInstancesOutput, lastPage bool) bool {
			for _, r := range result.Reservations {
				for _, instance := range r.Instances {
					instanceState := instanceState{
						InstanceId: *instance.InstanceId,
						State:      *instance.State.Name,
						LaunchTime: instance.LaunchTime,
						Tags:       instance.Tags,
					}
					if instance.PublicIpAddress != nil {
						instanceState.PublicIP = *instance.PublicIpAddress
					}

					if instance.KeyName != nil {
						instanceState.KeyName = *instance.KeyName
					}

					if len(instance.NetworkInterfaces) != 0 {

						var networkInterfaces []*string
						for _, ni := range instance.NetworkInterfaces {
							networkInterfaces = append(networkInterfaces, ni.NetworkInterfaceId)

						}
						input := &ec2.DescribeNetworkInterfacesInput{
							NetworkInterfaceIds: networkInterfaces,
						}

						result, err := svc.DescribeNetworkInterfaces(input)
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
							log.Fatal(err)
						}

						for _, ni := range result.NetworkInterfaces {
							if ni.SubnetId != nil && len(*ni.SubnetId) != 0 {
								instanceState.SubnetIds = append(instanceState.SubnetIds, *ni.SubnetId)
								// ni.PrivateIpAddresses is a superset of ni.PrivateIpAddress
								/*if ni.PrivateIpAddress != nil {
									instanceState.PrivateIPAddresses = append(instanceState.PrivateIPAddresses, *ni.PrivateIpAddress)
									log.Printf("%s %s\n", *ni.NetworkInterfaceId, *ni.PrivateIpAddress)
								}*/
								if len(ni.PrivateIpAddresses) != 0 {
									for _, ip := range ni.PrivateIpAddresses {
										instanceState.PrivateIPAddresses = append(instanceState.PrivateIPAddresses, *ip.PrivateIpAddress)
									}
								}

								if len(ni.Groups) != 0 {
									for _, sg := range ni.Groups {
										instanceState.SecurityGroups = append(instanceState.SecurityGroups, sg)

									}

								}
							}
						}
					}

					if instance.VpcId != nil {
						instanceState.VpcID = *instance.VpcId
					}

					for _, tag := range instance.Tags {
						if *tag.Key == "Name" {
							instanceState.Name = *tag.Value
						}
					}
					instanceStates = append(instanceStates, &instanceState)
				}
			}
			return lastPage
		})
	if err != nil {
		log.Fatal(err)
	}

	return instanceStates
}

func getInstanceKeyPairName(instanceID string) string {

	instanceIds := []*string{&instanceID}
	instanceData := getEC2InstanceData(nil, instanceIds...)

	for _, instance := range instanceData {
		if len(instance.KeyName) != 0 {
			return instance.KeyName
		}
	}
	return ""
}

func rdpWindowsHelper(instanceID string, PrivateIP bool, PublicIP bool, ShowCommand bool, KeyPath string, password string) {

	var instanceIds []*string
	if len(instanceID) != 0 {
		instanceIds = append(instanceIds, &instanceID)
	}

	instanceData := getEC2InstanceData(nil, instanceIds...)

	if len(instanceData) == 0 {
		log.Fatal("Instance data couldn't be retrieved")
	} else if len(instanceData) > 1 {
		idx, _ := fuzzyfinder.Find(instanceData, func(i int) string {
			return fmt.Sprintf("[%s] %s %s", instanceData[i].InstanceId, instanceData[i].Name, instanceData[i].State)
		})
		instanceID = instanceData[idx].InstanceId
	}

	var cmdToExecute string
	var cmdArgs []string

	if runtime.GOOS == "windows" {
		if len(password) == 0 {
			password = getWindowsPasswordHelper(instanceID, KeyPath)
		}

		cmdToExecute = "cmdkey.exe"

		if PublicIP {
			cmdArgs = append(cmdArgs, fmt.Sprintf("/add:%s", instanceData[0].PublicIP))
		} else if PrivateIP {
			if len(instanceData[0].PrivateIPAddresses) != 1 {
				panic("Instance has multiple private IP addresses. Not supported yet.")
			}
			cmdArgs = append(cmdArgs, fmt.Sprintf("/add:%s", instanceData[0].PrivateIPAddresses[0]))
		}
		cmdArgs = append(cmdArgs, "/user:Administrator")
		cmdArgs = append(cmdArgs, fmt.Sprintf("/pass:%s", password))

		if ShowCommand {
			log.Print(cmdToExecute, cmdArgs[0], cmdArgs[1], cmdArgs[2])
		} else {
			cmd := exec.Command(cmdToExecute, cmdArgs...)
			err := cmd.Run()
			if err != nil {
				log.Print(err)
			}
		}

		// Clear the slice
		cmdArgs = cmdArgs[:0]

		cmdToExecute = "mstsc.exe"

		if PublicIP {
			cmdArgs = append(cmdArgs, fmt.Sprintf("/v:%s", instanceData[0].PublicIP))
		} else if PrivateIP {
			if len(instanceData[0].PrivateIPAddresses) != 1 {
				panic("Instance has multiple private IP addresses. Not supported yet.")
			}
			cmdArgs = append(cmdArgs, fmt.Sprintf("/add:%s", instanceData[0].PrivateIPAddresses[0]))
		}
		cmdArgs = append(cmdArgs, "/noConsentPrompt")

		if ShowCommand {
			log.Print(cmdToExecute, cmdArgs[0], cmdArgs[1])
		} else {
			log.Printf("Password: %s\n", password)
			cmd := exec.Command(cmdToExecute, cmdArgs...)
			err := cmd.Run()
			if err != nil {
				log.Print(err)
			}
		}
	}
}

func getWindowsPasswordHelper(instanceID string, privateKeyPath string) string {

	if len(privateKeyPath) == 0 {
		usr, _ := user.Current()
		homeDir := usr.HomeDir

		keyPairName := getInstanceKeyPairName(instanceID)
		if len(keyPairName) == 0 {
			log.Fatal("Instance not launched using a key pair")
		}
		log.Printf("No key specified. The instance was launched using keypair: %s\n", keyPairName)
		privateKeyPath = filepath.Join(fmt.Sprintf("%s/.ssh/", homeDir), fmt.Sprintf("%s.pem", keyPairName))
		log.Printf("Attempting to find private key in %s\n", privateKeyPath)
	}

	sess := createSession()
	svc := ec2.New(sess)

	passwordInput := ec2.GetPasswordDataInput{
		InstanceId: &instanceID,
	}

	result, err := svc.GetPasswordData(&passwordInput)
	if err != nil {
		log.Fatal(err)
	}

	password, err := decryptWindowsPassword(privateKeyPath, *result.PasswordData)
	if err != nil {
		log.Fatal(err)
	}
	return password
}

func getAutoScalingGroups(params *autoscaling.DescribeAutoScalingGroupsInput) []*autoscaling.Group {

	var autoScalingGroups []*autoscaling.Group

	sess := createSession()
	svc := autoscaling.New(sess)
	err := svc.DescribeAutoScalingGroupsPages(params,
		func(result *autoscaling.DescribeAutoScalingGroupsOutput, lastPage bool) bool {
			// When we support multiple ASG names, this will be a way
			// to list all the instances attached to the ASGs
			for _, group := range result.AutoScalingGroups {
				autoScalingGroups = append(autoScalingGroups, group)
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
	}
	return autoScalingGroups
}

func displayEC2Interactive(instanceData []*instanceState) {
	selectedData := selectEC2InstanceInteractive(instanceData)
	displayFixedInstanceDetails(selectedData)
}

func selectEC2InstanceInteractive(instanceData []*instanceState) *instanceState {

	idx, _ := fuzzyfinder.Find(instanceData,
		func(i int) string {
			return fmt.Sprintf("[%s] %s %s", instanceData[i].InstanceId, instanceData[i].Name, instanceData[i].State)
		},
		fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if i == -1 {
				return ""
			}
			now := time.Now()
			uptime := now.Sub(*instanceData[i].LaunchTime)

			tags = ""
			for _, tag := range instanceData[i].Tags {
				tags = tags + fmt.Sprintf("%s: %s\n", *tag.Key, *tag.Value)
			}
			return fmt.Sprintf("Instance ID: %s (%s)\nUptime: %s \nPrivate IP: %s\nPublic IP: %s\nSubnet: %s\nVPC: %s \n\n%s",
				instanceData[i].InstanceId,
				instanceData[i].Name,
				uptime,
				instanceData[i].PrivateIPAddresses,
				instanceData[i].PublicIP,
				instanceData[i].SubnetIds,
				instanceData[i].VpcID,
				tags,
			)
		}))
	return instanceData[idx]
}

func getDMSReplicationTasks() []*databasemigrationservice.ReplicationTask {
	svc := databasemigrationservice.New(session.New())
	input := &databasemigrationservice.DescribeReplicationTasksInput{}

	result, err := svc.DescribeReplicationTasks(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case databasemigrationservice.ErrCodeResourceNotFoundFault:
				fmt.Println(databasemigrationservice.ErrCodeResourceNotFoundFault, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return nil
	}
	return result.ReplicationTasks
}

func displayDMSTaskStatusInteractive(taskData []*databasemigrationservice.ReplicationTask) {

	_, _ = fuzzyfinder.Find(taskData,
		func(i int) string {
			return fmt.Sprintf("[%s] - %s", *taskData[i].ReplicationTaskIdentifier, *taskData[i].Status)
		},
		fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if i == -1 {
				return ""
			}

			tableStatistics := getTableStatistics(*taskData[i].ReplicationTaskArn)
			pendingValidation := 0
			validated := 0
			mismatched := 0

			for _, stat := range tableStatistics {
				switch *stat.ValidationState {
				case "Validated":
					validated += 1
				case "Mismatched records":
					mismatched += 1
				case "Pending records":
					pendingValidation += 1
				case "Pending validation":
					pendingValidation += 1
				}

			}
			//now := time.Now()
			//uptime := now.Sub(*taskData[i].LaunchTime)
			statusDetails := ""
			if *taskData[i].Status != "running" {
				statusDetails = *taskData[i].StopReason
			}

			return fmt.Sprintf(
				"Task Identifier: %s (%s)\nStatus Details: %s \nMigration Type: %s\n\n"+
					"Task Statistics \n\nFullLoadProgressPercent: %v\nTablesErrored: %v\nTablesLoaded: %v\nTablesLoading: %v\nTablesQueued: %v\n\n"+
					"Table Validation statistics\n\nValidated: %v\nMismatch: %v\nPending: %v\n",
				*taskData[i].ReplicationTaskIdentifier,
				*taskData[i].Status,
				statusDetails,
				*taskData[i].MigrationType,
				*taskData[i].ReplicationTaskStats.FullLoadProgressPercent,
				*taskData[i].ReplicationTaskStats.TablesErrored,
				*taskData[i].ReplicationTaskStats.TablesLoaded,
				*taskData[i].ReplicationTaskStats.TablesLoading,
				*taskData[i].ReplicationTaskStats.TablesQueued,
				validated,
				mismatched,
				pendingValidation,
			)
		}))
	//return instanceData[idx]
}

func getTableStatistics(taskArn string) []*databasemigrationservice.TableStatistics {
	svc := databasemigrationservice.New(session.New())
	input := &databasemigrationservice.DescribeTableStatisticsInput{
		ReplicationTaskArn: aws.String(taskArn),
	}

	result, err := svc.DescribeTableStatistics(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case databasemigrationservice.ErrCodeResourceNotFoundFault:
				fmt.Println(databasemigrationservice.ErrCodeResourceNotFoundFault, aerr.Error())
			case databasemigrationservice.ErrCodeInvalidResourceStateFault:
				fmt.Println(databasemigrationservice.ErrCodeInvalidResourceStateFault, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return nil
	}

	return result.TableStatistics
}
