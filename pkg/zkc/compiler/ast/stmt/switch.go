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
package stmt

import (
	"strings"

	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/expr"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/symbol"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/variable"
)

// SwitchBranch represents a given branch.
//
// Note: the Cases slice should only contain Constants, no variables
type SwitchBranch[S symbol.Symbol[S]] struct {
	IsDefault bool
	Cases     []expr.Expr[S]
	Body      []Stmt[S]
}

// Switch represents a switch block of the form:
//
//	switch (discr) {
//		case a, b, ..., z: { branch_az }	// 1st case
//		case A, B, ..., Z: { branch_AZ }	// 2nd case, etc ...
//		default: { branch_default }		// optional default branch
//	}
type Switch[S symbol.Symbol[S]] struct {
	// Discriminant dictates the case
	Discriminant expr.Expr[S]
	// Branches contains all the non default branches
	Branches []SwitchBranch[S]
}

// Uses implementation for Stmt interface.
func (p *Switch[S]) Uses() []variable.Id {
	var reads []variable.Id
	// Collect variables read by the argument
	bits := p.Discriminant.LocalUses()
	for iter := bits.Iter(); iter.HasNext(); {
		reads = append(reads, iter.Next())
	}

	// Collect variables from non default branches
	for _, branch := range p.Branches {
		for _, statement := range branch.Body {
			reads = append(reads, statement.Uses()...)
		}
	}

	return reads
}

// Definitions implementation for Stmt interface.
func (p *Switch[S]) Definitions() []variable.Id {
	var writes []variable.Id

	for _, branch := range p.Branches {
		for _, statement := range branch.Body {
			writes = append(writes, statement.Definitions()...)
		}
	}

	return writes
}

func (p *Switch[S]) String(env variable.Map[S]) string {
	var b strings.Builder
	b.WriteString("switch (")
	b.WriteString(p.Discriminant.String(env))
	b.WriteString(") {\n\t case _: { ... }\n\t default: { ... } }")

	return b.String()
}
