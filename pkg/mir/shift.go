// Copyright Consensys Software Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with
// the License. You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
package mir

import (
	"fmt"
	"math"
	"reflect"
)

// ShiftRangeOfTerm returns the minimum and maximum shift value used anywhere in
// the given term.
func shiftRangeOfTerm(e Term) (int, int) {
	//
	switch e := e.(type) {
	case *Add:
		return shiftRangeOfTerms(e.Args)
	case *Cast:
		return shiftRangeOfTerm(e.Arg)
	case *Constant:
		return math.MaxInt, math.MinInt
	case *ColumnAccess:
		return e.Shift, e.Shift
	case *Exp:
		return shiftRangeOfTerm(e.Arg)
	case *Mul:
		return shiftRangeOfTerms(e.Args)
	case *Norm:
		return shiftRangeOfTerm(e.Arg)
	case *Sub:
		return shiftRangeOfTerms(e.Args)
	default:
		name := reflect.TypeOf(e).Name()
		panic(fmt.Sprintf("unknown MIR expression \"%s\"", name))
	}
}

func shiftRangeOfTerms(terms []Term) (int, int) {
	minShift := math.MaxInt
	maxShift := math.MinInt
	//
	for _, t := range terms {
		tMin, tMax := shiftRangeOfTerm(t)
		minShift = min(minShift, tMin)
		maxShift = max(maxShift, tMax)
	}
	//
	return minShift, maxShift
}

func shiftRangeOfEquations(equations []Equation) (int, int) {
	minShift := math.MaxInt
	maxShift := math.MinInt
	// Do left-hand sides
	for _, eq := range equations {
		tMin, tMax := shiftRangeOfTerm(eq.lhs)
		minShift = min(minShift, tMin)
		maxShift = max(maxShift, tMax)
	}
	// do right-hand sides
	for _, eq := range equations {
		tMin, tMax := shiftRangeOfTerm(eq.rhs)
		minShift = min(minShift, tMin)
		maxShift = max(maxShift, tMax)
	}
	//
	return minShift, maxShift
}
