package widget

import (
	"math"

	"github.com/consensys/go-corset/pkg/util/termio"
)

// TableSource is an abstraction used by a table to determine what values to put in each cell.
type TableSource interface {
	// Width returns the width of a given column.
	ColumnWidth(col uint) uint
	// Get content of given cell in table.
	CellAt(col, row uint) termio.FormattedText
}

// Table is a grid of cells of varying width.
type Table struct {
	source TableSource
}

// NewTable constructs a new table with a given source.
func NewTable(source TableSource) *Table {
	return &Table{source}
}

// GetHeight of this widget, where MaxUint indicates widget expands to take as
// much as it can.
func (p *Table) GetHeight() uint {
	return math.MaxUint
}

// SetSource sets the table source.
func (p *Table) SetSource(source TableSource) {
	p.source = source
}

// Render this widget on the given canvas.
func (p *Table) Render(canvas termio.Canvas) {
	// Determine canvas dimensions
	width, height := canvas.GetDimensions()
	//
	xpos := uint(0)
	//
	for col := uint(0); xpos < width; col++ {
		colWidth := p.source.ColumnWidth(col)
		//
		for row := uint(0); row < height; row++ {
			cell := p.source.CellAt(col, row)
			cell.Clip(0, colWidth)
			canvas.Write(xpos, row, cell)
		}
		//
		xpos += colWidth + 1
	}
}
