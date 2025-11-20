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
package trace

import (
	"bytes"
	"cmp"
	"encoding/gob"
	"fmt"
	"math"
)

// CellRef identifies a unique cell within a given table.
type CellRef struct {
	// Column index for the cell
	Column ColumnRef
	// Row index for the cell
	Row int
}

// NewCellRef constructs a new cell reference.
func NewCellRef(Column ColumnRef, Row int) CellRef {
	return CellRef{Column, Row}
}

// Cmp implementation for the set.Comparable interface. This allows a CellRef to
// be used in an AnySortedSet.
func (p CellRef) Cmp(q CellRef) int {
	var c = p.Column.Cmp(q.Column)
	//
	if c == 0 {
		c = cmp.Compare(p.Row, q.Row)
	}
	//
	return c
}

// ============================================================================

// ColumnRef abstracts a complete (i.e. global) Column identifier.
type ColumnRef struct {
	// Module containing this Column
	mid ModuleId
	// Column index within that module
	rid ColumnId
}

// NewColumnRef constructs a new Column reference from the given module and
// Column identifiers.
func NewColumnRef(mid ModuleId, rid ColumnId) ColumnRef {
	return ColumnRef{mid, rid}
}

// NewIndexedColumnRef constructs a new Column reference from the given column
// index computed using Index() and the given width.
func NewIndexedColumnRef(index uint, width uint) ColumnRef {
	mid := index % width
	rid := index / width
	//
	return NewColumnRef(mid, NewColumnId(rid))
}

// Cmp implementation for the set.Comparable interface. This allows a CellRef to
// be used in an AnySortedSet.
func (p ColumnRef) Cmp(q ColumnRef) int {
	var c = cmp.Compare(p.mid, q.mid)
	//
	if c == 0 {
		c = cmp.Compare(p.rid.Unwrap(), q.rid.Unwrap())
	}
	//
	return c
}

// Index returns a unique index for this column, assuming a given number of
// modules.
func (p ColumnRef) Index(nModules uint) uint {
	return p.mid + (nModules * p.rid.index)
}

// Module returns the module identifier of this Column reference.
func (p ColumnRef) Module() ModuleId {
	return p.mid
}

// Column returns the Column identifier of this Column reference.
func (p ColumnRef) Column() ColumnId {
	return p.rid
}

// Register returns the register (i.e. column) identifier of this reference.
// Since this type is also used in schema, this function is included here for
// convenience.
func (p ColumnRef) Register() ColumnId {
	return p.rid
}

// ModuleId abstracts the notion of a "module identifier"
type ModuleId = uint

// ColumnId captures the notion of a column index.  That is, for each
// module, every Column is allocated a given index starting from 0.  The
// purpose of the wrapper is avoid confusion between uint values and things
// which are expected to identify Columns.
type ColumnId struct {
	index uint
}

// NewColumnId constructs a new Column ID from a given raw index.
func NewColumnId(index uint) ColumnId {
	return ColumnId{index}
}

// NewUnusedColumnId constructs something akin to a null reference.  This is
// used in some situations where we may (or may not) want to refer to a specific
// Column.
func NewUnusedColumnId() ColumnId {
	return ColumnId{math.MaxUint}
}

// Cmp implementation for the Comparable interface
func (p ColumnId) Cmp(o ColumnId) int {
	return cmp.Compare(p.index, o.index)
}

// Unwrap returns the underlying Column index.
func (p ColumnId) Unwrap() uint {
	if p.index == math.MaxUint {
		panic("attempt to unwrap unused Column id")
	}
	//
	return p.index
}

// Id returns the underlying register id for this identifier.
func (p ColumnId) Id() ColumnId {
	return p
}

// IsUsed checks whether this corresponds to a valid Column index.
func (p ColumnId) IsUsed() bool {
	return p.index != math.MaxUint
}

// AccessOf constructs a relative column identifier from this identifier by
// including a bitwidth and relative shift.
func (p ColumnId) AccessOf(bitwidth uint) ColumnAccessId {
	//
	return ColumnAccessId{p, bitwidth, bitwidth, 0}
}

func (p ColumnId) String() string {
	return fmt.Sprintf("%d", p.index)
}

// ============================================================================
// ColumnAccessId
// ============================================================================

// ColumnAccessId is a wrapper around a column Id which adds a "relative shift"
// and a "bitwidth access".  That is, it identifies a column on a relative row
// from the given row.  It also enables direct column casting.
type ColumnAccessId struct {
	id ColumnId
	// bitwidth is the declared bitwidth of this column.
	bitwidth uint
	// maskwidth is the portion of the column actually used (which maybe less
	// than the declared width).
	maskwidth uint
	// shift is the relative shift applied to determine which row of the column
	// is being accessed.
	shift int
}

// Cmp implementation for the Comparable interface
func (p ColumnAccessId) Cmp(o ColumnAccessId) int {
	if c := p.id.Cmp(o.id); c != 0 {
		return c
	} else if c := cmp.Compare(p.shift, o.shift); c != 0 {
		return c
	} else if c := cmp.Compare(p.bitwidth, o.bitwidth); c != 0 {
		return c
	}
	//
	return cmp.Compare(p.maskwidth, o.maskwidth)
}

// Id returns returns the underlying register id.
func (p ColumnAccessId) Id() ColumnId {
	return p.id
}

// BitWidth returns the declared bitwidth of the variable being accessed.
// Observe that the actual width of this access may be smaller than this if a
// mask is being applied.
func (p ColumnAccessId) BitWidth() uint {
	return p.bitwidth
}

// MaskWidth returns the portion of the underlying column / register actually
// read by this access.  For example, given a register of type u16 we might only
// be accessing the first u8 portion.
func (p ColumnAccessId) MaskWidth() uint {
	return p.maskwidth
}

// Shift shifts the given column access by a given (relative) amount.
func (p ColumnAccessId) Shift(shift int) ColumnAccessId {
	return ColumnAccessId{
		p.id, p.bitwidth, p.maskwidth, p.shift + shift,
	}
}

// Mask masks the given column access by a given number of bits.
func (p ColumnAccessId) Mask(mask uint) ColumnAccessId {
	if mask > p.bitwidth {
		panic(fmt.Sprintf("invalid mask (u%d > u%d)", mask, p.bitwidth))
	}
	//
	return ColumnAccessId{
		p.id, p.bitwidth, mask, p.shift,
	}
}

// RelativeShift returns the relative shift for this register.
func (p ColumnAccessId) RelativeShift() int {
	return p.shift
}

// Unwrap returns the underlying Column index.
func (p ColumnAccessId) Unwrap() uint {
	return p.id.Unwrap()
}

// ============================================================================
// Encoding / Decoding
// ============================================================================

// GobEncode an option.  This allows it to be marshalled into a binary form.
func (p ColumnId) GobEncode() (data []byte, err error) {
	var (
		buffer     bytes.Buffer
		gobEncoder = gob.NewEncoder(&buffer)
	)
	//
	if err := gobEncoder.Encode(&p.index); err != nil {
		return nil, err
	}
	// Done
	return buffer.Bytes(), nil
}

// GobDecode a previously encoded option
func (p *ColumnId) GobDecode(data []byte) error {
	var (
		buffer     = bytes.NewBuffer(data)
		gobDecoder = gob.NewDecoder(buffer)
	)
	//
	if err := gobDecoder.Decode(&p.index); err != nil {
		return err
	}
	// Success!
	return nil
}

// GobEncode an option.  This allows it to be marshalled into a binary form.
func (p ColumnRef) GobEncode() (data []byte, err error) {
	var (
		rid        = p.rid.Unwrap()
		buffer     bytes.Buffer
		gobEncoder = gob.NewEncoder(&buffer)
	)
	//
	if err := gobEncoder.Encode(&p.mid); err != nil {
		return nil, err
	}
	//
	if err := gobEncoder.Encode(&rid); err != nil {
		return nil, err
	}
	// Done
	return buffer.Bytes(), nil
}

// GobDecode a previously encoded option
func (p *ColumnRef) GobDecode(data []byte) error {
	var (
		rid        uint
		buffer     = bytes.NewBuffer(data)
		gobDecoder = gob.NewDecoder(buffer)
	)
	//
	if err := gobDecoder.Decode(&p.mid); err != nil {
		return err
	}
	//
	if err := gobDecoder.Decode(&rid); err != nil {
		return err
	}
	// Construct reg id
	p.rid = NewColumnId(rid)
	// Success!
	return nil
}
