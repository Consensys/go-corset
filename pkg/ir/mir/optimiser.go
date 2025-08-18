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
	"reflect"

	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/math"
)

// OptimisationConfig provides a mechanism for controlling how optimisations are
// applied during MIR lowering.
type OptimisationConfig struct {
	// InverseEliminationLevel sets an upper bound on the range cardinality at
	// which inverses will be eliminated in favour of constraints.  A level of 0
	// means no inverses will be eliminated, a range of 1 means only trivial
	// ranges (i.e. {-1,0}, {0,1} and {-1,0,1}) will be eliminated; Otherwise,
	// the level indicates the range cardinality.  For example, level 2 means
	// any range of cardinality 2 is eliminated (e.g. {1,2}, {5,6}, etc).
	InverseEliminiationLevel uint
	// MaxRangeConstraint determines the largest bitwidth for which range
	// constraints are translated into AIR range constraints, versus  using a
	// horizontal bitwidth gadget.
	MaxRangeConstraint uint
	// ShiftNormalisation is an optimisation for inverse columns involving
	// shifts.
	ShiftNormalisation bool
	// LimitlessTypeProofs enables the use of type proofs which exploit the
	// limitless prover. Specifically, modules with a recursive structure are
	// created specifically for the purpose of checking types.
	LimitlessTypeProofs bool
}

// OPTIMISATION_LEVELS provides a set of precanned optimisation configurations.
// Here 0 implies no optimisation and, otherwise, increasing levels implies
// increasingly aggressive optimisation (though that doesn't mean they will
// always improve performance).
var OPTIMISATION_LEVELS = []OptimisationConfig{
	// Level 0 == nothing enabled
	{0, 8, false, false},
	// Level 1 == minimal optimisations applied.
	{1, 16, true, true},
}

// DEFAULT_OPTIMISATION_INDEX gives the index of the default optimisation level
// in OPTIMISATION_LEVELS.
var DEFAULT_OPTIMISATION_INDEX = uint(1)

// DEFAULT_OPTIMISATION_LEVEL provides a default level of optimisation which
// should be used in most cases.
var DEFAULT_OPTIMISATION_LEVEL = OPTIMISATION_LEVELS[DEFAULT_OPTIMISATION_INDEX]

// attempt to eliminate normalisations by undertaking a range analysis on their
// arguments to see whether they have sufficiently small ranges.
func eliminateNormalisationInTerm[F field.Element[F]](term Term[F], module schema.Module[F],
	cfg OptimisationConfig) Term[F] {
	switch term := term.(type) {
	case *Add[F]:
		args := eliminateNormalisationInTerms(term.Args, module, cfg)
		return &Add[F]{Args: args}
	case *Cast[F]:
		arg := eliminateNormalisationInTerm(term.Arg, module, cfg)
		return &Cast[F]{Arg: arg, BitWidth: term.BitWidth}
	case *Constant[F]:
		return term
	case *RegisterAccess[F]:
		return term
	case *Exp[F]:
		arg := eliminateNormalisationInTerm(term.Arg, module, cfg)
		return &Exp[F]{Arg: arg, Pow: term.Pow}
	case *Mul[F]:
		args := eliminateNormalisationInTerms(term.Args, module, cfg)
		return &Mul[F]{Args: args}
	case *Norm[F]:
		return eliminateNormalisationInNorm(term.Arg, module, cfg)
	case *Sub[F]:
		args := eliminateNormalisationInTerms(term.Args, module, cfg)
		return &Sub[F]{Args: args}
	case *VectorAccess[F]:
		return term
	default:
		name := reflect.TypeOf(term).Name()
		panic(fmt.Sprintf("unknown MIR expression \"%s\"", name))
	}
}

func eliminateNormalisationInTerms[F field.Element[F]](terms []Term[F], module schema.Module[F],
	cfg OptimisationConfig) []Term[F] {
	nterms := make([]Term[F], len(terms))
	//
	for i, t := range terms {
		nterms[i] = eliminateNormalisationInTerm(t, module, cfg)
	}
	//
	return nterms
}

func eliminateNormalisationInNorm[F field.Element[F]](arg Term[F], module schema.Module[F], cfg OptimisationConfig) Term[F] {
	bounds := arg.ValueRange(module)
	// optimise argument
	arg = eliminateNormalisationInTerm(arg, module, cfg)
	// Check whether normalisation actually required.  For example, if the
	// argument is just a binary column then a normalisation is not actually
	// required.
	if cfg.InverseEliminiationLevel > 0 && bounds.Within(math.NewInterval64(0, 1)) {
		// arg ∈ {0,1} ==> normalised already :)
		return arg
	} else if cfg.InverseEliminiationLevel > 0 && bounds.Within(math.NewInterval64(-1, 1)) {
		// arg ∈ {-1,0,1} ==> (arg*arg) ∈ {0,1}
		return &Mul[F]{Args: []Term[F]{arg, arg}}
	}
	// Nothing happening
	return &Norm[F]{Arg: arg}
}
