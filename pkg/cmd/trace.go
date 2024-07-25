package cmd

import (
	"fmt"
	"math"
	"os"
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
		tr := readTraceFile(args[0])
		list := getFlag(cmd, "list")
		stats := getFlag(cmd, "stats")
		print := getFlag(cmd, "print")
		padding := getUint(cmd, "pad")
		start := getUint(cmd, "start")
		end := getUint(cmd, "end")
		max_width := getUint(cmd, "max-width")
		filter := getString(cmd, "filter")
		output := getString(cmd, "out")
		// construct filters
		if filter != "" {
			tr = filterColumns(tr, filter)
		}
		if padding != 0 {
			trace.PadColumns(tr, padding)
		}
		if list {
			listColumns(tr)
		}
		if stats {
			summaryStats(tr)
		}
		//
		if output != "" {
			writeTraceFile(output, tr)
		}

		if print {
			printTrace(start, end, max_width, tr)
		}
	},
}

func init() {
	rootCmd.AddCommand(traceCmd)
	traceCmd.Flags().BoolP("list", "l", false, "list only the columns in the trace file")
	traceCmd.Flags().Bool("stats", false, "print summary information about the trace file")
	traceCmd.Flags().BoolP("print", "p", false, "print entire trace file")
	traceCmd.Flags().Uint("pad", 0, "add a given number of padding rows (to each module)")
	traceCmd.Flags().UintP("start", "s", 0, "filter out rows below this")
	traceCmd.Flags().UintP("end", "e", math.MaxUint, "filter out this and all following rows")
	traceCmd.Flags().Uint("max-width", 32, "specify maximum display width for a column")
	traceCmd.Flags().StringP("out", "o", "", "Specify output file to write trace")
	traceCmd.Flags().StringP("filter", "f", "", "Filter columns beginning with prefix")
}

// Construct a new trace containing only those columns from the original who
// name begins with the given prefix.
func filterColumns(tr trace.Trace, prefix string) trace.Trace {
	n := tr.Columns().Len()
	builder := trace.NewBuilder()
	// Initialise modules in the builder to ensure module indices are preserved
	// across traces.
	for i := uint(0); i < n; i++ {
		ith := tr.Columns().Get(i)
		name := tr.Modules().Get(ith.Context().Module()).Name()

		if !builder.HasModule(name) {
			if _, err := builder.Register(name, ith.Height()); err != nil {
				panic(err)
			}
		}
	}
	// Now create the columns.
	for i := uint(0); i < n; i++ {
		qName := QualifiedColumnName(i, tr)
		//
		if strings.HasPrefix(qName, prefix) {
			ith := tr.Columns().Get(i)
			// Copy column data
			data := ith.Data().Clone()
			err := builder.Add(qName, ith.Padding(), data)
			// Sanity check
			if err != nil {
				panic(err)
			}
		}
	}
	// Done
	return builder.Build()
}
func printTrace(start uint, end uint, max_width uint, tr trace.Trace) {
	cols := tr.Columns()
	n := tr.Columns().Len()
	height := min(trace.MaxHeight(tr), end) - start
	tbl := util.NewTablePrinter(1+height, 1+n)

	for j := uint(0); j < height; j++ {
		tbl.Set(j+1, 0, fmt.Sprintf("#%d", j+start))
	}

	for i := uint(0); i < n; i++ {
		ith := cols.Get(i)
		tbl.Set(0, i+1, QualifiedColumnName(i, tr))

		if start < ith.Height() {
			ith_height := min(ith.Height(), end) - start
			for j := uint(0); j < ith_height; j++ {
				tbl.Set(j+1, i+1, ith.Get(int(j+start)).String())
			}
		}
	}
	//
	tbl.SetMaxWidth(max_width)
	tbl.Print()
}

func listColumns(tr trace.Trace) {
	// Determine number of columns
	m := 1 + uint(len(colSummarisers))
	// Determine number of rows
	n := tr.Columns().Len()
	// Go!
	tbl := util.NewTablePrinter(m, n)

	for i := uint(0); i < n; i++ {
		ith := tr.Columns().Get(i)
		row := make([]string, m)
		row[0] = QualifiedColumnName(i, tr)
		// Add summarises
		for j := 0; j < len(colSummarisers); j++ {
			row[j+1] = colSummarisers[j].summary(ith)
		}
		tbl.SetRow(i, row...)
	}
	//
	tbl.SetMaxWidth(64)
	tbl.Print()
}

func summaryStats(tr trace.Trace) {
	m := uint(len(trSummarisers))
	tbl := util.NewTablePrinter(2, m)
	// Go!
	for i := uint(0); i < m; i++ {
		ith := trSummarisers[i]
		summary := ith.summary(tr)
		tbl.SetRow(i, ith.name, summary)
	}
	//
	tbl.SetMaxWidth(64)
	tbl.Print()
}

// ============================================================================
// Column Summarisers
// ============================================================================

// ColSummariser abstracts the notion of a function which summarises the
// contents of a given column.
type ColSummariser struct {
	name    string
	summary func(trace.Column) string
}

var colSummarisers []ColSummariser = []ColSummariser{
	{"count", rowSummariser},
	{"width", widthSummariser},
	{"bytes", bytesSummariser},
	{"unique", uniqueSummariser},
}

func rowSummariser(col trace.Column) string {
	return fmt.Sprintf("%d rows", col.Data().Len())
}

func widthSummariser(col trace.Column) string {
	return fmt.Sprintf("%d bits", col.Data().BitWidth())
}

func bytesSummariser(col trace.Column) string {
	bitwidth := col.Data().BitWidth()
	byteWidth := bitwidth / 8
	// Determine proper bytewidth
	if bitwidth%8 != 0 {
		byteWidth++
	}

	return fmt.Sprintf("%d bytes", col.Data().Len()*byteWidth)
}

func uniqueSummariser(col trace.Column) string {
	data := col.Data()
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
	summary func(trace.Trace) string
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
		summary: func(tr trace.Trace) string {
			count := 0
			for i := uint(0); i < tr.Columns().Len(); i++ {
				ithWidth := tr.Columns().Get(i).Data().BitWidth()
				if ithWidth >= lowWidth && ithWidth <= highWidth {
					count++
				}
			}
			return fmt.Sprintf("%d", count)
		},
	}
}
