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
package cmd

import (
	"fmt"

	"github.com/consensys/go-corset/pkg/binfile"
	"github.com/consensys/go-corset/pkg/ir"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/field/bls12_377"
)

// SchemaStack is an abstraction for working with a stack of schemas, where each
// is layer is a refinement of its parent.  For example, the micro assembly
// layer is a refinement of the macro assembly layer.  Likewise, the Arithmetic
// Intermediate Representation is a refinement of the Mid-level Intermediate
// Representation, etc.
type SchemaStack[F field.Element[F]] struct {
	// Binfile represents the top of this stack.
	binfile binfile.BinaryFile
	// The various (abstract) layers which are refined from the binfile.
	abstractSchemas []schema.AnySchema[bls12_377.Element]
	// The various (concrete) layers which are refined from the abstract layers.
	concreteSchemas []schema.AnySchema[F]
	// Register mapping used
	mapping schema.LimbsMap
	// Name of IR used for corresponding schema
	names []string
	// Configuration for trace expansion
	traceBuilder ir.TraceBuilder[F]
}

// AbstractSchemas returns the stack of abstract schemas according to the
// selected layers, where higher-level layers come first.
func (p *SchemaStack[F]) AbstractSchemas() []schema.AnySchema[bls12_377.Element] {
	return p.abstractSchemas
}

// BinaryFile returns the binary file representing the top of this stack.
func (p *SchemaStack[F]) BinaryFile() *binfile.BinaryFile {
	return &p.binfile
}

// ConcreteSchemas returns the stack of concrete schemas according to the selected
// layers, where higher-level layers come first.
func (p *SchemaStack[F]) ConcreteSchemas() []schema.AnySchema[F] {
	return p.concreteSchemas
}

// ConcreteSchemaOf returns the schema associated with the given IR representation.  If
// there is no match, this will panic.
func (p *SchemaStack[F]) ConcreteSchemaOf(ir string) schema.AnySchema[F] {
	m := len(p.abstractSchemas)
	//
	for i, n := range p.names[m:] {
		if n == ir {
			return p.concreteSchemas[i]
		}
	}
	//
	panic(fmt.Sprintf("schema for %s not found", ir))
}

// HasUniqueSchema determines whether or not we have exactly one schema.
func (p *SchemaStack[F]) HasUniqueSchema() bool {
	return len(p.concreteSchemas) == 1
}

// RegisterMapping returns the register mapping used to split registers
// according to the given field configuration.
func (p *SchemaStack[F]) RegisterMapping() schema.LimbsMap {
	return p.mapping
}

// UniqueConcreteSchema returns the first schema on the stack which, when
// HasUniqueSchema() holds, means the uniquely specified schema.
func (p *SchemaStack[F]) UniqueConcreteSchema() schema.AnySchema[F] {
	return p.concreteSchemas[0]
}

// LowestConcreteSchema returns the last (i.e. lowest) schema on the stack.
func (p *SchemaStack[F]) LowestConcreteSchema() schema.AnySchema[F] {
	n := len(p.concreteSchemas) - 1
	return p.concreteSchemas[n]
}

// ConcreteIrName returns a human-readable anacronym of the IR used to generate the
// corresponding SCHEMA.
func (p *SchemaStack[F]) ConcreteIrName(index uint) string {
	return p.names[len(p.abstractSchemas)+int(index)]
}

// TraceBuilder returns a configured trace builder.
func (p *SchemaStack[F]) TraceBuilder() ir.TraceBuilder[F] {
	return p.traceBuilder
}
