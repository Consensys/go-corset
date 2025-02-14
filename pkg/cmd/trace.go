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
	"fmt"
	"math"
	"os"
	"regexp"
	"strings"

	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/hash"
	"github.com/spf13/cobra"
)

// traceCmd represents the trace command for manipulating traces.
var traceCmd = &cobra.Command{
	Use:   "trace [flags] trace_file [constraint_file(s)]",
	Short: "Operate on a trace file.",
	Long: `Operate on a trace file, such as converting
	it from one format (e.g. lt) to another (e.g. json),
	or filtering out modules, or listing columns, etc.`,
	Run: func(cmd *cobra.Command, args []string) {
		var traces [][]trace.RawColumn
		expand := GetFlag(cmd, "expand")
		// Sanity check
		if (expand && len(args) != 2) || (!expand && len(args) != 1) {
			fmt.Println(cmd.UsageString())
			os.Exit(1)
		}
		// Parse trace
		list := GetFlag(cmd, "list")
		defensive := GetFlag(cmd, "defensive")
		stats := GetFlag(cmd, "stats")
		stdlib := !GetFlag(cmd, "no-stdlib")
		includes := GetStringArray(cmd, "include")
		print := GetFlag(cmd, "print")
		start := GetUint(cmd, "start")
		end := GetUint(cmd, "end")
		max_width := GetUint(cmd, "max-width")
		filter := GetString(cmd, "filter")
		output := GetString(cmd, "out")
		air := GetFlag(cmd, "air")
		mir := GetFlag(cmd, "mir")
		hir := GetFlag(cmd, "hir")
		batched := GetFlag(cmd, "batched")
		// Parse trace file(s)
		if batched {
			// batched mode
			traces = ReadBatchedTraceFile(args[0])
		} else {
			// unbatched (i.e. normal) mode
			columns := ReadTraceFile(args[0])
			traces = [][]trace.RawColumn{columns}
		}
		//
		if expand && !air && !mir && !hir {
			fmt.Println("must specify --hir/mir/air for trace expansion")
			os.Exit(2)
		} else if expand {
			level := determineAbstractionLevel(air, mir, hir)
			for i, cols := range traces {
				traces[i] = expandWithConstraints(level, cols, stdlib, defensive, args[1:])
			}
		} else if defensive {
			fmt.Println("cannot apply defensive padding without trace expansion")
			os.Exit(2)
		}
		// Now manipulate traces
		for i := range traces {
			// construct filters
			if filter != "" {
				traces[i] = filterColumns(traces[i], filter)
			}
			if start != 0 || end != math.MaxUint {
				sliceColumns(traces[i], start, end)
			}
			if list {
				listColumns(traces[i], includes)
			}
			if stats {
				summaryStats(traces[i])
			}

			if print {
				printTrace(start, max_width, traces[i])
			}
		}
		// Write out results (if requested)
		if output != "" {
			writeBatchedTracesFile(output, traces...)
		}
	},
}

func init() {
	rootCmd.AddCommand(traceCmd)
	traceCmd.Flags().BoolP("list", "l", false, "list only the columns in the trace file")
	traceCmd.Flags().StringArrayP("include", "i", []string{"lines", "bitwidth", "bytes", "elements"},
		fmt.Sprintf("specify information to include in column listing: %s", summariserOptions()))
	traceCmd.Flags().Bool("stats", false, "print summary information about the trace file")
	traceCmd.Flags().BoolP("print", "p", false, "print entire trace file")
	traceCmd.Flags().BoolP("expand", "e", false, "perform trace expansion (schema required)")
	traceCmd.Flags().Bool("defensive", false, "perform defensive padding (schema required)")
	traceCmd.Flags().Uint("start", 0, "filter out rows below this")
	traceCmd.Flags().Uint("end", math.MaxUint, "filter out this and all following rows")
	traceCmd.Flags().Uint("max-width", 32, "specify maximum display width for a column")
	traceCmd.Flags().StringP("out", "o", "", "Specify output file to write trace")
	traceCmd.Flags().StringP("filter", "f", "", "Filter columns matching regex")
	traceCmd.Flags().Bool("hir", false, "expand to HIR level")
	traceCmd.Flags().Bool("mir", false, "expand to MIR level")
	traceCmd.Flags().Bool("air", false, "expand to AIR level")
	traceCmd.Flags().Bool("batched", false,
		"specify trace file is batched (i.e. contains multiple traces, one for each line)")
}

const air_LEVEL = 0
const mir_LEVEL = 1
const hir_LEVEL = 2

func determineAbstractionLevel(air, mir, hir bool) int {
	switch {
	case air && !mir && !hir:
		return air_LEVEL
	case !air && mir && !hir:
		return mir_LEVEL
	case !air && !mir && hir:
		return hir_LEVEL
	case !air && !mir && !hir:
		fmt.Println("must specify target level (hir/mir/air) for trace expansion")
	default:
		fmt.Println("conflicting target level (hir/mir/air) for trace expansion")
	}
	//nolint:revive
	os.Exit(2)
	panic("unreachable")
}

func expandWithConstraints(level int, cols []trace.RawColumn, stdlib bool, defensive bool,
	filenames []string) []trace.RawColumn {
	//
	var schema sc.Schema
	//
	binfile := ReadConstraintFiles(stdlib, false, false, filenames)
	//
	switch level {
	case hir_LEVEL:
		schema = &binfile.Schema
	case mir_LEVEL:
		schema = binfile.Schema.LowerToMir()
	case air_LEVEL:
		schema = binfile.Schema.LowerToMir().LowerToAir()
	default:
		panic("unreachable")
	}
	// Done
	return expandColumns(cols, schema, defensive)
}

func expandColumns(cols []trace.RawColumn, schema sc.Schema, defensive bool) []trace.RawColumn {
	builder := sc.NewTraceBuilder(schema).Expand(true).Defensive(defensive)
	tr, errs := builder.Build(cols)
	//
	if len(errs) > 0 {
		for _, err := range errs {
			fmt.Println(err)
		}
		//
		os.Exit(1)
	}
	// Convert back to raw column array
	rcols := make([]trace.RawColumn, tr.Width())
	//
	for i := range rcols {
		ith := tr.Column(uint(i))
		module := tr.Modules().Nth(ith.Context().Module())
		//
		rcols[i] = trace.RawColumn{
			Module: module.Name(),
			Name:   ith.Name(),
			Data:   ith.Data(),
		}
	}
	//
	return rcols
}

// Construct a new trace containing only those columns from the original who
// name begins with the given prefix.
func filterColumns(cols []trace.RawColumn, regex string) []trace.RawColumn {
	r, err := regexp.Compile(regex)
	// Check for error
	if err != nil {
		panic(err)
	}
	//
	ncols := make([]trace.RawColumn, 0)
	// Now create the columns.
	for i := 0; i < len(cols); i++ {
		name := trace.QualifiedColumnName(cols[i].Module, cols[i].Name)
		if r.MatchString(name) {
			ncols = append(ncols, cols[i])
		}
	}
	// Done
	return ncols
}

// Construct a new trace where all columns are sliced to a given region.  In
// some cases, that might mean the column becomes entirely empty.
func sliceColumns(cols []trace.RawColumn, start uint, end uint) {
	// Now slice them columns.
	for i := 0; i < len(cols); i++ {
		ith := cols[i]
		s := min(start, ith.Data.Len())
		e := min(end, ith.Data.Len())
		cols[i] = trace.RawColumn{
			Module: ith.Module,
			Name:   ith.Name,
			Data:   ith.Data.Slice(s, e),
		}
	}
}

func printTrace(start uint, max_width uint, cols []trace.RawColumn) {
	n := uint(len(cols))
	height := maxHeightColumns(cols)
	tbl := util.NewTablePrinter(1+height, 1+n)

	for j := uint(0); j < height; j++ {
		tbl.Set(j+1, 0, fmt.Sprintf("#%d", j+start))
	}

	for i := uint(0); i < n; i++ {
		ith := cols[i].Data
		tbl.Set(0, i+1, cols[i].QualifiedName())

		for j := uint(0); j < ith.Len(); j++ {
			jth := ith.Get(j)

			tbl.Set(j+1, i+1, jth.Text(16))
		}
	}
	//
	tbl.SetMaxWidths(max_width)
	tbl.Print()
}

func listColumns(tr []trace.RawColumn, includes []string) {
	summarisers := selectColumnSummarisers(includes)
	m := 1 + uint(len(summarisers))
	n := uint(len(tr))
	// Go!
	tbl := util.NewTablePrinter(m, n+1)
	c := make(chan util.Pair[uint, []string], n)
	// Set titles
	tbl.Set(0, 0, "Column")

	for i := uint(0); i < uint(len(summarisers)); i++ {
		tbl.Set(i+1, 0, summarisers[i].name)
	}
	// Compute data
	for i := uint(0); i < n; i++ {
		// Launch summarisers
		go func(index uint) {
			// Apply summarisers to column
			row := summariseColumn(tr[index], summarisers)
			// Package result
			c <- util.NewPair(index, row)
		}(i)
	}
	// Collect results
	for i := uint(0); i < n; i++ {
		// Read packaged result from channel
		res := <-c
		// Set row
		tbl.SetRow(res.Left+1, res.Right...)
	}
	//
	tbl.SetMaxWidths(64)
	tbl.Print()
}

func selectColumnSummarisers(includes []string) []ColSummariser {
	includes = flattenIncludes(includes)
	summarisers := make([]ColSummariser, len(includes))
	// Iterate included summarisers
	for i, ss := range includes {
		// Look them up
		for _, cs := range colSummarisers {
			if cs.name == ss {
				summarisers[i] = cs
				break
			}
		}
		// Sanity check we found something
		if summarisers[i].name != ss {
			panic(fmt.Sprintf("unknown column summariser: %s", ss))
		}
	}
	// Done
	return summarisers
}

func flattenIncludes(includes []string) []string {
	count := 0
	// Determine total number of columns
	for _, s := range includes {
		extras := strings.Count(s, ",")
		if extras > 0 {
			count += extras
		}

		count++
	}
	// Expand (if necessary)
	if count != len(includes) {
		nincludes := make([]string, count)
		index := 0
		// Process each include
		for _, s := range includes {
			if strings.Contains(s, ",") {
				for _, t := range strings.Split(s, ",") {
					nincludes[index] = t
					index++
				}
			} else {
				nincludes[index] = s
				index++
			}
		}
		// Done
		includes = nincludes
	}
	// Done
	return includes
}

func summariseColumn(column trace.RawColumn, summarisers []ColSummariser) []string {
	m := 1 + uint(len(summarisers))
	//
	row := make([]string, m)
	row[0] = column.QualifiedName()
	// Generate each summary
	for j := 0; j < len(summarisers); j++ {
		row[j+1] = summarisers[j].summary(column)
	}
	// Done
	return row
}

func summaryStats(tr []trace.RawColumn) {
	m := uint(len(trSummarisers))
	tbl := util.NewTablePrinter(2, m)
	// Go!
	for i := uint(0); i < m; i++ {
		ith := trSummarisers[i]
		summary := ith.summary(tr)
		tbl.SetRow(i, ith.name, summary)
	}
	//
	tbl.SetMaxWidths(64)
	tbl.Print()
}

// ============================================================================
// Column Summarisers
// ============================================================================

// ColSummariser abstracts the notion of a function which summarises the
// contents of a given column.
type ColSummariser struct {
	name        string
	description string
	summary     func(trace.RawColumn) string
}

var colSummarisers []ColSummariser = []ColSummariser{
	{"lines", "line count for column", lineCountSummariser},
	{"bitwidth", "bitwidth for column as specified in trace file", bitWidthSummariser},
	{"bytes", "total bytes required for column", bytesSummariser},
	{"elements", "number of unique elements in column", uniqueElementsSummariser},
	{"entropy", "number of lines in column whose value differs from previous line", entropySummariser},
}

// Used to show the available options on the command-line.
func summariserOptions() string {
	summarisers := "\n"
	//
	for _, s := range colSummarisers {
		summarisers = fmt.Sprintf("%s--- %s (%s)\n", summarisers, s.name, s.description)
	}
	//
	return summarisers
}

func lineCountSummariser(col trace.RawColumn) string {
	return fmt.Sprintf("%d", col.Data.Len())
}

func bitWidthSummariser(col trace.RawColumn) string {
	return fmt.Sprintf("%d", col.Data.BitWidth())
}

func bytesSummariser(col trace.RawColumn) string {
	bitwidth := col.Data.BitWidth()
	byteWidth := bitwidth / 8
	// Determine proper bytewidth
	if bitwidth%8 != 0 {
		byteWidth++
	}

	return fmt.Sprintf("%d", col.Data.Len()*byteWidth)
}

func uniqueElementsSummariser(col trace.RawColumn) string {
	data := col.Data
	elems := hash.NewSet[hash.BytesKey](data.Len() / 2)
	// Add all the elements
	for i := uint(0); i < data.Len(); i++ {
		bytes := util.FrElementToBytes(data.Get(i))
		elems.Insert(hash.NewBytesKey(bytes[:]))
	}
	// Done
	return fmt.Sprintf("%d", elems.Size())
}

func entropySummariser(col trace.RawColumn) string {
	data := col.Data
	entropy := 0.0
	//
	if data.Len() > 0 {
		last := data.Get(0)
		count := 1
		// Count all rows which have same value as previous row.
		for i := uint(1); i < data.Len(); i++ {
			ith := data.Get(i)
			if last.Cmp(&ith) == 0 {
				count++
			}
		}
		// Calculate entropy
		entropy = float64(count) / float64(data.Len())
		entropy *= 100
	}
	// Done
	return fmt.Sprintf("%2.1f%%", entropy)
}

// ============================================================================
// Trace Summarisers
// ============================================================================

type traceSummariser struct {
	name    string
	summary func([]trace.RawColumn) string
}

var trSummarisers []traceSummariser = []traceSummariser{
	trWidthSummariser(1, 8),
	trWidthSummariser(9, 16),
	trWidthSummariser(17, 32),
	trWidthSummariser(33, 128),
	trWidthSummariser(129, 256),
}

func trWidthSummariser(lowWidth uint, highWidth uint) traceSummariser {
	return traceSummariser{
		name: fmt.Sprintf("# Columns (%d..%d bits)", lowWidth, highWidth),
		summary: func(tr []trace.RawColumn) string {
			count := 0
			for i := 0; i < len(tr); i++ {
				ithWidth := tr[i].Data.BitWidth()
				if ithWidth >= lowWidth && ithWidth <= highWidth {
					count++
				}
			}
			return fmt.Sprintf("%d", count)
		},
	}
}
