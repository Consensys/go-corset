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
package memory

import (
	"github.com/consensys/go-corset/pkg/schema/register"
	"github.com/consensys/go-corset/pkg/zkc/vm/word"
)

// RandomAccess represents  a form of memory where each cell can be read &
// written multiple times without restrictions.  The size of the memory expands
// dynamically to include any cell which is written, where cells are initialised
// with zero.
type RandomAccess[W word.Word[W]] struct {
	Array[W]
}

// NewRandomAccess constructs an empty random-access memory.
func NewRandomAccess[W word.Word[W]](name string, registers []register.Register) *RandomAccess[W] {
	return &RandomAccess[W]{
		newArray[W](name, registers),
	}
}
