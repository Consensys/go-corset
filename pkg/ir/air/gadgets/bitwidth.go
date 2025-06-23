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
	"github.com/consensys/go-corset/pkg/ir"
	"github.com/consensys/go-corset/pkg/ir/air"
	"github.com/consensys/go-corset/pkg/ir/assignment"
	"github.com/consensys/go-corset/pkg/schema"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
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
	// Enables the use of type proofs which exploit the
	// limitless prover. Specifically, modules with a recursive structure are
	// created specifically for the purpose of checking types.
	limitless bool
	// Schema into which constraints are placed.
	schema *air.SchemaBuilder
}

// NewBitwidthGadget constructs a new bitwidth gadget.
func NewBitwidthGadget(schema *air.SchemaBuilder) *BitwidthGadget {
	return &BitwidthGadget{
		maxRangeConstraint: 8,
		limitless:          false,
		schema:             schema,
	}
}

// WithMaxRangeConstraint determines the cutoff for range cosntraints.
func (p *BitwidthGadget) WithMaxRangeConstraint(width uint) *BitwidthGadget {
	p.maxRangeConstraint = width
	return p
}

// WithLimitless enables or disables use of limitless type proofs.
func (p *BitwidthGadget) WithLimitless(flag bool) *BitwidthGadget {
	p.limitless = flag
	return p
}

// Constrain ensures all values in a given register fit within a given bitwidth.
func (p *BitwidthGadget) Constrain(ref sc.RegisterRef, bitwidth uint) {
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
		access := ir.RawRegisterAccess[air.Term](ref.Register(), 0)
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
func (p *BitwidthGadget) applyBinaryGadget(ref sc.RegisterRef) {
	var (
		module   = p.schema.Module(ref.Module())
		register = module.Register(ref.Register())
		handle   = fmt.Sprintf("%s:u1", register.Name)
	)
	// Construct X
	X := ir.NewRegisterAccess[air.Term](ref.Register(), 0)
	// Construct X == 0
	X_eq0 := ir.Subtract(X, ir.Const64[air.Term](0))
	// Construct X == 0
	X_eq1 := ir.Subtract(X, ir.Const64[air.Term](1))
	// Construct (X==0) ∨ (X==1)
	X_X_m1 := ir.Product(X_eq0, X_eq1)
	// Done!
	module.AddConstraint(
		air.NewVanishingConstraint(handle, module.Id(), util.None[int](), X_X_m1))
}

// ApplyHorizontalBitwidthGadget ensures all values in a given column fit within
// a given bitwidth.  This is implemented using a *horizontal byte
// decomposition* which adds n columns and a vanishing constraint (where n*8 >=
// bitwidth).
func (p *BitwidthGadget) applyHorizontalBitwidthGadget(ref sc.RegisterRef, bitwidth uint) {
	var (
		module       = p.schema.Module(ref.Module())
		register     = module.Register(ref.Register())
		lookupHandle = fmt.Sprintf("%s:u%d", register.Name, bitwidth)
	)
	// Allocate computed byte registers in the given module, and add required
	// range constraints.
	byteRegisters := allocateByteRegisters(register.Name, bitwidth, module)
	// Build up the decomposition sum
	sum := buildDecompositionTerm(bitwidth, byteRegisters)
	// Construct X == (X:0 * 1) + ... + (X:n * 2^n)
	X := ir.NewRegisterAccess[air.Term](ref.Register(), 0)
	//
	eq := ir.Subtract(X, sum)
	// Construct column name
	module.AddConstraint(
		air.NewVanishingConstraint(lookupHandle, module.Id(), util.None[int](), eq))
	// Add decomposition assignment
	module.AddAssignment(&byteDecomposition{register.Name, bitwidth, ref, byteRegisters})
}

// ApplyRecursiveBitwidthGadget ensures all values in a given column fit within
// a given bitwidth. This is implemented using a combination of reference tables
// and lookups.  Specifically, if the width is below 16bits, then a static
// reference table is created along with a corresponding lookup,  Otherwise, a
// recursive procedure is applied whereby a table is created for the given width
// which divides each value into two smaller values.  This procedure is then
// recursively applied to those columns, etc.
func (p *BitwidthGadget) applyRecursiveBitwidthGadget(ref sc.RegisterRef, bitwidth uint) {
	var (
		module       = p.schema.Module(ref.Module())
		register     = module.Register(ref.Register())
		proofHandle  = fmt.Sprintf("u%d", bitwidth)
		lookupHandle = fmt.Sprintf("%s:u%d", register.Name, bitwidth)
	)
	// Recursive case.
	mid, ok := p.schema.HasModule(proofHandle)
	//
	if !ok {
		mid = p.constructTypeProof(proofHandle, bitwidth)
	}
	// Add lookup constraint for register into proof
	sources := []*air.ColumnAccess{ir.RawRegisterAccess[air.Term](ref.Register(), 0)}
	// NOTE: 0th column always assumed to hold full value, with others
	// representing limbs, etc.
	targets := []*air.ColumnAccess{ir.RawRegisterAccess[air.Term](sc.NewRegisterId(0), 0)}
	//
	module.AddConstraint(
		air.NewLookupConstraint(lookupHandle, mid, targets, module.Id(), sources))
	// Add column to assignment so its proof is included
	typeModule := p.schema.Module(mid)
	//
	decomposition := typeModule.Assignments().Next().(*typeDecomposition)
	decomposition.AddSource(ref)
}

func (p *BitwidthGadget) constructTypeProof(handle string, bitwidth uint) sc.ModuleId {
	var (
		// Create new module for this type proof
		mid    = p.schema.NewModule(handle, 1)
		module = p.schema.Module(mid)
		// Determine limb widths.
		loWidth, hiWidth = determineLimbSplit(bitwidth)
	)
	// Construct registers and their decompositions
	vid := module.NewRegister(sc.NewComputedRegister("V", bitwidth))
	vidLo := module.NewRegister(sc.NewComputedRegister("V'0", loWidth))
	vidHi := module.NewRegister(sc.NewComputedRegister("V'1", hiWidth))
	// Compute 2^loWidth to use as coefficient
	coeff := fr.NewElement(2)
	field.Pow(&coeff, uint64(loWidth))
	// Ensure lo/hi are decomposition of original
	module.AddConstraint(
		air.NewVanishingConstraint("decomposition", mid, util.None[int](),
			ir.Subtract(
				ir.NewRegisterAccess[air.Term](vid, 0),
				ir.Sum(
					ir.NewRegisterAccess[air.Term](vidLo, 0),
					ir.Product(ir.Const[air.Term](coeff), ir.NewRegisterAccess[air.Term](vidHi, 0)),
				),
			)))
	// Recursively proof lo/hi columns
	p.Constrain(sc.NewRegisterRef(mid, vidLo), loWidth)
	p.Constrain(sc.NewRegisterRef(mid, vidHi), hiWidth)
	// Construct corresponding register refs
	v := sc.NewRegisterRef(mid, vid)
	vLo := sc.NewRegisterRef(mid, vidLo)
	vHi := sc.NewRegisterRef(mid, vidHi)
	// Add (initially empty) assignment
	module.AddAssignment(&typeDecomposition{
		loWidth, hiWidth,
		nil,
		[]sc.RegisterRef{v, vLo, vHi},
	})
	// Done
	return mid
}

// ============================================================================
// Type Decomposition Assignment
// ============================================================================

type typeDecomposition struct {
	// Limb widths of decomposition.
	loWidth, hiWidth uint
	// Source registers being decomposed which represent the set of all the
	// columns in the entire system being proved to have the given bitwidth.
	sources []sc.RegisterRef
	// Target registers for decomposition.  There are always exactly three
	// registers, where the first holds the original value, followed by the
	// least significant (lo) limb and, finally, the most significant (hi) limb.
	targets []sc.RegisterRef
}

// AddSource adds a new source column to this decomposition.
func (p *typeDecomposition) AddSource(source sc.RegisterRef) {
	p.sources = append(p.sources, source)
}

// Compute computes the values of columns defined by this assignment.
// This requires computing the value of each byte column in the decomposition.
func (p *typeDecomposition) Compute(tr trace.Trace, schema sc.AnySchema) ([]trace.ArrayColumn, error) {
	// Read inputs
	sources := assignment.ReadRegisters(tr, p.sources...)
	// Combine all sources
	combined := combineSources(p.loWidth+p.hiWidth, sources)
	// Generate decomposition
	data := computeDecomposition(p.loWidth, p.hiWidth, combined)
	// Write outputs
	targets := assignment.WriteRegisters(schema, p.targets, data)
	// Done
	return targets, nil
}

// Bounds determines the well-definedness bounds for this assignment for both
// the negative (left) or positive (right) directions.  For example, consider an
// expression such as "(shift X -1)".  This is technically undefined for the
// first row of any trace and, by association, any constraint evaluating this
// expression on that first row is also undefined (and hence must pass).
func (p *typeDecomposition) Bounds(_ sc.ModuleId) util.Bounds {
	return util.EMPTY_BOUND
}

// Consistent performs some simple checks that the given schema is consistent.
// This provides a double check of certain key properties, such as that
// registers used for assignments are large enough, etc.
func (p *typeDecomposition) Consistent(schema sc.AnySchema) []error {
	return nil
}

// RegistersExpanded identifies registers expanded by this assignment.
func (p *typeDecomposition) RegistersExpanded() []sc.RegisterRef {
	return nil
}

// RegistersRead returns the set of columns that this assignment depends upon.
// That can include both input columns, as well as other computed columns.
func (p *typeDecomposition) RegistersRead() []sc.RegisterRef {
	return p.sources
}

// RegistersWritten identifies registers assigned by this assignment.
func (p *typeDecomposition) RegistersWritten() []sc.RegisterRef {
	return p.targets
}

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (p *typeDecomposition) Lisp(schema sc.AnySchema) sexp.SExp {
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
type byteDecomposition struct {
	// Handle for identifying this assignment
	handle string
	// Width of decomposition.
	bitwidth uint
	// The source register being decomposed
	source sc.RegisterRef
	// Target registers holding the decomposition
	targets []sc.RegisterRef
}

// Compute computes the values of columns defined by this assignment.
// This requires computing the value of each byte column in the decomposition.
func (p *byteDecomposition) Compute(tr trace.Trace, schema sc.AnySchema) ([]trace.ArrayColumn, error) {
	var n = uint(len(p.targets))
	// Read inputs
	sources := assignment.ReadRegisters(tr, p.source)
	// Apply native function
	data := byteDecompositionNativeFunction(n, sources)
	// Write outputs
	targets := assignment.WriteRegisters(schema, p.targets, data)
	//
	return targets, nil
}

// Bounds determines the well-definedness bounds for this assignment for both
// the negative (left) or positive (right) directions.  For example, consider an
// expression such as "(shift X -1)".  This is technically undefined for the
// first row of any trace and, by association, any constraint evaluating this
// expression on that first row is also undefined (and hence must pass).
func (p *byteDecomposition) Bounds(_ sc.ModuleId) util.Bounds {
	return util.EMPTY_BOUND
}

// Consistent performs some simple checks that the given schema is consistent.
// This provides a double check of certain key properties, such as that
// registers used for assignments are large enough, etc.
func (p *byteDecomposition) Consistent(schema sc.AnySchema) []error {
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
func (p *byteDecomposition) RegistersExpanded() []sc.RegisterRef {
	return nil
}

// RegistersRead returns the set of columns that this assignment depends upon.
// That can include both input columns, as well as other computed columns.
func (p *byteDecomposition) RegistersRead() []sc.RegisterRef {
	return []sc.RegisterRef{p.source}
}

// RegistersWritten identifies registers assigned by this assignment.
func (p *byteDecomposition) RegistersWritten() []sc.RegisterRef {
	return p.targets
}

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (p *byteDecomposition) Lisp(schema sc.AnySchema) sexp.SExp {
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

// Allocate n byte registers, each of which requires a suitable range
// constraint.
func allocateByteRegisters(prefix string, bitwidth uint, module *air.ModuleBuilder) []sc.RegisterRef {
	var n = bitwidth / 8
	//
	if bitwidth == 0 {
		panic("zero byte decomposition encountered")
	}
	// Account for asymetric case
	if bitwidth%8 != 0 {
		n++
	}
	// Allocate target register ids
	targets := make([]schema.RegisterRef, n)
	// Allocate byte registers
	for i := uint(0); i < n; i++ {
		name := fmt.Sprintf("%s:%d", prefix, i)
		byteRegister := schema.NewComputedRegister(name, min(8, bitwidth))
		// Allocate byte register
		rid := module.NewRegister(byteRegister)
		targets[i] = sc.NewRegisterRef(module.Id(), rid)
		// Add suitable range constraint
		ith_access := ir.RawRegisterAccess[air.Term](rid, 0)
		//
		module.AddConstraint(
			air.NewRangeConstraint(name, module.Id(), *ith_access, byteRegister.Width))
		//
		bitwidth -= 8
	}
	//
	return targets
}

func buildDecompositionTerm(bitwidth uint, byteRegisters []sc.RegisterRef) air.Term {
	var (
		// Determine ranges required for the give bitwidth
		ranges = splitColumnRanges(bitwidth)
		// Initialise array of terms
		terms = make([]air.Term, len(byteRegisters))
		// Initialise coefficient
		coefficient = fr.NewElement(1)
	)

	// Construct Columns
	for i, ref := range byteRegisters {
		// Create Column + Constraint
		reg := ir.NewRegisterAccess[air.Term](ref.Register(), 0)
		terms[i] = ir.Product(reg, ir.Const[air.Term](coefficient))
		// Update coefficient
		coefficient.Mul(&coefficient, &ranges[i])
	}
	// Construct (X:0 * 1) + ... + (X:n * 2^n)
	return ir.Sum(terms...)
}

func splitColumnRanges(nbits uint) []fr.Element {
	var (
		n      = nbits / 8
		m      = int64(nbits % 8)
		ranges []fr.Element
		fr256  = fr.NewElement(256)
	)
	//
	if m == 0 {
		ranges = make([]fr.Element, n)
	} else {
		var last fr.Element
		// Most significant column has smaller range.
		ranges = make([]fr.Element, n+1)
		// Determine final range
		last.Exp(fr.NewElement(2), big.NewInt(m))
		//
		ranges[n] = last
	}
	//
	for i := uint(0); i < n; i++ {
		ranges[i] = fr256
	}
	//
	return ranges
}

func byteDecompositionNativeFunction(n uint, sources []field.FrArray) []field.FrArray {
	var (
		source  = sources[0]
		targets = make([]field.FrArray, n)
		height  = source.Len()
	)
	// Sanity check
	if len(sources) != 1 {
		panic("too many source columns for byte decomposition")
	}
	// Initialise columns
	for i := range n {
		// Construct a byte array for ith byte
		targets[i] = field.NewFrArray(height, 8)
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
func decomposeIntoBytes(val fr.Element, n uint) []fr.Element {
	// Construct return array
	elements := make([]fr.Element, n)
	// Determine bytes of this value (in big endian form).
	bytes := val.Bytes()
	m := uint(len(bytes) - 1)
	// Convert each byte into a field element
	for i := uint(0); i < n; i++ {
		j := m - i
		ith := fr.NewElement(uint64(bytes[j]))
		elements[i] = ith
	}

	// Done
	return elements
}
