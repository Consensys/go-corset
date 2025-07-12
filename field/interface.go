package field

import "fmt"

// An Element of a prime-order field.
type Element[Operand any] interface {
	Add(y Operand) Operand      // Add x+y
	Sub(y Operand) Operand      // Sub x-y
	AddUint32(y uint32) Operand // AddUint32 x+y. It's the canonical way to create a new element with value y.
	ToUint32() uint32           // ToUint32 returns the numerical value of x.
	Mul(y Operand) Operand      // Mul x*y
	Cmp(y Operand) int          // Cmp returns 1 if x > y, 0 if x = y, and -1 if x < y.
	Double() Operand            // Double 2x
	Half() Operand              // Half x/2
	Inverse() Operand           // Inverse x⁻¹, or 0 if x = 0.
	fmt.Stringer
	Text(base int) string // Text returns the numerical value of x in the given base.
}
