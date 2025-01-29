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
			os.Exit(1)
		}
		//
		inspect(schema, trace)
	},
}

// Inspect a given trace using a given schema.
func inspect(schema sc.Schema, trace tr.Trace) {
	// Construct inspector window
	inspector := construct(schema, trace)
	// Render inspector
	if err := inspector.Render(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	//
	time.Sleep(5 * time.Second)
	//
	inspector.Close()
}

func construct(schema sc.Schema, trace tr.Trace) *Inspector {
	term, err := termio.NewTerminal()
	// Check whether successful
	if err != nil {
		fmt.Println(error.Error(err))
		os.Exit(1)
	}
	// Construct inspector state
	return NewInspector(term, schema, trace)
}

// ==================================================================
// Inspector
// ==================================================================

// Inspector provides the necessary pacjkage
type Inspector struct {
	term   *termio.Terminal
	schema sc.Schema
	trace  tr.Trace
	// Modules
	views []moduleView
	//
	tabs  *widget.Tabs
	table *widget.Table
	// Selected module
	module uint
}

// NewInspector constructs a new inspector on given terminal.
func NewInspector(term *termio.Terminal, schema sc.Schema, trace tr.Trace) *Inspector {
	tabs, table := initInspectorWidgets(term, schema)
	nmods := schema.Modules().Count()
	views := make([]moduleView, nmods)
	// initialise module views
	for i := uint(0); i < trace.Width(); i++ {
		mid := trace.Column(i).Context().Module()
		views[mid].columns = append(views[mid].columns, i)
	}
	//
	inspector := &Inspector{term, schema, trace, views, tabs, table, 0}
	table.SetSource(inspector)
	//
	return inspector
}

// Render the inspector to the given terminal
func (p *Inspector) Render() error {
	return p.term.Render()
}

// Close the inspector.
func (p *Inspector) Close() error {
	return p.term.Restore()
}

// ColumnWidth gets the width of a given column in the main table of the
// inspector.  Note that columns here are table columns, not trace columns.
func (p *Inspector) ColumnWidth(col uint) uint {
	//return p.views[p.module].widths[col]
	return 10
}

// CellAt returns the contents of a given cell in the main table of the
// inspector.
func (p *Inspector) CellAt(col, row uint) string {
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

// TableDimensions returns the maxium dimensions of the main table of the
// inspector.
func (p *Inspector) TableDimensions() (uint, uint) {
	nrows := p.trace.Height(tr.NewContext(p.module, 1))
	ncols := uint(len(p.views[p.module].columns))

	return nrows, ncols
}

func initInspectorWidgets(term *termio.Terminal, schema sc.Schema) (tabs *widget.Tabs, table *widget.Table) {
	tabs = initInspectorTabs(schema)
	table = widget.NewTable(nil)
	//
	term.Add(tabs)
	term.Add(widget.NewSeparator("⎯"))
	term.Add(table)
	return tabs, table
}

func initInspectorTabs(schema sc.Schema) *widget.Tabs {
	var titles []string
	for i := schema.Modules(); i.HasNext(); {
		titles = append(titles, i.Next().Name)
	}
	//
	return widget.NewTabs(titles...)
}

type moduleView struct {
	widths  []uint
	columns []uint
}

//nolint:errcheck
func init() {
	rootCmd.AddCommand(inspectCmd)
}
