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
	"github.com/consensys/go-corset/pkg/schema/constraint/interleaving"
	"github.com/consensys/go-corset/pkg/schema/constraint/permutation"
	"github.com/consensys/go-corset/pkg/schema/constraint/ranged"
	"github.com/consensys/go-corset/pkg/schema/constraint/sorted"
	"github.com/consensys/go-corset/pkg/schema/module"
	"github.com/consensys/go-corset/pkg/schema/register"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/field"
)

// Subdivide implementation for the FieldAgnostic interface.
func subdivideAssertion[F field.Element[F]](c Assertion[F], _ module.LimbsMap) Assertion[F] {
	// TODO: implement this
	return c
}

// Subdivide implementation for the FieldAgnostic interface.
func subdivideInterleaving[F field.Element[F]](c InterleavingConstraint[F], mapping module.LimbsMap,
) InterleavingConstraint[F] {
	var (
		targetModule = mapping.Module(c.TargetContext)
		sourceModule = mapping.Module(c.SourceContext)
		target       = splitVectorAccess(c.Target, targetModule)
		sources      = splitVectorAccesses(c.Sources, sourceModule)
	)
	// Done
	return interleaving.NewConstraint(c.Handle, c.TargetContext, c.SourceContext, target, sources)
}

// Subdivide implementation for the FieldAgnostic interface.
func subdividePermutation[F field.Element[F]](c PermutationConstraint[F], mapping module.LimbsMap,
) PermutationConstraint[F] {
	var (
		module  = mapping.Module(c.Context)
		sources []register.Id
		targets []register.Id
	)
	//
	for i := range len(c.Sources) {
		var (
			sourceLimbs = module.LimbIds(c.Sources[i])
			targetLimbs = module.LimbIds(c.Targets[i])
		)
		// Sanity check for now
		if len(sourceLimbs) != len(targetLimbs) {
			panic("encountered irregular permutation constraint")
		}
		//
		sources = append(sources, sourceLimbs...)
		targets = append(targets, targetLimbs...)
	}
	//
	return permutation.NewConstraint[F](c.Handle, c.Context, targets, sources)
}

// Subdivide implementation for the FieldAgnostic interface.
func subdivideRange[F field.Element[F]](c RangeConstraint[F], mapping module.LimbsMap) RangeConstraint[F] {
	var (
		modmap    = mapping.Module(c.Context)
		terms     []*RegisterAccess[F]
		bitwidths []uint
	)
	//
	for i, source := range c.Sources {
		var (
			split    = splitRawRegisterAccess(source, modmap)
			bitwidth = c.Bitwidths[i]
		)
		// Include all registers
		terms = append(terms, split...)
		// Split bitwidths
		for _, jth := range split {
			var limbWidth = jth.MaskWidth()
			//
			bitwidths = append(bitwidths, min(bitwidth, limbWidth))
			//
			if bitwidth >= limbWidth {
				bitwidth -= limbWidth
			} else {
				bitwidth = 0
			}
		}
	}
	//
	return ranged.NewConstraint(c.Handle, c.Context, terms, bitwidths)
}

// Subdivide implementation for the FieldAgnostic interface.
func subdivideSorted[F field.Element[F]](c SortedConstraint[F], mapping module.LimbsMap) SortedConstraint[F] {
	var (
		modmap   = mapping.Module(c.Context)
		signs    []bool
		sources  []*RegisterAccess[F]
		selector = util.None[*RegisterAccess[F]]()
		bitwidth uint
	)
	// Split sources
	for i, source := range c.Sources {
		var split = splitRawRegisterAccess(source, modmap)
		// Append in reverse order to ensure most signicant limb comes first.
		for j := len(split); j > 0; j-- {
			var (
				jth       = split[j-1]
				limbWidth = modmap.Limb(jth.Register()).Width
			)
			//
			sources = append(sources, jth)
			// Update sign (if applicable)
			if i < len(c.Signs) {
				signs = append(signs, c.Signs[i])
			}
			// Update bitwidth
			bitwidth = max(bitwidth, min(limbWidth, jth.MaskWidth()))
		}
	}
	// Split optional selector
	if c.Selector.HasValue() {
		tmp := splitRawRegisterAccess(c.Selector.Unwrap(), modmap)
		//
		if len(tmp) != 1 {
			panic(fmt.Sprintf("encountered irregular selectored with %d limbs.", len(tmp)))
		}
		//
		selector = util.Some(tmp[0])
	}
	// Done
	return sorted.NewConstraint(c.Handle, c.Context, bitwidth, selector, sources, signs, c.Strict)
}

func splitTerm[F field.Element[F]](expr Term[F], mapping register.LimbsMap) Term[F] {
	switch t := expr.(type) {
	case *Add[F]:
		return term.Sum(splitTerms(t.Args, mapping)...)
	case *Constant[F]:
		return t
	case *RegisterAccess[F]:
		return splitRegisterAccess(t, mapping)
	case *Mul[F]:
		return term.Product(splitTerms(t.Args, mapping)...)
	case *Sub[F]:
		return term.Subtract(splitTerms(t.Args, mapping)...)
	case *VectorAccess[F]:
		return splitVectorAccess(t, mapping)
	default:
		panic("unreachable")
	}
}

func splitTerms[F field.Element[F]](terms []Term[F], mapping register.LimbsMap) []Term[F] {
	var nterms []Term[F] = make([]Term[F], len(terms))
	//
	for i := range len(terms) {
		nterms[i] = splitTerm(terms[i], mapping)
	}
	//
	return nterms
}

func splitRegisterAccess[F field.Element[F]](expr *RegisterAccess[F], mapping register.LimbsMap) Term[F] {
	var (
		// Construct appropriate terms
		terms = splitRawRegisterAccess(expr, mapping)
	)
	// Check whether vector required, or not
	if len(terms) == 1 {
		// NOTE: we cannot return the original term directly, as its index may
		// differ under the limb mapping.
		return terms[0]
	}
	//
	return term.NewVectorAccess(terms)
}

func splitVectorAccesses[F field.Element[F]](terms []*VectorAccess[F], mapping register.LimbsMap) []*VectorAccess[F] {
	var (
		nterms = make([]*VectorAccess[F], len(terms))
	)
	// Split sources
	for i, src := range terms {
		nterms[i] = splitVectorAccess(src, mapping)
	}
	//
	return nterms
}

func splitVectorAccess[F field.Element[F]](expr *VectorAccess[F], mapping register.LimbsMap) *VectorAccess[F] {
	var terms []*RegisterAccess[F]
	//
	for _, v := range expr.Vars {
		var ith = splitRawRegisterAccess(v, mapping)
		//
		terms = append(terms, ith...)
	}
	//
	return term.RawVectorAccess(terms)
}

func splitRawRegisterAccesses[F field.Element[F]](terms []*RegisterAccess[F], mapping register.LimbsMap,
) []*VectorAccess[F] {
	//
	var (
		vecs = make([]*VectorAccess[F], len(terms))
	)
	//
	for i, t := range terms {
		ith := splitRawRegisterAccess(t, mapping)
		vecs[i] = term.RawVectorAccess(ith)
	}
	//
	return vecs
}

func splitRawRegisterAccess[F field.Element[F]](expr *RegisterAccess[F], mapping register.LimbsMap,
) []*RegisterAccess[F] {
	//
	var (
		// Determine limbs for this register
		limbs = mapping.LimbIds(expr.Register())
		// Construct appropriate terms
		terms []*RegisterAccess[F]
		//
		bitwidth = expr.MaskWidth()
	)
	//
	for _, limbId := range limbs {
		var (
			limb      = mapping.Limb(limbId)
			limbWidth = min(bitwidth, limb.Width)
		)
		//
		if limbWidth > 0 {
			// Construct register access
			ith := term.RawRegisterAccess[F, Term[F]](limbId, limb.Width, expr.RelativeShift())
			// Mask off any unrequired bits
			terms = append(terms, ith.Mask(limbWidth))
		}
		//
		bitwidth -= limbWidth
	}
	//
	return terms
}
