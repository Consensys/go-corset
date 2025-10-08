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
	"github.com/consensys/go-corset/pkg/asm"
	"github.com/consensys/go-corset/pkg/binfile"
	cmd "github.com/consensys/go-corset/pkg/cmd/util"
	"github.com/consensys/go-corset/pkg/cmd/view"
	"github.com/consensys/go-corset/pkg/corset"
	"github.com/consensys/go-corset/pkg/ir"
	"github.com/consensys/go-corset/pkg/trace"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/trace/lt"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/array"
	"github.com/consensys/go-corset/pkg/util/collection/hash"
	"github.com/consensys/go-corset/pkg/util/collection/pool"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/field/bls12_377"
	"github.com/consensys/go-corset/pkg/util/field/gf251"
	"github.com/consensys/go-corset/pkg/util/field/gf8209"
	"github.com/consensys/go-corset/pkg/util/field/koalabear"
	"github.com/consensys/go-corset/pkg/util/termio"
	"github.com/consensys/go-corset/pkg/util/termio/widget"
	"github.com/consensys/go-corset/pkg/util/word"
	log "github.com/sirupsen/logrus"
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
		runFieldAgnosticCmd(cmd, args, traceCmds)
	},
}

// Available instances
var traceCmds = []FieldAgnosticCmd{
	{field.GF_251, runTraceCmd[gf251.Element]},
	{field.GF_8209, runTraceCmd[gf8209.Element]},
	{field.KOALABEAR_16, runTraceCmd[koalabear.Element]},
	{field.BLS12_377, runTraceCmd[bls12_377.Element]},
}

func runTraceCmd[F field.Element[F]](cmd *cobra.Command, args []string) {
	var (
		ltTraces []lt.TraceFile
		traces   []tr.Trace[F]
		cfg      TraceConfig
	)
	// Configure log level
	if GetFlag(cmd, "verbose") {
		log.SetLevel(log.DebugLevel)
	}
	// Parse trace
	//columns := GetFlag(cmd, "columns")
	batched := GetFlag(cmd, "batched")
	//modules := GetFlag(cmd, "modules")
	//stats := GetFlag(cmd, "stats")
	//includes := GetStringArray(cmd, "include")
	print := GetFlag(cmd, "print")
	cfg.startRow = GetUint(cmd, "start")
	cfg.endRow = GetUint(cmd, "end")
	cfg.maxCellWidth = GetUint(cmd, "max-width")
	padding := GetUint(cmd, "padding")
	//filter := GetString(cmd, "filter")
	output := GetString(cmd, "out")
	metadata := GetFlag(cmd, "metadata")
	ltv2 := GetFlag(cmd, "ltv2")
	//sort := GetUint(cmd, "sort")
	// Read in constraint files
	stacker := *getSchemaStack[F](cmd, SCHEMA_OPTIONAL, args[1:]...)
	stack := stacker.Build()
	builder := stack.TraceBuilder().WithPadding(padding)
	// Extract debug information (if available)
	cfg.sourceMap, _ = binfile.GetAttribute[*corset.SourceMap](stacker.BinaryFile())
	// Extract register mapping (for limbs)
	cfg.mapping = stack.RegisterMapping()
	// Parse trace file(s)
	if batched {
		// batched mode
		ltTraces = ReadBatchedTraceFile(args[0])
	} else {
		// unbatched (i.e. normal) mode
		ltTraces = []lt.TraceFile{ReadTraceFile(args[0])}
		// Print meta-data (if requested)
		if metadata {
			printTraceFileHeader(&ltTraces[0].Header)
		}
	}
	//
	if builder.Expanding() && !stack.HasUniqueSchema() {
		fmt.Println("must specify one of --asm/uasm/mir/air")
		os.Exit(2)
	} else if !builder.Expanding() {
		fmt.Println("non-expanding trace command currently unsupported")
		os.Exit(2)
	}
	// Expand all the traces
	for _, cols := range ltTraces {
		traces = append(traces, expandLtTrace(cols, stack, builder))
	}
	// Now manipulate traces
	for i := range ltTraces {
		// construct filters
		// if filter != "" {
		// 	traces[i] = filterColumns(traces[i], filter)
		// }

		// if start != 0 || end != math.MaxUint {
		// 	sliceColumns(traces[i], start, end)
		// }

		// if columns {
		// 	listColumns(max_width, sort, traces[i], includes)
		// }

		// if modules {
		// 	listModules(max_width, sort, traces[i])
		// }

		// if stats {
		// 	summaryStats(traces[i])
		// }

		if print {
			printTrace(cfg, traces[i])
		}
	}
	// Write out results (if requested)
	if output != "" {
		// Convert all traces back to lt files.
		for i := range traces {
			ltTraces[i] = seqReconstructRawTrace(ltTraces[i].Header.MetaData, traces[i])
			// Upgrade to ltv2 if requested
			if ltv2 {
				ltTraces[i].Header.MajorVersion = lt.LTV2_MAJOR_VERSION
			}
		}
		//
		writeBatchedTracesFile(output, ltTraces...)
	}
}

func init() {
	rootCmd.AddCommand(traceCmd)
	traceCmd.Flags().BoolP("columns", "c", false, "show column stats for the trace file")
	traceCmd.Flags().BoolP("modules", "m", false, "show module stats for the trace file")
	traceCmd.Flags().StringArrayP("include", "i", []string{"lines", "bitwidth", "bytes", "elements"},
		fmt.Sprintf("specify information to include in column listing: %s", summariserOptions()))
	traceCmd.Flags().Bool("stats", false, "show overall stats for the trace file")
	traceCmd.Flags().BoolP("print", "p", false, "print entire trace file")
	traceCmd.Flags().Uint("sort", 0, "sort table column")
	traceCmd.Flags().Uint("start", 0, "filter out rows below this")
	traceCmd.Flags().Uint("end", math.MaxUint, "filter out this and all following rows")
	traceCmd.Flags().Uint("max-width", 32, "specify maximum display width for a column")
	traceCmd.Flags().Uint("padding", 0, "specify amount of (front) padding to apply")
	traceCmd.Flags().StringP("out", "o", "", "Specify output file to write trace")
	traceCmd.Flags().StringP("filter", "f", "", "Filter columns matching regex")
	traceCmd.Flags().Bool("batched", false,
		"specify trace file is batched (i.e. contains multiple traces, one for each line)")
	traceCmd.Flags().Bool("metadata", false, "Print embedded metadata")
	traceCmd.Flags().Bool("ltv2", false, "Use ltv2 file format")
}

// TraceConfig packages together useful things for the various supported
// options.
type TraceConfig struct {
	mapping       sc.LimbsMap
	sourceMap     *corset.SourceMap
	maxCellWidth  uint
	maxTitleWidth uint
	limbs         bool
	startRow      uint
	endRow        uint
}

// RawColumn provides a convenient alias
type RawColumn = lt.Column[word.BigEndian]

func expandLtTrace[F field.Element[F]](tf lt.TraceFile, stack cmd.SchemaStack[F], bldr ir.TraceBuilder[F],
) tr.Trace[F] {
	//
	var (
		schema = stack.BinaryFile().Schema
		errors []error
		tr     trace.Trace[F]
	)
	// Apply trace propagation
	if bldr.Expanding() {
		perf := util.NewPerfStats()
		//
		tf, errors = asm.Propagate(schema, tf, true)
		//
		perf.Log("Trace propagation")
	}
	//
	if len(errors) == 0 {
		// Construct expanded trace
		tr, errors = bldr.Build(stack.UniqueConcreteSchema(), tf)
	}
	// Handle errors
	if len(errors) > 0 {
		for _, err := range errors {
			fmt.Println(err)
		}
		//
		os.Exit(1)
	}
	// Now, reconstruct it!
	return tr
}

// NOTE: parallelising this algorithm did not improve performance as there is
// (presumably) too much contention.  A better solution would be to construct
// local heaps for each column in parallel and then merge them at the end.  But
// that remains complex since the indexing will be different.
func seqReconstructRawTrace[F field.Element[F]](metadata []byte, tr trace.Trace[F]) lt.TraceFile {
	var (
		perf     = util.NewPerfStats()
		expanded = make([]lt.Module[word.BigEndian], tr.Width())
		// Construct fresh heap for this trace
		heap       = pool.NewLocalHeap[word.BigEndian]()
		arrBuilder = array.NewDynamicBuilder(heap)
	)
	//
	for mid := range tr.Width() {
		var (
			module  = tr.Module(mid)
			columns = make([]lt.Column[word.BigEndian], module.Width())
		)
		// Initialise modules
		// Dispatch go-routines
		for cid := range module.Width() {
			col := module.Column(cid)
			//
			columns[cid] = lt.Column[word.BigEndian]{
				Name: col.Name(),
				Data: array.CloneArray(col.Data(), &arrBuilder),
			}
		}
		//
		expanded[mid] = lt.Module[word.BigEndian]{
			Name:    module.Name(),
			Columns: columns,
		}
	}
	//
	perf.Log("Trace reconstruction")
	//
	return lt.NewTraceFile(metadata, *heap, expanded)
}

// Construct a new trace containing only those columns from the original who
// name begins with the given prefix.
func filterColumns(tf lt.TraceFile, regex string) lt.TraceFile {
	var (
		r, err  = regexp.Compile(regex)
		modules = make([]lt.Module[word.BigEndian], len(tf.Modules))
	)
	// Check for error
	if err != nil {
		panic(err)
	}
	//
	for i, ith := range tf.Modules {
		var columns []lt.Column[word.BigEndian]
		// Now create the columns.
		for _, jth := range ith.Columns {
			name := trace.QualifiedColumnName(ith.Name, jth.Name)
			if r.MatchString(name) {
				columns = append(columns, jth)
			}
		}
		// Construct new (potentially empty) module
		modules[i] = lt.Module[word.BigEndian]{Name: ith.Name, Columns: columns}
	}
	// Done
	return lt.NewTraceFile(tf.Header.MetaData, tf.Heap, modules)
}

// Construct a new trace where all columns are sliced to a given region.  In
// some cases, that might mean the column becomes entirely empty.
func sliceColumns(tf lt.TraceFile, start uint, end uint) {
	// Now slice them columns.
	for i, ith := range tf.Modules {
		for j, jth := range ith.Columns {
			s := min(start, jth.Data.Len())
			e := min(end, jth.Data.Len())
			// Not pretty, but it works :)
			data := jth.Data.Slice(s, e).(array.MutArray[word.BigEndian])
			//
			tf.Modules[i].Columns[j] = lt.Column[word.BigEndian]{
				Name: jth.Name,
				Data: data,
			}
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

func printTrace[F field.Element[F]](cfg TraceConfig, trace tr.Trace[F]) {
	// Construct trace window
	window := view.NewBuilder[F](cfg.mapping).
		WithCellWidth(cfg.maxCellWidth).
		WithSourceMap(*cfg.sourceMap).
		Build(trace)
	// Print all windows
	for i := range window.Width() {
		ith := window.Module(i)
		// Construct & configure printer
		tp := widget.NewTable(window.Module(i))
		// Print out module name
		if window.Width() > 1 && ith.Data().Name() != "" {
			fmt.Printf("%s:\n", ith.Data().Name())
		}
		// Print out report
		tp.Print()
		fmt.Println()
	}
}

func listModules(max_width uint, sort_col uint, tf lt.TraceFile) {
	var (
		//
		summarisers = moduleSumarisers
		m           = 1 + uint(len(summarisers))
		n           = uint(len(tf.Modules))
		// Go!
		tbl = termio.NewFormattedTable(m, n+1)
	)
	// Set column titles
	for i := uint(0); i < uint(len(summarisers)); i++ {
		tbl.Set(i+1, 0, termio.NewText(summarisers[i].name))
	}
	// Compute column data
	for i, mod := range tf.Modules {
		row := summariseModule(mod, moduleSumarisers)
		// Set row
		tbl.SetRow(uint(i+1), row...)
	}
	//
	tbl.SetMaxWidths(max_width)
	tbl.Sort(1, termio.NewTableSorter().
		SortNumericalColumn(sort_col).
		Invert())
	tbl.Print(true)
}

func listColumns(max_width, sort_col uint, tf lt.TraceFile, includes []string) {
	var (
		summarisers = selectColumnSummarisers(includes)
		m           = 1 + uint(len(summarisers))
		n           = lt.NumberOfColumns(tf.Modules)
		// Go!
		tbl   = termio.NewFormattedTable(m, n+1)
		c     = make(chan util.Pair[uint, []termio.FormattedText], n)
		index uint
	)
	// Set titles
	tbl.Set(0, 0, termio.NewText("Column"))

	for i := uint(0); i < uint(len(summarisers)); i++ {
		tbl.Set(i+1, 0, termio.NewText(summarisers[i].name))
	}
	// Compute data
	for _, ith := range tf.Modules {
		for _, jth := range ith.Columns {
			// Launch summarisers
			go func(index uint) {
				// Apply summarisers to column
				row := summariseColumn(ith.Name, jth, summarisers)
				// Package result
				c <- util.NewPair(index, row)
			}(index)
			//
			index = index + 1
		}
	}
	// Collect results
	for range n {
		// Read packaged result from channel
		res := <-c
		// Set row
		tbl.SetRow(res.Left+1, res.Right...)
	}
	//
	tbl.SetMaxWidths(max_width)
	tbl.Sort(1, termio.NewTableSorter().
		SortNumericalColumn(sort_col).
		Invert())
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

func summariseModule(mod lt.Module[word.BigEndian], summarisers []ModuleSummariser) []termio.FormattedText {
	var (
		m   = 1 + uint(len(summarisers))
		row = make([]termio.FormattedText, m)
	)
	//
	row[0] = termio.NewText(mod.Name)
	//
	for j, s := range summarisers {
		row[j+1] = termio.NewText(s.summary(mod.Columns))
	}
	//
	return row
}

func summariseColumn(module string, column RawColumn, summarisers []ColumnSummariser) []termio.FormattedText {
	m := 1 + uint(len(summarisers))
	//
	row := make([]termio.FormattedText, m)
	row[0] = termio.NewText(fmt.Sprintf("%s.%s", module, column.Name))
	// Generate each summary
	for j := 0; j < len(summarisers); j++ {
		row[j+1] = termio.NewText(summarisers[j].summary(column))
	}
	// Done
	return row
}

func summaryStats(tf lt.TraceFile) {
	m := uint(len(trSummarisers))
	tbl := termio.NewFormattedTable(2, m)
	// Go!
	for i := uint(0); i < m; i++ {
		ith := trSummarisers[i]
		summary := ith.summary(tf.Modules)
		tbl.SetRow(i, termio.NewText(ith.name), termio.NewText(summary))
	}
	//
	tbl.SetMaxWidths(64)
	tbl.Print(true)
}

// ============================================================================
// Module Summarisers
// ============================================================================

// ModuleSummariser abstracts the notion of a function which summarises the
// contents of a given column.
type ModuleSummariser struct {
	name        string
	description string
	summary     func([]RawColumn) string
}

var moduleSumarisers = []ModuleSummariser{
	{"columns", "column count for module", moduleColumnSummariser},
	{"lines", "line count for module", moduleLineSummariser},
	{"bitwidth", "bitwidth of module", moduleBitwidthSummariser},
	{"cells", "total number of cells traced for module", moduleCountSummariser},
	{"nonzero", "total number of nonzero cells traced for module", moduleNonZeroCounter},
	{"bytes", "total number of bytes traced for module", moduleBytesSummariser},
}

func moduleCountSummariser(columns []RawColumn) string {
	count := 0

	for _, col := range columns {
		count += int(col.Data.Len())
	}
	//
	return fmt.Sprintf("%d", count)
}

func moduleColumnSummariser(columns []RawColumn) string {
	return fmt.Sprintf("%d", len(columns))
}

func moduleLineSummariser(columns []RawColumn) string {
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

func moduleBitwidthSummariser(columns []RawColumn) string {
	total := uint(0)
	//
	for _, c := range columns {
		total += bitwidth(c.Data)
	}
	//
	return fmt.Sprintf("%d", total)
}

func moduleBytesSummariser(columns []RawColumn) string {
	total := uint(0)
	//
	for _, c := range columns {
		bitwidth := bitwidth(c.Data)
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

func moduleNonZeroCounter(columns []RawColumn) string {
	count := uint(0)

	for _, col := range columns {
		count += nonZeroCount(col)
	}
	//
	return fmt.Sprintf("%d", count)
}

func bitwidth[T any](arr array.Array[T]) uint {
	if arr.BitWidth() == math.MaxUint {
		return uint(fr.Modulus().BitLen())
	}

	return arr.BitWidth()
}

// ============================================================================
// Column Summarisers
// ============================================================================

// ColumnSummariser abstracts the notion of a function which summarises the
// contents of a given column.
type ColumnSummariser struct {
	name        string
	description string
	summary     func(RawColumn) string
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

func columnCountSummariser(col RawColumn) string {
	return fmt.Sprintf("%d", col.Data.Len())
}

func columnBitwidthSummariser(col RawColumn) string {
	return fmt.Sprintf("%d", bitwidth(col.Data))
}

func columnBytesSummariser(col RawColumn) string {
	bitwidth := bitwidth(col.Data)
	byteWidth := bitwidth / 8
	// Determine proper bytewidth
	if bitwidth%8 != 0 {
		byteWidth++
	}

	return fmt.Sprintf("%d", col.Data.Len()*byteWidth)
}

func uniqueElementsSummariser(col RawColumn) string {
	data := col.Data
	elems := hash.NewSet[word.BigEndian](data.Len() / 2)
	// Add all the elements
	for i := uint(0); i < data.Len(); i++ {
		elems.Insert(data.Get(i))
	}
	// Done
	return fmt.Sprintf("%d", elems.Size())
}

func entropySummariser(col RawColumn) string {
	data := col.Data
	entropy := 0.0
	//
	if data.Len() > 0 {
		var (
			last  word.BigEndian = data.Get(0)
			count                = 1
		)
		// Count all rows whose value differs from previous row.
		for i := uint(1); i < data.Len(); i++ {
			ith := data.Get(i)
			if last.Cmp(ith) != 0 {
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

func nonZeroCounter(col RawColumn) string {
	return fmt.Sprintf("%d", nonZeroCount(col))
}

func nonZeroCount(col RawColumn) uint {
	var (
		count = uint(0)
		data  = col.Data
	)
	//
	if data.Len() > 0 {
		// Count all rows which have same value as previous row.
		for i := uint(1); i < data.Len(); i++ {
			ith := data.Get(i).AsBigInt()
			if ith.Sign() != 0 {
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
	summary func([]lt.Module[word.BigEndian]) string
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

func trRawCellCount(modules []lt.Module[word.BigEndian]) uint {
	total := uint(0)
	//
	for _, ith := range modules {
		for _, jth := range ith.Columns {
			total += jth.Data.Len()
		}
	}
	//
	return total
}

func trRawCellCountSummariser(modules []lt.Module[word.BigEndian]) string {
	total := trRawCellCount(modules)
	return fmt.Sprintf("%d", total)
}

const one_K = 1000
const one_M = one_K * one_K
const one_G = one_M * one_M

func trCellCountSummariser(modules []lt.Module[word.BigEndian]) string {
	total := trRawCellCount(modules)
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
		summary: func(tr []lt.Module[word.BigEndian]) string {
			count := 0
			for _, ith := range tr {
				for _, jth := range ith.Columns {
					ithWidth := bitwidth(jth.Data)
					if ithWidth >= lowWidth && ithWidth <= highWidth {
						count++
					}
				}
			}
			return fmt.Sprintf("%d", count)
		},
	}
}
