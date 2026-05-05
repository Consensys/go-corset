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
	"testing"

	cmd_util "github.com/consensys/go-corset/pkg/cmd/zkc"
	"github.com/consensys/go-corset/pkg/util/collection/array"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast"
	"github.com/consensys/go-corset/pkg/zkc/compiler/codegen"
	"github.com/consensys/go-corset/pkg/zkc/vm/machine"
	"github.com/consensys/go-corset/pkg/zkc/vm/memory"
	"github.com/consensys/go-corset/pkg/zkc/vm/word"
)

// CheckValid checks that a given source file compiles without any errors.
// nolint
func CheckValid(t *testing.T, test, ext string, fields ...field.Config) {
	// Enable testing each trace in parallel
	t.Parallel()
	//
	if len(fields) == 0 {
		panic("at least one target field is required")
	}
	// Check for each field requested
	for _, field := range fields {
		checkValidInternal(t, test, ext, field)
	}
}

func checkValidInternal(t *testing.T, test, ext string, field field.Config) {
	var filename = fmt.Sprintf("%s/%s.%s", TestDir, test, ext)
	// Compile source file into Abstract Syntax Tree form.
	program := cmd_util.CompileSourceFiles(field, filename)
	// Compile program into boot machine
	vm, errs := program.Compile(codegen.DEFAULT_CONFIG.Field(field))
	for _, err := range errs {
		t.Errorf("%s", err.Error())
	}
	//
	if len(errs) > 0 {
		return
	}
	// Search for tests
	for _, cfg := range TESTFILE_EXTENSIONS {
		// Check suitable field
		if cfg.field == nil || *cfg.field == field {
			// Read tests from file
			tests := ReadTestsFile(cfg, test)
			// Run all tests
			runTestCases(t, program, vm, tests)
		}
	}
}

func runTestCases(t *testing.T, program ast.Program, bootVm *machine.Word[word.Uint], tests []TestCase) {
	for _, test := range tests {
		runTestCase(t, program, bootVm, test)
	}
}

func runTestCase(t *testing.T, program ast.Program, vm *machine.Word[word.Uint], test TestCase) {
	var (
		err                   error
		inputs, outputs, errs = program.DecodeInputsOutputs(test.data)
	)
	// Execute machine
	if err = vm.Boot("main", inputs); err == nil {
		// Execute it
		if _, err = machine.ExecuteAll(vm, 1024); err == nil && test.expected {
			// Check outputs match
			errs = append(errs, checkExpectedOutputs(outputs, vm)...)
		} else if err == nil && !test.expected {
			errs = append(errs, fmt.Errorf("test accepted incorrectly"))
		} else if !test.expected {
			// prevent error as this was expected
			err = nil
		}
	}
	// Include single error
	if err != nil {
		errs = append(errs, err)
	}
	// Fail if errors found
	for _, err := range errs {
		t.Errorf("%s:%d %v", test.filename, test.line, err)
	}
}

func checkExpectedOutputs(outputs map[string][]word.Uint, vm *machine.Word[word.Uint]) []error {
	var errors []error
	//
	for _, m := range vm.Modules() {
		// Check whether this is an output memory or not.
		if m, ok := m.(*memory.WriteOnce[word.Uint]); ok {
			if output, ok := outputs[m.Name()]; ok {
				if c := array.Compare(output, m.Contents()); c != 0 {
					errors = append(errors, fmt.Errorf("incorrect output (expected %v, actual %v)", output, m.Contents()))
				}
			}
		}
	}
	//
	return errors
}

// TestConfig provides a simple mechanism for searching for testfiles.
type TestConfig struct {
	extension string
	expected  bool
	// Indicates extension only suitable for specific field.  If nil, then
	// suitable for all fields.
	field *field.Config
}

// TESTFILE_EXTENSIONS identifies the possible file extensions used for
// different test inputs.
var TESTFILE_EXTENSIONS []TestConfig = []TestConfig{
	// should all pass
	{"accepts", true, nil},
	{"accepts.bz2", true, nil},
	{"gf_251.accepts", true, &field.GF_251},
	{"gf_8209.accepts", true, &field.GF_8209},
	{"koalabear_16.accepts", true, &field.KOALABEAR_16},
	{"bls12_377.accepts", true, &field.BLS12_377},
	// should all fail
	{"rejects", false, nil},
	{"rejects.bz2", false, nil},
	{"gf_251.rejects", false, &field.GF_251},
	{"gf_8209.rejects", false, &field.GF_8209},
	{"koalabear_16.rejects", false, &field.KOALABEAR_16},
	{"bls12_377.rejects", false, &field.BLS12_377},
}
