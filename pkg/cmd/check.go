package cmd

import (
	"fmt"
	"math"
	"os"

	"github.com/consensys/go-corset/pkg/hir"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	log "github.com/sirupsen/logrus"
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
		// Configure log level
		if getFlag(cmd, "debug") {
			log.SetLevel(log.DebugLevel)
		}
		//
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
		cfg.batchSize = getUint(cmd, "batch")
		//
		stats := util.NewPerfStats()
		// TODO: support true ranges
		cfg.padding.Left = cfg.padding.Right
		// Parse constraints
		hirSchema = readSchemaFile(args[1])
		// Parse trace file
		columns := readTraceFile(args[0])
		//
		stats.Log("Reading trace files")
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
	// Size of constraint batches to execute in parallel
	batchSize uint
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
	trHIR, errsHIR := checkTrace("HIR", cols, hirSchema, cfg)
	//
	if errsHIR != nil {
		reportErrors(true, "HIR", trHIR, errsHIR, cfg)
		os.Exit(1)
	}
}

func checkTraceWithLoweringMir(cols []trace.RawColumn, hirSchema *hir.Schema, cfg checkConfig) {
	// Lower HIR => MIR
	mirSchema := hirSchema.LowerToMir()
	// Check trace
	trMIR, errsMIR := checkTrace("MIR", cols, mirSchema, cfg)
	//
	if errsMIR != nil {
		reportErrors(true, "MIR", trMIR, errsMIR, cfg)
		os.Exit(1)
	}
}

func checkTraceWithLoweringAir(cols []trace.RawColumn, hirSchema *hir.Schema, cfg checkConfig) {
	// Lower HIR => MIR
	mirSchema := hirSchema.LowerToMir()
	// Lower MIR => AIR
	airSchema := mirSchema.LowerToAir()
	trAIR, errsAIR := checkTrace("AIR", cols, airSchema, cfg)
	//
	if errsAIR != nil {
		reportErrors(true, "AIR", trAIR, errsAIR, cfg)
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
	trHIR, errsHIR := checkTrace("HIR", cols, hirSchema, cfg)
	trMIR, errsMIR := checkTrace("MIR", cols, mirSchema, cfg)
	trAIR, errsAIR := checkTrace("AIR", cols, airSchema, cfg)
	//
	if errsHIR != nil || errsMIR != nil || errsAIR != nil {
		reportErrors(true, "HIR", trHIR, errsHIR, cfg)
		reportErrors(true, "MIR", trMIR, errsMIR, cfg)
		reportErrors(true, "AIR", trAIR, errsAIR, cfg)
		os.Exit(1)
	}
}

func checkTrace(ir string, cols []trace.RawColumn, schema sc.Schema, cfg checkConfig) (trace.Trace, []error) {
	builder := sc.NewTraceBuilder(schema).Expand(cfg.expand).Parallel(cfg.parallelExpansion)
	//
	for n := cfg.padding.Left; n <= cfg.padding.Right; n++ {
		stats := util.NewPerfStats()
		tr, errs := builder.Padding(n).Build(cols)

		stats.Log("Expanding trace columns")
		// Check for errors
		if tr == nil || (cfg.strict && len(errs) > 0) {
			return tr, errs
		} else if len(errs) > 0 {
			reportErrors(false, ir, tr, errs, cfg)
		}
		// Validate trace.
		stats = util.NewPerfStats()
		//
		if err := validationCheck(tr, schema); err != nil {
			return tr, []error{err}
		}
		// Check trace.
		stats.Log("Validating trace")
		stats = util.NewPerfStats()
		//
		if err := sc.Accepts(cfg.batchSize, schema, tr); err != nil {
			return tr, []error{err}
		}

		stats.Log("Checking constraints")
	}
	// Done
	return nil, nil
}

// Validate that values held in trace columns match the expected type.  This is
// really a sanity check that the trace is not malformed.
func validationCheck(tr trace.Trace, schema sc.Schema) error {
	var err error

	schemaCols := schema.Columns()
	// Construct a communication channel for errors.
	c := make(chan error, tr.Width())
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
		go func() {
			// Send outcome back
			c <- validateColumn(colType, col, mod)
		}()
	}
	// Collect up all the results
	for i := uint(0); i < tr.Width(); i++ {
		// Read from channel
		if e := <-c; e != nil {
			err = e
		}
	}
	// Done
	return err
}

// Validate that all elements of a given column are within the given type.
func validateColumn(colType sc.Type, col trace.Column, mod sc.Module) error {
	for j := 0; j < int(col.Data().Len()); j++ {
		jth := col.Get(j)
		if !colType.Accept(jth) {
			qualColName := trace.QualifiedColumnName(mod.Name(), col.Name())
			return fmt.Errorf("row %d of column %s is out-of-bounds (%s)", j, qualColName, jth.String())
		}
	}
	// success
	return nil
}

func reportErrors(error bool, ir string, tr trace.Trace, errs []error, cfg checkConfig) {
	if cfg.report && tr != nil {
		trace.PrintTrace(tr)
	}
	// Construct set to ensure deduplicate errors
	set := make(map[string]bool, len(errs))
	//
	for _, err := range errs {
		key := fmt.Sprintf("%s (%s)", err, ir)
		set[key] = true
	}
	// Report each one
	for e := range set {
		if error {
			log.Errorln(e)
		} else {
			log.Warnln(e)
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
	checkCmd.Flags().BoolP("debug", "d", false, "report debug logs")
	checkCmd.Flags().BoolP("quiet", "q", false, "suppress output (e.g. warnings)")
	checkCmd.Flags().Bool("sequential", false, "perform sequential trace expansion")
	checkCmd.Flags().Uint("padding", 0, "specify amount of (front) padding to apply")
	checkCmd.Flags().UintP("batch", "b", math.MaxUint, "specify batch size for constraint checking")
	checkCmd.Flags().Int("spillage", -1,
		"specify amount of splillage to account for (where -1 indicates this should be inferred)")
}
