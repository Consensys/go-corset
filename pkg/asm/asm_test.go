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
package asm

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"os"
	"strings"
	"testing"

	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/source"
)

func Test_Pow(t *testing.T) {
	check(t, "pow")
}

func Test_Wcp(t *testing.T) {
	check(t, "wcp")
}

// ===================================================================
// Test Helpers
// ===================================================================

// Determines the (relative) location of the test directory.  That is
// where the corset test files (lisp) and the corresponding traces
// (accepts/rejects) are found.
const TestDir = "../../testdata/asm"

// For a given set of constraints, check that all traces which we
// expect to be accepted are accepted, and all traces that we expect
// to be rejected are rejected.
func check(t *testing.T, test string) {
	var (
		filename = fmt.Sprintf("%s.zkasm", test)
	)
	// Enable testing each trace in parallel
	t.Parallel()
	// Read constraints file
	bytes, err := os.ReadFile(fmt.Sprintf("%s/%s", TestDir, filename))
	// Check test file read ok
	if err != nil {
		t.Fatal(err)
	}
	// Package up as source file
	srcfile := source.NewSourceFile(filename, bytes)
	// Parse terms into an HIR schema
	fns, errs := Assemble(srcfile)
	// Check terms parsed ok
	if len(errs) > 0 {
		t.Fatalf("Error parsing %s: %v\n", filename, errs)
	} else if len(fns) == 0 {
		t.Fatalf("Empty test file: %s\n", filename)
	} else if len(fns) > 1 {
		t.Fatalf("Multi-function tests not (yet) supported: %s\n", filename)
	}
	// Record how many tests executed.
	nTests := 0
	// Iterate possible testfile extensions
	for i, cfg := range TESTFILE_EXTENSIONS {
		// Construct test filename
		testFilename := fmt.Sprintf("%s/%s.%s", TestDir, test, cfg.extension)
		traces := readTracesFile(testFilename, fns)
		// Run tests
		asmID := traceId{"ASM", test, cfg.expected, cfg.expand, cfg.validate, i + 1, 0}
		checkTraces(t, asmID, traces, fns...)
		// Record how many tests we found
		nTests += len(traces)
	}
	// Sanity check at least one trace found.
	if nTests == 0 {
		panic(fmt.Sprintf("missing any tests for %s", test))
	}
}

// Check the given traces for all function instances.
func checkTraces(t *testing.T, id traceId, traces [][][]FunctionInstance, fns ...Function) {
	for _, tr := range traces {
		for i, instance := range tr {
			checkFunction(t, id, uint(i), instance, fns...)
		}
	}
}

// Check the given traces for a particular function instance.
func checkFunction(t *testing.T, id traceId, index uint, insts []FunctionInstance, fns ...Function) {
	// Check each instance
	for _, inst := range insts {
		// Initialise a new interpreter
		interpreter := NewInterpreter(fns...)
		// Execute the program
		outputs := interpreter.Execute(index, inst.Inputs)
		// Checkout results
		for r, actual := range outputs {
			expected, ok := inst.Outputs[r]
			outcome := expected.Cmp(&actual) == 0
			// Check actual output matches expected output
			if !ok {
				t.Errorf("Missing output (%s)", r)
			} else if id.expected && !outcome {
				err := fmt.Sprintf("output %s was %s, but expected %s", r, actual.String(), expected.String())
				t.Errorf("Trace rejected incorrectly (%s, %s, line %d with padding %d): %s",
					id.ir, id.test, id.line, id.padding, err)
			} else if !id.expected && !outcome {
				return
			}
		}
		//
		if len(outputs) != len(inst.Outputs) {
			err := fmt.Sprintf("incorrect number of outputs (was %d but expected %d)", len(outputs), len(inst.Outputs))
			t.Errorf("Trace rejected incorrectly (%s, %s, line %d with padding %d): %s",
				id.ir, id.test, id.line, id.padding, err)
		} else if !id.expected {
			t.Errorf("Trace accepted incorrectly (%s, %s, line %d with padding %d)",
				id.ir, id.test, id.line, id.padding)
		}
	}
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
	// Identifies the line number within the test file that the failing trace
	// original.
	line int
	// Identifies how much padding has been added to the expanded trace.
	padding uint
}

// TestConfig provides a simple mechanism for searching for testfiles.
type TestConfig struct {
	extension string
	expected  bool
	expand    bool
	validate  bool
}

var TESTFILE_EXTENSIONS []TestConfig = []TestConfig{
	// should all pass
	{"accepts", true, true, true},
	{"accepts.bz2", true, true, true},
	{"auto.accepts", true, true, true},
	{"expanded.accepts", true, false, false},
	// should all fail
	{"rejects", false, true, false},
	{"rejects.bz2", false, true, false},
	{"auto.rejects", false, true, false},
	{"expanded.rejects", false, false, false},
}

type Traces map[string]Trace
type Trace map[string][]big.Int

// readTracesFile reads a file containing zero or more traces expressed as JSON, where
// each trace is on a separate line.
func readTracesFile(filename string, fns []Function) [][][]FunctionInstance {
	lines := util.ReadInputFile(filename)
	traces := make([][][]FunctionInstance, len(lines))
	// Read constraints line by line
	for i, line := range lines {
		// Parse input line as JSON
		if line != "" && !strings.HasPrefix(line, ";;") {
			tr, err := readTrace([]byte(line), fns)
			if err != nil {
				msg := fmt.Sprintf("%s:%d: %s", filename, i+1, err)
				panic(msg)
			}

			traces[i] = tr
		}
	}

	return traces
}

func readTrace(bytes []byte, fns []Function) ([][]FunctionInstance, error) {
	var (
		err    error
		traces Traces
	)
	// Unmarshall
	jsonErr := json.Unmarshal(bytes, &traces)
	if jsonErr != nil {
		return nil, jsonErr
	}
	//
	instances := make([][]FunctionInstance, len(fns))
	//
	for i, fn := range fns {
		tr, ok := traces[fn.Name]
		// Sanity check
		if !ok {
			return nil, fmt.Errorf("missing inputs/outputs for function %s\n", fn.Name)
		}
		//
		instances[i], err = readTraceInstances(tr, fn)
		//
		if err != nil {
			return nil, err
		}
	}
	//
	return instances, nil
}

func readTraceInstances(trace Trace, fn Function) ([]FunctionInstance, error) {
	var (
		height uint = math.MaxUint
		count       = 0
	)
	// Initialise register map
	for _, reg := range fn.Registers {
		is_ioreg := (reg.Kind == INPUT_REGISTER || reg.Kind == OUTPUT_REGISTER)
		//
		if _, ok := trace[reg.Name]; !ok && is_ioreg {
			return nil, fmt.Errorf("missing register from trace: %s", reg.Name)
		} else if is_ioreg {
			count++
		}
	}
	// Sanity check no extra registers.
	if len(trace) != count {
		return nil, fmt.Errorf("too many registers in trace (was %d expected %d)", len(trace), count)
	}
	// Sanity check register heights
	for k, vs := range trace {
		n := uint(len(vs))
		if height == math.MaxUint {
			height = n
		} else if height != n {
			return nil, fmt.Errorf("invalid register height: %s", k)
		}
	}
	//
	instances := make([]FunctionInstance, height)
	// Parse the trace
	for i := uint(0); i < height; i++ {
		// Initialise ith function instance
		var instance FunctionInstance
		//
		instance.Inputs = make(map[string]big.Int)
		instance.Outputs = make(map[string]big.Int)

		for _, reg := range fn.Registers {
			is_ioreg := (reg.Kind == INPUT_REGISTER || reg.Kind == OUTPUT_REGISTER)
			// Only consider input / output registers
			if is_ioreg {
				v := trace[reg.Name][i]
				// Check bitwidth
				if v.Cmp(reg.Bound()) >= 0 {
					return nil, fmt.Errorf("value %s out-of-bounds for %dbit register %s", v.String(), reg.Width, reg.Name)
				}
				// Assign as input or output
				if reg.Kind == INPUT_REGISTER {
					instance.Inputs[reg.Name] = v
				} else {
					instance.Outputs[reg.Name] = v
				}
			}
		}
		// Assign ith instance
		instances[i] = instance
	}
	//
	return instances, nil
}
