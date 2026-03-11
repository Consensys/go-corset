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
package data

import (
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/symbol"
)

// SubtypeOf performs a subtype check, denoted "t1 <: t2", in a given typing
// environment.  A subtype check can be view as a subset relationship, where "t1
// <: t2" is read as saying: the set of values in t1 is a subset of those in t2.
// For example, it follows that "u4 <: u8" holds since "{0..15} ⊆ {0..255}"
// holds.  Furthermore, "u4 <: u1" does not hold as "{0..15} ⊆ {0..1}" does not.
// Finally, the subtype operator follows the general algebraic properties of the
// subset operator.  For example, it is reflesive and, hence, "u4 <: u4" holds.
//
// The only real challenge is the presence of open (i.e. existential) types such
// as "u4+". For example, does "u4+ <: u8" hold?   To understand this, we can
// interpret "u4+" as saying "this can be any type which is at least a u4" Thus,
// the query "u4+ <: u8" can be read as saying "is there a type which is at
// least a u4 that is a subtype of u8?".  The answer, of course, is yes: u8.
func SubtypeOf[S symbol.Symbol[S]](t1, t2 Type[S], env Environment[S]) bool {
	// Resolve alias types so we compare underlying types from the Ref.
	if at2 := t2.AsAlias(env); at2 != nil && at2.Ref != nil {
		return SubtypeOf(t1, at2.Resolve(env), env)
	}

	if at1 := t1.AsAlias(env); at1 != nil && at1.Ref != nil {
		return SubtypeOf(at1.Resolve(env), t2, env)
	}

	switch t1 := t1.(type) {
	case *UnsignedInt[S]:
		if t := t2.AsUint(env); t != nil {
			return t1.BitWidth() <= t.BitWidth()
		}
	case *Tuple[S]:
		if t := t2.AsTuple(env); t != nil {
			if t1.Width() != t.Width() {
				return false
			}
			//
			for i := range t1.Width() {
				if !SubtypeOf(t1.Ith(i), t.Ith(i), env) {
					return false
				}
			}
			//
			return true
		}
	}
	//
	return false
}

// EquiTypes performs an equivalence check, denoted "t1 ~ t2", in a given typing
// environment.  In most cases, this amounts to check that the two types are
// equivalent.  For example, "u8 ~ u8" holds but "u4 ~ u8" does not.  The only
// real challenge is the presence of open (i.e. existential) types such as
// "u8+".  For such types, we have that e.g. "u8 ~ u4+" hold and, likewise, that
// "u8+ ~ u16".  To understand this, we can interpret "u8+" as saying "this can
// be any type which is at least a u8" Thus, the query "u8+ ~ u16" can be read
// as saying "is there a type which is at least a u8 equivalent to a u16?".  The
// answer, of course, is yes: u16. However, we note that "u16+ ~ u8" does not
// hold.
func EquiTypes[S symbol.Symbol[S]](t1, t2 Type[S], env Environment[S]) bool {
	switch t1 := t1.(type) {
	case *UnsignedInt[S]:
		if t := t2.AsUint(env); t != nil {
			return t1.BitWidth() == t.BitWidth() ||
				(t1.IsOpen() && t1.BitWidth() <= t.BitWidth()) ||
				(t.IsOpen() && t.BitWidth() <= t1.BitWidth())
		}
	case *Tuple[S]:
		if t := t2.AsTuple(env); t != nil {
			if t1.Width() != t.Width() {
				return false
			}
			//
			for i := range t1.Width() {
				if !EquiTypes(t1.Ith(i), t.Ith(i), env) {
					return false
				}
			}
			//
			return true
		}
	}
	//
	return false
}
