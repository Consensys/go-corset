package util

import (
	"fmt"
)

// TablePrinter is useful for printing tables to the terminal.
type TablePrinter struct {
	widths        []uint
	rows          [][]string
	escapes       [][]string
	enableEscapes bool
}

// NewTablePrinter constructs a new table with given dimensions.
func NewTablePrinter(width uint, height uint) *TablePrinter {
	widths := make([]uint, width)
	rows := make([][]string, height)
	escapes := make([][]string, height)
	// Construct the table
	for i := uint(0); i < height; i++ {
		rows[i] = make([]string, width)
		escapes[i] = make([]string, width)
	}

	return &TablePrinter{widths, rows, escapes, true}
}

// Set the contents of a given cell in this table
func (p *TablePrinter) Set(col uint, row uint, val string) {
	p.widths[col] = max(p.widths[col], uint(len(val)))
	p.rows[row][col] = val
}

// SetEscape set the colour to use when printing the contents of a given cell
func (p *TablePrinter) SetEscape(col uint, row uint, escape string) {
	p.escapes[row][col] = escape
}

// AnsiEscapes enables or disables the use of ANSI escapes (e.g. for showing
// colour).  Disabling escapes is useful in environments that don't support
// escapes as, otherwise, you get a lot of visible excape characters being
// printed.
func (p *TablePrinter) AnsiEscapes(enable bool) {
	p.enableEscapes = enable
}

// SetRow sets the contents of an entire row in this table
func (p *TablePrinter) SetRow(row uint, vals ...string) {
	if len(vals) != len(p.widths) {
		panic("incorrect number of columns")
	}
	// Update column widths
	for i := 0; i < len(p.widths); i++ {
		p.widths[i] = max(p.widths[i], uint(len(vals[i])))
	}
	// Done
	p.rows[row] = vals
}

// SetMaxWidths puts an upper bound on the width of any column.
func (p *TablePrinter) SetMaxWidths(width uint) {
	for i := uint(0); i < uint(len(p.widths)); i++ {
		p.SetMaxWidth(i, width)
	}
}

// SetMaxWidth puts an upper bound on the width of any column.
func (p *TablePrinter) SetMaxWidth(col uint, width uint) {
	p.widths[col] = min(p.widths[col], width)
}

// Print the table.
func (p *TablePrinter) Print() {
	//
	for i := 0; i < len(p.rows); i++ {
		row := p.rows[i]
		escapes := p.escapes[i]
		//
		for j, col := range row {
			jth := col
			jth_width := p.widths[j]
			jth_escape := escapes[j]
			// Print colour (if applicable)
			if p.enableEscapes && jth_escape != "" {
				fmt.Print(jth_escape)
			}
			// Print data
			if uint(len(col)) > jth_width {
				jth = col[0 : jth_width-2]
				fmt.Printf(" %*s..", jth_width-2, jth)
			} else {
				fmt.Printf(" %*s", jth_width, jth)
			}
			// Cancel colour (if applicable)
			if p.enableEscapes && jth_escape != "" {
				fmt.Print(ResetAnsiEscape().Build())
			}

			fmt.Print(" |")
		}

		fmt.Println()
	}
}
