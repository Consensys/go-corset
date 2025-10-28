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
package assignment

import (
	"github.com/consensys/go-corset/pkg/schema/module"
	"github.com/consensys/go-corset/pkg/schema/register"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util/collection/array"
	"github.com/consensys/go-corset/pkg/util/field"
)

// ReadRegisterRefs reads the values for a given set of registers from a trace.
func ReadRegisterRefs[F field.Element[F]](trace tr.Trace[F], regs ...register.Refs) [][]array.Array[F] {
	var (
		targets = make([][]array.Array[F], len(regs))
	)
	// Read registers
	for i, ref := range regs {
		targets[i] = ReadRegisters(trace, ref.Module(), ref.Registers()...)
	}
	//
	return targets
}

// ReadRegisters reads the values for a given set of registers from a trace.
func ReadRegisters[F field.Element[F]](trace tr.Trace[F], mid module.Id, regs ...register.Id) []array.Array[F] {
	var (
		targets = make([]array.Array[F], len(regs))
	)
	// Read registers
	for i, rid := range regs {
		targets[i] = trace.Module(mid).Column(rid.Unwrap()).Data()
	}
	//
	return targets
}

// ReadRegistersRef reads the values for a given set of registers from a trace.
func ReadRegistersRef[F field.Element[F]](trace tr.Trace[F], regs ...register.Ref) []array.Array[F] {
	var (
		targets = make([]array.Array[F], len(regs))
	)
	// Read registers
	for i, ref := range regs {
		mid, rid := ref.Module(), ref.Register().Unwrap()
		targets[i] = trace.Module(mid).Column(rid).Data()
	}
	//
	return targets
}

// ReadPadding reads the padding values determined for a given set of registers in a trace.
func ReadPadding[F field.Element[F]](trace tr.Trace[F], regs ...register.Ref) []F {
	var (
		targets = make([]F, len(regs))
	)
	// Read registers
	for i, ref := range regs {
		mid, rid := ref.Module(), ref.Register().Unwrap()
		targets[i] = trace.Module(mid).Column(rid).Padding()
	}
	//
	return targets
}

// SubdivideRegisterRefs subdivides a set of register references according to a
// given mapping.
func SubdivideRegisterRefs[F field.Element[F]](mapping module.LimbsMap, refs ...register.Refs) []register.Refs {
	var (
		nrefs = make([]register.Refs, len(refs))
	)
	//
	for i, ref := range refs {
		nrefs[i] = ref.Apply(mapping.Module(ref.Module()))
	}
	//
	return nrefs
}
