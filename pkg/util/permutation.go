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
	//"fmt"

	"slices"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

// ArePermutationOf checks whether or not a set of given destination columns are
// a valid permutation of a given set of source columns.  The number of source
// and target columns must match.  Likewise, they are expected to have the same
// height. This function does not modify any columns (though it does allocate
// intermediate arrays).
//
// This function operators by cloning the arrays, sorting them and checking they
// are the same.
func ArePermutationOf[T Array[fr.Element]](dst []T, src []T) bool {
	if len(dst) != len(src) {
		return false
	}
	// Determine geometry
	ncols := len(dst)
	nrows := dst[0].Len()
	// Rotate input arrays
	dstCopy := rotate(dst, ncols, nrows)
	srcCopy := rotate(src, ncols, nrows)
	// Sort rotated arrays
	slices.SortFunc(dstCopy, permutationFunc)
	slices.SortFunc(srcCopy, permutationFunc)
	// Check rotated arrays match
	return Equals2d(dstCopy, srcCopy)
}

func permutationFunc(lhs []fr.Element, rhs []fr.Element) int {
	for i := 0; i < len(lhs); i++ {
		// Compare ith elements
		c := lhs[i].Cmp(&rhs[i])
		// Check whether same
		if c != 0 {
			// Positive
			return c
		}
	}
	// Identical
	return 0
}

// PermutationSort sorts an array of columns in row-wise fashion.  For
// example, suppose consider [ [0,4,3,3], [1,2,4,3] ].  We can imagine
// that this is first transformed into an array of rows (i.e.
// [[0,1],[4,2],[3,4],[3,3]]) and then sorted lexicographically (to
// give [[0,1],[3,3],[3,4],[4,2]]).  This is then projected back into
// the original column-wise formulation, to give: [[0,3,3,4],
// [1,3,4,2]].
//
// A further complication is that the direction of sorting for each
// columns is determined by its sign.
//
// NOTE: the current implementation is not intended to be particularly
// efficient.  In particular, would be better to do the sort directly
// on the columns array without projecting into the row-wise form.
func PermutationSort[T Array[fr.Element]](cols []T, signs []bool) {
	n := cols[0].Len()
	m := len(cols)
	// Rotate input matrix
	rows := rotate(cols, m, n)
	// Perform the permutation sort
	slices.SortFunc(rows, func(l []fr.Element, r []fr.Element) int {
		return permutationSortFunc(l, r, signs)
	})
	// Project back
	for i := uint(0); i < n; i++ {
		row := rows[i]
		for j := 0; j < m; j++ {
			cols[j].Set(i, row[j])
		}
	}
}

// AreLexicographicallySorted checks whether one or more columns are
// lexicographically sorted according to the given signs.  This operation does
// not modify or clone either array.
func AreLexicographicallySorted(cols [][]fr.Element, signs []bool) bool {
	ncols := len(cols)
	nrows := len(cols[0])

	for i := 1; i < nrows; i++ {
		for j := 0; j < ncols; j++ {
			// Compare ith elements
			c := cols[j][i].Cmp(&cols[j][i-1])
			// Check whether same
			if signs[j] && c < 0 {
				return false
			} else if !signs[j] && c > 0 {
				return false
			} else if c != 0 {
				return true
			}
		}
	}

	return true
}

func permutationSortFunc(lhs []fr.Element, rhs []fr.Element, signs []bool) int {
	for i := 0; i < len(signs); i++ {
		// Compare ith elements
		c := lhs[i].Cmp(&rhs[i])
		// Check whether same
		if c != 0 {
			if signs[i] {
				// Positive
				return c
			}
			// Negative
			return -c
		}
	}
	// Identical
	return 0
}

// Clone and rotate a 2-dimensional array assuming a given geometry.
func rotate[T Array[fr.Element]](src []T, ncols int, nrows uint) [][]fr.Element {
	// Copy outer arrays
	dst := make([][]fr.Element, nrows)
	// Copy inner arrays
	for i := uint(0); i < nrows; i++ {
		row := make([]fr.Element, ncols)
		for j := 0; j < ncols; j++ {
			row[j] = src[j].Get(i)
		}

		dst[i] = row
	}
	//
	return dst
}
