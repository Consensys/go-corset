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

	"github.com/consensys/go-corset/pkg/cmd/corset"
	"github.com/consensys/go-corset/pkg/ir"
	"github.com/consensys/go-corset/pkg/schema/module"
	"github.com/consensys/go-corset/pkg/trace/lt"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/field/bls12_377"
	"github.com/consensys/go-corset/pkg/util/field/gf251"
	"github.com/consensys/go-corset/pkg/util/field/gf8209"
	"github.com/consensys/go-corset/pkg/util/field/koalabear"
	"github.com/consensys/go-corset/pkg/util/source"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast"
	"github.com/consensys/go-corset/pkg/zkc/compiler/codegen"
	"github.com/consensys/go-corset/pkg/zkc/constraints"
	"github.com/consensys/go-corset/pkg/zkc/vm"
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

func runExecuteCmd[F field.Element[F]](cmd *cobra.Command, args []string, field field.Config) {
	var (
		// compiler config
		config = codegen.DEFAULT_CONFIG.
			Vectorize(GetFlag(cmd, "vectorize")).
			Field(field)
		// output file for trace
		output = GetString(cmd, "output")
		// check constraints
		check = GetFlag(cmd, "check")
		// identify whether tracing required or not.
		tracing = check || output != ""
		// machine used for execution
		wm *vm.WordMachine[vm.Uint]
		//
		tf lt.TraceFile
	)
	// Construct trace builder
	builder := ir.NewTraceBuilder[F]().
		WithValidation(true).
		WithDefensivePadding(true).
		WithExpansion(true).
		WithParallelism(true).
		WithBatchSize(1024)
	//
	input := ParseInputFile(args[0])
	// Compile source files, or print errors
	program := CompileSourceFiles(field, args[1:]...)
	// Execute program (in either fast or slow mode)
	if tracing {
		wm, tf = executeAndTrace("main", config, program, input)
	} else {
		wm = executeNoTrace("main", config, program, input)
	}
	// Generate output
	if output == "" {
		printOutput(program, wm)
	} else {
		WriteTraceFile(output, tf)
	}
	// Check constraints (if requested)
	if check {
		checkConstraints(builder, field, wm, tf)
	}
}

func executeNoTrace(mainFn string, config codegen.Config, program ast.Program, input map[string][]byte,
) *vm.WordMachine[vm.Uint] {
	//
	var (
		wm     *vm.WordMachine[vm.Uint]
		errors []error
	)
	//
	if wm, errors = executeIrProgram(mainFn, config, program, input, vm.EmptyBaseObserver{}); len(errors) > 0 {
		// Log errors
		for _, e := range errors {
			log.Error(fmt.Sprintf("%s (IR)", e))
		}
		//
		os.Exit(4)
	}
	// Done
	return wm
}

func executeAndTrace(mainFn string, config codegen.Config, program ast.Program, input map[string][]byte,
) (*vm.WordMachine[vm.Uint], lt.TraceFile) {
	//
	var (
		wm       *vm.WordMachine[vm.Uint]
		errors   []error
		observer vm.TraceObserver[vm.Uint, *vm.WordMachine[vm.Uint]]
	)
	//
	if wm, errors = executeIrProgram(mainFn, config, program, input, &observer); len(errors) > 0 {
		// Log errors
		for _, e := range errors {
			log.Error(fmt.Sprintf("%s (IR)", e))
		}
		//
		os.Exit(4)
	}
	// Done
	return wm, observer.Trace(wm)
}

func printOutput(program ast.Program, wm *vm.WordMachine[vm.Uint]) {
	// Collect raw outputs from write-once memories
	rawOutputs := make(map[string][]vm.Uint)
	//
	for _, m := range wm.Modules() {
		if output, ok := m.(vm.InputOutputMemory[vm.Uint]); ok && output.IsWriteOnly() {
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

func checkConstraints[F field.Element[F]](builder ir.TraceBuilder[F], config field.Config, wm *vm.WordMachine[vm.Uint],
	tf lt.TraceFile) {
	//
	var cfg corset.CheckConfig
	// Set sensible defaults (for now)
	cfg.Report = true
	cfg.ReportCellWidth = 32
	cfg.ReportTitleWidth = 40
	cfg.ReportPadding = 2
	cfg.ReportLimbs = true
	cfg.ReportComputed = true
	cfg.AnsiEscapes = true
	// Lower to field machine
	fvm := vm.LowerWordMachine[vm.Uint, F](config, wm)
	// Generate MIR constraints
	avm := constraints.GenerateMirConstraints(fvm)
	// Construct limbs map
	mapping := module.NewLimbsMap[F](config, avm.Modules().Collect()...)
	// Register mappin
	builder = builder.WithRegisterMapping(mapping)
	// check the trace
	if !corset.CheckTrace("MIR", avm, tf, builder, cfg) {
		os.Exit(4)
	}
}

func executeIrProgram[V vm.BaseObserver[vm.Uint]](mainFn string, config codegen.Config, program ast.Program,
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
			} else if _, err := execute(wm, 1, view); err != nil {
				errors = append(errors, err)
			}
		}
	}
	// return machine + errors (if errors)
	return wm, errors
}

func execute[W vm.Word[W], V vm.BaseObserver[W]](machine *vm.WordMachine[W], n uint, observer V) (uint, error) {
	var (
		nsteps uint
	)
	//
	observer.Initialise(machine)
	//
	for {
		// observe pre execution
		observer.PreExecution(machine)
		// Execute upto n steps
		m, err := machine.Execute(n)
		// observe pre execution
		observer.PostExecution(machine)
		// update the tally
		nsteps += m
		// check for termination
		if err != nil || m < n {
			return nsteps, err
		}
	}
}

// ============================================================================
// Machine observers
// ============================================================================

// ============================================================================
// Misc
// ============================================================================

//nolint:errcheck
func init() {
	rootCmd.AddCommand(executeCmd)
	executeCmd.Flags().StringP("output", "o", "", "specify output file for writing trace")
	executeCmd.Flags().BoolP("check", "c", false, "check generated trace against constraints")
}
