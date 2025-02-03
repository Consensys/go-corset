package cmd

import (
	"fmt"
	"os"
	"reflect"

	"github.com/consensys/go-corset/pkg/air"
	"github.com/consensys/go-corset/pkg/hir"
	"github.com/consensys/go-corset/pkg/mir"
	"github.com/consensys/go-corset/pkg/schema"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/assignment"
	"github.com/consensys/go-corset/pkg/util"
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
		//
		hir := GetFlag(cmd, "hir")
		mir := GetFlag(cmd, "mir")
		air := GetFlag(cmd, "air")
		stats := GetFlag(cmd, "stats")
		stdlib := !GetFlag(cmd, "no-stdlib")
		debug := GetFlag(cmd, "debug")
		legacy := GetFlag(cmd, "legacy")
		// Parse constraints
		binfile := readSchema(stdlib, debug, legacy, args)
		// Print constraints
		if stats {
			printStats(&binfile.Schema, hir, mir, air)
		} else {
			printSchemas(&binfile.Schema, hir, mir, air)
		}
	},
}

func init() {
	rootCmd.AddCommand(debugCmd)
	debugCmd.Flags().Bool("hir", false, "Print constraints at HIR level")
	debugCmd.Flags().Bool("mir", false, "Print constraints at MIR level")
	debugCmd.Flags().Bool("air", false, "Print constraints at AIR level")
	debugCmd.Flags().Bool("stats", false, "Print summary information")
	debugCmd.Flags().Bool("debug", false, "enable debugging constraints")
}

func printSchemas(hirSchema *hir.Schema, hir bool, mir bool, air bool) {
	if hir {
		printSchema(hirSchema)
	}

	if mir {
		printSchema(hirSchema.LowerToMir())
	}

	if air {
		printSchema(hirSchema.LowerToMir().LowerToAir())
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

func printStats(hirSchema *hir.Schema, hir bool, mir bool, air bool) {
	schemas := make([]schema.Schema, 0)
	mirSchema := hirSchema.LowerToMir()
	airSchema := mirSchema.LowerToAir()
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
	reflect.TypeOf((*assignment.ComputedColumn[air.Expr])(nil))}

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

func typeOfCounter[T any](iter util.Iterator[T], dyntype reflect.Type) int {
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
