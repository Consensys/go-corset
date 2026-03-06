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
package validate

import (
	"fmt"
	"math/big"
	"reflect"

	"github.com/consensys/go-corset/pkg/util/source"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/expr"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/lval"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/stmt"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/symbol"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/variable"
	"github.com/consensys/go-corset/pkg/zkc/compiler/validate/typing"
)

// Stmt is a convenient alias.
type Stmt = stmt.Stmt[symbol.Resolved]

// LVal is a convenient alias.
type LVal = lval.LVal[symbol.Resolved]

// Typing validates that each declaration in a program is "correctly typed". For
// example, the following constant declaration is ill-typed:
//
// constant u8 MAX_U8 = 256
//
// The reason being that the constant's value does not fit in a u8.  Likewise,
// this set of declarations are ill-typed:
//
//	 public input u8[u8] ROM
//
//		function f(val u8) -> (r u10) {
//		  r = ROM + 1
//		  return
//		}
//
// The problem above is that ROM is an input memory and, hence, ROM+1 does not
// make sense.
func Typing(program ast.Program, srcmaps source.Maps[any]) []source.SyntaxError {
	var (
		errors []source.SyntaxError
		typer  = TypeChecker{program, srcmaps}
	)
	//
	for _, d := range program.Components() {
		switch d := d.(type) {
		case *ast.Constant:
			errors = append(errors, typer.typeConstant(*d)...)
		case *ast.Function:
			errors = append(errors, typer.typeFunction(*d)...)
		case *ast.Memory:
			// ignore
		default:
			panic(fmt.Sprintf("unknown component: %s", reflect.TypeOf(d).String()))
		}
	}
	//
	return errors
}

// TypeChecker embodies information needed for type checking a given program.
type TypeChecker struct {
	program ast.Program
	srcmaps source.Maps[any]
}

func (p *TypeChecker) lookup(id symbol.Resolved) ast.Declaration {
	return p.program.Component(id.Index)
}

func (p *TypeChecker) typeConstant(c ast.Constant) []source.SyntaxError {
	var (
		lhs_bits    = c.DataType.BitWidth()
		rhs, errors = p.typeExpression(c.ConstExpr, variable.ArrayMap())
	)
	// Sanity check
	if len(errors) != 0 {
		return errors
	} else if rhs_bits := rhs.AsUint().BitWidth(); lhs_bits < rhs_bits {
		return p.srcmaps.SyntaxErrors(c.ConstExpr, fmt.Sprintf("bit overflow (u%d into u%d)", rhs_bits, lhs_bits))
	}
	//
	return nil
}

func (p *TypeChecker) typeFunction(fn ast.Function) []source.SyntaxError {
	var errors []source.SyntaxError

	for _, s := range fn.Code {
		switch s := s.(type) {
		case *stmt.Assign[symbol.Resolved]:
			errors = append(errors, p.typeAssignment(s, &fn)...)
		case *stmt.IfGoto[symbol.Resolved]:
			errors = append(errors, p.typeIfGoto(s, &fn)...)
		}
	}
	//
	return errors
}

func (p *TypeChecker) typeAssignment(s *stmt.Assign[symbol.Resolved], env variable.Map) []source.SyntaxError {
	var (
		lhs_bits    uint
		rhs, errors = p.typeExpression(s.Source, env)
	)
	// determine lhs width
	for _, target := range s.Targets {
		switch t := target.(type) {
		case *lval.Variable[symbol.Resolved]:
			lhs_bits += env.Variable(t.Id).BitWidth()
		case *lval.MemAccess[symbol.Resolved]:
			panic("todo")
		default:
			panic("unknown lval encountered")
		}
	}
	// check
	if len(errors) != 0 {
		return errors
	} else if rhs_bits := rhs.AsUint().BitWidth(); lhs_bits < rhs_bits {
		return p.srcmaps.SyntaxErrors(s, fmt.Sprintf("bit overflow (u%d into u%d)", rhs_bits, lhs_bits))
	}
	//
	return checkTargets(s, env, p.srcmaps)
}

// CheckTargetRegisters performs some simple checks on a set of target registers
// being written.  Firstly, they cannot be input registers (as this are always
// constant).  Secondly, we cannot write to the same register more than once
// (i.e. a conflicting write).
func checkTargets(s *stmt.Assign[symbol.Resolved], env variable.Map, srcmaps source.Maps[any]) []source.SyntaxError {
	var targets []variable.Id
	//
	for _, id := range s.Targets {
		targets = append(targets, lval.Definitions(id)...)
	}
	//
	for i, id := range targets {
		ith := env.Variable(id)
		//
		if ith.IsParameter() {
			return srcmaps.SyntaxErrors(s, fmt.Sprintf("cannot write parameter %s", ith.Name))
		}
		//
		for j := i + 1; j < len(targets); j++ {
			if targets[i] == targets[j] {
				return srcmaps.SyntaxErrors(s, fmt.Sprintf("conflicting write to %s", ith.Name))
			}
		}
	}
	//
	return nil
}

func (p *TypeChecker) typeIfGoto(s *stmt.IfGoto[symbol.Resolved], env variable.Map) []source.SyntaxError {
	return p.typeCondition(s.Cond, env)
}

func (p *TypeChecker) typeCondition(e ast.Condition, env variable.Map) []source.SyntaxError {
	switch e := e.(type) {
	case *expr.Cmp[symbol.Resolved]:
		return p.typeCmp(e, env)
	default:
		return p.srcmaps.SyntaxErrors(e, "unknown condition")
	}
}

func (p *TypeChecker) typeCmp(e *expr.Cmp[symbol.Resolved], env variable.Map) []source.SyntaxError {
	var (
		lhs, lerrs = p.typeExpression(e.Left, env)
		rhs, rerrs = p.typeExpression(e.Right, env)
	)
	// Check left-hand side
	if lerrs == nil && lhs.AsUint() == nil {
		lerrs = p.srcmaps.SyntaxErrors(e.Left, "expected uint")
	}
	// Check right-hand side
	if rerrs == nil && rhs.AsUint() == nil {
		rerrs = p.srcmaps.SyntaxErrors(e.Right, "expected uint")
	}
	//
	return append(lerrs, rerrs...)
}

func (p *TypeChecker) typeExpression(e ast.Expr, env variable.Map) (typing.Type, []source.SyntaxError) {
	switch e := e.(type) {
	case *expr.Add[symbol.Resolved]:
		return p.typeAdd(e, env)
	case *expr.Const[symbol.Resolved]:
		return p.typeConst(e, env)
	case *expr.LocalAccess[symbol.Resolved]:
		return p.typeLocalAccess(e, env)
	case *expr.Mul[symbol.Resolved]:
		return p.typeMul(e, env)
	case *expr.ExternAccess[symbol.Resolved]:
		return p.typeExternAccess(e, env)
	case *expr.Sub[symbol.Resolved]:
		return p.typeSub(e, env)
	default:
		return nil, p.srcmaps.SyntaxErrors(e, "unknown expression")
	}
}

func (p *TypeChecker) typeExpressions(exprs []ast.Expr, env variable.Map) ([]typing.Type, []source.SyntaxError) {
	var (
		types  = make([]typing.Type, len(exprs))
		errors []source.SyntaxError
	)
	//
	for i, e := range exprs {
		var errs []source.SyntaxError
		//
		types[i], errs = p.typeExpression(e, env)
		//
		errors = append(errors, errs...)
	}
	//
	return types, errors
}

func (p *TypeChecker) typeAdd(e *expr.Add[symbol.Resolved], env variable.Map) (typing.Type, []source.SyntaxError) {
	var (
		args, errs = p.typeExpressions(e.Exprs, env)
		max        *typing.Uint
	)
	//
	if len(errs) > 0 {
		return nil, errs
	}
	//
	for i, t := range args {
		if ut := t.AsUint(); ut == nil {
			return nil, p.srcmaps.SyntaxErrors(e.Exprs[i], "expected uint")
		} else if i == 0 {
			max = ut
		} else {
			max.Add(ut)
		}
	}
	//
	e.SetBitWidth(max.BitWidth())
	//
	return max, nil
}

func (p *TypeChecker) typeConst(e *expr.Const[symbol.Resolved], env variable.Map) (typing.Type, []source.SyntaxError) {
	return &typing.Uint{MaxValue: e.Constant}, nil
}

func (p *TypeChecker) typeLocalAccess(e *expr.LocalAccess[symbol.Resolved], env variable.Map,
) (typing.Type, []source.SyntaxError) {
	//
	var (
		bound    = big.NewInt(2)
		bitwidth = env.Variable(e.Variable).BitWidth()
	)
	// compute 2^bitwidth
	bound.Exp(bound, big.NewInt(int64(bitwidth)), nil)
	// Subtract 1 because interval is inclusive.
	bound.Sub(bound, big.NewInt(1))
	//
	e.SetBitWidth(bitwidth)
	//
	return &typing.Uint{MaxValue: *bound}, nil
}

func (p *TypeChecker) typeMul(e *expr.Mul[symbol.Resolved], env variable.Map) (typing.Type, []source.SyntaxError) {
	var (
		args, errs = p.typeExpressions(e.Exprs, env)
		max        *typing.Uint
	)
	//
	if len(errs) > 0 {
		return nil, errs
	}
	//
	for i, t := range args {
		if ut := t.AsUint(); ut == nil {
			return nil, p.srcmaps.SyntaxErrors(e.Exprs[i], "expected uint")
		} else if i == 0 {
			max = ut
		} else {
			max.Mul(ut)
		}
	}
	//
	e.SetBitWidth(max.BitWidth())
	//
	return max, nil
}

func (p *TypeChecker) typeExternAccess(e *expr.ExternAccess[symbol.Resolved], env variable.Map,
) (typing.Type, []source.SyntaxError) {
	// Lookup the symbol
	var extern = p.lookup(e.Name)
	// Decide what kind of symbol it is
	switch t := extern.(type) {
	case *ast.Constant:
		return p.typeConstantAccess(t, e, env)
	case *ast.Memory:
		return p.typeMemoryAccess(t, e, env)
	case *ast.Function:
		return p.typeFunctionAccess(t, e, env)
	default:
		return nil, p.srcmaps.SyntaxErrors(e, "unknown symbol type")
	}
}
func (p *TypeChecker) typeConstantAccess(c *ast.Constant, e *expr.ExternAccess[symbol.Resolved], env variable.Map,
) (typing.Type, []source.SyntaxError) {
	var bound = big.NewInt(2)
	// NOTE: no need to sanity check expected number of arguments, as this is
	// done during linking.
	bitwidth := c.DataType.BitWidth()
	// compute 2^bitwidth
	bound.Exp(bound, big.NewInt(int64(bitwidth)), nil)
	// Subtract 1 because interval is inclusive.
	bound.Sub(bound, big.NewInt(1))
	//
	e.SetBitWidth(bitwidth)
	//
	return &typing.Uint{MaxValue: *bound}, nil
}

func (p *TypeChecker) typeMemoryAccess(c *ast.Memory, e *expr.ExternAccess[symbol.Resolved], env variable.Map,
) (typing.Type, []source.SyntaxError) {
	// type arguments
	_, errors := p.typeExpressions(e.Args, env)
	// TODO: type check returns
	return typing.FromVariables(c.Data...), errors
}

func (p *TypeChecker) typeFunctionAccess(c *ast.Function, e *expr.ExternAccess[symbol.Resolved], env variable.Map,
) (typing.Type, []source.SyntaxError) {
	panic("todo --- function accesses")
}

func (p *TypeChecker) typeSub(e *expr.Sub[symbol.Resolved], env variable.Map) (typing.Type, []source.SyntaxError) {
	var (
		args, errs = p.typeExpressions(e.Exprs, env)
		max        *typing.Uint
		min        *typing.Uint
	)
	//
	if len(errs) > 0 {
		return nil, errs
	}
	//
	for i, t := range args {
		if ut := t.AsUint(); ut == nil {
			return nil, p.srcmaps.SyntaxErrors(e.Exprs[i], "expected uint")
		} else if i == 0 {
			max = ut
		} else if i == 1 {
			min = ut
		} else {
			min.Add(ut)
		}
	}
	//
	e.SetBitWidths(min.BitWidth(), max.BitWidth())
	//
	return max, nil
}
