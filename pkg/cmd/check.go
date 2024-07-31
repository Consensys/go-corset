package cmd

import (
	"fmt"
	"os"

	"github.com/consensys/go-corset/pkg/hir"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
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
		cfg.strict = !getFlag(cmd, "warn")
		cfg.quiet = getFlag(cmd, "quiet")
		cfg.padding.Right = getUint(cmd, "padding")
		cfg.parallelExpansion = !getFlag(cmd, "sequential")
		// TODO: support true ranges
		cfg.padding.Left = cfg.padding.Right
		// Parse constraints
		hirSchema = readSchemaFile(args[1])
		// Parse trace file
		columns := readTraceFile(args[0])
		// Go!
		checkTraceWithLowering(columns, hirSchema, cfg)
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
	// Suppress output (e.g. warnings)
	quiet bool
	// Specified whether strict checking is performed or not.  This is enabled
	// by default, and ensures the tool fails with an error in any unexpected or
	// unusual case.
	strict bool
	// Specifies whether or not to perform trace expansion.  Trace expansion is
	// not required when a "raw" trace is given which already includes all
	// implied columns.
	expand bool
	// Specifies whether or not to report details of the failure (e.g. for
	// debugging purposes).
	report bool
	// Perform trace expansion in parallel (or not)
	parallelExpansion bool
}

// Check a given trace is consistently accepted (or rejected) at the different
// IR levels.
func checkTraceWithLowering(cols []trace.RawColumn, schema *hir.Schema, cfg checkConfig) {
	if !cfg.hir && !cfg.mir && !cfg.air {
		// Process together
		checkTraceWithLoweringDefault(cols, schema, cfg)
	} else {
		// Process individually
		if cfg.hir {
			checkTraceWithLoweringHir(cols, schema, cfg)
		}

		if cfg.mir {
			checkTraceWithLoweringMir(cols, schema, cfg)
		}

		if cfg.air {
			checkTraceWithLoweringAir(cols, schema, cfg)
		}
	}
}

func checkTraceWithLoweringHir(cols []trace.RawColumn, hirSchema *hir.Schema, cfg checkConfig) {
	trHIR, errsHIR := checkTrace(cols, hirSchema, cfg)
	//
	if errsHIR != nil {
		reportErrors("HIR", trHIR, errsHIR, cfg)
		os.Exit(1)
	}
}

func checkTraceWithLoweringMir(cols []trace.RawColumn, hirSchema *hir.Schema, cfg checkConfig) {
	// Lower HIR => MIR
	mirSchema := hirSchema.LowerToMir()
	// Check trace
	trMIR, errsMIR := checkTrace(cols, mirSchema, cfg)
	//
	if errsMIR != nil {
		reportErrors("MIR", trMIR, errsMIR, cfg)
		os.Exit(1)
	}
}

func checkTraceWithLoweringAir(cols []trace.RawColumn, hirSchema *hir.Schema, cfg checkConfig) {
	// Lower HIR => MIR
	mirSchema := hirSchema.LowerToMir()
	// Lower MIR => AIR
	airSchema := mirSchema.LowerToAir()
	trAIR, errsAIR := checkTrace(cols, airSchema, cfg)
	//
	if errsAIR != nil {
		reportErrors("AIR", trAIR, errsAIR, cfg)
		os.Exit(1)
	}
}

// The default check allows one to compare all levels against each other and
// look for any discrepenacies.
func checkTraceWithLoweringDefault(cols []trace.RawColumn, hirSchema *hir.Schema, cfg checkConfig) {
	// Lower HIR => MIR
	mirSchema := hirSchema.LowerToMir()
	// Lower MIR => AIR
	airSchema := mirSchema.LowerToAir()
	//
	trHIR, errsHIR := checkTrace(cols, hirSchema, cfg)
	trMIR, errsMIR := checkTrace(cols, mirSchema, cfg)
	trAIR, errsAIR := checkTrace(cols, airSchema, cfg)
	//
	if errsHIR != nil || errsMIR != nil || errsAIR != nil {
		reportErrors("HIR", trHIR, errsHIR, cfg)
		reportErrors("MIR", trMIR, errsMIR, cfg)
		reportErrors("AIR", trAIR, errsAIR, cfg)
		os.Exit(1)
	}
}

func checkTrace(cols []trace.RawColumn, schema sc.Schema, cfg checkConfig) (trace.Trace, []error) {
	builder := sc.NewTraceBuilder(schema).Expand(cfg.expand).Parallel(cfg.parallelExpansion)
	//
	for n := cfg.padding.Left; n <= cfg.padding.Right; n++ {
		tr, errs := builder.Padding(n).Build(cols)
		// Check for errors
		if tr == nil || (cfg.strict && len(errs) > 0) {
			return tr, errs
		} else if len(errs) > 0 {
			reportWarnings(errs, cfg)
		}
		// Validate trace.  Observe that this is only done for
		if err := validationCheck(tr, schema); err != nil {
			return tr, []error{err}
		}
		// Check whether accepted or not.
		if err := sc.Accepts(schema, tr); err != nil {
			return tr, []error{err}
		}
	}
	// Done
	return nil, nil
}

// Validate that values held in trace columns match the expected type.  This is
// really a sanity check that the trace is not malformed.
func validationCheck(tr trace.Trace, schema sc.Schema) error {
	schemaCols := schema.Columns()
	// Check each column in turn
	for i := uint(0); i < tr.Width(); i++ {
		// Extract ith column
		col := tr.Column(i)
		// Extract schema for ith column
		scCol := schemaCols.Next()
		// Determine enclosing module
		mod := schema.Modules().Nth(scCol.Context().Module())
		// Extract type for ith column
		colType := scCol.Type()
		// Check elements
		for j := 0; j < int(tr.Height(scCol.Context())); j++ {
			jth := col.Get(j)
			if !colType.Accept(jth) {
				qualColName := trace.QualifiedColumnName(mod.Name(), col.Name())
				return fmt.Errorf("row %d of column %s is out-of-bounds (%s)", j, qualColName, jth)
			}
		}
	}
	// Done
	return nil
}

func reportErrors(ir string, tr trace.Trace, errs []error, cfg checkConfig) {
	for _, err := range errs {
		reportError(ir, tr, err, cfg)
	}
}

func reportError(ir string, tr trace.Trace, err error, cfg checkConfig) {
	if cfg.report && tr != nil {
		trace.PrintTrace(tr)
	}

	if err != nil {
		fmt.Printf("[ERROR] %s: %s\n", ir, err)
	} else {
		fmt.Printf("[ERROR] Trace should have been rejected at %s level.\n", ir)
	}
}

func reportWarnings(errs []error, cfg checkConfig) {
	if !cfg.quiet {
		for _, err := range errs {
			fmt.Printf("[WARNING] %s\n", err)
		}
	}
}

func init() {
	rootCmd.AddCommand(checkCmd)
	checkCmd.Flags().Bool("report", false, "report details of failure for debugging")
	checkCmd.Flags().Bool("raw", false, "assume input trace already expanded")
	checkCmd.Flags().Bool("hir", false, "check at HIR level")
	checkCmd.Flags().Bool("mir", false, "check at MIR level")
	checkCmd.Flags().Bool("air", false, "check at AIR level")
	checkCmd.Flags().BoolP("warn", "w", false, "report warnings instead of failing for certain errors"+
		"(e.g. unknown columns in the trace)")
	checkCmd.Flags().BoolP("quiet", "q", false, "suppress output (e.g. warnings)")
	checkCmd.Flags().Bool("sequential", false, "perform sequential trace expansion")
	checkCmd.Flags().Uint("padding", 0, "specify amount of (front) padding to apply")
	checkCmd.Flags().Int("spillage", -1,
		"specify amount of splillage to account for (where -1 indicates this should be inferred)")
}
