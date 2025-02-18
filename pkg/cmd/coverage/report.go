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
package coverage

import (
	"fmt"

	sc "github.com/consensys/go-corset/pkg/schema"
)

// Report provides a raw form of coverage data which can subsequently be
// manipulated to provide a "user friendly" printed report.
type Report struct {
	// Schema for which coverage is being generated
	schema sc.Schema
	// Raw coverage data
	coverage sc.CoverageMap
	// Set of column calculations being used to construct report.
	calcs []ColumnCalc
}

// NewReport creates a new report for a given set of calculated metrics,
// coverage data and schema.
func NewReport(calcs []ColumnCalc, coverage sc.CoverageMap, schema sc.Schema) *Report {
	return &Report{schema, coverage, calcs}
}

// Row returns the formatted columns for a given constraint group.
func (p *Report) Row(group ConstraintGroup) []string {
	row := make([]string, len(p.calcs))
	//
	constraints := group.Select(p.schema)
	//
	for i, calc := range p.calcs {
		value := calc.Constructor(constraints, p.coverage, p.schema)
		row[i] = fmt.Sprintf("%d", value)
	}
	// fill in data here :)
	return row
}
