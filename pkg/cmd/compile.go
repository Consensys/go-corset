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
	"strings"

	"github.com/consensys/go-corset/pkg/corset"
	"github.com/consensys/go-corset/pkg/util/collection/typed"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var compileCmd = &cobra.Command{
	Use:   "compile [flags] constraint_file(s)",
	Short: "compile constraints into a binary package.",
	Long: `Compile a given set of constraint file(s) into a single binary package which can
	 be subsequently used without requiring a full compilation step.`,
	Run: func(cmd *cobra.Command, args []string) {
		var corsetConfig corset.CompilationConfig
		// Configure log level
		if GetFlag(cmd, "verbose") {
			log.SetLevel(log.DebugLevel)
		}
		//
		corsetConfig.Stdlib = !GetFlag(cmd, "no-stdlib")
		corsetConfig.Debug = GetFlag(cmd, "debug")
		corsetConfig.Legacy = GetFlag(cmd, "legacy")
		corsetConfig.Strict = GetFlag(cmd, "strict")
		output := GetString(cmd, "output")
		defines := GetStringArray(cmd, "define")
		// Parse constraints
		binfile := ReadConstraintFiles(corsetConfig, args)
		// Write metadata
		if err := binfile.Header.SetMetaData(buildMetadata(defines)); err != nil {
			fmt.Printf("error writing metadata: %s\n", err.Error())
			os.Exit(1)
		}
		// Serialise as a gob file.
		WriteBinaryFile(binfile, false, output)
	},
}

func buildMetadata(items []string) typed.Map {
	metadata := make(map[string]any)
	//
	for _, item := range items {
		split := strings.Split(item, "=")
		if len(split) != 2 {
			fmt.Printf("malformed definition \"%s\"\n", item)
			os.Exit(2)
		}
		//
		metadata[split[0]] = split[1]
	}
	//
	return typed.NewMap(metadata)
}

//nolint:errcheck
func init() {
	rootCmd.AddCommand(compileCmd)
	compileCmd.Flags().Bool("debug", false, "enable debugging constraints")
	compileCmd.Flags().StringP("output", "o", "a.bin", "specify output file.")
	compileCmd.Flags().StringArrayP("define", "D", []string{}, "define metadata attribute.")
	compileCmd.MarkFlagRequired("output")
}
