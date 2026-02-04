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
	"strings"

	"github.com/consensys/go-corset/pkg/schema/register"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/bit"
)

// WriteState identifiers, for a set of registers, whether each may have been
// written (i.e. maybe assigned) and, furthermore, whether it has definitely
// been written (i.e. definitely assigned).  Observe that definitely assigned
// registers are also always maybe assigned registers.
type WriteState struct {
	maxRegister    uint
	definiteWrites bit.Set
	maybeWrites    bit.Set
}

// Clone this write state producing an otherwise identical but physically
// disjoint state.
func (p *WriteState) Clone() WriteState {
	var nst WriteState
	//
	nst.maxRegister = p.maxRegister
	nst.definiteWrites.Union(p.definiteWrites)
	nst.maybeWrites.Union(p.maybeWrites)
	//
	return nst
}

// Write constructs a write state representing the give state after a set of
// writes have occurred.
func (p WriteState) Write(regs ...register.Id) WriteState {
	var nst = p.Clone()
	//
	for _, r := range regs {
		nst.maxRegister = max(nst.maxRegister, r.Unwrap())
		nst.definiteWrites.Insert(r.Unwrap())
		nst.maybeWrites.Insert(r.Unwrap())
	}
	//
	return nst
}

// MaybeAssigned determines whether or not a give register may have been
// assigned.
func (p WriteState) MaybeAssigned(reg register.Id) bool {
	return p.maybeWrites.Contains(reg.Unwrap())
}

// DefinitelyAssigned determines whether or not a give register has definitely
// been assigned.  Observe that, for any register r, it follows that
// DefinitelyAssigned(r) implies MaybeAssigned(r).
func (p WriteState) DefinitelyAssigned(reg register.Id) bool {
	return p.definiteWrites.Contains(reg.Unwrap())
}

func (p *WriteState) String(rmap register.Map) string {
	var (
		builder strings.Builder
		first   = true
		nRegs   = uint(len(rmap.Registers()))
	)
	//
	builder.WriteString("{")
	//
	for i := uint(0); i < nRegs; i++ {
		if !p.maybeWrites.Contains(i) {
			continue
		}
		//
		if !first {
			builder.WriteString(",")
		} else {
			first = false
		}
		//
		if !p.definiteWrites.Contains(i) {
			builder.WriteString("?")
		}
		//
		var (
			rid  = register.NewId(i)
			name = rmap.Register(rid).Name()
		)
		//
		builder.WriteString(name)
	}
	//
	builder.WriteString("}")
	//
	return builder.String()
}

// JoinWriteStates combines two write states together.
func JoinWriteStates(l, r WriteState) WriteState {
	var nst WriteState
	//
	nst.maxRegister = max(l.maxRegister, r.maxRegister)
	//
	for i := range nst.maxRegister + 1 {
		rid := register.NewId(i)
		//
		if l.DefinitelyAssigned(rid) && r.DefinitelyAssigned(rid) {
			nst.definiteWrites.Insert(i)
			nst.maybeWrites.Insert(i)
		} else if l.MaybeAssigned(rid) || r.MaybeAssigned(rid) {
			nst.maybeWrites.Insert(i)
		}
	}
	//
	return nst
}

// WriteMap is used to provide information about writes occurring within the
// micro instruction.  This is necessary, for example, to determine when
// forwarding should be applied.  Specifically, on entry to each micro-code,
// this identifies which variables may have been written at the given point.
// Furthermore, it allows us to distinguish between those which may have been
// written and those which have definitely been written.
type WriteMap struct {
	// For each micro-code, identifes the write state on entry to that micro-code.
	states []util.Option[WriteState]
}

// NewWriteMap constructs an empty write map for a given number of micro-codes.
func NewWriteMap(nCodes uint) WriteMap {
	var states = make([]util.Option[WriteState], nCodes)
	//
	if nCodes > 0 {
		states[0] = util.Some(WriteState{})
	}
	//
	return WriteMap{states: states}
}

// StateOf returns the current write state on entry to the given micro-code.
func (p *WriteMap) StateOf(i uint) WriteState {
	return p.states[i].Unwrap()
}

// JoinInto updates the write state for a given micro-code to include that from
// another branch.
func (p *WriteMap) JoinInto(i uint, st WriteState) {
	var (
		ith = p.states[i]
		nst WriteState
	)
	// Check for bottom
	if ith.HasValue() {
		nst = JoinWriteStates(st, ith.Unwrap())
	} else {
		nst = st.Clone()
	}
	//
	p.states[i] = util.Some(nst)
}

func (p *WriteMap) String(rmap register.Map) string {
	var builder strings.Builder
	//
	for _, st := range p.states {
		if st.HasValue() {
			val := st.Unwrap()
			builder.WriteString(val.String(rmap))
		} else {
			builder.WriteString("⊥")
		}
	}
	//
	return builder.String()
}
