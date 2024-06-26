package util

import "fmt"

// TablePrinter is useful for printing tables to the terminal.
type TablePrinter struct {
	widths []uint
	rows   [][]string
}

// NewTablePrinter constructs a new table with given dimensions.
func NewTablePrinter(width uint, height uint) *TablePrinter {
	widths := make([]uint, width)
	rows := make([][]string, height)
	// Construct the table
	for i := uint(0); i < height; i++ {
		rows[i] = make([]string, width)
	}

	return &TablePrinter{widths, rows}
}

// Set the contents of a given cell in this table
func (p *TablePrinter) Set(col uint, row uint, val string) {
	p.widths[col] = max(p.widths[col], uint(len(val)))
	p.rows[row][col] = val
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

// SetMaxWidth puts an upper bound on the width of any column.
func (p *TablePrinter) SetMaxWidth(m uint) {
	for i := 0; i < len(p.widths); i++ {
		p.widths[i] = min(p.widths[i], m)
	}
}

// Print the table.
func (p *TablePrinter) Print() {
	for i := 0; i < len(p.rows); i++ {
		row := p.rows[i]
		for j, col := range row {
			jth := col
			jth_width := p.widths[j]

			if uint(len(col)) > jth_width {
				jth = col[0:jth_width]
			}

			fmt.Printf(" %*s |", jth_width, jth)
		}

		fmt.Println()
	}
}
