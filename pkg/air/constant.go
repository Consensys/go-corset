package air

import (
	"fmt"
	"reflect"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
)

func constantOfTerm(e Term) *fr.Element {
	switch e := e.(type) {
	case *Add:
		return nil
	case *Constant:
		return &e.Value
	case *ColumnAccess:
		return nil
	case *Sub:
		return nil
	case *Mul:
		return nil
	default:
		name := reflect.TypeOf(e).Name()
		panic(fmt.Sprintf("unknown AIR expression \"%s\"", name))
	}
}
