package ir

import (
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
)

// An n-ary sum
type Add[T any] struct {
	// Arguments for the addition operation.  At least one
	// argument is required.
	arguments []T
}

// An n-ary subtraction operator (x - y - z).
type Sub[T any] struct {
	// Arguments for the subtraction operation.  At least one
	// argument is required.
	arguments []T
}

// An n-ary product
type Mul[T any] struct {
	// Arguments for the multiplication operation.  At least one
	// argument is required.
	arguments []T
}

// ===================================================================
// Constant
// ===================================================================

// A constant value used within an expression tree.
type Constant interface { Value() *fr.Element }

// ===================================================================
// Column Access
// ===================================================================

// Represents reading the value held at a given column in the tabular
// context.  Furthermore, the current row maybe shifted up (or down)
// by a given amount.  For example, consider this table:
//
//   +-----+-----+
// k |STAMP| CT  |
//   +-----+-----+
// 0 |  0  |  9  |
//   +-----+-----+
// 1 |  1  |  0  |
//   +-----+-----+
//
// Suppose we are evaluating a constraint on row k=1 which contains
// the column accesses "STAMP(0)" and "CT(-1)".  Then, STAMP(0)=1 and
// CT(-1)=9.
type ColumnAccess interface {
	// Column name
	Column() string
	// Amount to shift which can be either negative or positive.
	Shift() int
}

// ===================================================================
// Other
// ===================================================================

// Returns the (optional) true branch when the condition evaluates to zero, and
// the (optional false branch otherwise.
type IfZero[T comparable] struct {
	// Elements contained within this list.
	condition T
	// True branch (optional).
	trueBranch T
	// False branch (optional).
	falseBranch T
}

type List[T comparable] struct {
	// Elements contained within this list.
	elements []T
}

// Normalise the result of a given expression to be either 0 or 1.  More
// specifically, a normalised expression evaluates to 0 iff the original
// expression evaluates to 0.  Otherwise, it evaluates to 1.
type Normalise[T comparable] struct {
	expr T
}
