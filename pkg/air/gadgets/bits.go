package gadgets

import (
	"fmt"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/air"
	"github.com/consensys/go-corset/pkg/table"
)

// ApplyBinaryGadget adds a binarity constraint for a given column in the schema
// which enforces that all values in the given column are either 0 or 1. For a
// column X, this corresponds to the vanishing constraint X * (X-1) == 0.
func ApplyBinaryGadget(col string, schema *air.Schema) {
	// Construct X
	X := air.NewColumnAccess(col, 0)
	// Construct X-1
	X_m1 := X.Sub(air.NewConst64(1))
	// Construct X * (X-1)
	X_X_m1 := X.Mul(X_m1)
	// Done!
	schema.AddVanishingConstraint(col, nil, X_X_m1)
}

// ApplyBitwidthGadget ensures all values in a given column fit within a given
// number of bits.  This is implemented using a *byte decomposition* which adds
// n columns and a vanishing constraint (where n*8 >= nbits).
func ApplyBitwidthGadget(col string, nbits uint, schema *air.Schema) {
	if nbits%8 != 0 {
		panic("asymmetric bitwidth constraints not yet supported")
	} else if nbits == 0 {
		panic("zero bitwidth constraint encountered")
	}
	// Calculate how many bytes required.
	n := nbits / 8
	es := make([]air.Expr, n)
	fr256 := fr.NewElement(256)
	coefficient := fr.NewElement(1)
	// Construct Columns
	for i := uint(0); i < n; i++ {
		// Determine name for the ith byte column
		colName := fmt.Sprintf("%s:%d", col, i)
		// Create Column + Constraint
		schema.AddColumn(colName, true)
		schema.AddRangeConstraint(colName, &fr256)
		es[i] = air.NewColumnAccess(colName, 0).Mul(air.NewConstCopy(&coefficient))
		// Update coefficient
		coefficient.Mul(&coefficient, &fr256)
	}
	// Construct (X:0 * 1) + ... + (X:n * 2^n)
	sum := &air.Add{Args: es}
	// Construct X == (X:0 * 1) + ... + (X:n * 2^n)
	X := air.NewColumnAccess(col, 0)
	eq := X.Equate(sum)
	// Construct column name
	schema.AddVanishingConstraint(fmt.Sprintf("%s:u%d", col, nbits), nil, eq)
	// Finally, add the necessary byte decomposition computation.
	schema.AddComputation(table.NewByteDecomposition(col, nbits))
}

// AddBitArray adds an array of n bit columns using a given prefix, including
// the necessary binarity constraints.
func AddBitArray(prefix string, count int, schema *air.Schema) []string {
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
