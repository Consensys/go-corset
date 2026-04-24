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
	"bufio"

	"github.com/consensys/go-corset/pkg/util/termio"
	"github.com/consensys/go-corset/pkg/zkc/vm/word"
)

type TracePrinter[W word.Word[W]] struct {
	out bufio.Writer
	// Formatting to use for the Program Counter
	pcFormat termio.AnsiEscape
	// Formatting to use for the body of the instruction
	insnFormat termio.AnsiEscape
	// Formatting to use for the debugging values
	valueFormat termio.AnsiEscape
}

func NewInstructionPrinter[W word.Word[W]]() TracePrinter[W] {
	return TracePrinter[W]{
		pcFormat:    termio.NewAnsiEscape().FgColour(termio.TERM_YELLOW),
		insnFormat:  termio.NewAnsiEscape().FgColour(termio.TERM_WHITE),
		valueFormat: termio.NewAnsiEscape().Fg256Colour(250),
	}
}

// PrintAll prints one (or more) execution steps.
func (p *TracePrinter[W]) PrintAll(steps []ExecutionStep[W]) {

}

// Print exactly one execution step
func (p *TracePrinter[W]) Print(step ExecutionStep[W]) {

}
