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
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/spf13/cobra"
)

// getMetadataCmd represents the ec2 get metadata commond
var ec2GetMetadataCmd = &cobra.Command{
	Use:   "get-metadata",
	Short: "Get EC2 Metadata",
	Long:  `Get EC2 metadata`,
	Run: func(cmd *cobra.Command, args []string) {

    },
}


func init() {
	ec2Cmd.AddCommand(ec2GetMetadataCmd)
}
