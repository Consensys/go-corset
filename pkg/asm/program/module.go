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
package program

import (
	"github.com/consensys/go-corset/pkg/asm/io"
	"github.com/consensys/go-corset/pkg/schema"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/util/collection/iter"
	"github.com/consensys/go-corset/pkg/util/field"
)

// Module programs a wrapper around an io.Function which makes it look like a
// schema.Module.
type Module[F field.Element[F], T io.Instruction[T]] struct {
	id       sc.ModuleId
	function io.Function[T]
}

// NewModule constructs a new wrapper around a given io.Function instance.
func NewModule[F field.Element[F], T io.Instruction[T]](id sc.ModuleId, function io.Function[T]) *Module[F, T] {
	return &Module[F, T]{id, function}
}

// Function returns the underlying function being wrapped.
func (p *Module[F, T]) Function() io.Function[T] {
	return p.function
}

// ============================================================================
// schema.RegisterMap
// ============================================================================

// Name implementation for schema.RegisterMap interface.
func (p *Module[F, T]) Name() string {
	return p.function.Name()
}

// HasRegister implementation for schema.RegisterMap interface.
func (p *Module[F, T]) HasRegister(name string) (schema.RegisterId, bool) {
	return p.function.HasRegister(name)
}

// Register implementation for schema.RegisterMap interface.
func (p *Module[F, T]) Register(id schema.RegisterId) schema.Register {
	return p.function.Register(id)
}

// Registers implementation for schema.RegisterMap interface.
func (p *Module[F, T]) Registers() []schema.Register {
	return p.function.Registers()
}

// ============================================================================
// schema.Module
// ============================================================================

// Assignments implementation for schema.Module interface.
func (p *Module[F, T]) Assignments() iter.Iterator[schema.Assignment[F]] {
	panic("unsupported operation")
}

// AllowPadding implementation for schema.Module interface.
func (p *Module[F, T]) AllowPadding() bool {
	return false
}

// Constraints implementation for schema.Module interface.
func (p *Module[F, T]) Constraints() iter.Iterator[schema.Constraint[F]] {
	panic("unsupported operation")
}

// Consistent implementation for schema.Module interface.
func (p *Module[F, T]) Consistent(fieldWidth uint, schema schema.AnySchema[F]) []error {
	return nil
}

// LengthMultiplier implementation for schema.Module interface.
func (p *Module[F, T]) LengthMultiplier() uint {
	return 1
}

// IsPublic implementation for schema.Module interface.
func (p *Module[F, T]) IsPublic() bool {
	return p.function.IsPublic()
}

// IsSynthetic implementation for schema.Module interface.
func (p *Module[F, T]) IsSynthetic() bool {
	return false
}

// Substitute any matchined labelled constants within this module
func (p *Module[F, T]) Substitute(mapping map[string]F) {
	// For now, this is a no-operation because assembly has no concept of
	// labelled constants.  In the future, we might expect this to change.
}

// Width implementation for schema.Module interface.
func (p *Module[F, T]) Width() uint {
	return uint(len(p.function.Registers()))
}
