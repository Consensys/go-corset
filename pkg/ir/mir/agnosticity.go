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
	"github.com/consensys/go-corset/pkg/ir"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/util/field"
)

// Subdivide implementation for the FieldAgnostic interface.
func subdivideAssertion[F field.Element[F]](c Assertion[F], _ schema.LimbsMap) Assertion[F] {
	// TODO: implement this
	return c
}

// Subdivide implementation for the FieldAgnostic interface.
func subdivideInterleaving[F field.Element[F]](c InterleavingConstraint[F], _ schema.LimbsMap,
) InterleavingConstraint[F] {
	// TODO: implement this
	return c
}

// Subdivide implementation for the FieldAgnostic interface.
func subdividePermutation[F field.Element[F]](c PermutationConstraint[F], _ schema.LimbsMap,
) PermutationConstraint[F] {
	// TODO: implement this
	return c
}

// Subdivide implementation for the FieldAgnostic interface.
func subdivideRange[F field.Element[F]](c RangeConstraint[F], _ schema.LimbsMap) RangeConstraint[F] {
	// TODO: implement this
	return c
}

// Subdivide implementation for the FieldAgnostic interface.
func subdivideSorted[F field.Element[F]](c SortedConstraint[F], _ schema.LimbsMap) SortedConstraint[F] {
	// TODO: implement this
	return c
}

func splitTerm[F field.Element[F]](term Term[F], mapping schema.RegisterAllocator) Term[F] {
	switch t := term.(type) {
	case *Add[F]:
		return ir.Sum(splitTerms(t.Args, mapping)...)
	case *Constant[F]:
		return t
	case *RegisterAccess[F]:
		return splitRegisterAccess(t, mapping)
	case *Mul[F]:
		return ir.Product(splitTerms(t.Args, mapping)...)
	case *Sub[F]:
		return ir.Subtract(splitTerms(t.Args, mapping)...)
	case *VectorAccess[F]:
		return splitVectorAccess(t, mapping)
	default:
		panic("unreachable")
	}
}

func splitTerms[F field.Element[F]](terms []Term[F], mapping schema.RegisterAllocator) []Term[F] {
	var nterms []Term[F] = make([]Term[F], len(terms))
	//
	for i := range len(terms) {
		nterms[i] = splitTerm(terms[i], mapping)
	}
	//
	return nterms
}

func splitRegisterAccess[F field.Element[F]](term *RegisterAccess[F], mapping schema.RegisterAllocator) Term[F] {
	var (
		// Determine limbs for this register
		limbs = mapping.LimbIds(term.Register)
		// Construct appropriate terms
		terms = make([]*RegisterAccess[F], len(limbs))
	)
	//
	for i, limb := range limbs {
		terms[i] = &ir.RegisterAccess[F, Term[F]]{Register: limb, Shift: term.Shift}
	}
	// Check whether vector required, or not
	if len(limbs) == 1 {
		// NOTE: we cannot return the original term directly, as its index may
		// differ under the limb mapping.
		return terms[0]
	}
	//
	return ir.NewVectorAccess(terms)
}

func splitVectorAccess[F field.Element[F]](term *VectorAccess[F], mapping schema.RegisterAllocator) Term[F] {
	var terms []*RegisterAccess[F]
	//
	for _, v := range term.Vars {
		for _, limb := range mapping.LimbIds(v.Register) {
			term := &ir.RegisterAccess[F, Term[F]]{Register: limb, Shift: v.Shift}
			terms = append(terms, term)
		}
	}
	//
	return ir.NewVectorAccess(terms)
}

func splitRawRegisterAccesses[F field.Element[F]](terms []*RegisterAccess[F], mapping schema.RegisterLimbsMap,
) []*VectorAccess[F] {
	//
	var (
		vecs = make([]*VectorAccess[F], len(terms))
	)
	//
	for i, term := range terms {
		vecs[i] = splitRawRegisterAccess(term, mapping)
	}
	//
	return vecs
}

func splitRawRegisterAccess[F field.Element[F]](term *RegisterAccess[F], mapping schema.RegisterLimbsMap,
) *VectorAccess[F] {
	//
	var (
		// Determine limbs for this register
		limbs = mapping.LimbIds(term.Register)
		// Construct appropriate terms
		terms = make([]*RegisterAccess[F], len(limbs))
	)
	//
	for i, limb := range limbs {
		terms[i] = &ir.RegisterAccess[F, Term[F]]{Register: limb, Shift: term.Shift}
	}
	//
	return ir.RawVectorAccess(terms)
}
