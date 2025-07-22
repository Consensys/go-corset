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

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/corset"
	"github.com/consensys/go-corset/pkg/mir"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/trace/lt"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/hash"
	"github.com/consensys/go-corset/pkg/util/collection/set"
	"github.com/consensys/go-corset/pkg/util/termio"
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
		var (
			traces       [][]trace.RawColumn
			corsetConfig corset.CompilationConfig
		)
		//
		expand := GetFlag(cmd, "expand")
		// Sanity check
		if (expand && len(args) != 2) || (!expand && len(args) != 1) {
			fmt.Println(cmd.UsageString())
			os.Exit(1)
		}
		//
		optimisation := GetUint(cmd, "opt")
		// Set optimisation level
		if optimisation >= uint(len(mir.OPTIMISATION_LEVELS)) {
			fmt.Printf("invalid optimisation level %d\n", optimisation)
			os.Exit(2)
		}
		//
		optConfig := mir.OPTIMISATION_LEVELS[optimisation]
		// Parse trace
		columns := GetFlag(cmd, "columns")
		modules := GetFlag(cmd, "modules")
		defensive := GetFlag(cmd, "defensive")
		validate := GetFlag(cmd, "validate")
		stats := GetFlag(cmd, "stats")
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
		metadata := GetFlag(cmd, "metadata")
		//
		corsetConfig.Stdlib = !GetFlag(cmd, "no-stdlib")
		corsetConfig.Legacy = GetFlag(cmd, "legacy")
		// override for legacy lookups
		if cmd.Flags().Lookup("legacy-lookups").Changed {
			optConfig.LegacyLookups = GetFlag(cmd, "legacy-lookups")
		}
		// Parse trace file(s)
		if batched {
			// batched mode
			traces = ReadBatchedTraceFile(args[0])
		} else {
			// unbatched (i.e. normal) mode
			tracefile := ReadTraceFile(args[0])
			traces = [][]trace.RawColumn{tracefile.Columns}
			// Print meta-data (if requested)
			if metadata {
				printTraceFileHeader(&tracefile.Header)
			}
		}
		//
		if expand && !air && !mir && !hir {
			fmt.Println("must specify --hir/mir/air for trace expansion")
			os.Exit(2)
		} else if expand {
			level := determineAbstractionLevel(air, mir, hir)
			for i, cols := range traces {
				traces[i] = expandWithConstraints(level, cols, corsetConfig, validate, defensive, args[1:], optConfig)
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
			if columns {
				listColumns(max_width, traces[i], includes)
			}
			if modules {
				listModules(max_width, traces[i])
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
	traceCmd.Flags().BoolP("columns", "c", false, "show column stats for the trace file")
	traceCmd.Flags().BoolP("modules", "m", false, "show module stats for the trace file")
	traceCmd.Flags().StringArrayP("include", "i", []string{"lines", "bitwidth", "bytes", "elements"},
		fmt.Sprintf("specify information to include in column listing: %s", summariserOptions()))
	traceCmd.Flags().Bool("stats", false, "show overall stats for the trace file")
	traceCmd.Flags().BoolP("print", "p", false, "print entire trace file")
	traceCmd.Flags().BoolP("expand", "e", false, "perform trace expansion (schema required)")
	traceCmd.Flags().Bool("defensive", false, "perform defensive padding (schema required)")
	traceCmd.Flags().Bool("validate", true, "apply trace validation")
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
	traceCmd.Flags().Bool("metadata", false, "Print embedded metadata")
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

func expandWithConstraints(level int, cols []trace.RawColumn, corsetConfig corset.CompilationConfig, validate bool,
	defensive bool, filenames []string, optConfig mir.OptimisationConfig) []trace.RawColumn {
	//
	var schema sc.Schema
	//
	binfile := ReadConstraintFiles(corsetConfig, filenames)
	//
	switch level {
	case hir_LEVEL:
		schema = &binfile.Schema
	case mir_LEVEL:
		schema = binfile.Schema.LowerToMir()
	case air_LEVEL:
		schema = binfile.Schema.LowerToMir().LowerToAir(optConfig)
	default:
		panic("unreachable")
	}
	// Done
	return expandColumns(cols, schema, validate, defensive)
}

func expandColumns(cols []trace.RawColumn, schema sc.Schema, validate bool, defensive bool) []trace.RawColumn {
	builder := sc.NewTraceBuilder(schema).Expand(true).Validate(validate).Defensive(defensive)
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

func printTraceFileHeader(header *lt.Header) {
	fmt.Printf("Format: %d.%d\n", header.MajorVersion, header.MinorVersion)
	// Attempt to parse metadata
	metadata, err := header.GetMetaData()
	//
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	} else if !metadata.IsEmpty() {
		fmt.Println("Metadata:")
		//
		printTypedMetadata(1, metadata)
	}
}

func printTrace(start uint, max_width uint, cols []trace.RawColumn) {
	n := uint(len(cols))
	height := maxHeightColumns(cols)
	tbl := termio.NewTablePrinter(1+height, 1+n)

	for j := uint(0); j < height; j++ {
		tbl.Set(j+1, 0, termio.NewText(fmt.Sprintf("#%d", j+start)))
	}

	for i := uint(0); i < n; i++ {
		ith := cols[i].Data
		tbl.Set(0, i+1, termio.NewText(cols[i].QualifiedName()))

		for j := uint(0); j < ith.Len(); j++ {
			jth := ith.Get(j)

			tbl.Set(j+1, i+1, termio.NewText(jth.Text(16)))
		}
	}
	//
	tbl.SetMaxWidths(max_width)
	tbl.Print(true)
}

func listModules(max_width uint, tr []trace.RawColumn) {
	// Organise traces by their module ID
	traces, modules := organiseTracesByModule(tr)
	//
	summarisers := moduleSumarisers
	m := 1 + uint(len(summarisers))
	n := uint(len(modules))
	// Go!
	tbl := termio.NewTablePrinter(m, n+1)
	// Set column titles
	for i := uint(0); i < uint(len(summarisers)); i++ {
		tbl.Set(i+1, 0, termio.NewText(summarisers[i].name))
	}
	// Compute column data
	for i, mod := range modules {
		row := make([]termio.FormattedText, m)
		//
		row[0] = termio.NewText(mod)
		//
		for j, s := range summarisers {
			row[j+1] = termio.NewText(s.summary(traces[mod]))
		}
		//
		tbl.SetRow(uint(i+1), row...)
	}
	//
	tbl.SetMaxWidths(max_width)
	tbl.Print(true)
}

func listColumns(max_width uint, tr []trace.RawColumn, includes []string) {
	summarisers := selectColumnSummarisers(includes)
	m := 1 + uint(len(summarisers))
	n := uint(len(tr))
	// Go!
	tbl := termio.NewTablePrinter(m, n+1)
	c := make(chan util.Pair[uint, []termio.FormattedText], n)
	// Set titles
	tbl.Set(0, 0, termio.NewText("Column"))

	for i := uint(0); i < uint(len(summarisers)); i++ {
		tbl.Set(i+1, 0, termio.NewText(summarisers[i].name))
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
	tbl.SetMaxWidths(max_width)
	tbl.Print(true)
}

func selectColumnSummarisers(includes []string) []ColumnSummariser {
	includes = flattenIncludes(includes)
	summarisers := make([]ColumnSummariser, len(includes))
	// Iterate included summarisers
	for i, ss := range includes {
		// Look them up
		for _, cs := range columnSummarisers {
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

func summariseColumn(column trace.RawColumn, summarisers []ColumnSummariser) []termio.FormattedText {
	m := 1 + uint(len(summarisers))
	//
	row := make([]termio.FormattedText, m)
	row[0] = termio.NewText(column.QualifiedName())
	// Generate each summary
	for j := 0; j < len(summarisers); j++ {
		row[j+1] = termio.NewText(summarisers[j].summary(column))
	}
	// Done
	return row
}

func summaryStats(tr []trace.RawColumn) {
	m := uint(len(trSummarisers))
	tbl := termio.NewTablePrinter(2, m)
	// Go!
	for i := uint(0); i < m; i++ {
		ith := trSummarisers[i]
		summary := ith.summary(tr)
		tbl.SetRow(i, termio.NewText(ith.name), termio.NewText(summary))
	}
	//
	tbl.SetMaxWidths(64)
	tbl.Print(true)
}

func organiseTracesByModule(columns []trace.RawColumn) (map[string][]trace.RawColumn, []string) {
	keys := set.NewSortedSet[string]()
	mapping := make(map[string][]trace.RawColumn)
	//
	for _, col := range columns {
		mod := col.Module
		traces := mapping[mod]
		//
		mapping[mod] = append(traces, col)
		// Insert module name
		keys.Insert(mod)
	}
	//
	return mapping, keys.Iter().Collect()
}

// ============================================================================
// Module Summarisers
// ============================================================================

// ModuleSummariser abstracts the notion of a function which summarises the
// contents of a given column.
type ModuleSummariser struct {
	name        string
	description string
	summary     func([]trace.RawColumn) string
}

var moduleSumarisers = []ModuleSummariser{
	{"columns", "column count for module", moduleColumnSummariser},
	{"lines", "line count for module", moduleLineSummariser},
	{"bitwidth", "bitwidth of module", moduleBitwidthSummariser},
	{"cells", "total number of cells traced for module", moduleCountSummariser},
	{"nonzero", "total number of nonzero cells traced for module", moduleNonZeroCounter},
	{"bytes", "total number of bytes traced for module", moduleBytesSummariser},
}

func moduleCountSummariser(columns []trace.RawColumn) string {
	count := 0

	for _, col := range columns {
		count += int(col.Data.Len())
	}
	//
	return fmt.Sprintf("%d", count)
}

func moduleColumnSummariser(columns []trace.RawColumn) string {
	return fmt.Sprintf("%d", len(columns))
}

func moduleLineSummariser(columns []trace.RawColumn) string {
	var lines uint

	if len(columns) == 0 {
		lines = 0
	} else {
		lines = math.MaxUint
		// NOTE: we take the minimum here because its possible that some columns
		// have a multiplier, which means their length is a longer than the
		// others.
		for _, c := range columns {
			lines = min(lines, c.Data.Len())
		}
	}
	//
	return fmt.Sprintf("%d", lines)
}

func moduleBitwidthSummariser(columns []trace.RawColumn) string {
	total := uint(0)
	//
	for _, c := range columns {
		total += c.Data.BitWidth()
	}
	//
	return fmt.Sprintf("%d", total)
}

func moduleBytesSummariser(columns []trace.RawColumn) string {
	total := uint(0)
	//
	for _, c := range columns {
		bitwidth := c.Data.BitWidth()
		byteWidth := bitwidth / 8
		// Determine proper bytewidth
		if bitwidth%8 != 0 {
			byteWidth++
		}
		//
		total += c.Data.Len() * byteWidth
	}
	//
	return fmt.Sprintf("%d", total)
}

func moduleNonZeroCounter(columns []trace.RawColumn) string {
	count := uint(0)

	for _, col := range columns {
		count += nonZeroCount(col)
	}
	//
	return fmt.Sprintf("%d", count)
}

// ============================================================================
// Column Summarisers
// ============================================================================

// ColumnSummariser abstracts the notion of a function which summarises the
// contents of a given column.
type ColumnSummariser struct {
	name        string
	description string
	summary     func(trace.RawColumn) string
}

var columnSummarisers = []ColumnSummariser{
	{"lines", "line count for column", columnCountSummariser},
	{"bitwidth", "bitwidth for column as specified in trace file", columnBitwidthSummariser},
	{"bytes", "total bytes required for column", columnBytesSummariser},
	{"elements", "number of unique elements in column", uniqueElementsSummariser},
	{"entropy", "number of lines in column whose value differs from previous line", entropySummariser},
	{"nonzero", "number of lines in column whose value is not zero", nonZeroCounter},
}

// Used to show the available options on the command-line.
func summariserOptions() string {
	summarisers := "\n"
	//
	for _, s := range columnSummarisers {
		summarisers = fmt.Sprintf("%s--- %s (%s)\n", summarisers, s.name, s.description)
	}
	//
	return summarisers
}

func columnCountSummariser(col trace.RawColumn) string {
	return fmt.Sprintf("%d", col.Data.Len())
}

func columnBitwidthSummariser(col trace.RawColumn) string {
	return fmt.Sprintf("%d", col.Data.BitWidth())
}

func columnBytesSummariser(col trace.RawColumn) string {
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
		var (
			last  fr.Element = data.Get(0)
			count            = 1
		)
		// Count all rows whose value differs from previous row.
		for i := uint(1); i < data.Len(); i++ {
			ith := data.Get(i)
			if last.Cmp(&ith) != 0 {
				count++
			}
			//
			last = ith
		}
		// Calculate entropy
		entropy = float64(count*100) / float64(data.Len())
	}
	// Done
	return fmt.Sprintf("%2.1f%%", entropy)
}

func nonZeroCounter(col trace.RawColumn) string {
	return fmt.Sprintf("%d", nonZeroCount(col))
}

func nonZeroCount(col trace.RawColumn) uint {
	var (
		count = uint(0)
		zero  = fr.NewElement(0)
		data  = col.Data
	)
	//
	if data.Len() > 0 {
		// Count all rows which have same value as previous row.
		for i := uint(1); i < data.Len(); i++ {
			ith := data.Get(i)
			if ith.Cmp(&zero) != 0 {
				count++
			}
		}
	}
	//
	return count
}

// ============================================================================
// Trace Summarisers
// ============================================================================

type traceSummariser struct {
	name    string
	summary func([]trace.RawColumn) string
}

var trSummarisers []traceSummariser = []traceSummariser{
	{"Cells", trCellCountSummariser},
	{"Cells (raw)", trRawCellCountSummariser},
	trWidthSummariser(1, 8),
	trWidthSummariser(9, 16),
	trWidthSummariser(17, 32),
	trWidthSummariser(33, 128),
	trWidthSummariser(129, 256),
}

func trRawCellCount(cols []trace.RawColumn) uint {
	total := uint(0)
	//
	for _, col := range cols {
		total += col.Data.Len()
	}
	//
	return total
}

func trRawCellCountSummariser(cols []trace.RawColumn) string {
	total := trRawCellCount(cols)
	return fmt.Sprintf("%d", total)
}

const one_K = 1000
const one_M = one_K * one_K
const one_G = one_M * one_M

func trCellCountSummariser(cols []trace.RawColumn) string {
	total := trRawCellCount(cols)
	//
	switch {
	case total > one_G:
		val := float64(total) / one_G
		return fmt.Sprintf("%.01fG", val)
	case total > one_M:
		val := float64(total) / one_M
		return fmt.Sprintf("%.01fM", val)
	case total > one_K:
		val := float64(total) / one_K
		return fmt.Sprintf("%.01fK", val)
	default:
		return fmt.Sprintf("%d", total)
	}
}

func trWidthSummariser(lowWidth uint, highWidth uint) traceSummariser {
	return traceSummariser{
		name: fmt.Sprintf("Columns (%d..%d bits)", lowWidth, highWidth),
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
