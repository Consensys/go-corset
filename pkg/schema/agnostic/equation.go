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
package agnostic

import sc "github.com/consensys/go-corset/pkg/schema"

// Equation provides a generic notion of an equation between two polynomials.
// An equation is in *balanced form* if neither side contains a negative
// coefficient.
type Equation struct {
	// Left hand side.
	LeftHandSide Polynomial
	// Right hand side.
	RightHandSide Polynomial
}

// NewEquation simply constructs a new equation.
func NewEquation(lhs Polynomial, rhs Polynomial) Equation {
	return Equation{lhs, rhs}
}

// Split an equation according to a given field bandwidth.  This creates one
// or more equations implementing the original which operate safely within the
// given bandwidth.
func (p *Equation) Split(bandwidth uint, env sc.RegisterAllocator) []Equation {
	return []Equation{*p}
}
