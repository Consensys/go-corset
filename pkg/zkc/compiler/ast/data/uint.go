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
package data

import "fmt"

// UnsignedInt captures the fundamental data type of the language.
type UnsignedInt struct {
	bitwidth uint
}

// NewUnsignedInt constructs an unsigned int type of a given width.
func NewUnsignedInt(bitwidth uint) *UnsignedInt {
	return &UnsignedInt{bitwidth}
}

// BitWidth implementation for Type interface
func (p *UnsignedInt) BitWidth() uint {
	return p.bitwidth
}

// Flattern implementation for Type interface
func (p *UnsignedInt) Flattern(prefix string, constructor func(name string, bitwidth uint)) {
	constructor(prefix, p.bitwidth)
}

func (p *UnsignedInt) String() string {
	return fmt.Sprintf("u%d", p.bitwidth)
}
