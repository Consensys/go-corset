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

// OpCode identifies an instruction exactly
type OpCode uint8

const (
	// ========================================================================
	// A-type instructions
	// ========================================================================

	// INT_ADD instruction
	INT_ADD OpCode = iota
	// INT_SUB instruction
	INT_SUB
	// INT_MUL instruction
	INT_MUL
	// INT_DIV instruction
	INT_DIV
	// INT_REM instruction
	INT_REM
	// INT_CAST instruction
	INT_CAST
	// FIELD_ADD instruction
	FIELD_ADD
	// FIELD_SUB instruction
	FIELD_SUB
	// FIELD_MUL instruction
	FIELD_MUL
	// FIELD_CAST instruction
	FIELD_CAST
	// BIT_AND instruction
	BIT_AND
	// BIT_OR instruction
	BIT_OR
	// BIT_XOR instruction
	BIT_XOR
	// BIT_NOT instruction
	BIT_NOT
	// BIT_SHL (shift left) instruction
	BIT_SHL
	// BIT_SHR (shift right) instruction
	BIT_SHR
	// BIT_CONCAT (concatenation) instruction
	BIT_CONCAT
	// BIT_DESTRUCT (destructuring) instruction
	BIT_DESTRUCT
	// SKIP_IF instruction
	SKIP_IF
	// SKIP instruction
	SKIP
	// JUMP (unconditional) instruction
	JUMP
	// FAIL instruction
	FAIL
	// CALL instruction
	CALL
	// MEMORY_READ instruction
	MEMORY_READ
	// MEMORY_WRITE instruction
	MEMORY_WRITE
	// RETURN instruction
	RETURN
	// VECTOR instruction
	VECTOR
	// DEBUG instruction
	DEBUG
)
