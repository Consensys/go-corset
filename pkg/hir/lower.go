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
		p := lowerToConstraint(c.Property, mirSchema, p)
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
		mir_constraint := lowerToConstraint(v.Constraint, mirSchema, hirSchema)
		// Add translated constraint
		mirSchema.AddVanishingConstraint(v.Handle, 0, v.Context, v.Domain, mir_constraint)
	} else if v, ok := c.(RangeConstraint); ok {
		mir_expr := lowerToUnit(v.Expr, mirSchema, hirSchema)
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
		from[i] = lowerToUnit(c.Sources[i], mirSchema, hirSchema)
		into[i] = lowerToUnit(c.Targets[i], mirSchema, hirSchema)
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
		selector = util.Some(lowerToUnit(c.Selector.Unwrap(), mirSchema, hirSchema))
	}
	// Convert general expressions into unit expressions.
	for i := 0; i < len(sources); i++ {
		sources[i] = lowerToUnit(c.Sources[i], mirSchema, hirSchema)
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
	b := extractExpression(e.Term, mirSchema, hirSchema)
	//
	return b.Simplify()
}

// Extract the "condition" of an expression.  Every expression can be view as a
// conditional constraint of the form "if c then e", where "c" is the condition.
// This is allowed to return nil if the body is unconditional.
func extractConstraint(t Term, mirSchema *mir.Schema, hirSchema *Schema) mir.Constraint {
	switch e := t.(type) {
	case *Cast:
		return extractConstraint(e.Arg, mirSchema, hirSchema)
	case *Connective:
		constraints := make([]mir.Constraint, len(e.Args))
		//
		for i, t := range e.Args {
			constraints[i] = extractConstraint(t, mirSchema, hirSchema)
		}
		//
		if e.Sign {
			return mir.Disjunct(constraints...)
		}
		//
		return mir.Conjunct(constraints...)
	case *Equation:
		lhs := extractExpression(e.Lhs, mirSchema, hirSchema)
		rhs := extractExpression(e.Rhs, mirSchema, hirSchema)
		//
		switch e.Kind {
		case EQUALS:
			return mir.Equals(lhs, rhs)
		case NOT_EQUALS:
			return mir.NotEquals(lhs, rhs)
		case LESS_THAN:
			return mir.LessThan(lhs, rhs)
		case LESS_THAN_EQUALS:
			return mir.LessThanOrEquals(lhs, rhs)
		case GREATER_THAN_EQUALS:
			return mir.GreaterThanOrEquals(lhs, rhs)
		case GREATER_THAN:
			return mir.GreaterThan(lhs, rhs)
		default:
			panic("unreachable")
		}
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
	case *Not:
		arg := extractConstraint(e.Arg, mirSchema, hirSchema)
		return mir.Negate(arg)
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

func extractExpression(e Term, mirSchema *mir.Schema, hirSchema *Schema) mir.Expr {
	switch e := e.(type) {
	case *Add:
		args := extractBodies(e.Args, mirSchema, hirSchema)
		return mir.Sum(args...)
	case *Cast:
		arg := extractExpression(e.Arg, mirSchema, hirSchema)
		return mir.CastOf(arg, e.BitWidth)
	case *Constant:
		return mir.NewConst(e.Value)
	case *ColumnAccess:
		return mir.NewColumnAccess(e.Column, e.Shift)
	case *Exp:
		arg := extractExpression(e.Arg, mirSchema, hirSchema)
		return mir.Exponent(arg, e.Pow)
	case *IfZero:
		// Translate the condition
		c := extractCondition(true, e.Condition, mirSchema, hirSchema)
		neg_c := extractCondition(false, e.Condition, mirSchema, hirSchema)
		//
		if e.TrueBranch == nil || e.FalseBranch == nil {
			// Expansion should ensure this case does not exist.  This is necessary
			// to ensure exactly one expression is generated from this expression.
			panic(fmt.Sprintf("unbalanced condition encountered (%s)", lispOfTerm(e, hirSchema)))
		}
		//
		tb := extractExpression(e.TrueBranch, mirSchema, hirSchema)
		fb := extractExpression(e.FalseBranch, mirSchema, hirSchema)
		//
		tb = mir.Product(neg_c, tb)
		fb = mir.Product(c, fb)
		//
		return mir.Sum(tb, fb)
	case *LabelledConstant:
		return mir.NewConst(e.Value)
	case *Mul:
		args := extractBodies(e.Args, mirSchema, hirSchema)
		return mir.Product(args...)
	case *Norm:
		arg := extractExpression(e.Arg, mirSchema, hirSchema)
		return mir.Normalise(arg)
	case *Sub:
		args := extractBodies(e.Args, mirSchema, hirSchema)
		return mir.Subtract(args...)
	default:
		name := reflect.TypeOf(e).Name()
		panic(fmt.Sprintf("unknown HIR expression \"%s\"", name))
	}
}

func extractCondition(sign bool, e Term, mirSchema *mir.Schema, hirSchema *Schema) mir.Expr {
	switch e := e.(type) {
	case *IfZero:
		panic("todo")
	case *Connective:
		if sign == e.Sign {
			// Disjunction
			args := extractConditions(sign, e.Args, mirSchema, hirSchema)
			//
			return mir.Product(args...)
		}
		// Conjunction
		args := extractConditions(!sign, e.Args, mirSchema, hirSchema)
		// P && Q ==> !(!P || Q!) ==> 1 - ~(!P || !Q)
		return mir.Subtract(mir.NewConst64(1), mir.Normalise(mir.Product(args...)))
	case *Equation:
		l := extractExpression(e.Lhs, mirSchema, hirSchema)
		r := extractExpression(e.Rhs, mirSchema, hirSchema)
		t := mir.Normalise(mir.Subtract(l, r))
		//
		kind := invertAtomicEquation(sign, e.Kind)
		//
		switch kind {
		case EQUALS:
			return t
		case NOT_EQUALS:
			return mir.Subtract(mir.NewConst64(1), t)
		default:
			panic("unreachable")
		}
	default:
		name := reflect.TypeOf(e).Name()
		panic(fmt.Sprintf("unknown HIR expression \"%s\"", name))
	}
}

func extractConditions(sign bool, es []Term, mirSchema *mir.Schema, hirSchema *Schema) []mir.Expr {
	exprs := make([]mir.Expr, len(es))
	//
	for i, e := range es {
		exprs[i] = extractCondition(sign, e, mirSchema, hirSchema)
	}
	//
	return exprs
}

func invertAtomicEquation(sign bool, kind uint8) uint8 {
	if sign {
		return kind
	}
	//
	switch kind {
	case EQUALS:
		return NOT_EQUALS
	case NOT_EQUALS:
		return EQUALS
	case LESS_THAN:
		return GREATER_THAN_EQUALS
	case LESS_THAN_EQUALS:
		return GREATER_THAN
	case GREATER_THAN:
		return LESS_THAN_EQUALS
	case GREATER_THAN_EQUALS:
		return LESS_THAN
	default:
		panic("unreachable")
	}
}

// Extract a vector of expanded expressions to the MIR level.
func extractBodies(es []Term, mirSchema *mir.Schema, hirSchema *Schema) []mir.Expr {
	rs := make([]mir.Expr, len(es))

	for i, e := range es {
		rs[i] = extractExpression(e, mirSchema, hirSchema)
	}
	//
	return rs
}
