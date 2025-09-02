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
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace/lt"
	"github.com/consensys/go-corset/pkg/util/field"
)

// AlignTrace performs "trace alignment" on a given trace file.  That is, it
// ensures: firstly, the order in which modules occur in the trace file matches
// (i.e. aligns with) those in the given schema; secondly, it ensures that the
// columns within each module match (i.e. align with) those of the corresponding
// schema module.  If any columns or modules are missing, then one or more
// errors will be reported.
func AlignTrace[F field.Element[F]](schema sc.AnySchema[F], tf lt.TraceFile) []error {
	panic("todo")
}

// func alignRawTrace[F field.Element[F]](expanded bool, schema sc.AnySchema[F], modmap map[string]uint,
// 	rawTrace []lt.Module[F]) ([][]trace.RawColumn[F], []error) {
// 	//
// 	var (
// 		// Errs contains the set of filling errors which are accumulated
// 		errs []error
// 		//
// 		seen map[columnKey]bool = make(map[columnKey]bool, 0)
// 	)
// 	//
// 	colmap, modules := initialiseColumnMap(expanded, schema)
// 	// Assign data from each input column given
// 	for _, col := range cols {
// 		// Lookup the module
// 		if _, ok := modmap[col.Module]; !ok {
// 			errs = append(errs, fmt.Errorf("unknown module '%s' in trace", col.Module))
// 		} else {
// 			key := columnKey{col.Module, col.Name}
// 			// Determine enclosiong module height
// 			cid, ok := colmap[key]
// 			// More sanity checks
// 			if !ok {
// 				errs = append(errs, fmt.Errorf("unknown column '%s' in trace", col.QualifiedName()))
// 			} else if _, ok := seen[key]; ok {
// 				errs = append(errs, fmt.Errorf("duplicate column '%s' in trace", col.QualifiedName()))
// 			} else {
// 				seen[key] = true
// 				modules[cid.module][cid.column] = col
// 			}
// 		}
// 	}
// 	// Sanity check everything was assigned
// 	for i, m := range modules {
// 		mod := schema.Module(uint(i))
// 		//
// 		for j, c := range m {
// 			rid := sc.NewRegisterId(uint(j))
// 			reg := mod.Register(rid)
// 			//
// 			if reg.IsInputOutput() && c.Data == nil {
// 				errs = append(errs, fmt.Errorf("missing input/output column '%s' from trace", c.QualifiedName()))
// 			} else if expanded && c.Data == nil {
// 				errs = append(errs, fmt.Errorf("missing computed column '%s' from expanded trace", c.QualifiedName()))
// 			}
// 		}
// 	}
// 	//
// 	return modules, errs
// }

// func initialiseModuleMap[F any](schema sc.AnySchema[F]) map[string]uint {
// 	modmap := make(map[string]uint, 100)
// 	// Initialise modules
// 	for i := uint(0); i != schema.Width(); i++ {
// 		m := schema.Module(i)
// 		// Sanity check module
// 		if _, ok := modmap[m.Name()]; ok {
// 			panic(fmt.Sprintf("duplicate module '%s' in schema", m.Name()))
// 		}

// 		modmap[m.Name()] = i
// 	}
// 	// Done
// 	return modmap
// }

// func initialiseColumnMap[F field.Element[F]](expanded bool, schema sc.AnySchema[F]) (map[columnKey]columnId,
// 	[][]trace.RawColumn[F]) {
// 	//
// 	var (
// 		colmap  = make(map[columnKey]columnId, 100)
// 		modules = make([][]trace.RawColumn[F], schema.Width())
// 	)
// 	// Initialise modules
// 	for i := uint(0); i != schema.Width(); i++ {
// 		m := schema.Module(i)
// 		columns := make([]trace.RawColumn[F], m.Width())
// 		//
// 		for j := uint(0); j != m.Width(); j++ {
// 			var (
// 				rid = sc.NewRegisterId(j)
// 				col = m.Register(rid)
// 				key = columnKey{m.Name(), col.Name}
// 				id  = columnId{i, j}
// 			)
// 			//
// 			if _, ok := colmap[key]; ok {
// 				panic(fmt.Sprintf("duplicate column '%s' in schema", trace.QualifiedColumnName(m.Name(), col.Name)))
// 			}
// 			// Add initially empty column
// 			columns[j] = trace.RawColumn[F]{
// 				Module: m.Name(),
// 				Name:   col.Name,
// 				Data:   nil,
// 			}
// 			// Set column as expected if appropriate.
// 			if expanded || col.IsInputOutput() {
// 				colmap[key] = id
// 			}
// 		}
// 		// Initialise empty columns for this module.
// 		modules[i] = columns
// 	}
// 	// Done
// 	return colmap, modules
// }
