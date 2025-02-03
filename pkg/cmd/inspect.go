package cmd

import (
	"fmt"
	"os"

	"github.com/consensys/go-corset/pkg/binfile"
	"github.com/consensys/go-corset/pkg/cmd/inspector"
	"github.com/consensys/go-corset/pkg/corset/compiler"
	sc "github.com/consensys/go-corset/pkg/schema"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/termio"
	"github.com/spf13/cobra"
)

var inspectCmd = &cobra.Command{
	Use:   "inspect [flags] trace_file constraint_file(s)",
	Short: "Inspect a trace file",
	Long:  `Inspect a trace file using an interactive (terminal-based) environment`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 2 {
			fmt.Println(cmd.UsageString())
			os.Exit(1)
		}
		defensive := GetFlag(cmd, "defensive")
		//
		stats := util.NewPerfStats()
		// Parse constraints
		binf := ReadConstraintFiles(true, false, false, args[1:])
		// Sanity check debug information is available.
		if _, ok := binfile.GetAttribute[*compiler.SourceMap](binf); !ok {
			panic("missing source map information from binary constraints file")
		}
		//
		stats.Log("Reading constraints file")
		// Parse trace file
		columns := ReadTraceFile(args[0])
		//
		stats.Log("Reading trace file")
		//
		builder := sc.NewTraceBuilder(&binf.Schema).Expand(true).Defensive(defensive).Parallel(true)
		//
		trace, errors := builder.Build(columns)
		//
		if len(errors) == 0 {
			// Run the inspector.
			errors = inspect(binf, trace)
		}
		// Sanity check what happened
		if len(errors) > 0 {
			for _, err := range errors {
				fmt.Println(err)
			}
			os.Exit(1)
		}
	},
}

// Inspect a given trace using a given schema.
func inspect(binf *binfile.BinaryFile, trace tr.Trace) []error {
	// Construct inspector window
	inspector, err := construct(binf, trace)
	// Check error
	if err != nil {
		return []error{err}
	}
	// Render inspector
	if err := inspector.Render(); err != nil {
		return []error{err}
	}
	//
	return inspector.Start()
}

func construct(binf *binfile.BinaryFile, trace tr.Trace) (*inspector.Inspector, error) {
	term, err := termio.NewTerminal()
	// Check whether successful
	if err != nil {
		fmt.Println(error.Error(err))
		os.Exit(1)
	}
	// Construct inspector state
	return inspector.NewInspector(term, binf, trace)
}

//nolint:errcheck
func init() {
	rootCmd.AddCommand(inspectCmd)
	inspectCmd.Flags().Bool("defensive", true, "enable / disable defensive padding")
}
