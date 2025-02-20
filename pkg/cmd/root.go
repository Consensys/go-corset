// Copyright Consensys Software Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with
// the License. You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
package cmd

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/spf13/cobra"
)

// Version is filled when building with make, but *not* when installing via "go
// install".
var Version string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "go-corset",
	Short: "A compiler for the Corset language.",
	Long:  "A compiler (and general toolbox) for the Corset language.",
	Run: func(cmd *cobra.Command, args []string) {
		if GetFlag(cmd, "version") {
			fmt.Print("go-corset ")
			if Version != "" {
				// Built via "make"
				fmt.Printf("%s", Version)
			} else if info, ok := debug.ReadBuildInfo(); ok {
				// Built via "go install"
				fmt.Printf("%s", info.Main.Version)
			} else {
				// Unknown, perhaps "go run"
				fmt.Printf("(unknown version)")
			}
			fmt.Println()
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.Flags().Bool("version", false, "Report version of this executable")
	rootCmd.PersistentFlags().Bool("legacy", true, "use legacy binary format")
	rootCmd.PersistentFlags().Bool("strict", true, "use strict typing mode")
	rootCmd.PersistentFlags().Bool("no-stdlib", false, "prevent standard library from being included")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "increase logging verbosity")
	rootCmd.PersistentFlags().UintP("opt", "O", 1, "set optimisation level")
}
