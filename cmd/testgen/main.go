package main

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/hir"
	sc "github.com/consensys/go-corset/pkg/schema"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/trace/json"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func main() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "testgen",
	Short: "Test generation utility for go-corset.",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Println(cmd.UsageString())
			os.Exit(1)
		}
		model := args[0]
		// Lookup model
		for _, m := range models {
			if m.Name == model {
				// Read schema
				filename := fmt.Sprintf("%s.lisp", m.Name)
				schema := readSchemaFile(path.Join("testdata", filename))
				// Generate & split traces
				valid, invalid := generateTestTraces(m, schema)
				// Write out
				writeTestTraces(m, "accepts", schema, valid)
				writeTestTraces(m, "rejects", schema, invalid)
				os.Exit(0)
			}
		}
		//
		fmt.Printf("unknown model \"%s\"\n", model)
		os.Exit(1)
	},
}

// Model represents a hard-coded oracle for a given test.
type Model struct {
	// Name of the model in question
	Name string
	// Predicate for determining which trace to accept
	Oracle func(sc.Schema, tr.Trace) bool
}

var models []Model = []Model{
	{"memory", memoryModel},
}

// Generate test traces
func generateTestTraces(model Model, schema sc.Schema) ([]tr.Trace, []tr.Trace) {
	// NOTE: This is really a temporary solution for now.  It doesn't handle
	// length multipliers.  It doesn't allow for modules with different heights.
	// It uses a fixed pool.
	pool := []fr.Element{fr.NewElement(0), fr.NewElement(1), fr.NewElement(2)}
	//
	enumerator := sc.NewTraceEnumerator(2, schema, pool)
	valid := make([]tr.Trace, 0)
	invalid := make([]tr.Trace, 0)
	// Generate and split the traces
	for enumerator.HasNext() {
		trace := enumerator.Next()
		// Check whether trace is valid or not (according to the oracle)
		if model.Oracle(schema, trace) {
			valid = append(valid, trace)
		} else {
			invalid = append(invalid, trace)
		}
	}
	// Done
	return valid, invalid
}

func writeTestTraces(model Model, ext string, schema sc.Schema, traces []tr.Trace) {
	var sb strings.Builder
	// Construct filename
	filename := fmt.Sprintf("testdata/%s.auto.%s", model.Name, ext)
	// Generate lines
	for _, trace := range traces {
		raw := traceToColumns(schema, trace)
		json := json.ToJsonString(raw)
		sb.WriteString(json)
		sb.WriteString("\n")
	}
	// Write the file
	if err := os.WriteFile(filename, []byte(sb.String()), 0644); err != nil {
		panic(err)
	}
	// Log what happened
	log.Infof("Wrote %s\n", filename)
}

// Convert a trace into an array of raw columns.
func traceToColumns(schema sc.Schema, trace tr.Trace) []tr.RawColumn {
	ncols := schema.InputColumns().Count()
	cols := make([]tr.RawColumn, ncols)
	i := 0
	// Convert each column
	for iter := schema.InputColumns(); iter.HasNext(); {
		sc_col := iter.Next()
		// Lookup the column data
		tr_col := findColumn(sc_col.Context().Module(), sc_col.Name(), schema, trace)
		// Determine module name
		mod := schema.Modules().Nth(sc_col.Context().Module())
		// Assignt the raw colmn
		cols[i] = tr.RawColumn{Module: mod.Name(), Name: sc_col.Name(), Data: tr_col.Data()}
		//
		i++
	}
	//
	return cols
}

func readSchemaFile(filename string) *hir.Schema {
	// Read schema file
	bytes, err := os.ReadFile(filename)
	// Handle errors
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// Attempt to parse schema
	schema, err2 := hir.ParseSchemaString(string(bytes))
	// Check whether parsed successfully or not
	if err2 == nil {
		// Ok
		return schema
	}
	// Errors
	fmt.Println(err2)
	os.Exit(1)
	// unreachable
	return nil
}

func findColumn(mod uint, col string, schema sc.Schema, trace tr.Trace) tr.Column {
	cid, ok := sc.ColumnIndexOf(schema, mod, col)
	if !ok {
		panic(fmt.Sprintf("unknown column \"%s\"", col))
	}
	// Done
	return trace.Column(cid)
}

// ============================================================================
// Models
// ============================================================================

func memoryModel(schema sc.Schema, trace tr.Trace) bool {
	TWO_1 := fr.NewElement(2)
	TWO_8 := fr.NewElement(256)
	TWO_16 := fr.NewElement(65536)
	TWO_32 := fr.NewElement(4294967296)
	//
	PC := findColumn(0, "PC", schema, trace).Data()
	RW := findColumn(0, "RW", schema, trace).Data()
	ADDR := findColumn(0, "ADDR", schema, trace).Data()
	VAL := findColumn(0, "VAL", schema, trace).Data()
	// Configure memory model
	memory := make(map[fr.Element]fr.Element, 0)
	//
	for i := uint(0); i < PC.Len(); i++ {
		pc_i := PC.Get(i)
		rw_i := RW.Get(i)
		addr_i := ADDR.Get(i)
		val_i := VAL.Get(i)
		// Type constraints
		t_pc := pc_i.Cmp(&TWO_16) < 0
		t_rw := rw_i.Cmp(&TWO_1) < 0
		t_addr := addr_i.Cmp(&TWO_32) < 0
		t_val := val_i.Cmp(&TWO_8) < 0
		// Check type constraints
		if !(t_pc && t_rw && t_addr && t_val) {
			return false
		}
		// Heartbeat 1
		h1 := i != 0 || pc_i.IsZero()
		// Heartbeat 2
		h2 := i == 0 || pc_i.IsZero() || isIncremented(PC.Get(i-1), pc_i)
		// Heartbeat 3
		h3 := i == 0 || !pc_i.IsZero() || PC.Get(i-1) == pc_i
		// Heartbeat 4
		h4 := !pc_i.IsZero() || (rw_i.IsZero() && addr_i.IsZero() && val_i.IsZero())
		// Check heartbeat constraints
		if !(h1 && h2 && h3 && h4) {
			return false
		}
		// Check reading / writing
		if rw_i.IsOne() {
			// Write
			memory[addr_i] = val_i
		} else {
			v := memory[addr_i]
			// Check read matches
			if v.Cmp(&val_i) != 0 {
				return false
			}
		}
	}
	// Success
	return true
}

// ============================================================================
// Helpers
// ============================================================================

// Check a given element is the previous element plus one.
func isIncremented(before fr.Element, after fr.Element) bool {
	after.Sub(&after, &before)
	//
	return after.IsOne()
}
