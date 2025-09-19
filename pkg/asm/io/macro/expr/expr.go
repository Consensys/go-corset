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
package expr

import (
	"encoding/gob"
	"fmt"
	"math/big"
	"strings"

	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/agnostic"
	"github.com/consensys/go-corset/pkg/util/collection/bit"
	"github.com/consensys/go-corset/pkg/util/math"
)

var (
	biZERO big.Int = *big.NewInt(0)
	biONE  big.Int = *big.NewInt(1)
)

// Expr represents an arbitrary expression used within an instruction.
type Expr interface {
	// Evaluate this expression in a given environment producing a given value.
	Eval([]big.Int) big.Int
	// Polynomial returns this expression flatterned into a polynomial form.
	Polynomial() agnostic.Polynomial
	// RegistersRead returns the set of registers read by this expression
	RegistersRead() bit.Set
	// String returns a string representation of this expression in a given base.
	String(mapping schema.RegisterMap) string
	// ValueRange returns the interval of values that this term can evaluate to.
	// For terms accessing registers, this is determined by the declared width of
	// the register.
	ValueRange(mapping schema.RegisterMap) math.Interval
}

// String provides a generic facility for converting an expression into a
// suitable string.
func String(e Expr, mapping schema.RegisterMap) string {
	var (
		exprs    []Expr
		operator string
		builder  strings.Builder
	)
	//
	switch e := e.(type) {
	case *Add:
		operator = "+"
		exprs = e.Exprs
	case *Const:
		return stringOfConstant(e.Constant, e.Base)
	case *Mul:
		exprs = e.Exprs
		operator = "*"
	case *RegAccess:
		return mapping.Register(e.Register).Name
	case *Sub:
		exprs = e.Exprs
		operator = "-"
	default:
		panic("unreachable")
	}
	//
	for i, e := range exprs {
		if i != 0 {
			builder.WriteString(" ")
			builder.WriteString(operator)
			builder.WriteString(" ")
		}
		//
		if needsBraces(e) {
			builder.WriteString("(")
			builder.WriteString(String(e, mapping))
			builder.WriteString(")")
		} else {
			builder.WriteString(String(e, mapping))
		}
	}
	//
	return builder.String()
}

func stringOfConstant(val big.Int, base uint) string {
	switch base {
	case 2:
		return fmt.Sprintf("0b%s", val.Text(2))
	case 16:
		return fmt.Sprintf("0x%s", val.Text(16))
	default:
		return val.String()
	}
}

func needsBraces(e Expr) bool {
	switch e.(type) {
	case *Const:
		return false
	case *RegAccess:
		return false
	default:
		return true
	}
}

func init() {
	gob.Register(Expr(&Add{}))
	gob.Register(Expr(&Const{}))
	gob.Register(Expr(&Mul{}))
	gob.Register(Expr(&RegAccess{}))
	gob.Register(Expr(&Sub{}))
}
