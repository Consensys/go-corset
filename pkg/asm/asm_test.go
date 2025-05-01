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
	"fmt"
	"math"
	"os"
	"testing"

	"github.com/consensys/go-corset/pkg/mir"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util/source"
)

func Test_Counter(t *testing.T) {
	check(t, "counter")
}
func Test_Max(t *testing.T) {
	check(t, "max")
}
func Test_Pow(t *testing.T) {
	check(t, "pow")
}

func Test_Wcp(t *testing.T) {
	check(t, "wcp")
}

// ===================================================================
// Test Helpers
// ===================================================================

// Determines the maximum amount of padding to use when testing.  Specifically,
// every trace is tested with varying amounts of padding upto this value.
const MAX_PADDING uint = 2

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
	fns, _, errs := Parse(srcfile)
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
	for _, cfg := range TESTFILE_EXTENSIONS {
		// Construct test filename
		testFilename := fmt.Sprintf("%s/%s.%s", TestDir, test, cfg.extension)
		traces := ReadBatchedTraceFile(testFilename, fns)
		// Run tests
		checkTraces(t, test, cfg, traces, fns...)
		checkIrTraces(t, test, cfg, traces, fns...)
		// Record how many tests we found
		nTests += len(traces)
	}
	// Sanity check at least one trace found.
	if nTests == 0 {
		panic(fmt.Sprintf("missing any tests for %s", test))
	}
}

// Check the given traces for all function instances.
func checkTraces(t *testing.T, test string, cfg TestConfig, traces [][]FunctionInstance, fns ...MacroFunction) {
	//
	for i, tr := range traces {
		id := traceId{"ASM", test, cfg.expected, i + 1, 0}

		for _, instance := range tr {
			checkFunction(t, id, instance, fns...)
		}
	}
}

// Check a given set of tests have an expected outcome (i.e. are
// either accepted or rejected) by a given set of constraints.
func checkIrTraces(t *testing.T, test string, cfg TestConfig, instances [][]FunctionInstance, fns ...MacroFunction) {
	var (
		maxPadding = MAX_PADDING
		builder    = NewTraceBuilder(fns...)
		traces     [][]trace.RawColumn
	)
	//
	binFile, errs := NewCompiler().Compile(fns...)
	hirSchema := &binFile.Schema
	//
	if len(errs) > 0 {
		// should be unreachable.
		t.Fatalf("Error compiling %s: %v\n", test, errs)
	}
	//
	for _, inst := range instances {
		traces = append(traces, builder.Build(inst))
	}
	//
	for i, tr := range traces {
		if tr != nil {
			// Lower HIR => MIR
			mirSchema := hirSchema.LowerToMir()
			// Lower MIR => AIR
			airSchema := mirSchema.LowerToAir(mir.DEFAULT_OPTIMISATION_LEVEL)
			// Align trace with schema, and check whether expanded or not.
			for padding := uint(0); padding <= maxPadding; padding++ {
				// Construct trace identifiers
				hirID := traceId{"HIR", test, cfg.expected, i + 1, padding}
				mirID := traceId{"MIR", test, cfg.expected, i + 1, padding}
				airID := traceId{"AIR", test, cfg.expected, i + 1, padding}
				// Only HIR / MIR constraints for traces which must be
				// expanded.  They don't really make sense otherwise.
				checkTrace(t, tr, hirID, hirSchema)
				checkTrace(t, tr, mirID, mirSchema)
				// Always check AIR constraints
				checkTrace(t, tr, airID, airSchema)
			}
		}
	}
}

func checkTrace(t *testing.T, inputs []trace.RawColumn, id traceId, schema sc.Schema) {
	// Construct the trace
	tr, errs := sc.NewTraceBuilder(schema).
		Padding(id.padding).
		Parallel(true).
		Build(inputs)
	// Sanity check construction
	if len(errs) > 0 {
		t.Errorf("Trace expansion failed (%s, %s, line %d with padding %d): %s",
			id.ir, id.test, id.line, id.padding, errs)
	} else {
		// Check Constraints
		_, errs1 := sc.Accepts(false, 100, schema, tr)
		// Check assertions
		_, errs2 := sc.Asserts(true, 100, schema, tr)
		errs := append(errs1, errs2...)
		// Determine whether trace accepted or not.
		accepted := len(errs) == 0
		// Process what happened versus what was supposed to happen.
		if !accepted && id.expected {
			//table.PrintTrace(tr)
			t.Errorf("Trace rejected incorrectly (%s, %s, line %d with padding %d): %s",
				id.ir, id.test, id.line, id.padding, errs)
		} else if accepted && !id.expected {
			//printTrace(tr)
			t.Errorf("Trace accepted incorrectly (%s, %s, line %d with padding %d)",
				id.ir, id.test, id.line, id.padding)
		}
	}
}

// Check the given traces for a particular function instance.
func checkFunction(t *testing.T, id traceId, instance FunctionInstance, fns ...MacroFunction) {
	outcome, err := CheckInstance(instance, fns)
	//
	if outcome == math.MaxUint {
		t.Errorf("Failure (%s, line %d with padding %d): %s",
			id.test, id.line, id.padding, err)
	} else if outcome == 0 && !id.expected {
		t.Errorf("Trace accepted incorrectly (%s, line %d with padding %d)",
			id.test, id.line, id.padding)
	} else if outcome != 0 && id.expected {
		t.Errorf("Trace rejected incorrectly (%s, line %d with padding %d)",
			id.test, id.line, id.padding)
	}
}

// A trace identifier uniquely identifies a specific trace within a given test.
// This is used to provide debug information about a trace failure.
// Specifically, so the user knows which line in which file caused the problem.
type traceId struct {
	ir string
	// Identifies the test name.  From this, the test filename can be determined
	// in conjunction with the expected outcome.
	test string
	// Identifies whether this trace should be accepted (true) or rejected
	// (false).
	expected bool
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
}

var TESTFILE_EXTENSIONS []TestConfig = []TestConfig{
	// should all pass
	{"accepts", true},
	{"accepts.bz2", true},
	{"auto.accepts", true},
	{"expanded.accepts", true},
	// should all fail
	{"rejects", false},
	{"rejects.bz2", false},
	{"auto.rejects", false},
	{"expanded.rejects", false},
}
