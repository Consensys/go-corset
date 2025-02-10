package mir

import (
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/util"
)

// Expr represents an expression in the Mid-Level Intermediate Representation
// (MIR).  Expressions at this level have a one-2-one correspondance with
// expressions in the AIR level.  However, some expressions at this level do not
// exist at the AIR level (e.g. normalise) and are "compiled out" by introducing
// appropriate computed columns and constraints.
type Expr interface {
	util.Boundable
	sc.Evaluable

	// IntRange computes a conservative approximation for the set of possible
	// values that this expression can evaluate to.
	IntRange(schema sc.Schema) *util.Interval
}
