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
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/consensys/go-corset/pkg/binfile"
	"github.com/consensys/go-corset/pkg/ir"
	"github.com/consensys/go-corset/pkg/schema"
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
	// Binfile represents the top of this stack.
	binfile util.Option[binfile.BinaryFile]
	// The various (abstract) layers which are refined from the binfile.
	abstractSchemas []schema.AnySchema[word.BigEndian]
	// The various (concrete) layers which are refined from the abstract layers.
	concreteSchemas []schema.AnySchema[F]
	// Register mapping used
	mapping module.LimbsMap
	// Name of IR used for corresponding schema
	names []string
	// Configuration for trace expansion
	traceBuilder ir.TraceBuilder[F]
}

// AbstractSchemas returns the stack of abstract schemas according to the
// selected layers, where higher-level layers come first.
func (p *SchemaStack[F]) AbstractSchemas() []schema.AnySchema[word.BigEndian] {
	return p.abstractSchemas
}

// BinaryFile returns the binary file representing the top of this stack.
func (p *SchemaStack[F]) BinaryFile() *binfile.BinaryFile {
	bf := p.binfile.Unwrap()
	return &bf
}

<<<<<<< HEAD:pkg/cmd/corset/util/schema_stack.go
// HasConcreteSchema returns true if there is at least one concrete schema..
func (p *SchemaStack[F]) HasConcreteSchema() bool {
	return len(p.concreteSchemas) > 0
}

// ConcreteSchema returns the stack of concrete schemas according to the selected
=======
// Clone this stack producing a physically disjoint but otherwise identical
// stack.  The purpose of this is to ensure not interference between runs.
func (p *SchemaStack[F]) Clone() SchemaStack[F] {
	var (
		binfile         util.Option[binfile.BinaryFile]
		abstractSchemas = cloneSchemas(p.abstractSchemas)
		concreteSchemas = cloneSchemas(p.concreteSchemas)
	)
	//
	if p.binfile.HasValue() {
		binfile = util.Some(p.binfile.Unwrap().Clone())
	}
	//
	return SchemaStack[F]{
		binfile,
		abstractSchemas,
		concreteSchemas,
		p.mapping,
		p.names,
		p.traceBuilder,
	}
}

// ConcreteSchemas returns the stack of concrete schemas according to the selected
>>>>>>> 4d1b4745 (support SchemaStack.Clone()):pkg/cmd/util/schema_stack.go
// layers, where higher-level layers come first.
func (p *SchemaStack[F]) ConcreteSchema() schema.AnySchema[F] {
	var n = len(p.concreteSchemas) - 1
	return p.concreteSchemas[n]
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

// RegisterMapping returns the register mapping used to split registers
// according to the given field configuration.
func (p *SchemaStack[F]) RegisterMapping() module.LimbsMap {
	return p.mapping
}

// ConcreteIrName returns a human-readable anacronym of the IR used to generate the
// corresponding SCHEMA.
func (p *SchemaStack[F]) ConcreteIrName() string {
	var n = len(p.concreteSchemas) - 1
	return p.names[len(p.abstractSchemas)+n]
}

// TraceBuilder returns a configured trace builder.
func (p *SchemaStack[F]) TraceBuilder() ir.TraceBuilder[F] {
	return p.traceBuilder
}

// Perform a deep copy of a schema by encoding it into bytes, and then decoding
// it from those bytes back into a fresh object.
func cloneSchemas[F field.Element[F]](schemas []schema.AnySchema[F]) []schema.AnySchema[F] {
	var nschemas = make([]schema.AnySchema[F], len(schemas))
	//
	for i, s := range schemas {
		nschemas[i] = decodeSchema[F](encodeSchema[F](s))
	}
	//
	return nschemas
}

func encodeSchema[F field.Element[F]](schema schema.AnySchema[F]) []byte {
	var (
		buffer     bytes.Buffer
		gobEncoder *gob.Encoder = gob.NewEncoder(&buffer)
	)
	// Encode schema
	if err := gobEncoder.Encode(&schema); err != nil {
		panic(err.Error())
	}
	// Done
	return buffer.Bytes()
}

func decodeSchema[F field.Element[F]](data []byte) (r schema.AnySchema[F]) {
	var (
		buffer = bytes.NewBuffer(data)
		// Looks good, proceed.
		decoder = gob.NewDecoder(buffer)
	)
	// Finally, decode the schema itself
	if err := decoder.Decode(&r); err != nil {
		panic(err.Error())
	}
	//
	return r
}
