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
	Use:   "test [flags] constraint_file",
	Short: "Test a set of constraints (e.g. using randomly inputs). [EXPERIMENTAL]",
	Long: `Check a given trace against a set of constraints.
	Constraints can be given either as lisp or bin files.  The goal is to check for
	 properties which don't hold on valid traces`,
	Run: func(cmd *cobra.Command, args []string) {
		var cfg checkConfig
		var hirSchema *hir.Schema

		if len(args) != 1 {
			fmt.Println(cmd.UsageString())
			os.Exit(1)
		}
		// Configure log level
		if GetFlag(cmd, "verbose") {
			log.SetLevel(log.DebugLevel)
		}
		legacy := GetFlag(cmd, "legacy")
		// Setup check config
		cfg.air = GetFlag(cmd, "air")
		cfg.mir = GetFlag(cmd, "mir")
		cfg.hir = GetFlag(cmd, "hir")
		cfg.expand = !GetFlag(cmd, "raw")
		cfg.stdlib = !GetFlag(cmd, "no-stdlib")
		cfg.report = GetFlag(cmd, "report")
		cfg.reportPadding = GetUint(cmd, "report-context")
		// cfg.strict = !GetFlag(cmd, "warn")
		// cfg.quiet = GetFlag(cmd, "quiet")
		cfg.padding.Right = GetUint(cmd, "padding")
		cfg.parallelExpansion = !GetFlag(cmd, "sequential")
		cfg.batchSize = GetUint(cmd, "batch")
		cfg.ansiEscapes = GetFlag(cmd, "ansi-escapes")
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
		hirSchema = readSchema(cfg.stdlib, false, legacy, args)
		//
		stats.Log("Reading constraints file")
		//
		if ok := runTests(2, cfg, hirSchema); !ok {
			// Error signal
			os.Exit(1)
		}
	},
}

func runTests(nrows uint, cfg checkConfig, hirSchema *hir.Schema) bool {
	ok := true
	// TODO: this only tests the happy path.
	for iter := initTraceEnumerator(nrows, hirSchema); iter.HasNext(); {
		// Read out next trace to test
		trace := iter.Next()
		// Test this specific trace
		ok = testTraceWithLowering(trace, hirSchema, cfg) && ok
	}
	//
	return ok
}

// Check a given trace is consistently accepted (or rejected) at the different
// IR levels.
func testTraceWithLowering(trace tr.Trace, schema *hir.Schema, cfg checkConfig) bool {
	ok := true
	// Check whether assertions hold for this trace
	asserts := sc.Asserts(cfg.batchSize, schema, trace)
	// Process individually
	if cfg.hir {
		ok = testTrace("HIR", asserts, trace, schema, cfg) && ok
	}

	if cfg.mir {
		ok = testTrace("MIR", asserts, trace, schema.LowerToMir(), cfg) && ok
	}

	if cfg.air {
		ok = testTrace("AIR", asserts, trace, schema.LowerToMir().LowerToAir(), cfg) && ok
	}

	return ok
}

func testTrace(ir string, asserts []sc.Failure, trace tr.Trace, schema sc.Schema, cfg checkConfig) bool {
	ok := true
	//
	for n := cfg.padding.Left; n <= cfg.padding.Right; n++ {
		// Check constraints
		if errs := sc.Accepts(cfg.batchSize, schema, trace); len(asserts) > 0 && len(errs) == 0 {
			// Trace accepts, but at least one assertion has failed.
			reportFailures(ir, asserts, trace, cfg)
			// Indicate all is not well
			ok = false
		}
	}
	// Done
	return ok
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
	testCmd.Flags().Bool("sequential", false, "perform sequential trace expansion")
	testCmd.Flags().Uint("padding", 0, "specify amount of (front) padding to apply")
	testCmd.Flags().UintP("batch", "b", math.MaxUint, "specify batch size for constraint checking")
	testCmd.Flags().Int("spillage", -1,
		"specify amount of splillage to account for (where -1 indicates this should be inferred)")
	testCmd.Flags().Bool("ansi-escapes", true, "specify whether to allow ANSI escapes or not (e.g. for colour reports)")
}
