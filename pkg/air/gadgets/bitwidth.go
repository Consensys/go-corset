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
package gadgets

import (
	"fmt"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/air"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/assignment"
	"github.com/consensys/go-corset/pkg/schema/constraint"
	"github.com/consensys/go-corset/pkg/trace"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/iter"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// BitwidthGadget is a general-purpose mechanism for enforcing type constraints
// (i.e. that a given register has a given bitwidth).  Depending on the width
// and config used, this generates appropriate constraints and/or modules as
// necessary for enforcing bitwidth constraints.
type BitwidthGadget struct {
	// Determines the largest bitwidth for which range constraints are
	// translated into AIR range constraints, versus  using a horizontal
	// bitwidth gadget.
	maxRangeConstraint uint
	// Disables the use of type proofs which exploit the limitless prover.
	// Specifically, modules with a recursive structure are created specifically
	// for the purpose of checking types.
	legacy bool
	// Schema into which constraints are placed.
	schema *air.Schema
}

// NewBitwidthGadget constructs a new bitwidth gadget.
func NewBitwidthGadget(schema *air.Schema) *BitwidthGadget {
	return &BitwidthGadget{
		maxRangeConstraint: 8,
		legacy:             false,
		schema:             schema,
	}
}

// WithMaxRangeConstraint determines the cutoff for range cosntraints.
func (p *BitwidthGadget) WithMaxRangeConstraint(width uint) *BitwidthGadget {
	p.maxRangeConstraint = width
	return p
}

// WithLegacyTypeProofs disables (or enables) use of limitless type proofs.
func (p *BitwidthGadget) WithLegacyTypeProofs(flag bool) *BitwidthGadget {
	p.legacy = flag
	return p
}

// Constrain ensures all values in a given register fit within a given bitwidth.
func (p *BitwidthGadget) Constrain(col uint, bitwidth uint) {
	// Base cases
	switch {
	case bitwidth <= 1:
		p.applyBinaryGadget(col)
		return
	case bitwidth <= p.maxRangeConstraint:
		// Add range constraint
		p.schema.AddRangeConstraint(col, 0, bitwidth)
		// Done
		return
	case p.legacy:
		// NOTE: this should be deprecated once the limitless prover is well
		// established.
		p.applyHorizontalBitwidthGadget(col, bitwidth)
	default:
		p.applyRecursiveBitwidthGadget(col, bitwidth)
	}
}

// ApplyBinaryGadget adds a binarity constraint for a given column in the schema
// which enforces that all values in the given column are either 0 or 1. For a
// column X, this corresponds to the vanishing constraint X * (X-1) == 0.
func (p *BitwidthGadget) applyBinaryGadget(col uint) {
	var (
		// Identify target column
		column = p.schema.Columns().Nth(col)
		// Construct column handle
		handle = fmt.Sprintf("%s:u1", column.Name)
	)
	// Construct X
	X := air.NewColumnAccess(col, 0)
	// Construct X-1
	X_m1 := X.Sub(air.NewConst64(1))
	// Construct X * (X-1)
	X_X_m1 := X.Mul(X_m1)
	// Done!
	p.schema.AddVanishingConstraint(handle, 0, column.Context, util.None[int](), X_X_m1)
}

// ApplyHorizontalBitwidthGadget ensures all values in a given column fit within
// a given bitwidth.  This is implemented using a *horizontal byte
// decomposition* which adds n columns and a vanishing constraint (where n*8 >=
// bitwidth).
func (p *BitwidthGadget) applyHorizontalBitwidthGadget(col uint, bitwidth uint) {
	var (
		// Determine ranges required for the give bitwidth
		ranges, widths = splitColumnRanges(bitwidth)
		// Identify number of columns required.
		n = uint(len(ranges))
	)
	// Sanity check
	if bitwidth == 0 {
		panic("zero bitwidth constraint encountered")
	}
	// Identify target column
	column := p.schema.Columns().Nth(col)
	// Calculate how many bytes required.
	es := make([]air.Expr, n)
	name := column.Name
	coefficient := fr.NewElement(1)
	// Add decomposition assignment
	index := p.schema.AddAssignment(
		assignment.NewByteDecomposition(name, column.Context, col, bitwidth))
	// Construct Columns
	for i := uint(0); i < n; i++ {
		// Create Column + Constraint
		es[i] = air.NewColumnAccess(index+i, 0).Mul(air.NewConst(coefficient))

		p.schema.AddRangeConstraint(index+i, 0, widths[i])
		// Update coefficient
		coefficient.Mul(&coefficient, &ranges[i])
	}
	// Construct (X:0 * 1) + ... + (X:n * 2^n)
	sum := air.Sum(es...)
	// Construct X == (X:0 * 1) + ... + (X:n * 2^n)
	X := air.NewColumnAccess(col, 0)
	eq := X.Equate(sum)
	// Construct column name
	p.schema.AddVanishingConstraint(
		fmt.Sprintf("%s:u%d", name, bitwidth), 0, column.Context, util.None[int](), eq)
}

// ApplyRecursiveBitwidthGadget ensures all values in a given column fit within
// a given bitwidth. This is implemented using a combination of reference tables
// and lookups.  Specifically, if the width is below 16bits, then a static
// reference table is created along with a corresponding lookup,  Otherwise, a
// recursive procedure is applied whereby a table is created for the given width
// which divides each value into two smaller values.  This procedure is then
// recursively applied to those columns, etc.
func (p *BitwidthGadget) applyRecursiveBitwidthGadget(col uint, bitwidth uint) {
	var (
		column       = p.schema.Columns().Nth(col)
		proofHandle  = fmt.Sprintf(":u%d", bitwidth)
		lookupHandle = fmt.Sprintf("%s:u%d", column.Name, bitwidth)
	)
	// Recursive case.
	mid, ok := p.schema.Modules().Find(func(m sc.Module) bool {
		return m.Name == proofHandle
	})
	//
	if !ok {
		// Construct proof module where first column is the target, etc.
		mid = p.constructTypeProof(proofHandle, bitwidth)
	}
	// Identify the target column of proof module
	vid := p.findTargetColumn(mid)
	// Configure source vector
	source := constraint.UnfilteredLookupVector(column.Context, &air.ColumnAccess{Column: col, Shift: 0})
	// Configure target vector
	target := constraint.UnfilteredLookupVector(tr.NewContext(mid, 1), &air.ColumnAccess{Column: vid, Shift: 0})
	// Add lookup constraint
	p.schema.AddLookupConstraint(lookupHandle, source, target)
	// Add column to assignment so its proof is included
	p.schema.Assignments().Find(func(a sc.Assignment) bool {
		// Check whether the matching type decomposition
		if proof, ok := a.(*typeDecomposition); ok && proof.Bitwidth() == bitwidth {
			// match
			proof.AddSource(col)
			// terminate search
			return true
		}
		//
		return false
	})
}

func (p *BitwidthGadget) constructTypeProof(handle string, bitwidth uint) uint {
	var (
		// Create new module for this type proof
		mid = p.schema.AddModule(handle)
		//
		ctx = tr.NewContext(mid, 1)
		// Determine limb widths.
		loWidth, hiWidth = determineLimbSplit(bitwidth)
	)
	// Add (initially empty) assignment
	vid := p.schema.AddAssignment(newTypeDecomposition(ctx, loWidth, hiWidth))
	vidLo := vid + 1
	vidHi := vid + 2
	// Compute 2^loWidth to use as coefficient
	coeff := fr.NewElement(2)
	util.Pow(&coeff, uint64(loWidth))
	// Ensure lo/hi are decomposition of original
	p.schema.AddVanishingConstraint("decomposition", 0, ctx, util.None[int](),
		air.Subtract(
			air.NewColumnAccess(vid, 0),
			air.Sum(
				air.NewColumnAccess(vidLo, 0),
				air.Product(air.NewConst(coeff), air.NewColumnAccess(vidHi, 0)),
			),
		))
	// Recursively prove lo/hi columns
	p.Constrain(vidLo, loWidth)
	p.Constrain(vidHi, hiWidth)
	//dev Done
	return mid
}

// Determine the index of the target column for the given type proof.  That is,
// the column which holds the values being range checked.
func (p *BitwidthGadget) findTargetColumn(mid uint) uint {
	// Determining the first column index of an assignment is pretty easy.  We
	// just look for the first occurring column whose context matches the target
	// module.
	cid, ok := p.schema.Columns().Find(func(m sc.Column) bool {
		return m.Context.ModuleId == mid
	})
	// Sanity check
	if !ok {
		mod := p.schema.Modules().Nth(mid)
		panic(fmt.Sprintf("missing target column for type proof %s", mod.Name))
	}
	// Done!
	return cid
}

// ============================================================================
// Type Decomposition Assignment
// ============================================================================

type typeDecomposition struct {
	// Limb widths of decomposition.
	loWidth, hiWidth uint
	// Source registers being decomposed which represent the set of all the
	// columns in the entire system being proved to have the given bitwidth.
	sources []uint
	// Target registers for decomposition.  There are always exactly three
	// registers, where the first holds the original value, followed by the
	// least significant (lo) limb and, finally, the most significant (hi) limb.
	targets []sc.Column
}

func newTypeDecomposition(context trace.Context, loWidth, hiWidth uint) *typeDecomposition {
	return &typeDecomposition{
		loWidth, hiWidth,
		nil, // initially empty set of source columns
		[]sc.Column{
			sc.NewColumn(context, "V", sc.NewUintType(loWidth+hiWidth)),
			sc.NewColumn(context, "V'0", sc.NewUintType(loWidth)),
			sc.NewColumn(context, "V'1", sc.NewUintType(hiWidth)),
		},
	}
}

// AddSource adds a new source column to this decomposition.
func (p *typeDecomposition) AddSource(source uint) {
	p.sources = append(p.sources, source)
}

// Bitwidth returns the bitwidth being enforced by this type decompsition.
func (p *typeDecomposition) Bitwidth() uint {
	return p.loWidth + p.hiWidth
}

// Context returns the evaluation context for this declaration.
func (p *typeDecomposition) Context() trace.Context {
	return p.targets[0].Context
}

// Columns returns the columns declared by this byte decomposition (in the order
// of declaration).
func (p *typeDecomposition) Columns() iter.Iterator[sc.Column] {
	return iter.NewArrayIterator(p.targets)
}

// IsComputed Determines whether or not this declaration is computed.
func (p *typeDecomposition) IsComputed() bool {
	return true
}

// Compute computes the values of columns defined by this assignment.
// This requires computing the value of each byte column in the decomposition.
func (p *typeDecomposition) ComputeColumns(tr trace.Trace) ([]trace.ArrayColumn, error) {
	// Read all input columns
	sources := readSources(tr, p.sources...)
	// Combine all sources to eliminate duplicates
	combined := combineSources(p.loWidth+p.hiWidth, sources)
	// Generate decomposition
	data := computeDecomposition(p.loWidth, p.hiWidth, combined)
	//
	return writeTargets(p.targets, data), nil
}

// Bounds determines the well-definedness bounds for this assignment for both
// the negative (left) or positive (right) directions.  For example, consider an
// expression such as "(shift X -1)".  This is technically undefined for the
// first row of any trace and, by association, any constraint evaluating this
// expression on that first row is also undefined (and hence must pass).
func (p *typeDecomposition) Bounds() util.Bounds {
	return util.EMPTY_BOUND
}

// CheckConsistency performs some simple checks that the given schema is consistent.
// This provides a double check of certain key properties, such as that
// registers used for assignments are large enough, etc.
func (p *typeDecomposition) CheckConsistency(schema sc.Schema) error {
	return nil
}

// RegistersRead returns the set of columns that this assignment depends upon.
// That can include both input columns, as well as other computed columns.
func (p *typeDecomposition) Dependencies() []uint {
	return p.sources
}

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (p *typeDecomposition) Lisp(schema sc.Schema) sexp.SExp {
	var (
		targets = sexp.EmptyList()
		sources = sexp.EmptyList()
	)
	//
	for _, t := range p.targets {
		targets.Append(sexp.NewList([]sexp.SExp{
			// name
			sexp.NewSymbol(t.QualifiedName(schema)),
			// type
			sexp.NewSymbol(t.DataType.String()),
		}))
	}
	//
	for _, s := range p.sources {
		ithName := sc.QualifiedName(schema, s)
		sources.Append(sexp.NewSymbol(ithName))
	}
	//
	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("decompose"),
		targets,
		sources,
	})
}

// ============================================================================
// Helpers (for recursive)
// ============================================================================

// Determine the split of limbs for the given bitwidth.  For example, 33bits
// could be broken into 16bit and 17bit limbs, or into 8bit and 25bit limbs or
// into 32bit and 1bit limbs.  The current algorithm ensures the least
// significant limb is always a power of 2, and it tries to balance the limbs as
// much as possible (i.e. to reduce the tree depth).
func determineLimbSplit(bitwidth uint) (uint, uint) {
	var (
		pivot      = bitwidth / 2
		loMaxWidth = uint(1)
		loMinWidth = uint(1)
	)
	// Find nearest power of 2 (upper bound)
	for ; loMaxWidth < pivot; loMaxWidth = loMaxWidth * 2 {
		loMinWidth = loMaxWidth
	}
	// Decide which option gives better balance
	lowerDelta := pivot - loMinWidth
	upperDelta := loMaxWidth - pivot
	//
	if lowerDelta < upperDelta {
		return loMinWidth, bitwidth - loMinWidth
	}
	//
	return loMaxWidth, bitwidth - loMaxWidth
}

func readSources(trace tr.Trace, sources ...uint) []field.FrArray {
	var (
		targets = make([]field.FrArray, len(sources))
	)
	// Read registers
	for i, col := range sources {
		targets[i] = trace.Column(col).Data()
	}
	//
	return targets
}

func writeTargets(targets []sc.Column, data []field.FrArray) []tr.ArrayColumn {
	var (
		padding = fr.NewElement(0)
		columns = make([]tr.ArrayColumn, len(targets))
	)
	// Write outputs
	for i, target := range targets {
		columns[i] = tr.NewArrayColumn(target.Context, target.Name, data[i], padding)
	}
	//
	return columns
}

// Combine all values from the given source registers into a single array of
// data, whilst eliminating duplicates.
func combineSources(bitwidth uint, sources []field.FrArray) field.FrArray {
	var arr = field.NewFrIndexArray(0, bitwidth)
	// Always include zero to work around limitations of FrIndexArray.  This is
	// not actually inefficient, since all columns are subject to an initial
	// padding row anyway.
	arr.Append(fr.NewElement(0))
	//
	for _, src := range sources {
		for i := range src.Len() {
			ith := src.Get(i)
			// Add ith item if not already seen.
			if _, ok := arr.IndexOf(ith); !ok {
				arr.Append(src.Get(i))
			}
		}
	}
	// Done
	return arr
}

func computeDecomposition(loWidth, hiWidth uint, vArr field.FrArray) []field.FrArray {
	// FIXME: using an index array here ensures the underlying data is
	// represented using a full field element, rather than e.g. some smaller
	// number of bytes.  This is needed to handle reject tests which can produce
	// values outside the range of the computed register, but which we still
	// want to check are actually rejected (i.e. since they are simulating what
	// an attacker might do).
	var (
		vLoArr = field.NewFrIndexArray(vArr.Len(), loWidth)
		vHiArr = field.NewFrIndexArray(vArr.Len(), hiWidth)
	)
	//
	for i := range vArr.Len() {
		ith := vArr.Get(i)
		lo, hi := decompose(loWidth, ith)
		vLoArr.Set(i, lo)
		vHiArr.Set(i, hi)
	}
	//
	return []field.FrArray{vArr, vLoArr, vHiArr}
}

// Decompose a given field element into its least and most significant limbs,
// based on the required bitwidth for the least significant limb.
func decompose(loWidth uint, ith fr.Element) (fr.Element, fr.Element) {
	// Extract bytes from element
	var (
		bytes      = ith.Bytes()
		loFr, hiFr fr.Element
	)
	// Sanity check assumption
	if loWidth%8 != 0 {
		panic("unreachable")
	}
	//
	n := 32 - (loWidth / 8)
	hiFr.SetBytes(bytes[:n])
	loFr.SetBytes(bytes[n:])
	//
	return loFr, hiFr
}

// ============================================================================
// Helpers (for horizontal)
// ============================================================================

func splitColumnRanges(nbits uint) ([]fr.Element, []uint) {
	var (
		n      = nbits / 8
		m      = int64(nbits % 8)
		ranges []fr.Element
		widths []uint
		fr256  = fr.NewElement(256)
	)
	//
	if m == 0 {
		ranges = make([]fr.Element, n)
		widths = make([]uint, n)
	} else {
		var last fr.Element
		// Most significant column has smaller range.
		ranges = make([]fr.Element, n+1)
		widths = make([]uint, n+1)
		// Determine final range
		last.Exp(fr.NewElement(2), big.NewInt(m))
		//
		ranges[n] = last
		widths[n] = uint(m)
	}
	//
	for i := uint(0); i < n; i++ {
		ranges[i] = fr256
		widths[i] = 8
	}
	//
	return ranges, widths
}
