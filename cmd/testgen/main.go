package main

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	util "github.com/consensys/go-corset/pkg/cmd"
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
	rootCmd.Flags().Uint("min-elem", 0, "Minimum element")
	rootCmd.Flags().Uint("max-elem", 2, "Maximum element")
	rootCmd.Flags().Uint("min-lines", 1, "Minimum number of lines")
	rootCmd.Flags().Uint("max-lines", 4, "Maximum number of lines")
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
		var cfg TestGenConfig
		// Lookup model
		cfg.model = findModel(args[0])
		cfg.min_elem = util.GetUint(cmd, "min-elem")
		cfg.max_elem = util.GetUint(cmd, "max-elem")
		cfg.min_lines = util.GetUint(cmd, "min-lines")
		cfg.max_lines = util.GetUint(cmd, "max-lines")
		// Read schema
		filename := fmt.Sprintf("%s.lisp", cfg.model.Name)
		schema := readSchemaFile(path.Join("testdata", filename))
		// Generate & split traces
		valid, invalid := generateTestTraces(cfg, schema)
		// Write out
		writeTestTraces(cfg.model, "accepts", schema, valid)
		writeTestTraces(cfg.model, "rejects", schema, invalid)
		os.Exit(0)

	},
}

// TestGenConfig encapsulates configuration related to test generation.
type TestGenConfig struct {
	model     Model
	min_elem  uint
	max_elem  uint
	min_lines uint
	max_lines uint
}

// OracleFn defines function which determines whether or not a given trace is accepted by the model (or not).
type OracleFn = func(sc.Schema, tr.Trace) bool

// Model represents a hard-coded oracle for a given test.
type Model struct {
	// Name of the model in question
	Name string
	// Predicate for determining which trace to accept
	Oracle OracleFn
}

var models []Model = []Model{
	{"bit_decomposition", bitDecompositionModel},
	{"byte_decomposition", byteDecompositionModel},
	{"memory", memoryModel},
	{"word_sorting", wordSortingModel},
	{"counter", functionalModel("STAMP", counterModel)},
}

func findModel(name string) Model {
	for _, m := range models {
		if m.Name == name {
			return m
		}
	}
	//
	panic(fmt.Sprintf("unknown model \"%s\"", name))
}

// Generate test traces
func generateTestTraces(cfg TestGenConfig, schema sc.Schema) ([]tr.Trace, []tr.Trace) {
	// NOTE: This is really a temporary solution for now.  It doesn't handle
	// length multipliers.  It doesn't allow for modules with different heights.
	// It uses a fixed pool.
	pool := generatePool(cfg)
	valid := make([]tr.Trace, 0)
	invalid := make([]tr.Trace, 0)
	//
	for n := cfg.min_lines; n < cfg.max_lines; n++ {
		enumerator := sc.NewTraceEnumerator(n, schema, pool)
		// Generate and split the traces
		for enumerator.HasNext() {
			trace := enumerator.Next()
			// Check whether trace is valid or not (according to the oracle)
			if cfg.model.Oracle(schema, trace) {
				valid = append(valid, trace)
			} else {
				invalid = append(invalid, trace)
			}
		}
	}
	// Done
	return valid, invalid
}

func generatePool(cfg TestGenConfig) []fr.Element {
	n := cfg.max_elem - cfg.min_elem + 1
	elems := make([]fr.Element, n)
	// Iterate values
	for i := uint(0); i != n; i++ {
		val := uint64(cfg.min_elem + i)
		elems[i] = fr.NewElement(val)
	}
	// Done
	return elems
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
	log.Infof("Wrote %s (%d traces)\n", filename, len(traces))
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

func functionalModel(stamp string, model func(uint, uint, sc.Schema, tr.Trace) bool) OracleFn {
	return func(schema sc.Schema, trace tr.Trace) bool {
		// Lookup stamp column
		STAMP := findColumn(0, stamp, schema, trace).Data()
		// Check STAMP initially zero
		if STAMP.Len() > 0 {
			STAMP_0 := STAMP.Get(0)
			if !STAMP_0.IsZero() {
				return false
			}
		}
		// Set initial frame
		start := uint(0)
		current := fr.NewElement(0)
		i := uint(1)
		// Split frames
		for ; i < STAMP.Len(); i++ {
			stamp_i := STAMP.Get(i)
			// Look for frame boundary
			if stamp_i.Cmp(&current) != 0 {
				// Check stamp incremented
				if !isIncremented(current, stamp_i) {
					return false
				}
				// Check whether valid frame (or padding)
				if !current.IsZero() && !model(start, i-1, schema, trace) {
					return false
				}
				// Reset for next frame
				start = i
				current = stamp_i
			}
		}
		// Handle final frame
		if !current.IsZero() && !model(start, i-1, schema, trace) {
			return false
		}
		//
		return true
	}
}

// ============================================================================
// Models
// ============================================================================
func bitDecompositionModel(schema sc.Schema, trace tr.Trace) bool {
	TWO_1 := fr.NewElement(2)
	TWO_2 := fr.NewElement(4)
	TWO_3 := fr.NewElement(8)
	TWO_4 := fr.NewElement(16)
	//
	NIBBLE := findColumn(0, "NIBBLE", schema, trace).Data()
	BIT_0 := findColumn(0, "BIT_0", schema, trace).Data()
	BIT_1 := findColumn(0, "BIT_1", schema, trace).Data()
	BIT_2 := findColumn(0, "BIT_2", schema, trace).Data()
	BIT_3 := findColumn(0, "BIT_3", schema, trace).Data()
	//
	for i := uint(0); i < NIBBLE.Len(); i++ {
		NIBBLE_i := NIBBLE.Get(i)
		BIT_0_i := BIT_0.Get(i)
		BIT_1_i := BIT_1.Get(i)
		BIT_2_i := BIT_2.Get(i)
		BIT_3_i := BIT_3.Get(i)
		// Check types
		t_NIBBLE := NIBBLE_i.Cmp(&TWO_4) < 0
		t_BIT_0 := BIT_0_i.Cmp(&TWO_1) < 0
		t_BIT_1 := BIT_1_i.Cmp(&TWO_1) < 0
		t_BIT_2 := BIT_2_i.Cmp(&TWO_1) < 0
		t_BIT_3 := BIT_3_i.Cmp(&TWO_1) < 0
		// Check type constraints
		if !(t_NIBBLE && t_BIT_0 && t_BIT_1 && t_BIT_2 && t_BIT_3) {
			return false
		}
		//
		b1 := mul(BIT_1_i, TWO_1)
		b2 := mul(BIT_2_i, TWO_2)
		b3 := mul(BIT_3_i, TWO_3)
		sum := add(add(add(b3, b2), b1), BIT_0_i)
		// Check decomposition matches
		if NIBBLE_i.Cmp(&sum) != 0 {
			return false
		}
	}
	// Success
	return true
}

func byteDecompositionModel(schema sc.Schema, trace tr.Trace) bool {
	TWO_8 := fr.NewElement(256)
	//
	ST := findColumn(0, "ST", schema, trace).Data()
	CT := findColumn(0, "CT", schema, trace).Data()
	BYTE := findColumn(0, "BYTE", schema, trace).Data()
	ARG := findColumn(0, "ARG", schema, trace).Data()
	//
	padding := true
	//
	for i := uint(0); i < ST.Len(); i++ {
		st_i := ST.Get(i)
		ct_i := CT.Get(i)
		byte_i := BYTE.Get(i)
		arg_i := ARG.Get(i)
		// Type constraints
		t_byte := byte_i.Cmp(&TWO_8) < 0
		// Check type constraints
		if !t_byte {
			return false
		}
		//
		if padding && st_i.IsZero() {

		} else if padding && !st_i.IsZero() {
			padding = false
		} else if st_i.IsZero() {
			return false
		}
		//
		if i+1 < ST.Len() {
			st_ip1 := ST.Get(i + 1)
			if !(eq(st_i, st_ip1) || eq(add_const(st_i, 1), st_ip1)) {
				return false
			}
		}
		// Check other constraints
		if ct_i.IsZero() && !eq(arg_i, byte_i) {
			return false
		} else if !ct_i.IsZero() && !eq(arg_i, add(byte_i, mul(TWO_8, BYTE.Get(i-1)))) {
			return false
		} else if !padding && i+1 < ST.Len() && !eq(add_const(ct_i, 1), CT.Get(i+1)) {
			return false
		}
	}
	// Success
	return true
}

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

func wordSortingModel(schema sc.Schema, trace tr.Trace) bool {
	TWO_8 := fr.NewElement(256)
	//
	X := findColumn(0, "X", schema, trace).Data()
	Delta := findColumn(0, "Delta", schema, trace).Data()
	Byte_0 := findColumn(0, "Byte_0", schema, trace).Data()
	Byte_1 := findColumn(0, "Byte_1", schema, trace).Data()
	//
	for i := uint(0); i < X.Len(); i++ {
		X_i := X.Get(i)
		Delta_i := Delta.Get(i)
		Byte_0_i := Byte_0.Get(i)
		Byte_1_i := Byte_1.Get(i)
		tmp := add(mul(Byte_1_i, TWO_8), Byte_0_i)
		//
		if Delta_i.Cmp(&tmp) != 0 {
			return false
		} else if i > 0 {
			X_im1 := X.Get(i - 1)
			diff := sub(X_i, X_im1)

			if Delta_i.Cmp(&diff) != 0 {
				return false
			}
		}
	}
	// Success
	return true
}

// ============================================================================
// Functional Models
// ============================================================================

func counterModel(first uint, last uint, schema sc.Schema, trace tr.Trace) bool {
	CT := findColumn(0, "CT", schema, trace).Data()
	// All frames in this model must have length 4
	if last-first != 3 {
		return false
	}
	//
	for i := first; i <= last; i++ {
		ct_i := CT.Get(i)
		expected := fr.NewElement(uint64(i - first))
		// Check counter matches expected valid
		if ct_i.Cmp(&expected) != 0 {
			return false
		}
	}
	//
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

func add(lhs fr.Element, rhs fr.Element) fr.Element {
	lhs.Add(&lhs, &rhs)
	return lhs
}

func add_const(lhs fr.Element, rhs uint64) fr.Element {
	return add(lhs, fr.NewElement(rhs))
}

func sub(lhs fr.Element, rhs fr.Element) fr.Element {
	lhs.Sub(&lhs, &rhs)
	return lhs
}

func mul(lhs fr.Element, rhs fr.Element) fr.Element {
	lhs.Mul(&lhs, &rhs)
	return lhs
}

func eq(lhs fr.Element, rhs fr.Element) bool {
	d := sub(lhs, rhs)
	return d.IsZero()
}
