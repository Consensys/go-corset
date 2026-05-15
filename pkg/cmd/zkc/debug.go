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
package zkc

import (
	"fmt"

	"github.com/consensys/go-corset/pkg/cmd/zkc/debug"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/field/bls12_377"
	"github.com/consensys/go-corset/pkg/util/field/gf251"
	"github.com/consensys/go-corset/pkg/util/field/gf8209"
	"github.com/consensys/go-corset/pkg/util/field/koalabear"
	"github.com/consensys/go-corset/pkg/util/source"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast"
	"github.com/consensys/go-corset/pkg/zkc/compiler/codegen"
	"github.com/consensys/go-corset/pkg/zkc/vm"
	"github.com/spf13/cobra"
)

var debugCmd = &cobra.Command{
	Use:   "debug [flags] input.json file1.zkc file2.zkc ...",
	Short: "Debug a zkc program.",
	Long:  `Debug a zkc program to produce a set of outputs a from given a set of inputs.`,
	Run: func(cmd *cobra.Command, args []string) {
		runFieldAgnosticCmd(cmd, args, debugCmds)
	},
}

// Available instances
var debugCmds = []FieldAgnosticCmd{
	{field.GF_251, runDebugCmd[gf251.Element]},
	{field.GF_8209, runDebugCmd[gf8209.Element]},
	{field.KOALABEAR_16, runDebugCmd[koalabear.Element]},
	{field.BLS12_377, runDebugCmd[bls12_377.Element]},
}

func runDebugCmd[F field.Element[F]](cmd *cobra.Command, args []string, field field.Config) {
	var (
		// compiler config
		config = codegen.DEFAULT_CONFIG.
			Vectorize(GetFlag(cmd, "vectorize")).
			Field(field)
	)
	//
	input := ParseInputFile(args[0])
	// Compile source files, or print errors
	program := CompileSourceFiles(field, args[1:]...)
	//
	observer := debug.TraceObserver[vm.Uint]{}
	//
	ExecuteIrProgram("main", config, program, input, &observer)
	//
	fmt.Println()
}

// ExecuteIrProgram provides a generic means of executing a given program with a
// given view.  This can return a nil machine if compilation failed.  However,
// it can also return a valid machine with errors in the case it compiled
// successfully, but failed during execution.
func ExecuteIrProgram[V vm.BaseObserver[vm.Uint]](mainFn string, config codegen.Config, program ast.Program,
	input map[string][]byte, view V,
) (*vm.WordMachine[vm.Uint], []error) {
	var (
		wm        *vm.WordMachine[vm.Uint]
		bigInputs map[string][]vm.Uint
		errors    []error
	)
	// Execute machine in chunks of 1K steps
	if bigInputs, _, errors = program.DecodeInputsOutputs(input); len(errors) == 0 {
		// Build our machine
		var compileErrs []source.SyntaxError

		wm, compileErrs = program.Compile(config)
		for _, e := range compileErrs {
			errors = append(errors, &e)
		}
		//
		if len(errors) == 0 {
			if err := wm.Boot(mainFn, bigInputs); err != nil {
				errors = append(errors, err)
			} else if _, err := vm.ExecuteAndObserve(wm, 1, view); err != nil {
				errors = append(errors, err)
			}
		}
	}
	// return machine + errors (if errors)
	return wm, errors
}

// ============================================================================
// Misc
// ============================================================================

//nolint:errcheck
func init() {
	rootCmd.AddCommand(debugCmd)
}
