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
package bexp

import "math/big"

// Term represents an abstraction over boolean expressions.
type Term[T any] interface {
	// Set this term to a new variable.
	Variable(string) T
	// Set this term to a new constant.
	Number(big.Int) T
	// Logical
	Or(...T) T
	// And(...Proposition) Proposition
	// Not(...Proposition) Proposition
	// Relational
	Equals(T) T
	NotEquals(T) T
	// Arithmetic
}

/*
type Arithmetic interface {
	Const(big.Int) Arithmetic
	Add(...Arithmetic) Arithmetic
	Mul(...Arithmetic) Arithmetic
	//
	Equals(Arithmetic) Proposition
	NotEquals(Arithmetic) Proposition
	LessThan(Arithmetic) Proposition
	LessThanOrEqual(Arithmetic) Proposition
}
*/
