package cmd

import (
	"fmt"
	"math"
	"os"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/hir"
	sc "github.com/consensys/go-corset/pkg/schema"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test [flags] trace_file constraint_file",
	Short: "Test a set of constraints (e.g. using randomly inputs).",
	Long: `Check a given trace against a set of constraints.
	Traces can be given either as JSON or binary lt files.
	Constraints can be given either as lisp or bin files.`,
	Run: func(cmd *cobra.Command, args []string) {
		var cfg checkConfig
		var hirSchema *hir.Schema

		if len(args) != 1 {
			fmt.Println(cmd.UsageString())
			os.Exit(1)
		}
		// Configure log level
		if getFlag(cmd, "debug") {
			log.SetLevel(log.DebugLevel)
		}
		// Setup check config
		cfg.air = getFlag(cmd, "air")
		cfg.mir = getFlag(cmd, "mir")
		cfg.hir = getFlag(cmd, "hir")
		cfg.expand = !getFlag(cmd, "raw")
		cfg.report = getFlag(cmd, "report")
		cfg.reportPadding = getUint(cmd, "report-context")
		// cfg.strict = !getFlag(cmd, "warn")
		// cfg.quiet = getFlag(cmd, "quiet")
		cfg.padding.Right = getUint(cmd, "padding")
		cfg.parallelExpansion = !getFlag(cmd, "sequential")
		cfg.batchSize = getUint(cmd, "batch")
		cfg.ansiEscapes = getFlag(cmd, "ansi-escapes")
		// TODO: support true ranges
		cfg.padding.Left = cfg.padding.Right
		// Normalise IRs
		if !cfg.hir && !cfg.mir && !cfg.air {
			// If IR not specified default to running all.
			cfg.hir, cfg.mir, cfg.air = true, true, true
		}
		//
		stats := util.NewPerfStats()
		// Parse constraints
		hirSchema = readSchemaFile(args[0])
		//
		stats.Log("Reading constraints file")
		//
		if errs := runTests(2, cfg, hirSchema); len(errs) > 0 {
			// Report errors
			for _, e := range errs {
				log.Error(e)
			}
			// Error signal
			os.Exit(1)
		}
	},
}

func runTests(nrows uint, cfg checkConfig, hirSchema *hir.Schema) []error {
	errors := []error{}
	// TODO: this only tests the happy path.
	for iter := initTraceEnumerator(nrows, hirSchema); iter.HasNext(); {
		// Read out next trace to test
		trace := iter.Next()
		// Detemine whether trace should be considered valid
		// Test it
		if errs := testTraceWithLowering(trace, hirSchema, cfg); len(errs) > 0 {
			errors = append(errors, errs...)
		}
	}
	//
	return errors
}

// Check a given trace is consistently accepted (or rejected) at the different
// IR levels.
func testTraceWithLowering(trace tr.Trace, schema *hir.Schema, cfg checkConfig) []error {
	errs := []error{}
	// Check whether assertions hold for this trace
	valid := sc.Asserts(cfg.batchSize, schema, trace) == nil
	// Process individually
	if cfg.hir {
		errs = testTrace("HIR", valid, trace, schema, cfg)
	}

	if cfg.mir {
		errs = append(errs, testTrace("MIR", valid, trace, schema.LowerToMir(), cfg)...)
	}

	if cfg.air {
		errs = append(errs, testTrace("AIR", valid, trace, schema.LowerToMir().LowerToAir(), cfg)...)
	}

	return errs
}

func testTrace(ir string, valid bool, trace tr.Trace, schema sc.Schema, cfg checkConfig) []error {
	errors := []error{}
	//
	for n := cfg.padding.Left; n <= cfg.padding.Right; n++ {
		// Check constraints
		if errs := sc.Accepts(cfg.batchSize, schema, trace); valid && len(errs) > 0 {
			// Trace rejected, but should have been accepted
			err := fmt.Errorf("rejected incorrectly: %s (%s)", trace, ir)
			errors = append(errors, err)
		} else if !valid && len(errs) == 0 {
			// Trace accepted, but should have been rejected
			err := fmt.Errorf("accepted incorrectly: %s (%s)", trace, ir)
			errors = append(errors, err)
		}
	}
	// Done
	return errors
}

// Constructs a (lazy) enumerator over the set of traces to be used for testing.
func initTraceEnumerator(nrows uint, hirSchema *hir.Schema) util.Enumerator[tr.Trace] {
	// NOTE: This is really a temporary solution for now.  It doesn't handle
	// length multipliers.  It doesn't allow for modules with different heights.
	// It uses a fixed pool.
	pool := []fr.Element{fr.NewElement(0), fr.NewElement(1), fr.NewElement(2),
		fr.NewElement(3), fr.NewElement(4), fr.NewElement(5)}
	// Done
	return sc.NewTraceEnumerator(nrows, hirSchema, pool)
}

func init() {
	rootCmd.AddCommand(testCmd)
	testCmd.Flags().Bool("report", false, "report details of failure for debugging")
	testCmd.Flags().Uint("report-context", 2, "specify number of rows to show eitherside of failure in report")
	testCmd.Flags().Bool("raw", false, "assume input trace already expanded")
	testCmd.Flags().Bool("hir", false, "check at HIR level")
	testCmd.Flags().Bool("mir", false, "check at MIR level")
	testCmd.Flags().Bool("air", false, "check at AIR level")
	// testCmd.Flags().BoolP("warn", "w", false, "report warnings instead of failing for certain errors"+
	// 	"(e.g. unknown columns in the trace)")
	testCmd.Flags().BoolP("debug", "d", false, "report debug logs")
	//testCmd.Flags().BoolP("quiet", "q", false, "suppress output (e.g. warnings)")
	testCmd.Flags().Bool("sequential", false, "perform sequential trace expansion")
	testCmd.Flags().Uint("padding", 0, "specify amount of (front) padding to apply")
	testCmd.Flags().UintP("batch", "b", math.MaxUint, "specify batch size for constraint checking")
	testCmd.Flags().Int("spillage", -1,
		"specify amount of splillage to account for (where -1 indicates this should be inferred)")
	testCmd.Flags().Bool("ansi-escapes", true, "specify whether to allow ANSI escapes or not (e.g. for colour reports)")
}
