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

	"github.com/consensys/go-corset/pkg/util/collection/bit"
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
		effects     bit.Set
		rhs, errors = p.typeExpression(c.DataType, c.ConstExpr, variable.ArrayMap[symbol.Resolved](), effects)
	)
	// Sanity check
	if len(errors) != 0 {
		return errors
	}
	// Subtype check
	return p.checkEquiTypes(rhs, c.DataType, c.ConstExpr)
}

func (p *TypeChecker) typeFunction(fn decl.ResolvedFunction) []source.SyntaxError {
	var (
		errors  []source.SyntaxError
		effects bit.Set
	)
	// initialise effects
	for _, s := range fn.Effects {
		effects.Insert(s.Index)
	}
	// type instructions
	for _, s := range fn.Code {
		switch s := s.(type) {
		case *stmt.Assign[symbol.Resolved]:
			errors = append(errors, p.typeAssignment(s, &fn, effects)...)
		case *stmt.IfGoto[symbol.Resolved]:
			errors = append(errors, p.typeIfGoto(s, &fn, effects)...)
		case *stmt.Printf[symbol.Resolved]:
			errors = append(errors, p.typePrintf(s, &fn, effects)...)
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
		errors       []source.SyntaxError
		dataType     = variable.DescriptorsToType(c.Data...)
		emptyEffects bit.Set
	)
	//
	for _, v := range c.Contents {
		valType, errs := p.typeExpression(dataType, v, variable.ArrayMap[symbol.Resolved](), emptyEffects)
		if len(errs) != 0 {
			errors = append(errors, errs...)
		} else {
			errors = append(errors, p.checkEquiTypes(valType, dataType, v)...)
		}
	}
	//
	return errors
}

func (p *TypeChecker) typeAssignment(s *stmt.Assign[symbol.Resolved], env VariableMap, effects bit.Set,
) []source.SyntaxError {
	var (
		errors  []source.SyntaxError
		sources = []expr.Expr[symbol.Resolved]{s.Source}
	)
	// Sanity check assignment arity
	if len(s.Targets) != 0 && len(s.Targets) < len(sources) {
		return p.srcmaps.SyntaxErrors(s, fmt.Sprintf("insufficient target variables (expected %d)", len(sources)))
	} else if len(s.Targets) > len(sources) {
		return p.srcmaps.SyntaxErrors(s, fmt.Sprintf("too many target variables (expected %d)", len(sources)))
	} else if len(s.Targets) == 0 {
		// Special case for empty targets.  This can only arise for a function
		// call which does not assign any return values.  This ensures that, in
		// such case, the source expression is typed.
		_, errors = p.typeExpression(nil, s.Source, env, effects)
		//
		return errors
	}
	// Check each in turn
	for i, lval := range s.Targets {
		var (
			rhs             = sources[i]
			lval_t, lhsErrs = p.typeLval(lval, env, effects)
			rhs_t, rhsErrs  = p.typeExpression(lval_t, rhs, env, effects)
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

func (p *TypeChecker) typeLval(target LVal, env VariableMap, effects bit.Set) (Type, []source.SyntaxError) {
	// determine lhs width
	switch t := target.(type) {
	case *lval.Variable[symbol.Resolved]:
		var bitwidth uint
		// Special case single variables
		if len(t.Ids) == 1 {
			return env.Variable(t.Ids[0]).DataType, nil
		}
		// Consider destructurings
		for _, id := range t.Ids {
			id_t := env.Variable(id).DataType.AsUint(p.env)
			// Check whether have integer type or not
			if id_t == nil {
				return nil, p.srcmaps.SyntaxErrors(target, "expected integer type")
			}
			//
			bitwidth += id_t.BitWidth()
		}
		//
		return data.NewUnsignedInt[symbol.Resolved](bitwidth, false), nil
	case *lval.MemAccess[symbol.Resolved]:
		// Lookup the symbol
		var extern = p.lookup(t.Name)
		//
		switch e := extern.(type) {
		case *decl.ResolvedConstant:
			return nil, p.srcmaps.SyntaxErrors(target, "cannot assign constant")
		case *decl.ResolvedMemory:
			return p.typeMemoryLVal(e, t, env, effects)
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
	env VariableMap, effects bit.Set) (Type, []source.SyntaxError) {
	var errors []source.SyntaxError
	// Initial sanity checks
	if len(e.Args) != len(c.Address) {
		return nil, p.srcmaps.SyntaxErrors(e,
			fmt.Sprintf("mismatched arguments (expected %d, found %d)", len(c.Address), len(e.Args)))
	} else if !effects.Contains(e.Name.Index) && c.IsReadable() && c.IsWriteable() {
		return nil, p.srcmaps.SyntaxErrors(e, "read/write memory not visible here")
	}
	// check argument types
	for i, e := range e.Args {
		var (
			ith         = c.Address[i].DataType
			arg_t, errs = p.typeExpression(c.Address[i].DataType, e, env, effects)
		)
		//
		errors = append(errors, errs...)
		// Subtype check (if no other errors)
		if len(errs) == 0 {
			errors = append(errors, p.checkEquiTypes(arg_t, ith, e)...)
		}
	}
	// Done
	return variable.DescriptorsToType(c.Data...), errors
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

func (p *TypeChecker) typeIfGoto(s *stmt.IfGoto[symbol.Resolved], env VariableMap, effects bit.Set,
) []source.SyntaxError {
	return p.typeCondition(s.Cond, env, effects)
}

func (p *TypeChecker) typePrintf(s *stmt.Printf[symbol.Resolved], env VariableMap, effects bit.Set,
) []source.SyntaxError {
	var errs []source.SyntaxError
	//
	for _, e := range s.Arguments {
		ith, ierrs := p.typeExpression(nil, e, env, effects)
		//
		if len(ierrs) == 0 && ith.AsUint(p.env) == nil {
			errs = append(errs, *p.srcmaps.SyntaxError(e, "expected uint"))
		} else if len(ierrs) == 0 && ith.AsUint(p.env).IsOpen() {
			errs = append(errs, *p.srcmaps.SyntaxError(e, "concrete type required"))
		} else {
			errs = append(errs, ierrs...)
		}
	}
	//
	return errs
}

func (p *TypeChecker) typeCondition(e expr.ResolvedCondition, env VariableMap, effects bit.Set) []source.SyntaxError {
	switch e := e.(type) {
	case *expr.Cmp[symbol.Resolved]:
		return p.typeCmp(e, env, effects)
	default:
		return p.srcmaps.SyntaxErrors(e, "unknown condition")
	}
}

// typeTernaryCondition type-checks the condition of a ternary expression.
// After lowering, the condition is always a single Cmp node.
func (p *TypeChecker) typeTernaryCondition(e expr.Resolved, env VariableMap, effects bit.Set) []source.SyntaxError {
	if cmp, ok := e.(*expr.Cmp[symbol.Resolved]); ok {
		return p.typeCmp(cmp, env, effects)
	}

	return p.srcmaps.SyntaxErrors(e, "invalid ternary condition")
}

func (p *TypeChecker) typeCmp(e *expr.Cmp[symbol.Resolved], env VariableMap, effects bit.Set) []source.SyntaxError {
	var (
		lhs, lerrs = p.typeExpression(nil, e.Left, env, effects)
		rhs, rerrs = p.typeExpression(lhs, e.Right, env, effects)
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

func (p *TypeChecker) typeExpression(expected Type, e expr.Resolved, env VariableMap, effects bit.Set,
) (actual Type, errs []source.SyntaxError) {
	switch e := e.(type) {
	case *expr.Cast[symbol.Resolved]:
		actual, errs = p.typeCastExpression(expected, e, env, effects)
	case *expr.Concat[symbol.Resolved]:
		actual, errs = p.typeConcatExpression(expected, e, env, effects)
	case *expr.Add[symbol.Resolved]:
		actual, errs = p.typeArithmeticExpression(expected, e.Exprs, env, effects)
	case *expr.BitwiseAnd[symbol.Resolved]:
		actual, errs = p.typeArithmeticExpression(expected, e.Exprs, env, effects)
	case *expr.Const[symbol.Resolved]:
		actual, errs = p.typeConst(expected, e, env)
	case *expr.LocalAccess[symbol.Resolved]:
		actual, errs = p.typeLocalAccess(e, env)
	case *expr.Mul[symbol.Resolved]:
		actual, errs = p.typeArithmeticExpression(expected, e.Exprs, env, effects)
	case *expr.BitwiseNot[symbol.Resolved]:
		actual, errs = p.typeBitwiseNot(expected, e, env, effects)
	case *expr.BitwiseOr[symbol.Resolved]:
		actual, errs = p.typeArithmeticExpression(expected, e.Exprs, env, effects)
	case *expr.ExternAccess[symbol.Resolved]:
		actual, errs = p.typeExternAccess(e, env, effects)
	case *expr.Shl[symbol.Resolved]:
		actual, errs = p.typeShiftExpression(expected, e.Exprs, env, effects)
	case *expr.Shr[symbol.Resolved]:
		actual, errs = p.typeShiftExpression(expected, e.Exprs, env, effects)
	case *expr.Div[symbol.Resolved]:
		actual, errs = p.typeArithmeticExpression(expected, e.Exprs, env, effects)
	case *expr.Rem[symbol.Resolved]:
		actual, errs = p.typeArithmeticExpression(expected, e.Exprs, env, effects)
	case *expr.Sub[symbol.Resolved]:
		actual, errs = p.typeArithmeticExpression(expected, e.Exprs, env, effects)
	case *expr.Xor[symbol.Resolved]:
		actual, errs = p.typeArithmeticExpression(expected, e.Exprs, env, effects)
	case *expr.Ternary[symbol.Resolved]:
		errs = append(errs, p.typeTernaryCondition(e.Cond, env, effects)...)
		tt, terrs := p.typeExpression(expected, e.IfTrue, env, effects)
		ft, ferrs := p.typeExpression(expected, e.IfFalse, env, effects)

		errs = append(append(errs, terrs...), ferrs...)
		if len(errs) == 0 {
			errs = p.checkEquiTypes(ft, tt, e.IfFalse)
			actual = tt
		}

	default:
		return nil, p.srcmaps.SyntaxErrors(e, "invalid expression")
	}
	// Associate type
	e.SetType(actual)
	//
	return actual, errs
}

func (p *TypeChecker) typeUintExpressions(t Type, exprs []expr.Resolved, env VariableMap, effects bit.Set,
) (Type, []source.SyntaxError) {
	var (
		errors []source.SyntaxError
		res    *data.UnsignedInt[symbol.Resolved]
	)
	//
	for i, e := range exprs {
		ith_t, errs := p.typeExpression(t, e, env, effects)
		//
		if len(errs) > 0 {
			errors = append(errors, errs...)
		} else if i == 0 && ith_t.AsUint(p.env) == nil {
			return nil, append(errors, *p.srcmaps.SyntaxError(exprs[i], "expected uint"))
		} else if i == 0 {
			res = ith_t.AsUint(p.env)
		} else if len(errors) > 0 {
			// skip type checking
		} else if errs := p.checkEquiTypes(ith_t, res, exprs[i]); len(errs) > 0 {
			errors = append(errors, errs...)
		} else {
			res = res.Join(exprs[i].Type().AsUint(p.env))
		}
	}
	//
	return res, errors
}

func (p *TypeChecker) typeArithmeticExpression(t Type, exprs []expr.Resolved, env VariableMap, effects bit.Set,
) (Type, []source.SyntaxError) {
	var res, errors = p.typeUintExpressions(t, exprs, env, effects)
	//
	return res, errors
}

func (p *TypeChecker) typeConcatExpression(t Type, e *expr.Concat[symbol.Resolved], env VariableMap, effects bit.Set,
) (Type, []source.SyntaxError) {
	var (
		errors   []source.SyntaxError
		bitwidth uint
	)
	//
	for _, e := range e.Exprs {
		ith_t, errs := p.typeExpression(t, e, env, effects)
		//
		if len(errs) > 0 {
			errors = append(errors, errs...)
		} else if ith_t := ith_t.AsUint(p.env); ith_t == nil {
			return nil, append(errors, *p.srcmaps.SyntaxError(e, "expected uint"))
		} else if ith_t := ith_t.AsUint(p.env); ith_t.IsOpen() {
			return nil, append(errors, *p.srcmaps.SyntaxError(e, "expected fixed-width uint"))
		} else {
			bitwidth += ith_t.BitWidth()
		}
	}
	//
	return data.NewUnsignedInt[symbol.Resolved](bitwidth, false), nil
}

func (p *TypeChecker) typeCastExpression(t Type, e *expr.Cast[symbol.Resolved], env VariableMap, effects bit.Set,
) (Type, []source.SyntaxError) {
	var (
		srcType, errors = p.typeExpression(nil, e.Expr, env, effects)
	)
	//
	if len(errors) == 0 && !data.SubtypeOf(e.CastType, srcType, p.env) && !data.SubtypeOf(srcType, e.CastType, p.env) {
		errors = p.srcmaps.SyntaxErrors(e.Expr, fmt.Sprintf("expected type %s", e.CastType.String(p.env)))
	}
	//
	return e.CastType, errors
}

// typeShiftExpression types a shift expression (Shl or Shr). The result type
// is that of the first (value) operand. All operands must be uint and the
// shift amount must have a compatible type with the value being shifted.
func (p *TypeChecker) typeShiftExpression(t Type, exprs []expr.Resolved, env VariableMap, effects bit.Set,
) (Type, []source.SyntaxError) {
	var (
		arg, errors = p.typeExpression(t, exprs[0], env, effects)
		_, errs2    = p.typeUintExpressions(nil, exprs[1:], env, effects)
		res         *data.UnsignedInt[symbol.Resolved]
	)
	// Sanity check argument
	if len(errors) > 0 {
		// don't type check as other problems
	} else if res = arg.AsUint(p.env); res == nil {
		return nil, append(errors, *p.srcmaps.SyntaxError(exprs[0], "expected uint"))
	}
	//
	return res, append(errors, errs2...)
}

func (p *TypeChecker) typeBitwiseNot(t Type, e *expr.BitwiseNot[symbol.Resolved], env VariableMap, effects bit.Set,
) (Type, []source.SyntaxError) {
	t, errs := p.typeExpression(t, e.Expr, env, effects)
	//
	if len(errs) > 0 {
		return nil, errs
	} else if t.AsUint(p.env) == nil {
		return nil, p.srcmaps.SyntaxErrors(e.Expr, "expected uint")
	}

	return t, nil
}

func (p *TypeChecker) typeConst(t Type, e *expr.Const[symbol.Resolved], env VariableMap) (Type, []source.SyntaxError) {
	var (
		bitwidth = uint(e.Constant.BitLen())
		actual   = data.NewUnsignedInt[symbol.Resolved](bitwidth, true)
	)
	//
	if t == nil {
		return actual, nil
	}
	//
	return t, p.checkEquiTypes(actual, t, e)
}

func (p *TypeChecker) typeLocalAccess(e *expr.LocalAccess[symbol.Resolved], env VariableMap,
) (Type, []source.SyntaxError) {
	//
	return env.Variable(e.Variable).DataType, nil
}

func (p *TypeChecker) typeExternAccess(e *expr.ExternAccess[symbol.Resolved], env VariableMap, effects bit.Set,
) (Type, []source.SyntaxError) {
	// Lookup the symbol
	var extern = p.lookup(e.Name)
	// Decide what kind of symbol it is
	switch t := extern.(type) {
	case *decl.ResolvedConstant:
		return p.typeConstantAccess(t)
	case *decl.ResolvedMemory:
		return p.typeMemoryAccess(t, e, env, effects)
	case *decl.ResolvedFunction:
		return p.typeFunctionCall(t, e, env, effects)
	case *decl.ResolvedTypeAlias:
		return p.typeAlias(e)
	default:
		return nil, p.srcmaps.SyntaxErrors(e, "unknown symbol type")
	}
}
func (p *TypeChecker) typeConstantAccess(c *decl.ResolvedConstant) (Type, []source.SyntaxError) {
	return c.DataType, nil
}

func (p *TypeChecker) typeMemoryAccess(c *decl.ResolvedMemory, e *expr.ExternAccess[symbol.Resolved],
	env VariableMap, effects bit.Set) (Type, []source.SyntaxError) {
	var errors []source.SyntaxError
	//
	if len(e.Args) != len(c.Address) {
		return nil, p.srcmaps.SyntaxErrors(e,
			fmt.Sprintf("mismatched arguments (expected %d, found %d)", len(c.Address), len(e.Args)))
	} else if c.IsReadable() && c.IsWriteable() && !effects.Contains(e.Name.Index) {
		return nil, p.srcmaps.SyntaxErrors(e, "read/write memory not visible here")
	}
	// check argument types
	for i, arg := range e.Args {
		var (
			ith         = c.Address[i].DataType
			ith_t, errs = p.typeExpression(ith, arg, env, effects)
		)
		//
		errors = append(errors, errs...)
		// Perofmr subtype check (if no other errors)
		if len(errs) == 0 {
			// Subtype check
			errors = append(errors, p.checkEquiTypes(ith_t, ith, e.Args[i])...)
		}
	}
	// Done
	return variable.DescriptorsToType(c.Data...), errors
}

func (p *TypeChecker) typeAlias(e *expr.ExternAccess[symbol.Resolved]) (Type, []source.SyntaxError) {
	return nil, p.srcmaps.SyntaxErrors(e, "not an expression")
}

func (p *TypeChecker) typeFunctionCall(c *decl.ResolvedFunction, e *expr.ExternAccess[symbol.Resolved],
	env VariableMap, effects bit.Set) (Type, []source.SyntaxError) {
	var (
		errors []source.SyntaxError
		n      = uint(len(e.Args))
	)
	//
	if n != c.NumInputs {
		return nil, p.srcmaps.SyntaxErrors(e,
			fmt.Sprintf("mismatched arguments (expected %d, found %d)", c.NumInputs, n))
	}
	// check argument types
	for i, arg := range e.Args {
		var (
			ith         = c.Variables[i].DataType
			ith_t, errs = p.typeExpression(ith, arg, env, effects)
		)
		//
		errors = append(errors, errs...)
		// Subtype check (if no other errors)
		if len(errs) == 0 {
			errors = append(errors, p.checkEquiTypes(ith_t, ith, e.Args[i])...)
		}
	}
	// Sanity check callee effects visible at call site
	for _, effect := range c.Effects {
		if !effects.Contains(effect.Index) {
			errors = append(errors,
				*p.srcmaps.SyntaxError(e, fmt.Sprintf("read/write memory \"%s\" not visible here", effect.Name)))
		}
	}
	// Done
	return variable.DescriptorsToType(c.Outputs()...), errors
}

func (p *TypeChecker) checkEquiTypes(lhs, rhs Type, node any) []source.SyntaxError {
	if !data.EquiTypes(lhs, rhs, p.env) {
		return p.srcmaps.SyntaxErrors(node, fmt.Sprintf("expected type %s", rhs.String(p.env)))
	}
	//
	return nil
}
