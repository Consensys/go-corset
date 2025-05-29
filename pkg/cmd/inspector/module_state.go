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
package inspector

import (
	"github.com/consensys/go-corset/pkg/corset"
	"github.com/consensys/go-corset/pkg/ir/mir"
	"github.com/consensys/go-corset/pkg/util"
)

// SourceColumn provides key information to the inspector about source-level
// columns and their mapping to registers at the HIR level (i.e. columns we
// would find in the trace).
type SourceColumn struct {
	// Column name
	Name string
	// Determines whether this is a Computed column.
	Computed bool
	// Selector determines when column active.
	Selector mir.Term
	// Display modifier
	Display uint
	// Register to which this column is allocated
	Register uint
}

// ExtractSourceColumns extracts source column descriptions for a given module
// based on the corset source mapping.  This is particularly useful when you
// want to show the original name for a column (e.g. when its in a perspective),
// rather than the raw register name.
func ExtractSourceColumns(path util.Path, selector mir.Term, columns []corset.SourceColumn,
	submodules []corset.SourceModule) []SourceColumn {
	//
	var srcColumns []SourceColumn
	//
	for _, col := range columns {
		name := path.Extend(col.Name).String()[1:]
		srcCol := SourceColumn{name, col.Computed, selector, col.Display, col.Register}
		srcColumns = append(srcColumns, srcCol)
	}
	//
	for _, submod := range submodules {
		subpath := path.Extend(submod.Name)
		subSrcColumns := ExtractSourceColumns(*subpath, submod.Selector, submod.Columns, submod.Submodules)
		srcColumns = append(srcColumns, subSrcColumns...)
	}
	//
	return srcColumns
}
