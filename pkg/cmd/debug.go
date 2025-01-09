package cmd

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/consensys/go-corset/pkg/hir"
	"github.com/consensys/go-corset/pkg/schema"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/util"
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
		hir := GetFlag(cmd, "hir")
		mir := GetFlag(cmd, "mir")
		air := GetFlag(cmd, "air")
		stats := GetFlag(cmd, "stats")
		stdlib := !GetFlag(cmd, "no-stdlib")
		debug := GetFlag(cmd, "debug")
		legacy := GetFlag(cmd, "legacy")
		// Parse constraints
		hirSchema := readSchema(stdlib, debug, legacy, args)
		// Print constraints
		if stats {
			printStats(hirSchema, hir, mir, air)
		} else {
			printSchemas(hirSchema, hir, mir, air)
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
	mirSchema := hirSchema.LowerToMir()
	airSchema := mirSchema.LowerToAir()

	if hir {
		printSchema(hirSchema)
	}

	if mir {
		printSchema(mirSchema)
	}

	if air {
		printSchema(airSchema)
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
	constraintCounter("Constraints", "*constraint.VanishingConstraint"),
	constraintCounter("Lookups", "*constraint.LookupConstraint"),
	constraintCounter("Permutations", "*constraint.PermutationConstraint"),
	constraintCounter("Types", "*constraint.TypeConstraint"),
	constraintCounter("Range", "*constraint.RangeConstraint"),
	// Assignments
	assignmentCounter("Decompositions", "*assignment.ByteDecomposition"),
	assignmentCounter("Computed Columns", "*assignment.ComputedColumn"),
	assignmentCounter("Committed Columns", "*assignment.DataColumn"),
	assignmentCounter("Interleavings", "*assignment.Interleaving"),
	assignmentCounter("Lexicographic Orderings", "*assignment.LexicographicSort"),
	assignmentCounter("Sorted Permutations", "*assignment.SortedPermutation"),
	// Column Width
	columnWidthSummariser(1, 1),
	columnWidthSummariser(2, 4),
	columnWidthSummariser(5, 8),
	columnWidthSummariser(9, 16),
	columnWidthSummariser(17, 32),
	columnWidthSummariser(33, 64),
	columnWidthSummariser(65, 128),
	columnWidthSummariser(129, 256),
}

func constraintCounter(title string, prefix string) schemaSummariser {
	return schemaSummariser{
		name: title,
		summary: func(schema sc.Schema) int {
			return typeOfCounter(schema.Constraints(), prefix)
		},
	}
}

func assignmentCounter(title string, prefix string) schemaSummariser {
	return schemaSummariser{
		name: title,
		summary: func(schema sc.Schema) int {
			return typeOfCounter(schema.Declarations(), prefix)
		},
	}
}

func typeOfCounter[T any](iter util.Iterator[T], prefix string) int {
	count := 0

	for iter.HasNext() {
		ith := iter.Next()
		if isTypeOf(ith, prefix) {
			count++
		}
	}

	return count
}

func isTypeOf(obj any, prefix string) bool {
	dyntype := reflect.TypeOf(obj)
	// Check whether dynanic type matches prefix
	return strings.HasPrefix(dyntype.String(), prefix)
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
