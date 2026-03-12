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
	"encoding/hex"
	"fmt"
	"os"

	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/field/bls12_377"
	"github.com/consensys/go-corset/pkg/util/field/gf251"
	"github.com/consensys/go-corset/pkg/util/field/gf8209"
	"github.com/consensys/go-corset/pkg/util/field/koalabear"
	"github.com/consensys/go-corset/pkg/util/source"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast"
	"github.com/consensys/go-corset/pkg/zkc/vm/machine"
	"github.com/consensys/go-corset/pkg/zkc/vm/memory"
	"github.com/consensys/go-corset/pkg/zkc/vm/word"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var executeCmd = &cobra.Command{
	Use:     "execute [flags] input.json file1.zkc file2.zkc ...",
	Short:   "Execute a zkc program.",
	Long:    `Execute a zkc program to produce a set of outputs a from given a set of inputs.`,
	Aliases: []string{"exec"},
	Run: func(cmd *cobra.Command, args []string) {
		runFieldAgnosticCmd(cmd, args, executeCmds)
	},
}

// Available instances
var executeCmds = []FieldAgnosticCmd{
	{field.GF_251, runExecuteCmd[gf251.Element]},
	{field.GF_8209, runExecuteCmd[gf8209.Element]},
	{field.KOALABEAR_16, runExecuteCmd[koalabear.Element]},
	{field.BLS12_377, runExecuteCmd[bls12_377.Element]},
}

func runExecuteCmd[F field.Element[F]](cmd *cobra.Command, args []string) {
	//
	ir := GetFlag(cmd, "ir")
	//
	input := ParseInputFile(args[0])
	// Compile source files, or print errors
	program := CompileSourceFiles(args[1:]...)
	//
	if ir {
		executeIrProgram("main", program, input)
	}
}

func executeIrProgram(mainFn string, program ast.Program, input map[string][]byte) {
	var (
		vm        *machine.Base[word.Uint]
		bigInputs map[string][]word.Uint
		errors    []error
	)
	// Execute machine in chunks of 1K steps
	if bigInputs, _, errors = program.DecodeInputsOutputs(input); len(errors) == 0 {
		// Build our machine
		var compileErrs []source.SyntaxError

		vm, compileErrs = program.Compile()
		for _, e := range compileErrs {
			errors = append(errors, &e)
		}
		//
		if len(errors) == 0 {
			if err := vm.Boot(mainFn, bigInputs); err != nil {
				errors = append(errors, err)
			} else if _, err := machine.ExecuteAll(vm, 1024); err != nil {
				// NOTE: determine stack trace!
				errors = append(errors, err)
			}
		}
	}
	// Exit with failure (if errors)
	if len(errors) > 0 {
		// Log errors
		for _, e := range errors {
			log.Error(fmt.Sprintf("%s (IR)", e))
		}
		//
		os.Exit(4)
	}
	// Collect raw outputs from write-once memories
	rawOutputs := make(map[string][]word.Uint)
	for _, m := range vm.Modules() {
		if output, ok := m.(*memory.WriteOnce[word.Uint]); ok {
			rawOutputs[output.Name()] = output.Contents()
		}
	}
	// Encode outputs back to bytes
	encodedOutputs, encErrors := program.EncodeInputsOutputs(rawOutputs)
	//
	for _, e := range encErrors {
		log.Error(fmt.Sprintf("%s (IR)", e))
	}
	// Write output
	for name, bytes := range encodedOutputs {
		fmt.Printf("%s = 0x%s\n", name, hex.EncodeToString(bytes))
	}
}

//nolint:errcheck
func init() {
	rootCmd.AddCommand(executeCmd)
	executeCmd.Flags().Bool("ir", false, "execute intermediate representation (IR)")
}
