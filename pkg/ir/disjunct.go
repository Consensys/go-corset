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
package ir

import (
	"github.com/consensys/go-corset/pkg/trace"
)

// Disjunct erpresents the logical OR of zero or more terms.  Observe that if
// there are no terms, then this is equivalent to logical falsehood.
type Disjunct[T LogicalTerm[T]] struct {
	disjuncts []T
}

// TestAt implementation for Testable interface.
func (p *Disjunct[T]) TestAt(k int, tr trace.Module) (bool, error) {
	//
	for _, disjunct := range p.disjuncts {
		val, _, err := disjunct.TestAt(k, tr)
		//
		if err != nil {
			return val, err
		} else if val {
			// Success
			return val, nil
		}
	}
	// Failure
	return false, nil
}
