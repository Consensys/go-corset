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
package builder

import (
	"fmt"

	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	tr "github.com/consensys/go-corset/pkg/trace"
)

// SequentialTraceExpansion expands a given trace according to a given schema.
// More specifically, that means computing the actual values for any
// assignments.  This is done using a straightforward sequential algorithm.
func SequentialTraceExpansion(schema sc.AnySchema, trace *trace.ArrayTrace) error {
	var (
		err      error
		expander = NewExpander(schema.Width(), schema.Assignments())
	)
	// Compute each assignment in turn
	for !expander.Done() {
		var cols []tr.ArrayColumn
		// Get next assignment
		ith := expander.Next(1)[0]
		// Compute ith assignment(s)
		if cols, err = ith.Compute(trace, schema); err != nil {
			return err
		}
		// Fill all computed columns
		fillComputedColumns(ith.RegistersWritten(), cols, trace)
	}
	// Done
	return nil
}

// SequentialTraceValidation validates that values held in trace columns match
// the expected type.  This is really a sanity check that the trace is not
// malformed.
func SequentialTraceValidation(schema sc.AnySchema, tr trace.Trace) []error {
	var errors []error
	//
	for i := uint(0); i < max(schema.Width(), tr.Width()); i++ {
		// Sanity checks first
		if i >= tr.Width() {
			err := fmt.Errorf("module %s missing from trace", schema.Module(i).Name())
			errors = append(errors, err)
		} else if i >= schema.Width() {
			err := fmt.Errorf("unknown module %s in trace", tr.Module(i).Name())
			errors = append(errors, err)
		} else {
			var (
				scMod = schema.Module(i)
				trMod = tr.Module(i)
			)
			// Validate module
			errors = append(errors, sequentialModuleValidation(scMod, trMod)...)
		}
	}
	// Done
	return errors
}

func sequentialModuleValidation(scMod sc.Module, trMod trace.Module) []error {
	var (
		errors []error
		// Extract module registers
		registers = scMod.Registers()
	)
	// Sanity check
	if scMod.Name() != trMod.Name() {
		err := fmt.Errorf("misaligned module during trace expansion (%s vs %s)", scMod.Name(), trMod.Name())
		errors = append(errors, err)
	} else {
		for i := uint(0); i < max(trMod.Width(), scMod.Width()); i++ {
			// Sanity checks first
			if i >= trMod.Width() {
				err := fmt.Errorf("register %s.%s missing from trace", trMod.Name(), registers[i].Name)
				errors = append(errors, err)
			} else if i >= scMod.Width() {
				err := fmt.Errorf("unknown register %s.%s in trace", trMod.Name(), trMod.Column(i).Name())
				errors = append(errors, err)
			} else {
				var (
					rid               = sc.NewRegisterId(i)
					reg  sc.Register  = scMod.Register(rid)
					data trace.Column = trMod.Column(i)
				)
				// Sanity check data has expected bitwidth
				if err := validateColumnBitWidth(reg.Width, data, scMod); err != nil {
					errors = append(errors, err)
				}
			}
		}
	}
	// Done
	return errors
}
