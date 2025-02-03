package cmd

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
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
		binfile := readSchema(true, false, false, args[1:])
		//
		stats.Log("Reading constraints file")
		// Parse trace file
		columns := readTraceFile(args[0])
		//
		stats.Log("Reading trace file")
		//
		builder := sc.NewTraceBuilder(&binfile.Schema).Expand(true).Parallel(true)
		//
		trace, errors := builder.Build(columns)
		//
		if len(errors) == 0 {
			// Run the inspector.
			errors = inspect(&binfile.Schema, trace)
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
	return inspector.Start()
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

// DEFAULT_MODE sets the default command bar, and allows the user to navigate
// the trace.
const DEFAULT_MODE = 0

// NUMERIC_INPUT_MODE is where the user is entering a numberic value (e.g. to
// specify the row for a goto command).
const NUMERIC_INPUT_MODE = 1

// TEXT_INPUT_MODE is where the user is entering a text value (e.g. for a column
// filter).
const TEXT_INPUT_MODE = 2

// STATUS_MODE means the commandbar is notifying the user with a message for a
// short period of time.
const STATUS_MODE = 3

// Inspector provides the necessary pacjkage
type Inspector struct {
	width  uint
	height uint
	//
	term   *termio.Terminal
	schema sc.Schema
	trace  tr.Trace
	// Modules
	views []moduleView
	// Widgets
	tabs      *widget.Tabs
	table     *widget.Table
	cmdBar    *widget.TextLine
	statusBar *widget.TextLine
	// The stack of "modes" in which the inspector is operating.  The root modes
	// is the first in the stack.  When this is terminated, then the inspector
	// closes.
	modes []InspectorMode
}

// InspectorMode identifies a mode in which the inspector is operating.  The
// default mode is for navigating the trace, but other modes are available for
// receiving input from the user or displaying error messages, etc.
type InspectorMode interface {
	// Activate is called when this mode becomes active.  This happens when the
	// mode is first entered, but can also happen subsequently when a child mode
	// exits and results in this mode being reactivated.
	Activate(*Inspector)
	// Clock is called on every clock tick.  This gives the mode an opportunity
	// to do something if it wishes to.
	Clock(*Inspector)
	// KeyPressed in the inspector and received by this mode.
	KeyPressed(*Inspector, uint16) bool
}

// NewInspector constructs a new inspector on given terminal.
func NewInspector(term *termio.Terminal, schema sc.Schema, trace tr.Trace) *Inspector {
	tabs, table, cmdbar, statusbar := initInspectorWidgets(term, schema)
	nmods := schema.Modules().Count()
	views := make([]moduleView, nmods)
	// initialise module views
	for i := uint(0); i < trace.Width(); i++ {
		mid := trace.Column(i).Context().Module()
		views[mid].trColumns = append(views[mid].trColumns, i)
	}
	// Finalise the module view.
	for i := range views {
		views[i].finalise(trace)
	}
	//
	inspector := &Inspector{0, 0, term, schema, trace, views, tabs, table, cmdbar, statusbar, nil}
	table.SetSource(inspector)
	// Put the inspector into default mode.
	inspector.EnterMode(&NavigationMode{})
	//
	return inspector
}

// Clock the inspector
func (p *Inspector) Clock() error {
	mode := len(p.modes) - 1
	nWidth, nHeight := p.term.GetSize()
	// Pass on clock
	p.modes[mode].Clock(p)
	// Only force rerender if dimensions have changed.
	if nWidth != p.width || nHeight != p.height {
		// Update cached dimensions
		p.width, p.height = nWidth, nHeight
		// Render
		return p.Render()
	}
	//
	return nil
}

// Render the inspector to the given terminal
func (p *Inspector) Render() error {
	return p.term.Render()
}

// Close the inspector.
func (p *Inspector) Close() error {
	return p.term.Restore()
}

// EnterMode pushes a new mode onto the mode stack.
func (p *Inspector) EnterMode(mode InspectorMode) {
	// Append mode to stack of active modes
	p.modes = append(p.modes, mode)
	// Activate mode
	mode.Activate(p)
}

// KeyPressed allows the inspector to react to a key being pressed by the user.
func (p *Inspector) KeyPressed(key uint16) bool {
	var n = len(p.modes) - 1
	//
	if p.modes[n].KeyPressed(p, key) {
		p.modes = p.modes[0:n]
		//
		if n > 0 {
			// Reactivate mode
			p.modes[n-1].Activate(p)
		}
	}
	// Exit when the mode stack is empty.
	return len(p.modes) == 0
}

// Access currently selected view
func (p *Inspector) currentView() *moduleView {
	module := p.tabs.Selected()
	// Action change
	return &p.views[module]
}

// Actions goto row mode
func (p *Inspector) gotoRow(row uint) bool {
	module := p.tabs.Selected()
	// Action change
	return p.views[module].setTrRowOffset(row)
}

// filter columns based on a regex
func (p *Inspector) filterColumns(regex *regexp.Regexp) bool {
	module := p.tabs.Selected()
	p.views[module].applyColumnFilter(p.trace, regex, true)
	// Success
	return true
}

func (p *Inspector) clearColumnFilter() bool {
	module := p.tabs.Selected()
	regex := regexp.MustCompile("")
	p.views[module].applyColumnFilter(p.trace, regex, false)
	// Success
	return true
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
	trRow := col + view.trRowOffset
	width := uint(0)
	//
	if col == 0 {
		width = min(colWidths[0], 48)
	} else if trRow < uint(len(colWidths)) {
		width = min(colWidths[trRow], maxWidth)
	}
	//
	return width
}

// CellAt returns the contents of a given cell in the main table of the
// inspector.
func (p *Inspector) CellAt(col, row uint) termio.FormattedText {
	// Determine currently selected module
	module := p.tabs.Selected()
	view := &p.views[module]
	// Calculate trace offsets
	trCol := min(row-1+view.trColOffset, uint(len(view.trFilteredColumns)))
	trRow := min(col-1+view.trRowOffset, uint(len(view.tabColumnWidths)))
	//
	if col == 0 && row == 0 {
		return termio.NewText(" ")
	} else if row == 0 {
		val := fmt.Sprintf("%d", trRow)
		return termio.NewColouredText(val, termio.TERM_BLUE)
	} else if trCol >= uint(len(view.trFilteredColumns)) {
		// Overrun columns
		return termio.NewText("")
	} else if col == 0 {
		cid := view.trFilteredColumns[trCol]
		name := p.schema.Columns().Nth(cid).Name
		//
		return termio.NewColouredText(name, termio.TERM_BLUE)
	}
	// Determine trace column
	trColumn := view.trFilteredColumns[trCol]
	// Extract cell value
	val := p.trace.Column(trColumn).Get(int(trRow))
	// Convert value into appropriate form.  For now, this is always
	// hexadecimal.
	hex := fmt.Sprintf("0x%s", val.Text(16))
	runes := []rune(hex)
	//
	if len(runes) > int(view.maxTabColWidth) {
		runes := runes[0:view.maxTabColWidth]
		runes[view.maxTabColWidth-1] = '.'
		runes[view.maxTabColWidth-2] = '.'
	}
	//
	text := termio.NewText(string(runes))
	//
	text.Format(cellColour(val))
	//
	return text
}

// Start provides a read / update / render loop.
func (p *Inspector) Start() []error {
	var errors []error
	// Start clock timer
	clk := time.NewTicker(500 * time.Millisecond)
	//
	go func() {
		for {
			// Receive clock signal
			<-clk.C
			// Force render
			//nolint:errcheck
			p.Clock()
		}
	}()
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

// This algorithm is based on that used in the original tool.  To understand
// this algorithm, you need to look at the 256 colour table for ANSI escape
// codes.  It actually does make sense, even if it doesn't appear to.
func cellColour(val fr.Element) termio.AnsiEscape {
	if val.IsZero() {
		return termio.NewAnsiEscape().FgColour(termio.TERM_WHITE)
	}
	// Compute a simple hash of the bytes making up the value in question.
	col := uint(0)
	for _, b := range val.Bytes() {
		col = col ^ uint(b)
	}
	// Select suitable background colour based on hash, whilst also ensuring
	// contrast with the foreground colour.
	bg_col := (col % (213 - 16))
	escape := termio.NewAnsiEscape().Bg256Colour(16 + bg_col)
	//
	if bg_col%36 > 18 {
		escape = escape.FgColour(termio.TERM_BLACK)
	}
	//
	return escape
}

// ==================================================================
// Navigation Mode
// ==================================================================

// NavigationMode is the default mode of the inspector.  In this mode, the user
// is navigating the trace in the normal fashion.
type NavigationMode struct {
}

// Activate navigation mode by setting the command bar to show the navigation
// commands.
func (p *NavigationMode) Activate(parent *Inspector) {
	parent.cmdBar.Clear()
	parent.cmdBar.Add(termio.NewColouredText("[g]", termio.TERM_YELLOW))
	parent.cmdBar.Add(termio.NewText("oto :: "))
	parent.cmdBar.Add(termio.NewColouredText("[f]", termio.TERM_YELLOW))
	parent.cmdBar.Add(termio.NewText("ilter :: "))
	parent.cmdBar.Add(termio.NewColouredText("[#]", termio.TERM_YELLOW))
	parent.cmdBar.Add(termio.NewText("clear filter :: "))
	//p.cmdbar.Add(termio.NewFormattedText("[p]erspectives"))
	parent.cmdBar.Add(termio.NewColouredText("[q]", termio.TERM_RED))
	parent.cmdBar.Add(termio.NewText("uit"))
	//
	parent.statusBar.Clear()
}

// Clock navitation mode, which does nothing at this time.
func (p *NavigationMode) Clock(parent *Inspector) {

}

// KeyPressed in navigation mode, which either adjusts our view of the trace
// table or fires off some command.
func (p *NavigationMode) KeyPressed(parent *Inspector, key uint16) bool {
	module := parent.tabs.Selected()
	//
	switch key {
	case termio.TAB:
		parent.tabs.Select(module + 1)
	case termio.BACKTAB:
		parent.tabs.Select(module - 1)
	case termio.CURSOR_UP:
		col := parent.views[module].trColOffset
		parent.views[module].setTrColumnOffset(col - 1)
	case termio.CURSOR_DOWN:
		col := parent.views[module].trColOffset
		parent.views[module].setTrColumnOffset(col + 1)
	case termio.CURSOR_LEFT:
		row := parent.views[module].trRowOffset
		parent.views[module].setTrRowOffset(row - 1)
	case termio.CURSOR_RIGHT:
		row := parent.views[module].trRowOffset
		parent.views[module].setTrRowOffset(row + 1)
	// quit
	case 'q':
		return true
	// goto command
	case 'g':
		parent.EnterMode(p.gotoInputMode(parent))
	case 'f':
		parent.EnterMode(p.filterInputMode(parent))
	case '#':
		parent.clearColumnFilter()
	}
	//
	return false
}

func (p *NavigationMode) gotoInputMode(parent *Inspector) InspectorMode {
	prompt := termio.NewColouredText("[history ↑/↓] row? ", termio.TERM_YELLOW)
	history := parent.currentView().targetRowHistory
	history_index := uint(len(history))
	//
	return newInputMode(prompt, history_index, history, newUintHandler(parent.gotoRow))
}

func (p *NavigationMode) filterInputMode(parent *Inspector) InspectorMode { //
	prompt := termio.NewColouredText("[history ↑/↓] regex? ", termio.TERM_YELLOW)
	// Determine current active filter
	filter := parent.currentView().columnFilter
	history := parent.currentView().columnFilterHistory
	history_index := uint(len(history))
	//
	if filter != "" {
		history_index--
	}
	//
	return newInputMode(prompt, history_index, history, newRegexHandler(parent.filterColumns))
}

// ==================================================================
// Input Mode
// ==================================================================

// InputMode is where the user is entering some information (e.g. row for
// executing a goto command).
type InputMode[T any] struct {
	// prompt to show user
	prompt termio.FormattedText
	// input text being accumulated whilst in input mode.
	input []byte
	// history of options for this input mode.
	history []string
	// history index
	index uint
	// parser responsible for checking whether input is valid (or not).
	handler InputHandler[T]
}

// InputHandler provides a generic way of handling input, including a mechanism
// for checking that input is well formed.
type InputHandler[T any] interface {
	// Convert attempts to convert the input string into a valid value.
	Convert(string) (T, bool)
	// Apply the given input, which will activate some kind of callback.
	Apply(T)
}

func newInputMode[T any](prompt termio.FormattedText, index uint, history []string,
	handler InputHandler[T]) *InputMode[T] {
	var input []byte
	// Determine whether to show item from history
	if index >= uint(len(history)) {
		input = []byte{}
	} else {
		input = []byte(history[index])
	}
	// Done
	return &InputMode[T]{prompt, input, history, index, handler}
}

// Activate navigation mode by setting the command bar to show the navigation
// commands.
func (p *InputMode[T]) Activate(parent *Inspector) {
	parent.cmdBar.Clear()
	parent.cmdBar.Add(p.prompt)
	// Add current filter
	colour := termio.TERM_GREEN
	input := string(p.input)
	//
	if _, ok := p.handler.Convert(input); !ok {
		colour = termio.TERM_RED
	}
	//
	parent.cmdBar.Add(termio.NewColouredText(input, colour))
}

// Clock navitation mode, which does nothing at this time.
func (p *InputMode[T]) Clock(parent *Inspector) {
	// Nothing to do.
}

// KeyPressed in input mode simply updates the input, or exits the mode if
// either "ESC" or enter are pressed.
func (p *InputMode[T]) KeyPressed(parent *Inspector, key uint16) bool {
	switch {
	case key == termio.ESC:
		return true
	case key == termio.BACKSPACE || key == termio.DEL:
		if len(p.input) > 0 {
			n := len(p.input) - 1
			p.input = p.input[0:n]
		}
	case key == termio.CARRIAGE_RETURN:
		input := string(p.input)
		// Attempt conversion
		if val, ok := p.handler.Convert(input); ok {
			// Looks good, to fire the value
			p.handler.Apply(val)
		}
		// Success
		return true
	case key == termio.CURSOR_UP:
		if p.index > 0 {
			p.index--
			p.input = []byte(p.history[p.index])
		}
	case key == termio.CURSOR_DOWN:
		p.index++
		// Check for end-of-history
		if p.index >= uint(len(p.history)) {
			p.index = uint(len(p.history))
			p.input = []byte{}
		} else {
			p.input = []byte(p.history[p.index])
		}
	case key >= 32 && key <= 126:
		p.input = append(p.input, byte(key))
	}
	// Update displayed text
	p.Activate(parent)
	//
	return false
}

type uintHandler struct {
	callback func(uint) bool
}

func newUintHandler(callback func(uint) bool) InputHandler[uint] {
	return &uintHandler{callback}
}

func (p *uintHandler) Convert(input string) (uint, bool) {
	val, err := strconv.Atoi(input)
	//
	if val < 0 || err != nil {
		return 0, false
	}
	//
	return uint(val), true
}

func (p *uintHandler) Apply(value uint) {
	p.callback(value)
}

type regexHandler struct {
	callback func(*regexp.Regexp) bool
}

func newRegexHandler(callback func(*regexp.Regexp) bool) InputHandler[*regexp.Regexp] {
	return &regexHandler{callback}
}

func (p *regexHandler) Convert(input string) (*regexp.Regexp, bool) {
	if regex, err := regexp.Compile(input); err == nil {
		return regex, true
	}

	return nil, false
}

func (p *regexHandler) Apply(regex *regexp.Regexp) {
	p.callback(regex)
}

// ==================================================================
// Helpers
// ==================================================================

func initInspectorWidgets(term *termio.Terminal, schema sc.Schema) (tabs *widget.Tabs,
	table *widget.Table, cmdbar *widget.TextLine, statusbar *widget.TextLine) {
	//
	tabs = initInspectorTabs(schema)
	table = widget.NewTable(nil)
	cmdbar = widget.NewText()
	statusbar = widget.NewText()
	//
	term.Add(tabs)
	term.Add(widget.NewSeparator("⎯"))
	term.Add(table)
	term.Add(widget.NewSeparator("⎯"))
	term.Add(cmdbar)
	term.Add(statusbar)
	//
	return tabs, table, cmdbar, statusbar
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
	// Identifies trace columns in this module.
	trColumns []uint
	// Identifies trace filtered in this module.
	trFilteredColumns []uint
	// Row offset into trace
	trRowOffset uint
	// Column offset into trace
	trColOffset uint
	// History for goto row commands
	targetRowHistory []string
	// Active column filter
	columnFilter string
	// Set of column filters used.
	columnFilterHistory []string
}

func (p *moduleView) setTrColumnOffset(colOffset uint) {
	// Only set when it makes sense
	if colOffset < uint(len(p.trFilteredColumns)) {
		p.trColOffset = colOffset
	}
}

func (p *moduleView) setTrRowOffset(rowOffset uint) bool {
	// Only set when it makes sense
	if rowOffset < uint(len(p.tabColumnWidths)) {
		rowOffsetStr := fmt.Sprintf("%d", rowOffset)
		p.trRowOffset = rowOffset
		p.targetRowHistory = history_append(p.targetRowHistory, rowOffsetStr)
		//
		return true
	}
	// failed
	return false
}

// Finalise the module view, for example by computing all the column widths.
func (p *moduleView) finalise(tr tr.Trace) {
	// First table column always for trace column headers.
	nTableCols := uint(0)
	// Determine height of columns in this module, keeping in mind that some
	// columns might have multipliers in play.
	for _, col := range p.trColumns {
		column := tr.Column(col)
		nTableCols = max(nTableCols, column.Data().Len())

		p.trFilteredColumns = append(p.trFilteredColumns, col)
	}
	//
	p.tabColumnWidths = make([]uint, nTableCols+1)
	//
	for _, col := range p.trFilteredColumns {
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

// Apply a new column filter to this module view.
func (p *moduleView) applyColumnFilter(tr tr.Trace, regex *regexp.Regexp, history bool) {
	// Reset filter
	p.trFilteredColumns = nil
	// Apply filter
	for _, col := range p.trColumns {
		if name := tr.Column(col).Name(); regex.MatchString(name) {
			p.trFilteredColumns = append(p.trFilteredColumns, col)
		}
	}
	// Update selection and history
	p.columnFilter = regex.String()
	//
	if history {
		p.columnFilterHistory = history_append(p.columnFilterHistory, regex.String())
	}
}

// History append will append a given item to the end of the history.  However,
// if that item already existed in the history, then that is removed.  This is
// to avoid duplicates in the history.
func history_append[T comparable](history []T, item T) []T {
	// Remove previous entry (if applicable)
	history = util.RemoveMatching(history, func(ith T) bool { return ith == item })
	// Add item to end
	return append(history, item)
}

//nolint:errcheck
func init() {
	rootCmd.AddCommand(inspectCmd)
}
