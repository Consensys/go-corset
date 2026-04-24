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
package debug

import (
	"github.com/consensys/go-corset/pkg/zkc/vm/machine"
	"github.com/consensys/go-corset/pkg/zkc/vm/word"
)

// ExecutionTrace captures the steps arising when executing a given program.
type ExecutionTrace[W word.Word[W]] struct {
	Steps []ExecutionStep[W]
}

// ExecutionStep captures the minimal amount of information
type ExecutionStep[W word.Word[W]] struct {
	Kind ExecutionKind
	// Id of function executing this step.
	Fun uint
	// pc position within function
	Pc machine.ProgramCounter
	// targets / source values
	Values []W
}

// ExecutionKind provides a means of classifying different execution steps.
type ExecutionKind uint

const (
	// EXEC captures a normal instruction execution.  That is, the execution of
	// an instruction which simply proceeds to the next.
	EXEC ExecutionKind = iota
	// ENTER captures the start of a function call.  In such cases, the values
	// of the step correspond to the call arguments.
	ENTER
	// RETURN captures the end of a function call.  In such case, the values of
	// the step correspond to the call returns.
	RETURN
)
