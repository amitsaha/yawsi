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

package main

import "os"
import "fmt"
import "github.com/amitsaha/yawsi/cmd"
import "strings"

func main() {
	if len(os.Getenv("COMP_LINE")) != 0 {
		if len(os.Getenv("COMP_DEBUG")) != 0 {
			fmt.Printf("%#v\n", os.Args[1:])
		}
		// whose commands is being completed
		if os.Args[1] == "yawsi" {
			// no sub-commands yet: yawsi <TAB>
			if os.Args[2] == "" && os.Args[3] == "yawsi" {
				for _, cmd := range cmd.RootCmd.Commands() {
					fmt.Printf("%s\n", cmd.Use)
				}
			}

			// yawsi <valid-subcommand> <TAB>
			if len(os.Args[2]) == 0 && len(os.Args[3]) != 0 && os.Args[3] != "yawsi" {
				for _, cmd := range cmd.RootCmd.Commands() {
					if cmd.Use == os.Args[3] {
						for _, cmd := range cmd.Commands() {
							fmt.Printf("%s\n", cmd.Use)
						}
						for _, arg := range cmd.ValidArgs {
							fmt.Printf("%s\n", arg)
						}
						flags := cmd.NonInheritedFlags()
                                                fmt.Printf("Flags: %#v\n", flags)

					}
				}
			}
			// yawsi <valid-subcommand> <valid-subcommand/arg><TAB>
			if len(os.Args[2]) != 0 && len(os.Args[3]) != 0 && os.Args[3] != "yawsi" {
				for _, cmd := range cmd.RootCmd.Commands() {
					if cmd.Use == os.Args[3] {
						for _, cmd := range cmd.Commands() {
							if strings.HasPrefix(cmd.Use, os.Args[2]) || os.Args[2] == cmd.Use {
								fmt.Printf("%s\n", cmd.Use)
							}
						}
						for _, arg := range cmd.ValidArgs {
							fmt.Printf("%s\n", arg)
						}

					}
				}
			}

		}
	} else {
		cmd.Execute()
	}
}
