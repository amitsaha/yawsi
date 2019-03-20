package cmd

import (
	"time"

	"github.com/aws/aws-sdk-go/service/ec2"
)

var instanceIds string
var verboseOutput bool
var debugOutput bool

// RouteContainer represents a set of routes and whether it
// is associated to the VPC main route table or not
type RouteContainer struct {
	Main         bool
	RouteTableId string
	Routes       []*ec2.Route
}

// Embdes ec2.IpPermission and adds an additional field
// to mark whether this is an inbound or outbound security group rule
type SecurityGroupRule struct {
	permission *ec2.IpPermission
	egress     bool
}

type instanceState struct {
	InstanceId string
	State      string
	LaunchTime *time.Time
	KeyName    string
	Name       string

	Tags               []*ec2.Tag
	VpcID              string
	PublicIP           string
	SecurityGroups     []*ec2.GroupIdentifier
	SecurityGroupRules []*SecurityGroupRule

	SubnetIds []string

	// Map of network interface id to subnet id
	NetworkInterfaces map[string]string

	PrivateIPAddresses []string

	// Map of subnet ID to SubnetCIDR
	SubnetCIDRs map[string]string
	// Map of subnet ID to NetworkACLs
	NetworkAcls map[string]*ec2.NetworkAcl

	Routes []*RouteContainer
}

type checkResult struct {
	Result      bool
	DisplayText string
	Metadata    map[string]interface{}
}

func newCheckResult() checkResult {
	result := checkResult{Result: false}
	result.Metadata = make(map[string]interface{})
	return result
}
