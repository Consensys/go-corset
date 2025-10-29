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
package array

import "fmt"

// Vector represents a vector of identically sized arrays.  This just makes it
// easier to work with such vectors, providing various bits of helpful
// functionality.
type Vector[T any] struct {
	length uint
	limbs  []Array[T]
}

// VectorOf constructs a new vector array from a given set of arrays.  All
// supplied arrays are expected to have the same height, otherwise this will
// panic.
func VectorOf[T any](arrays ...Array[T]) Vector[T] {
	var length uint
	//
	for i, arr := range arrays {
		if i == 0 {
			length = arr.Len()
		} else if length != arr.Len() {
			panic(fmt.Sprintf("incompatible array lengths (%d vs %d)", length, arr.Len()))
		}
	}
	//
	return Vector[T]{length, arrays}
}

// All checks whether, for all limbs, the given predicate holds at a given
// index.
func (p *Vector[T]) All(index uint, f Predicate[T]) bool {
	for _, limb := range p.limbs {
		if !f(limb.Get(index)) {
			return false
		}
	}
	//
	return true
}

// Clone each array in this vector, producing a vector of mutable arrays.
func (p *Vector[T]) Clone() MutVector[T] {
	var limbs = make([]MutArray[T], len(p.limbs))
	//
	for i, ith := range p.limbs {
		limbs[i] = ith.Clone()
	}
	//
	return MutVector[T]{p.length, limbs}
}

// EmptyClone construct a vector of mutable arrays which are all empty, but have
// matching dimensions (i.e. height and bitwidth).
func (p *Vector[T]) EmptyClone(builder Builder[T]) MutVector[T] {
	var limbs = make([]MutArray[T], len(p.limbs))
	//
	for i, ith := range p.limbs {
		limbs[i] = builder.NewArray(p.length, ith.BitWidth())
	}
	//
	return MutVector[T]{p.length, limbs}
}

// Limb returns the ith limb in this vector.
func (p *Vector[T]) Limb(index uint) Array[T] {
	return p.limbs[index]
}

// Read the value of each limb at the given index into the given array.
func (p *Vector[T]) Read(index uint, values []T) {
	// TODO: should be able to relax following.
	if p.Width() != uint(len(values)) {
		panic("incompatible value array")
	}
	//
	for i, limb := range p.limbs {
		values[i] = limb.Get(index)
	}
}

// Len returns the length of each array in this vector.
func (p *Vector[T]) Len() uint {
	return p.length
}

// Some checks whether there exists a limb for which the given predicate holds
// at a given index.
func (p *Vector[T]) Some(index uint, f Predicate[T]) bool {
	for _, limb := range p.limbs {
		if f(limb.Get(index)) {
			return true
		}
	}
	//
	return false
}

// Width returns the number of arrays in this vector.
func (p *Vector[T]) Width() uint {
	return uint(len(p.limbs))
}

// ============================================================================
// Mutable VecArray
// ============================================================================

// MutVector represents a vector of mutable arrays.  As such, this supports not
// only reading operations but also writing operations.
type MutVector[T any] struct {
	length uint
	limbs  []MutArray[T]
}

// NewMutVector creates a new mutable vector with a given number of limbs as
// determined by the given bitwidths.
func NewMutVector[T any](height uint, bitwidths []uint, builder Builder[T]) MutVector[T] {
	var limbs = make([]MutArray[T], len(bitwidths))
	//
	for i, bitwidth := range bitwidths {
		limbs[i] = builder.NewArray(height, bitwidth)
	}
	//
	return MutVector[T]{height, limbs}
}

// MutVectorOf constructs a mutable vector from a given set of mutable arrays.
// All supplied arrays are expected to have the same height, otherwise this will
// panic.
func MutVectorOf[T any](arrays ...MutArray[T]) MutVector[T] {
	var length uint
	//
	for i, arr := range arrays {
		if i == 0 {
			length = arr.Len()
		} else if length != arr.Len() {
			panic(fmt.Sprintf("incompatible array lengths (%d vs %d)", length, arr.Len()))
		}
	}
	//
	return MutVector[T]{length, arrays}
}

// Len returns the length of each array in this vector.
func (p *MutVector[T]) Len() uint {
	return p.length
}

// Read the value of each limb at the given index into the given array.
func (p *MutVector[T]) Read(index uint, values []T) {
	// TODO: should be able to relax following.
	if p.Width() != uint(len(values)) {
		panic("incompatible value array")
	}
	//
	for i, limb := range p.limbs {
		values[i] = limb.Get(index)
	}
}

// Unwrap provides access to the underlying arrays.
func (p *MutVector[T]) Unwrap() []MutArray[T] {
	return p.limbs
}

// Width returns the number of arrays in this vector.
func (p *MutVector[T]) Width() uint {
	return uint(len(p.limbs))
}

// Write the value of each element in the given array into the corresponding
// limb at the given index.
func (p *MutVector[T]) Write(index uint, values []T) {
	// TODO: should be able to relax following.
	if p.Width() != uint(len(values)) {
		panic("incompatible value array")
	}
	//
	for i, limb := range p.limbs {
		limb.Set(index, values[i])
	}
}

// ============================================================================
// Helpers
// ============================================================================

// WidthOfVectors returns the total number of limbs across all vectors.
func WidthOfVectors[T any](vectors ...Vector[T]) uint {
	var totalWidth uint
	//
	for _, v := range vectors {
		totalWidth += v.Width()
	}
	//
	return totalWidth
}

// BitwidthOfVectors returns the maximum bitwidth of each limb across a given
// set of vectors.
func BitwidthOfVectors[T any](vectors ...Vector[T]) []uint {
	var (
		bitwidths = make([]uint, MaxWidthOfVectors(vectors...))
	)
	//
	for _, v := range vectors {
		for i := range v.Width() {
			bitwidths[i] = max(bitwidths[i], v.Limb(i).BitWidth())
		}
	}
	//
	return bitwidths
}

// MaxWidthOfVectors returns the maximum width of any vector in a given set of
// zero or more vectors.
func MaxWidthOfVectors[T any](vectors ...Vector[T]) uint {
	var maxWidth uint
	//
	for _, vec := range vectors {
		maxWidth = max(maxWidth, vec.Width())
	}
	//
	return maxWidth
}
