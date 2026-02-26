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
package stmt

import (
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/variable"
)

// Fail signals an exceptional return from the enclosing function.
type Fail[E any] struct {
	// dummy is included to force Return structs to be stored in the heap.
	//nolint
	Dummy uint
}

// Buses implementation for Instruction interface
func (p *Fail[E]) Buses() []E {
	return nil
}

// Uses implementation for Instruction interface.
func (p *Fail[E]) Uses() []variable.Id {
	return nil
}

// Definitions implementation for Instruction interface.
func (p *Fail[E]) Definitions() []variable.Id {
	return nil
}

func (p *Fail[E]) String(_ variable.Map) string {
	return "fail"
}

// Validate implementation for Instruction interface.
func (p *Fail[E]) Validate(env variable.Map) error {
	return nil
}
