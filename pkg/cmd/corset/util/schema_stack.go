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
package util

import (
	"github.com/consensys/go-corset/pkg/asm"
	"github.com/consensys/go-corset/pkg/binfile"
	"github.com/consensys/go-corset/pkg/ir"
	"github.com/consensys/go-corset/pkg/schema/module"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/word"
)

// SchemaStack is an abstraction for working with a stack of schemas, where each
// is layer is a refinement of its parent.  For example, the micro assembly
// layer is a refinement of the macro assembly layer.  Likewise, the Arithmetic
// Intermediate Representation is a refinement of the Mid-level Intermediate
// Representation, etc.
type SchemaStack[F field.Element[F]] struct {
	header     binfile.Header
	attributes []binfile.Attribute
	compiled   []binfile.CompiledSchema
	mapping    module.LimbsMap
	// Configuration for trace expansion
	traceBuilder ir.TraceBuilder[F]
}

// Attributes returns the set of attributes for the binary file being generated.
func (p *SchemaStack[F]) Attributes() []binfile.Attribute {
	return p.attributes
}

// BinaryFile constructs a suitable binary file from this schema.
func (p *SchemaStack[F]) BinaryFile() *binfile.BinaryFile {
	var (
		root     = p.AbstractSchema()
		compiled util.Option[binfile.CompiledSchema]
	)
	//
	if len(p.compiled) > 1 {
		compiled = util.Some[binfile.CompiledSchema](p.ConcreteSchema())
	}
	//
	return binfile.NewBinaryFile(p.header.MetaData, p.attributes, root, compiled)
}

// Header returns a suitable header for the binary file to be generated
func (p *SchemaStack[F]) Header() binfile.Header {
	return p.header
}

// AbstractSchema returns the top-level (i.e. most abstract) schema for system of
// constraints being described.
func (p *SchemaStack[F]) AbstractSchema() asm.MacroHirProgram {
	var (
		schema = binfile.ExtractSchema[word.BigEndian](p.compiled[0])
		// Should always be safe as schema stacker enforces invariant that
		// top-level program comes first.
		program = schema.(*asm.MacroHirProgram)
	)
	//
	return *program
}

// HasConcreteSchema determines whether or not a schema was requested which
// operates over the concrete field F.
func (p *SchemaStack[F]) HasConcreteSchema() bool {
	for _, c := range p.compiled {
		if c.Name == "MIR" || c.Name == "AIR" {
			return true
		}
	}
	//
	return false
}

// ConcreteMapping returns the mapping of registers at the abstract level down
// to those at the concrete level.  This only makes sense when
// HasConcreteSchema() holds.
func (p *SchemaStack[F]) ConcreteMapping() module.LimbsMap {
	return p.mapping
}

// FindSchema returns the compiled schema corresponding to the given name.
func (p *SchemaStack[F]) FindSchema(name string) binfile.CompiledSchema {
	for _, c := range p.compiled {
		if c.Name == name {
			return c
		}
	}
	//
	panic("schema not found")
}

// ConcreteSchema returns the most concrete schema within this stack.
func (p *SchemaStack[F]) ConcreteSchema() binfile.CompiledSchema {
	n := len(p.compiled)
	//
	return p.compiled[n-1]
}

// TraceBuilder returns a configured trace builder.
func (p *SchemaStack[F]) TraceBuilder() ir.TraceBuilder[F] {
	return p.traceBuilder
}
