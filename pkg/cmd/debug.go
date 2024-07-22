package cmd

import (
	"fmt"
	"os"

	"github.com/consensys/go-corset/pkg/air"
	"github.com/consensys/go-corset/pkg/hir"
	"github.com/consensys/go-corset/pkg/mir"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/assignment"
	"github.com/consensys/go-corset/pkg/schema/constraint"
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
		hir := getFlag(cmd, "hir")
		mir := getFlag(cmd, "mir")
		air := getFlag(cmd, "air")
		stats := getFlag(cmd, "stats")
		// Parse constraints
		hirSchema := readSchemaFile(args[0])

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
}

func printSchemas(hirSchema *hir.Schema, hir bool, mir bool, air bool) {
	mirSchema := hirSchema.LowerToMir()
	airSchema := mirSchema.LowerToAir()

	if hir {
		printSchema(airSchema)
	}

	if mir {
		printSchema(airSchema)
	}

	if air {
		printSchema(airSchema)
	}
}

// Print out all declarations included in a given
func printSchema(schema schema.Schema) {
	for i := schema.Declarations(); i.HasNext(); {
		fmt.Println(i.Next())
	}

	for i := schema.Constraints(); i.HasNext(); {
		fmt.Println(i.Next())
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
			row[j+1] = ith.summary(schemas[j])
		}
		tbl.SetRow(i, row...)
	}
	//
	tbl.SetMaxWidth(64)
	tbl.Print()
}

// ============================================================================
// Schema Summarisers
// ============================================================================

type schemaSummariser struct {
	name    string
	summary func(schema.Schema) string
}

var schemaSummarisers []schemaSummariser = []schemaSummariser{
	// Constraints
	{"Constraints", vanishingSummariser},
	{"Lookups", lookupSummariser},
	{"Permutations", constraintSummariser[*constraint.PermutationConstraint]},
	{"Types", constraintSummariser[*constraint.TypeConstraint]},
	{"Ranges", constraintSummariser[*constraint.RangeConstraint]},
	// Assignments
	{"Decomposition", assignmentSummariser[*assignment.ByteDecomposition]},
	{"Computed Columns", computedColumnSummariser},
	{"Committed Columns", assignmentSummariser[*assignment.DataColumn]},
	{"Sorted Permutations", assignmentSummariser[*assignment.SortedPermutation]},
	{"Interleave", assignmentSummariser[*assignment.Interleaving]},
	{"Lexicographic Sort", assignmentSummariser[*assignment.LexicographicSort]},
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

func vanishingSummariser(sc schema.Schema) string {
	count := constraintCounter[air.VanishingConstraint](sc)
	count += constraintCounter[mir.VanishingConstraint](sc)
	count += constraintCounter[hir.VanishingConstraint](sc)

	return fmt.Sprintf("%d", count)
}

func lookupSummariser(sc schema.Schema) string {
	count := constraintCounter[air.LookupConstraint](sc)
	count += constraintCounter[mir.LookupConstraint](sc)
	count += constraintCounter[hir.LookupConstraint](sc)

	return fmt.Sprintf("%d", count)
}

func constraintSummariser[T any](sc schema.Schema) string {
	count := constraintCounter[T](sc)
	return fmt.Sprintf("%d", count)
}

func constraintCounter[T any](sc schema.Schema) int {
	count := 0

	for c := sc.Constraints(); c.HasNext(); {
		ith := c.Next()
		if _, ok := ith.(T); ok {
			count++
		}
	}

	return count
}

func computedColumnSummariser(sc schema.Schema) string {
	count := assignmentCounter[*assignment.ComputedColumn[air.Expr]](sc)
	count += assignmentCounter[*assignment.ComputedColumn[mir.Expr]](sc)
	count += assignmentCounter[*assignment.ComputedColumn[mir.Expr]](sc)

	return fmt.Sprintf("%d", count)
}

func assignmentSummariser[T any](sc schema.Schema) string {
	count := assignmentCounter[T](sc)
	return fmt.Sprintf("%d", count)
}

func assignmentCounter[T any](sc schema.Schema) int {
	count := 0

	for c := sc.Declarations(); c.HasNext(); {
		ith := c.Next()
		if _, ok := ith.(T); ok {
			count++
		}
	}

	return count
}

func columnWidthSummariser(lowWidth uint, highWidth uint) schemaSummariser {
	return schemaSummariser{
		name: fmt.Sprintf("Columns (%d..%d bits)", lowWidth, highWidth),
		summary: func(sc schema.Schema) string {
			count := 0
			for i := sc.Columns(); i.HasNext(); {
				ith := i.Next()
				ithWidth := ith.Type().BitWidth()
				if ithWidth >= lowWidth && ithWidth <= highWidth {
					count++
				}
			}
			return fmt.Sprintf("%d", count)
		},
	}
}
