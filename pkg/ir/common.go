package ir

import (
	"errors"
	"fmt"
	"math/big"
	"unicode"
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
type Constant interface {
	Value() *big.Int
}

// Attempt to parse a string into a constant value.  This will only
// succeed if the string corresponds to a numeric value.
func StringToConstant(symbol string) (*big.Int,error) {
	num := new(big.Int)
	num,ok := num.SetString(symbol,10)
	if ok { return num,nil }
	return nil,errors.New("invalid constant")
}

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

// Attempt to parse a string into a column access (with a default
// shift of 0).  This will only success if the symbol is a valid
// column name.
func StringToColumnAccess(symbol string) (string,int,error) {
	if ValidColumnName(symbol) {
		return symbol,0,nil
	}
	return "",0,errors.New("invalid column access")
}

// Convert a slice representing a shift expression "(shift c n)" into
// a column access for column "c" with shift "n".  This will fail
// unless there are exactly two arguments, with the first being a
// column access and the second being a constant.
func SliceToShiftAccess[T comparable](args []T) (string,int,error) {
	var msg string
	// Sanity check sufficient arguments
	if len(args) != 2 {
		msg = fmt.Sprintf("Incorrect number of shift arguments: {%d}",len(args))
	} else {
		// Extract parameters
		c,ok1 := any(args[0]).(ColumnAccess)
		n,ok2 := any(args[1]).(Constant)
		// Sanit check this make sense
		if ok1 && ok2 && n.Value().IsInt64() {
			n := int(n.Value().Int64())
			return c.Column(),c.Shift()+n,nil
		} else if !ok1 {
			msg = fmt.Sprintf("Shift column malformed: {%s}",any(args[0]))
		} else {
			msg = fmt.Sprintf("Shift amount malformed: {%s}",n)
		}
	}
	return "", 0, errors.New(msg)
}

// Check whether a given column name is made up fom characters,
// digits or "_" and does not start with a digit.
func ValidColumnName(s string) bool {
	for i,c := range s {
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
