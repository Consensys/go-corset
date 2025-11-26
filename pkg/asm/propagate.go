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
	"math/big"
	"strings"

	"github.com/consensys/go-corset/pkg/asm/io"
	"github.com/consensys/go-corset/pkg/ir"
	"github.com/consensys/go-corset/pkg/ir/mir"
	"github.com/consensys/go-corset/pkg/schema"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/trace/lt"
	"github.com/consensys/go-corset/pkg/util/collection/array"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/word"
)

// RawModule provides a convenient alias
type RawModule = lt.Module[word.BigEndian]

// PropagateAll propagates secondary (i.e. derivative) function instances
// throughout one or more traces.  NOTES:
//
// Parallelism?
// Validation?
// Batch size?
// Recursion limit (to prevent infinite loops)
func PropagateAll[F field.Element[F], T io.Instruction[T]](p MixedProgram[F, T], ts []lt.TraceFile,
) ([]lt.TraceFile, []error) {
	var (
		errors  []error
		ntraces = make([]lt.TraceFile, len(ts))
	)
	//
	for i, trace := range ts {
		var errs []error
		// NOTE: its possible to get empty traces which arise from comment lines
		// in a trace batch.  Whilst it is kind of awkward, we want to preserve
		// the empty traces as this helps error reporting with respect to line
		// numbers.
		if trace.Modules != nil {
			ntraces[i], errs = Propagate(p, trace)
			errors = append(errors, errs...)
		}
	}
	//
	return ntraces, errors
}

// Propagate secondary (i.e. derivative) function instances throughout a trace.
// For example, suppose two functions "f(x)" and "g(y)", where f(x) calls g(y).
// Then, consider a trace which contains exactly one instance "f(a)=b". Since
// f(x) calls g(y) we may (depending on the exact implementation of f(x) and the
// parameter given) require a secondary instance, say "g(y)=c", to be added to
// make the trace complete with respect to the original instance. Trace
// propagation is about figuring out what secondary instances are required, and
// adding them to the trace.
//
// NOTES:
//
// Parallelism?
// Validation?
// Batch size?
// Recursion limit (to prevent infinite loops)
func Propagate[F field.Element[F], T io.Instruction[T]](p MixedProgram[F, T], trace lt.TraceFile,
) (lt.TraceFile, []error) {
	// Construct suitable executior for the given program
	var (
		errors []error
		n      = uint(len(p.program.Functions()))
		//
		executor = io.NewExecutor(p.program)
		// Clone heap in trace file, since will mutate this.
		heap = trace.Heap.Clone()
	)
	// Perform trace alignment
	trace.Modules, errors = ir.AlignTrace(p.Modules().Collect(), trace.Modules, true)
	// Sanity check for errors
	if len(errors) > 0 {
		return lt.TraceFile{}, errors
	}
	// Write seed instances
	errors = writeInstances(p, n, trace.Modules, executor)
	// Read out generated instances
	modules := readInstances(&heap, p.program, executor)
	// Append external modules (which are unaffected by propagation).
	modules = append(modules, trace.Modules[n:]...)
	// Done
	return lt.NewTraceFile(trace.Header.MetaData, heap, modules), errors
}

// WriteInstances writes all of the instances defined in the given trace columns
// into the executor which, in turn, forces it to execute the relevant
// functions, and functions they call, etc.
func writeInstances[F field.Element[F], T io.Instruction[T]](p MixedProgram[F, T], n uint,
	trace []lt.Module[word.BigEndian], executor *io.Executor[T]) []error {
	//
	var errors []error
	//
	for i, m := range trace[:n] {
		errs := writeFunctionInstances(uint(i), p.program, m, executor)
		errors = append(errors, errs...)
	}
	// Write all from non-assembly modules
	for i, m := range trace[n:] {
		var extern = p.externs[i]
		// Write instances from any external calls
		for _, call := range extractExternalCalls(extern) {
			errs := writeExternCall(call, p.program, m, executor)
			errors = append(errors, errs...)
		}
	}
	//
	return errors
}

func writeFunctionInstances[T io.Instruction[T]](fid uint, p io.Program[T], mod RawModule,
	executor *io.Executor[T]) []error {
	//
	var (
		height  = mod.Height()
		fn      = p.Function(fid)
		inputs  = make([]big.Int, fn.NumInputs())
		outputs = make([]big.Int, fn.NumOutputs())
		errors  []error
	)
	// Invoke padding instance
	extractFunctionPadding(fn.Registers(), inputs, outputs)
	// Execute function call to produce outputs
	errors = executeAndCheck(fid, fn.Name(), inputs, outputs, executor)
	// Invoke each user-defined instance in turn
	for i := range height {
		// Extract function inputs
		extractFunctionColumns(i, mod, inputs, outputs)
		// Execute function call to produce outputs
		errs := executeAndCheck(fid, fn.Name(), inputs, outputs, executor)
		errors = append(errors, errs...)
	}
	//
	return errors
}

// Extract any external function calls found within the given module, returning
// them as an array.
func extractExternalCalls[F field.Element[F], M sc.Module[F]](extern M) []mir.FunctionCall[F] {
	var calls []mir.FunctionCall[F]
	//
	for iter := extern.Constraints(); iter.HasNext(); {
		c := iter.Next()
		// This should always hold
		if hc, ok := c.(mir.Constraint[F]); ok {
			// Check whether its a call or not
			if call, ok := hc.Unwrap().(mir.FunctionCall[F]); ok {
				// Yes, so record it
				calls = append(calls, call)
			}
		}
	}
	//
	return calls
}

// Write any function instances arising from the given call.
func writeExternCall[F field.Element[F], T io.Instruction[T]](call mir.FunctionCall[F], p io.Program[T], mod RawModule,
	executor *io.Executor[T]) []error {
	//
	var (
		trMod   = &ltModuleAdaptor[F]{mod}
		height  = mod.Height()
		fn      = p.Function(call.Callee)
		inputs  = make([]big.Int, fn.NumInputs())
		outputs = make([]big.Int, fn.NumOutputs())
		errors  []error
	)
	//
	if call.Selector.HasValue() {
		var selector = call.Selector.Unwrap()
		// Invoke each user-defined instance in turn
		for i := range height {
			// execute if selector enabled
			if enabled, _, err := selector.TestAt(int(i), trMod, nil); enabled {
				// Extract external columns
				extractExternColumns(int(i), call, trMod, inputs, outputs)
				// Execute function call to produce outputs
				errs := executeAndCheck(call.Callee, fn.Name(), inputs, outputs, executor)
				errors = append(errors, errs...)
			} else if err != nil {
				errors = append(errors, err)
			}
		}
	} else {
		// Invoke each user-defined instance in turn
		for i := range height {
			// Extract external columns
			extractExternColumns(int(i), call, trMod, inputs, outputs)
			// Execute function call to produce outputs
			errs := executeAndCheck(call.Callee, fn.Name(), inputs, outputs, executor)
			errors = append(errors, errs...)
		}
	}
	//
	return errors
}

func executeAndCheck[T io.Instruction[T]](fid uint, name string, inputs, outputs []big.Int,
	executor *io.Executor[T]) []error {
	var (
		errors []error
		// Execute function call to produce actual outputs
		actual = executor.Read(fid, inputs)
	)
	// Sanity actual outputs match expected outputs
	for i := range len(outputs) {
		given := outputs[i]
		computed := actual[i]
		// Check input value
		if given.Cmp(&computed) != 0 {
			ins := toArgumentString(inputs)
			outs := toArgumentString(outputs)
			acts := toArgumentString(actual)
			errors = append(errors, fmt.Errorf("inconsistent instance %s(%s)=%s in trace (expected %s(%s)=%s)",
				name, ins, outs, name, ins, acts))
		}
	}
	//
	return errors
}

func extractFunctionColumns(row uint, mod RawModule, inputs, outputs []big.Int) {
	var (
		numInputs  = uint(len(inputs))
		numOutputs = uint(len(outputs))
	)
	//
	for i := range numInputs {
		var (
			col   = mod.Columns[i]
			input big.Int
		)
		// Assign value
		input.SetBytes(col.Data.Get(row).Bytes())
		//
		inputs[i] = input
	}
	//
	for i := range numOutputs {
		var (
			col    = mod.Columns[i+numInputs]
			output big.Int
		)
		// Assign value
		output.SetBytes(col.Data.Get(row).Bytes())
		//
		outputs[i] = output
	}
}

func extractExternColumns[F field.Element[F]](row int, call mir.FunctionCall[F], mod trace.Module[F],
	inputs, outputs []big.Int) []error {
	//
	// Extract function arguments
	errs1 := extractExternTerms(row, call.Arguments, mod, inputs)
	// Extract function returns
	errs2 := extractExternTerms(row, call.Returns, mod, outputs)
	//
	return append(errs1, errs2...)
}

func extractExternTerms[F field.Element[F]](row int, terms []mir.Term[F], mod trace.Module[F], values []big.Int,
) []error {
	var errors []error
	//
	for i, arg := range terms {
		var (
			ith      big.Int
			val, err = arg.EvalAt(row, mod, nil)
		)
		ith.SetBytes(val.Bytes())
		values[i] = ith
		//
		errors = append(errors, err)
	}
	//
	return errors
}
func extractFunctionPadding(registers []schema.Register, inputs, outputs []big.Int) {
	var numInputs = len(inputs)
	//
	for i := range len(inputs) {
		inputs[i] = registers[i].Padding
	}

	for i := range len(outputs) {
		outputs[i] = registers[i+numInputs].Padding
	}
}

// ReadInstances simply traverses all internal states generated within the
// executor and converts them back into raw columns.
func readInstances[T io.Instruction[T]](heap *lt.WordHeap, p io.Program[T], executor *io.Executor[T],
) []lt.Module[word.BigEndian] {
	var (
		modules = make([]lt.Module[word.BigEndian], len(p.Functions()))
		builder = array.NewDynamicBuilder(heap)
	)
	//
	for i := range p.Functions() {
		fn := p.Function(uint(i))
		instances := executor.Instances(uint(i))
		modules[i] = readFunctionInstances(fn, instances, &builder)
	}
	//
	return modules
}

func readFunctionInstances[T io.Instruction[T]](fn io.Function[T], instances []io.FunctionInstance,
	builder array.Builder[word.BigEndian]) lt.Module[word.BigEndian] {
	var (
		registers = fn.Registers()
		columns   = make([]lt.Column[word.BigEndian], fn.NumInputs()+fn.NumOutputs())
	)
	//
	for i := range columns {
		data := readFunctionInputOutputs(uint(i), registers, instances, builder)
		//
		columns[i] = lt.Column[word.BigEndian]{
			Name: registers[i].Name,
			Data: data,
		}
	}
	//
	return lt.Module[word.BigEndian]{
		Name:    fn.Name(),
		Columns: columns,
	}
}

func readFunctionInputOutputs(arg uint, registers []io.Register, instances []io.FunctionInstance,
	builder array.Builder[word.BigEndian]) array.MutArray[word.BigEndian] {
	var (
		height = uint(len(instances))
		arr    = builder.NewArray(height, registers[arg].Width)
	)
	//
	for i, instance := range instances {
		var (
			ith big.Int = instance.Get(arg)
			w   word.BigEndian
		)
		//
		arr = arr.Set(uint(i), w.SetBytes(ith.Bytes()))
	}
	//
	return arr
}

func toArgumentString(args []big.Int) string {
	var builder strings.Builder
	//
	for i, arg := range args {
		if i != 0 {
			builder.WriteString(", ")
		}

		builder.WriteString(arg.String())
	}
	//
	return builder.String()
}

// The purpose of the lt adaptor is to make an lt.TraceFile look like a Trace.
// In general, this is not safe.  However, we use this once we already know that
// the trace has been aligned.  Also, it is only used in a specific context.
type ltModuleAdaptor[F field.Element[F]] struct {
	module lt.Module[word.BigEndian]
}

func (p *ltModuleAdaptor[F]) Name() string {
	return p.module.Name
}

func (p *ltModuleAdaptor[F]) Width() uint {
	return uint(len(p.module.Columns))
}

func (p *ltModuleAdaptor[F]) Height() uint {
	return p.module.Height()
}

func (p *ltModuleAdaptor[F]) Column(cid uint) trace.Column[F] {
	return &ltColumnAdaptor[F]{p.module.Columns[cid]}
}

func (p *ltModuleAdaptor[F]) ColumnOf(col string) trace.Column[F] {
	panic("unsupported operation")
}

type ltColumnAdaptor[F field.Element[F]] struct {
	column lt.Column[word.BigEndian]
}

func (p *ltColumnAdaptor[F]) Name() string {
	return p.column.Name
}

func (p *ltColumnAdaptor[F]) Get(row int) F {
	var (
		v = p.column.Data.Get(uint(row))
		w F
	)
	// Convert
	return w.SetBytes(v.Bytes())
}

func (p *ltColumnAdaptor[F]) Data() array.Array[F] {
	panic("unsupported operation")
}

func (p *ltColumnAdaptor[F]) Padding() F {
	panic("unsupported operation")
}
