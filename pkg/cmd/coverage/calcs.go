package coverage

import sc "github.com/consensys/go-corset/pkg/schema"

// DEFAULT_CALCS provides the detault set of calculations which can be used.
var DEFAULT_CALCS []ColumnCalc = []ColumnCalc{}

// ColumnCalc represents a calculation which can be done for a given constraint.
type ColumnCalc struct {
	Name        string
	Constructor func([]ConstraintId, sc.CoverageMap, sc.Schema) uint
}
