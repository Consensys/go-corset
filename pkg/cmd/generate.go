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

	"github.com/consensys/go-corset/pkg/binfile"
	"github.com/consensys/go-corset/pkg/cmd/generate"
	"github.com/consensys/go-corset/pkg/corset"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate [flags] constraint_file(s)",
	Short: "generate suitable Java class(es) for integration.",
	Long:  `Generate suitable Java class(es) for integration with a Java-based tracer generator.`,
	Run: func(cmd *cobra.Command, args []string) {
		var (
			corsetConfig corset.CompilationConfig
			genInterface bool
			source       string
			err          error
		)
		// Configure log level
		if GetFlag(cmd, "verbose") {
			log.SetLevel(log.DebugLevel)
		}
		//
		corsetConfig.Stdlib = !GetFlag(cmd, "no-stdlib")
		corsetConfig.Legacy = GetFlag(cmd, "legacy")
		asmConfig := parseLoweringConfig(cmd)
		filename := GetString(cmd, "output")
		pkgname := GetString(cmd, "package")
		//
		if inteface := GetString(cmd, "interface"); inteface != "" {
			genInterface = true
			filename = inteface
		}
		// Parse constraints
		binf := ReadConstraintFiles(corsetConfig, asmConfig, args)
		// Sanity check debug information is available.
		srcmap, srcmap_ok := binfile.GetAttribute[*corset.SourceMap](binf)
		//
		if !srcmap_ok {
			fmt.Printf("constraints file(s) \"%s\" missing source map", args[1])
		} else if genInterface {
			source, err = generate.JavaTraceInterface(filename, pkgname, srcmap)
		} else {
			// NOTE: assume defensive padding is enabled.
			spillage := determineConservativeSpillage(true, &binf.Schema)
			// Generate appropriate Java source
			source, err = generate.JavaTraceClass(filename, pkgname, spillage, srcmap, binf)
		}
		// check for errors / write out file.
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		} else if err := os.WriteFile(filename, []byte(source), 0644); err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	},
}

//nolint:errcheck
func init() {
	rootCmd.AddCommand(generateCmd)
	generateCmd.Flags().StringP("output", "o", "Trace.java", "specify output file.")
	generateCmd.Flags().StringP("interface", "i", "", "generate interface file.")
	generateCmd.Flags().StringP("extend", "e", "Trace", "specify interface to extend or implement.")
	generateCmd.Flags().StringP("package", "p", "", "specify Java package.")
}
