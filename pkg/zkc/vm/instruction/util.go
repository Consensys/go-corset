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
package instruction

import (
	"fmt"
	"strings"

	"github.com/consensys/go-corset/pkg/schema/register"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/zkc/vm/word"
)

// RegistersToString returns a string representation for zero or more registers
// separated by a comma.
func registersToString(env register.Map, regs ...register.Id) string {
	var (
		builder strings.Builder
		n       = uint(len(env.Registers()))
	)
	//
	for i := 0; i < len(regs); i++ {
		var rid = regs[i]
		//
		if i != 0 {
			builder.WriteString(", ")
		}
		//
		if rid.Unwrap() < n {
			builder.WriteString(env.Register(rid).Name())
		} else {
			builder.WriteString("??")
		}
	}
	//
	return builder.String()
}

func expressionToString[W word.Word[W]](op string, regs []register.Id, constant W, env register.Map) string {
	var builder strings.Builder
	//
	for i := 0; i < len(regs); i++ {
		var rid = regs[i]
		//
		builder.WriteString(env.Register(rid).Name())
		builder.WriteString(" ")
		builder.WriteString(op)
		builder.WriteString(" ")
	}
	//
	builder.WriteString(fmt.Sprintf("0x%s", constant.Text(16)))
	//
	return builder.String()
}

// CheckTargetRegisters performs some simple checks on a set of target registers
// being written.  Firstly, they cannot be input registers (as this are always
// constant).  Secondly, we cannot write to the same register more than once
// (i.e. a conflicting write).
func checkTargetRegisters(config field.Config, regs register.Map, targets ...register.Id) []error {
	var errors []error
	//
	for i, target := range targets {
		ith := regs.Register(target)
		//
		if ith.IsInput() {
			errors = append(errors, fmt.Errorf("cannot write input register %s", ith.Name()))
		} else if ith.Width() > config.BandWidth {
			errors = append(errors, fmt.Errorf("register %s exceeds maximum width (u%d > u%d)",
				ith.Name(), ith.Width(), config.RegisterWidth))
		}
		//
		for j := i + 1; j < len(targets); j++ {
			if target == targets[j] {
				errors = append(errors, fmt.Errorf("conflicting write to register %s", ith.Name()))
			}
		}
	}
	// check targets fit within bandwidth
	if sumTargetBits(targets, regs) > config.BandWidth {
		errors = append(errors, fmt.Errorf("target registers exceed available bandwidth"))
	}
	//
	return errors
}

// Sum the total number of bits used by the given set of target registers.
func sumTargetBits(targets []register.Id, regs register.Map) uint {
	sum := uint(0)
	//
	for _, target := range targets {
		sum += regs.Register(target).Width()
	}
	//
	return sum
}
