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
package agnostic

import (
	"fmt"

	sc "github.com/consensys/go-corset/pkg/schema"
)

// NumberOfLimbs determines the number of register limbs required for a given
// bitwidth. For example, a 64bit register splits into two limbs for a maximum
// register width of 32bits. Observe that an e.g. 60bit register also splits
// into two limbs here as well, where the most significant limb is 28bits wide
// and the least significant is 32bits width.
func NumberOfLimbs(maxRegisterWidth uint, registerWidth uint) uint {
	n := registerWidth / maxRegisterWidth
	m := registerWidth % maxRegisterWidth
	// Check for uneven split
	if m != 0 {
		return n + 1
	}
	//
	return n
}

// CommonLimbWidth returns the "common" limb width when splitting a given
// register for a given maximum width.  Assume a given register splits into n
// limbs.  Assuming n > 1, then n-1 of these will have the same "common" width.
// The remaining limb is referred to as the residue, and may have a different
// width (usually smaller, but this is not a requirement).
func CommonLimbWidth(maxRegisterWidth uint, registerWidth uint) uint {
	var (
		// Determine how many limbs required
		n = NumberOfLimbs(maxRegisterWidth, registerWidth)
		// Determine average limb width
		width = registerWidth / n
		//
		acc = uint(1)
	)
	// Now, round up our average limb width to the nearest power of two.  This
	// is because we "prefer" widths to be powers of two.
	for ; acc < width; acc = acc * 2 {
	}
	//
	return acc
}

// WidthsOfLimbs returns the limb bitwidths corresponding to a given set of
// identifiers.
func WidthsOfLimbs(mapping sc.RegisterLimbsMap, lids []sc.LimbId) []uint {
	var (
		widths []uint = make([]uint, len(lids))
	)
	//
	for i, lid := range lids {
		widths[i] = mapping.Limb(lid).Width
	}
	//
	return widths
}

// CombinedWidthOfLimbs returns the combined bitwidth of all limbs.  For example,
// suppose we have three limbs: x:u8, y:u8, z:u11.  Then the combined width is
// 8+8+11=27.
func CombinedWidthOfLimbs(mapping sc.RegisterLimbsMap, limbs ...sc.LimbId) uint {
	var (
		width uint
	)
	//
	for _, lid := range limbs {
		width += mapping.Limb(lid).Width
	}
	//
	return width
}

// SplitIntoLimbs splits a register into a number of limbs with the given maximum
// bitwidth.  For the resulting array, the least significant register is first.
// Since registers are always split to the maximum width as much as possible, it
// is only the most significant register which may (in some cases) have fewer
// bits than the maximum allowed.
func SplitIntoLimbs(maxWidth uint, r sc.Register) []sc.Register {
	var (
		nlimbs = NumberOfLimbs(maxWidth, r.Width)
		limbs  = make([]sc.Register, nlimbs)
		width  = r.Width
		// Split padding value
		padding = SplitConstant(r.Padding, LimbWidths(maxWidth, r.Width)...)
	)
	// Special case when register doesn't require splitting.  This is useful
	// because we want to retain the original register name exactly.
	if nlimbs == 1 {
		return []sc.Register{r}
	}
	//
	maxWidth = CommonLimbWidth(maxWidth, width)
	//
	for i := uint(0); i < nlimbs; i++ {
		ith_name := fmt.Sprintf("%s'%d", r.Name, i)
		ith_width := min(maxWidth, width)
		limbs[i] = sc.Register{
			Name:    ith_name,
			Kind:    r.Kind,
			Width:   ith_width,
			Padding: padding[i],
		}
		//
		width -= maxWidth
	}
	//
	return limbs
}

// LimbWidths determines the limb widths for any register of the given size.
func LimbWidths(maxWidth, regWidth uint) []uint {
	var (
		nlimbs     = NumberOfLimbs(maxWidth, regWidth)
		limbWidths = make([]uint, nlimbs)
	)
	//
	maxWidth = CommonLimbWidth(maxWidth, regWidth)
	//
	for i := uint(0); i < nlimbs; i++ {
		limbWidths[i] = min(maxWidth, regWidth)
		regWidth -= maxWidth
	}
	//
	return limbWidths
}
