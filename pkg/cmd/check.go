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
	"os"
	"runtime"
	"runtime/pprof"

	"github.com/consensys/go-corset/pkg/asm"
	"github.com/consensys/go-corset/pkg/binfile"
	cmd_util "github.com/consensys/go-corset/pkg/cmd/util"
	"github.com/consensys/go-corset/pkg/cmd/view"
	"github.com/consensys/go-corset/pkg/corset"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/constraint"
	"github.com/consensys/go-corset/pkg/schema/constraint/lookup"
	"github.com/consensys/go-corset/pkg/schema/constraint/ranged"
	"github.com/consensys/go-corset/pkg/schema/constraint/vanishing"
	"github.com/consensys/go-corset/pkg/schema/module"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/trace/lt"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/set"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/field/bls12_377"
	"github.com/consensys/go-corset/pkg/util/field/gf251"
	"github.com/consensys/go-corset/pkg/util/field/gf8209"
	"github.com/consensys/go-corset/pkg/util/field/koalabear"
	"github.com/consensys/go-corset/pkg/util/termio/widget"
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
		runFieldAgnosticCmd(cmd, args, checkCmds)
	},
}

// Available instances
var checkCmds = []FieldAgnosticCmd{
	{field.GF_251, runCheckCmd[gf251.Element]},
	{field.GF_8209, runCheckCmd[gf8209.Element]},
	{field.KOALABEAR_16, runCheckCmd[koalabear.Element]},
	{field.BLS12_377, runCheckCmd[bls12_377.Element]},
}

func runCheckCmd[F field.Element[F]](cmd *cobra.Command, args []string) {
	var cfg checkConfig

	if len(args) < 2 {
		fmt.Println(cmd.UsageString())
		os.Exit(1)
	}
	// Configure log level
	if GetFlag(cmd, "verbose") {
		log.SetLevel(log.DebugLevel)
	}
	// Configure CPU profiling (if requested)
	startCpuProfiling(cmd)
	//
	batched := GetFlag(cmd, "batched")
	//
	cfg.padding.Right = GetUint(cmd, "padding")
	cfg.report = GetFlag(cmd, "report")
	cfg.reportPadding = GetUint(cmd, "report-context")
	cfg.reportLimbs = GetFlag(cmd, "report-limbs")
	cfg.reportCellWidth = GetUint(cmd, "report-cellwidth")
	cfg.reportTitleWidth = GetUint(cmd, "report-titlewidth")
	cfg.ansiEscapes = GetFlag(cmd, "ansi-escapes")
	// TODO: support true ranges
	cfg.padding.Left = cfg.padding.Right
	// Read in constraint files
	schemas := *getSchemaStack[F](cmd, SCHEMA_DEFAULT_AIR, args[1:]...)
	// enable / disable coverage
	if covfile := GetString(cmd, "coverage"); covfile != "" {
		cfg.coverage = util.Some(covfile)
	}
	//
	tracefile := args[0]
	//
	checkWithLegacyPipeline(cfg, batched, tracefile, schemas)
	// Write memory profiling (if requested)
	writeMemProfile(cmd)
	// Stop cpu profiling (if was requested)
	stopCpuProfiling(cmd)
}

func startCpuProfiling(cmd *cobra.Command) {
	if filename := GetString(cmd, "cpuprof"); filename != "" {
		f, err := os.Create(filename)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		//
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
	}
}

func stopCpuProfiling(cmd *cobra.Command) {
	if filename := GetString(cmd, "cpuprof"); filename != "" {
		pprof.StopCPUProfile()
	}
}

func writeMemProfile(cmd *cobra.Command) {
	if filename := GetString(cmd, "memprof"); filename != "" {
		f, err := os.Create(filename)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		//nolint
		defer f.Close()
		//
		runtime.GC()
		//
		if err := pprof.Lookup("allocs").WriteTo(f, 0); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
	}
}

// check config encapsulates certain parameters to be used when
// checking traces.
type checkConfig struct {
	// Corset source mapping (maybe nil if non available).
	corsetSourceMap *corset.SourceMap
	// Specifies whether to use coverage testing and, if so, where to write the
	// coverage data.
	coverage util.Option[string]
	// Specifies the range of padding values to check
	padding util.Pair[uint, uint]
	// Specifies whether or not to report details of the failure (e.g. for
	// debugging purposes).
	report bool
	// Specifies the number of additional rows to show eitherside of the failing
	// area. This essentially allows more contextual information to be shown.
	reportPadding uint
	// Specifies the width of a cell to show.
	reportCellWidth uint
	// Specifies the width of a column title to show.
	reportTitleWidth uint
	// Specifies whether or not to show raw limbs
	reportLimbs bool
	// Enable ansi escape codes in reports
	ansiEscapes bool
}

// Check raw constraints using the legacy pipeline.
func checkWithLegacyPipeline[F field.Element[F]](cfg checkConfig, batched bool, tracefile string,
	schemas cmd_util.SchemaStacker[F]) {
	//
	var (
		errors    []error
		traces    []lt.TraceFile
		ok        bool = true
		expanding      = schemas.TraceBuilder().Expanding()
	)
	//
	stats := util.NewPerfStats()
	// Extract debug information (if available)
	cfg.corsetSourceMap, _ = binfile.GetAttribute[*corset.SourceMap](schemas.BinaryFile())
	//
	stats.Log("Reading constraints file")
	// Parse trace file(s)
	if batched {
		// batched mode
		traces = ReadBatchedTraceFile(tracefile)
	} else {
		// unbatched (i.e. normal) mode
		traces = []lt.TraceFile{ReadTraceFile(tracefile)}
	}
	//
	schema := schemas.BinaryFile().Schema
	// Apply trace propagation
	if expanding {
		perf := util.NewPerfStats()
		//
		traces, errors = asm.PropagateAll(schema, traces, expanding)
		//
		perf.Log("Trace propagation")
	}
	// Go!
	if len(errors) == 0 {
		ok = checkTraces(traces, schemas, cfg) && ok
	}
	// Handle errors
	if !ok || len(errors) > 0 {
		for _, err := range errors {
			log.Errorf("%s\n", err.Error())
		}
		//
		os.Exit(1)
	}
}

func checkTraces[F field.Element[F]](traces []lt.TraceFile, stacker cmd_util.SchemaStacker[F], cfg checkConfig) bool {
	//
	for _, tf := range traces {
		//
		for n := cfg.padding.Left; n <= cfg.padding.Right; n++ {
			// Configure stack.  This is important to ensure true separation
			// between runs (e.g. for the io.Executor).
			stack := stacker.Build()
			// configure trace builder
			builder := stack.TraceBuilder().WithPadding(n)
			// Run each concrete schema separately
			for i, schema := range stack.ConcreteSchemas() {
				ir := stack.ConcreteIrName(uint(i))
				stats := util.NewPerfStats()
				trace, errs := builder.Build(schema, tf)
				// Log cost of expansion
				stats.Log("Expanding trace columns")
				// Report any errors
				reportErrors(ir, errs)
				// Check whether considered unrecoverable
				if trace == nil || len(errs) > 0 {
					return false
				}
				//
				stats = util.NewPerfStats()
				// Check constraints
				if errs := sc.Accepts(builder.Parallelism(), builder.BatchSize(), schema, trace); len(errs) > 0 {
					reportFailures(ir, errs, trace, builder.Mapping(), cfg)
					return false
				}

				stats.Log("Checking constraints")
			}
		}
	}
	// Done
	return true
}

// Report constraint failures, whilst providing contextual information (when requested).
func reportFailures[F field.Element[F]](ir string, failures []sc.Failure, trace tr.Trace[F], mapping module.LimbsMap,
	cfg checkConfig) {
	//
	var (
		errs = make([]error, len(failures))
	)
	//
	for i, f := range failures {
		errs[i] = errors.New(f.Message())
	}
	// First, log errors
	reportErrors(ir, errs)
	// Second, produce report (if requested)
	if cfg.report {
		for _, f := range failures {
			reportFailure(f, trace, mapping, cfg)
		}
	}
}

// Print a human-readable report detailing the given failure
func reportFailure[F field.Element[F]](failure sc.Failure, trace tr.Trace[F], mapping module.LimbsMap,
	cfg checkConfig) {
	//
	if f, ok := failure.(*vanishing.Failure[F]); ok {
		cells := f.RequiredCells(trace)
		fmt.Printf("failing constraint %s:\n", f.Handle)
		reportRelevantCells(cells, trace, mapping, cfg)
	} else if f, ok := failure.(*ranged.Failure[F]); ok {
		cells := f.RequiredCells(trace)
		fmt.Printf("failing range constraint %s:\n", f.Handle)
		reportRelevantCells(cells, trace, mapping, cfg)
	} else if f, ok := failure.(*lookup.Failure[F]); ok {
		cells := f.RequiredCells(trace)
		fmt.Printf("failing lookup constraint %s:\n", f.Handle)
		reportRelevantCells(cells, trace, mapping, cfg)
	} else if f, ok := failure.(*constraint.AssertionFailure[F]); ok {
		cells := f.RequiredCells(trace)
		fmt.Printf("failing assertion %s:\n", f.Handle)
		reportRelevantCells(cells, trace, mapping, cfg)
	} else if f, ok := failure.(*constraint.InternalFailure[F]); ok {
		cells := f.RequiredCells(trace)
		fmt.Printf("%s in %s:\n", f.Error, f.Handle)
		reportRelevantCells(cells, trace, mapping, cfg)
	}
}

// Print a human-readable report detailing the given failure with a vanishing constraint.
func reportRelevantCells[F field.Element[F]](cells *set.AnySortedSet[tr.CellRef], trace tr.Trace[F],
	mapping module.LimbsMap, cfg checkConfig) {
	// Construct trace window
	window := view.NewBuilder[F](mapping).
		WithLimbs(cfg.reportLimbs).
		WithCellWidth(cfg.reportCellWidth).
		WithTitleWidth(cfg.reportTitleWidth).
		WithSourceMap(*cfg.corsetSourceMap).
		WithFormatting(view.NewCellFormatter(*cells, cfg.ansiEscapes)).
		Build(trace)
	// Focus window on those cells relevant to the failure
	window = window.Filter(view.FilterForCells(*cells, cfg.reportPadding))
	// Print all windows
	for i := range window.Width() {
		var (
			ith = window.Module(i)
			// Construct & configure printer
			tp = widget.NewTable(window.Module(i))
			//
			name = ith.Data().Name().String()
		)
		// Print out module name
		if window.Width() > 1 && name != "" {
			fmt.Printf("%s:\n", name)
		}
		// Print out report
		tp.Print()
		fmt.Println()
	}
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
	checkCmd.Flags().Bool("report-limbs", false, "specify whether to show register limbs in report")
	checkCmd.Flags().Uint("report-cellwidth", 32, "specify max number of bytes to show in a given cell in report")
	checkCmd.Flags().Uint("report-titlewidth", 40, "specify maximum width of column titles in report")
	//
	checkCmd.Flags().String("coverage", "", "write JSON coverage data to file")
	checkCmd.Flags().Uint("padding", 0, "specify amount of (front) padding to apply")
	checkCmd.Flags().Bool("batched", false,
		"specify trace file is batched (i.e. contains multiple traces, one for each line)")
	checkCmd.Flags().Bool("ansi-escapes", true, "specify whether to allow ANSI escapes or not (e.g. for colour reports)")
	// profiling commands'
	checkCmd.Flags().String("cpuprof", "", "write cpu profile to `file`")
	checkCmd.Flags().String("memprof", "", "write memory profile to `file`")
}
