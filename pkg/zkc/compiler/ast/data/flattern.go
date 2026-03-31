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

	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/symbol"
)

// Flattern this type into a set of one or more registers, using a given
// prefix.  For example, a variable "x [2]u8" is flatterned into "x$0 u8"
// and "x$1 u8", etc.   NOTE: should a typing cycle exist involving the given type,
// then this will enter an infinite loop.
func Flattern[S symbol.Symbol[S]](t Type[S], prefix string, env Environment[S],
	constructor func(name string, bitwidth uint)) {
	//
	switch t := t.(type) {
	case *UnsignedInt[S]:
		constructor(prefix, t.bitwidth)
	case *Alias[S]:
		Flattern(t.Resolve(env), prefix, env, constructor)
	case *FixedArray[S]:
		Flattern(t.Resolve(env), prefix, env, constructor)
	case *Tuple[S]:
		for i, element := range t.elements {
			ith := fmt.Sprintf("%s$%d", prefix, i)
			Flattern(element, ith, env, constructor)
		}
	default:
		//
		panic(fmt.Sprintf("unknown type encountered: %s", t.String(env)))
	}
}
