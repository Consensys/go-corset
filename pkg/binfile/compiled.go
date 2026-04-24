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
package binfile

import (
	"bytes"
	"encoding/gob"

	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/module"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/word"
)

// CompiledSchema represents a schema compiled for a given field configuration.
// If no field configuration is provided, then the target field is
// word.BigEndian (i.e. this is an abstract schema).
type CompiledSchema struct {
	// Config idenfies the field configuration to which this schema is compiled.
	// This is needed to ensure the compiled schema is deserialised under the
	// correct field.
	Config util.Option[field.Config]
	// Name identifies the Intermediate Representation to which this schema is
	// compiled.  It is really only for debugging purposes.
	Name string
	// Mapping identies how registers in the top-level schema are mapped into
	// one (or more) limbs in this schema.
	Mapping module.LimbsMap
	// Databytes of the compiled schema.
	Bytes []byte
}

// NewAbstractSchema constructs a new abstract schema with a given handle.  Such
// a schema is field agnostic.
func NewAbstractSchema(name string, schema schema.AnySchema[word.BigEndian]) CompiledSchema {
	var (
		buffer     bytes.Buffer
		gobEncoder *gob.Encoder = gob.NewEncoder(&buffer)
	)
	// Encode (optional) compiled schema
	if err := gobEncoder.Encode(&schema); err != nil {
		panic(err)
	}
	// Done
	return CompiledSchema{util.None[field.Config](), name, nil, buffer.Bytes()}
}

// NewConcreteSchema constructs a new (concrete) schema with a given handle.
// Such a schema is suitable only for the given field configuration.
func NewConcreteSchema[F any](config field.Config, name string, mapping module.LimbsMap, schema schema.AnySchema[F],
) CompiledSchema {
	var (
		buffer     bytes.Buffer
		gobEncoder *gob.Encoder = gob.NewEncoder(&buffer)
	)
	// Encode (optional) compiled schema
	if err := gobEncoder.Encode(&schema); err != nil {
		panic(err)
	}
	// Done
	return CompiledSchema{util.Some(config), name, mapping, buffer.Bytes()}
}

// ExtractSchema extracts an actual schema from a compiled schema.
func ExtractSchema[F any](s CompiledSchema) schema.AnySchema[F] {
	var (
		err error
		//
		buffer   = bytes.NewBuffer(s.Bytes)
		decoder  = gob.NewDecoder(buffer)
		compiled schema.AnySchema[F]
	)
	// Sanity check
	if err = decoder.Decode(&compiled); err != nil {
		panic(err)
	}
	//
	return compiled
}
