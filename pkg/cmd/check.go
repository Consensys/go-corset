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
		cfg.report = getFlag(cmd, "report")
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
	// Specifies whether or not to report details of the failure (e.g. for
	// debugging purposes).
	report bool
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
	trHIR, errHIR := checkTrace(tr, hirSchema, cfg)
	//
	if errHIR != nil {
		reportError("HIR", trHIR, errHIR, cfg)
		os.Exit(1)
	}
}

func checkTraceWithLoweringMir(tr *table.ArrayTrace, hirSchema *hir.Schema, cfg checkConfig) {
	// Lower HIR => MIR
	mirSchema := hirSchema.LowerToMir()
	// Check trace
	trMIR, errMIR := checkTrace(tr, mirSchema, cfg)
	//
	if errMIR != nil {
		reportError("MIR", trMIR, errMIR, cfg)
		os.Exit(1)
	}
}

func checkTraceWithLoweringAir(tr *table.ArrayTrace, hirSchema *hir.Schema, cfg checkConfig) {
	// Lower HIR => MIR
	mirSchema := hirSchema.LowerToMir()
	// Lower MIR => AIR
	airSchema := mirSchema.LowerToAir()
	trAIR, errAIR := checkTrace(tr, airSchema, cfg)
	//
	if errAIR != nil {
		reportError("AIR", trAIR, errAIR, cfg)
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
	trHIR, errHIR := checkTrace(tr, hirSchema, cfg)
	trMIR, errMIR := checkTrace(tr, mirSchema, cfg)
	trAIR, errAIR := checkTrace(tr, airSchema, cfg)
	//
	if errHIR != nil || errMIR != nil || errAIR != nil {
		strHIR := toErrorString(errHIR)
		strMIR := toErrorString(errMIR)
		strAIR := toErrorString(errAIR)
		// At least one error encountered.
		if strHIR == strMIR && strMIR == strAIR {
			fmt.Println(errHIR)
		} else {
			reportError("HIR", trHIR, errHIR, cfg)
			reportError("MIR", trMIR, errMIR, cfg)
			reportError("AIR", trAIR, errAIR, cfg)
		}

		os.Exit(1)
	}
}

func checkTrace(tr *table.ArrayTrace, schema table.Schema, cfg checkConfig) (table.Trace, error) {
	if cfg.expand {
		// Clone to prevent interefence with subsequent checks
		tr = tr.Clone()
		// Apply spillage
		if cfg.spillage >= 0 {
			// Apply user-specified spillage
			tr.Pad(uint(cfg.spillage))
		} else {
			// Apply default inferred spillage
			tr.Pad(schema.RequiredSpillage())
		}
		// Expand trace
		if err := schema.ExpandTrace(tr); err != nil {
			return tr, err
		}
	}
	// Check whether padding requested
	if cfg.padding.Left == 0 && cfg.padding.Right == 0 {
		// No padding requested.  Therefore, we can avoid a clone in this case.
		return tr, schema.Accepts(tr)
	}
	// Apply padding
	for n := cfg.padding.Left; n <= cfg.padding.Right; n++ {
		// Prevent interference
		ptr := tr.Clone()
		// Apply padding
		ptr.Pad(n)
		// Check whether accepted or not.
		if err := schema.Accepts(ptr); err != nil {
			return ptr, err
		}
	}
	// Done
	return nil, nil
}

func toErrorString(err error) string {
	if err == nil {
		return ""
	}

	return err.Error()
}

func reportError(ir string, tr table.Trace, err error, cfg checkConfig) {
	if cfg.report {
		table.PrintTrace(tr)
	}

	if err != nil {
		fmt.Printf("%s: %s\n", ir, err)
	} else {
		fmt.Printf("Trace should have been rejected at %s level.\n", ir)
	}
}

func init() {
	rootCmd.AddCommand(checkCmd)
	checkCmd.Flags().Bool("report", false, "report details of failure for debugging")
	checkCmd.Flags().Bool("raw", false, "assume input trace already expanded")
	checkCmd.Flags().Bool("hir", false, "check at HIR level")
	checkCmd.Flags().Bool("mir", false, "check at MIR level")
	checkCmd.Flags().Bool("air", false, "check at AIR level")
	checkCmd.Flags().Uint("padding", 0, "specify amount of (front) padding to apply")
	checkCmd.Flags().Int("spillage", -1,
		"specify amount of splillage to account for (where -1 indicates this should be inferred)")
}
