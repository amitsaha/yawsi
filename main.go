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

import (
	"fmt"
	"os"
	"strings"

	"github.com/amitsaha/yawsi/cmd"
	"github.com/spf13/pflag"
)

// Function copied from github.com/spf13/cobra
func nonCompletableFlag(flag *pflag.Flag) bool {
	return flag.Hidden || len(flag.Deprecated) > 0
}

func main() {
	// handle bash autocomplete
	if len(os.Getenv("COMP_LINE")) != 0 {
		if len(os.Getenv("COMP_DEBUG")) != 0 {
			fmt.Printf("%#v\n", os.Getenv("COMP_LINE"))
		}
		compLine := strings.Split(os.Getenv("COMP_LINE"), " ")

		// we only handle auto complete, if we are the command
		// being invoked
		if compLine[0] == cmd.RootCmd.Name() {
			var cmdArgs []string
			if len(compLine) > 1 {
				cmdArgs = compLine[1:]
			} else {
				cmdArgs = compLine[0:]
			}
			c, _, _ := cmd.RootCmd.Find(cmdArgs)
			suggestions := c.SuggestionsFor(cmdArgs[len(cmdArgs)-1])
			for _, s := range suggestions {
				fmt.Printf("%s\n", s)
			}

			localNonPersistentFlags := c.LocalNonPersistentFlags()
			c.NonInheritedFlags().VisitAll(func(flag *pflag.Flag) {
				if nonCompletableFlag(flag) {
					return
				}
				if strings.HasPrefix(fmt.Sprintf("--%s", flag.Name), cmdArgs[len(cmdArgs)-1]) {
					fmt.Printf("--%s\n", flag.Name)
				}
				if localNonPersistentFlags.Lookup(flag.Name) != nil {
					if strings.HasPrefix(fmt.Sprintf("--%s", flag.Name), cmdArgs[len(cmdArgs)-1]) {
						fmt.Printf("--%s\n", flag.Name)
					}
				}
			})

			c.InheritedFlags().VisitAll(func(flag *pflag.Flag) {
				if nonCompletableFlag(flag) {
					return
				}
				if strings.HasPrefix(flag.Name, cmdArgs[len(cmdArgs)-1]) {
					fmt.Printf("--%s\n", flag.Name)
				}
				if len(flag.Shorthand) > 0 {
					if strings.HasPrefix(flag.Name, cmdArgs[len(cmdArgs)-1]) {
						fmt.Printf("--%s\n", flag.Name)
					}
				}
			})

		}

	} else {
		cmd.Execute()
	}
}
