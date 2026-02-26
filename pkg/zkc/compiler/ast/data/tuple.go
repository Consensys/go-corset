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

import (
	"fmt"
	"strings"
)

// Tuple represents the composition of zero or more types together.  For
// example, (u8,u16) is a tuple type consisting of two elements: the first being
// a u8, and the second being a u16.  The overall width of the tuple is
// therefore the sum of the widths of the elements.
type Tuple struct {
	elements []Type
}

// NewTuple constructs a new tuple type.
func NewTuple(elements ...Type) *Tuple {
	return &Tuple{elements}
}

// BitWidth implementation for Type interface
func (p *Tuple) BitWidth() uint {
	var sum uint
	//
	for _, element := range p.elements {
		sum += element.BitWidth()
	}
	//
	return sum
}

// Flattern implementation for Type interface
func (p *Tuple) Flattern(prefix string, constructor func(name string, bitwidth uint)) {
	for i, element := range p.elements {
		ith := fmt.Sprintf("%s$%d", prefix, i)
		element.Flattern(ith, constructor)
	}
}

func (p *Tuple) String() string {
	var builder strings.Builder
	//
	builder.WriteString("(")
	//
	for i, element := range p.elements {
		if i != 0 {
			builder.WriteString(",")
		}
		//
		builder.WriteString(element.String())
	}
	//
	builder.WriteString(")")
	//
	return builder.String()
}
