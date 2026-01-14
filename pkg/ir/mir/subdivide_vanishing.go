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
package mir

import (
	"fmt"

	"github.com/consensys/go-corset/pkg/ir/term"
	"github.com/consensys/go-corset/pkg/schema/agnostic"
	"github.com/consensys/go-corset/pkg/schema/constraint/vanishing"
	"github.com/consensys/go-corset/pkg/schema/register"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/word"
	log "github.com/sirupsen/logrus"
)

// EXPLODING_MULTIPLIER determines the multipler to use for logging "exploding"
// constraints.  This is essentially an aid to debugging.
var EXPLODING_MULTIPLIER = uint(10)

// Subdivide implementation for the FieldAgnostic interface.
func (p *Subdivider[F]) subdivideVanishing(vc VanishingConstraint[F]) VanishingConstraint[F] {
	//
	var (
		modmap = p.mapping.Module(vc.Context)
		// Extract allocator
		alloc = p.FreshAllocator(vc.Context)
		// Split all registers occurring in the logical term.
		targets, context = splitLogicalTerm(vc.Constraint, term.True[F, LogicalTerm[F]](), modmap, alloc)
		// Determine size of original tree
		n = sizeOfTree(vc.Constraint, modmap)
		// Determine size of split tree
		m = sizeOfTree(targets, alloc) + sizeOfTree(context, alloc)
		//
		multiplier = float64(m) / float64(n)
	)
	// Check for any exploding constraints
	if multiplier > float64(EXPLODING_MULTIPLIER) {
		multiplier := fmt.Sprintf("%.2f", multiplier)
		log.Debug("exploding (x", multiplier, ") constraint \"", vc.Handle, "\" in module \"", modmap.Name(), "\" detected.")
	}
	// Flush allocator
	p.FlushAllocator(vc.Context, alloc)
	//
	return vanishing.NewConstraint(vc.Handle, vc.Context, vc.Domain, term.Conjunction(context, targets).Simplify(false))
}

func splitLogicalTerm[F field.Element[F]](expr LogicalTerm[F], path LogicalTerm[F], mapping register.LimbsMap,
	env agnostic.RegisterAllocator) (target LogicalTerm[F], context LogicalTerm[F]) {
	//
	switch t := expr.(type) {
	case *Conjunct[F]:
		targets, context := splitLogicalTerms(t.Args, path, mapping, env)
		return term.Conjunction(targets...), context
	case *Disjunct[F]:
		targets, context := splitLogicalTerms(t.Args, path, mapping, env)
		return term.Disjunction(targets...), context
	case *Equal[F]:
		return splitEquality(true, t.Lhs, t.Rhs, path, mapping, env)
	case *Ite[F]:
		condition, ctx1 := splitLogicalTerm(t.Condition, path, mapping, env)
		truePath := term.Conjunction(condition, path)
		falsePath := term.Conjunction(condition.Negate(), path)
		trueBranch, ctx2 := splitOptionalLogicalTerm(t.TrueBranch, truePath, mapping, env)
		falseBranch, ctx3 := splitOptionalLogicalTerm(t.FalseBranch, falsePath, mapping, env)
		//
		return term.IfThenElse(condition, trueBranch, falseBranch), term.Conjunction(ctx1, ctx2, ctx3)
	case *Negate[F]:
		target, context = splitLogicalTerm(t.Arg, path, mapping, env)
		return term.Negation(target), context
	case *NotEqual[F]:
		return splitEquality(false, t.Lhs, t.Rhs, path, mapping, env)
	default:
		panic("unreachable")
	}
}

func splitOptionalLogicalTerm[F field.Element[F]](expr LogicalTerm[F], path LogicalTerm[F],
	mapping register.LimbsMap, env agnostic.RegisterAllocator) (target LogicalTerm[F], context LogicalTerm[F]) {
	//
	if expr == nil {
		return nil, term.True[F, LogicalTerm[F]]()
	}
	//
	return splitLogicalTerm(expr, path, mapping, env)
}

func splitLogicalTerms[F field.Element[F]](terms []LogicalTerm[F], path LogicalTerm[F],
	mapping register.LimbsMap, env agnostic.RegisterAllocator) (targets []LogicalTerm[F], context LogicalTerm[F]) {
	//
	var (
		nterms = make([]LogicalTerm[F], len(terms))
		nctx   = make([]LogicalTerm[F], len(terms))
	)
	//
	for i := range len(terms) {
		nterms[i], nctx[i] = splitLogicalTerm(terms[i], path, mapping, env)
	}
	//
	return nterms, term.Conjunction(nctx...)
}

func splitEquality[F field.Element[F]](sign bool, lhs, rhs Term[F], path LogicalTerm[F], mapping register.LimbsMap,
	env agnostic.RegisterAllocator) (target LogicalTerm[F], context LogicalTerm[F]) {
	//
	var (
		alloc = newCtxRegisterAllocator(env, path)
		// Split terms accordingl to mapping, and translate into polynomials
		left  = termToPolynomial(subdivideTerm(lhs, mapping), mapping.LimbsMap())
		right = termToPolynomial(subdivideTerm(rhs, mapping), mapping.LimbsMap())
		// Construct equality for spltting
		equation = agnostic.NewEquation(left, right)
		// Split the equation
		tgtEqns, ctxEqns = equation.Split(mapping.Field(), alloc)
		// Prepare resulting conjunct / disjunct
		tgtTerms = make([]LogicalTerm[F], len(tgtEqns))
		ctxTerms = make([]LogicalTerm[F], len(ctxEqns))
	)
	// Translate target equations
	for i, eq := range tgtEqns {
		// reconstruct original term
		l := polynomialToTerm[F](eq.LeftHandSide)
		r := polynomialToTerm[F](eq.RightHandSide)
		//
		if sign {
			tgtTerms[i] = term.Equals[F, LogicalTerm[F]](l, r)
		} else {
			tgtTerms[i] = term.NotEquals[F, LogicalTerm[F]](l, r)
		}
	}
	// Translate contextual equations
	for i, eq := range ctxEqns {
		// reconstruct original term
		l := polynomialToTerm[F](eq.LeftHandSide)
		r := polynomialToTerm[F](eq.RightHandSide)
		//
		ctxTerms[i] = term.Equals[F, LogicalTerm[F]](l, r)
	}
	// construct contextual constraints
	context = term.IfThenElse(path, term.Conjunction(ctxTerms...), nil)
	// Done (for now)
	if sign {
		return term.Conjunction(tgtTerms...), context
	}
	//
	return term.Disjunction(tgtTerms...), context
}

func sizeOfTree[F field.Element[F]](term LogicalTerm[F], mapping register.Map) uint {
	switch t := term.(type) {
	case *Conjunct[F]:
		return sizeOfTrees(t.Args, mapping)
	case *Disjunct[F]:
		return sizeOfTrees(t.Args, mapping)
	case *Equal[F]:
		return 1
	case *Ite[F]:
		size := sizeOfTree(t.Condition, mapping)
		//
		if t.TrueBranch != nil {
			size += sizeOfTree(t.TrueBranch, mapping)
		}
		//
		if t.FalseBranch != nil {
			size += sizeOfTree(t.FalseBranch, mapping)
		}
		//
		return size
	case *Negate[F]:
		return sizeOfTree(t.Arg, mapping)
	case *NotEqual[F]:
		return 1
	default:
		panic("unknown logical term encountered")
	}
}

func sizeOfTrees[F field.Element[F]](terms []LogicalTerm[F], mapping register.Map) uint {
	var size uint
	//
	for _, term := range terms {
		size += sizeOfTree(term, mapping)
	}
	//
	return size
}

// ============================================================================
// Contextual Register Allocator
// ============================================================================

// Contextual Register Allocator is a register allocator which ensures that
// allocated assignments have proper context (i.e. a proper path condition).
type ctxRegisterAllocator struct {
	alloc   agnostic.RegisterAllocator
	context term.LogicalComputation[word.BigEndian]
}

func newCtxRegisterAllocator[F field.Element[F]](alloc agnostic.RegisterAllocator, path LogicalTerm[F],
) *ctxRegisterAllocator {
	return &ctxRegisterAllocator{alloc, toLogicalComputation(path.Simplify(false))}
}

// Allocate implementation for the RegisterAllocator interface
func (p *ctxRegisterAllocator) Allocate(prefix string, width uint) register.Id {
	return p.alloc.Allocate(prefix, width)
}

// AllocateWithN implementation for the RegisterAllocator interface
func (p *ctxRegisterAllocator) AllocateN(prefix string, widths []uint) []register.Id {
	return p.alloc.AllocateN(prefix, widths)
}

// AllocateWith implementation for the RegisterAllocator interface
func (p *ctxRegisterAllocator) AllocateWith(prefix string, width uint, assignment Computation) register.Id {
	// Add path condition
	assignment = term.IfElse(p.context, assignment, term.Const64[word.BigEndian, Computation](0))
	// Allocate as before
	return p.alloc.AllocateWith(prefix, width, assignment)
}

// AllocateWithN implementation for the RegisterAllocator interface
func (p *ctxRegisterAllocator) AllocateWithN(prefix string, assignment Computation, widths ...uint) []register.Id {
	// Add path condition
	assignment = term.IfElse(p.context, assignment, term.Const64[word.BigEndian, Computation](0))
	// Allocate as before
	return p.alloc.AllocateWithN(prefix, assignment, widths...)
}

// Assign implementation for the RegisterAllocator interface
func (p *ctxRegisterAllocator) Assignments() []util.Pair[[]register.Id, Computation] {
	return p.alloc.Assignments()
}

// Name implementation for RegisterMapping interface
func (p *ctxRegisterAllocator) Name() trace.ModuleName {
	return p.alloc.Name()
}

// HasRegister implementation for RegisterMap interface.
func (p *ctxRegisterAllocator) HasRegister(name string) (register.Id, bool) {
	return p.alloc.HasRegister(name)
}

// Register implementation for RegisterMap interface.
func (p *ctxRegisterAllocator) Register(rid register.Id) register.Register {
	return p.alloc.Register(rid)
}

// Registers implementation for RegisterMap interface.
func (p *ctxRegisterAllocator) Registers() []register.Register {
	return p.alloc.Registers()
}

// Reset implementation for RegisterAllocator interface.
func (p *ctxRegisterAllocator) Reset(n uint) {
	p.alloc.Reset(n)
}

func (p *ctxRegisterAllocator) String() string {
	return p.alloc.String()
}

// ============================================================================
// Computation Conversion
// ============================================================================

func toLogicalComputation[F field.Element[F]](t LogicalTerm[F]) term.LogicalComputation[word.BigEndian] {
	switch t := t.(type) {
	case *Conjunct[F]:
		args := toLogicalComputations(t.Args)
		return term.Conjunction(args...)
	case *Disjunct[F]:
		args := toLogicalComputations(t.Args)
		return term.Disjunction(args...)
	case *Equal[F]:
		lhs := toComputation[F](t.Lhs)
		rhs := toComputation[F](t.Rhs)

		return term.Equals[word.BigEndian, LogicalComputation](lhs, rhs)
	case *Ite[F]:
		var trueBranch, falseBranch LogicalComputation

		condition := toLogicalComputation[F](t.Condition)
		//
		if t.TrueBranch != nil {
			trueBranch = toLogicalComputation[F](t.TrueBranch)
		}
		//
		if t.FalseBranch != nil {
			falseBranch = toLogicalComputation[F](t.FalseBranch)
		}
		//
		return term.IfThenElse(condition, trueBranch, falseBranch)
	case *Negate[F]:
		arg := toLogicalComputation[F](t.Arg)
		return term.Negation(arg)
	case *NotEqual[F]:
		lhs := toComputation[F](t.Lhs)
		rhs := toComputation[F](t.Rhs)

		return term.NotEquals[word.BigEndian, LogicalComputation](lhs, rhs)
	default:
		panic(fmt.Sprintf("unknown computation encountered: %s", t.Lisp(false, nil).String(false)))
	}
}

func toLogicalComputations[F field.Element[F]](terms []LogicalTerm[F]) []LogicalComputation {
	var computations = make([]LogicalComputation, len(terms))
	//
	for i, t := range terms {
		computations[i] = toLogicalComputation[F](t)
	}
	//
	return computations
}

func toComputation[F field.Element[F]](t Term[F]) term.Computation[word.BigEndian] {
	switch t := t.(type) {
	case *Add[F]:
		args := toComputations[F](t.Args)
		return term.Sum(args...)
	case *Constant[F]:
		var value word.BigEndian

		return term.Const[word.BigEndian, Computation](value.SetBytes(t.Value.Bytes()))
	case *Mul[F]:
		args := toComputations[F](t.Args)
		return term.Product(args...)
	case *RegisterAccess[F]:
		return term.RawRegisterAccess[word.BigEndian, Computation](
			t.Register(), t.BitWidth(), t.RelativeShift()).Mask(t.MaskWidth())
	case *Sub[F]:
		args := toComputations[F](t.Args)
		return term.Subtract(args...)
	case *VectorAccess[F]:
		var nterms = make([]*term.RegisterAccess[word.BigEndian, Computation], len(t.Vars))
		//
		for i, v := range t.Vars {
			nterms[i] = term.RawRegisterAccess[word.BigEndian, Computation](
				v.Register(), v.BitWidth(), v.RelativeShift()).Mask(v.MaskWidth())
		}
		//
		return term.NewVectorAccess(nterms)
	default:
		panic(fmt.Sprintf("unknown computation encountered: %s", t.Lisp(false, nil).String(false)))
	}
}

func toComputations[F field.Element[F]](terms []Term[F]) []Computation {
	var computations = make([]Computation, len(terms))
	//
	for i, t := range terms {
		computations[i] = toComputation[F](t)
	}
	//
	return computations
}
