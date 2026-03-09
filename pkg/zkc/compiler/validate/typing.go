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
	"reflect"

	"github.com/consensys/go-corset/pkg/util/source"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/data"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/expr"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/lval"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/stmt"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/symbol"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/variable"
)

// Stmt is a convenient alias.
type Stmt = stmt.Stmt[symbol.Resolved]

// LVal is a convenient alias.
type LVal = lval.LVal[symbol.Resolved]

// Type is a convenient alias.
type Type = data.Type[symbol.Resolved]

// VariableMap is a convenient alias.
type VariableMap = variable.Map[symbol.Resolved]

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
		env    data.Environment[symbol.Resolved]
		errors []source.SyntaxError
		typer  = TypeChecker{program, env, srcmaps}
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
	env     data.Environment[symbol.Resolved]
	srcmaps source.Maps[any]
}

func (p *TypeChecker) lookup(id symbol.Resolved) ast.Declaration {
	return p.program.Component(id.Index)
}

func (p *TypeChecker) typeConstant(c ast.Constant) []source.SyntaxError {
	var (
		rhs, errors = p.typeExpression(c.ConstExpr, variable.ArrayMap[symbol.Resolved]())
	)
	// Sanity check
	if len(errors) != 0 {
		return errors
	} else if !data.SubtypeOf(rhs, c.DataType, p.env) {
		return p.srcmaps.SyntaxErrors(c.ConstExpr, fmt.Sprintf("%s not subtype of %s", rhs.String(), c.DataType.String()))
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

func (p *TypeChecker) typeAssignment(s *stmt.Assign[symbol.Resolved], env VariableMap) []source.SyntaxError {
	var (
		errors  []source.SyntaxError
		sources = []expr.Expr[symbol.Resolved]{s.Source}
	)
	// Sanity check assignment arity
	if len(s.Targets) < len(sources) {
		return p.srcmaps.SyntaxErrors(s, fmt.Sprintf("insufficient target variables (expected %d)", len(sources)))
	} else if len(s.Targets) > len(sources) {
		return p.srcmaps.SyntaxErrors(s, fmt.Sprintf("too many target variables (expected %d)", len(sources)))
	}
	// Check each in turn
	for i, lval := range s.Targets {
		var (
			rhs             = sources[i]
			lval_t, lhsErrs = p.typeLval(lval, env)
			rhs_t, rhsErrs  = p.typeExpression(rhs, env)
		)
		//
		if len(lhsErrs) != 0 || len(rhsErrs) != 0 {
			errors = append(errors, lhsErrs...)
			errors = append(errors, rhsErrs...)
		} else if !data.SubtypeOf(rhs_t, lval_t, p.env) {
			err := *p.srcmaps.SyntaxError(rhs, fmt.Sprintf("cannot use %s as %s in assignment", rhs_t, lval_t))
			errors = append(errors, err)
		}
	}
	//
	return append(errors, checkTargets(s, env, p.srcmaps)...)
}

func (p *TypeChecker) typeLval(target LVal, env VariableMap) (Type, []source.SyntaxError) {
	// determine lhs width
	switch t := target.(type) {
	case *lval.Variable[symbol.Resolved]:
		return env.Variable(t.Id).DataType, nil
	case *lval.MemAccess[symbol.Resolved]:
		panic("todo")
	default:
		return nil, p.srcmaps.SyntaxErrors(target, "unknown lval")
	}
}

// CheckTargetRegisters performs some simple checks on a set of target registers
// being written.  Firstly, they cannot be input registers (as this are always
// constant).  Secondly, we cannot write to the same register more than once
// (i.e. a conflicting write).
func checkTargets(s *stmt.Assign[symbol.Resolved], env VariableMap, srcmaps source.Maps[any]) []source.SyntaxError {
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

func (p *TypeChecker) typeIfGoto(s *stmt.IfGoto[symbol.Resolved], env VariableMap) []source.SyntaxError {
	return p.typeCondition(s.Cond, env)
}

func (p *TypeChecker) typeCondition(e ast.Condition, env VariableMap) []source.SyntaxError {
	switch e := e.(type) {
	case *expr.Cmp[symbol.Resolved]:
		return p.typeCmp(e, env)
	default:
		return p.srcmaps.SyntaxErrors(e, "unknown condition")
	}
}

func (p *TypeChecker) typeCmp(e *expr.Cmp[symbol.Resolved], env VariableMap) []source.SyntaxError {
	var (
		lhs, lerrs = p.typeExpression(e.Left, env)
		rhs, rerrs = p.typeExpression(e.Right, env)
	)
	// Check left-hand side
	if len(lerrs) == 0 && lhs.AsUint() == nil {
		lerrs = p.srcmaps.SyntaxErrors(e.Left, "expected uint")
	}
	// Check right-hand side
	if len(rerrs) == 0 && rhs.AsUint() == nil {
		rerrs = p.srcmaps.SyntaxErrors(e.Right, "expected uint")
	}
	// Check matching types
	if len(lerrs)+len(rerrs) == 0 && !data.SubtypeOf(lhs, rhs, p.env) && !data.SubtypeOf(rhs, lhs, p.env) {
		return p.srcmaps.SyntaxErrors(e.Right, fmt.Sprintf("expected type %s", lhs.String()))
	}
	//
	return append(lerrs, rerrs...)
}

func (p *TypeChecker) typeExpression(e ast.Expr, env VariableMap) (t Type, errs []source.SyntaxError) {
	switch e := e.(type) {
	case *expr.Add[symbol.Resolved]:
		t, errs = p.typeArithmeticExpression(e.Exprs, env)
	case *expr.Const[symbol.Resolved]:
		t, errs = p.typeConst(e, env)
	case *expr.LocalAccess[symbol.Resolved]:
		t, errs = p.typeLocalAccess(e, env)
	case *expr.Mul[symbol.Resolved]:
		t, errs = p.typeArithmeticExpression(e.Exprs, env)
	case *expr.ExternAccess[symbol.Resolved]:
		t, errs = p.typeExternAccess(e, env)
	case *expr.Sub[symbol.Resolved]:
		t, errs = p.typeArithmeticExpression(e.Exprs, env)
	default:
		return nil, p.srcmaps.SyntaxErrors(e, "unknown expression")
	}
	// Associate type
	e.SetType(t)
	//
	return t, errs
}

func (p *TypeChecker) typeExpressions(exprs []ast.Expr, env VariableMap) ([]Type, []source.SyntaxError) {
	var (
		types  = make([]Type, len(exprs))
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

func (p *TypeChecker) typeArithmeticExpression(exprs []ast.Expr, env VariableMap) (Type, []source.SyntaxError) {
	var (
		args, errs = p.typeExpressions(exprs, env)
		res        Type
	)
	//
	if len(errs) > 0 {
		return nil, errs
	}
	//
	for i, t := range args {
		if i == 0 && t.AsUint() == nil {
			return nil, p.srcmaps.SyntaxErrors(exprs[i], "expected uint")
		} else if i == 0 {
			res = t
		} else if !data.SubtypeOf(res, t, p.env) && !data.SubtypeOf(t, res, p.env) {
			return nil, p.srcmaps.SyntaxErrors(exprs[i], fmt.Sprintf("expected type %s", res.String()))
		}
	}
	//
	return res, nil
}

func (p *TypeChecker) typeConst(e *expr.Const[symbol.Resolved], env VariableMap) (Type, []source.SyntaxError) {
	var (
		bitwidth = uint(e.Constant.BitLen())
	)
	//
	return data.NewUnsignedInt[symbol.Resolved](bitwidth, true), nil
}

func (p *TypeChecker) typeLocalAccess(e *expr.LocalAccess[symbol.Resolved], env VariableMap,
) (Type, []source.SyntaxError) {
	//
	return env.Variable(e.Variable).DataType, nil
}

func (p *TypeChecker) typeExternAccess(e *expr.ExternAccess[symbol.Resolved], env VariableMap,
) (Type, []source.SyntaxError) {
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
func (p *TypeChecker) typeConstantAccess(c *ast.Constant, e *expr.ExternAccess[symbol.Resolved], env VariableMap,
) (Type, []source.SyntaxError) {
	return c.DataType, nil
}

func (p *TypeChecker) typeMemoryAccess(c *ast.Memory, e *expr.ExternAccess[symbol.Resolved], env VariableMap,
) (Type, []source.SyntaxError) {
	var args, errs = p.typeExpressions(e.Args, env)
	//
	if len(args) != len(c.Address) {
		return nil, p.srcmaps.SyntaxErrors(e,
			fmt.Sprintf("mismatched arguments (expected %d, found %d)", len(c.Address), len(args)))
	} else if len(errs) == 0 {
		// check argument types
		for i := range args {
			ith := c.Address[i].DataType
			if !data.SubtypeOf(args[i], ith, p.env) {
				errs = append(errs, *p.srcmaps.SyntaxError(e.Args[i], fmt.Sprintf("expected type %s", ith.String())))
			}
		}
	}
	// Done
	return variable.DescriptorsToType(c.Data...), errs
}

func (p *TypeChecker) typeFunctionAccess(c *ast.Function, e *expr.ExternAccess[symbol.Resolved], env VariableMap,
) (Type, []source.SyntaxError) {
	panic("todo --- function accesses")
}
