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

// // SplittingEnvironment is used to assist with register splitting.
type SplittingEnvironment interface {
	schema.RegisterMapping
	// BandWidth returns the maximum bandwidth available in the underlying
	// field.  This cannot be smaller than the maximum register width.
	BandWidth() uint
	// MaxWidth returns the maximum permitted register width.
	MaxWidth() uint
	// AllocateCarryRegister allocates a carry flag to hold bits which "overflow" the
	// left-hand side of an assignment (i.e. where sourceWidth is greater than
	// targetWidth).
	AllocateCarryRegister(targetWidth uint, sourceWidth uint) RegisterId
}

// NewSplittingEnvironment constructs a new splitting environment.
func NewSplittingEnvironment(mapping schema.RegisterMapping, bandwidth uint) SplittingEnvironment {
	panic("todo")
}
