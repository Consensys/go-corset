package cmd

import (
	"errors"
	"fmt"
	"math"
	"os"

	"github.com/consensys/go-corset/pkg/hir"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/constraint"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// computeCmd represents the compute command
var checkCmd = &cobra.Command{
	Use:   "check [flags] trace_file constraint_file(s)",
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
		if GetFlag(cmd, "verbose") {
			log.SetLevel(log.DebugLevel)
		}
		legacy := GetFlag(cmd, "legacy")
		//
		cfg.air = GetFlag(cmd, "air")
		cfg.mir = GetFlag(cmd, "mir")
		cfg.hir = GetFlag(cmd, "hir")
		cfg.defensive = GetFlag(cmd, "defensive")
		cfg.expand = !GetFlag(cmd, "raw")
		cfg.report = GetFlag(cmd, "report")
		cfg.reportPadding = GetUint(cmd, "report-context")
		cfg.reportCellWidth = GetUint(cmd, "report-cellwidth")
		cfg.spillage = GetInt(cmd, "spillage")
		cfg.strict = !GetFlag(cmd, "warn")
		cfg.stdlib = !GetFlag(cmd, "no-stdlib")
		cfg.debug = GetFlag(cmd, "debug")
		cfg.quiet = GetFlag(cmd, "quiet")
		cfg.padding.Right = GetUint(cmd, "padding")
		cfg.parallelExpansion = !GetFlag(cmd, "sequential")
		cfg.batchSize = GetUint(cmd, "batch")
		cfg.ansiEscapes = GetFlag(cmd, "ansi-escapes")
		// TODO: support true ranges
		cfg.padding.Left = cfg.padding.Right
		if !cfg.hir && !cfg.mir && !cfg.air {
			// If IR not specified default to running all.
			cfg.hir, cfg.mir, cfg.air = true, true, true
		}
		//
		stats := util.NewPerfStats()
		// Parse constraints
		hirSchema = readSchema(cfg.stdlib, cfg.debug, legacy, args[1:])
		//
		stats.Log("Reading constraints file")
		// Parse trace file
		columns := readTraceFile(args[0])
		//
		stats.Log("Reading trace file")
		// Go!
		if !checkTraceWithLowering(columns, hirSchema, cfg) {
			os.Exit(1)
		}
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
	// Determines whether or not to apply "defensive padding" to every module.
	defensive bool
	// Determines how much spillage to account for.  This gives the user the
	// ability to override the inferred default.  A negative value indicates
	// this default should be used.
	spillage int
	// Determines how much padding to use
	padding util.Pair[uint, uint]
	// Determines whether or not to enable debugging constraints
	debug bool
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
	// Specifies whether or not to include the standard library.  The default is
	// to include it.
	stdlib bool
	// Specifies whether or not to report details of the failure (e.g. for
	// debugging purposes).
	report bool
	// Specifies the number of additional rows to show eitherside of the failing
	// area. This essentially allows more contextual information to be shown.
	reportPadding uint
	// Specifies the width of a cell to show.
	reportCellWidth uint
	// Perform trace expansion in parallel (or not)
	parallelExpansion bool
	// Size of constraint batches to execute in parallel
	batchSize uint
	// Enable ansi escape codes in reports
	ansiEscapes bool
}

// Check a given trace is consistently accepted (or rejected) at the different
// IR levels.
func checkTraceWithLowering(cols []tr.RawColumn, schema *hir.Schema, cfg checkConfig) bool {
	res := true
	// Process individually
	if cfg.hir {
		res = checkTrace("HIR", cols, schema, cfg)
	}

	if cfg.mir {
		res = checkTrace("MIR", cols, schema.LowerToMir(), cfg) && res
	}

	if cfg.air {
		res = checkTrace("AIR", cols, schema.LowerToMir().LowerToAir(), cfg) && res
	}

	return res
}

func checkTrace(ir string, cols []tr.RawColumn, schema sc.Schema, cfg checkConfig) bool {
	builder := sc.NewTraceBuilder(schema).
		Defensive(cfg.defensive).
		Expand(cfg.expand).
		Parallel(cfg.parallelExpansion).
		BatchSize(cfg.batchSize)
	//
	for n := cfg.padding.Left; n <= cfg.padding.Right; n++ {
		stats := util.NewPerfStats()
		trace, errs := builder.Padding(n).Build(cols)
		// Log cost of expansion
		stats.Log("Expanding trace columns")
		// Report any errors
		reportErrors(cfg.strict, ir, errs)
		// Check whether considered unrecoverable
		if trace == nil || (cfg.strict && len(errs) > 0) {
			return false
		}
		//
		stats = util.NewPerfStats()
		// Check constraints
		if errs := sc.Accepts(cfg.batchSize, schema, trace); len(errs) > 0 {
			reportFailures(ir, errs, trace, cfg)
			return false
		}
		// Check assertions
		if errs := sc.Asserts(cfg.batchSize, schema, trace); len(errs) > 0 {
			reportFailures(ir, errs, trace, cfg)
			return false
		}

		stats.Log("Checking constraints")
	}
	// Done
	return true
}

// Report constraint failures, whilst providing contextual information (when requested).
func reportFailures(ir string, failures []sc.Failure, trace tr.Trace, cfg checkConfig) {
	errs := make([]error, len(failures))
	for i, f := range failures {
		errs[i] = errors.New(f.Message())
	}
	// First, log errors
	reportErrors(true, ir, errs)
	// Second, produce report (if requested)
	if cfg.report {
		for _, f := range failures {
			reportFailure(f, trace, cfg)
		}
	}
}

// Print a human-readable report detailing the given failure
func reportFailure(failure sc.Failure, trace tr.Trace, cfg checkConfig) {
	if f, ok := failure.(*constraint.VanishingFailure); ok {
		cells := f.RequiredCells(trace)
		reportConstraintFailure("constraint", f.Handle, cells, trace, cfg)
	} else if f, ok := failure.(*sc.AssertionFailure); ok {
		cells := f.RequiredCells(trace)
		reportConstraintFailure("assertion", f.Handle, cells, trace, cfg)
	}
}

// Print a human-readable report detailing the given failure with a vanishing constraint.
func reportConstraintFailure(kind string, handle string, cells *util.AnySortedSet[tr.CellRef],
	trace tr.Trace, cfg checkConfig) {
	var start uint = math.MaxUint
	// Determine all (input) cells involved in evaluating the given constraint
	end := uint(0)
	// Determine row bounds
	for _, c := range cells.ToArray() {
		start = min(start, uint(c.Row))
		end = max(end, uint(c.Row))
	}
	// Determine columns to show
	cols := util.NewSortedSet[uint]()
	for _, c := range cells.ToArray() {
		cols.Insert(c.Column)
	}
	// Construct & configure printer
	tp := tr.NewPrinter().Start(start).End(end).MaxCellWidth(cfg.reportCellWidth).Padding(cfg.reportPadding)
	// Determine whether to enable ANSI escapes (e.g. for colour in the terminal)
	tp = tp.AnsiEscapes(cfg.ansiEscapes)
	// Filter out columns not used in evaluating the constraint.
	tp = tp.Columns(func(col uint, trace tr.Trace) bool {
		return cols.Contains(col)
	})
	// Highlight failing cells
	tp = tp.Highlight(func(cell tr.CellRef, trace tr.Trace) bool {
		return cells.Contains(cell)
	})
	// Print out report
	fmt.Printf("failing %s %s:\n", kind, handle)
	tp.Print(trace)
	fmt.Println()
}

func reportErrors(error bool, ir string, errs []error) {
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
	checkCmd.Flags().Uint("report-context", 2, "specify number of rows to show eitherside of failure in report")
	checkCmd.Flags().Uint("report-cellwidth", 32, "specify max number of bytes to show in a given cell in the report")
	checkCmd.Flags().Bool("raw", false, "assume input trace already expanded")
	checkCmd.Flags().Bool("hir", false, "check at HIR level")
	checkCmd.Flags().Bool("mir", false, "check at MIR level")
	checkCmd.Flags().Bool("air", false, "check at AIR level")
	checkCmd.Flags().BoolP("warn", "w", false, "report warnings instead of failing for certain errors"+
		"(e.g. unknown columns in the trace)")
	checkCmd.Flags().Bool("no-stdlib", false, "prevents the standard library from being included")
	checkCmd.Flags().Bool("debug", false, "enable debugging constraints")
	checkCmd.Flags().BoolP("quiet", "q", false, "suppress output (e.g. warnings)")
	checkCmd.Flags().Bool("sequential", false, "perform sequential trace expansion")
	checkCmd.Flags().Bool("defensive", true, "automatically apply defensive padding to every module")
	checkCmd.Flags().Uint("padding", 0, "specify amount of (front) padding to apply")
	checkCmd.Flags().UintP("batch", "b", math.MaxUint, "specify batch size for constraint checking")
	checkCmd.Flags().Int("spillage", -1,
		"specify amount of splillage to account for (where -1 indicates this should be inferred)")
	checkCmd.Flags().Bool("ansi-escapes", true, "specify whether to allow ANSI escapes or not (e.g. for colour reports)")
}
