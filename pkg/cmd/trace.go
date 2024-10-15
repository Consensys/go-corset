package cmd

import (
	"fmt"
	"math"
	"os"
	"regexp"
	"strings"

	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/spf13/cobra"
)

// traceCmd represents the trace command for manipulating traces.
var traceCmd = &cobra.Command{
	Use:   "trace [flags] trace_file",
	Short: "Operate on a trace file.",
	Long: `Operate on a trace file, such as converting
	it from one format (e.g. lt) to another (e.g. json),
	or filtering out modules, or listing columns, etc.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Println(cmd.UsageString())
			os.Exit(1)
		}
		// Parse trace
		cols := readTraceFile(args[0])
		list := getFlag(cmd, "list")
		stats := getFlag(cmd, "stats")
		includes := getStringArray(cmd, "include")
		print := getFlag(cmd, "print")
		start := getUint(cmd, "start")
		end := getUint(cmd, "end")
		max_width := getUint(cmd, "max-width")
		filter := getString(cmd, "filter")
		output := getString(cmd, "out")
		// construct filters
		if filter != "" {
			cols = filterColumns(cols, filter)
		}
		if list {
			listColumns(cols, includes)
		}
		if stats {
			summaryStats(cols)
		}
		//
		if output != "" {
			writeTraceFile(output, cols)
		}

		if print {
			printTrace(start, end, max_width, cols)
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
	traceCmd.Flags().Uint("start", 0, "filter out rows below this")
	traceCmd.Flags().Uint("end", math.MaxUint, "filter out this and all following rows")
	traceCmd.Flags().Uint("max-width", 32, "specify maximum display width for a column")
	traceCmd.Flags().StringP("out", "o", "", "Specify output file to write trace")
	traceCmd.Flags().StringP("filter", "f", "", "Filter columns matching regex")
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

func printTrace(start uint, end uint, max_width uint, cols []trace.RawColumn) {
	n := uint(len(cols))
	height := min(maxHeightColumns(cols), end) - start
	tbl := util.NewTablePrinter(1+height, 1+n)

	for j := uint(0); j < height; j++ {
		tbl.Set(j+1, 0, fmt.Sprintf("#%d", j+start))
	}

	for i := uint(0); i < n; i++ {
		ith := cols[i].Data
		tbl.Set(0, i+1, cols[i].QualifiedName())

		if start < ith.Len() {
			ith_height := min(ith.Len(), end) - start
			for j := uint(0); j < ith_height; j++ {
				jth := ith.Get(j + start)

				tbl.Set(j+1, i+1, jth.Text(16))
			}
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
	elems := util.NewHashSet[util.BytesKey](data.Len() / 2)
	// Add all the elements
	for i := uint(0); i < data.Len(); i++ {
		bytes := util.FrElementToBytes(data.Get(i))
		elems.Insert(util.NewBytesKey(bytes[:]))
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
