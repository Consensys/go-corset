package cmd

import (
	"fmt"
	"os"

	"github.com/consensys/go-corset/pkg/hir"
	"github.com/consensys/go-corset/pkg/table"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/spf13/cobra"
)

// computeCmd represents the compute command
var checkCmd = &cobra.Command{
	Use:   "check [flags] trace_file constraint_file",
	Short: "Check a given trace against a set of constraints.",
	Long: `Check a given trace against a set of constraints.
	Traces can be given either as JSON or binary lt files.
	Constraints can be given either as lisp or bin files.`,
	Run: func(cmd *cobra.Command, args []string) {
		var trace *table.ArrayTrace
		var hirSchema *hir.Schema
		var cfg checkConfig

		if len(args) != 2 {
			fmt.Println(cmd.UsageString())
			os.Exit(1)
		}
		cfg.air = getFlag(cmd, "air")
		cfg.mir = getFlag(cmd, "mir")
		cfg.hir = getFlag(cmd, "hir")
		cfg.expand = !getFlag(cmd, "raw")
		cfg.spillage = getInt(cmd, "spillage")
		cfg.padding.Right = getUint(cmd, "padding")
		// TODO: support true ranges
		cfg.padding.Left = cfg.padding.Right
		// Parse trace
		trace = readTraceFile(args[0])
		// Parse constraints
		hirSchema = readSchemaFile(args[1])
		// Go!
		checkTraceWithLowering(trace, hirSchema, cfg)
	},
}

// check config encapsulates certain parameters to be used when
// checking traces.
type checkConfig struct {
	// Performing checking at HIR level
	hir bool
	// Performing checking at MIR level
	mir bool
	// Performing checking at AIR level
	air bool
	// Determines how much spillage to account for.  This gives the user the
	// ability to override the inferred default.  A negative value indicates
	// this default should be used.
	spillage int
	// Determines how much padding to use
	padding util.Pair[uint, uint]
	// Specifies whether or not to perform trace expansion.  Trace expansion is
	// not required when a "raw" trace is given which already includes all
	// implied columns.
	expand bool
}

// Check a given trace is consistently accepted (or rejected) at the different
// IR levels.
func checkTraceWithLowering(tr *table.ArrayTrace, schema *hir.Schema, cfg checkConfig) {
	if !cfg.hir && !cfg.mir && !cfg.air {
		// Process together
		checkTraceWithLoweringDefault(tr, schema, cfg)
	} else {
		// Process individually
		if cfg.hir {
			checkTraceWithLoweringHir(tr, schema, cfg)
		}

		if cfg.mir {
			checkTraceWithLoweringMir(tr, schema, cfg)
		}

		if cfg.air {
			checkTraceWithLoweringAir(tr, schema, cfg)
		}
	}
}

func checkTraceWithLoweringHir(tr *table.ArrayTrace, hirSchema *hir.Schema, cfg checkConfig) {
	errHIR := checkTrace(tr, hirSchema, cfg)
	//
	if errHIR != nil {
		reportError(errHIR, "HIR")
		os.Exit(1)
	}
}

func checkTraceWithLoweringMir(tr *table.ArrayTrace, hirSchema *hir.Schema, cfg checkConfig) {
	// Lower HIR => MIR
	mirSchema := hirSchema.LowerToMir()
	// Check trace
	errMIR := checkTrace(tr, mirSchema, cfg)
	//
	if errMIR != nil {
		reportError(errMIR, "MIR")
		os.Exit(1)
	}
}

func checkTraceWithLoweringAir(tr *table.ArrayTrace, hirSchema *hir.Schema, cfg checkConfig) {
	// Lower HIR => MIR
	mirSchema := hirSchema.LowerToMir()
	// Lower MIR => AIR
	airSchema := mirSchema.LowerToAir()
	errAIR := checkTrace(tr, airSchema, cfg)
	//
	if errAIR != nil {
		reportError(errAIR, "AIR")
		os.Exit(1)
	}
}

// The default check allows one to compare all levels against each other and
// look for any discrepenacies.
func checkTraceWithLoweringDefault(tr *table.ArrayTrace, hirSchema *hir.Schema, cfg checkConfig) {
	// Lower HIR => MIR
	mirSchema := hirSchema.LowerToMir()
	// Lower MIR => AIR
	airSchema := mirSchema.LowerToAir()
	//
	errHIR := checkTrace(tr, hirSchema, cfg)
	errMIR := checkTrace(tr, mirSchema, cfg)
	errAIR := checkTrace(tr, airSchema, cfg)
	//
	if errHIR != nil || errMIR != nil || errAIR != nil {
		strHIR := toErrorString(errHIR)
		strMIR := toErrorString(errMIR)
		strAIR := toErrorString(errAIR)
		// At least one error encountered.
		if strHIR == strMIR && strMIR == strAIR {
			fmt.Println(errHIR)
		} else {
			reportError(errHIR, "HIR")
			reportError(errMIR, "MIR")
			reportError(errAIR, "AIR")
		}

		os.Exit(1)
	}
}

func checkTrace(tr *table.ArrayTrace, schema table.Schema, cfg checkConfig) error {
	if cfg.expand {
		// Clone to prevent interefence with subsequent checks
		tr = tr.Clone()
		// Apply spillage
		if cfg.spillage >= 0 {
			// Apply user-specified spillage
			table.FrontPadWithZeros(uint(cfg.spillage), tr)
		} else {
			// Apply default inferred spillage
			table.FrontPadWithZeros(schema.RequiredSpillage(), tr)
		}
		// Expand trace
		if err := schema.ExpandTrace(tr); err != nil {
			return err
		}
	}
	// Check whether padding requested
	if cfg.padding.Left == 0 && cfg.padding.Right == 0 {
		// No padding requested.  Therefore, we can avoid a clone in this case.
		return schema.Accepts(tr)
	}
	// Apply padding
	for n := cfg.padding.Left; n <= cfg.padding.Right; n++ {
		// Prevent interference
		ptr := tr.Clone()
		// Apply padding
		schema.ApplyPadding(n, ptr)
		fmt.Println(ptr.String())
		// Check whether accepted or not.
		if err := schema.Accepts(ptr); err != nil {
			return err
		}
	}
	// Done
	return nil
}

func toErrorString(err error) string {
	if err == nil {
		return ""
	}

	return err.Error()
}

func reportError(err error, ir string) {
	if err != nil {
		fmt.Printf("%s: %s\n", ir, err)
	} else {
		fmt.Printf("Trace should have been rejected at %s level.\n", ir)
	}
}

func init() {
	rootCmd.AddCommand(checkCmd)
	checkCmd.Flags().Bool("raw", false, "assume input trace already expanded")
	checkCmd.Flags().Bool("hir", false, "check at HIR level")
	checkCmd.Flags().Bool("mir", false, "check at MIR level")
	checkCmd.Flags().Bool("air", false, "check at AIR level")
	checkCmd.Flags().Uint("padding", 0, "specify amount of (front) padding to apply")
	checkCmd.Flags().Int("spillage", -1,
		"specify amount of splillage to account for (where -1 indicates this should be inferred)")
}
