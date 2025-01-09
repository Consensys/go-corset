package cmd

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/consensys/go-corset/pkg/binfile"
	"github.com/consensys/go-corset/pkg/corset"
	"github.com/consensys/go-corset/pkg/hir"
	"github.com/consensys/go-corset/pkg/sexp"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/trace/json"
	"github.com/consensys/go-corset/pkg/trace/lt"
	"github.com/spf13/cobra"
)

// GetFlag gets an expected flag, or panic if an error arises.
func GetFlag(cmd *cobra.Command, flag string) bool {
	r, err := cmd.Flags().GetBool(flag)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}

	return r
}

// GetInt gets an expectedsigned integer, or panic if an error arises.
func GetInt(cmd *cobra.Command, flag string) int {
	r, err := cmd.Flags().GetInt(flag)
	if err != nil {
		fmt.Println(err)
		os.Exit(3)
	}

	return r
}

// GetUint gets an expected unsigned integer, or panic if an error arises.
func GetUint(cmd *cobra.Command, flag string) uint {
	r, err := cmd.Flags().GetUint(flag)
	if err != nil {
		fmt.Println(err)
		os.Exit(4)
	}

	return r
}

// GetString gets an expected string, or panic if an error arises.
func GetString(cmd *cobra.Command, flag string) string {
	r, err := cmd.Flags().GetString(flag)
	if err != nil {
		fmt.Println(err)
		os.Exit(4)
	}

	return r
}

// GetStringArray gets an expected string array, or panic if an error arises.
func GetStringArray(cmd *cobra.Command, flag string) []string {
	r, err := cmd.Flags().GetStringArray(flag)
	if err != nil {
		fmt.Println(err)
		os.Exit(4)
	}

	return r
}

// Write a given trace file to disk
func writeTraceFile(filename string, columns []trace.RawColumn) {
	var err error

	var bytes []byte
	// Check file extension
	ext := path.Ext(filename)
	//
	switch ext {
	case ".json":
		js := json.ToJsonString(columns)
		//
		if err = os.WriteFile(filename, []byte(js), 0644); err == nil {
			return
		}
	case ".lt":
		bytes, err = lt.ToBytes(columns)
		//
		if err == nil {
			if err = os.WriteFile(filename, bytes, 0644); err == nil {
				return
			}
		}
	default:
		err = fmt.Errorf("Unknown trace file format: %s", ext)
	}
	// Handle error
	fmt.Println(err)
	os.Exit(4)
}

// Parse a trace file using a parser based on the extension of the filename.
func readTraceFile(filename string) []trace.RawColumn {
	var tr []trace.RawColumn
	// Read data file
	bytes, err := os.ReadFile(filename)
	// Check success
	if err == nil {
		// Check file extension
		ext := path.Ext(filename)
		//
		switch ext {
		case ".json":
			tr, err = json.FromBytes(bytes)
			if err == nil {
				return tr
			}
		case ".lt":
			tr, err = lt.FromBytes(bytes)
			if err == nil {
				return tr
			}
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

// Read the constraints file, whilst optionally including the standard library.
func readSchema(stdlib bool, debug bool, legacy bool, filenames []string) *hir.Schema {
	if len(filenames) == 0 {
		fmt.Println("source or binary constraint(s) file required.")
		os.Exit(5)
	} else if len(filenames) == 1 && path.Ext(filenames[0]) == ".bin" {
		// Single (binary) file supplied
		return readBinaryFile(legacy, filenames[0])
	}
	// Must be source files
	return readSourceFiles(stdlib, debug, filenames)
}

// Read a "bin" file.
func readBinaryFile(legacy bool, filename string) *hir.Schema {
	var schema *hir.Schema
	// Read schema file
	data, err := os.ReadFile(filename)
	// Handle errors
	if err == nil && legacy {
		// Read the binary file
		schema, err = binfile.HirSchemaFromJson(data)
	} else if err == nil {
		// Read the Gob file
		buffer := bytes.NewBuffer(data)
		decoder := gob.NewDecoder(buffer)
		err = decoder.Decode(&schema)
	}
	// Return if no errors
	if err == nil {
		return schema
	}
	// Handle error & exit
	fmt.Println(err)
	os.Exit(2)
	// unreachable
	return nil
}

// Parse a set of source files and compile them into a single schema.  This can
// result, for example, in a syntax error, etc.
func readSourceFiles(stdlib bool, debug bool, filenames []string) *hir.Schema {
	srcfiles := make([]*sexp.SourceFile, len(filenames))
	// Read each file
	for i, n := range filenames {
		// Read source file
		bytes, err := os.ReadFile(n)
		// Sanity check for errors
		if err != nil {
			fmt.Println(err)
			os.Exit(3)
		}
		//
		srcfiles[i] = sexp.NewSourceFile(n, bytes)
	}
	// Parse and compile source files
	schema, errs := corset.CompileSourceFiles(stdlib, debug, srcfiles)
	// Check for any errors
	if len(errs) == 0 {
		return schema
	}
	// Report errors
	for _, err := range errs {
		printSyntaxError(&err)
	}
	// Fail
	os.Exit(4)
	// unreachable
	return nil
}

func writeHirSchema(schema *hir.Schema, legacy bool, filename string) {
	var (
		buffer     bytes.Buffer
		gobEncoder *gob.Encoder = gob.NewEncoder(&buffer)
	)
	// Sanity checks
	if legacy {
		// Currently, there is no support for this.
		fmt.Println("legacy binary format not supported for writing")
	}
	// Encode schema
	if err := gobEncoder.Encode(schema); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	// Write file
	if err := os.WriteFile(filename, buffer.Bytes(), 0644); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

// Print a syntax error with appropriate highlighting.
func printSyntaxError(err *sexp.SyntaxError) {
	span := err.Span()
	line := err.FirstEnclosingLine()
	lineOffset := span.Start() - line.Start()
	// Calculate length (ensures don't overflow line)
	length := min(line.Length()-lineOffset, span.Length())
	// Print error + line number
	fmt.Printf("%s:%d:%d-%d %s\n", err.SourceFile().Filename(),
		line.Number(), 1+lineOffset, 1+lineOffset+length, err.Message())
	// Print separator line
	fmt.Println()
	// Print line
	fmt.Println(line.String())
	// Print indent (todo: account for tabs)
	fmt.Print(strings.Repeat(" ", lineOffset))
	// Print highlight
	fmt.Println(strings.Repeat("^", length))
}

func maxHeightColumns(cols []trace.RawColumn) uint {
	h := uint(0)
	// Iterate over modules
	for _, col := range cols {
		h = max(h, col.Data.Len())
	}
	// Done
	return h
}
