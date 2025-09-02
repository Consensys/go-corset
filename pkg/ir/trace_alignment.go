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
package ir

import (
	"fmt"

	sc "github.com/consensys/go-corset/pkg/schema"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/trace/lt"
)

// AlignTrace performs "trace alignment" on a given trace file.  That is, it
// ensures: firstly, the order in which modules occur in the trace file matches
// (i.e. aligns with) those in the given schema; secondly, it ensures that the
// columns within each module match (i.e. align with) those of the corresponding
// schema module.  If any columns or modules are missing, then one or more
// errors will be reported.
//
// NOTE: alignment is impacted by whether or not the trace is being expanded or
// not. Specifically, expanding traces don't need to include data for computed
// columns, since these will be added during expansion.
func AlignTrace[F any](schema sc.AnySchema[F], modules []lt.Module[F], expanding bool,
) ([]lt.Module[F], []error) {
	var errors []error
	// First, align modules
	if modules, errors = alignModules(schema, modules, expanding); len(errors) > 0 {
		return nil, errors
	}
	// Second, align columns within modules
	for i := range schema.Width() {
		cols, errs := alignColumns(schema.Module(i), modules[i].Columns, expanding)
		errors = append(errors, errs...)
		modules[i].Columns = cols
	}
	// Done
	return modules, errors
}

func alignModules[F any](schema sc.AnySchema[F], modules []lt.Module[F], expanding bool) ([]lt.Module[F], []error) {
	//
	var (
		modmap = make(map[string]uint)
		nmods  = make([]lt.Module[F], max(schema.Width()))
		errs   []error
	)
	// Initialise module mapping
	for i := range schema.Width() {
		ith := schema.Module(i)
		nmods[i].Name = ith.Name()
		modmap[ith.Name()] = i
	}
	// Rearrange layout
	for _, m := range modules {
		if index, ok := modmap[m.Name]; ok {
			nmods[index] = m
		} else if expanding {
			errs = append(errs, fmt.Errorf("unknown module '%s' in trace", m.Name))
		}
	}
	//
	return nmods, errs
}

func alignColumns[F any](mod sc.Module[F], columns []lt.Column[F], expanding bool) ([]lt.Column[F], []error) {
	var (
		// Errs contains the set of filling errors which are accumulated
		errs []error
		//
		colmap = make(map[string]uint, mod.Width())
		seen   = make([]bool, mod.Width())
		//
		ncols = make([]lt.Column[F], mod.Width())
	)
	// Initialise column map
	for i := range mod.Width() {
		ith := mod.Register(sc.NewRegisterId(i))
		ncols[i].Name = ith.Name
		colmap[ith.Name] = i
	}
	// Assign data for each column given
	for _, col := range columns {
		// Determine enclosiong module height
		cid, ok := colmap[col.Name]
		// More sanity checks
		if !ok {
			errs = append(errs, fmt.Errorf("unknown column '%s' in trace", tr.QualifiedColumnName(mod.Name(), col.Name)))
		} else if ok := seen[cid]; ok {
			errs = append(errs, fmt.Errorf("duplicate column '%s' in trace", tr.QualifiedColumnName(mod.Name(), col.Name)))
		} else {
			seen[cid] = true
			ncols[cid] = col
		}
	}
	// Sanity check everything was assigned
	for i := range mod.Width() {
		var (
			reg = mod.Register(sc.NewRegisterId(i))
			col = ncols[i]
		)
		//
		if reg.IsInputOutput() && col.Data == nil {
			name := tr.QualifiedColumnName(mod.Name(), reg.Name)
			errs = append(errs, fmt.Errorf("missing input/output column '%s' from trace", name))
		} else if !expanding && col.Data == nil {
			name := tr.QualifiedColumnName(mod.Name(), reg.Name)
			errs = append(errs, fmt.Errorf("missing computed column '%s' from expanded trace", name))
		}
	}
	//
	return ncols, errs
}
