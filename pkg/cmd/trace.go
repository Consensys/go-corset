package cmd

import (
	"fmt"
	"os"

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
		output := getString(cmd, "out")
		//
		if list {
			listColumns(trace)
		}
		//
		if output != "" {
			writeTraceFile(output, trace)
		}
	},
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

func init() {
	rootCmd.AddCommand(traceCmd)
	traceCmd.Flags().BoolP("list", "l", false, "detail the columns in the trace file")
	traceCmd.Flags().StringP("out", "o", "", "Specify output file to write trace")
}
