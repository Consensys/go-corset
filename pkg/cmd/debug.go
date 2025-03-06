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
	"reflect"

	"github.com/consensys/go-corset/pkg/air"
	"github.com/consensys/go-corset/pkg/binfile"
	"github.com/consensys/go-corset/pkg/corset"
	"github.com/consensys/go-corset/pkg/hir"
	"github.com/consensys/go-corset/pkg/mir"
	"github.com/consensys/go-corset/pkg/schema"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/assignment"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/iter"
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
		stats := GetFlag(cmd, "stats")
		attrs := GetFlag(cmd, "attributes")
		stdlib := !GetFlag(cmd, "no-stdlib")
		debug := GetFlag(cmd, "debug")
		legacy := GetFlag(cmd, "legacy")
		metadata := GetFlag(cmd, "metadata")
		constants := GetFlag(cmd, "constants")
		externs := GetStringArray(cmd, "set")
		// Parse constraints
		binfile := ReadConstraintFiles(stdlib, debug, legacy, args)
		// Apply any user-specified values for externalised constants.
		applyExternOverrides(externs, binfile)
		// Print constant info (if requested)
		if constants {
			printExternalisedConstants(binfile)
		}
		// Print meta-data (if requested)
		if metadata {
			printBinaryFileHeader(&binfile.Header)
		}
		// Print stats (if requested)
		if stats {
			printStats(&binfile.Schema, hir, mir, air, optConfig)
		}
		// Print embedded attributes (if requested
		if attrs {
			printAttributes(binfile.Attributes)
		}
		//
		if !stats && !attrs {
			printSchemas(&binfile.Schema, hir, mir, air, optConfig)
		}
	},
}

func init() {
	rootCmd.AddCommand(debugCmd)
	debugCmd.Flags().Bool("air", false, "Print constraints at AIR level")
	debugCmd.Flags().Bool("attributes", false, "Print attribute information")
	debugCmd.Flags().Bool("constants", false, "Print information about externalised constants")
	debugCmd.Flags().Bool("debug", false, "enable debugging constraints")
	debugCmd.Flags().Bool("hir", false, "Print constraints at HIR level")
	debugCmd.Flags().Bool("metadata", false, "Print embedded metadata")
	debugCmd.Flags().Bool("mir", false, "Print constraints at MIR level")
	debugCmd.Flags().Bool("stats", false, "Print summary information")
	debugCmd.Flags().StringArrayP("set", "s", []string{}, "set value of externalised constant.")
}

func printSchemas(hirSchema *hir.Schema, hir bool, mir bool, air bool, optConfig mir.OptimisationConfig) {
	if hir {
		printSchema(hirSchema)
	}

	if mir {
		printSchema(hirSchema.LowerToMir())
	}

	if air {
		printSchema(hirSchema.LowerToMir().LowerToAir(optConfig))
	}
}

// Print out all declarations included in a given
func printSchema(schema schema.Schema) {
	for i := schema.Declarations(); i.HasNext(); {
		ith := i.Next()
		fmt.Println(ith.Lisp(schema).String(true))
	}

	for i := schema.Constraints(); i.HasNext(); {
		ith := i.Next()
		fmt.Println(ith.Lisp(schema).String(true))
	}
}

func printAttributes(attrs []binfile.Attribute) {
	for _, attr := range attrs {
		fmt.Printf("attribute \"%s\":\n", attr.AttributeName())
	}
}

func printExternalisedConstants(binf *binfile.BinaryFile) {
	fmt.Println("External constants:")
	// Sanity check debug information is available.
	srcmap, srcmap_ok := binfile.GetAttribute[*corset.SourceMap](binf)
	//
	if !srcmap_ok {
		fmt.Println("\t(no information available)")
		return
	}
	//
	printExternalisedModuleConstants(1, srcmap.Root)
}

func printExternalisedModuleConstants(indent uint, mod corset.SourceModule) {
	first := true
	// print constants in this module.
	for _, c := range mod.Constants {
		if c.Extern {
			if first && mod.Name != "" {
				printIndent(indent)
				fmt.Printf("%s:\n", mod.Name)
				//
				indent++
			}
			//
			printIndent(indent)
			//
			if c.DataType != nil {
				fmt.Printf("%s (%s): %s\n", c.Name, c.DataType.String(), c.Value.String())
			} else {
				fmt.Printf("%s: %s\n", c.Name, c.Value.String())
			}
			//
			first = false
		}
	}
	// traverse submodules
	for _, m := range mod.Submodules {
		printExternalisedModuleConstants(indent, m)
	}
}

func printStats(hirSchema *hir.Schema, hir bool, mir bool, air bool, optConfig mir.OptimisationConfig) {
	schemas := make([]schema.Schema, 0)
	mirSchema := hirSchema.LowerToMir()
	airSchema := mirSchema.LowerToAir(optConfig)
	// Construct columns
	if hir {
		schemas = append(schemas, hirSchema)
	}

	if mir {
		schemas = append(schemas, mirSchema)
	}

	if air {
		schemas = append(schemas, airSchema)
	}
	//
	n := 1 + uint(len(schemas))
	m := uint(len(schemaSummarisers))
	tbl := util.NewTablePrinter(n, m)
	// Go!
	for i := uint(0); i < m; i++ {
		ith := schemaSummarisers[i]
		row := make([]string, n)
		row[0] = ith.name

		for j := 0; j < len(schemas); j++ {
			count := ith.summary(schemas[j])
			row[j+1] = fmt.Sprintf("%d", count)
		}

		tbl.SetRow(i, row...)
	}
	//
	tbl.SetMaxWidths(64)
	tbl.Print()
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

// ============================================================================
// Schema Summarisers
// ============================================================================

type schemaSummariser struct {
	name    string
	summary func(schema.Schema) int
}

var schemaSummarisers []schemaSummariser = []schemaSummariser{
	// Constraints
	constraintCounter("Constraints", vanishingConstraints...),
	constraintCounter("Lookups", lookupConstraints...),
	constraintCounter("Permutations", permutationConstraints...),
	constraintCounter("Range", rangeConstraints...),
	// Assignments
	assignmentCounter("Decompositions", reflect.TypeOf((*assignment.ByteDecomposition)(nil))),
	assignmentCounter("Committed Columns", reflect.TypeOf((*assignment.DataColumn)(nil))),
	assignmentCounter("Computed Columns", computedColumns...),
	assignmentCounter("Computation Columns", reflect.TypeOf((*assignment.Computation)(nil))),
	assignmentCounter("Interleavings", reflect.TypeOf((*assignment.Interleaving)(nil))),
	assignmentCounter("Lexicographic Orderings", reflect.TypeOf((*assignment.LexicographicSort)(nil))),
	assignmentCounter("Sorted Permutations", reflect.TypeOf((*assignment.SortedPermutation)(nil))),
	// Columns
	columnCounter(),
	columnWidthSummariser(1, 1),
	columnWidthSummariser(2, 4),
	columnWidthSummariser(5, 8),
	columnWidthSummariser(9, 16),
	columnWidthSummariser(17, 32),
	columnWidthSummariser(33, 64),
	columnWidthSummariser(65, 128),
	columnWidthSummariser(129, 256),
}

var vanishingConstraints = []reflect.Type{
	reflect.TypeOf((hir.VanishingConstraint)(nil)),
	reflect.TypeOf((mir.VanishingConstraint)(nil)),
	reflect.TypeOf((air.VanishingConstraint)(nil))}

var lookupConstraints = []reflect.Type{
	reflect.TypeOf((hir.LookupConstraint)(nil)),
	reflect.TypeOf((mir.LookupConstraint)(nil)),
	reflect.TypeOf((air.LookupConstraint)(nil))}

var rangeConstraints = []reflect.Type{
	reflect.TypeOf((hir.RangeConstraint)(nil)),
	reflect.TypeOf((mir.RangeConstraint)(nil)),
	reflect.TypeOf((air.RangeConstraint)(nil))}

var permutationConstraints = []reflect.Type{
	// permutation constraints only exist at AIR level
	reflect.TypeOf((air.PermutationConstraint)(nil))}

var computedColumns = []reflect.Type{
	// permutation constraints only exist at AIR level
	reflect.TypeOf((*assignment.ComputedColumn)(nil))}

func constraintCounter(title string, types ...reflect.Type) schemaSummariser {
	return schemaSummariser{
		name: title,
		summary: func(schema sc.Schema) int {
			sum := 0
			for _, t := range types {
				sum += typeOfCounter(schema.Constraints(), t)
			}
			return sum
		},
	}
}

func assignmentCounter(title string, types ...reflect.Type) schemaSummariser {
	return schemaSummariser{
		name: title,
		summary: func(schema sc.Schema) int {
			sum := 0
			for _, t := range types {
				sum += typeOfCounter(schema.Declarations(), t)
			}
			return sum
		},
	}
}

func typeOfCounter[T any](iter iter.Iterator[T], dyntype reflect.Type) int {
	count := 0

	for iter.HasNext() {
		ith := iter.Next()
		if dyntype == reflect.TypeOf(ith) {
			count++
		}
	}

	return count
}

func columnCounter() schemaSummariser {
	return schemaSummariser{
		name: "Columns (all)",
		summary: func(sc schema.Schema) int {
			count := 0
			for i := sc.Columns(); i.HasNext(); {
				i.Next()
				count++
			}
			return count
		},
	}
}

func columnWidthSummariser(lowWidth uint, highWidth uint) schemaSummariser {
	return schemaSummariser{
		name: fmt.Sprintf("Columns (%d..%d bits)", lowWidth, highWidth),
		summary: func(sc schema.Schema) int {
			count := 0
			for i := sc.Columns(); i.HasNext(); {
				ith := i.Next()
				ithWidth := ith.DataType.BitWidth()
				if ithWidth >= lowWidth && ithWidth <= highWidth {
					count++
				}
			}
			return count
		},
	}
}
