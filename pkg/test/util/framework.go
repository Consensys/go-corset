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
package util

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/consensys/go-corset/pkg/asm"
	"github.com/consensys/go-corset/pkg/binfile"
	cmd_util "github.com/consensys/go-corset/pkg/cmd/util"
	"github.com/consensys/go-corset/pkg/corset"
	"github.com/consensys/go-corset/pkg/ir"
	"github.com/consensys/go-corset/pkg/ir/mir"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/agnostic"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/trace/json"
	"github.com/consensys/go-corset/pkg/util"
)

// TestDir determines the (relative) location of the test directory.  That is
// where the corset test files (lisp) and the corresponding traces
// (accepts/rejects) are found.
const TestDir = "../../testdata"

// MAX_PADDING determines the maximum amount of padding to use when testing.
// Specifically, every trace is tested with varying amounts of padding upto this
// value.
const MAX_PADDING uint = 7

// Check that all traces which we expect to be accepted are accepted by a given
// set of constraints, and all traces that we expect to be rejected are
// rejected.  A default field is used for these tests (BLS12_377)
func Check(t *testing.T, stdlib bool, test string) {
	CheckWithFields(t, stdlib, test, agnostic.BLS12_377)
}

// CheckWithFields checks that all traces which we expect to be accepted are
// accepted by a given set of constraints, and all traces that we expect to be
// rejected are rejected.  All fields provided are tested against.
func CheckWithFields(t *testing.T, stdlib bool, test string, fields ...agnostic.FieldConfig) {
	// Sanity check
	if len(fields) == 0 {
		panic("no field configurations")
	}
	// Enable testing each trace in parallel
	t.Parallel()
	//
	for _, field := range fields {
		checkWithField(t, stdlib, test, field)
	}
}

func checkWithField(t *testing.T, stdlib bool, test string, field agnostic.FieldConfig) {
	var (
		filenames = matchSourceFiles(test)
		// Configure the stack
		stack = getSchemaStack(stdlib, field, filenames...)
	)
	// Record how many tests executed.
	nTests := 0
	// Iterate possible testfile extensions
	for _, cfg := range TESTFILE_EXTENSIONS {
		var traces [][]trace.BigEndianColumn
		// Construct test filename
		testFilename := fmt.Sprintf("%s/%s.%s", TestDir, test, cfg.extension)
		// Read traces from file
		traces = ReadTracesFile(testFilename)
		// Run tests
		binCheckTraces(t, testFilename, cfg, traces, stack)
		// Record how many tests we found
		nTests += len(traces)
	}
	// Sanity check at least one trace found.
	if nTests == 0 {
		panic(fmt.Sprintf("missing any tests for %s", test))
	}
}

func binCheckTraces(t *testing.T, test string, cfg Config,
	traces [][]trace.BigEndianColumn, stack cmd_util.SchemaStack) {
	// Run checks using schema compiled from source
	for _, opt := range cfg.optlevels {
		// Set optimisation level
		stack.WithOptimisationConfig(mir.OPTIMISATION_LEVELS[opt])
		// Configure stack
		stack.Apply(*stack.BinaryFile())
		// Apply stack
		checkTraces(t, test, MAX_PADDING, opt, cfg, traces, stack)
	}
	// Construct binary schema
	if binSchema := encodeDecodeSchema(t, *stack.BinaryFile()); binSchema != nil {
		// Choose any valid optimisation level
		opt := cfg.optlevels[0]
		// Set optimisation level
		stack.WithOptimisationConfig(mir.OPTIMISATION_LEVELS[opt])
		// Reset the stack for given binary file
		stack.Apply(*binSchema)
		// Run checks using schema from binary file.  Observe, to try and reduce
		// overhead of repeating all the tests we don't consider padding.
		checkTraces(t, test, 0, opt, cfg, traces, stack)
	}
}

// Check a given set of tests have an expected outcome (i.e. are
// either accepted or rejected) by a given set of constraints.
func checkTraces(t *testing.T, test string, maxPadding uint, opt uint, cfg Config,
	traces [][]trace.BigEndianColumn, stack cmd_util.SchemaStack) {
	// For unexpected traces, we never want to explore padding (because that's
	// the whole point of unexpanded traces --- they are raw).
	if !cfg.expand {
		maxPadding = 0
	}
	//
	for i, tr := range traces {
		if tr != nil {
			for _, ir := range []string{"MIR", "AIR"} {
				// Align trace with schema, and check whether expanded or not.
				for padding := uint(0); padding <= maxPadding; padding++ {
					// Construct trace identifier
					id := traceId{ir, test, cfg.expected, cfg.expand, cfg.validate, opt, i + 1, padding}
					//
					if cfg.expand || ir == "AIR" {
						// Always check if expansion required, otherwise
						// only check AIR constraints.
						checkTrace(t, tr, id, stack.SchemaOf(ir), stack.RegisterMapping())
					}
				}
			}
		}
	}
}

func checkTrace[C sc.Constraint](t *testing.T, inputs []trace.BigEndianColumn, id traceId,
	schema sc.Schema[C], mapping sc.RegisterMappings) {
	//
	// Construct the trace
	tr, errs := ir.NewTraceBuilder().
		WithExpansion(id.expand).
		WithValidation(id.validate).
		WithPadding(id.padding).
		WithParallelism(true).
		WithRegisterMapping(mapping).
		Build(sc.Any(schema), inputs)
	// Sanity check construction
	if len(errs) > 0 {
		t.Errorf("Trace expansion failed (%s [O%d], %s, line %d with padding %d): %s",
			id.ir, id.optimisation, id.test, id.line, id.padding, errs)
	} else {
		// Check Constraints
		errs := sc.Accepts(true, 100, schema, tr)
		// Determine whether trace accepted or not.
		accepted := len(errs) == 0
		// Process what happened versus what was supposed to happen.
		if !accepted && id.expected {
			//table.PrintTrace(tr)
			t.Errorf("Trace rejected incorrectly (%s [O%d], %s, line %d with padding %d): %s",
				id.ir, id.optimisation, id.test, id.line, id.padding, errs)
		} else if accepted && !id.expected {
			//printTrace(tr)
			t.Errorf("Trace accepted incorrectly (%s [O%d], %s, line %d with padding %d)",
				id.ir, id.optimisation, id.test, id.line, id.padding)
		}
	}
}

// SRC_EXTENSIONS identifies the set of currently recognised extensions for
// constraint source files.
var SRC_EXTENSIONS = []string{"lisp", "zkasm"}

// This identifies matching source files.
func matchSourceFiles(test string) []string {
	var filenames []string
	//
	for _, ext := range SRC_EXTENSIONS {
		filename := fmt.Sprintf("%s/%s.%s", TestDir, test, ext)
		if _, err := os.Stat(filename); err == nil {
			filenames = append(filenames, filename)
		}
	}
	//
	return filenames
}

// Config provides a simple mechanism for searching for testfiles.
type Config struct {
	extension string
	expected  bool
	expand    bool
	validate  bool
	optlevels []uint
}

var allOptLevels = []uint{0, 1}
var defaultOptLevel = []uint{1}

// TESTFILE_EXTENSIONS identifies the possible file extensions used for
// different test inputs.
var TESTFILE_EXTENSIONS []Config = []Config{
	// should all pass
	{"accepts", true, true, true, allOptLevels},
	{"accepts.bz2", true, true, true, allOptLevels},
	{"auto.accepts", true, true, true, allOptLevels},
	{"expanded.accepts", true, false, false, allOptLevels},
	{"expanded.O1.accepts", true, false, false, defaultOptLevel},
	// should all fail
	{"rejects", false, true, false, allOptLevels},
	{"rejects.bz2", false, true, false, allOptLevels},
	{"auto.rejects", false, true, false, allOptLevels},
	{"expanded.rejects", false, false, false, allOptLevels},
	{"expanded.O1.rejects", false, false, false, defaultOptLevel},
}

// A trace identifier uniquely identifies a specific trace within a given test.
// This is used to provide debug information about a trace failure.
// Specifically, so the user knows which line in which file caused the problem.
type traceId struct {
	// Identifies the Intermediate Representation tested against.
	ir string
	// Identifies the test name.  From this, the test filename can be determined
	// in conjunction with the expected outcome.
	test string
	// Identifies whether this trace should be accepted (true) or rejected
	// (false).
	expected bool
	// Identifies whether this trace should be expanded (or not).
	expand bool
	// Identifies whether this trace should be validate (or not).
	validate bool
	// Optimisation level
	optimisation uint
	// Identifies the line number within the test file that the failing trace
	// original.
	line int
	// Identifies how much padding has been added to the expanded trace.
	padding uint
}

// ReadTracesFile reads a file containing zero or more traces expressed as JSON, where
// each trace is on a separate line.
func ReadTracesFile(filename string) [][]trace.BigEndianColumn {
	lines := util.ReadInputFile(filename)
	traces := make([][]trace.BigEndianColumn, len(lines))
	// Read constraints line by line
	for i, line := range lines {
		// Parse input line as JSON
		if line != "" && !strings.HasPrefix(line, ";;") {
			// Read traces
			tr, err := json.FromBytes([]byte(line))
			//
			if err != nil {
				msg := fmt.Sprintf("%s:%d: %s", filename, i+1, err)
				panic(msg)
			}

			traces[i] = tr
		}
	}

	return traces
}

// This is a little test to ensure the binary file format (specifically the
// binary encoder / decoder) works as expected.
func encodeDecodeSchema(t *testing.T, binf binfile.BinaryFile) *binfile.BinaryFile {
	var nbinf binfile.BinaryFile
	// Turn the binary file into bytes
	bytes, err := binf.MarshalBinary()
	// Encode schema
	if err != nil {
		t.Error(err)
		return nil
	}
	// Decode schema
	if err := nbinf.UnmarshalBinary(bytes); err != nil {
		t.Error(err)
		return nil
	}
	//
	return &nbinf
}

func getSchemaStack(stdlib bool, field agnostic.FieldConfig, filenames ...string) cmd_util.SchemaStack {
	//
	var (
		stack        cmd_util.SchemaStack
		corsetConfig corset.CompilationConfig
		asmConfig    asm.LoweringConfig
	)
	// Configure corset for testing
	corsetConfig.Legacy = true
	corsetConfig.Stdlib = stdlib
	// Configure asm for lowering
	asmConfig.Vectorize = true
	asmConfig.MaxFieldWidth = field.FieldBandWidth
	asmConfig.MaxRegisterWidth = field.RegisterWidth
	//
	stack.
		WithCorsetConfig(corsetConfig).
		WithAssemblyConfig(asmConfig).
		WithLayer(cmd_util.MACRO_ASM_LAYER).
		WithLayer(cmd_util.MICRO_ASM_LAYER).
		WithLayer(cmd_util.MIR_LAYER).
		WithLayer(cmd_util.AIR_LAYER)
	// Read in all specified constraint files.
	stack.Read(filenames...)
	//
	return stack
}
