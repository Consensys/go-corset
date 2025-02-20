package coverage

import (
	sc "github.com/consensys/go-corset/pkg/schema"
	log "github.com/sirupsen/logrus"
)

// DEFAULT_CALCS provides the detault set of calculations which can be used.
var DEFAULT_CALCS []ColumnCalc = []ColumnCalc{
	{"covered", coverageCalculator},
	{"branches", branchesCalculator},
	{"coverage", percentCalculator},
}

// ColumnCalc represents a calculation which can be done for a given constraint.
type ColumnCalc struct {
	Name        string
	Constructor func([]sc.Constraint, sc.CoverageMap, sc.Schema) uint
}

func branchesCalculator(constraints []sc.Constraint, cov sc.CoverageMap, schema sc.Schema) uint {
	total := uint(0)
	//
	for _, c := range constraints {
		total += c.Branches()
	}
	//
	return total
}

func coverageCalculator(constraints []sc.Constraint, cov sc.CoverageMap, schema sc.Schema) uint {
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
	return total
}

func percentCalculator(constraints []sc.Constraint, cov sc.CoverageMap, schema sc.Schema) uint {
	branches := branchesCalculator(constraints, cov, schema)
	covered := coverageCalculator(constraints, cov, schema)
	// Sanity check
	if branches == 0 {
		return 0
	}
	//
	return (covered * 100) / branches
}
