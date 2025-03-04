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

	"github.com/consensys/go-corset/pkg/mir"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/util"
)

// LowerToMir lowers (or refines) an HIR table into an MIR schema.  That means
// lowering all the columns and constraints, whilst adding additional columns /
// constraints as necessary to preserve the original semantics.
func (p *Schema) LowerToMir() *mir.Schema {
	mirSchema := mir.EmptySchema()
	// Copy modules
	for _, mod := range p.modules {
		mirSchema.AddModule(mod.Name)
	}
	// Lower columns
	for _, input := range p.inputs {
		col := input.(DataColumn)
		mirSchema.AddDataColumn(col.Context(), col.Name(), col.Type())
	}
	// Lower assignments (nothing to do here)
	for _, a := range p.assignments {
		mirSchema.AddAssignment(a)
	}
	// Lower constraints
	for _, c := range p.constraints {
		lowerConstraintToMir(c, mirSchema)
	}
	// Copy property assertions.  Observe, these do not require lowering
	// because they are already MIR-level expressions.
	for _, c := range p.assertions {
		properties := lowerToConstraints(c.Property.Expr, mirSchema)
		for _, p := range properties {
			mirSchema.AddPropertyAssertion(c.Handle, c.Context, p)
		}
	}
	//
	return mirSchema
}

func lowerConstraintToMir(c sc.Constraint, schema *mir.Schema) {
	// Check what kind of constraint we have
	if v, ok := c.(LookupConstraint); ok {
		lowerLookupConstraint(v, schema)
	} else if v, ok := c.(VanishingConstraint); ok {
		mir_constraints := lowerToConstraints(v.Constraint.Expr, schema)
		// Add individual constraints arising
		for i, mir_consrtaint := range mir_constraints {
			schema.AddVanishingConstraint(v.Handle, uint(i), v.Context, v.Domain, mir_consrtaint)
		}
	} else if v, ok := c.(RangeConstraint); ok {
		mir_exprs := v.Expr.LowerTo(schema)
		// Add individual constraints arising
		for i, mir_expr := range mir_exprs {
			schema.AddRangeConstraint(v.Handle, uint(i), v.Context, mir_expr, v.Bound)
		}
	} else if v, ok := c.(SortedConstraint); ok {
		lowerSortedConstraint(v, schema)
	} else {
		// Should be unreachable as no other constraint types can be added to a
		// schema.
		panic("unreachable")
	}
}

func lowerLookupConstraint(c LookupConstraint, schema *mir.Schema) {
	from := make([]mir.Expr, len(c.Sources))
	into := make([]mir.Expr, len(c.Targets))
	// Convert general expressions into unit expressions.
	for i := 0; i < len(from); i++ {
		from[i] = lowerUnitTo(c.Sources[i], schema)
		into[i] = lowerUnitTo(c.Targets[i], schema)
	}
	//
	schema.AddLookupConstraint(c.Handle, c.SourceContext, c.TargetContext, from, into)
}

func lowerSortedConstraint(c SortedConstraint, schema *mir.Schema) {
	var (
		selector util.Option[mir.Expr] = util.None[mir.Expr]()
		sources                        = make([]mir.Expr, len(c.Sources))
	)
	// Convert (optional) selector expression
	if c.Selector.HasValue() {
		selector = util.Some(lowerUnitTo(c.Selector.Unwrap(), schema))
	}
	// Convert general expressions into unit expressions.
	for i := 0; i < len(sources); i++ {
		sources[i] = lowerUnitTo(c.Sources[i], schema)
	}
	//
	schema.AddSortedConstraint(c.Handle, c.Context, c.BitWidth, selector, sources, c.Signs, c.Strict)
}

// Lower an expression which is expected to lower into a single expression.
// This will panic if the unit expression is malformed (i.e. does not lower
// into a single expression).
func lowerUnitTo(e UnitExpr, schema *mir.Schema) mir.Expr {
	exprs := lowerTo(e.Expr, schema)

	if len(exprs) != 1 {
		panic("invalid unitary expression")
	}

	return exprs[0]
}

// ============================================================================
// lowerTo
// ============================================================================

// Lowers a given expression to the MIR level.  The expression is first expanded
// into one or more target expressions. Furthermore, conditions must be "lifted"
// to the root.
func lowerToConstraints(e Expr, schema *mir.Schema) []mir.Constraint {
	// First expand expression
	es := expand(e.Term, schema)
	// Now lower each one (carefully)
	mes := make([]mir.Constraint, len(es))
	//
	for i, e := range es {
		c := extractCondition(e, schema)
		b := extractBody(e, schema)
		mes[i] = mir.Disjunct(c, b.EqualsZero())
	}
	// Done
	return mes
}

// Lowers a given expression to the MIR level.  The expression is first expanded
// into one or more target expressions. Furthermore, conditions must be "lifted"
// to the root.
func lowerTo(e Expr, schema *mir.Schema) []mir.Expr {
	// First expand expression
	es := expand(e.Term, schema)
	// Now lower each one (carefully)
	mes := make([]mir.Expr, len(es))
	//
	for i, e := range es {
		c := extractCondition(e, schema).AsExpr()
		b := extractBody(e, schema)
		mes[i] = mir.Product(c, b).Simplify()
	}
	// Done
	return mes
}

// Extract the "condition" of an expression.  Every expression can be view as a
// conditional constraint of the form "if c then e", where "c" is the condition.
// This is allowed to return nil if the body is unconditional.
func extractCondition(e Term, schema *mir.Schema) mir.Constraint {
	switch e := e.(type) {
	case *Add:
		return extractConditions(e.Args, schema)
	case *Cast:
		return extractCondition(e.Arg, schema)
	case *Constant:
		return mir.TRUE
	case *ColumnAccess:
		return mir.TRUE
	case *Exp:
		return extractCondition(e.Arg, schema)
	case *IfZero:
		return extractIfZeroCondition(e, schema)
	case *Mul:
		return extractConditions(e.Args, schema)
	case *Norm:
		return extractCondition(e.Arg, schema)
	case *Sub:
		return extractConditions(e.Args, schema)
	default:
		name := reflect.TypeOf(e).Name()
		panic(fmt.Sprintf("unknown HIR expression \"%s\"", name))
	}
}

func extractConditions(es []Term, schema *mir.Schema) mir.Constraint {
	var r mir.Constraint = mir.TRUE
	//
	for _, e := range es {
		r = mir.Disjunct(r, extractCondition(e, schema))
	}
	//
	return r
}

// Extracting from conditional expressions is slightly more complex than others,
// so it gets a case of its own.
func extractIfZeroCondition(e *IfZero, schema *mir.Schema) mir.Constraint {
	var (
		bc mir.Constraint
		cb mir.Constraint
	)
	// Lower condition
	cc := extractCondition(e.Condition, schema)
	body := extractBody(e.Condition, schema)
	// Add conditions arising
	if e.TrueBranch != nil && e.FalseBranch != nil {
		// Expansion should ensure this case does not exist.  This is necessary
		// to ensure exactly one expression is generated from this expression.
		panic(fmt.Sprintf("unexpanded expression (%s)", lispOfTerm(e, schema)))
	} else if e.TrueBranch != nil {
		// Lower conditional's arising from body
		cb = body.NotEqualsZero()
		bc = extractCondition(e.TrueBranch, schema)
	} else {
		// Lower conditional's arising from body
		cb = body.EqualsZero()
		bc = extractCondition(e.FalseBranch, schema)
	}
	//
	return mir.Disjunct(cc, cb, bc)
}

// Translate the "body" of an expression.  Every expression can be view as a
// conditional constraint of the form "if c then e", where "e" is the
// constraint.
func extractBody(e Term, schema *mir.Schema) mir.Expr {
	switch e := e.(type) {
	case *Add:
		return mir.Sum(extractBodies(e.Args, schema)...)
	case *Cast:
		return mir.CastOf(extractBody(e.Arg, schema), e.BitWidth)
	case *Constant:
		return mir.NewConst(e.Value)
	case *ColumnAccess:
		return mir.NewColumnAccess(e.Column, e.Shift)
	case *Exp:
		return mir.Exponent(extractBody(e.Arg, schema), e.Pow)
	case *IfZero:
		if e.TrueBranch != nil && e.FalseBranch != nil {
			// Expansion should ensure this case does not exist.  This is necessary
			// to ensure exactly one expression is generated from this expression.
			panic(fmt.Sprintf("unexpanded expression (%s)", lispOfTerm(e, schema)))
		} else if e.TrueBranch != nil {
			return extractBody(e.TrueBranch, schema)
		}
		// Done
		return extractBody(e.FalseBranch, schema)
	case *Mul:
		return mir.Product(extractBodies(e.Args, schema)...)
	case *Norm:
		return mir.Normalise(extractBody(e.Arg, schema))
	case *Sub:
		return mir.Subtract(extractBodies(e.Args, schema)...)
	default:
		name := reflect.TypeOf(e).Name()
		panic(fmt.Sprintf("unknown HIR expression \"%s\"", name))
	}
}

// Extract a vector of expanded expressions to the MIR level.
func extractBodies(es []Term, schema *mir.Schema) []mir.Expr {
	rs := make([]mir.Expr, len(es))
	for i, e := range es {
		rs[i] = extractBody(e, schema)
	}

	return rs
}

// ============================================================================
// expand
// ============================================================================

// Expand a term into one or more terms by eliminating lists and
// breaking down conditions.  For example, a list such as say "(begin (- X Y) (-
// Y Z))" is broken down into two distinct expressions "(- X Y)" and "(- Y Z)".
// Likewise, a condition such as "(if X Y Z)" is broken down into two
// expressions "(if X Y)" and "(ifnot X Z)".  These are necessary steps for the
// conversion into a lower-level form.
func expand(e Term, schema sc.Schema) []Term {
	switch e := e.(type) {
	case *Add:
		return expandAdd(e, schema)
	case *Cast:
		return expandCast(e, schema)
	case *Constant:
		return []Term{e}
	case *ColumnAccess:
		return []Term{e}
	case *Mul:
		return expandMul(e, schema)
	case *List:
		return expandList(e, schema)
	case *Exp:
		return expandExp(e, schema)
	case *Norm:
		return expandNorm(e, schema)
	case *IfZero:
		return expandIfZero(e, schema)
	case *Sub:
		return expandSub(e, schema)
	}
	// Should be unreachable
	panic(fmt.Sprintf("unknown expression: %s", lispOfTerm(e, schema)))
}

func expandSub(e *Sub, schema sc.Schema) []Term {
	return expandWithNaryConstructor(e.Args, func(nargs []Term) Term {
		return &Sub{Args: nargs}
	}, schema)
}

func expandIfZero(e *IfZero, schema sc.Schema) []Term {
	ees := make([]Term, 0)
	// Expand true branch with condition
	if e.TrueBranch != nil {
		ees = expandWithBinaryConstructor(e.Condition, e.TrueBranch, func(c Term, tb Term) Term {
			return &IfZero{c, tb, nil}
		}, schema)
	}
	// Expand false branch with condition
	if e.FalseBranch != nil {
		fes := expandWithBinaryConstructor(e.Condition, e.FalseBranch, func(c Term, fb Term) Term {
			return &IfZero{c, nil, fb}
		}, schema)
		ees = append(ees, fes...)
	}
	// Dne
	return ees
}

func expandNorm(e *Norm, schema sc.Schema) []Term {
	ees := expand(e.Arg, schema)
	for i, ee := range ees {
		ees[i] = &Norm{ee}
	}

	return ees
}

func expandExp(e *Exp, schema sc.Schema) []Term {
	ees := expand(e.Arg, schema)
	for i, ee := range ees {
		ees[i] = &Exp{ee, e.Pow}
	}

	return ees
}

func expandList(e *List, schema sc.Schema) []Term {
	ees := make([]Term, 0)
	for _, arg := range e.Args {
		ees = append(ees, expand(arg, schema)...)
	}

	return ees
}

func expandAdd(e *Add, schema sc.Schema) []Term {
	return expandWithNaryConstructor(e.Args, func(nargs []Term) Term {
		var args []Term
		// Flatten nested sums
		for _, e := range nargs {
			if a, ok := e.(*Add); ok {
				args = append(args, a.Args...)
			} else {
				args = append(args, e)
			}
		}
		// Done
		return &Add{Args: args}
	}, schema)
}

func expandCast(e *Cast, schema sc.Schema) []Term {
	ees := expand(e.Arg, schema)
	for i, ee := range ees {
		ees[i] = &Cast{ee, e.BitWidth}
	}

	return ees
}

func expandMul(e *Mul, schema sc.Schema) []Term {
	return expandWithNaryConstructor(e.Args, func(nargs []Term) Term {
		var args []Term
		// Flatten nested products
		for _, e := range nargs {
			if a, ok := e.(*Mul); ok {
				args = append(args, a.Args...)
			} else {
				args = append(args, e)
			}
		}
		// Done
		return &Mul{Args: args}
	}, schema)
}

type binaryConstructor func(Term, Term) Term
type naryConstructor func([]Term) Term

// LowerWithBinaryConstructor is a generic mechanism for lowering down to a binary expression.
func expandWithBinaryConstructor(lhs Term, rhs Term, create binaryConstructor, schema sc.Schema) []Term {
	var res []Term
	// Lower all three expressions
	is := expand(lhs, schema)
	js := expand(rhs, schema)

	// Now construct
	for i := 0; i < len(is); i++ {
		for j := 0; j < len(js); j++ {
			// Construct binary expression
			expr := create(is[i], js[j])
			// Append to the end
			res = append(res, expr)
		}
	}

	return res
}

// LowerWithNaryConstructor performs the cross-product expansion of an nary HIR
// expression. This is necessary because each argument of that expression will
// itself turn into one or more MIR expressions.  For example, consider lowering
// the following HIR expression:
//
// > (if X Y Z) + 10
//
// Here, (if X Y Z) will lower into two MIR expressions: (1-NORM(X))*Y and X*Z.
// Thus, we need to generate two MIR expressions for our example:
//
// > ((1 - NORM(X)) * Y) + 10 > (X * Y) + 10
//
// Finally, consider an expression such as the following:
//
// > (if X Y Z) + (if A B C)
//
// This will expand into *four* MIR expressions (i.e. the cross product of the
// left and right ifs).
func expandWithNaryConstructor(args []Term, constructor naryConstructor, schema sc.Schema) []Term {
	// Accumulator is initially empty
	acc := make([]Term, len(args))
	// Start from the first argument
	return expandWithNaryConstructorHelper(0, acc, args, constructor, schema)
}

// LowerWithNaryConstructorHelper manages progress through the cross-product expansion.
// Specifically, "i" determines how much of args has been lowered thus
// far, whilst "acc" represents the current array being generated.
func expandWithNaryConstructorHelper(i int, acc []Term, args []Term,
	constructor naryConstructor, schema sc.Schema) []Term {
	if i == len(acc) {
		// Base Case
		nacc := make([]Term, len(acc))
		// Clone the slice because it is used as a temporary
		// working storage during the expansion.
		copy(nacc, acc)
		// Apply the constructor to produce the appropriate
		// mir.Expr.
		return []Term{constructor(nacc)}
	}

	// Recursive Case
	var nargs []Term

	for _, ith := range expand(args[i], schema) {
		acc[i] = ith
		iths := expandWithNaryConstructorHelper(i+1, acc, args, constructor, schema)
		nargs = append(nargs, iths...)
	}

	return nargs
}
