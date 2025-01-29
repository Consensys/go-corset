package cmd

import (
	"fmt"
	"os"

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
		trace, errors := builder.Build(columns)
		//
		if len(errors) == 0 {
			// Run the inspector.
			errors = inspect(schema, trace)
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
func inspect(schema sc.Schema, trace tr.Trace) []error {
	// Construct inspector window
	inspector := construct(schema, trace)
	// Render inspector
	if err := inspector.Render(); err != nil {
		return []error{err}
	}
	//
	return inspector.Loop()
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
	// Widgets
	tabs  *widget.Tabs
	table *widget.Table
}

// NewInspector constructs a new inspector on given terminal.
func NewInspector(term *termio.Terminal, schema sc.Schema, trace tr.Trace) *Inspector {
	tabs, table := initInspectorWidgets(term, schema)
	nmods := schema.Modules().Count()
	views := make([]moduleView, nmods)
	// initialise module views
	for i := uint(0); i < trace.Width(); i++ {
		mid := trace.Column(i).Context().Module()
		views[mid].trColumnIds = append(views[mid].trColumnIds, i)
	}
	// Finalise the module view.
	for i := range views {
		views[i].finalise(trace)
	}
	//
	inspector := &Inspector{term, schema, trace, views, tabs, table}
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

// KeyPressed allows the inspector to react to a key being pressed by the user.
func (p *Inspector) KeyPressed(key uint16) bool {
	module := p.tabs.Selected()
	//
	switch key {
	case termio.TAB:
		p.tabs.Select(module + 1)
	case termio.BACKTAB:
		p.tabs.Select(module - 1)
	case termio.CURSOR_UP:
		col := p.views[module].trColOffset
		p.views[module].setTrColumnOffset(col - 1)
	case termio.CURSOR_DOWN:
		col := p.views[module].trColOffset
		p.views[module].setTrColumnOffset(col + 1)
	case termio.CURSOR_LEFT:
		row := p.views[module].trRowOffset
		p.views[module].setTrRowOffset(row - 1)
	case termio.CURSOR_RIGHT:
		row := p.views[module].trRowOffset
		p.views[module].setTrRowOffset(row + 1)
	case 'q':
		return true
	}
	//
	return false
}

// ==================================================================
// TableSource
// ==================================================================

// ColumnWidth gets the width of a given column in the main table of the
// inspector.  Note that columns here are table columns, not trace columns.
func (p *Inspector) ColumnWidth(col uint) uint {
	module := p.tabs.Selected()
	view := p.views[module]
	colWidths := view.tabColumnWidths
	maxWidth := view.maxTabColWidth
	//
	trRow := min(col-1+view.trRowOffset, uint(len(view.tabColumnWidths)))
	width := maxWidth
	//
	if col == 0 {
		width = colWidths[col] + 1
	} else if trRow < uint(len(colWidths)) {
		width = colWidths[trRow] + 1
	}
	// Default
	return min(width, maxWidth) + 1
}

// CellAt returns the contents of a given cell in the main table of the
// inspector.
func (p *Inspector) CellAt(col, row uint) string {
	// Determine currently selected module
	module := p.tabs.Selected()
	view := &p.views[module]
	// Calculate trace offsets
	trCol := min(row-1+view.trColOffset, uint(len(view.trColumnIds)))
	trRow := min(col-1+view.trRowOffset, uint(len(view.tabColumnWidths)))
	//
	if col == 0 && row == 0 {
		return " "
	} else if row == 0 {
		return fmt.Sprintf("%d", trRow)
	} else if trCol >= uint(len(view.trColumnIds)) {
		// Overrun columns
		return ""
	} else if col == 0 {
		cid := view.trColumnIds[trCol]
		// Determine column name
		return p.schema.Columns().Nth(cid).Name
	}
	// Determine trace column
	trColumn := view.trColumnIds[trCol]
	// Extract cell value
	val := p.trace.Column(trColumn).Get(int(trRow - 1))
	//
	return fmt.Sprintf("0x%s", val.Text(16))
}

// Loop provides a read / update / render loop.
func (p *Inspector) Loop() []error {
	var errors []error
	//
	for {
		if key, err := p.term.ReadKey(); err != nil {
			errors = append(errors, err)
			break
		} else if exit := p.KeyPressed(key); exit {
			break
		}
		// Rerender window
		if err := p.Render(); err != nil {
			errors = append(errors, err)
			break
		}
	}
	// Attempt to restore terminal state
	if err := p.term.Restore(); err != nil {
		errors = append(errors, err)
	}
	// Done
	return errors
}

// ==================================================================
// Helpers
// ==================================================================

func initInspectorWidgets(term *termio.Terminal, schema sc.Schema) (tabs *widget.Tabs, table *widget.Table) {
	tabs = initInspectorTabs(schema)
	table = widget.NewTable(nil)
	//
	term.Add(tabs)
	term.Add(widget.NewSeparator("⎯"))
	term.Add(table)
	//
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
	// Identifies table column tabColumnWidths. Notice that these columns are table
	// columns, not trace columns!
	tabColumnWidths []uint
	// Current maximum width for a table column
	maxTabColWidth uint
	// Identifies trace trColumnIds in this module.
	trColumnIds []uint
	// Row offset into trace
	trRowOffset uint
	// Column offset into trace
	trColOffset uint
}

func (p *moduleView) setTrColumnOffset(colOffset uint) {
	// Only set when it makes sense
	if colOffset < uint(len(p.trColumnIds)) {
		p.trColOffset = colOffset
	}
}

func (p *moduleView) setTrRowOffset(rowOffset uint) {
	// Only set when it makes sense
	if rowOffset < uint(len(p.tabColumnWidths)) {
		p.trRowOffset = rowOffset
	}
}

// Finalise the module view, for example by computing all the column widths.
func (p *moduleView) finalise(tr tr.Trace) {
	// First table column always for trace column headers.
	nTableCols := uint(0)
	// Determine height of columns in this module, keeping in mind that some
	// columns might have multipliers in play.
	for _, col := range p.trColumnIds {
		column := tr.Column(col)
		nTableCols = max(nTableCols, column.Data().Len())
	}
	//
	p.tabColumnWidths = make([]uint, nTableCols+1)
	//
	for _, col := range p.trColumnIds {
		column := tr.Column(col)
		length := len(column.Name())
		data := column.Data()
		p.tabColumnWidths[0] = max(p.tabColumnWidths[0], uint(length))
		//
		for i := uint(0); i < data.Len(); i++ {
			val := data.Get(i)
			str := fmt.Sprintf("0x%s", val.Text(16))
			width := uint(len(str))
			p.tabColumnWidths[i+1] = max(p.tabColumnWidths[i+1], width)
		}
	}
	// Final configuration stuff
	p.maxTabColWidth = 16
}

//nolint:errcheck
func init() {
	rootCmd.AddCommand(inspectCmd)
}
