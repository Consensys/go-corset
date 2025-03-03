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
package air

import (
	"fmt"
	"reflect"
)

// shiftTerm shifts all variable accesses in a given term by a given amount.
// This can be used to normalise shifting in certain circumstances.
func shiftTerm(e Term, shift int) Term {
	//
	switch e := e.(type) {
	case *Add:
		return &Add{Args: shiftTerms(e.Args, shift)}
	case *Constant:
		return e
	case *ColumnAccess:
		return &ColumnAccess{Column: e.Column, Shift: e.Shift + shift}
	case *Mul:
		return &Mul{Args: shiftTerms(e.Args, shift)}
	case *Sub:
		return &Sub{Args: shiftTerms(e.Args, shift)}
	default:
		name := reflect.TypeOf(e).Name()
		panic(fmt.Sprintf("unknown AIR expression \"%s\"", name))
	}
}

func shiftTerms(terms []Term, shift int) []Term {
	nterms := make([]Term, len(terms))
	//
	for i := range terms {
		nterms[i] = shiftTerm(terms[i], shift)
	}
	//
	return nterms
}
