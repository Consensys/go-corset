package cmd

import (
	"fmt"
	"os"
	"path"

	"github.com/consensys/go-corset/pkg/binfile"
	"github.com/consensys/go-corset/pkg/hir"
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
			schema, err = hir.ParseSchemaSExp(string(bytes))
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
	fmt.Println(err)
	os.Exit(2)
	// unreachable
	return nil
}
