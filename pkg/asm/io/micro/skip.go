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
package micro

import (
	"fmt"
	"math/big"

	"github.com/consensys/go-corset/pkg/asm/io"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/agnostic"
)

// Skip microcode performs a conditional skip over a given number of codes. The
// condition is either that two registers are equal, or that they are not equal.
// This has two variants: register-register; and, register-constant.  The latter
// is indiciated when the right register is marked as UNUSED.
type Skip struct {
	// Left and right comparisons
	Left, Right io.RegisterId
	//
	Constant big.Int
	// Skip
	Skip uint
}

// Clone this micro code.
func (p *Skip) Clone() Code {
	var constant big.Int
	//
	constant.Set(&p.Constant)
	//
	return &Skip{
		Left:     p.Left,
		Right:    p.Right,
		Constant: constant,
		Skip:     p.Skip,
	}
}

// MicroExecute a given micro-code, using a given local state.  This may update
// the register values, and returns either the number of micro-codes to "skip
// over" when executing the enclosing instruction or, if skip==0, a destination
// program counter (which can signal return of enclosing function).
func (p *Skip) MicroExecute(state io.State) (uint, uint) {
	var (
		lhs = state.Load(p.Left)
		rhs *big.Int
	)
	//
	if p.Right.IsUsed() {
		rhs = state.Load(p.Right)
	} else {
		rhs = &p.Constant
	}
	//
	if lhs.Cmp(rhs) != 0 {
		return 1 + p.Skip, 0
	} else {
		return 1, 0
	}
}

// RegistersRead returns the set of registers read by this instruction.
func (p *Skip) RegistersRead() []io.RegisterId {
	if p.Right.IsUsed() {
		return []io.RegisterId{p.Left, p.Right}
	}
	//
	return []io.RegisterId{p.Left}
}

// RegistersWritten returns the set of registers written by this instruction.
func (p *Skip) RegistersWritten() []io.RegisterId {
	return nil
}

// Split this micro code using registers of arbirary width into one or more
// micro codes using registers of a fixed maximum width.
func (p *Skip) Split(mapping schema.RegisterLimbsMap, _ schema.RegisterAllocator) []Code {
	// NOTE: we can assume left and right have matching bitwidths
	var (
		lhsLimbs = mapping.LimbIds(p.Left)
		ncodes   []Code
		n        = uint(len(lhsLimbs))
		skip     = p.Skip + n - 1
	)
	//
	if p.Right.IsUsed() {
		rhsLimbs := mapping.LimbIds(p.Right)
		//
		for i := range n {
			ncode := &Skip{lhsLimbs[i], rhsLimbs[i], p.Constant, skip - i}
			ncodes = append(ncodes, ncode)
		}
	} else {
		lhsLimbWidths := agnostic.WidthsOfLimbs(mapping, lhsLimbs)
		constantLimbs := agnostic.SplitConstant(p.Constant, lhsLimbWidths...)
		//
		for i := range n {
			ncode := &Skip{lhsLimbs[i], schema.NewUnusedRegisterId(), constantLimbs[i], skip - i}
			ncodes = append(ncodes, ncode)
		}
	}

	return ncodes
}

func (p *Skip) String(fn schema.RegisterMap) string {
	var (
		l = fn.Register(p.Left).Name
	)
	//
	if p.Right.IsUsed() {
		return fmt.Sprintf("skip %s!=%s %d", l, fn.Register(p.Right).Name, p.Skip)
	}
	//
	return fmt.Sprintf("skip %s!=%s %d", l, p.Constant.String(), p.Skip)
}

// Validate checks whether or not this instruction is correctly balanced.
func (p *Skip) Validate(fieldWidth uint, fn schema.RegisterMap) error {
	var lw = fn.Register(p.Left).Width
	//
	if p.Right.IsUsed() {
		rw := fn.Register(p.Right).Width
		//
		if lw != rw {
			return fmt.Errorf("bit mismatch (u%d vs u%d)", lw, rw)
		}
	} else {
		cw := uint(p.Constant.BitLen())
		//
		if lw < cw {
			return fmt.Errorf("bit overflow (u%d into u%d)", lw, cw)
		}
	}
	//
	return nil
}
