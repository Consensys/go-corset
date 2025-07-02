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
package io

import (
	"github.com/consensys/go-corset/pkg/schema"
)

// SplittingEnvironment is used to assist with register splitting.
type SplittingEnvironment struct {
	// Mapping provides the mapping of registers before splitting to their limbs
	// after splitting.
	mapping schema.RegisterMapping
	// BandWidth represents the maximum bandwidth available in the underlying
	// field.
	bandwidth uint
}

// NewSplittingEnvironment constructs a new splitting environment.
func NewSplittingEnvironment(mapping schema.RegisterMapping, bandwidth uint) SplittingEnvironment {
	return SplittingEnvironment{mapping, bandwidth}
}

// AllocateCarryRegister allocates a carry flag to hold bits which "overflow" the
// left-hand side of an assignment (i.e. where sourceWidth is greater than
// targetWidth).
func (p SplittingEnvironment) AllocateCarryRegister(targetWidth uint, sourceWidth uint) RegisterId {
	// Sanity check new register is not larger than maximum register width!
	panic("todo")
}

// BandWidth returns the maximum bandwidth available in the underlying
// field.  This cannot be smaller than the maximum register width.
func (p SplittingEnvironment) BandWidth() uint {
	return p.bandwidth
}

// LimbIds implementation for the schema.RegisterMapping interface.
func (p SplittingEnvironment) LimbIds(reg schema.RegisterId) []schema.LimbId {
	return p.mapping.LimbIds(reg)
}

// Limb implementation for the schema.RegisterMapping interface.
func (p SplittingEnvironment) Limb(reg schema.LimbId) schema.Limb {
	return p.mapping.Limb(reg)
}

// Limbs implementation for the schema.RegisterMapping interface.
func (p SplittingEnvironment) Limbs() []schema.Limb {
	return p.mapping.Limbs()
}
