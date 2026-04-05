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
package instruction

import (
	"github.com/consensys/go-corset/pkg/schema/register"
	"github.com/consensys/go-corset/pkg/util/field"
)

// Return performs an unconditional branch to a given target instructon.
type Return[W any] struct {
}

// Uses implementation for Instruction interface.
func (p *Return[W]) Uses() []register.Id {
	return nil
}

// Definitions implementation for Instruction interface.
func (p *Return[W]) Definitions() []register.Id {
	return nil
}

func (p *Return[W]) String(_ SystemMap[W]) string {
	return "ret"
}

// Validate implementation for Instruction interface.
func (p *Return[W]) Validate(_ field.Config, _ SystemMap[W]) []error {
	return nil
}

// MicroValidate implementation for Instruction interface.
func (p *Return[W]) MicroValidate(_ uint, _ field.Config, _ SystemMap[W]) []error {
	return nil
}
