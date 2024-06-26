package json

import (
	"encoding/json"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
)

// FromBytes parses a trace expressed in JSON notation.  For example, {"X":
// [0], "Y": [1]} is a trace containing one row of data each for two columns "X"
// and "Y".
func FromBytes(bytes []byte) (*trace.ArrayTrace, error) {
	var zero fr.Element = fr.NewElement((0))

	var rawData map[string][]*big.Int
	// Unmarshall
	jsonErr := json.Unmarshal(bytes, &rawData)
	if jsonErr != nil {
		return nil, jsonErr
	}

	trace := trace.EmptyArrayTrace()

	for name, rawInts := range rawData {
		// Translate raw bigints into raw field elements
		rawElements := util.ToFieldElements(rawInts)
		// Add new column to the trace
		trace.AddColumn(name, rawElements, &zero)
	}
	// Done.
	return trace, nil
}
