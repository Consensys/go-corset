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
package builder

import (
	"fmt"

	"github.com/consensys/go-corset/pkg/schema/module"
	"github.com/consensys/go-corset/pkg/schema/register"
	"github.com/consensys/go-corset/pkg/trace/lt"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/array"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/word"
)

// TraceSplitting splits a given set of raw columns according to a given
// register mapping or, otherwise, simply lowers them.
func TraceSplitting[F field.Element[F]](parallel bool, tf lt.TraceFile, mapping module.LimbsMap) (array.Builder[F],
	[]lt.Module[F], []error) {
	//
	var (
		stats   = util.NewPerfStats()
		builder array.Builder[F]
		modules []lt.Module[F]
		err     []error
	)
	//
	if parallel {
		builder, modules, err = parallelTraceSplitting[F](tf, mapping)
	} else {
		builder, modules, err = sequentialTraceSplitting[F](tf, mapping)
	}
	//
	stats.Log("Trace splitting")
	//
	return builder, modules, err
}

func sequentialTraceSplitting[F field.Element[F]](tf lt.TraceFile, mapping module.LimbsMap) (array.Builder[F],
	[]lt.Module[F], []error) {
	//
	var (
		modules = make([]lt.Module[F], len(tf.Modules))
		// Allocate fresh array builder
		builder = array.NewStaticBuilder[F]()
		errors  []error
	)
	//
	for i, ith := range tf.Modules {
		var (
			columns []lt.Column[F] // Access mapping for enclosing module
			modmap  = mapping.ModuleOf(ith.Name)
		)
		//
		for _, jth := range ith.Columns {
			split, errs := splitRawColumn(jth, builder, modmap)
			columns = append(columns, split...)
			errors = append(errors, errs...)
		}
		//
		modules[i] = lt.Module[F]{Name: ith.Name, Columns: columns}
	}
	//
	return builder, modules, errors
}

func parallelTraceSplitting[F field.Element[F]](tf lt.TraceFile, mapping module.LimbsMap) (array.Builder[F],
	[]lt.Module[F], []error) {
	//
	var (
		ncols = lt.NumberOfColumns(tf.Modules)
		//
		splits = make([][][]lt.Column[F], len(tf.Modules))
		// Allocate fresh array builder
		builder = array.NewStaticBuilder[F]()
		// Construct a communication channel split columns.
		c = make(chan splitResult[F], ncols)
		//
		errors []error
	)
	// Split column concurrently
	for i, ith := range tf.Modules {
		// Access mapping for enclosing module
		modmap := mapping.ModuleOf(ith.Name)
		// Initiali split array
		splits[i] = make([][]lt.Column[F], len(ith.Columns))
		//
		for j, jth := range ith.Columns {
			// Start go-routine for this column
			go func(mid, cid int, column lt.Column[word.BigEndian], mapping module.LimbsMap) {
				// Send outcome back
				data, errors := splitRawColumn(column, builder, modmap)
				c <- splitResult[F]{mid, cid, data, errors}
			}(i, j, jth, mapping)
		}
	}
	// Collect results
	for range ncols {
		// Read from channel
		res := <-c
		// Assign split
		splits[res.module][res.column] = res.data
		errors = append(errors, res.errors...)
	}
	// Flatten split
	return builder, flatten(tf, splits), errors
}

func flatten[W any](tf lt.TraceFile, splits [][][]lt.Column[W]) []lt.Module[W] {
	var (
		modules = make([]lt.Module[W], len(splits))
		index   = 0
	)
	// Flattern all columns
	for i, ith := range splits {
		var columns []lt.Column[W]
		//
		for _, jth := range ith {
			columns = append(columns, jth...)
			index++
		}
		//
		modules[i] = lt.Module[W]{Name: tf.Modules[i].Name, Columns: columns}
	}
	//
	return modules
}

// SplitRawColumn splits a given raw column using the given register mapping.
func splitRawColumn[F field.Element[F]](col lt.Column[word.BigEndian], builder array.Builder[F],
	modmap register.LimbsMap) ([]lt.Column[F], []error) {
	//
	var (
		height uint
		//
		reg, regExists = modmap.HasRegister(col.Name)
	)
	// Check whether register is known
	if !regExists {
		// Unknown register --- this is an error
		return nil, []error{fmt.Errorf("unknown register \"%s\"", col.Name)}
	}
	// Calculate register height.  Observe that computed registers will have nil
	// for their data at this point since they haven't been computed yet.
	if col.Data != nil {
		height = col.Data.Len()
	}
	// Determine register id for this column (we can assume it exists)
	limbIds := modmap.LimbIds(reg)
	//
	if len(limbIds) == 1 {
		// No, this register was not split into any limbs.  Therefore, no need
		// to split the column into any limbs.
		return []lt.Column[F]{lowerRawColumn(col, builder)}, nil
	}
	// Yes, must split into two or more limbs of given widths.
	limbWidths := register.WidthsOfLimbs(modmap, modmap.LimbIds(reg))
	// Determine limbs of this register
	limbs := register.LimbsOf(modmap, limbIds)
	// Construct temporary place holder for new array data.
	arrays := make([]array.MutArray[F], len(limbIds))
	//
	for i, limb := range limbs {
		arrays[i] = builder.NewArray(height, limb.Width)
	}
	// Deconstruct all data
	for i := range height {
		// Extract ith data
		ith := col.Data.Get(i)
		// Assign split components
		setSplitWord(ith, i, arrays, limbWidths)
	}
	// Construct final columns
	columns := make([]lt.Column[F], len(limbIds))
	// Construct final columns
	for i, limb := range limbs {
		columns[i] = lt.Column[F]{
			Name: limb.Name,
			Data: arrays[i]}
	}
	// Done
	return columns, nil
}

// split a given field element into a given set of limbs, where the least
// significant comes first.  NOTE: this is really a temporary function which
// should be eliminated when RawColumn is moved away from fr.Element.
func setSplitWord[F field.Element[F]](val word.BigEndian, row uint, arrays []array.MutArray[F], widths []uint) {
	// FIXME: following is not efficient, as it allocates memory and does quite
	// a lot of work overall.
	var elements = field.SplitWord[F](val, widths)
	//
	for i := range widths {
		arrays[i].Set(row, elements[i])
	}
}

// SplitResult is returned by worker threads during parallel trace splitting.
type splitResult[W any] struct {
	module int
	column int
	data   []lt.Column[W]
	errors []error
}
