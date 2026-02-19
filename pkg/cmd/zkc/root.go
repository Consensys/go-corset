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
package zkc

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/spf13/cobra"
)

// Version is filled when building with make, but *not* when installing via "go
// install".
var Version string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "zkc",
	Short: "A compiler for the ZkC language.",
	Long:  "A compiler (and general toolbox) for the ZkC language.",
	Run: func(cmd *cobra.Command, args []string) {
		if GetFlag(cmd, "version") {
			fmt.Print("zkc ")
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

// FieldAgnosticCmd represents a command to be executed for a given field.
type FieldAgnosticCmd struct {
	Field    field.Config
	Function func(*cobra.Command, []string)
}

// Run a field agnostic top-level command.
func runFieldAgnosticCmd(cmd *cobra.Command, args []string, cmds []FieldAgnosticCmd) {
	var (
		fieldName = GetString(cmd, "field")
		// Field configuration
		config = field.GetConfig(fieldName)
	)
	// Sanity check
	if config == nil {
		fmt.Printf("unknown field \"%s\"\n", fieldName)
		os.Exit(3)
	}
	// Find command to dispatch
	for _, c := range cmds {
		if c.Field == *config {
			// Match
			c.Function(cmd, args)
			// Done
			return
		}
	}
	//
	fmt.Printf("field %s unsupported for command '%s'\n", fieldName, cmd.Name())
	os.Exit(2)
}

func init() {
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "increase logging verbosity")
	rootCmd.PersistentFlags().Bool("vectorize", true, "Apply instruction vectorization")
	rootCmd.PersistentFlags().String("field", "BLS12_377", "prime field to use throughout")
}
