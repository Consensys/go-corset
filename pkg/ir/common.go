package ir

import (
	"errors"
	"fmt"
	"math/big"
	"unicode"
)

// Add An n-ary sum
type Add[T any] struct {
	// Arguments for the addition operation.  At least one
	// argument is required.
	arguments []T
}

// Sub An n-ary subtraction operator (x - y - z).
type Sub[T any] struct {
	// Arguments for the subtraction operation.  At least one
	// argument is required.
	arguments []T
}

// Mul An n-ary product
type Mul[T any] struct {
	// Arguments for the multiplication operation.  At least one
	// argument is required.
	arguments []T
}

// ===================================================================
// Constant
// ===================================================================

// Constant is a constant value used within an expression tree.
type Constant struct {
	Value *big.Int
}

// StringToConstant attempts to parse a string into a constant value. This will only
// succeed if the string corresponds to a numeric value.
func StringToConstant(symbol string) (*Constant, error) {
	var num *big.Int

	num, ok := num.SetString(symbol, 10)
	if ok {
		return &Constant{num}, nil
	}

	return nil, errors.New("invalid constant")
}

// ===================================================================
// Column Access
// ===================================================================

// ColumnAccess represents reading the value held at a given column in the tabular
// context.  Furthermore, the current row maybe shifted up (or down)
// by a given amount.  For example, consider this table:
//
//	+-----+-----+
//
// k |STAMP| CT  |
//
//	+-----+-----+
//
// 0 |  0  |  9  |
//
//	+-----+-----+
//
// 1 |  1  |  0  |
//
//	+-----+-----+
//
// Suppose we are evaluating a constraint on row k=1 which contains
// the column accesses "STAMP(0)" and "CT(-1)".  Then, STAMP(0)=1 and
// CT(-1)=9.
type ColumnAccess struct {
	// Column name
	Column string
	// Amount to shift which can be either negative or positive.
	Shift int
}

// StringToColumnAccess attempts to parse a string into a column access (with a default
// shift of 0).  This will only success if the symbol is a valid
// column name.
func StringToColumnAccess(symbol string) (*ColumnAccess, error) {
	if ValidColumnName(symbol) {
		return &ColumnAccess{
			Column: symbol,
			Shift:  0,
		}, nil
	}

	return nil, errors.New("invalid column access")
}

// SliceToShiftAccess converts a slice representing a shift expression "(shift c n)" into
// a column access for column "c" with shift "n".  This will fail
// unless there are exactly two arguments, with the first being a
// column access and the second being a constant.
func SliceToShiftAccess[T comparable](args []T) (*ColumnAccess, error) {
	var msg string
	// Sanity check sufficient arguments
	if len(args) != 2 {
		msg = fmt.Sprintf("Incorrect number of shift arguments: {%d}", len(args))
	} else {
		// Extract parameters
		c, ok1 := any(args[0]).(*ColumnAccess)
		n, ok2 := any(args[1]).(*Constant)
		// Sanity check this make sense
		if ok1 && ok2 && n.Value.IsInt64() {
			n := int(n.Value.Int64())
			return &ColumnAccess{c.Column, c.Shift + n}, nil
		} else if !ok1 {
			msg = fmt.Sprintf("Shift column malformed: {%s}", any(args[0]))
		} else {
			msg = fmt.Sprintf("Shift amount malformed: {%s}", n)
		}
	}

	return nil, errors.New(msg)
}

// ValidColumnName checks whether a given column name is made up from characters,
// digits or "_" and does not start with a digit.
func ValidColumnName(s string) bool {
	for i, c := range s {
		if unicode.IsLetter(c) || c == '_' {
			// OK
		} else if i != 0 && unicode.IsNumber(c) {
			// Also OK
		} else {
			// Otherwise, not OK.
			return false
		}
	}

	return true
}

// ===================================================================
// Other
// ===================================================================

// IfZero returns the (optional) true branch when the condition evaluates to zero, and
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
