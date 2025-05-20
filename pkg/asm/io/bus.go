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
package io

// Bus describes an I/O bus referred to within a function.  Every function can
// connect with zero or more buses.  For example, making a function call
// requires a bus for the target function.  Each bus consists of some number of
// _address lines_ and some number of _data lines_.  Reading a value from the
// bus requires setting the address lines, then reading the data lines.
// Likewise, put a value onto the bus requires setting both the address and data
// lines.
type Bus struct {
	// Bus name
	Name string
	// Global bus identifier.  This uniquely identifies the bus across all
	// functions and components.
	BusId uint
	// Determines the address lines of this bus.
	address []uint
	// Determiunes the data lines of this bus.
	data []uint
}

// IsUnlinked checks whether a given bus has been linked already or not.
func (p *Bus) IsUnlinked() bool {
	return p.BusId == UNKNOWN_BUS
}

// UnlinkedBus constructs a new bus which is not yet connected with anything.
// Rather it simply has a name which will be used later to establish the
// connection.
func UnlinkedBus(name string) Bus {
	return Bus{name, UNKNOWN_BUS, nil, nil}
}

// NewBus constructs a new bus with the given components.
func NewBus(name string, id uint, address []uint, data []uint) Bus {
	return Bus{name, id, address, data}
}

// Address returns the "address lines" for this bus.  That is, the registers
// which hold the various components of the address.
func (p *Bus) Address() []uint {
	return p.address
}

// Data returns the "data lines" for this bus.  That is, the registers which
// hold the various data values (either being read or written).
func (p *Bus) Data() []uint {
	return p.data
}

// AddressData returns the "address" and "data" lines for this bus (in that
// order).  That is, the registers which hold the various components of the
// address.
func (p *Bus) AddressData() []uint {
	return append(p.address, p.data...)
}
