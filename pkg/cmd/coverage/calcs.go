package coverage

import (
	"fmt"

	sc "github.com/consensys/go-corset/pkg/schema"
	log "github.com/sirupsen/logrus"
)

// DEFAULT_CALCS provides the detault set of calculations which can be used.
var DEFAULT_CALCS []ColumnCalc = []ColumnCalc{
	{"covered", coveredCalc},
	{"branches", branchesCalc},
	{"coverage", coverageCalc},
}

// ColumnCalc represents a calculation which can be done for a given constraint.
type ColumnCalc struct {
	Name        string
	Constructor func([]sc.Constraint, sc.CoverageMap, sc.Schema) CalcValue
}

func branchesCalc(constraints []sc.Constraint, cov sc.CoverageMap, schema sc.Schema) CalcValue {
	return &IntegerValue{branchesCalculation(constraints)}
}

func coveredCalc(constraints []sc.Constraint, cov sc.CoverageMap, schema sc.Schema) CalcValue {
	return &IntegerValue{coveredCalculation(constraints, cov)}
}

func coverageCalc(constraints []sc.Constraint, cov sc.CoverageMap, schema sc.Schema) CalcValue {
	return &FloatValue{percentCalculation(constraints, cov)}
}

func branchesCalculation(constraints []sc.Constraint) int {
	total := uint(0)
	//
	for _, c := range constraints {
		total += c.Branches()
	}
	//
	return int(total)
}

func coveredCalculation(constraints []sc.Constraint, cov sc.CoverageMap) int {
	total := uint(0)
	//
	for _, c := range constraints {
		mid := c.Contexts()[0].Module()
		name, num := c.Name()
		bitsets := cov.CoverageOf(mid, name)
		// Check it lines up
		if num < uint(len(bitsets)) {
			// Determine coverage
			total += bitsets[num].Count()
		} else {
			log.Warnf("*** missing coverage data for constraint %s.%d\n", name, num)
		}
	}
	//
	return int(total)
}

func percentCalculation(constraints []sc.Constraint, cov sc.CoverageMap) float64 {
	branches := branchesCalculation(constraints)
	covered := coveredCalculation(constraints, cov)
	// Sanity check
	if branches == 0 {
		return 0
	}
	//
	return float64(100*covered) / float64(branches)
}

// CalcValue provides a wrapper around a specific kind of value.
type CalcValue interface {
	String() string
}

// IntegerValue is an example of a CalcValue which is just a plain number.
type IntegerValue struct {
	value int
}

func (p *IntegerValue) String() string {
	return fmt.Sprintf("%d", p.value)
}

// FloatValue is an example of a CalcValue which is just a plain number.
type FloatValue struct {
	value float64
}

func (p *FloatValue) String() string {
	return fmt.Sprintf("%0.1f", p.value)
}
