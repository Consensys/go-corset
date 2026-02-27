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
	"fmt"

	"github.com/consensys/go-corset/pkg/schema/register"
)

// Goto performs an unconditional branch to a given target instructon.
type Goto struct {
	Target uint
}

// Uses implementation for Instruction interface.
func (p *Goto) Uses() []register.Id {
	return nil
}

// Definitions implementation for Instruction interface.
func (p *Goto) Definitions() []register.Id {
	return nil
}

func (p *Goto) String(_ register.Map) string {
	return fmt.Sprintf("goto %d", p.Target)
}

// Validate implementation for Instruction interface.
func (p *Goto) Validate(env register.Map) error {
	return nil
}
