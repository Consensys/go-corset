package cmd

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/consensys/go-corset/pkg/binfile"
	"github.com/consensys/go-corset/pkg/hir"
	"github.com/consensys/go-corset/pkg/sexp"
	"github.com/consensys/go-corset/pkg/table"
	"github.com/spf13/cobra"
)

// Get an expected flag, or panic if an error arises.
func getFlag(cmd *cobra.Command, flag string) bool {
	r, err := cmd.Flags().GetBool(flag)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}

	return r
}

// Parse a trace file using a parser based on the extension of the filename.
func readTraceFile(filename string) *table.ArrayTrace {
	bytes, err := os.ReadFile(filename)
	if err == nil {
		// Check file extension
		ext := path.Ext(filename)
		//
		switch ext {
		case ".json":
			trace, err := table.ParseJsonTrace(bytes)
			if err == nil {
				return trace
			}
		case ".lt":
			panic("Support for lt trace files not implemented (yet).")
		default:
			err = fmt.Errorf("Unknown trace file format: %s", ext)
		}
	}
	// Handle error
	fmt.Println(err)
	os.Exit(2)
	// unreachable
	return nil
}

// Parse a constraints schema file using a parser based on the extension of the
// filename.
func readSchemaFile(filename string) *hir.Schema {
	var schema *hir.Schema
	// Read schema file
	bytes, err := os.ReadFile(filename)
	// Handle errors
	if err == nil {
		// Check file extension
		ext := path.Ext(filename)
		//
		switch ext {
		case ".lisp":
			// Parse bytes into an S-Expression
			schema, err = hir.ParseSchemaString(string(bytes))
			if err == nil {
				return schema
			}
		case ".bin":
			schema, err = binfile.HirSchemaFromJson(bytes)
			if err == nil {
				return schema
			}
		default:
			err = fmt.Errorf("Unknown trace file format: %s\n", ext)
		}
	}
	// Handle error
	if e, ok := err.(*sexp.ParseError); ok {
		printSyntaxError(filename, e.Message, e.Index, e.Index+1, string(bytes))
	} else {
		fmt.Println(err)
	}

	os.Exit(2)
	// unreachable
	return nil
}

// Print a syntax error with appropriate highlighting.
func printSyntaxError(filename string, msg string, start int, end int, text string) {
	line, offset, num := findEnclosingLine(start, text)
	// Print error + line number
	fmt.Printf("%s:%d: %s\n", filename, num, msg)
	// Print line
	fmt.Println(line)
	// Print indent (todo: account for tabs)
	fmt.Print(strings.Repeat(" ", start-offset-1))
	// Print highlight
	fmt.Println(strings.Repeat("^", end-start))
}

// Determine the enclosing line for the given index in a string.
func findEnclosingLine(index int, text string) (string, int, int) {
	num := 1
	start := 0
	// Handle case where we've reached the end-of-file unexpectedly.  This
	// essentially means the error is reported at the end of the last physical
	// line.
	if index >= len(text) {
		index = index - 1
	}
	// Find the line.
	for i := 0; i < len(text); i++ {
		if i == index {
			end := findEndOfLine(index, text)
			return text[start:end], start, num
		} else if text[i] == '\n' {
			num++
			start = i + 1
		}
	}
	// Should be impossible
	panic("unreachable")
}

// Find the end of the enclosing line
func findEndOfLine(index int, text string) int {
	for i := index; i < len(text); i++ {
		if text[i] == '\n' {
			return i
		}
	}
	// No end in sight!
	return len(text)
}
