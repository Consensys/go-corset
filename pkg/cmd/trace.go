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
		trace := readTraceFile(args[0])
		list := getFlag(cmd, "list")
		print := getFlag(cmd, "print")
		padding := getUint(cmd, "pad")
		start := getUint(cmd, "start")
		end := getUint(cmd, "end")
		max_width := getUint(cmd, "max-width")
		filter := getString(cmd, "filter")
		output := getString(cmd, "out")
		//
		if filter != "" {
			trace = filterColumns(trace, filter)
		}
		if padding != 0 {
			trace.Pad(padding)
		}
		if list {
			listColumns(trace)
		}
		//
		if output != "" {
			writeTraceFile(output, trace)
		}

		if print {
			printTrace(start, end, max_width, trace)
		}
	},
}

// Construct a new trace containing only those columns from the original who
// name begins with the given prefix.
func filterColumns(tr trace.Trace, prefix string) trace.Trace {
	ntr := trace.EmptyArrayTrace()
	//
	for i := uint(0); i < tr.Width(); i++ {
		ith := tr.ColumnByIndex(i)
		if strings.HasPrefix(ith.Name(), prefix) {
			ntr.Add(ith)
		}
	}
	// Done
	return ntr
}

func listColumns(tr trace.Trace) {
	tbl := util.NewTablePrinter(3, tr.Width())

	for i := uint(0); i < tr.Width(); i++ {
		ith := tr.ColumnByIndex(i)
		elems := fmt.Sprintf("%d rows", ith.Height())
		bytes := fmt.Sprintf("%d bytes", ith.Width()*ith.Height())
		tbl.SetRow(i, ith.Name(), elems, bytes)
	}

	//
	tbl.SetMaxWidth(64)
	tbl.Print()
}

func printTrace(start uint, end uint, max_width uint, tr trace.Trace) {
	height := min(tr.Height(), end) - start
	tbl := util.NewTablePrinter(1+height, 1+tr.Width())

	for j := uint(0); j < height; j++ {
		tbl.Set(j+1, 0, fmt.Sprintf("#%d", j+start))
	}

	for i := uint(0); i < tr.Width(); i++ {
		ith := tr.ColumnByIndex(i)
		tbl.Set(0, i+1, ith.Name())

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
