package trace

import (
	"fmt"
	"unicode/utf8"
)

// PrintTrace prints a trace in a more human-friendly fashion.
func PrintTrace(tr Trace) {
	n := tr.Width()
	//
	rows := make([][]string, n)
	for i := uint(0); i < n; i++ {
		rows[i] = traceColumnData(tr, i)
	}
	//
	widths := traceRowWidths(tr.Height(), rows)
	//
	printHorizontalRule(widths)
	//
	for _, r := range rows {
		printTraceRow(r, widths)
		printHorizontalRule(widths)
	}
}

func traceColumnData(tr Trace, col uint) []string {
	n := tr.Height()
	data := make([]string, n+2)
	data[0] = fmt.Sprintf("#%d", col)
	data[1] = tr.ColumnByIndex(col).Name()

	for row := 0; row < int(n); row++ {
		data[row+2] = tr.ColumnByIndex(col).Get(row).String()
	}

	return data
}

func traceRowWidths(height uint, rows [][]string) []int {
	widths := make([]int, height+2)

	for _, row := range rows {
		for i, col := range row {
			w := utf8.RuneCountInString(col)
			widths[i] = max(w, widths[i])
		}
	}

	return widths
}

func printTraceRow(row []string, widths []int) {
	for i, col := range row {
		fmt.Printf(" %*s |", widths[i], col)
	}

	fmt.Println()
}

func printHorizontalRule(widths []int) {
	for _, w := range widths {
		fmt.Print("-")

		for i := 0; i < w; i++ {
			fmt.Print("-")
		}
		fmt.Print("-+")
	}

	fmt.Println()
}
