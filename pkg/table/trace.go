package table

import (
	"encoding/json"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/util"
)

// Acceptable represents an element which can "accept" a trace, or either reject
// with an error (or eventually perhaps report a warning).
type Acceptable interface {
	Accepts(Trace) error
}

// Column describes an individual column of data within a trace table.
type Column interface {
	// Get the name of this column
	Name() string
	// Return the height (i.e. number of rows) of this column.
	Height() uint
	// Return the raw data stored in this column.
	Data() []*fr.Element
	// Return the value to use for padding this column.
	Padding() *fr.Element
	// Get the value at a given row in this column.  If the row is
	// out-of-bounds, then the column's padding value is returned instead.
	// Thus, this function always succeeds.
	Get(row int) *fr.Element
}

// Trace describes a set of named columns.  Columns are not required to have the
// same height and can be either "data" columns or "computed" columns.
type Trace interface {
	// Attempt to align this trace with a given schema.  This means ensuring the
	// order of columns in this trace matches the order in the schema.  Thus,
	// column indexes used by constraints in the schema can directly access in
	// this trace (i.e. without name lookup).  Alignment can fail, however, if
	// there is a mismatch between columns in the trace and those expected by
	// the schema.
	AlignWith(schema Schema) error
	// Add a new column of data
	AddColumn(name string, data []*fr.Element, padding *fr.Element)
	// ColumnByIndex returns the ith column in this trace.
	ColumnByIndex(uint) Column
	// ColumnByName returns the data of a given column in order that it can be
	// inspected.  If the given column does not exist, then nil is returned.
	ColumnByName(name string) Column
	// Check whether this trace contains data for the given column.
	HasColumn(name string) bool
	// Pad each column in this trace with n items at the front.  An iterator over
	// the padding values to use for each column must be given.
	Pad(n uint)
	// Determine the height of this table, which is defined as the
	// height of the largest column.
	Height() uint
	// Get the number of columns in this trace.
	Width() uint
}

// ConstraintsAcceptTrace determines whether or not one or more groups of
// constraints accept a given trace.  It returns the first error or warning
// encountered.
func ConstraintsAcceptTrace[T Acceptable](trace Trace, constraints []T) error {
	for _, c := range constraints {
		err := c.Accepts(trace)
		if err != nil {
			return err
		}
	}
	//
	return nil
}

// ===================================================================
// JSON Parser
// ===================================================================

// ParseJsonTrace parses a trace expressed in JSON notation.  For example, {"X":
// [0], "Y": [1]} is a trace containing one row of data each for two columns "X"
// and "Y".
func ParseJsonTrace(bytes []byte) (*ArrayTrace, error) {
	var zero fr.Element = fr.NewElement((0))

	var rawData map[string][]*big.Int
	// Unmarshall
	jsonErr := json.Unmarshal(bytes, &rawData)
	if jsonErr != nil {
		return nil, jsonErr
	}

	trace := EmptyArrayTrace()

	for name, rawInts := range rawData {
		// Translate raw bigints into raw field elements
		rawElements := util.ToFieldElements(rawInts)
		// Add new column to the trace
		trace.AddColumn(name, rawElements, &zero)
	}
	// Done.
	return trace, nil
}
