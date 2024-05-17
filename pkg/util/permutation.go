package util

import (
	//"fmt"
	"slices"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
)

// IsPermutationOf checks whether or not a given destination column is a valid
// permutation of a given source column.  This function does not modify either
// column (though it does allocate an intermediate array).
//
// This function operators by cloning both arrays, sorting them and checking
// they are the same.
func IsPermutationOf(dst []*fr.Element, src []*fr.Element) bool {
	if len(dst) != len(src) {
		return false
	}
	// Copy arrays
	dstCopy := make([]*fr.Element, len(dst))
	srcCopy := make([]*fr.Element, len(src))

	copy(dstCopy, dst)
	copy(srcCopy, src)
	// Sort arrays
	slices.SortFunc(dstCopy, func(l *fr.Element, r *fr.Element) int { return l.Cmp(r) })
	slices.SortFunc(srcCopy, func(l *fr.Element, r *fr.Element) int { return l.Cmp(r) })
	// Check they are equal
	for i := 0; i < len(dst); i++ {
		if dstCopy[i].Cmp(srcCopy[i]) != 0 {
			return false
		}
	}
	// Match
	return true
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
func PermutationSort(cols [][]*fr.Element, signs []bool) {
	n := len(cols[0])
	m := len(cols)
	//
	rows := make([][]*fr.Element, n)
	// project into row-wise form
	for i := 0; i < n; i++ {
		row := make([]*fr.Element, m)
		for j := 0; j < m; j++ {
			row[j] = cols[j][i]
		}

		rows[i] = row
	}
	// Perform the permutation sort
	slices.SortFunc(rows, func(l []*fr.Element, r []*fr.Element) int {
		return permutationSortFunc(l, r, signs)
	})
	// Project back
	for i := 0; i < n; i++ {
		row := rows[i]
		for j := 0; j < m; j++ {
			cols[j][i] = row[j]
		}
	}
}

// AreLexicographicallySorted checks whether one or more columns are
// lexicographically sorted according to the given signs.  This operation does
// not modify or clone either array.
func AreLexicographicallySorted(cols [][]*fr.Element, signs []bool) bool {
	ncols := len(cols)
	nrows := len(cols[0])

	for i := 1; i < nrows; i++ {
		for j := 0; j < ncols; j++ {
			// Compare ith elements
			c := cols[j][i].Cmp(cols[j][i-1])
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

func permutationSortFunc(lhs []*fr.Element, rhs []*fr.Element, signs []bool) int {
	for i := 0; i < len(lhs); i++ {
		// Compare ith elements
		c := lhs[i].Cmp(rhs[i])
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
