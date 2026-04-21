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

// SwitchCase represents a given branch
//
// Note: the Cases slice should only contain Constants, no variables
type SwitchCase[S symbol.Symbol[S]] struct {
	Cases []expr.Const[S]
	Body  []Stmt[S]
}

// Switch represents a switch block of the form:
//
//	switch (arg) {
//		case a, b, ..., z: { branch_az }    // one SwitchCase
//		case A, B, ..., Z: { branch_AZ }    // another SwitchCase
//		default: { branch_default }
//
// default must always be provided
type Switch[S symbol.Symbol[S]] struct {
	// Argument dictates the case
	Argument expr.Expr[S]
	// Branches contains all the non default branches
	Branches []SwitchCase[S]
	// DefaultBranch contains the default branch
	DefaultBranch SwitchCase[S]
}

// Uses implementation for Stmt interface.
func (p *Switch[S]) Uses() []variable.Id {
	var reads []variable.Id
	// Collect variables read by the argument
	bits := p.Argument.LocalUses()
	for iter := bits.Iter(); iter.HasNext(); {
		reads = append(reads, iter.Next())
	}

	// Collect variables from non default branches
	//
	for _, branch := range p.Branches {
		for _, statement := range branch.Body {
			reads = append(reads, statement.Uses()...)
		}
	}

	// Collect from the default branch
	for _, statement := range p.DefaultBranch.Body {
		reads = append(reads, statement.Uses()...)
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

	for _, statement := range p.DefaultBranch.Body {
		writes = append(writes, statement.Definitions()...)
	}

	return writes
}

func (p *Switch[S]) String(env variable.Map[S]) string {
	var b strings.Builder
	b.WriteString("switch (")
	b.WriteString(p.Argument.String(env))
	b.WriteString(") {\n\t case _: { ... }\n\t default: { ... } }")

	return b.String()
}
