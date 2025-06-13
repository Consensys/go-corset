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
	"math"
	"os"

	"github.com/consensys/go-corset/pkg/cmd/debug"
	cmd_util "github.com/consensys/go-corset/pkg/cmd/util"
	"github.com/consensys/go-corset/pkg/corset"
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

func printAttributes(schemas cmd_util.SchemaStack) {
	binfile := schemas.BinaryFile()
	// Print attributes
	for _, attr := range binfile.Attributes {
		fmt.Printf("attribute \"%s\":\n", attr.AttributeName())
		//
		if attr.AttributeName() == "CorsetSourceMap" {
			printSourceMap(attr.(*corset.SourceMap))
		}
	}
}

func printSourceMap(srcmap *corset.SourceMap) {
	printSourceMapModule(1, srcmap.Root)
}

func printSourceMapModule(indent uint, module corset.SourceModule) {
	//
	fmt.Println()
	printIndent(indent)
	//
	if module.Virtual {
		fmt.Printf("virtual ")
	}
	//
	fmt.Printf("module \"%s\":\n", module.Name)
	//
	indent++
	// Print constants
	for _, c := range module.Constants {
		printIndent(indent)
		//
		if c.Extern {
			fmt.Printf("extern\t")
		} else {
			fmt.Printf("const\t")
		}
		//
		if c.Bitwidth != math.MaxUint {
			fmt.Printf("u%d ", c.Bitwidth)
		}
		//
		fmt.Printf("%s = %s\n", c.Name, &c.Value)
	}
	// Print columns
	for _, c := range module.Columns {
		printIndent(indent)
		fmt.Printf("u%d\t%s\t[", c.Bitwidth, c.Name)
		//
		for i, a := range sourceColumnAttrs(c) {
			if i == 0 {
				fmt.Print(a)
			} else {
				fmt.Printf(", %s", a)
			}
		}

		fmt.Println("]")
	}
	// Print submodules
	for _, m := range module.Submodules {
		printSourceMapModule(indent, m)
	}
}

func sourceColumnAttrs(col corset.SourceColumn) []string {
	var attrs []string
	//
	attrs = append(attrs, fmt.Sprintf("r%d", col.Register))
	//
	if col.Multiplier != 1 {
		attrs = append(attrs, fmt.Sprintf("×%d", col.Multiplier))
	}
	//
	if col.Computed {
		attrs = append(attrs, "computed")
	}
	//
	if col.MustProve {
		attrs = append(attrs, "proved")
	}
	//
	switch col.Display {
	case corset.DISPLAY_HEX:
		attrs = append(attrs, "hex")
	case corset.DISPLAY_DEC:
		attrs = append(attrs, "dec")
	case corset.DISPLAY_BYTES:
		attrs = append(attrs, "bytes")
	}
	//
	return attrs
}

func printSpillage(schemas cmd_util.SchemaStack, defensive bool) {
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

func printBinaryFileHeader(schemas cmd_util.SchemaStack) {
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
