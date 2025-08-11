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
	"math"
	"math/big"

	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/word"
)

// TraceValidation validates that values held in trace columns match the
// expected type.  This is really a sanity check that the trace is not
// malformed.
func TraceValidation(parallel bool, schema sc.AnySchema, tr tr.Trace[word.BigEndian]) []error {
	var (
		errors []error
		// Start timer
		stats = util.NewPerfStats()
	)
	// Validate expanded trace
	if parallel {
		// Run (parallel) trace validation
		errors = ParallelTraceValidation(schema, tr)
	} else {
		// Run (sequential) trace validation
		errors = SequentialTraceValidation(schema, tr)
	}
	// Log stats
	stats.Log("Trace validation")
	//
	return errors
}

// SequentialTraceValidation validates that values held in trace columns match
// the expected type.  This is really a sanity check that the trace is not
// malformed.
func SequentialTraceValidation[W word.Word[W]](schema sc.AnySchema, tr trace.Trace[W]) []error {
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

// ParallelTraceValidation validates that values held in trace columns match the
// expected type.  This is really a sanity check that the trace is not
// malformed.
func ParallelTraceValidation[W word.Word[W]](schema sc.AnySchema, trace tr.Trace[W]) []error {
	var (
		errors []error
		// Construct a communication channel for errors.
		c = make(chan error, tr.NumberOfColumns(trace))
		// Number of columns to validate
		ntodo = uint(0)
	)
	// Check each module in turn
	for mid := uint(0); mid < trace.Width(); mid++ {
		var (
			scMod = schema.Module(mid)
			trMod = trace.Module(mid)
		)
		// Check each column within each module
		for i := uint(0); i < trMod.Width(); i++ {
			rid := sc.NewRegisterId(i)
			// Check elements
			go func(reg sc.Register, data tr.Column[W]) {
				// Send outcome back
				c <- validateColumnBitWidth(reg.Width, data, scMod)
			}(scMod.Register(rid), trMod.Column(i))
			//
			ntodo++
		}
	}
	// Collect up all the results
	for i := uint(0); i < ntodo; i++ {
		// Read from channel
		if e := <-c; e != nil {
			errors = append(errors, e)
		}
	}
	// Done
	return errors
}

func sequentialModuleValidation[W word.Word[W]](scMod sc.Module, trMod trace.Module[W]) []error {
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
					rid  = sc.NewRegisterId(i)
					reg  = scMod.Register(rid)
					data = trMod.Column(i)
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

// Validate that all elements of a given column fit within a given bitwidth.
func validateColumnBitWidth[W word.Word[W]](bitwidth uint, col tr.Column[W], mod sc.Module) error {
	// Sanity check bitwidth can be checked.
	if bitwidth == math.MaxUint {
		// This indicates a column which has no fixed bitwidth but, rather, uses
		// the entire field element.  The only situation this arises in practice
		// is for columns holding the multiplicative inverse of some other
		// column.
		return nil
	} else if col.Data() == nil {
		panic(fmt.Sprintf("column %s is unassigned", col.Name()))
	}
	//
	for j := 0; j < int(col.Data().Len()); j++ {
		var jth big.Int

		jth.SetBytes(col.Get(j).Bytes())
		//
		if uint(jth.BitLen()) > bitwidth {
			qualColName := trace.QualifiedColumnName(mod.Name(), col.Name())
			return fmt.Errorf("row %d of column %s is out-of-bounds (%s)", j, qualColName, jth.String())
		}
	}
	// success
	return nil
}
