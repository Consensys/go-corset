package cmd

import (
	"fmt"
	"os"

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
		//
		fmt.Printf("Tracefile has %d columns\n", trace.Width())
	},
}

func init() {
	rootCmd.AddCommand(traceCmd)
}
