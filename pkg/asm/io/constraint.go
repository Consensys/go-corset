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
package io

import (
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/bit"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// Constraint represents a wrapper around an instruction in order for it to
// conform to the constraint interface.
type Constraint[T Instruction[T]] Function[T]

// Accepts implementation for schema.Constraint interface.
func (p Constraint[T]) Accepts(trace.Trace) (bit.Set, sc.Failure) {
	panic("todo")
}

// Bounds implementation for schema.Constraint interface.
func (p Constraint[T]) Bounds(module uint) util.Bounds {
	panic("todo")
}

// Consistent implementation for schema.Constraint interface.
func (p Constraint[T]) Consistent(sc.AnySchema) []error {
	panic("todo")
}

// Contexts implementation for schema.Constraint interface.
func (p Constraint[T]) Contexts() []sc.ModuleId {
	panic("todo")
}

// Name implementation for schema.Constraint interface.
func (p Constraint[T]) Name() string {
	panic("todo")
}

// Lisp implementation for schema.Constraint interface.
func (p Constraint[T]) Lisp(schema sc.AnySchema) sexp.SExp {
	//
	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("function"),
		sexp.NewSymbol(p.name),
	})
}

// Substitute implementation for schema.Constraint interface.
func (p Constraint[T]) Substitute(map[string]fr.Element) {
	// Do nothing since assembly instructions do not (at the time of writing)
	// employ labelled constants.
}
