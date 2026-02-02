package schema

import "math/big"

type Map interface {
// Read a set of values at a given address on a bus.  The exact meaning of
// this depends upon the I/O peripheral connected to the bus.  For example,
// if its a function then the function is executed with the given address as
// its arguments, producing some number of outputs.  Likewise, if its a
// memory, then this will return the current value stored in that address,
// etc.
Read(bus uint, address []big.Int) []big.Int
// Write a set of values to a given address on a bus.  This only makes sense
// for writeable memory, such Random Access Memory (RAM).  In contrast,
// functions and Read-Only Memory (ROM) are not considered writeable.
Write(bus uint, address []big.Int, values []big.Int)
}