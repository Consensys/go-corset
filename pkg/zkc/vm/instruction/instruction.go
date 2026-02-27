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
)

// Instruction provides an abstract notion of a "machine instruction".  That is, a single atomic unit which can be
type Instruction[W any] interface {
	// Uses returns the set of variables used (i.e. read) by this instruction.
	Uses() []register.Id
	// Definitions returns the set of variables registers defined (i.e. written)
	// by this instruction.
	Definitions() []register.Id
	// Validate that this instruction is well-formed.  For example, that it is
	// balanced, that there are no conflicting writes, that all temporaries have
	// been allocated, etc.
	Validate(env register.Map) error
	// Provide human readable form of instruction
	String(env register.Map) string
}
