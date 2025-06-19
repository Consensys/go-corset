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
	"math"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util/collection/bit"
	"github.com/consensys/go-corset/pkg/util/field"
)

// Fill a set of columns with their computed results.  The column index is that
// of the first column in the sequence, and subsequent columns are index
// consecutively.
func fillComputedColumns(refs []sc.RegisterRef, cols []tr.ArrayColumn, trace *tr.ArrayTrace) {
	var resized bit.Set
	// Add all columns
	for i, ref := range refs {
		var (
			rid    = ref.Column().Unwrap()
			module = trace.RawModule(ref.Module())
			dst    = module.Column(rid)
			col    = cols[i]
		)
		// Sanity checks
		if dst.Name() != col.Name() {
			mod := module.Name()
			panic(fmt.Sprintf("misaligned computed register %s.%s during trace expansion", mod, col.Name()))
		}
		// Looks good
		if module.FillColumn(rid, col.Data(), col.Padding()) {
			// Register module as being resized.
			resized.Insert(ref.Module())
		}
	}
	// Finalise resized modules
	for iter := resized.Iter(); iter.HasNext(); {
		module := trace.RawModule(iter.Next())
		module.Resize()
	}
}

// Validate that all elements of a given column fit within a given bitwidth.
func validateColumnBitWidth(bitwidth uint, col tr.Column, mod sc.Module) error {
	// Sanity check bitwidth can be checked.
	if bitwidth == math.MaxUint {
		// This indicates a column which has no fixed bitwidth but, rather, uses
		// the entire field element.  The only situation this arises in practice
		// is for columns holding the multiplicative inverse of some other
		// column.
		return nil
	} else if col.Data() == nil {
		panic(fmt.Sprintf("column %s is unassigned", col.Name()))
	}
	//
	var frBound fr.Element = fr.NewElement(2)
	// Compute 2^n
	field.Pow(&frBound, uint64(bitwidth))
	//
	for j := 0; j < int(col.Data().Len()); j++ {
		var jth = col.Get(j)
		//
		if jth.Cmp(&frBound) >= 0 {
			qualColName := trace.QualifiedColumnName(mod.Name(), col.Name())
			return fmt.Errorf("row %d of column %s is out-of-bounds (%s)", j, qualColName, jth.String())
		}
	}
	// success
	return nil
}
