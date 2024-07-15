package cmd

import (
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
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
		//
		if output != "" {
			writeTraceFile(output, tr)
		}

		if print {
			printTrace(start, end, max_width, tr)
		}
	},
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
			data := make([]*fr.Element, ith.Height())
			//
			for j := 0; j < int(ith.Height()); j++ {
				data[j] = ith.Get(j)
			}

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

func listColumns(tr trace.Trace) {
	n := tr.Columns().Len()
	tbl := util.NewTablePrinter(3, n)

	for i := uint(0); i < n; i++ {
		ith := tr.Columns().Get(i)
		elems := fmt.Sprintf("%d rows", ith.Height())
		bytes := fmt.Sprintf("%d bytes", ith.Width()*ith.Height())
		tbl.SetRow(i, QualifiedColumnName(i, tr), elems, bytes)
	}

	//
	tbl.SetMaxWidth(64)
	tbl.Print()
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

func init() {
	rootCmd.AddCommand(traceCmd)
	traceCmd.Flags().BoolP("list", "l", false, "list only the columns in the trace file")
	traceCmd.Flags().BoolP("print", "p", false, "print entire trace file")
	traceCmd.Flags().Uint("pad", 0, "add a given number of padding rows (to each module)")
	traceCmd.Flags().UintP("start", "s", 0, "filter out rows below this")
	traceCmd.Flags().UintP("end", "e", math.MaxUint, "filter out this and all following rows")
	traceCmd.Flags().Uint("max-width", 32, "specify maximum display width for a column")
	traceCmd.Flags().StringP("out", "o", "", "Specify output file to write trace")
	traceCmd.Flags().StringP("filter", "f", "", "Filter columns beginning with prefix")
}
