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
package hir

import (
	"fmt"
	"reflect"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	sc "github.com/consensys/go-corset/pkg/schema"
)

func substituteConstraint(mapping map[string]fr.Element, constraint sc.Constraint) {
	switch e := constraint.(type) {
	case LookupConstraint:
		// Substitute through source expressions
		for _, source := range e.Sources {
			substituteExpression(mapping, source)
		}
		// Substitute through target expressions
		for _, target := range e.Targets {
			substituteExpression(mapping, target)
		}
	case RangeConstraint:
		substituteExpression(mapping, e.Expr)
	case SortedConstraint:
		if e.Selector.HasValue() {
			substituteExpression(mapping, e.Selector.Unwrap())
		}
		// Substitute through source expressions
		for _, source := range e.Sources {
			substituteExpression(mapping, source)
		}
	case VanishingConstraint:
		substituteExpression(mapping, e.Constraint)
	default:
		name := reflect.TypeOf(e)
		panic(fmt.Sprintf("unknown HIR constraint \"%s\"", name.String()))
	}
}

func substituteExpression(mapping map[string]fr.Element, e Expr) {
	substituteTerm(mapping, e.Term)
}

func substituteTerm(mapping map[string]fr.Element, e Term) {
	switch e := e.(type) {
	case *Add:
		substituteTerms(mapping, e.Args)
	case *Cast:
		substituteTerm(mapping, e.Arg)
	case *Constant:
		// do not
	case *ColumnAccess:
		// do nout
	case *Exp:
		substituteTerm(mapping, e.Arg)
	case *IfZero:
		substituteTerm(mapping, e.Condition)
		// Subsitute true branch (if applicable)
		if e.TrueBranch != nil {
			substituteTerm(mapping, e.TrueBranch)
		}
		// Subsitute false branch (if applicable)
		if e.FalseBranch != nil {
			substituteTerm(mapping, e.FalseBranch)
		}
	case *LabelledConstant:
		// Attempt to apply substitution
		if nval, ok := mapping[e.Label]; ok {
			e.Value = nval
		}
	case *List:
		substituteTerms(mapping, e.Args)
	case *Mul:
		substituteTerms(mapping, e.Args)
	case *Norm:
		substituteTerm(mapping, e.Arg)
	case *Sub:
		substituteTerms(mapping, e.Args)
	default:
		name := reflect.TypeOf(e).Name()
		panic(fmt.Sprintf("unknown HIR expression \"%s\"", name))
	}
}

func substituteTerms(mapping map[string]fr.Element, terms []Term) {
	for _, term := range terms {
		substituteTerm(mapping, term)
	}
}
