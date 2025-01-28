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
	// Construct inspector state
	inspector := newInspector(schema, trace)
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
	schema sc.Schema
	trace  tr.Trace
	// Modules
	views []moduleView
	// Selected module
	module uint
}

func newInspector(schema sc.Schema, trace tr.Trace) *inspector {
	nmods := schema.Modules().Count()
	views := make([]moduleView, nmods)
	// initialise module views
	for i := uint(0); i < trace.Width(); i++ {
		mid := trace.Column(i).Context().Module()
		views[mid].columns = append(views[mid].columns, i)
	}
	//
	return &inspector{schema, trace, views, 0}
}

func (p *inspector) ColumnWidth(col uint) uint {
	//return p.views[p.module].widths[col]
	return 10
}

func (p *inspector) CellAt(col, row uint) string {
	view := &p.views[p.module]
	if row >= uint(len(view.columns)) {
		return "???"
	} else if col == 0 {
		cid := view.columns[row]
		// Determine column name
		return p.schema.Columns().Nth(cid).Name
	}
	//
	return "x"
}

func (p *inspector) TableDimensions() (uint, uint) {
	nrows := p.trace.Height(tr.NewContext(p.module, 1))
	ncols := uint(len(p.views[p.module].columns))

	return nrows, ncols
}

type moduleView struct {
	widths  []uint
	columns []uint
}

//nolint:errcheck
func init() {
	rootCmd.AddCommand(inspectCmd)
}
