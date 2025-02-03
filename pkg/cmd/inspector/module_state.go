package inspector

import (
	"fmt"
	"regexp"
	"slices"

	"github.com/consensys/go-corset/pkg/corset/compiler"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
)

// ModuleState provides state regarding how to display the trace for a given
// module, including related aspects like filter histories, etc.
type ModuleState struct {
	// Identifies trace columns in this module.
	columns []compiler.SourceColumnMapping
	// Active module view
	view ModuleView
	// History for goto row commands
	targetRowHistory []string
	// Active column filter
	columnFilter string
	// Set of column filters used.
	columnFilterHistory []string
}

func (p *ModuleState) setColumnOffset(colOffset uint) {
	p.view.SetColumn(colOffset)
}

func (p *ModuleState) setRowOffset(rowOffset uint) bool {
	if p.view.SetRow(rowOffset) {
		// Update history
		rowOffsetStr := fmt.Sprintf("%d", rowOffset)
		p.targetRowHistory = history_append(p.targetRowHistory, rowOffsetStr)
		//
		return true
	}
	// failed
	return false
}

// Finalise the module view, for example by computing all the column widths.
func (p *ModuleState) finalise(trace tr.Trace) {
	// Sort all column names so that, for example, columns in the same
	// perspective are grouped together.
	slices.SortFunc(p.columns, func(l compiler.SourceColumnMapping, r compiler.SourceColumnMapping) int {
		return l.Column.Name.Compare(r.Column.Name)
	})
	// Final configuration stuff
	p.view.maxRowWidth = 16
	// Initialise the view
	p.view.SetActiveColumns(trace, p.columns)
}

// Apply a new column filter to the module view.  This determines which columns
// are currently visible.
func (p *ModuleState) applyColumnFilter(trace tr.Trace, regex *regexp.Regexp, history bool) {
	filteredColumns := make([]compiler.SourceColumnMapping, 0)
	// Apply filter
	for _, col := range p.columns {
		// Check whether it matches the regex or not.
		if name := col.Column.Name.String(); regex.MatchString(name) {
			filteredColumns = append(filteredColumns, col)
		}
	}
	// Update the view
	p.view.SetActiveColumns(trace, filteredColumns)
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
