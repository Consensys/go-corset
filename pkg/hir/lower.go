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
		lowerConstraintToMir(c, mirSchema, p)
	}
	// Copy property assertions.  Observe, these do not require lowering
	// because they are already MIR-level expressions.
	for _, c := range p.assertions {
		properties := lowerToConstraints(c.Property.Expr, mirSchema, p)
		for _, p := range properties {
			mirSchema.AddPropertyAssertion(c.Handle, c.Context, p)
		}
	}
	//
	return mirSchema
}

func lowerConstraintToMir(c sc.Constraint, mirSchema *mir.Schema, hirSchema *Schema) {
	// Check what kind of constraint we have
	if v, ok := c.(LookupConstraint); ok {
		lowerLookupConstraint(v, mirSchema, hirSchema)
	} else if v, ok := c.(VanishingConstraint); ok {
		mir_constraints := lowerToConstraints(v.Constraint.Expr, mirSchema, hirSchema)
		// Add individual constraints arising
		for i, mir_constraint := range mir_constraints {
			mirSchema.AddVanishingConstraint(v.Handle, uint(i), v.Context, v.Domain, mir_constraint)
		}
	} else if v, ok := c.(RangeConstraint); ok {
		mir_exprs := lowerTo(v.Expr.Expr, mirSchema, hirSchema)
		// Add individual constraints arising
		for i, mir_expr := range mir_exprs {
			mirSchema.AddRangeConstraint(v.Handle, uint(i), v.Context, mir_expr, v.Bound)
		}
	} else if v, ok := c.(SortedConstraint); ok {
		lowerSortedConstraint(v, mirSchema, hirSchema)
	} else {
		// Should be unreachable as no other constraint types can be added to a
		// schema.
		panic("unreachable")
	}
}

func lowerLookupConstraint(c LookupConstraint, mirSchema *mir.Schema, hirSchema *Schema) {
	from := make([]mir.Expr, len(c.Sources))
	into := make([]mir.Expr, len(c.Targets))
	// Convert general expressions into unit expressions.
	for i := 0; i < len(from); i++ {
		from[i] = lowerUnitTo(c.Sources[i], mirSchema, hirSchema)
		into[i] = lowerUnitTo(c.Targets[i], mirSchema, hirSchema)
	}
	//
	mirSchema.AddLookupConstraint(c.Handle, c.SourceContext, c.TargetContext, from, into)
}

func lowerSortedConstraint(c SortedConstraint, mirSchema *mir.Schema, hirSchema *Schema) {
	var (
		selector util.Option[mir.Expr] = util.None[mir.Expr]()
		sources                        = make([]mir.Expr, len(c.Sources))
	)
	// Convert (optional) selector expression
	if c.Selector.HasValue() {
		selector = util.Some(lowerUnitTo(c.Selector.Unwrap(), mirSchema, hirSchema))
	}
	// Convert general expressions into unit expressions.
	for i := 0; i < len(sources); i++ {
		sources[i] = lowerUnitTo(c.Sources[i], mirSchema, hirSchema)
	}
	//
	mirSchema.AddSortedConstraint(c.Handle, c.Context, c.BitWidth, selector, sources, c.Signs, c.Strict)
}

// Lower an expression which is expected to lower into a single expression.
// This will panic if the unit expression is malformed (i.e. does not lower
// into a single expression).
func lowerUnitTo(e UnitExpr, mirSchema *mir.Schema, hirSchema *Schema) mir.Expr {
	exprs := lowerTo(e.Expr, mirSchema, hirSchema)

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
func lowerToConstraints(e Expr, mirSchema *mir.Schema, hirSchema *Schema) []mir.Constraint {
	// First expand expression
	es := expand(e.Term, hirSchema)
	// Now lower each one (carefully)
	mes := make([]mir.Constraint, len(es))
	//
	for i, e := range es {
		c := extractCondition(e, mirSchema, hirSchema)
		b := extractBody(e, mirSchema, hirSchema)
		mes[i] = mir.Disjunct(c, b.EqualsZero())
	}
	// Done
	return mes
}

// Lowers a given expression to the MIR level.  The expression is first expanded
// into one or more target expressions. Furthermore, conditions must be "lifted"
// to the root.
func lowerTo(e Expr, mirSchema *mir.Schema, hirSchema *Schema) []mir.Expr {
	// First expand expression
	es := expand(e.Term, hirSchema)
	// Now lower each one (carefully)
	mes := make([]mir.Expr, len(es))
	//
	for i, e := range es {
		c := extractCondition(e, mirSchema, hirSchema).AsExpr()
		b := extractBody(e, mirSchema, hirSchema)
		mes[i] = mir.Product(c, b).Simplify()
	}
	// Done
	return mes
}

// Extract the "condition" of an expression.  Every expression can be view as a
// conditional constraint of the form "if c then e", where "c" is the condition.
// This is allowed to return nil if the body is unconditional.
func extractCondition(e Term, mirSchema *mir.Schema, hirSchema *Schema) mir.Constraint {
	switch e := e.(type) {
	case *Add:
		return extractConditions(e.Args, mirSchema, hirSchema)
	case *Cast:
		return extractCondition(e.Arg, mirSchema, hirSchema)
	case *Constant:
		return mir.TRUE
	case *ColumnAccess:
		return mir.TRUE
	case *Exp:
		return extractCondition(e.Arg, mirSchema, hirSchema)
	case *IfZero:
		return extractIfZeroCondition(e, mirSchema, hirSchema)
	case *LabelledConstant:
		return mir.TRUE
	case *Mul:
		return extractConditions(e.Args, mirSchema, hirSchema)
	case *Norm:
		return extractCondition(e.Arg, mirSchema, hirSchema)
	case *Sub:
		return extractConditions(e.Args, mirSchema, hirSchema)
	default:
		name := reflect.TypeOf(e).Name()
		panic(fmt.Sprintf("unknown HIR expression \"%s\"", name))
	}
}

func extractConditions(es []Term, mirSchema *mir.Schema, hirSchema *Schema) mir.Constraint {
	var r mir.Constraint = mir.TRUE
	//
	for _, e := range es {
		r = mir.Disjunct(r, extractCondition(e, mirSchema, hirSchema))
	}
	//
	return r
}

// Extracting from conditional expressions is slightly more complex than others,
// so it gets a case of its own.
func extractIfZeroCondition(e *IfZero, mirSchema *mir.Schema, hirSchema *Schema) mir.Constraint {
	var (
		bc mir.Constraint
		cb mir.Constraint
	)
	// Lower condition
	cc := extractCondition(e.Condition, mirSchema, hirSchema)
	body := extractBody(e.Condition, mirSchema, hirSchema)
	// Add conditions arising
	if e.TrueBranch != nil && e.FalseBranch != nil {
		// Expansion should ensure this case does not exist.  This is necessary
		// to ensure exactly one expression is generated from this expression.
		panic(fmt.Sprintf("unexpanded expression (%s)", lispOfTerm(e, hirSchema)))
	} else if e.TrueBranch != nil {
		// Lower conditional's arising from body
		cb = body.NotEqualsZero()
		bc = extractCondition(e.TrueBranch, mirSchema, hirSchema)
	} else {
		// Lower conditional's arising from body
		cb = body.EqualsZero()
		bc = extractCondition(e.FalseBranch, mirSchema, hirSchema)
	}
	//
	return mir.Disjunct(cc, cb, bc)
}

// Translate the "body" of an expression.  Every expression can be view as a
// conditional constraint of the form "if c then e", where "e" is the
// constraint.
func extractBody(e Term, mirSchema *mir.Schema, hirSchema *Schema) mir.Expr {
	switch e := e.(type) {
	case *Add:
		return mir.Sum(extractBodies(e.Args, mirSchema, hirSchema)...)
	case *Cast:
		return mir.CastOf(extractBody(e.Arg, mirSchema, hirSchema), e.BitWidth)
	case *Constant:
		return mir.NewConst(e.Value)
	case *ColumnAccess:
		return mir.NewColumnAccess(e.Column, e.Shift)
	case *Exp:
		return mir.Exponent(extractBody(e.Arg, mirSchema, hirSchema), e.Pow)
	case *IfZero:
		if e.TrueBranch != nil && e.FalseBranch != nil {
			// Expansion should ensure this case does not exist.  This is necessary
			// to ensure exactly one expression is generated from this expression.
			panic(fmt.Sprintf("unexpanded expression (%s)", lispOfTerm(e, hirSchema)))
		} else if e.TrueBranch != nil {
			return extractBody(e.TrueBranch, mirSchema, hirSchema)
		}
		// Done
		return extractBody(e.FalseBranch, mirSchema, hirSchema)
	case *LabelledConstant:
		return mir.NewConst(e.Value)
	case *Mul:
		return mir.Product(extractBodies(e.Args, mirSchema, hirSchema)...)
	case *Norm:
		return mir.Normalise(extractBody(e.Arg, mirSchema, hirSchema))
	case *Sub:
		return mir.Subtract(extractBodies(e.Args, mirSchema, hirSchema)...)
	default:
		name := reflect.TypeOf(e).Name()
		panic(fmt.Sprintf("unknown HIR expression \"%s\"", name))
	}
}

// Extract a vector of expanded expressions to the MIR level.
func extractBodies(es []Term, mirSchema *mir.Schema, hirSchema *Schema) []mir.Expr {
	rs := make([]mir.Expr, len(es))
	for i, e := range es {
		rs[i] = extractBody(e, mirSchema, hirSchema)
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
func expand(e Term, hirSchema *Schema) []Term {
	switch e := e.(type) {
	case *Add:
		return expandAdd(e, hirSchema)
	case *Cast:
		return expandCast(e, hirSchema)
	case *Constant:
		return []Term{e}
	case *ColumnAccess:
		return []Term{e}
	case *LabelledConstant:
		return []Term{e}
	case *Mul:
		return expandMul(e, hirSchema)
	case *List:
		return expandList(e, hirSchema)
	case *Exp:
		return expandExp(e, hirSchema)
	case *Norm:
		return expandNorm(e, hirSchema)
	case *IfZero:
		return expandIfZero(e, hirSchema)
	case *Sub:
		return expandSub(e, hirSchema)
	}
	// Should be unreachable
	panic(fmt.Sprintf("unknown expression: %s", lispOfTerm(e, hirSchema)))
}

func expandSub(e *Sub, hirSchema *Schema) []Term {
	return expandWithNaryConstructor(e.Args, func(nargs []Term) Term {
		return &Sub{Args: nargs}
	}, hirSchema)
}

func expandIfZero(e *IfZero, hirSchema *Schema) []Term {
	ees := make([]Term, 0)
	// Expand true branch with condition
	if e.TrueBranch != nil {
		ees = expandWithBinaryConstructor(e.Condition, e.TrueBranch, func(c Term, tb Term) Term {
			return &IfZero{c, tb, nil}
		}, hirSchema)
	}
	// Expand false branch with condition
	if e.FalseBranch != nil {
		fes := expandWithBinaryConstructor(e.Condition, e.FalseBranch, func(c Term, fb Term) Term {
			return &IfZero{c, nil, fb}
		}, hirSchema)
		ees = append(ees, fes...)
	}
	// Dne
	return ees
}

func expandNorm(e *Norm, hirSchema *Schema) []Term {
	ees := expand(e.Arg, hirSchema)
	for i, ee := range ees {
		ees[i] = &Norm{ee}
	}

	return ees
}

func expandExp(e *Exp, hirSchema *Schema) []Term {
	ees := expand(e.Arg, hirSchema)
	for i, ee := range ees {
		ees[i] = &Exp{ee, e.Pow}
	}

	return ees
}

func expandList(e *List, hirSchema *Schema) []Term {
	ees := make([]Term, 0)
	for _, arg := range e.Args {
		ees = append(ees, expand(arg, hirSchema)...)
	}

	return ees
}

func expandAdd(e *Add, hirSchema *Schema) []Term {
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
	}, hirSchema)
}

func expandCast(e *Cast, hirSchema *Schema) []Term {
	ees := expand(e.Arg, hirSchema)
	for i, ee := range ees {
		ees[i] = &Cast{ee, e.BitWidth}
	}

	return ees
}

func expandMul(e *Mul, hirSchema *Schema) []Term {
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
	}, hirSchema)
}

type binaryConstructor func(Term, Term) Term
type naryConstructor func([]Term) Term

// LowerWithBinaryConstructor is a generic mechanism for lowering down to a binary expression.
func expandWithBinaryConstructor(lhs Term, rhs Term, create binaryConstructor, hirSchema *Schema) []Term {
	var res []Term
	// Lower all three expressions
	is := expand(lhs, hirSchema)
	js := expand(rhs, hirSchema)

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
func expandWithNaryConstructor(args []Term, constructor naryConstructor, hirSchema *Schema) []Term {
	// Accumulator is initially empty
	acc := make([]Term, len(args))
	// Start from the first argument
	return expandWithNaryConstructorHelper(0, acc, args, constructor, hirSchema)
}

// LowerWithNaryConstructorHelper manages progress through the cross-product expansion.
// Specifically, "i" determines how much of args has been lowered thus
// far, whilst "acc" represents the current array being generated.
func expandWithNaryConstructorHelper(i int, acc []Term, args []Term,
	constructor naryConstructor, hirSchema *Schema) []Term {
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

	for _, ith := range expand(args[i], hirSchema) {
		acc[i] = ith
		iths := expandWithNaryConstructorHelper(i+1, acc, args, constructor, hirSchema)
		nargs = append(nargs, iths...)
	}

	return nargs
}
