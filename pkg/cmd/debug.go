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

	"github.com/consensys/go-corset/pkg/cmd/debug"
	cmd_util "github.com/consensys/go-corset/pkg/cmd/util"
	"github.com/consensys/go-corset/pkg/corset"
	"github.com/consensys/go-corset/pkg/util/field/bls12_377"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var debugCmd = &cobra.Command{
	Use:   "debug [flags] constraint_file",
	Short: "print constraints at various levels of expansion.",
	Long: `Print a given set of constraints at specific levels of
	expansion in order to debug them.  Constraints can be given
	either as lisp or bin files.`,
	Run: func(cmd *cobra.Command, args []string) {
		//
		if len(args) < 1 {
			fmt.Println(cmd.UsageString())
			os.Exit(1)
		}
		// Configure log level
		if GetFlag(cmd, "verbose") {
			log.SetLevel(log.DebugLevel)
		}
		stats := GetFlag(cmd, "stats")
		attrs := GetFlag(cmd, "attributes")
		metadata := GetFlag(cmd, "metadata")
		constants := GetFlag(cmd, "constants")
		spillage := GetFlag(cmd, "spillage")
		textWidth := GetUint(cmd, "textwidth")
		// Read in constraint files
		schemas := *getSchemaStack(cmd, SCHEMA_DEFAULT_MIR, args...)
		// Print constant info (if requested)
		if constants {
			debug.PrintExternalisedConstants(schemas)
		}
		// Print spillage info (if requested)
		if spillage {
			printSpillage(schemas, true)
		}
		// Print meta-data (if requested)
		if metadata {
			printBinaryFileHeader(schemas)
		}
		// Print stats (if requested)
		if stats {
			debug.PrintStats(schemas)
		}
		// Print embedded attributes (if requested
		if attrs {
			printAttributes(schemas)
		}
		//
		if !stats && !attrs {
			debug.PrintSchemas(schemas, textWidth)
		}
	},
}

func init() {
	rootCmd.AddCommand(debugCmd)
	debugCmd.Flags().Bool("attributes", false, "Print attribute information")
	debugCmd.Flags().Bool("constants", false, "Print information about externalised constants")
	debugCmd.Flags().Bool("metadata", false, "Print embedded metadata")
	debugCmd.Flags().Bool("stats", false, "Print summary information")
	debugCmd.Flags().Bool("spillage", false, "Print spillage information")
	debugCmd.Flags().Uint("textwidth", 130, "Set maximum textwidth to use")
}

func printAttributes(schemas cmd_util.SchemaStack[bls12_377.Element]) {
	binfile := schemas.BinaryFile()
	// Print attributes
	for _, attr := range binfile.Attributes {
		fmt.Printf("attribute \"%s\":\n", attr.AttributeName())
		//
		if attr.AttributeName() == "CorsetSourceMap" {
			debug.PrintSourceMap(attr.(*corset.SourceMap))
		}
	}
}

func printSpillage(schemas cmd_util.SchemaStack[bls12_377.Element], defensive bool) {
	// fmt.Println("Spillage:")
	// // Compute spillage for optimisation level
	// spillage := determineSpillage(&binf.Schema, defensive, optConfig)
	// // Define module ID
	// mid := uint(0)
	// // Iterate modules and print spillage
	// for i := uint(0); i < uint(len(spillage)); i++ {
	// 	name := binf.Schema.Module(i).Name()
	// 	//
	// 	if name == "" {
	// 		name = "<prelude>"
	// 	}
	// 	//
	// 	fmt.Printf("\t%s: %d\n", name, spillage[i])
	// 	//
	// 	mid++
	// }
	panic("todo")
}

func printBinaryFileHeader(schemas cmd_util.SchemaStack[bls12_377.Element]) {
	header := schemas.BinaryFile().Header
	//
	fmt.Printf("Format: %d.%d\n", header.MajorVersion, header.MinorVersion)
	// Attempt to parse metadata
	metadata, err := header.GetMetaData()
	//
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	} else if !metadata.IsEmpty() {
		fmt.Println("Metadata:")
		//
		printTypedMetadata(1, metadata)
	}
}
