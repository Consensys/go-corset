package gadgets

import (
	"fmt"
	"strings"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/air"
	"github.com/consensys/go-corset/pkg/table"
)

// ApplyLexicographicSortingGadget Add sorting constraints for a sequence of one
// or more columns.  Sorting is done lexicographically starting from the
// leftmost column.  For example, consider lexicographically sorting two columns
// X and Y (in that order) in ascending (i.e. positive direction).  Then sorting
// ensures (X[k-1] < X[k]) or (X[k-1] == X[k] and Y[k-1] <= Y[k]).  The sign for
// each column determines whether its sorted into ascending (i.e. positive) or
// descending (i.e. negative) order.
//
// To implement this sort, a kind of "bit multiplexing" is used.  Specifically,
// a bit column is associated with each column being sorted, where exactly one
// of these bits can be 1.  That bit identifies the leftmost column Ci where
// Ci[k-1] < C[k].  For all columns Cj where j < i, we must have Cj[k-1] =
// Cj[k].  If all bits are zero then all columns match their previous row.
// Finally, a delta column is used in a similar fashion as for the single column
// case (see above).  The delta value captures the difference Ci[k]-Ci[k-1] to
// ensure it is positive.  The delta column is constrained to a given bitwidth,
// with constraints added as necessary to ensure this.
func ApplyLexicographicSortingGadget(columns []string, signs []bool, bitwidth uint, schema *air.Schema) {
	// Check preconditions
	ncols := len(columns)
	if ncols != len(signs) {
		panic("Inconsistent number of columns and signs for lexicographic sort.")
	}
	// Add trace computation
	schema.AddComputation(&lexicographicSortExpander{columns, signs, bitwidth})
	// Construct a unique prefix for this sort.
	prefix := constructLexicographicSortingPrefix(columns, signs)
	deltaName := fmt.Sprintf("%s:delta", prefix)
	// Construct selecto bits.
	bits := addLexicographicSelectorBits(prefix, columns, schema)
	// Add delta column
	schema.AddColumn(deltaName, true)
	// Construct delta terms
	constraint := constructLexicographicDeltaConstraint(deltaName, bits, columns, signs)
	// Add delta constraint
	schema.AddVanishingConstraint(deltaName, nil, constraint)
	// Add necessary bitwidth constraints
	ApplyBitwidthGadget(deltaName, bitwidth, schema)
}

// Construct a unique identifier for the given sort.  This should not conflict
// with the identifier for any other sort.
func constructLexicographicSortingPrefix(columns []string, signs []bool) string {
	// Use string builder to try and make this vaguely efficient.
	var id strings.Builder
	// Concatenate column names with their signs.
	for i := 0; i < len(columns); i++ {
		id.WriteString(columns[i])

		if signs[i] {
			id.WriteString("+")
		} else {
			id.WriteString("-")
		}
	}
	// Done
	return id.String()
}

// Add lexicographic selector bits, including the necessary constraints.  Each
// selector bit is given a binarity constraint to ensure it is always either 1
// or 0.  A selector bit can only be set if all bits to its left are unset, and
// there is a strict difference between the two values for its colummn.
//
// NOTE: this implementation differs from the original corset which used an
// additional "Eq" bit to help ensure at most one selector bit was enabled.
func addLexicographicSelectorBits(prefix string, columns []string, schema *air.Schema) []string {
	ncols := len(columns)
	// Add bits and their binary constraints.
	bits := AddBitArray(prefix, ncols, schema)
	// Apply constraints to ensure at most one is set.
	terms := make([]air.Expr, ncols)
	for i := 0; i < ncols; i++ {
		terms[i] = air.NewColumnAccess(bits[i], 0)
		pterms := make([]air.Expr, i+1)
		qterms := make([]air.Expr, i)

		for j := 0; j < i; j++ {
			pterms[j] = air.NewColumnAccess(bits[j], 0)
			qterms[j] = air.NewColumnAccess(bits[j], 0)
		}
		// (∀j<=i.Bj=0) ==> C[k]=C[k-1]
		pterms[i] = air.NewColumnAccess(bits[i], 0)
		pDiff := air.NewColumnAccess(columns[i], 0).Sub(air.NewColumnAccess(columns[i], -1))
		pName := fmt.Sprintf("%s:%d:a", prefix, i)
		schema.AddVanishingConstraint(pName, nil, air.NewConst64(1).Sub(&air.Add{Args: pterms}).Mul(pDiff))
		// (∀j<i.Bj=0) ∧ Bi=1 ==> C[k]≠C[k-1]
		qDiff := Normalise(air.NewColumnAccess(columns[i], 0).Sub(air.NewColumnAccess(columns[i], -1)), schema)
		qName := fmt.Sprintf("%s:%d:b", prefix, i)
		// bi = 0 || C[k]≠C[k-1]
		constraint := air.NewColumnAccess(bits[i], 0).Mul(air.NewConst64(1).Sub(qDiff))

		if i != 0 {
			// (∃j<i.Bj≠0) || bi = 0 || C[k]≠C[k-1]
			constraint = air.NewConst64(1).Sub(&air.Add{Args: qterms}).Mul(constraint)
		}

		schema.AddVanishingConstraint(qName, nil, constraint)
	}

	sum := &air.Add{Args: terms}
	// (sum = 0) ∨ (sum = 1)
	constraint := sum.Mul(sum.Equate(air.NewConst64(1)))
	name := fmt.Sprintf("%s:xor", prefix)
	schema.AddVanishingConstraint(name, nil, constraint)

	return bits
}

// Construct the lexicographic delta constraint.  This states that the delta
// column either holds 0 or the difference Ci[k] - Ci[k-1] (adjusted
// appropriately for the sign) between the ith column whose multiplexor bit is
// set. This is assumes that multiplexor bits are mutually exclusive (i.e. at
// most is one).
func constructLexicographicDeltaConstraint(deltaName string, bits []string, columns []string, signs []bool) air.Expr {
	ncols := len(columns)
	// Construct delta terms
	terms := make([]air.Expr, ncols)
	Dk := air.NewColumnAccess(deltaName, 0)

	for i := 0; i < ncols; i++ {
		var Xdiff air.Expr
		// Ith bit column (at row k)
		Bk := air.NewColumnAccess(bits[i], 0)
		// Ith column (at row k)
		Xk := air.NewColumnAccess(columns[i], 0)
		// Ith column (at row k-1)
		Xkm1 := air.NewColumnAccess(columns[i], -1)
		if signs[i] {
			Xdiff = Xk.Sub(Xkm1)
		} else {
			Xdiff = Xkm1.Sub(Xk)
		}
		// if Bk then Xdiff
		terms[i] = Bk.Mul(Xdiff)
	}
	// Construct final constraint
	return Dk.Equate(&air.Add{Args: terms})
}

type lexicographicSortExpander struct {
	columns  []string
	signs    []bool
	bitwidth uint
}

// RequiredSpillage returns the minimum amount of spillage required to ensure
// valid traces are accepted in the presence of arbitrary padding.
func (p *lexicographicSortExpander) RequiredSpillage() uint {
	return uint(0)
}

// Accepts checks whether a given trace has the necessary columns
func (p *lexicographicSortExpander) Accepts(tr table.Trace) error {
	prefix := constructLexicographicSortingPrefix(p.columns, p.signs)
	deltaName := fmt.Sprintf("%s:delta", prefix)
	// Check delta column exists
	if !tr.HasColumn(deltaName) {
		return fmt.Errorf("Trace missing lexicographic delta column ({%s})", deltaName)
	}
	// Check selector columns exist
	for i := range p.columns {
		bitName := fmt.Sprintf("%s:%d", prefix, i)
		if !tr.HasColumn(bitName) {
			return fmt.Errorf("Trace missing lexicographic selector column ({%s})", bitName)
		}
	}
	//
	return nil
}

// Add columns as needed to support the LexicographicSortingGadget.  That
// includes the delta column, and the bit selectors.
func (p *lexicographicSortExpander) ExpandTrace(tr table.Trace) error {
	zero := fr.NewElement(0)
	one := fr.NewElement(1)
	// Exact number of columns involved in the sort
	ncols := len(p.columns)
	// Determine how many rows to be constrained.
	nrows := tr.Height()
	// Construct a unique prefix for this sort.
	prefix := constructLexicographicSortingPrefix(p.columns, p.signs)
	deltaName := fmt.Sprintf("%s:delta", prefix)
	// Initialise new data columns
	delta := make([]*fr.Element, nrows)
	bit := make([][]*fr.Element, ncols)

	for i := 0; i < ncols; i++ {
		bit[i] = make([]*fr.Element, nrows)
	}

	for i := uint(0); i < nrows; i++ {
		set := false
		// Initialise delta to zero
		delta[i] = &zero
		// Decide which row is the winner (if any)
		for j := 0; j < ncols; j++ {
			prev := tr.GetByName(p.columns[j], int(i-1))
			curr := tr.GetByName(p.columns[j], int(i))

			if !set && prev != nil && prev.Cmp(curr) != 0 {
				var diff fr.Element

				bit[j][i] = &one
				// Compute curr - prev
				if p.signs[j] {
					diff.Set(curr)
					delta[i] = diff.Sub(&diff, prev)
				} else {
					diff.Set(prev)
					delta[i] = diff.Sub(&diff, curr)
				}

				set = true
			} else {
				bit[j][i] = &zero
			}
		}
	}

	// Add delta column data
	tr.AddColumn(deltaName, delta, &zero)
	// Add bit column data
	for i := 0; i < ncols; i++ {
		bitName := fmt.Sprintf("%s:%d", prefix, i)
		tr.AddColumn(bitName, bit[i], &zero)
	}
	// Done.
	return nil
}

// String returns a string representation of this constraint.  This is primarily
// used for debugging.
func (p *lexicographicSortExpander) String() string {
	return fmt.Sprintf("(lexer (%s) (%v) :%d))", any(p.columns), p.signs, p.bitwidth)
}
