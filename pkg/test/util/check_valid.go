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
	"github.com/consensys/go-corset/pkg/zkc/vm"
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
	for _, f := range fields {
		checkValidInternal(t, test, ext, codegen.DEFAULT_CONFIG, f)
	}
}

// CheckValidWithConfig checks that a given source file compiles and runs correctly using the provided codegen config.
// nolint
func CheckValidWithConfig(t *testing.T, test, ext string, cfg codegen.Config, fields ...field.Config) {
	if len(fields) == 0 {
		panic("at least one target field is required")
	}
	// Check for each field requested
	for _, f := range fields {
		checkValidInternal(t, test, ext, cfg, f)
	}
}

func checkValidInternal(t *testing.T, test, ext string, cfg codegen.Config, field field.Config) {
	var filename = fmt.Sprintf("%s/%s.%s", TestDir, test, ext)
	// Compile source file into Abstract Syntax Tree form.
	program := cmd_util.CompileSourceFiles(field, filename)
	// Compile program into boot machine
	vm, errs := program.Compile(cfg.Field(field))
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

func runTestCases(t *testing.T, program ast.Program, wm *vm.WordMachine[vm.Uint], tests []TestCase) {
	for _, test := range tests {
		runTestCase(t, program, wm, test)
	}
}

func runTestCase(t *testing.T, program ast.Program, wm *vm.WordMachine[vm.Uint], test TestCase) {
	var (
		err                   error
		inputs, outputs, errs = program.DecodeInputsOutputs(test.data)
	)
	// Execute machine
	if err = wm.Boot("main", inputs); err == nil {
		// Execute it
		if _, err = vm.ExecuteAll(wm, 1024); err == nil && test.expected {
			// Check outputs match
			errs = append(errs, checkExpectedOutputs(outputs, wm)...)
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

// BenchZkcAccepts benchmarks one field by running every case from matching
// `.accepts` test files (same selection as CheckValid).
func BenchZkcAccepts(b *testing.B, test string, fields ...field.Config) {
	b.Helper()

	if len(fields) == 0 {
		fields = []field.Config{field.BLS12_377}
	}

	f := fields[0]

	filename := fmt.Sprintf("%s/%s.%s", TestDir, test, "zkc")
	program := cmd_util.CompileSourceFiles(f, filename)

	wm, errs := program.Compile(codegen.DEFAULT_CONFIG.Field(f))

	for _, err := range errs {
		b.Fatal(err)
	}

	var cases []TestCase

	for _, cfg := range TESTFILE_EXTENSIONS {
		if !cfg.expected {
			continue
		}

		if cfg.field != nil && *cfg.field != f {
			continue
		}

		cases = append(cases, ReadTestsFile(cfg, test)...)
	}

	if len(cases) == 0 {
		b.Fatalf("no accept test cases for %s", test)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, tc := range cases {
			benchMustRunAcceptCase(b, program, wm, tc)
		}
	}
}

func benchMustRunAcceptCase(b *testing.B, program ast.Program, wm *vm.WordMachine[vm.Uint], tc TestCase) {
	b.Helper()

	inputs, _, decodeErrs := program.DecodeInputsOutputs(tc.data)

	if len(decodeErrs) != 0 {
		b.Fatal(decodeErrs)
	}

	if err := wm.Boot("main", inputs); err != nil {
		b.Fatal(err)
	}

	if _, err := vm.ExecuteAll(wm, 1024); err != nil {
		b.Fatal(err)
	}
}

func checkExpectedOutputs(outputs map[string][]vm.Uint, wm *vm.WordMachine[vm.Uint]) []error {
	var errors []error
	//
	for _, m := range wm.Modules() {
		// Check whether this is an output memory or not.
		if m, ok := m.(vm.InputOutputMemory[vm.Uint]); ok && m.IsWriteOnly() {
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
