// Copyright Consensys Software Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with
// the License. You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	cmdutil "github.com/consensys/go-corset/pkg/cmd"
	"github.com/consensys/go-corset/pkg/corset"
	"github.com/consensys/go-corset/pkg/hir"
	sc "github.com/consensys/go-corset/pkg/schema"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/trace/json"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/source"
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
		cfg.min_elem = cmdutil.GetUint(cmd, "min-elem")
		cfg.max_elem = cmdutil.GetUint(cmd, "max-elem")
		cfg.min_lines = cmdutil.GetUint(cmd, "min-lines")
		cfg.max_lines = cmdutil.GetUint(cmd, "max-lines")
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
	{"byte_decomposition", fixedFunctionModel("ST", "CT", 4, byteDecompositionModel)},
	{"multiplier", multiplierModel},
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
		tr_col := findColumn(sc_col.Context.Module(), sc_col.Name, schema, trace)
		// Determine module name
		mod := schema.Modules().Nth(sc_col.Context.Module())
		// Assignt the raw column
		cols[i] = tr.RawColumn{Module: mod.Name, Name: sc_col.Name, Data: tr_col.Data()}
		//
		i++
	}
	//
	return cols
}

func readSchemaFile(filename string) *hir.Schema {
	var corsetConfig corset.CompilationConfig
	// Read schema file
	bytes, err := os.ReadFile(filename)
	// Handle errors
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// Package up as source file
	srcfile := source.NewSourceFile(filename, bytes)
	// Attempt to parse schema
	binfile, err2 := corset.CompileSourceFile(corsetConfig, srcfile)
	// Check whether parsed successfully or not
	if err2 == nil {
		// Ok
		return &binfile.Schema
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

// Fixed function model is a function model where each frame has a fixed number of rows.
func fixedFunctionModel(stamp string, clk string, n uint, model func(uint, uint, sc.Schema, tr.Trace) bool) OracleFn {
	clockedFn := func(first uint, last uint, schema sc.Schema, trace tr.Trace) bool {
		CLK := findColumn(0, clk, schema, trace).Data()
		// Check frame has expected size
		if last-first+1 != n {
			return false
		}
		// Check the counter
		for i := first; i <= last; i++ {
			clk_i := CLK.Get(i)
			expected := fr.NewElement(uint64(i - first))
			// Check counter matches expected valid
			if clk_i.Cmp(&expected) != 0 {
				return false
			}
		}
		// Chain the model
		return model(first, last, schema, trace)
	}
	// Final step
	return functionalModel(stamp, clockedFn)
}

func checkType(bitwidth uint64, name string, schema sc.Schema, trace tr.Trace) bool {
	// Determine 2^n
	two_n := fr.NewElement(2)
	util.Pow(&two_n, bitwidth)
	// Find column in question
	col := findColumn(0, name, schema, trace).Data()
	//
	for i := uint(0); i < col.Len(); i++ {
		ith := col.Get(i)
		if ith.Cmp(&two_n) >= 0 {
			return false
		}
	}
	//
	return true
}

func checkTypes(bitwidth uint64, schema sc.Schema, trace tr.Trace, names ...string) bool {
	for _, n := range names {
		if !checkType(bitwidth, n, schema, trace) {
			return false
		}
	}
	//
	return true
}

// ============================================================================
// Models
// ============================================================================
func bitDecompositionModel(schema sc.Schema, trace tr.Trace) bool {
	TWO_1 := fr.NewElement(2)
	TWO_2 := fr.NewElement(4)
	TWO_3 := fr.NewElement(8)
	//
	NIBBLE := findColumn(0, "NIBBLE", schema, trace).Data()
	BIT_0 := findColumn(0, "BIT_0", schema, trace).Data()
	BIT_1 := findColumn(0, "BIT_1", schema, trace).Data()
	BIT_2 := findColumn(0, "BIT_2", schema, trace).Data()
	BIT_3 := findColumn(0, "BIT_3", schema, trace).Data()
	// Check column types
	if !checkType(4, "NIBBLE", schema, trace) ||
		!checkType(2, "BIT_0", schema, trace) ||
		!checkType(2, "BIT_1", schema, trace) ||
		!checkType(2, "BIT_2", schema, trace) ||
		!checkType(2, "BIT_3", schema, trace) {
		return false
	}
	// Check decomposition
	for i := uint(0); i < NIBBLE.Len(); i++ {
		NIBBLE_i := NIBBLE.Get(i)
		BIT_0_i := BIT_0.Get(i)
		BIT_1_i := BIT_1.Get(i)
		BIT_2_i := BIT_2.Get(i)
		BIT_3_i := BIT_3.Get(i)
		//
		sum := add(mul(BIT_3_i, TWO_3), mul(BIT_2_i, TWO_2), mul(BIT_1_i, TWO_1), BIT_0_i)
		// Check decomposition matches
		if NIBBLE_i.Cmp(&sum) != 0 {
			return false
		}
	}
	// Success
	return true
}

func byteDecompositionModel(first uint, last uint, schema sc.Schema, trace tr.Trace) bool {
	TWO_8 := fr.NewElement(256)
	BYTE := findColumn(0, "BYTE", schema, trace).Data()
	ARG := findColumn(0, "ARG", schema, trace).Data()
	acc := fr.NewElement(0)
	// Iterate elements
	for i := first; i <= last; i++ {
		byte := BYTE.Get(i)
		arg := ARG.Get(i)
		acc := add(mul(acc, TWO_8), byte)
		// Check accumulator
		if acc.Cmp(&arg) != 0 {
			return false
		}
	}
	//
	return true
}
func multiplierModel(schema sc.Schema, trace tr.Trace) bool {
	TWO_4 := fr.NewElement(16)
	TWO_8 := fr.NewElement(256)
	TWO_12 := fr.NewElement(4096)
	//
	ARG1_0 := findColumn(0, "ARG1_0", schema, trace).Data()
	ARG1_1 := findColumn(0, "ARG1_1", schema, trace).Data()
	ARG2_0 := findColumn(0, "ARG2_0", schema, trace).Data()
	ARG2_1 := findColumn(0, "ARG2_1", schema, trace).Data()
	RES_0 := findColumn(0, "RES_0", schema, trace).Data()
	RES_1 := findColumn(0, "RES_1", schema, trace).Data()
	RES_2 := findColumn(0, "RES_2", schema, trace).Data()
	RES_3 := findColumn(0, "RES_3", schema, trace).Data()
	// Check column types
	if !checkTypes(4, schema, trace, "ARG1_0", "ARG1_1", "ARG2_0", "ARG2_1") ||
		!checkTypes(4, schema, trace, "RES_0", "RES_1", "RES_2", "RES_3") {
		return false
	}
	// Check decomposition
	for i := uint(0); i < RES_0.Len(); i++ {
		ARG1_0_i := ARG1_0.Get(i)
		ARG1_1_i := ARG1_1.Get(i)
		ARG2_0_i := ARG2_0.Get(i)
		ARG2_1_i := ARG2_1.Get(i)
		RES_0_i := RES_0.Get(i)
		RES_1_i := RES_1.Get(i)
		RES_2_i := RES_2.Get(i)
		RES_3_i := RES_3.Get(i)
		//
		res := add(mul(TWO_12, RES_3_i), mul(TWO_8, RES_2_i), mul(TWO_4, RES_1_i), RES_0_i)
		arg1 := add(mul(TWO_4, ARG1_1_i), ARG1_0_i)
		arg2 := add(mul(TWO_4, ARG2_1_i), ARG2_0_i)
		arg1.Mul(&arg1, &arg2)
		// Check decomposition matches
		if res.Cmp(&arg1) != 0 {
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

func add(items ...fr.Element) fr.Element {
	var acc = fr.NewElement(0)
	for _, item := range items {
		acc.Add(&acc, &item)
	}

	return acc
}

func sub(lhs fr.Element, rhs fr.Element) fr.Element {
	lhs.Sub(&lhs, &rhs)
	return lhs
}

func mul(lhs fr.Element, rhs fr.Element) fr.Element {
	lhs.Mul(&lhs, &rhs)
	return lhs
}
