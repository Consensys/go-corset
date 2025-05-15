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
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"os"
	"strings"

	"github.com/consensys/go-corset/pkg/asm/io"
	"github.com/consensys/go-corset/pkg/util"
)

type tracesMap map[string]traceMap
type traceMap map[string][]big.Int

// ReadBatchedTraceFile reads a file containing zero or more traces expressed as JSON, where
// each trace is on a separate line.
func ReadBatchedTraceFile(filename string, program MacroProgram) []MacroTrace {
	var (
		lines  = util.ReadInputFile(filename)
		traces []MacroTrace
	)
	// Read constraints line by line
	for i, line := range lines {
		// Parse input line as JSON
		if line != "" && !strings.HasPrefix(line, ";;") {
			tr, err := ReadTrace([]byte(line), program)
			if err != nil {
				msg := fmt.Sprintf("%s:%d: %s", filename, i+1, err)
				panic(msg)
			}
			//
			traces = append(traces, tr)
		}
	}

	return traces
}

// ReadTraceFile reads a file containing a single trace.
func ReadTraceFile(filename string, program MacroProgram) (MacroTrace, error) {
	// Read data file
	bytes, err := os.ReadFile(filename)
	// Check for errors
	if err != nil {
		return nil, err
	}
	//
	return ReadTrace(bytes, program)
}

// ReadTrace reads a given set of function instances from JSON-encoded byte
// sequence.
func ReadTrace(bytes []byte, program MacroProgram) (MacroTrace, error) {
	var (
		traces tracesMap
		fnMap  map[string]uint = make(map[string]uint, 0)
		trace  MacroTrace      = io.NewTrace(program)
	)
	// Unmarshall
	jsonErr := json.Unmarshal(bytes, &traces)
	if jsonErr != nil {
		return trace, jsonErr
	}
	// Build function map
	for i, fn := range program.Functions() {
		fnMap[fn.Name()] = uint(i)
	}
	// Read trace instances
	for f, tr := range traces {
		fid, ok := fnMap[f]
		if !ok {
			return nil, fmt.Errorf("unknown function %s in trace", f)
		}
		// Read instances
		fnInsts, err := readTraceInstances(tr, fid, program.Function(fid))
		if err != nil {
			return trace, err
		}
		// Insert all
		trace.InsertAll(fnInsts)
	}
	//
	return trace, nil
}

func readTraceInstances[T any](trace traceMap, fid uint, fn io.Function[T]) ([]io.FunctionInstance, error) {
	var (
		height uint = math.MaxUint
		count       = 0
	)
	// Initialise register map
	for _, reg := range fn.Registers() {
		is_ioreg := (reg.IsInput() || reg.IsOutput())
		//
		if _, ok := trace[reg.Name]; !ok && is_ioreg {
			return nil, fmt.Errorf("missing register from trace: %s", reg.Name)
		} else if is_ioreg {
			count++
		}
	}
	// Sanity check no extra registers.
	if len(trace) != count {
		return nil, fmt.Errorf("too many registers in trace (was %d expected %d)", len(trace), count)
	}
	// Sanity check register heights
	for k, vs := range trace {
		n := uint(len(vs))
		if height == math.MaxUint {
			height = n
		} else if height != n {
			return nil, fmt.Errorf("invalid register height: %s", k)
		}
	}
	//
	instances := make([]io.FunctionInstance, height)
	// Parse the trace
	for i := uint(0); i < height; i++ {
		// Initialise ith function instance
		var instance io.FunctionInstance
		//
		instance.Function = fid
		instance.Inputs = make(map[string]big.Int)
		instance.Outputs = make(map[string]big.Int)

		for _, reg := range fn.Registers() {
			is_ioreg := (reg.IsInput() || reg.IsOutput())
			// Only consider input / output registers
			if is_ioreg {
				v := trace[reg.Name][i]
				// Check bitwidth
				if v.Cmp(reg.Bound()) >= 0 {
					return nil, fmt.Errorf("value %s out-of-bounds for %dbit register %s", v.String(), reg.Width, reg.Name)
				}
				// Assign as input or output
				if reg.IsInput() {
					instance.Inputs[reg.Name] = v
				} else {
					instance.Outputs[reg.Name] = v
				}
			}
		}
		// Assign ith instance
		instances[i] = instance
	}
	//
	return instances, nil
}
