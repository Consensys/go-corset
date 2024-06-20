package cmd

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/consensys/go-corset/pkg/binfile"
	"github.com/consensys/go-corset/pkg/hir"
	"github.com/consensys/go-corset/pkg/sexp"
	"github.com/consensys/go-corset/pkg/trace"
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

// Get an expectedsigned integer, or panic if an error arises.
func getInt(cmd *cobra.Command, flag string) int {
	r, err := cmd.Flags().GetInt(flag)
	if err != nil {
		fmt.Println(err)
		os.Exit(3)
	}

	return r
}

// Get an expected unsigned integer, or panic if an error arises.
func getUint(cmd *cobra.Command, flag string) uint {
	r, err := cmd.Flags().GetUint(flag)
	if err != nil {
		fmt.Println(err)
		os.Exit(4)
	}

	return r
}

// Parse a trace file using a parser based on the extension of the filename.
func readTraceFile(filename string) *trace.ArrayTrace {
	var tr *trace.ArrayTrace
	// Read data file
	bytes, err := os.ReadFile(filename)
	// Check success
	if err == nil {
		// Check file extension
		ext := path.Ext(filename)
		//
		switch ext {
		case ".json":
			tr, err = trace.ParseJsonTrace(bytes)
			if err == nil {
				return tr
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
	if e, ok := err.(*sexp.SyntaxError); ok {
		printSyntaxError(filename, e, string(bytes))
	} else {
		fmt.Println(err)
	}

	os.Exit(2)
	// unreachable
	return nil
}

// Print a syntax error with appropriate highlighting.
func printSyntaxError(filename string, err *sexp.SyntaxError, text string) {
	span := err.Span()
	// Construct empty source map in order to determine enclosing line.
	srcmap := sexp.NewSourceMap[sexp.SExp]([]rune(text))
	//
	line := srcmap.FindFirstEnclosingLine(span)
	// Print error + line number
	fmt.Printf("%s:%d: %s\n", filename, line.Number(), err.Message())
	// Print separator line
	fmt.Println()
	// Print line
	fmt.Println(line.String())
	// Print indent (todo: account for tabs)
	lineOffset := span.Start() - line.Start()
	fmt.Print(strings.Repeat(" ", lineOffset))
	// Calculate length (ensures don't overflow line)
	length := min(line.Length()-lineOffset, span.Length())
	// Print highlight
	fmt.Println(strings.Repeat("^", length))
}
