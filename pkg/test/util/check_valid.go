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

	"github.com/consensys/go-corset/pkg/cmd/zkc"
	cmd_util "github.com/consensys/go-corset/pkg/cmd/zkc"
	"github.com/consensys/go-corset/pkg/ir/mir"
	"github.com/consensys/go-corset/pkg/schema/module"
	"github.com/consensys/go-corset/pkg/trace/lt"
	"github.com/consensys/go-corset/pkg/util/collection/array"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/field/bls12_377"
	"github.com/consensys/go-corset/pkg/util/field/gf251"
	"github.com/consensys/go-corset/pkg/util/field/gf8209"
	"github.com/consensys/go-corset/pkg/util/field/koalabear"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast"
	"github.com/consensys/go-corset/pkg/zkc/compiler/codegen"
	"github.com/consensys/go-corset/pkg/zkc/constraints"
	"github.com/consensys/go-corset/pkg/zkc/vm"
)

var (
	// DEFAULT_FIELDS set default fields for testing
	DEFAULT_FIELDS = []field.Config{field.BLS12_377, field.KOALABEAR_16}
	// DEFAULT_CONFIG sets a default testing configuration
	DEFAULT_CONFIG = Config{fields: DEFAULT_FIELDS, constraints: false, nativeLowering: true}
)

// Config for testing
type Config struct {
	// Fields to test over
	fields []field.Config
	// enable constraints checking, or not.
	constraints bool
	// enable testing for native lowering
	nativeLowering bool
}

// Fields determines which fields to test over.
func (p Config) Fields(fields ...field.Config) Config {
	p.fields = fields
	//
	return p
}

// Constraints determines whether or not to check constraints.
func (p Config) Constraints(flag bool) Config {
	p.constraints = flag
	//
	return p
}

// NativeLowering determines whether or not to test native lowerings as well.
func (p Config) NativeLowering(flag bool) Config {
	p.nativeLowering = flag
	//
	return p
}

// CheckValid checks that a given source file compiles without any errors.
// nolint
func CheckValid(t *testing.T, test, ext string, config Config) {
	// Enable testing each trace in parallel
	t.Parallel()
	//
	if len(config.fields) == 0 {
		panic("at least one target field is required")
	}
	// Check for each field requested
	for _, f := range config.fields {
		checkValidInternal(t, test, ext, codegen.DEFAULT_CONFIG, config.constraints, f)
		// check whether to enable lowering as well.
		if config.nativeLowering {
			checkValidInternal(t, test, ext, codegen.DEFAULT_CONFIG.LowerZkcNative(true), config.constraints, f)
		}
	}
}

func checkValidInternal(t *testing.T, test, ext string, codeCfg codegen.Config, constraints bool, field field.Config) {
	var filename = fmt.Sprintf("%s/%s.%s", TestDir, test, ext)
	// Compile source file into Abstract Syntax Tree form.
	program := cmd_util.CompileSourceFiles(field, filename)
	// Compile program into boot machine
	vm, errs := program.Compile(codeCfg.Field(field))
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
			// Run execution tests
			for _, test := range tests {
				runExecutionTest(t, program, vm, test)
			}
			// Run constraint tests
			if constraints {
				for _, test := range tests {
					// FIXME: support reject tests
					if test.expected {
						runConstraintTest(t, codeCfg, program, test, field)
					}
				}
			}
		}
	}
}

func runExecutionTest(t *testing.T, program ast.Program, wm *vm.WordMachine[vm.Uint], test TestCase) {
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

func runConstraintTest(t *testing.T, config codegen.Config, program ast.Program, test TestCase, f field.Config) {
	// Dispatch based on field config
	switch f {
	case field.GF_251:
		testConstraintsWithField[gf251.Element](t, config, program, test, f)
	case field.GF_8209:
		testConstraintsWithField[gf8209.Element](t, config, program, test, f)
	case field.KOALABEAR_16:
		testConstraintsWithField[koalabear.Element](t, config, program, test, f)
	case field.BLS12_377:
		testConstraintsWithField[bls12_377.Element](t, config, program, test, f)
	default:
		panic(fmt.Sprintf("unknown field configuration: %s", f.Name))
	}
}

func testConstraintsWithField[F field.Element[F]](t *testing.T, config codegen.Config, program ast.Program,
	test TestCase, f field.Config) {
	//
	var (
		wvm, tf = executeAndTrace(t, config, program, test.data)
		id      = traceId{f.Name, "MIR", test.filename,
			test.expected, true, true, mir.DEFAULT_OPTIMISATION_INDEX, true, int(test.line), 0}
	)
	// Lower to field machine
	fvm := vm.LowerWordMachine[vm.Uint, F](f, wvm)
	// lower to mir constraints
	avm := constraints.GenerateMirConstraints(fvm)
	// Construct limbs map
	mapping := module.NewLimbsMap[F](f, avm.Modules().Collect()...)
	// generate initial trace
	checkTrace(t, tf, id, avm, mapping)
}

func executeAndTrace(t *testing.T, config codegen.Config, program ast.Program, input map[string][]byte,
) (*vm.WordMachine[vm.Uint], lt.TraceFile) {
	//
	var (
		wm       *vm.WordMachine[vm.Uint]
		errors   []error
		observer vm.TraceObserver[vm.Uint, *vm.WordMachine[vm.Uint]]
	)
	//
	if wm, errors = zkc.ExecuteIrProgram("main", config, program, input, &observer); len(errors) > 0 {
		t.Errorf("%v", errors)
	}
	// Done
	return wm, observer.Trace(wm)
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
