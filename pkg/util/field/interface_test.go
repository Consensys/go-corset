package field

import (
	bls12_377 "github.com/consensys/go-corset/pkg/util/field/bls12-377"
	"github.com/consensys/go-corset/pkg/util/field/koalabear"
)

func init() {
	// make sure the interface is adhered to.
	_ = Element[koalabear.Element](koalabear.Element{})
	_ = Element[bls12_377.Element](bls12_377.Element{})
}
