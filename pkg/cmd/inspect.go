package cmd

import (
	"fmt"
	"os"
	"time"

	sc "github.com/consensys/go-corset/pkg/schema"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/termio"
	"github.com/consensys/go-corset/pkg/util/termio/widget"
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
		//
		stats := util.NewPerfStats()
		// Parse constraints
		schema := readSchema(true, false, false, args[1:])
		//
		stats.Log("Reading constraints file")
		// Parse trace file
		columns := readTraceFile(args[0])
		//
		stats.Log("Reading trace file")
		//
		builder := sc.NewTraceBuilder(schema).Expand(true).Parallel(true)
		//
		trace, errs := builder.Build(columns)
		//
		if len(errs) > 0 {
			fmt.Println(errs)
		}
		//
		inspect(schema, trace)
	},
}

// Inspect a given trace using a given schema.
func inspect(schema sc.Schema, trace tr.Trace) {
	// Construct inspector window
	term := construct(schema, trace)
	// Render inspector
	if err := term.Render(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	//
	time.Sleep(5 * time.Second)
	//
	term.Restore()
}

func construct(schema sc.Schema, trace tr.Trace) *termio.Terminal {
	term, err := termio.NewTerminal()
	// Check whether successful
	if err != nil {
		fmt.Println(error.Error(err))
		os.Exit(1)
	}
	inspector := &inspector{schema, trace}
	//
	term.Add(constructTabs(schema))
	term.Add(widget.NewSeparator("⎯"))
	term.Add(widget.NewTable(inspector))
	//
	return term
}

func constructTabs(schema sc.Schema) termio.Widget {
	var titles []string
	for i := schema.Modules(); i.HasNext(); {
		titles = append(titles, i.Next().Name)
	}
	//
	return widget.NewTabs(titles...)
}

// ==================================================================
// Inspector
// ==================================================================

type inspector struct {
	trace tr.Trace
	// Modules
	views []moduleView
	// Selected module
	module uint
}

func newInspector(schema sc.Schema, trace tr.Trace) *inspector {
	nmods := schema.Modules().Count()
	views := make([]moduleView, nmods)
	// initialise views
	for iter := schema.Columns(); iter.HasNext(); {
		ith := iter.Next()
		mid := ith.Context.Module()
		views[mid].columns = append(views[mid].columns, i))
	}
	//
	return &inspector{trace, views, 0}
}

func (p *inspector) ColumnWidth(col uint) uint {
	return 1
}

func (p *inspector) CellAt(col, row uint) string {
	return "x"
}

func (p *inspector) TableDimensions() (uint, uint) {
	// Find columns in the table
	cols, ok := p.schema.Columns().Find(func(c sc.Column) bool {
		return c.Context.Module() == p.module
	})
	//
	return 5, uint(len(cols))
}

type moduleView struct {
	widths  []uint
	columns []uint
}

//nolint:errcheck
func init() {
	rootCmd.AddCommand(inspectCmd)
}
