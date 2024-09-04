package cmd

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/consensys/go-corset/pkg/binfile"
	"github.com/consensys/go-corset/pkg/hir"
	"github.com/consensys/go-corset/pkg/sexp"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/trace/json"
	"github.com/consensys/go-corset/pkg/trace/lt"
	log "github.com/sirupsen/logrus"
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

// Get an expected string, or panic if an error arises.
func getString(cmd *cobra.Command, flag string) string {
	r, err := cmd.Flags().GetString(flag)
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
			err = fmt.Errorf("Unknown schema file format: %s\n", ext)
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

func maxHeightColumns(cols []trace.RawColumn) uint {
	h := uint(0)
	// Iterate over modules
	for _, col := range cols {
		h = max(h, col.Data.Len())
	}
	// Done
	return h
}

// PerfStats provides a snapshot of memory allocation at a given point in time.
type PerfStats struct {
	// Starting time
	startTime time.Time
	// Starting total memory allocation
	startMem uint64
	// Starting number of gc events
	startGc uint32
}

// NewPerfStats creates a new snapshot of the current amount of memory allocated.
func NewPerfStats() *PerfStats {
	var m runtime.MemStats

	startTime := time.Now()

	runtime.ReadMemStats(&m)

	return &PerfStats{startTime, m.TotalAlloc, m.NumGC}
}

// Log logs the difference between the state now and as it was when the PerfStats object was created.
func (p *PerfStats) Log(prefix string) {
	var m runtime.MemStats

	runtime.ReadMemStats(&m)
	alloc := (m.TotalAlloc - p.startMem) / 1024 / 1024 / 1024
	gcs := m.NumGC - p.startGc
	exectime := time.Since(p.startTime).Seconds()

	log.Debugf("%s took %0.2fs using %v Gb (%v GC events)", prefix, exectime, alloc, gcs)
}
