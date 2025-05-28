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
package schema

import (
	"fmt"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util/field"
)

// TraceBuilder provides a mechanical means of constructing a trace from a given
// schema and set of input columns.  The goal is to encapsulate all of the logic
// around building a trace.
type TraceBuilder struct {
	// Schema to be used when building the trace
	schema AnySchema
	// Indicates whether or not to perform defensive padding.  This is where
	// padding rows are appended and/or prepended to ensure no constraint in the
	// active region of the trace is clipped.  Whilst not strictly necessary,
	// this can be helpful for identifying invalid constraints which are only
	// exposed with a given amount of padding.
	defensive bool
	// Indicates whether or not to perform trace expansion.  The default should
	// be to apply trace expansion.  However, for testing purposes, it can be
	// useful to provide an already expanded trace to ensure a set of
	// constraints correctly rejects it.
	expand bool
	// Indicates whether or not to validate all column types.  That is, check
	// that the values supplied for all columns (both input and computed) are
	// within their declared type.
	validate bool
	// Determines the amount of padding to apply to each module in the trace.
	// At the moment, this is applied uniformly across all modules.  This is
	// somewhat cumbersome, and it would make sense to support different
	// protocols.  For example, one obvious protocol is to expand a module's
	// length upto a power-of-two.
	padding uint
}

// A column key is used as a key for the column map
type columnKey struct {
	module string
	column string
}

type columnId struct {
	module uint
	column uint
}

// NewTraceBuilder constructs a default trace builder.  The idea is that this
// could then be customized as needed following the builder pattern.
func NewTraceBuilder[C Constraint](schema Schema[C]) TraceBuilder {
	return TraceBuilder{Any(schema), true, true, true, 0}
}

// Defensive updates a given builder configuration to apply defensive padding
// (or not).
func (tb TraceBuilder) Defensive(flag bool) TraceBuilder {
	ntb := tb
	ntb.defensive = flag
	//
	return ntb
}

// Expand updates a given builder configuration to perform trace expansion (or
// not).
func (tb TraceBuilder) Expand(flag bool) TraceBuilder {
	ntb := tb
	ntb.expand = flag
	//
	return ntb
}

// Validate updates a given builder configuration to perform trace validation (or
// not).
func (tb TraceBuilder) Validate(flag bool) TraceBuilder {
	ntb := tb
	ntb.validate = flag
	//
	return ntb
}

// Padding updates a given builder configuration to use a given amount of padding
func (tb TraceBuilder) Padding(padding uint) TraceBuilder {
	ntb := tb
	ntb.padding = padding
	//
	return ntb
}

// Build attempts to construct a trace for a given schema, producing errors if
// there are inconsistencies (e.g. missing columns, duplicate columns, etc).
func (tb TraceBuilder) Build(cols []trace.RawColumn) (trace.Trace, []error) {
	tr, errors := initialiseTrace(tb.schema, cols)
	//
	if len(errors) > 0 {
		// Critical failure
		return nil, errors
	}
	// FIXME: this is where we need to apply trace expansion.
	//
	// else if tb.expand {
	//
	// }
	// Padding
	if tb.padding > 0 {
		padTraceColumns(tr, tb.padding)
	}
	//
	return tr, errors
}

func initialiseTrace(schema AnySchema, cols []trace.RawColumn) (*trace.ArrayTrace, []error) {
	var (
		// Initialise modules
		modmap  = initialiseModuleMap(schema)
		modules = make([]trace.ArrayModule, schema.Width())
	)
	//
	columns, errors := splitTraceColumns(schema, modmap, cols)
	//
	for i := uint(0); i != schema.Width(); i++ {
		var (
			errs []error
			name = schema.Module(i).Name()
		)

		modules[i], errs = fillTraceModule(i, name, columns[i])
		errors = append(errors, errs...)
	}
	// Done
	return trace.NewArrayTrace(modules), errors
}

func initialiseModuleMap(schema AnySchema) map[string]uint {
	modmap := make(map[string]uint, 100)
	// Initialise modules
	for i := uint(0); i != schema.Width(); i++ {
		m := schema.Module(i)
		// Sanity check module
		if _, ok := modmap[m.Name()]; ok {
			panic(fmt.Sprintf("duplicate module '%s' in schema", m.Name()))
		}

		modmap[m.Name()] = i
	}
	// Done
	return modmap
}

func splitTraceColumns(schema AnySchema, modmap map[string]uint,
	cols []trace.RawColumn) ([][]trace.RawColumn, []error) {
	//
	var (
		// Errs contains the set of filling errors which are accumulated
		errs []error
		//
		seen map[columnKey]bool = make(map[columnKey]bool, 0)
	)
	//
	colmap, modules := initialiseColumnMap(schema)
	// Assign data from each input column given
	for _, c := range cols {
		// Lookup the module
		if _, ok := modmap[c.Module]; !ok {
			errs = append(errs, fmt.Errorf("unknown module '%s' in trace", c.Module))
		} else {
			key := columnKey{c.Module, c.Name}
			// Determine enclosiong module height
			cid, ok := colmap[key]
			// More sanity checks
			if !ok {
				errs = append(errs, fmt.Errorf("unknown column '%s' in trace", c.QualifiedName()))
			} else if _, ok := seen[key]; ok {
				errs = append(errs, fmt.Errorf("duplicate column '%s' in trace", c.QualifiedName()))
			} else {
				seen[key] = true
				modules[cid.module][cid.column] = c
			}
		}
	}
	//
	return modules, errs
}

func initialiseColumnMap(schema AnySchema) (map[columnKey]columnId, [][]trace.RawColumn) {
	var (
		colmap  = make(map[columnKey]columnId, 100)
		modules = make([][]trace.RawColumn, schema.Width())
	)
	// Initialise modules
	for i := uint(0); i != schema.Width(); i++ {
		m := schema.Module(i)
		columns := make([]trace.RawColumn, m.Width())
		//
		for j := uint(0); j != m.Width(); j++ {
			col := m.Register(j)
			key := columnKey{m.Name(), col.Name}
			id := columnId{i, j}
			//
			if _, ok := colmap[key]; ok {
				panic(fmt.Sprintf("duplicate column '%s' in schema", trace.QualifiedColumnName(m.Name(), col.Name)))
			}
			//
			colmap[key] = id
			columns[i] = trace.RawColumn{Module: m.Name(), Name: col.Name, Data: nil}
		}
		// Initialise empty columns for this module.
		modules[i] = columns
	}
	// Done
	return colmap, modules
}

func fillTraceModule(mid uint, name string, rawColumns []trace.RawColumn) (trace.ArrayModule, []error) {
	var (
		traceColumns = make([]trace.ArrayColumn, len(rawColumns))
		zero         = fr.NewElement(0)
		errors       []error
	)
	//
	for i := range traceColumns {
		ith := rawColumns[i]
		ctx := trace.NewContext(mid, 1)
		data := ith.Data
		//
		if data == nil {
			err := fmt.Errorf("missing input column '%s.%s' in trace", ith.Name, ith.Name)
			errors = append(errors, err)
			// Fill with a column of height zero.
			data = field.NewFrArray(0, 256)
		}
		//
		traceColumns[i] = trace.NewArrayColumn(ctx, ith.Name, data, zero)
	}
	//
	return trace.NewArrayModule(name, traceColumns), errors
}

// PadColumns pads every column in a given trace with a given amount of (front)
// padding. Observe that this applies on top of any spillage and/or defensive
// padding already applied.
func padTraceColumns(tr *trace.ArrayTrace, padding uint) {
	n := tr.Modules().Count()
	// Iterate over modules
	for i := uint(0); i < n; i++ {
		tr.Pad(i, padding, 0)
	}
}
