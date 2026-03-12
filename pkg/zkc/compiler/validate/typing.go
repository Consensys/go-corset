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
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/decl"
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
		errors []source.SyntaxError
		typer  = TypeChecker{program, program.Environment(), srcmaps}
	)
	//
	for _, d := range program.Components() {
		switch d := d.(type) {
		case *decl.ResolvedConstant:
			errors = append(errors, typer.typeConstant(*d)...)
		case *decl.ResolvedFunction:
			errors = append(errors, typer.typeFunction(*d)...)
		case *decl.ResolvedMemory:
			errors = append(errors, typer.typeMemory(*d)...)
		case *decl.ResolvedTypeAlias:
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
	env     data.ResolvedEnvironment
	srcmaps source.Maps[any]
}

func (p *TypeChecker) lookup(id symbol.Resolved) decl.Resolved {
	return p.program.Component(id.Index)
}

func (p *TypeChecker) typeConstant(c decl.ResolvedConstant) []source.SyntaxError {
	var (
		rhs, errors = p.typeExpression(c.ConstExpr, variable.ArrayMap[symbol.Resolved]())
	)
	// Sanity check
	if len(errors) != 0 {
		return errors
	}
	// Subtype check
	return p.checkEquiTypes(rhs, c.DataType, c.ConstExpr)
}

func (p *TypeChecker) typeFunction(fn decl.ResolvedFunction) []source.SyntaxError {
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

func (p *TypeChecker) typeMemory(c decl.ResolvedMemory) []source.SyntaxError {
	if !c.IsStatic() {
		return nil
	}

	var (
		errors   []source.SyntaxError
		dataType = variable.DescriptorsToType(c.Data...)
	)
	//
	for _, v := range c.Contents {
		valBitwidth := uint(v.BitLen())
		valType := data.NewUnsignedInt[symbol.Resolved](valBitwidth, true)
		errors = append(errors, p.checkEquiTypes(valType, dataType, v)...)
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
		} else {
			//
			errors = append(errors, p.checkEquiTypes(rhs_t, lval_t, rhs)...)
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
		// Lookup the symbol
		var extern = p.lookup(t.Name)
		//
		switch e := extern.(type) {
		case *decl.ResolvedConstant:
			return nil, p.srcmaps.SyntaxErrors(target, "cannot assign constant")
		case *decl.ResolvedMemory:
			return p.typeMemoryLVal(e, t, env)
		case *decl.ResolvedFunction:
			return nil, p.srcmaps.SyntaxErrors(target, "cannot assign function")
		case *decl.ResolvedTypeAlias:
			return nil, p.srcmaps.SyntaxErrors(target, "cannot assign type alias")
		}
	}
	//
	return nil, p.srcmaps.SyntaxErrors(target, "unknown lval")
}

func (p *TypeChecker) typeMemoryLVal(c *decl.ResolvedMemory, e *lval.MemAccess[symbol.Resolved],
	env VariableMap) (Type, []source.SyntaxError) {
	var args, errs = p.typeExpressions(e.Args, env)
	//
	if len(args) != len(c.Address) {
		return nil, p.srcmaps.SyntaxErrors(e,
			fmt.Sprintf("mismatched arguments (expected %d, found %d)", len(c.Address), len(args)))
	} else if len(errs) == 0 {
		// check argument types
		for i := range args {
			ith := c.Address[i].DataType
			errs = append(errs, p.checkEquiTypes(args[i], ith, e.Args[i])...)
		}
	}
	// Done
	return variable.DescriptorsToType(c.Data...), errs
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

func (p *TypeChecker) typeCondition(e expr.ResolvedCondition, env VariableMap) []source.SyntaxError {
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
	if len(lerrs) == 0 && lhs.AsUint(p.env) == nil {
		lerrs = p.srcmaps.SyntaxErrors(e.Left, "expected uint")
	}
	// Check right-hand side
	if len(rerrs) == 0 && rhs.AsUint(p.env) == nil {
		rerrs = p.srcmaps.SyntaxErrors(e.Right, "expected uint")
	}
	// Check matching types
	if len(lerrs)+len(rerrs) == 0 {
		// Equivalence check
		return p.checkEquiTypes(rhs, lhs, e.Right)
	}
	//
	return append(lerrs, rerrs...)
}

func (p *TypeChecker) typeExpression(e expr.Resolved, env VariableMap) (t Type, errs []source.SyntaxError) {
	switch e := e.(type) {
	case *expr.Cast[symbol.Resolved]:
		t, errs = p.typeCastExpression(e, env)
	case *expr.Add[symbol.Resolved]:
		t, errs = p.typeArithmeticExpression(e.Exprs, env)
	case *expr.And[symbol.Resolved]:
		t, errs = p.typeArithmeticExpression(e.Exprs, env)
	case *expr.Const[symbol.Resolved]:
		t, errs = p.typeConst(e, env)
	case *expr.LocalAccess[symbol.Resolved]:
		t, errs = p.typeLocalAccess(e, env)
	case *expr.Mul[symbol.Resolved]:
		t, errs = p.typeArithmeticExpression(e.Exprs, env)
	case *expr.Not[symbol.Resolved]:
		t, errs = p.typeBitwiseNot(e, env)
	case *expr.Or[symbol.Resolved]:
		t, errs = p.typeArithmeticExpression(e.Exprs, env)
	case *expr.ExternAccess[symbol.Resolved]:
		t, errs = p.typeExternAccess(e, env)
	case *expr.Shl[symbol.Resolved]:
		t, errs = p.typeShiftExpression(e.Exprs, env)
	case *expr.Shr[symbol.Resolved]:
		t, errs = p.typeShiftExpression(e.Exprs, env)
	case *expr.Sub[symbol.Resolved]:
		t, errs = p.typeArithmeticExpression(e.Exprs, env)
	case *expr.Xor[symbol.Resolved]:
		t, errs = p.typeArithmeticExpression(e.Exprs, env)
	default:
		return nil, p.srcmaps.SyntaxErrors(e, "unknown expression")
	}
	// Associate type
	e.SetType(t)
	//
	return t, errs
}

func (p *TypeChecker) typeExpressions(exprs []expr.Resolved, env VariableMap) ([]Type, []source.SyntaxError) {
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

func (p *TypeChecker) typeArithmeticExpression(exprs []expr.Resolved, env VariableMap) (Type, []source.SyntaxError) {
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
		if i == 0 && t.AsUint(p.env) == nil {
			return nil, append(errs, *p.srcmaps.SyntaxError(exprs[i], "expected uint"))
		} else if i == 0 {
			res = t
		} else {
			errs = append(errs, p.checkEquiTypes(t, res, exprs[i])...)
		}
	}
	//
	return res, errs
}

func (p *TypeChecker) typeCastExpression(e *expr.Cast[symbol.Resolved], env VariableMap) (Type, []source.SyntaxError) {
	var (
		srcType, errs = p.typeExpression(e.Expr, env)
	)
	//
	if len(errs) == 0 && !data.SubtypeOf(e.CastType, srcType, p.env) && !data.SubtypeOf(srcType, e.CastType, p.env) {
		errs = p.srcmaps.SyntaxErrors(e.Expr, fmt.Sprintf("expected type %s", e.CastType.String(p.env)))
	}
	//
	return e.CastType, errs
}

// typeShiftExpression types a shift expression (Shl or Shr). The result type
// is that of the first (value) operand. All operands must be uint and the
// shift amount must have a compatible type with the value being shifted.
func (p *TypeChecker) typeShiftExpression(exprs []expr.Resolved, env VariableMap) (Type, []source.SyntaxError) {
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
		if t.AsUint(p.env) == nil {
			return nil, p.srcmaps.SyntaxErrors(exprs[i], "expected uint")
		} else if i == 0 {
			res = t
		} else if !data.SubtypeOf(res, t, p.env) && !data.SubtypeOf(t, res, p.env) {
			return nil, p.srcmaps.SyntaxErrors(exprs[i],
				fmt.Sprintf("expected type %s", res.String(p.env)))
		}
	}
	//
	return res, nil
}

func (p *TypeChecker) typeBitwiseNot(e *expr.Not[symbol.Resolved], env VariableMap) (Type, []source.SyntaxError) {
	t, errs := p.typeExpression(e.Expr, env)
	if len(errs) > 0 {
		return nil, errs
	} else if t.AsUint(p.env) == nil {
		return nil, p.srcmaps.SyntaxErrors(e.Expr, "expected uint")
	}

	return t, nil
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
	case *decl.ResolvedConstant:
		return p.typeConstantAccess(t, e, env)
	case *decl.ResolvedMemory:
		return p.typeMemoryAccess(t, e, env)
	case *decl.ResolvedFunction:
		return p.typeFunctionAccess(t, e, env)
	case *decl.ResolvedTypeAlias:
		return p.typeAlias(e)
	default:
		return nil, p.srcmaps.SyntaxErrors(e, "unknown symbol type")
	}
}
func (p *TypeChecker) typeConstantAccess(c *decl.ResolvedConstant, e *expr.ExternAccess[symbol.Resolved],
	env VariableMap) (Type, []source.SyntaxError) {
	return c.DataType, nil
}

func (p *TypeChecker) typeMemoryAccess(c *decl.ResolvedMemory, e *expr.ExternAccess[symbol.Resolved],
	env VariableMap) (Type, []source.SyntaxError) {
	var args, errs = p.typeExpressions(e.Args, env)
	//
	if len(args) != len(c.Address) {
		return nil, p.srcmaps.SyntaxErrors(e,
			fmt.Sprintf("mismatched arguments (expected %d, found %d)", len(c.Address), len(args)))
	} else if len(errs) == 0 {
		// check argument types
		for i := range args {
			ith := c.Address[i].DataType
			// Subtype check
			errs = append(errs, p.checkEquiTypes(args[i], ith, e.Args[i])...)
		}
	}
	// Done
	return variable.DescriptorsToType(c.Data...), errs
}

func (p *TypeChecker) typeAlias(e *expr.ExternAccess[symbol.Resolved]) (Type, []source.SyntaxError) {
	return nil, p.srcmaps.SyntaxErrors(e, "cannot assign type alias")
}

func (p *TypeChecker) typeFunctionAccess(c *decl.ResolvedFunction, e *expr.ExternAccess[symbol.Resolved],
	env VariableMap) (Type, []source.SyntaxError) {
	var (
		args, errs = p.typeExpressions(e.Args, env)
		n          = uint(len(args))
	)

	//
	if n != c.NumInputs {
		return nil, p.srcmaps.SyntaxErrors(e,
			fmt.Sprintf("mismatched arguments (expected %d, found %d)", c.NumInputs, n))
	} else if len(errs) == 0 {
		// check argument types
		for i := range args {
			ith := c.Variables[i].DataType
			// Subtype check
			errs = append(errs, p.checkEquiTypes(args[i], ith, e.Args[i])...)
		}
	}
	// Done
	return variable.DescriptorsToType(c.Outputs()...), errs
}

func (p *TypeChecker) checkEquiTypes(lhs, rhs Type, node any) []source.SyntaxError {
	if !data.EquiTypes(lhs, rhs, p.env) {
		return p.srcmaps.SyntaxErrors(node, fmt.Sprintf("expected type %s", rhs.String(p.env)))
	}
	//
	return nil
}
