// Copyright Consensys Software Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with
// the License. You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
package cmd

import (
	"errors"
	"fmt"
	"math"
	"os"
	"path"

	"github.com/consensys/go-corset/pkg/asm"
	"github.com/consensys/go-corset/pkg/asm/insn"
	"github.com/consensys/go-corset/pkg/corset"
	"github.com/consensys/go-corset/pkg/hir"
	"github.com/consensys/go-corset/pkg/mir"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/constraint"
	"github.com/consensys/go-corset/pkg/trace"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/set"
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
		var cfg checkConfig

		if len(args) != 2 {
			fmt.Println(cmd.UsageString())
			os.Exit(1)
		}
		// Configure log level
		if GetFlag(cmd, "verbose") {
			log.SetLevel(log.DebugLevel)
		}
		optimisation := GetUint(cmd, "opt")
		batched := GetFlag(cmd, "batched")
		//
		if optimisation >= uint(len(mir.OPTIMISATION_LEVELS)) {
			fmt.Printf("invalid optimisation level %d\n", optimisation)
			os.Exit(1)
		}
		//
		cfg.uasm = GetFlag(cmd, "uasm")
		cfg.air = GetFlag(cmd, "air")
		cfg.mir = GetFlag(cmd, "mir")
		cfg.hir = GetFlag(cmd, "hir")
		cfg.defensive = GetFlag(cmd, "defensive")
		cfg.validate = GetFlag(cmd, "validate")
		cfg.expand = !GetFlag(cmd, "raw")
		cfg.report = GetFlag(cmd, "report")
		cfg.reportPadding = GetUint(cmd, "report-context")
		cfg.reportCellWidth = GetUint(cmd, "report-cellwidth")
		cfg.spillage = GetInt(cmd, "spillage")
		cfg.corsetConfig.Stdlib = !GetFlag(cmd, "no-stdlib")
		cfg.corsetConfig.Debug = GetFlag(cmd, "debug")
		cfg.corsetConfig.Legacy = GetFlag(cmd, "legacy")
		cfg.padding.Right = GetUint(cmd, "padding")
		cfg.parallel = !GetFlag(cmd, "sequential")
		cfg.batchSize = GetUint(cmd, "batch")
		cfg.ansiEscapes = GetFlag(cmd, "ansi-escapes")
		cfg.optimisation = mir.OPTIMISATION_LEVELS[optimisation]
		cfg.asmConfig = parseLoweringConfig(cmd)
		externs := GetStringArray(cmd, "set")
		// TODO: support true ranges
		cfg.padding.Left = cfg.padding.Right
		// enable / disable coverage
		if covfile := GetString(cmd, "coverage"); covfile != "" {
			cfg.coverage = util.Some(covfile)
		}
		//
		tracefile := args[0]
		constraints := args[1:]
		// Determine which pipeline to use
		if len(constraints) == 1 && path.Ext(constraints[0]) == ".zkasm" {
			// Single (asm) file supplied
			checkWithAsmPipeline(cfg, args[0], constraints[0])
		} else {
			// Configure Intermediate Representations
			if !cfg.hir && !cfg.mir && !cfg.air {
				// If IR not specified default to running all.
				cfg.hir, cfg.mir, cfg.air = true, true, true
			}
			//
			checkWithLegacyPipeline(cfg, batched, externs, tracefile, constraints)
		}
	},
}

// check config encapsulates certain parameters to be used when
// checking traces.
type checkConfig struct {
	// Perform checking at µASM level
	uasm bool
	// Perform checking at HIR level
	hir bool
	// Perform checking at MIR level
	mir bool
	// Perform checking at AIR level
	air bool
	// Lowering config
	asmConfig asm.LoweringConfig
	// Set optimisation config to use.
	optimisation mir.OptimisationConfig
	// Determines whether or not to apply "defensive padding" to every module.
	defensive bool
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
	// Specifies whether or not to perform trace validation.  That is, to check
	// all input values are within expected bounds.
	validate bool
	// Corset compilation options
	corsetConfig corset.CompilationConfig
	// Specifies whether to use coverage testing and, if so, where to write the
	// coverage data.
	coverage util.Option[string]
	// Specifies whether or not to report details of the failure (e.g. for
	// debugging purposes).
	report bool
	// Specifies the number of additional rows to show eitherside of the failing
	// area. This essentially allows more contextual information to be shown.
	reportPadding uint
	// Specifies the width of a cell to show.
	reportCellWidth uint
	// Perform trace expansion in parallel (or not)
	parallel bool
	// Size of constraint batches to execute in parallel
	batchSize uint
	// Enable ansi escape codes in reports
	ansiEscapes bool
}

func checkWithAsmPipeline(cfg checkConfig, tracefile string, asmfiles ...string) {
	var (
		ok              bool = true
		macroProgram, _      = ReadAssemblyProgram(asmfiles...)
		microProgram         = macroProgram.Lower(cfg.asmConfig)
		trace                = ReadAssemblyTrace(tracefile, &macroProgram)
	)
	//
	for _, instance := range trace {
		// Macro check
		ok = checkFunctionInstance("ASM", instance, &macroProgram) && ok
		// Micro check
		if cfg.uasm {
			ok = checkFunctionInstance("µASM", instance, &microProgram) && ok
		}
	}
	//
	if cfg.hir || cfg.mir || cfg.air {
		binfile, errs := asm.NewCompiler().Compile(microProgram)
		// Check constraints
		if len(errs) > 0 {
			for _, err := range errs {
				printSyntaxError(&err)
			}
		} else {
			builder := asm.NewTraceBuilder(&microProgram)
			hirTrace := builder.Build(trace)
			ok, _ = checkTraceWithLowering([][]tr.RawColumn{hirTrace}, &binfile.Schema, cfg)
		}
	}
	//
	if !ok {
		os.Exit(4)
	}
}

func checkFunctionInstance[T insn.Instruction](ir string, instance asm.FunctionInstance, program asm.Program[T]) bool {
	// Macro check
	if outcome, err := asm.CheckInstance(instance, program); outcome == math.MaxUint {
		// Internal failure
		panic(err)
	} else if outcome != 0 {
		fmt.Printf("trace rejected (%s): %s\n", ir, err)
		return false
	}
	// success
	return true
}

// Check raw constraints using the legacy pipeline.
func checkWithLegacyPipeline(cfg checkConfig, batched bool, externs []string, tracefile string, constraints []string) {
	var traces [][]trace.RawColumn
	//
	stats := util.NewPerfStats()
	// Parse constraints
	binfile := ReadConstraintFiles(cfg.corsetConfig, cfg.asmConfig, constraints)
	//
	stats.Log("Reading constraints file")
	// Parse trace file(s)
	if batched {
		// batched mode
		traces = ReadBatchedTraceFile(tracefile)
	} else {
		// unbatched (i.e. normal) mode
		tracefile := ReadTraceFile(tracefile)
		traces = [][]trace.RawColumn{tracefile.Columns}
	}
	//
	stats.Log("Reading trace file")
	// Apply any user-specified values for externalised constants.
	applyExternOverrides(externs, binfile)
	// Go!
	ok, coverage := checkTraceWithLowering(traces, &binfile.Schema, cfg)
	//
	if !ok {
		os.Exit(1)
	} else if cfg.coverage.HasValue() {
		writeCoverageReport(cfg.coverage.Unwrap(), binfile, coverage, cfg.optimisation)
	}
}

// Check a given trace is consistently accepted (or rejected) at the different
// IR levels.
func checkTraceWithLowering(traces [][]tr.RawColumn, schema *hir.Schema, cfg checkConfig) (bool, [3]sc.CoverageMap) {
	var (
		tmp         bool
		airCoverage sc.CoverageMap
		mirCoverage sc.CoverageMap
		hirCoverage sc.CoverageMap
	)
	//
	res := true
	// Process individually
	if cfg.hir {
		res, hirCoverage = checkTrace("HIR", traces, schema, cfg)
	}

	if cfg.mir {
		tmp, mirCoverage = checkTrace("MIR", traces, schema.LowerToMir(), cfg)
		//
		res = res && tmp
	}

	if cfg.air {
		airSchema := schema.LowerToMir().LowerToAir(cfg.optimisation)
		tmp, airCoverage = checkTrace("AIR", traces, airSchema, cfg)
		//
		res = res && tmp
	}

	return res, [3]sc.CoverageMap{airCoverage, mirCoverage, hirCoverage}
}

func checkTrace(ir string, traces [][]tr.RawColumn, schema sc.Schema,
	cfg checkConfig) (bool, sc.CoverageMap) {
	//
	coverage := sc.NewBranchCoverage()
	builder := sc.NewTraceBuilder(schema).
		Validate(cfg.validate).
		Defensive(cfg.defensive).
		Expand(cfg.expand).
		Parallel(cfg.parallel).
		BatchSize(cfg.batchSize)
	//
	for _, cols := range traces {
		for n := cfg.padding.Left; n <= cfg.padding.Right; n++ {
			stats := util.NewPerfStats()
			trace, errs := builder.Padding(n).Build(cols)
			// Log cost of expansion
			stats.Log("Expanding trace columns")
			// Report any errors
			reportErrors(ir, errs)
			// Check whether considered unrecoverable
			if trace == nil || len(errs) > 0 {
				return false, coverage
			}
			//
			stats = util.NewPerfStats()
			// Check constraints
			if cov, errs := sc.Accepts(cfg.parallel, cfg.batchSize, schema, trace); len(errs) > 0 {
				reportFailures(ir, errs, trace, cfg)
				return false, coverage
			} else {
				coverage.Union(cov)
			}
			// Check assertions
			if _, errs := sc.Asserts(cfg.parallel, cfg.batchSize, schema, trace); len(errs) > 0 {
				reportFailures(ir, errs, trace, cfg)
				return false, coverage
			}

			stats.Log("Checking constraints")
		}
	}
	// Done
	return true, coverage
}

// Report constraint failures, whilst providing contextual information (when requested).
func reportFailures(ir string, failures []sc.Failure, trace tr.Trace, cfg checkConfig) {
	errs := make([]error, len(failures))
	for i, f := range failures {
		errs[i] = errors.New(f.Message())
	}
	// First, log errors
	reportErrors(ir, errs)
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
		fmt.Printf("failing constraint %s:\n", f.Handle)
		reportRelevantCells(cells, trace, cfg)
	} else if f, ok := failure.(*constraint.RangeFailure); ok {
		cells := f.RequiredCells(trace)
		fmt.Printf("failing range constraint %s:\n", f.Handle)
		reportRelevantCells(cells, trace, cfg)
	} else if f, ok := failure.(*constraint.LookupFailure); ok {
		cells := f.RequiredCells(trace)
		fmt.Printf("failing lookup constraint %s:\n", f.Handle)
		reportRelevantCells(cells, trace, cfg)
	} else if f, ok := failure.(*sc.AssertionFailure); ok {
		cells := f.RequiredCells(trace)
		fmt.Printf("failing assertion %s:\n", f.Handle)
		reportRelevantCells(cells, trace, cfg)
	} else if f, ok := failure.(*sc.InternalFailure); ok {
		cells := f.RequiredCells(trace)
		fmt.Printf("%s in %s:\n", f.Error, f.Handle)
		reportRelevantCells(cells, trace, cfg)
	}
}

// Print a human-readable report detailing the given failure with a vanishing constraint.
func reportRelevantCells(cells *set.AnySortedSet[tr.CellRef], trace tr.Trace, cfg checkConfig) {
	var start uint = math.MaxUint
	// Determine all (input) cells involved in evaluating the given constraint
	end := uint(0)
	// Determine row bounds
	for _, c := range cells.ToArray() {
		start = min(start, uint(c.Row))
		end = max(end, uint(c.Row))
	}
	// Determine columns to show
	cols := set.NewSortedSet[uint]()
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
	tp.Print(trace)
	fmt.Println()
}

func reportErrors(ir string, errs []error) {
	// Construct set to ensure deduplicate errors
	set := make(map[string]bool, len(errs))
	//
	for _, err := range errs {
		key := fmt.Sprintf("%s (%s)", err, ir)
		set[key] = true
	}
	// Report each one
	for e := range set {
		log.Errorln(e)
	}
}

func init() {
	rootCmd.AddCommand(checkCmd)
	checkCmd.Flags().Bool("report", false, "report details of failure for debugging")
	checkCmd.Flags().Uint("report-context", 2, "specify number of rows to show eitherside of failure in report")
	checkCmd.Flags().Uint("report-cellwidth", 32, "specify max number of bytes to show in a given cell in the report")
	checkCmd.Flags().Bool("raw", false, "assume input trace already expanded")
	checkCmd.Flags().Bool("uasm", false, "check at µASM level")
	checkCmd.Flags().Bool("hir", false, "check at HIR level")
	checkCmd.Flags().Bool("mir", false, "check at MIR level")
	checkCmd.Flags().Bool("air", false, "check at AIR level")
	checkCmd.Flags().Bool("no-stdlib", false, "prevents the standard library from being included")
	checkCmd.Flags().Bool("debug", false, "enable debugging constraints")
	checkCmd.Flags().Bool("sequential", false, "perform sequential trace expansion")
	checkCmd.Flags().Bool("defensive", true, "automatically apply defensive padding to every module")
	checkCmd.Flags().Bool("validate", true, "apply trace validation")
	checkCmd.Flags().String("coverage", "", "write JSON coverage data to file")
	checkCmd.Flags().Uint("padding", 0, "specify amount of (front) padding to apply")
	checkCmd.Flags().UintP("batch", "b", math.MaxUint, "specify batch size for constraint checking")
	checkCmd.Flags().Bool("batched", false,
		"specify trace file is batched (i.e. contains multiple traces, one for each line)")
	checkCmd.Flags().Int("spillage", -1,
		"specify amount of splillage to account for (where -1 indicates this should be inferred)")
	checkCmd.Flags().Bool("ansi-escapes", true, "specify whether to allow ANSI escapes or not (e.g. for colour reports)")
	checkCmd.Flags().StringArrayP("set", "S", []string{}, "set value of externalised constant.")
}
