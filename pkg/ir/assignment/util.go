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
	sc "github.com/consensys/go-corset/pkg/schema"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util/collection/array"
	"github.com/consensys/go-corset/pkg/util/field"
)

// ReadRegisters a given set of registers from a trace.
func ReadRegisters[F field.Element[F]](trace tr.Trace[F], regs ...sc.RegisterRef) []array.Array[F] {
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

// WriteRegisters a given set of registers from a trace.
func WriteRegisters[F field.Element[F], T array.Array[F]](schema sc.AnySchema, targets []sc.RegisterRef, data []T) []tr.ArrayColumn[F] {
	var (
		columns = make([]tr.ArrayColumn[F], len(targets))
		zero    F
	)
	// Write outputs
	for i, ref := range targets {
		ith := schema.Register(ref)
		columns[i] = tr.NewArrayColumn(ith.Name, data[i], zero)
	}
	//
	return columns
}

func toRegisterRefs(context sc.ModuleId, ids []sc.RegisterId) []sc.RegisterRef {
	var refs = make([]sc.RegisterRef, len(ids))
	//
	for i, id := range ids {
		refs[i] = sc.NewRegisterRef(context, id)
	}
	//
	return refs
}
