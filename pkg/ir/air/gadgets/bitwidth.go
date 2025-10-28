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

	"github.com/consensys/go-corset/pkg/ir/air"
	"github.com/consensys/go-corset/pkg/ir/assignment"
	"github.com/consensys/go-corset/pkg/ir/term"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/constraint/lookup"
	"github.com/consensys/go-corset/pkg/schema/module"
	"github.com/consensys/go-corset/pkg/schema/register"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/array"
	"github.com/consensys/go-corset/pkg/util/collection/hash"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
	"github.com/consensys/go-corset/pkg/util/word"
)

// BitwidthGadget is a general-purpose mechanism for enforcing type constraints
// (i.e. that a given register has a given bitwidth).  Depending on the width
// and config used, this generates appropriate constraints and/or modules as
// necessary for enforcing bitwidth constraints.
type BitwidthGadget[F field.Element[F]] struct {
	// Determines the largest bitwidth for which range constraints are
	// translated into AIR range constraints, versus  using a horizontal
	// bitwidth gadget.
	maxRangeConstraint uint
	// Enables the use of type proofs which exploit the
	// limitless prover. Specifically, modules with a recursive structure are
	// created specifically for the purpose of checking types.
	limitless bool
	// Schema into which constraints are placed.
	schema *air.SchemaBuilder[F]
}

// NewBitwidthGadget constructs a new bitwidth gadget.
func NewBitwidthGadget[F field.Element[F]](schema *air.SchemaBuilder[F]) *BitwidthGadget[F] {
	return &BitwidthGadget[F]{
		maxRangeConstraint: 8,
		limitless:          false,
		schema:             schema,
	}
}

// WithMaxRangeConstraint determines the cutoff for range cosntraints.
func (p *BitwidthGadget[F]) WithMaxRangeConstraint(width uint) *BitwidthGadget[F] {
	p.maxRangeConstraint = width
	return p
}

// WithLimitless enables or disables use of limitless type proofs.
func (p *BitwidthGadget[F]) WithLimitless(flag bool) *BitwidthGadget[F] {
	p.limitless = flag
	return p
}

// Constrain ensures all values in a given register fit within a given bitwidth.
func (p *BitwidthGadget[F]) Constrain(ref register.Ref, bitwidth uint) {
	var (
		module   = p.schema.Module(ref.Module())
		register = module.Register(ref.Register())
	)
	// Base cases
	switch {
	case bitwidth <= 1:
		p.applyBinaryGadget(ref)
		return
	case bitwidth <= p.maxRangeConstraint:
		handle := fmt.Sprintf("%s:u%d", register.Name, bitwidth)
		// Construct access to register
		access := term.RawRegisterAccess[F, air.Term[F]](ref.Register(), 0)
		// Add range constraint
		module.AddConstraint(air.NewRangeConstraint(handle, module.Id(), *access, bitwidth))
		// Done
		return
	case p.limitless:
		p.applyRecursiveBitwidthGadget(ref, bitwidth)
	default:
		// NOTE: this should be deprecated once the limitless prover is well
		// established.
		p.applyHorizontalBitwidthGadget(ref, bitwidth)
	}
}

// ApplyBinaryGadget adds a binarity constraint for a given column in the schema
// which enforces that all values in the given column are either 0 or 1. For a
// column X, this corresponds to the vanishing constraint X * (X-1) == 0.
func (p *BitwidthGadget[F]) applyBinaryGadget(ref register.Ref) {
	var (
		module   = p.schema.Module(ref.Module())
		register = module.Register(ref.Register())
		handle   = fmt.Sprintf("%s:u1", register.Name)
	)
	// Construct X
	X := term.NewRegisterAccess[F, air.Term[F]](ref.Register(), 0)
	// Construct X == 0
	X_eq0 := term.Subtract(X, term.Const64[F, air.Term[F]](0))
	// Construct X == 0
	X_eq1 := term.Subtract(X, term.Const64[F, air.Term[F]](1))
	// Construct (X==0) ∨ (X==1)
	X_X_m1 := term.Product(X_eq0, X_eq1)
	// Done!
	module.AddConstraint(
		air.NewVanishingConstraint(handle, module.Id(), util.None[int](), X_X_m1))
}

// ApplyHorizontalBitwidthGadget ensures all values in a given column fit within
// a given bitwidth.  This is implemented using a *horizontal byte
// decomposition* which adds n columns and a vanishing constraint (where n*8 >=
// bitwidth).
func (p *BitwidthGadget[F]) applyHorizontalBitwidthGadget(ref register.Ref, bitwidth uint) {
	var (
		module       = p.schema.Module(ref.Module())
		register     = module.Register(ref.Register())
		lookupHandle = fmt.Sprintf("%s:u%d", register.Name, bitwidth)
	)
	// Allocate computed byte registers in the given module, and add required
	// range constraints.
	byteRegisters := allocateByteRegisters(register.Name, bitwidth, module)
	// Build up the decomposition sum
	sum := buildDecompositionTerm[F](bitwidth, byteRegisters)
	// Construct X == (X:0 * 1) + ... + (X:n * 2^n)
	X := term.NewRegisterAccess[F, air.Term[F]](ref.Register(), 0)
	//
	eq := term.Subtract(X, sum)
	// Construct column name
	module.AddConstraint(
		air.NewVanishingConstraint(lookupHandle, module.Id(), util.None[int](), eq))
	// Add decomposition assignment
	module.AddAssignment(&byteDecomposition[F]{register.Name, bitwidth, ref, byteRegisters})
}

// ApplyRecursiveBitwidthGadget ensures all values in a given column fit within
// a given bitwidth. This is implemented using a combination of reference tables
// and lookups.  Specifically, if the width is below 16bits, then a static
// reference table is created along with a corresponding lookup,  Otherwise, a
// recursive procedure is applied whereby a table is created for the given width
// which divides each value into two smaller values.  This procedure is then
// recursively applied to those columns, etc.
func (p *BitwidthGadget[F]) applyRecursiveBitwidthGadget(ref register.Ref, bitwidth uint) {
	var (
		proofHandle  = module.Name{Name: fmt.Sprintf("u%d", bitwidth), Multiplier: 1}
		mod          = p.schema.Module(ref.Module())
		reg          = mod.Register(ref.Register())
		lookupHandle = fmt.Sprintf("%s:u%d", reg.Name, bitwidth)
	)
	// Recursive case.
	mid, ok := p.schema.HasModule(proofHandle)
	//
	if !ok {
		mid = p.constructTypeProof(proofHandle, bitwidth)
	}
	// Add lookup constraint for register into proof
	sourceAccesses := []*air.ColumnAccess[F]{
		// Source Value
		term.RawRegisterAccess[F, air.Term[F]](ref.Register(), 0)}
	// NOTE: 0th column always assumed to hold full value, with others
	// representing limbs, etc.
	targetAccesses := []*air.ColumnAccess[F]{
		// Target Value
		term.RawRegisterAccess[F, air.Term[F]](register.NewId(0), 0)}
	//
	targets := []lookup.Vector[F, *air.ColumnAccess[F]]{
		lookup.UnfilteredVector(mid, targetAccesses...)}
	sources := []lookup.Vector[F, *air.ColumnAccess[F]]{
		lookup.UnfilteredVector(mod.Id(), sourceAccesses...)}
	//
	mod.AddConstraint(air.NewLookupConstraint(lookupHandle, targets, sources))
	// Add column to assignment so its proof is included
	typeModule := p.schema.Module(mid)
	//
	decomposition := typeModule.Assignments().Next().(*typeDecomposition[F])
	decomposition.AddSource(ref)
}

func (p *BitwidthGadget[F]) constructTypeProof(handle module.Name, bitwidth uint) sc.ModuleId {
	var (
		// Create new module for this type proof
		mid    = p.schema.NewModule(handle, false, false, true)
		module = p.schema.Module(mid)
		// Determine limb widths.
		loWidth, hiWidth = determineLimbSplit(bitwidth)
		// Compute 2^loWidth to use as coefficient
		coeff = field.TwoPowN[F](loWidth)
		// Default padding
		zero big.Int
	)
	// Construct registers and their decompositions
	vid := module.NewRegister(register.NewComputed("V", bitwidth, zero))
	vidLo := module.NewRegister(register.NewComputed("V'0", loWidth, zero))
	vidHi := module.NewRegister(register.NewComputed("V'1", hiWidth, zero))
	// Ensure lo/hi are decomposition of original
	module.AddConstraint(
		air.NewVanishingConstraint("decomposition", mid, util.None[int](),
			term.Subtract(
				term.NewRegisterAccess[F, air.Term[F]](vid, 0),
				term.Sum(
					term.NewRegisterAccess[F, air.Term[F]](vidLo, 0),
					term.Product(term.Const[F, air.Term[F]](coeff),
						term.NewRegisterAccess[F, air.Term[F]](vidHi, 0)),
				),
			)))
	// Recursively proof lo/hi columns
	p.Constrain(register.NewRef(mid, vidLo), loWidth)
	p.Constrain(register.NewRef(mid, vidHi), hiWidth)
	// Construct corresponding register refs
	v := register.NewRef(mid, vid)
	vLo := register.NewRef(mid, vidLo)
	vHi := register.NewRef(mid, vidHi)
	// Add (initially empty) assignment
	module.AddAssignment(&typeDecomposition[F]{
		loWidth, hiWidth,
		nil,
		[]register.Ref{v, vLo, vHi},
	})
	// Done
	return mid
}

// ============================================================================
// Type Decomposition Assignment
// ============================================================================

type typeDecomposition[F field.Element[F]] struct {
	// Limb widths of decomposition.
	loWidth, hiWidth uint
	// Source registers being decomposed which represent the set of all the
	// columns in the entire system being proved to have the given bitwidth.
	sources []register.Ref
	// Target registers for decomposition.  There are always exactly three
	// registers, where the first holds the original value, followed by the
	// least significant (lo) limb and, finally, the most significant (hi) limb.
	targets []register.Ref
}

// AddSource adds a new source column to this decomposition.
func (p *typeDecomposition[F]) AddSource(source register.Ref) {
	p.sources = append(p.sources, source)
}

// Compute computes the values of columns defined by this assignment.
// This requires computing the value of each byte column in the decomposition.
func (p *typeDecomposition[F]) Compute(tr trace.Trace[F], schema sc.AnySchema[F],
) ([]array.MutArray[F], error) {
	// Read inputs
	sources := assignment.ReadRegistersRef(tr, p.sources...)
	padding := assignment.ReadPadding(tr, p.sources...)
	// Combine all sources
	combined := combineSources(p.loWidth+p.hiWidth, sources, padding, tr.Builder())
	// Generate decomposition
	data := computeDecomposition(p.loWidth, p.hiWidth, combined, tr.Builder())
	// Done
	return data, nil
}

// Bounds determines the well-definedness bounds for this assignment for both
// the negative (left) or positive (right) directions.  For example, consider an
// expression such as "(shift X -1)".  This is technically undefined for the
// first row of any trace and, by association, any constraint evaluating this
// expression on that first row is also undefined (and hence must pass).
func (p *typeDecomposition[F]) Bounds(_ sc.ModuleId) util.Bounds {
	return util.EMPTY_BOUND
}

// Consistent performs some simple checks that the given schema is consistent.
// This provides a double check of certain key properties, such as that
// registers used for assignments are large enough, etc.
func (p *typeDecomposition[F]) Consistent(schema sc.AnySchema[F]) []error {
	return nil
}

// RegistersExpanded identifies registers expanded by this assignment.
func (p *typeDecomposition[F]) RegistersExpanded() []register.Ref {
	return nil
}

// RegistersRead returns the set of columns that this assignment depends upon.
// That can include both input columns, as well as other computed columns.
func (p *typeDecomposition[F]) RegistersRead() []register.Ref {
	return p.sources
}

// RegistersWritten identifies registers assigned by this assignment.
func (p *typeDecomposition[F]) RegistersWritten() []register.Ref {
	return p.targets
}

// Substitute any matchined labelled constants within this assignment
func (p *typeDecomposition[F]) Substitute(mapping map[string]F) {
	// Nothing to do here.
}

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (p *typeDecomposition[F]) Lisp(schema sc.AnySchema[F]) sexp.SExp {
	var (
		targets = sexp.EmptyList()
		sources = sexp.EmptyList()
	)

	for _, ref := range p.targets {
		module := schema.Module(ref.Module())
		ith := module.Register(ref.Register())
		name := sexp.NewSymbol(ith.QualifiedName(module))
		datatype := sexp.NewSymbol(fmt.Sprintf("u%d", ith.Width))
		def := sexp.NewList([]sexp.SExp{name, datatype})
		targets.Append(def)
	}

	for _, ref := range p.sources {
		module := schema.Module(ref.Module())
		ith := module.Register(ref.Register())
		ithName := ith.QualifiedName(module)
		sources.Append(sexp.NewSymbol(ithName))
	}

	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("decompose"),
		targets,
		sources,
	})
}

// ============================================================================
// Byte Decomposition Assignment
// ============================================================================

// byteDecomposition is part of a range constraint for wide columns (e.g. u32)
// implemented using a byte decomposition.
type byteDecomposition[F field.Element[F]] struct {
	// Handle for identifying this assignment
	handle string
	// Width of decomposition.
	bitwidth uint
	// The source register being decomposed
	source register.Ref
	// Target registers holding the decomposition
	targets []register.Ref
}

// Compute computes the values of columns defined by this assignment.
// This requires computing the value of each byte column in the decomposition.
func (p *byteDecomposition[F]) Compute(tr trace.Trace[F], schema sc.AnySchema[F],
) ([]array.MutArray[F], error) {
	var n = uint(len(p.targets))
	// Read inputs
	sources := assignment.ReadRegistersRef(tr, p.source)
	// Apply native function
	data := byteDecompositionNativeFunction(n, sources, tr.Builder())
	//
	return data, nil
}

// Bounds determines the well-definedness bounds for this assignment for both
// the negative (left) or positive (right) directions.  For example, consider an
// expression such as "(shift X -1)".  This is technically undefined for the
// first row of any trace and, by association, any constraint evaluating this
// expression on that first row is also undefined (and hence must pass).
func (p *byteDecomposition[F]) Bounds(_ sc.ModuleId) util.Bounds {
	return util.EMPTY_BOUND
}

// Consistent performs some simple checks that the given schema is consistent.
// This provides a double check of certain key properties, such as that
// registers used for assignments are large enough, etc.
func (p *byteDecomposition[F]) Consistent(schema sc.AnySchema[F]) []error {
	var (
		bitwidth = schema.Register(p.source).Width
		total    = uint(0)
		errors   []error
	)
	//
	for _, ref := range p.targets {
		reg := schema.Module(ref.Module()).Register(ref.Register())
		total += reg.Width
	}
	//
	if total != bitwidth {
		err := fmt.Errorf("inconsistent byte decomposition (decomposed %d bits, but expected %d)", total, bitwidth)
		errors = append(errors, err)
	}
	//
	return errors
}

// RegistersExpanded identifies registers expanded by this assignment.
func (p *byteDecomposition[F]) RegistersExpanded() []register.Ref {
	return nil
}

// RegistersRead returns the set of columns that this assignment depends upon.
// That can include both input columns, as well as other computed columns.
func (p *byteDecomposition[F]) RegistersRead() []register.Ref {
	return []register.Ref{p.source}
}

// RegistersWritten identifies registers assigned by this assignment.
func (p *byteDecomposition[F]) RegistersWritten() []register.Ref {
	return p.targets
}

// Substitute any matchined labelled constants within this assignment
func (p *byteDecomposition[F]) Substitute(mapping map[string]F) {
	// Nothing to do here.
}

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (p *byteDecomposition[F]) Lisp(schema sc.AnySchema[F]) sexp.SExp {
	var (
		srcModule = schema.Module(p.source.Module())
		source    = srcModule.Register(p.source.Register())
		targets   = sexp.EmptyList()
	)
	//
	for _, t := range p.targets {
		tgtModule := schema.Module(t.Module())
		reg := tgtModule.Register(t.Register())
		targets.Append(sexp.NewList([]sexp.SExp{
			// name
			sexp.NewSymbol(reg.QualifiedName(tgtModule)),
			// type
			sexp.NewSymbol(fmt.Sprintf("u%d", reg.Width)),
		}))
	}

	return sexp.NewList(
		[]sexp.SExp{sexp.NewSymbol("decompose"),
			targets,
			sexp.NewSymbol(source.QualifiedName(srcModule)),
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

// Combine all values from the given source registers into a single array of
// data, whilst eliminating duplicates.
func combineSources[F field.Element[F]](bitwidth uint, sources []array.Array[F], padding []F,
	pool array.Builder[F]) array.MutArray[F] {
	//
	var (
		n    = sources[0].Len()
		arr  = pool.NewArray(0, bitwidth)
		seen = hash.NewSet[F](n)
	)
	// Add all values
	for _, src := range sources {
		// Add all values from column
		for i := range src.Len() {
			ith := src.Get(i)
			// Add item if not already seen
			if !seen.Contains(ith) {
				// record have seen item
				seen.Insert(ith)
				// append and record
				arr.Append(ith)
			}
		}
	}
	// Add all padding values
	for _, ith := range padding {
		// Add item if not already seen
		if !seen.Contains(ith) {
			// record have seen item
			seen.Insert(ith)
			// append and record
			arr.Append(ith)
		}
	}
	// Done
	return arr
}

func computeDecomposition[F field.Element[F]](loWidth, hiWidth uint, vArr array.MutArray[F],
	builder array.Builder[F]) []array.MutArray[F] {
	//
	var (
		vLoArr = builder.NewArray(vArr.Len(), loWidth)
		vHiArr = builder.NewArray(vArr.Len(), hiWidth)
	)
	//
	for i := range vArr.Len() {
		ith := vArr.Get(i)
		lo, hi := decompose(loWidth, ith)
		vLoArr.Set(i, lo)
		vHiArr.Set(i, hi)
	}
	//
	return []array.MutArray[F]{vArr, vLoArr, vHiArr}
}

// Decompose a given field element into its least and most significant limbs,
// based on the required bitwidth for the least significant limb.
func decompose[F field.Element[F]](loWidth uint, ith F) (F, F) {
	// Extract bytes from element
	var (
		bytes       = ith.Bytes()
		loByteWidth = loWidth / 8
		loFr, hiFr  F
		n           = uint(len(bytes))
	)
	// Sanity check assumption
	if loWidth%8 != 0 {
		panic("unreachable")
	}
	//
	if loByteWidth >= n {
		// no high bytes at all
		loFr = loFr.SetBytes(bytes)
	} else {
		// Determine pivot
		n = n - loByteWidth
		// Split bytes
		hiFr = hiFr.SetBytes(bytes[:n])
		loFr = loFr.SetBytes(bytes[n:])
	}
	//
	return loFr, hiFr
}

// ============================================================================
// Helpers (for horizontal)
// ============================================================================

// Allocate n byte registers, each of which requires a suitable range
// constraint.
func allocateByteRegisters[F field.Element[F]](prefix string, bitwidth uint, module *air.ModuleBuilder[F],
) []register.Ref {
	var (
		n    = bitwidth / 8
		zero big.Int
	)
	//
	if bitwidth == 0 {
		panic("zero byte decomposition encountered")
	}
	// Account for asymetric case
	if bitwidth%8 != 0 {
		n++
	}
	// Allocate target register ids
	targets := make([]register.Ref, n)
	// Allocate byte registers
	for i := uint(0); i < n; i++ {
		name := fmt.Sprintf("%s:%d", prefix, i)
		byteRegister := register.NewComputed(name, min(8, bitwidth), zero)
		// Allocate byte register
		rid := module.NewRegister(byteRegister)
		targets[i] = register.NewRef(module.Id(), rid)
		// Add suitable range constraint
		ith_access := term.RawRegisterAccess[F, air.Term[F]](rid, 0)
		//
		module.AddConstraint(
			air.NewRangeConstraint(name, module.Id(), *ith_access, byteRegister.Width))
		//
		bitwidth -= 8
	}
	//
	return targets
}

func buildDecompositionTerm[F field.Element[F]](bitwidth uint, byteRegisters []register.Ref) air.Term[F] {
	var (
		// Determine ranges required for the give bitwidth
		ranges = splitColumnRanges[F](bitwidth)
		// Initialise array of terms
		terms = make([]air.Term[F], len(byteRegisters))
		// Initialise coefficient
		coefficient F = field.One[F]()
	)
	// Construct Columns
	for i, ref := range byteRegisters {
		// Create Column + Constraint
		reg := term.NewRegisterAccess[F, air.Term[F]](ref.Register(), 0)
		terms[i] = term.Product(reg, term.Const[F, air.Term[F]](coefficient))
		// Update coefficient
		coefficient = coefficient.Mul(ranges[i])
	}
	// Construct (X:0 * 1) + ... + (X:n * 2^n)
	return term.Sum(terms...)
}

func splitColumnRanges[F field.Element[F]](nbits uint) []F {
	var (
		n      = nbits / 8
		m      = nbits % 8
		ranges []F
		// FIXME: following fails for very small fields like GF251!
		two8 F = field.Uint64[F](256)
	)
	//
	if m == 0 {
		ranges = make([]F, n)
	} else {
		// Most significant column has smaller range.
		ranges = make([]F, n+1)
		// Determine final range
		ranges[n] = field.TwoPowN[F](m)
	}
	//
	for i := range n {
		ranges[i] = two8
	}
	//
	return ranges
}

func byteDecompositionNativeFunction[F field.Element[F]](n uint, sources []array.Array[F],
	builder array.Builder[F]) []array.MutArray[F] {
	//
	var (
		source  = sources[0]
		targets = make([]array.MutArray[F], n)
		height  = source.Len()
	)
	// Sanity check
	if len(sources) != 1 {
		panic("too many source columns for byte decomposition")
	}
	// Initialise columns
	for i := range n {
		// Construct a byte array for ith byte
		targets[i] = builder.NewArray(height, 8)
	}
	// Decompose each row of each column
	for i := range height {
		ith := decomposeIntoBytes(source.Get(i), n)
		for j := uint(0); j < n; j++ {
			targets[j].Set(i, ith[j])
		}
	}
	//
	return targets
}

// Decompose a given element into n bytes in little endian form.  For example,
// decomposing 41b into 2 bytes gives [0x1b,0x04].
func decomposeIntoBytes[W word.Word[W]](val W, n uint) []W {
	// Construct return array
	elements := make([]W, n)
	// Determine bytes of this value (in big endian form).
	bytes := val.Bytes()
	//
	l := uint(len(bytes))
	//
	m := min(n, l)
	// Convert each byte into a field element
	for i := range m {
		var (
			ith W
			j   = l - i - 1
		)
		//
		elements[i] = ith.SetUint64(uint64(bytes[j]))
	}
	// Done
	return elements
}
