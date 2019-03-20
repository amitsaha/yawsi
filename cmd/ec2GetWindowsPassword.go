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

	"github.com/spf13/cobra"
)

// listAsgCmd represents the listAsg command
var getWindowsPassword = &cobra.Command{
	Use:   "get-windows-password",
	Short: "Get Windows Password",
	Run: func(cmd *cobra.Command, args []string) {
		log.Printf("%s\n", getWindowsPasswordHelper(args[0], KeyPath))
	},
	Args: cobra.ExactArgs(1),
}

var KeyPath string

func init() {
	ec2Cmd.AddCommand(getWindowsPassword)
	getWindowsPassword.Flags().StringVarP(&KeyPath, "key-path", "k", "", "Private Key to decrypt the password")
}
