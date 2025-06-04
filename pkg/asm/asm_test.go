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

	"github.com/consensys/go-corset/pkg/asm/io"
	"github.com/consensys/go-corset/pkg/asm/io/micro"
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
func Test_SlowPow(t *testing.T) {
	check(t, "slow_pow")
}
func Test_FastPow(t *testing.T) {
	check(t, "fast_pow")
}
func Test_RecPow(t *testing.T) {
	check(t, "rec_pow")
}
func Test_Wcp(t *testing.T) {
	check(t, "wcp")
}

func Test_Byte(t *testing.T) {
	check(t, "byte")
}

func Test_Shift(t *testing.T) {
	check(t, "shift")
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
		// default config (for now)
		loweringConfig = LoweringConfig{
			MaxFieldWidth:    252,
			MaxRegisterWidth: 128,
			Vectorize:        true,
		}
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
	// Parse terms into an assembly macroProgram
	macroProgram, _, errs := Assemble(*srcfile)
	// Check terms parsed ok
	if len(errs) > 0 {
		t.Fatalf("Error parsing %s: %v\n", filename, errs)
	} else if len(macroProgram.Functions()) == 0 {
		t.Fatalf("Empty test file: %s\n", filename)
	}
	// Record how many tests executed.
	nTests := 0
	// Iterate possible testfile extensions
	for _, cfg := range TESTFILE_EXTENSIONS {
		// Construct test filename
		testFilename := fmt.Sprintf("%s/%s.%s", TestDir, test, cfg.extension)
		macroTraces := ReadBatchedTraceFile(testFilename, macroProgram)
		microTraces := LowerTraces(loweringConfig, macroTraces...)
		// Check traces at ASM level
		checkTraces(t, test, "ASM", cfg, macroTraces)
		// Check traces at uASM level
		checkTraces(t, test, "µASM", cfg, microTraces)
		//
		if len(microTraces) > 0 {
			// Check traces at MIR/AIR levels
			checkIrTraces(t, test, cfg, microTraces)
		}
		// Record how many tests we found
		nTests += len(macroTraces)
	}
	// Sanity check at least one trace found.
	if nTests == 0 {
		panic(fmt.Sprintf("missing any tests for %s", test))
	}
}

// Check the given traces for all function instances.
func checkTraces[T io.Instruction[T]](t *testing.T, test string, ir string, cfg TestConfig, traces []io.Trace[T]) {
	//
	for i, tr := range traces {
		id := traceId{ir, test, cfg.expected, i + 1, 0}

		for _, instance := range tr.Instances() {
			checkFunction(t, id, instance, tr.Program())
		}
	}
}

// Check a given set of tests have an expected outcome (i.e. are
// either accepted or rejected) by a given set of constraints.
func checkIrTraces(t *testing.T, test string, cfg TestConfig, traces []io.Trace[micro.Instruction]) {
	// var (
	// 	maxPadding = MAX_PADDING
	// 	program    = traces[0].Program()
	// 	rawTraces  [][]trace.RawColumn
	// )
	// //
	// mirSchema := CompileBinary(fmt.Sprintf("%s.lisp", test), program)
	// //
	// for _, tr := range traces {
	// 	rawTraces = append(rawTraces, LowerMicroTrace(tr))
	// }
	// //
	// for i, tr := range rawTraces {
	// 	if tr != nil {
	// 		// Lower MIR => AIR
	// 		airSchema := mir.LowerToAir(&mirSchema, mir.DEFAULT_OPTIMISATION_LEVEL)
	// 		// Align trace with schema, and check whether expanded or not.
	// 		for padding := uint(0); padding <= maxPadding; padding++ {
	// 			// Construct trace identifiers
	// 			mirID := traceId{"MIR", test, cfg.expected, i + 1, padding}
	// 			airID := traceId{"AIR", test, cfg.expected, i + 1, padding}
	// 			// Only MIR constraints for traces which must be
	// 			// expanded.  They don't really make sense otherwise.
	// 			checkTrace(t, tr, mirID, mirSchema)
	// 			// Always check AIR constraints
	// 			checkTrace(t, tr, airID, airSchema)
	// 		}
	// 	}
	// }
	panic("todo")
}

func checkTrace[C sc.Constraint](t *testing.T, inputs []trace.RawColumn, id traceId, schema sc.Schema[C]) {
	// Construct the trace
	tr, errs := sc.NewTraceBuilder().
		WithPadding(id.padding).
		WithParallelism(true).
		Build(sc.Any(schema), inputs)
	// Sanity check construction
	if len(errs) > 0 {
		t.Errorf("Trace expansion failed (%s, %s, line %d with padding %d): %s",
			id.ir, id.test, id.line, id.padding, errs)
	} else {
		// Check Constraints
		errs := sc.Accepts(false, 100, schema, tr)
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
func checkFunction[T io.Instruction[T]](t *testing.T, id traceId, instance io.FunctionInstance, program io.Program[T]) {
	outcome, err := CheckInstance(instance, program)
	//
	if outcome == math.MaxUint {
		t.Errorf("Failure (%s, %s, line %d with padding %d): %s",
			id.ir, id.test, id.line, id.padding, err)
	} else if outcome == 0 && !id.expected {
		t.Errorf("Trace accepted incorrectly (%s, %s, line %d with padding %d)",
			id.ir, id.test, id.line, id.padding)
	} else if outcome != 0 && id.expected {
		t.Errorf("Trace rejected incorrectly (%s, %s, line %d with padding %d)",
			id.ir, id.test, id.line, id.padding)
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
