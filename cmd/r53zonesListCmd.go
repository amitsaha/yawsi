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
	"github.com/spf13/cobra"
)

// listVpcsCmd represents the list vpc command
var listR53ZonesCmd = &cobra.Command{
	Use:   "list-zones",
	Short: "List Route53 zones",
	Long: `List the Route53 zones
	
	To list the VPCs alongwith some key information:

		$ yawsi r53 list-zones		
	
	`,
	Run: func(cmd *cobra.Command, args []string) {
		ListR53Zones()
	},
}

func init() {
	r53Cmd.AddCommand(listR53ZonesCmd)
	//listR53ZonesCmd.Flags().BoolVarP(&vpcDetails, "details", "", false, "Show VPC details")
}
