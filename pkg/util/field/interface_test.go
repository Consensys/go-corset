package field

import (
	"math/rand"
	"testing"

	"github.com/consensys/go-corset/pkg/util/assert"
	bls12_377 "github.com/consensys/go-corset/pkg/util/field/bls12-377"
	"github.com/consensys/go-corset/pkg/util/field/koalabear"
)

func init() {
	// make sure the interface is adhered to.
	_ = Element[koalabear.Element](koalabear.Element{})
	_ = Element[bls12_377.Element](bls12_377.Element{})
}

func TestBatchInvert(t *testing.T) {
	s := make([]koalabear.Element, 4000)
	sInv := make([]koalabear.Element, len(s))
	scratch := make([]koalabear.Element, len(s))

	for i := range s {
		s[i] = koalabear.Element{rand.Uint32()}
		if s[i][0] >= koalabear.Modulus {
			s[i][0] = 0 // getting a zero with considerable probability
		}

		sInv[i] = s[i].Inverse()

		copy(scratch[:i], s)
		BatchInvert(scratch[:i])

		for j := range i {
			assert.Equal(t, sInv[j][0], scratch[j][0], "on slice %v, at index %d", s[:i], j)
		}
	}
}
