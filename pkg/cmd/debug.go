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
	"github.com/consensys/go-corset/pkg/cmd/debug"
	"github.com/consensys/go-corset/pkg/corset"
	"github.com/consensys/go-corset/pkg/mir"
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
		var corsetConfig corset.CompilationConfig
		//
		if len(args) != 1 {
			fmt.Println(cmd.UsageString())
			os.Exit(1)
		}
		// Configure log level
		if GetFlag(cmd, "verbose") {
			log.SetLevel(log.DebugLevel)
		}
		optimisation := GetUint(cmd, "opt")
		// Set optimisation level
		if optimisation >= uint(len(mir.OPTIMISATION_LEVELS)) {
			fmt.Printf("invalid optimisation level %d\n", optimisation)
			os.Exit(2)
		}
		//
		optConfig := mir.OPTIMISATION_LEVELS[optimisation]
		hir := GetFlag(cmd, "hir")
		mir := GetFlag(cmd, "mir")
		air := GetFlag(cmd, "air")
		masm := GetFlag(cmd, "asm")
		uasm := GetFlag(cmd, "uasm")
		stats := GetFlag(cmd, "stats")
		attrs := GetFlag(cmd, "attributes")
		metadata := GetFlag(cmd, "metadata")
		constants := GetFlag(cmd, "constants")
		externs := GetStringArray(cmd, "set")
		spillage := GetFlag(cmd, "spillage")
		corsetConfig.Stdlib = !GetFlag(cmd, "no-stdlib")
		corsetConfig.Debug = GetFlag(cmd, "debug")
		corsetConfig.Legacy = GetFlag(cmd, "legacy")
		asmConfig := parseLoweringConfig(cmd)
		// Parse constraints
		if masm || uasm {
			// Read in the assembly program
			program, _ := ReadAssemblyProgram(args...)
			// Print it out
			debug.PrintAssemblyProgram(uasm, asmConfig, program)
		} else {
			binfile := ReadConstraintFiles(corsetConfig, asmConfig, args)
			// Apply any user-specified values for externalised constants.
			applyExternOverrides(externs, binfile)
			// Print constant info (if requested)
			if constants {
				debug.PrintExternalisedConstants(binfile)
			}
			// Print spillage info (if requested)
			if spillage {
				printSpillage(binfile, true, optConfig)
			}
			// Print meta-data (if requested)
			if metadata {
				printBinaryFileHeader(&binfile.Header)
			}
			// Print stats (if requested)
			if stats {
				debug.PrintStats(&binfile.Schema, hir, mir, air, optConfig)
			}
			// Print embedded attributes (if requested
			if attrs {
				printAttributes(binfile.Attributes)
			}
			//
			if !stats && !attrs {
				debug.PrintSchemas(&binfile.Schema, hir, mir, air, optConfig)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(debugCmd)
	debugCmd.Flags().Bool("air", false, "Print constraints at AIR level")
	debugCmd.Flags().Bool("asm", false, "Print constraints at ASM level")
	debugCmd.Flags().Bool("attributes", false, "Print attribute information")
	debugCmd.Flags().Bool("constants", false, "Print information about externalised constants")
	debugCmd.Flags().Bool("debug", false, "enable debugging constraints")
	debugCmd.Flags().Bool("hir", false, "Print constraints at HIR level")
	debugCmd.Flags().Bool("metadata", false, "Print embedded metadata")
	debugCmd.Flags().Bool("mir", false, "Print constraints at MIR level")
	debugCmd.Flags().Bool("stats", false, "Print summary information")
	debugCmd.Flags().Bool("spillage", false, "Print spillage information")
	debugCmd.Flags().StringArrayP("set", "S", []string{}, "set value of externalised constant.")
	debugCmd.Flags().Bool("uasm", false, "Print constraints at micro ASM level")
}

func printAttributes(attrs []binfile.Attribute) {
	for _, attr := range attrs {
		fmt.Printf("attribute \"%s\":\n", attr.AttributeName())
	}
}

func printSpillage(binf *binfile.BinaryFile, defensive bool, optConfig mir.OptimisationConfig) {
	fmt.Println("Spillage:")
	// Compute spillage for optimisation level
	spillage := determineSpillage(&binf.Schema, defensive, optConfig)
	// Define module ID
	mid := uint(0)
	// Iterate modules and print spillage
	for i := uint(0); i < uint(len(spillage)); i++ {
		name := binf.Schema.Modules().Nth(i).Name
		//
		if name == "" {
			name = "<prelude>"
		}
		//
		fmt.Printf("\t%s: %d\n", name, spillage[i])
		//
		mid++
	}
}

func printBinaryFileHeader(header *binfile.Header) {
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
