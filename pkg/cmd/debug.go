package cmd

import (
	"fmt"
	"os"

	"github.com/consensys/go-corset/pkg/air"
	"github.com/consensys/go-corset/pkg/hir"
	"github.com/consensys/go-corset/pkg/mir"
	"github.com/consensys/go-corset/pkg/schema"
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
		stats := getFlag(cmd, "stats")
		hir := getFlag(cmd, "hir")
		mir := getFlag(cmd, "mir")
		air := getFlag(cmd, "air")
		// Parse constraints
		hirSchema := readSchemaFile(args[0])
		mirSchema := hirSchema.LowerToMir()
		airSchema := mirSchema.LowerToAir()
		// Print constraints
		if hir {
			printSchema(hirSchema, stats)
		}
		if mir {
			printSchema(mirSchema, stats)
		}
		if air {
			printSchema(airSchema, stats)
		}
	},
}

// Print out all declarations included in a given
func printSchema(schema schema.Schema, stats bool) {
	dataColumns := 0
	permutations := 0
	vanishing := 0
	ranges := 0
	assertions := 0
	computations := 0
	// Print declarations in order of appearance.
	for i := 0; i < schema.Size(); i++ {
		ith := schema.GetDeclaration(i)
		fmt.Println(ith.String())
		// Count stats
		if isDataColumn(ith) {
			dataColumns++
		} else if isPermutation(ith) {
			permutations++
		} else if isVanishing(ith) {
			vanishing++
		} else if isRange(ith) {
			ranges++
		} else {
			computations++
		}
	}
	//
	if stats {
		fmt.Println("--")
		fmt.Printf("%d column(s), %d permutation(s), %d constraint(s), %d range(s), %d assertion(s) and %d computation(s).\n",
			dataColumns, permutations, vanishing, ranges, assertions, computations)
	}
}

func isDataColumn(d schema.Declaration) bool {
	if _, ok := d.(air.DataColumn); ok {
		return true
	} else if _, ok := d.(mir.DataColumn); ok {
		return true
	} else if _, ok := d.(hir.DataColumn); ok {
		return true
	}

	return false
}

func isPermutation(d schema.Declaration) bool {
	if _, ok := d.(air.Permutation); ok {
		return true
	} else if _, ok := d.(mir.Permutation); ok {
		return true
	} else if _, ok := d.(hir.Permutation); ok {
		return true
	}

	return false
}

func isVanishing(d schema.Declaration) bool {
	if _, ok := d.(air.VanishingConstraint); ok {
		return true
	} else if _, ok := d.(mir.VanishingConstraint); ok {
		return true
	} else if _, ok := d.(hir.VanishingConstraint); ok {
		return true
	}

	return false
}

func isRange(d schema.Declaration) bool {
	if _, ok := d.(*schema.RangeConstraint); ok {
		return true
	}

	return false
}

func init() {
	rootCmd.AddCommand(debugCmd)
	debugCmd.Flags().BoolP("stats", "s", false, "Report statistics")
	debugCmd.Flags().Bool("hir", false, "Print constraints at HIR level")
	debugCmd.Flags().Bool("mir", false, "Print constraints at MIR level")
	debugCmd.Flags().Bool("air", false, "Print constraints at AIR level")
}
