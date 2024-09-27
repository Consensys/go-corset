package cmd

import (
	"fmt"
	"math"
	"os"
	"regexp"

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
			listColumns(cols)
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
	traceCmd.Flags().Bool("stats", false, "print summary information about the trace file")
	traceCmd.Flags().BoolP("print", "p", false, "print entire trace file")
	traceCmd.Flags().UintP("start", "s", 0, "filter out rows below this")
	traceCmd.Flags().UintP("end", "e", math.MaxUint, "filter out this and all following rows")
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

func listColumns(tr []trace.RawColumn) {
	m := 1 + uint(len(colSummarisers))
	n := uint(len(tr))
	// Go!
	tbl := util.NewTablePrinter(m, n)
	c := make(chan util.Pair[uint, []string], n)
	//
	for i := uint(0); i < n; i++ {
		// Launch summarisers
		go func(index uint) {
			// Apply summarisers to column
			row := summariseColumn(tr[index])
			// Package result
			c <- util.NewPair(index, row)
		}(i)
	}
	// Collect results
	for i := uint(0); i < n; i++ {
		// Read packaged result from channel
		res := <-c
		// Set row
		tbl.SetRow(res.Left, res.Right...)
	}
	//
	tbl.SetMaxWidths(64)
	tbl.Print()
}

func summariseColumn(column trace.RawColumn) []string {
	m := 1 + uint(len(colSummarisers))
	//
	row := make([]string, m)
	row[0] = column.QualifiedName()
	// Generate each summary
	for j := 0; j < len(colSummarisers); j++ {
		row[j+1] = colSummarisers[j].summary(column)
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
	name    string
	summary func(trace.RawColumn) string
}

var colSummarisers []ColSummariser = []ColSummariser{
	{"count", rowSummariser},
	{"width", widthSummariser},
	{"bytes", bytesSummariser},
	{"unique", uniqueSummariser},
}

func rowSummariser(col trace.RawColumn) string {
	return fmt.Sprintf("%d rows", col.Data.Len())
}

func widthSummariser(col trace.RawColumn) string {
	return fmt.Sprintf("%d bits", col.Data.BitWidth())
}

func bytesSummariser(col trace.RawColumn) string {
	bitwidth := col.Data.BitWidth()
	byteWidth := bitwidth / 8
	// Determine proper bytewidth
	if bitwidth%8 != 0 {
		byteWidth++
	}

	return fmt.Sprintf("%d bytes", col.Data.Len()*byteWidth)
}

func uniqueSummariser(col trace.RawColumn) string {
	data := col.Data
	elems := util.NewHashSet[util.BytesKey](data.Len() / 2)
	// Add all the elements
	for i := uint(0); i < data.Len(); i++ {
		bytes := util.FrElementToBytes(data.Get(i))
		elems.Insert(util.NewBytesKey(bytes[:]))
	}
	// Done
	return fmt.Sprintf("%d elements", elems.Size())
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
