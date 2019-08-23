package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/eks"

	"gopkg.in/yaml.v2"
)

func describeEKSCluster(input *eks.DescribeClusterInput) *eks.DescribeClusterOutput {
	sess := createSession()
	svc := eks.New(sess)

	result, err := svc.DescribeCluster(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case eks.ErrCodeResourceNotFoundException:
				fmt.Println(eks.ErrCodeResourceNotFoundException, aerr.Error())
			case eks.ErrCodeClientException:
				fmt.Println(eks.ErrCodeClientException, aerr.Error())
			case eks.ErrCodeServerException:
				fmt.Println(eks.ErrCodeServerException, aerr.Error())
			case eks.ErrCodeServiceUnavailableException:
				fmt.Println(eks.ErrCodeServiceUnavailableException, aerr.Error())
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
	return result
}

func DescribeEKSCluster(clusterName *string) *eks.DescribeClusterOutput {

	input := &eks.DescribeClusterInput{
		Name: clusterName,
	}
	return describeEKSCluster(input)
}

func listEKSCluster(input *eks.ListClustersInput) *eks.ListClustersOutput {
	sess := createSession()
	svc := eks.New(sess)

	result, err := svc.ListClusters(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case eks.ErrCodeResourceNotFoundException:
				fmt.Println(eks.ErrCodeResourceNotFoundException, aerr.Error())
			case eks.ErrCodeClientException:
				fmt.Println(eks.ErrCodeClientException, aerr.Error())
			case eks.ErrCodeServerException:
				fmt.Println(eks.ErrCodeServerException, aerr.Error())
			case eks.ErrCodeServiceUnavailableException:
				fmt.Println(eks.ErrCodeServiceUnavailableException, aerr.Error())
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
	return result
}

func ListEKSClusters(details bool) *eks.ListClustersOutput {

	input := &eks.ListClustersInput{}
	return listEKSCluster(input)
}

// Idea from http://www.studytrails.com/devops/kubernetes/local-dns-resolution-for-eks-with-private-endpoint/
func GetPrivateMasterIP(clusterName *string) *string {
	description := ec2.Filter{
		Name:   aws.String("description"),
		Values: []*string{aws.String("Amazon EKS " + *clusterName)},
	}

	input := &ec2.DescribeNetworkInterfacesInput{
		Filters: []*ec2.Filter{
			&description,
		},
	}
	result := GetNetworkInterfaces(input)
	if result != nil {
		return result.NetworkInterfaces[0].PrivateIpAddress
	}
	return nil
}

type KubeConfigType struct {
	APIVersion     string           `yaml:"apiVersion"`
	CurrentContext string           `yaml:"current-context"`
	Clusters       []ClusterConfig  `yaml:"clusters"`
	Contexts       []ClusterContext `yaml:"contexts"`
	Kind           string           `yaml:"kind"`
	Users          []ClusterUser    `yaml:"users"`
}

type ClusterConfig struct {
	Cluster map[string]string `yaml:"cluster"`
	Name    string            `yaml:"name"`
}

type ClusterContext struct {
	Context map[string]string `yaml:"context"`
	Name    string            `yaml:"name"`
}

type ClusterUser struct {
	Name string                            `yaml:"name"`
	User map[string]ClusterUserExecDetails `yaml:"user"`
}

type ClusterUserExecDetails struct {
	ApiVersion string   `yaml:"apiVersion"`
	Command    string   `yaml:"command"`
	Args       []string `yaml:"args"`
}

type kubeConfigData struct {
	IAMRoleArn                      *string
	ClusterCertificateAuthorityData *string
	ClusterEndpoint                 *string
	ClusterName                     *string
	ContextName                     string
	UserName                        string
}

func KubeConfigPath() (string, error) {
	homeDir := GetUserHomeDir()
	if len(homeDir) == 0 {
		return "", errors.New("Couldn't find user's home directory")
	}

	kubeConfigDir := path.Join(homeDir, ".kube")
	err := os.MkdirAll(kubeConfigDir, os.ModePerm)
	if err != nil {
		return "", err
	}

	kubeConfig := path.Join(kubeConfigDir, "config")
	return kubeConfig, nil

}
func GetKubeConfigData() ([]byte, error) {

	kubeConfig, err := KubeConfigPath()
	if err != nil {
		return nil, err
	}

	_, err = os.Stat(kubeConfig)
	if err != nil {
		return nil, nil
	}

	data, err := ioutil.ReadFile(kubeConfig)
	return data, err
}
func WriteKubeConfig(kd kubeConfigData) error {

	var isAdminConfig bool

	kConfig := KubeConfigType{}

	data, err := GetKubeConfigData()
	if err != nil {
		return err
	}
	if len(data) != 0 {
		err = yaml.Unmarshal(data, &kConfig)
		if err != nil {
			return err
		}

	}

	if !strings.HasSuffix(kd.UserName, "-admin") {
		isAdminConfig = false
	} else {
		isAdminConfig = true
	}

	kConfig.APIVersion = "v1"
	kConfig.Kind = "Config"

	newCluster := ClusterConfig{}
	newCluster.Name = *kd.ClusterName
	newCluster.Cluster = map[string]string{
		"server":                     *kd.ClusterEndpoint,
		"certificate-authority-data": *kd.ClusterCertificateAuthorityData,
	}

	var contextNamespace string

	if !isAdminConfig {
		contextNamespace = kd.UserName
	} else {
		contextNamespace = "default"
	}

	newClusterContext := ClusterContext{}
	newClusterContext.Name = kd.ContextName
	newClusterContext.Context = map[string]string{
		"cluster":   *kd.ClusterName,
		"user":      kd.UserName,
		"namespace": contextNamespace,
	}

	newClusterUser := ClusterUser{}
	newClusterUser.Name = kd.UserName

	newClusterUserExec := ClusterUserExecDetails{}
	newClusterUserExec.ApiVersion = "client.authentication.k8s.io/v1alpha1"
	newClusterUserExec.Command = "aws-iam-authenticator"
	newClusterUserExec.Args = []string{
		"token",
		"-i",
		*kd.ClusterName,
	}

	if !isAdminConfig {
		newClusterUserExec.Args = append(newClusterUserExec.Args, "-r")
		newClusterUserExec.Args = append(newClusterUserExec.Args, *kd.IAMRoleArn)
	}

	newClusterUser.User = map[string]ClusterUserExecDetails{
		"exec": newClusterUserExec,
	}

	var clusterPresent, contextPresent, userPresent bool
	for _, c := range kConfig.Clusters {
		if c.Cluster["server"] == newCluster.Cluster["server"] {
			clusterPresent = true
			break
		}
	}

	for _, c := range kConfig.Contexts {
		if c.Name == newClusterContext.Name && c.Context["cluster"] == newClusterContext.Context["cluster"] {
			contextPresent = true
			break
		}
	}

	for _, u := range kConfig.Users {
		if u.Name == newClusterUser.Name && u.User["exec"].Args[len(u.User["exec"].Args)-1] == newClusterUserExec.Args[len(newClusterUserExec.Args)-1] {
			userPresent = true
			break
		}
	}

	if !clusterPresent {
		kConfig.Clusters = append(kConfig.Clusters, newCluster)
	}

	if !contextPresent {
		kConfig.Contexts = append(kConfig.Contexts, newClusterContext)
	}

	if !userPresent {
		kConfig.Users = append(kConfig.Users, newClusterUser)
	}

	yamlData, err := yaml.Marshal(&kConfig)
	if err != nil {
		return err
	}

	kubeConfig, err := KubeConfigPath()
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(kubeConfig, yamlData, 0644)
	return err
}

//Write kubeconfig to file
func WriteKubeConfigToFile(clusterData *eks.DescribeClusterOutput, projectName string, environment string) error {

	d := kubeConfigData{}
	d.ClusterName = clusterData.Cluster.Name
	d.ClusterCertificateAuthorityData = clusterData.Cluster.CertificateAuthority.Data
	d.ClusterEndpoint = clusterData.Cluster.Endpoint

	// for non-admins
	if len(projectName) > 0 {
		if len(environment) == 0 {
			return errors.New("Must specify environment")
		}

		iamRoleArn := GetIAMRoleArnToAssume(projectName, environment)
		if iamRoleArn == nil {
			return fmt.Errorf("Unable to get IAM role: %s-%s-humans", projectName, environment)
		}

		d.IAMRoleArn = iamRoleArn
		projectEnv := fmt.Sprintf("%s-%s", projectName, environment)
		d.UserName = projectEnv
		d.ContextName = projectEnv

	} else {
		d.UserName = *d.ClusterName + "-admin"
		d.ContextName = *d.ClusterName + "-admin"
	}

	err := WriteKubeConfig(d)

	if err != nil {
		return err
	}

	fmt.Printf("Kubeconfig written")
	return nil
}

func GetKubeConfigContexts() []ClusterContext {
	kd, err := GetKubeConfigData()
	if err != nil {
		log.Fatal(err)
	}

	if len(kd) != 0 {
		kConfig := KubeConfigType{}

		err = yaml.Unmarshal(kd, &kConfig)
		if err != nil {
			log.Fatal(err)
		}

		return kConfig.Contexts

	}

	return nil
}

func SetKubeConfigCurrentContext(contextName string) error {
	kd, err := GetKubeConfigData()
	if err != nil {
		return err
	}

	if len(kd) != 0 {
		kConfig := KubeConfigType{}

		err = yaml.Unmarshal(kd, &kConfig)
		if err != nil {
			return err
		}

		for _, c := range kConfig.Contexts {
			if c.Name == contextName {
				kConfig.CurrentContext = c.Name
				break
			}
		}

		kubeConfig, err := KubeConfigPath()
		if err != nil {
			return err
		}
		yamlData, err := yaml.Marshal(&kConfig)
		if err != nil {
			return err
		}
		err = ioutil.WriteFile(kubeConfig, yamlData, 0644)
		return err

	} else {
		return errors.New("Empty kube config file")
	}

	return nil
}
