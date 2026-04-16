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
	"os"

	"github.com/consensys/go-corset/pkg/cmd/zkc/debug"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/field/bls12_377"
	"github.com/consensys/go-corset/pkg/util/field/gf251"
	"github.com/consensys/go-corset/pkg/util/field/gf8209"
	"github.com/consensys/go-corset/pkg/util/field/koalabear"
	"github.com/consensys/go-corset/pkg/util/termio"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast"
	"github.com/consensys/go-corset/pkg/zkc/vm/word"
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

func runDebugCmd[F field.Element[F]](cmd *cobra.Command, args []string) {
	// Check whether interactive mode (or not)
	interactive := GetFlag(cmd, "interactive")
	//
	input := ParseInputFile(args[0])
	// Compile source files, or print errors
	program := CompileSourceFiles(args[1:]...)
	//
	if interactive {
		runInteractiveDebugger[F](input, program)
		return
	}
	//
	observer := debug.TraceObserver[word.Uint]{}
	//
	executeIrProgram("main", program, input, &observer)
	//
	fmt.Println()
}

func runInteractiveDebugger[F field.Element[F]](input map[string][]byte, program ast.Program) []error {
	var (
		debugger = constructInteractiveDebugger[F]()
	)
	// Render inspector
	if err := debugger.Render(); err != nil {
		return []error{err}
	}
	//
	return debugger.Start()
}

func constructInteractiveDebugger[F field.Element[F]]() *debug.Debugger {
	//
	var (
		term, err = termio.NewTerminal()
		view      = &debug.TraceView{}
	)
	// Check whether successful
	if err == nil {
		// Construct inspector state
		return debug.NewDebugger(term, view)
	}

	fmt.Println(error.Error(err))
	os.Exit(1)
	// Unreachable
	return nil
}

// ============================================================================
// Misc
// ============================================================================

//nolint:errcheck
func init() {
	debugCmd.Flags().BoolP("interactive", "i", false, "enable interactive debugging")
	rootCmd.AddCommand(debugCmd)
}
