package cmd

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"path"
	"reflect"
	"text/template"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"
)

var subnetName string

type naclRulesSpec struct {
	SubnetName string `toml:"subnet_name"`
	Rules      []naclRule
}

type naclRule struct {
	NetworkACLID string `tf:"network_acl_id"`
	Egress       bool   `toml:"egress" tf:"egress" tf_type:"bool"`
	RuleNo       int64  `toml:"rule_no" tf:"rule_number" tf_type:"int"`
	RuleAction   string `toml:"rule_action" tf:"rule_action"`
	CidrBlock    string `toml:"cidr_block" tf:"cidr_block"`
	Protocol     string `toml:"protocol" tf:"protocol"`
	FromPort     int64  `toml:"from_port" tf:"from_port" tf_type:"int"`
	ToPort       int64  `toml:"to_port" tf:"to_port" tf_type:"int"`
}

func (r *naclRule) Validate() (bool, error) {
	if r.RuleAction != "allow" && r.RuleAction != "deny" {
		return false, errors.New("Invalid rule_action specified")
	}
	if r.RuleNo > 32767 {
		return false, errors.New("Rule number must be < 32767 for IPv4 addresses")
	}
	_, _, err := net.ParseCIDR(r.CidrBlock)
	if err != nil {
		return false, err
	}

	return true, nil
}

func getResourceName(rule naclRule) string {
	if rule.Egress {
		return fmt.Sprintf("rule_%s_egress_%v", subnetName, rule.RuleNo)
	} else {
		return fmt.Sprintf("rule_%s_ingress_%v", subnetName, rule.RuleNo)
	}
}

func renderRule(rule naclRule) string {
	var rendered string
	val := reflect.ValueOf(&rule).Elem()
	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		typeField := val.Type().Field(i)
		tag := typeField.Tag
		tfAttribute := tag.Get("tf")
		tfDataType := tag.Get("tf_type")
		if tfDataType == "int" || tfDataType == "bool" {
			rendered += fmt.Sprintf("    %s = %v\n", tfAttribute, valueField.Interface())
		} else {
			rendered += fmt.Sprintf("    %s = \"%v\"\n", tfAttribute, valueField.Interface())
		}
	}
	return rendered
}

func generateTfNaclRules(naclRules []naclRule) {

	funcMap := template.FuncMap{
		"getResourceName": getResourceName,
		"renderRule":      renderRule,
	}
	tmpl := template.New("aws_network_acl").Funcs(funcMap)
	tmpl, err := tmpl.Parse(`
# This is a generated file, do not hand edit. See README at the
# root of the repository

{{ range . }}resource "aws_network_acl_rule" "{{. | getResourceName}}" {

{{ . | renderRule }}

}
{{end}}
`)
	if err != nil {
		log.Fatal("Error parsing template: ", err)

	}

	outputFile, err := os.Create(path.Join(".", fmt.Sprintf("%s_nacls.tf", subnetName)))
	if err != nil {
		log.Fatal("Error creating output file", err)
	}

	err = tmpl.Execute(outputFile, naclRules)
	if err != nil {
		log.Fatal("Error executing template: ", err)

	}
}

var tfNaclCmd = &cobra.Command{
	Use:   "generate-nacl-rules",
	Short: "General NACL rules",
	Run: func(cmd *cobra.Command, args []string) {

		var naclRules naclRulesSpec
		if _, err := toml.DecodeFile(naclSpecPath, &naclRules); err != nil {
			fmt.Println("Error", err)
			return
		}
		subnetName = naclRules.SubnetName
		// We use the index only pattern here so that
		// we can modify the array elements to insert the
		// static value for NetworkAclID
		for i := range naclRules.Rules {
			if result, err := naclRules.Rules[i].Validate(); !result {
				log.Fatalf("Invalid rule specification: %#v\n%v\n", naclRules.Rules[i], err)
			}
			naclRules.Rules[i].NetworkACLID = fmt.Sprintf(`${lookup(local.network_acl_ids_map, "%s")}`, subnetName)
		}
		generateTfNaclRules(naclRules.Rules)
	},
}

var naclSpecPath string

func init() {
	tfCmd.AddCommand(tfNaclCmd)
	tfNaclCmd.Flags().StringVarP(&naclSpecPath, "nacl-spec", "", "", "NACL TOML spec")
}
