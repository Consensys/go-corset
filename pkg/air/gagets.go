package air

import (
	"fmt"
	"strings"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/table"
)

// Norm constructs an expression representing the normalised value of e.  That is,
// an expression which is 0 when e is 0, and 1 when e is non-zero.  This is done
// by introducing a synthetic column to hold the (pseudo) mutliplicative inverse
// of e.
func Norm(e Expr, tbl *Schema) Expr {
	// Construct pseudo multiplicative inverse of e.
	ie := ApplyPseudoInverseGadget(e, tbl)
	// Return e * e⁻¹.
	return e.Mul(ie)
}

// ApplyPseudoInverseGadget constructs an expression representing the
// (pseudo) multiplicative inverse of another expression.  Since this cannot be computed
// directly using arithmetic constraints, it is done by adding a new computed
// column which holds the multiplicative inverse.  Constraints are also added to
// ensure it really holds the inverted value.
func ApplyPseudoInverseGadget(e Expr, tbl *Schema) Expr {
	// Construct inverse computation
	ie := &Inverse{Expr: e}
	// Determine computed column name
	name := ie.String()
	// Add new column (if it does not already exist)
	if !tbl.HasColumn(name) {
		// Add (synthetic) computed column
		tbl.AddColumn(name, true)
		tbl.AddComputation(table.NewComputedColumn(name, ie))
	}

	// Construct 1/e
	inv_e := NewColumnAccess(name, 0)
	// Construct e/e
	e_inv_e := e.Mul(inv_e)
	// Construct 1 == e/e
	one_e_e := NewConst64(1).Equate(e_inv_e)
	// Construct (e != 0) ==> (1 == e/e)
	e_implies_one_e_e := e.Mul(one_e_e)
	// Construct (1/e != 0) ==> (1 == e/e)
	inv_e_implies_one_e_e := inv_e.Mul(one_e_e)
	// Ensure (e != 0) ==> (1 == e/e)
	l_name := fmt.Sprintf("[%s <=]", ie.String())
	tbl.AddVanishingConstraint(l_name, nil, e_implies_one_e_e)
	// Ensure (e/e != 0) ==> (1 == e/e)
	r_name := fmt.Sprintf("[%s =>]", ie.String())
	tbl.AddVanishingConstraint(r_name, nil, inv_e_implies_one_e_e)
	// Done
	return NewColumnAccess(name, 0)
}

// Inverse represents a computation which computes the multiplicative
// inverse of a given AIR expression.
type Inverse struct{ Expr Expr }

// EvalAt computes the multiplicative inverse of a given expression at a given
// row in the table.
func (e *Inverse) EvalAt(k int, tbl table.Trace) *fr.Element {
	inv := new(fr.Element)
	val := e.Expr.EvalAt(k, tbl)
	// Catch undefined case
	if val == nil {
		return nil
	}
	// Go syntax huh?
	return inv.Inverse(val)
}

func (e *Inverse) String() string {
	return fmt.Sprintf("(inv %s)", e.Expr)
}

// ApplyBinaryGadget adds a binarity constraint for a given column in the schema
// which enforces that all values in the given column are either 0 or 1. For a
// column X, this corresponds to the vanishing constraint X * (X-1) == 0.
func ApplyBinaryGadget(col string, schema *Schema) {
	// Construct X
	X := NewColumnAccess(col, 0)
	// Construct X-1
	X_m1 := X.Sub(NewConst64(1))
	// Construct X * (X-1)
	X_X_m1 := X.Mul(X_m1)
	// Done!
	schema.AddVanishingConstraint(col, nil, X_X_m1)
}

// ApplyBitwidthGadget ensures all values in a given column fit within a given
// number of bits.  This is implemented using a *byte decomposition* which adds
// n columns and a vanishing constraint (where n*8 >= nbits).
func ApplyBitwidthGadget(col string, nbits uint, schema *Schema) {
	if nbits%8 != 0 {
		panic("asymmetric bitwidth constraints not yet supported")
	} else if nbits == 0 {
		panic("zero bitwidth constraint encountered")
	}
	// Calculate how many bytes required.
	n := nbits / 8
	es := make([]Expr, n)
	fr256 := fr.NewElement(256)
	coefficient := fr.NewElement(1)
	// Construct Columns
	for i := uint(0); i < n; i++ {
		// Determine name for the ith byte column
		colName := fmt.Sprintf("%s:%d", col, i)
		// Create Column + Constraint
		schema.AddColumn(colName, true)
		schema.AddRangeConstraint(colName, &fr256)
		es[i] = NewColumnAccess(colName, 0).Mul(NewConstCopy(&coefficient))
		// Update coefficient
		coefficient.Mul(&coefficient, &fr256)
	}
	// Construct (X:0 * 1) + ... + (X:n * 2^n)
	sum := &Add{Args: es}
	// Construct X == (X:0 * 1) + ... + (X:n * 2^n)
	X := NewColumnAccess(col, 0)
	eq := X.Equate(sum)
	schema.AddVanishingConstraint(col, nil, eq)
	// Finally, add the necessary byte decomposition computation.
	schema.AddComputation(table.NewByteDecomposition(col, nbits))
}

// ApplyColumnSortingGadget Add sorting constraints for a column where the
// difference between any two rows (i.e. the delta) is constrained to fit within
// a given bitwidth.  The target column is assumed to have an appropriate
// (enforced) bitwidth to ensure overflow cannot arise.  The sorting constraint
// is either ascending (positively signed) or descending (negatively signed).  A
// delta column is added along with bitwidth constraints (where necessary) to
// ensure the delta is within the given width.
func ApplyColumnSortingGadget(column string, sign bool, bitwidth uint, schema *Schema) {
	var deltaName string
	// Configure computation
	Xk := NewColumnAccess(column, 0)
	Xkm1 := NewColumnAccess(column, -1)
	// Account for sign
	var Xdiff Expr
	if sign {
		Xdiff = Xk.Sub(Xkm1)
		deltaName = fmt.Sprintf("+%s", column)
	} else {
		Xdiff = Xkm1.Sub(Xk)
		deltaName = fmt.Sprintf("-%s", column)
	}
	// Add delta column
	schema.AddColumn(deltaName, true)
	// Add diff computation
	schema.AddComputation(table.NewComputedColumn(deltaName, Xdiff))
	// Add necessary bitwidth constraints
	ApplyBitwidthGadget(deltaName, bitwidth, schema)
	// Configure constraint: Delta[k] = X[k] - X[k-1]
	Dk := NewColumnAccess(deltaName, 0)
	schema.AddVanishingConstraint(deltaName, nil, Dk.Equate(Xdiff))
}

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
func ApplyLexicographicSortingGadget(columns []string, signs []bool, bitwidth uint, schema *Schema) {
	// Check preconditions
	ncols := len(columns)
	if ncols != len(signs) {
		panic("Inconsistent number of columns and signs for lexicographic sort.")
	}
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
	// FIXME: Add trace expansion computation
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
func addLexicographicSelectorBits(prefix string, columns []string, schema *Schema) []string {
	ncols := len(columns)
	// Add bits and their binary constraints.
	bits := AddBitArray(prefix, ncols, schema)
	// Apply constraints to ensure at most one is set.
	terms := make([]Expr, ncols)
	for i := 0; i < ncols; i++ {
		terms[i] = NewColumnAccess(bits[i], 0)
		pterms := make([]Expr, i+1)
		qterms := make([]Expr, i)

		for j := 0; j < i; j++ {
			pterms[j] = NewColumnAccess(columns[j], 0)
			qterms[j] = NewColumnAccess(columns[j], 0)
		}
		// (∀j<=i.Bj=0) ==> C[k]=C[k-1]
		pterms[i] = NewColumnAccess(columns[i], 0)
		pDiff := NewColumnAccess(columns[i], 0).Sub(NewColumnAccess(columns[i], -1))
		pName := fmt.Sprintf("%s:%d:0", prefix, i)
		schema.AddVanishingConstraint(pName, nil, NewConst64(1).Sub(&Add{pterms}).Mul(pDiff))
		// (∀j<i.Bj=0) ∧ Bi=1 ==> C[k]≠C[k-1]
		qDiff := Norm(NewColumnAccess(columns[i], 0).Sub(NewColumnAccess(columns[i], -1)), schema)
		qName := fmt.Sprintf("%s:%d:1", prefix, i)
		constraint := NewConst64(1).Sub(qDiff)

		if i != 0 {
			constraint = NewConst64(1).Sub(&Add{qterms}).Mul(NewColumnAccess(columns[i], 0).Mul(constraint))
			schema.AddVanishingConstraint(qName, nil, constraint)
		}

		schema.AddVanishingConstraint(qName, nil, constraint)
	}

	sum := &Add{Args: terms}
	// (sum = 0) ∨ (sum = 1)
	constraint := sum.Mul(sum.Equate(NewConst64(1)))
	name := fmt.Sprintf("%s:xor", prefix)
	schema.AddVanishingConstraint(name, nil, constraint)

	return bits
}

// Construct the lexicographic delta constraint.  This states that the delta
// column either holds 0 or the difference Ci[k] - Ci[k-1] (adjusted
// appropriately for the sign) between the ith column whose multiplexor bit is
// set. This is assumes that multiplexor bits are mutually exclusive (i.e. at
// most is one).
func constructLexicographicDeltaConstraint(deltaName string, bits []string, columns []string, signs []bool) Expr {
	ncols := len(columns)
	// Construct delta terms
	terms := make([]Expr, ncols)
	Dk := NewColumnAccess(deltaName, 0)

	for i := 0; i < ncols; i++ {
		var Xdiff Expr
		// Ith bit column (at row k)
		Bk := NewColumnAccess(bits[i], 0)
		// Ith column (at row k)
		Xk := NewColumnAccess(columns[i], 0)
		// Ith column (at row k-1)
		Xkm1 := NewColumnAccess(columns[i], -1)
		if signs[i] {
			Xdiff = Xk.Sub(Xkm1)
		} else {
			Xdiff = Xkm1.Sub(Xk)
		}
		// if Bk then Xdiff
		terms[i] = Bk.Mul(Xdiff)
	}
	// Construct final constraint
	return Dk.Equate(&Add{Args: terms})
}

// AddBitArray adds an array of n bit columns using a given prefix, including
// the necessary binarity constraints.
func AddBitArray(prefix string, count int, schema *Schema) []string {
	bits := make([]string, count)
	for i := 0; i < count; i++ {
		// Construct bit column name
		bits[i] = fmt.Sprintf("%s:%d", prefix, i)
		// Add (synthetic) column
		schema.AddColumn(bits[i], true)
		// Add binarity constraints (i.e. to enfoce that this column is a bit).
		ApplyBinaryGadget(bits[i], schema)
	}
	//
	return bits
}
