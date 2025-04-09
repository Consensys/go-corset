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
		p := lowerToConstraint(c.Property.Expr, mirSchema, p)
		mirSchema.AddPropertyAssertion(c.Handle, c.Context, p)
	}
	//
	return mirSchema
}

func lowerConstraintToMir(c sc.Constraint, mirSchema *mir.Schema, hirSchema *Schema) {
	// Check what kind of constraint we have
	if v, ok := c.(LookupConstraint); ok {
		lowerLookupConstraint(v, mirSchema, hirSchema)
	} else if v, ok := c.(VanishingConstraint); ok {
		mir_constraint := lowerToConstraint(v.Constraint.Expr, mirSchema, hirSchema)
		// Add translated constraint
		mirSchema.AddVanishingConstraint(v.Handle, 0, v.Context, v.Domain, mir_constraint)
	} else if v, ok := c.(RangeConstraint); ok {
		mir_expr := lowerToUnit(v.Expr.Expr, mirSchema, hirSchema)
		// Add individual constraints arising
		mirSchema.AddRangeConstraint(v.Handle, 0, v.Context, mir_expr, v.Bound)
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
		from[i] = lowerToUnit(c.Sources[i].Expr, mirSchema, hirSchema)
		into[i] = lowerToUnit(c.Targets[i].Expr, mirSchema, hirSchema)
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
		selector = util.Some(lowerToUnit(c.Selector.Unwrap().Expr, mirSchema, hirSchema))
	}
	// Convert general expressions into unit expressions.
	for i := 0; i < len(sources); i++ {
		sources[i] = lowerToUnit(c.Sources[i].Expr, mirSchema, hirSchema)
	}
	//
	mirSchema.AddSortedConstraint(c.Handle, c.Context, c.BitWidth, selector, sources, c.Signs, c.Strict)
}

// ============================================================================
// lowerTo
// ============================================================================

// Lowers a given expression to the MIR level.  The expression is first expanded
// into one or more target expressions. Furthermore, conditions must be "lifted"
// to the root.
func lowerToConstraint(e Expr, mirSchema *mir.Schema, hirSchema *Schema) mir.Constraint {
	return extractConstraint(e.Term, mirSchema, hirSchema)
}

// Lowers a given expression to the MIR level.  The expression is first expanded
// into one or more target expressions. Furthermore, conditions must be "lifted"
// to the root.
func lowerToUnit(e Expr, mirSchema *mir.Schema, hirSchema *Schema) mir.Expr {
	c, b := extractExpression(e.Term, mirSchema, hirSchema)
	c = mir.Negate(c)
	//
	exprs := c.AsExprs()
	//
	if len(exprs) != 1 {
		panic("attempting to lower non-unit expression")
	}
	//
	return mir.Product(exprs[0], b).Simplify()
}

// Extract the "condition" of an expression.  Every expression can be view as a
// conditional constraint of the form "if c then e", where "c" is the condition.
// This is allowed to return nil if the body is unconditional.
func extractConstraint(t Term, mirSchema *mir.Schema, hirSchema *Schema) mir.Constraint {
	switch e := t.(type) {
	case *Cast:
		return extractConstraint(e.Arg, mirSchema, hirSchema)
	case *Equation:
		cl, l := extractExpression(e.Lhs, mirSchema, hirSchema)
		cr, r := extractExpression(e.Rhs, mirSchema, hirSchema)
		//
		if e.Sign {
			return mir.Disjunct(cl, cr, mir.Equals(l, r))
		}
		//
		return mir.Disjunct(cl, cr, mir.NotEquals(l, r))
	case *IfZero:
		return extractIfZeroCondition(e, mirSchema, hirSchema)
	case *List:
		constraints := make([]mir.Constraint, len(e.Args))
		//
		for i, t := range e.Args {
			constraints[i] = extractConstraint(t, mirSchema, hirSchema)
		}
		//
		return mir.Conjunct(constraints...)
	default:
		panic(fmt.Sprintf("unknown HIR constraint \"%s\"", lispOfTerm(t, hirSchema).String(false)))
	}
}

func extractIfZeroCondition(e *IfZero, mirSchema *mir.Schema, hirSchema *Schema) mir.Constraint {
	var cases []mir.Constraint
	// Lower condition
	cc := extractConstraint(e.Condition, mirSchema, hirSchema)
	//
	if e.TrueBranch != nil {
		ncc := mir.Negate(cc)
		bc := extractConstraint(e.TrueBranch, mirSchema, hirSchema)
		cases = append(cases, mir.Disjunct(ncc, bc))
	}
	//
	if e.FalseBranch != nil {
		bc := extractConstraint(e.FalseBranch, mirSchema, hirSchema)
		cases = append(cases, mir.Disjunct(cc, bc))
	}
	//
	return mir.Conjunct(cases...)
}

func extractExpression(e Term, mirSchema *mir.Schema, hirSchema *Schema) (mir.Constraint, mir.Expr) {
	switch e := e.(type) {
	case *Add:
		c, args := extractBodies(e.Args, mirSchema, hirSchema)
		return c, mir.Sum(args...)
	case *Cast:
		c, arg := extractExpression(e.Arg, mirSchema, hirSchema)
		return c, mir.CastOf(arg, e.BitWidth)
	case *Constant:
		return mir.TRUE, mir.NewConst(e.Value)
	case *ColumnAccess:
		return mir.TRUE, mir.NewColumnAccess(e.Column, e.Shift)
	case *Exp:
		c, arg := extractExpression(e.Arg, mirSchema, hirSchema)
		return c, mir.Exponent(arg, e.Pow)
	case *IfZero:
		// var (
		// 	condition = extractConstraint(e.Condition, mirSchema, hirSchema)
		// 	bodycond  mir.Constraint
		// 	body      mir.Expr
		// )

		// if e.TrueBranch != nil && e.FalseBranch != nil {
		// 	// Expansion should ensure this case does not exist.  This is necessary
		// 	// to ensure exactly one expression is generated from this expression.
		// 	panic(fmt.Sprintf("unexpanded expression (%s)", lispOfTerm(e, hirSchema)))
		// } else if e.TrueBranch != nil {
		// 	bodycond, body = extractExpression(e.TrueBranch, mirSchema, hirSchema)
		// } else {
		// 	condition = mir.Negate(condition)
		// 	bodycond, body = extractExpression(e.FalseBranch, mirSchema, hirSchema)
		// }
		// //
		// return mir.Conjunct(condition, bodycond), body
		panic("not needed?")
	case *LabelledConstant:
		return mir.TRUE, mir.NewConst(e.Value)
	case *Mul:
		c, args := extractBodies(e.Args, mirSchema, hirSchema)
		return c, mir.Product(args...)
	case *Norm:
		c, arg := extractExpression(e.Arg, mirSchema, hirSchema)
		return c, mir.Normalise(arg)
	case *Sub:
		c, args := extractBodies(e.Args, mirSchema, hirSchema)
		return c, mir.Subtract(args...)
	default:
		name := reflect.TypeOf(e).Name()
		panic(fmt.Sprintf("unknown HIR expression \"%s\"", name))
	}
}

// Extract a vector of expanded expressions to the MIR level.
func extractBodies(es []Term, mirSchema *mir.Schema, hirSchema *Schema) (mir.Constraint, []mir.Expr) {
	rs := make([]mir.Expr, len(es))
	cs := make([]mir.Constraint, len(es))

	for i, e := range es {
		cs[i], rs[i] = extractExpression(e, mirSchema, hirSchema)
	}
	//
	return mir.Disjunct(cs...), rs
}
