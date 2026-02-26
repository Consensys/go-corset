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
	"math/big"
	"os"

	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/field/bls12_377"
	"github.com/consensys/go-corset/pkg/util/field/gf251"
	"github.com/consensys/go-corset/pkg/util/field/gf8209"
	"github.com/consensys/go-corset/pkg/util/field/koalabear"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast"
	"github.com/consensys/go-corset/pkg/zkc/vm/machine"
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
	program := CompileSourceFiles(args[1:])
	//
	if ir {
		executeIrProgram(program, input)
	}
}

func executeIrProgram(program ast.Program, input map[string][]byte) {
	var (
		vm     machine.Core[big.Int, ast.Instruction]
		errors []error
	)
	// Execute machine in chunks of 1K steps
	if vm, errors = program.BootMachine(input, "main"); len(errors) == 0 {
		if _, err := machine.ExecuteAll(vm, 1024); err != nil {
			// NOTE: determine stack trace!
			errors = append(errors, err)
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
	// write output
	for i := range vm.State().NumOutputs() {
		var output = vm.State().Output(i)
		//
		fmt.Printf("%s", output.Name())
		//
		for i, val := range output.Contents() {
			if i != 0 {
				fmt.Printf(", ")
			}
			//
			fmt.Printf("0x%s", val.Text(16))
		}
		//
		fmt.Println()
	}
}

//nolint:errcheck
func init() {
	rootCmd.AddCommand(executeCmd)
	executeCmd.Flags().Bool("ir", false, "execute intermediate representation (IR)")
}
