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
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/symbol"
)

// SubtypeOf performs a subtype check, reporting whether or not t1 <: t2.
func SubtypeOf[S symbol.Symbol[S]](t1, t2 Type[S], env Environment[S]) bool {
	switch t1 := t1.(type) {
	case *UnsignedInt[S]:
		if t := t2.AsUint(); t != nil {
			return t1.Width() == t.Width() || (t1.IsOpen() && t1.Width() < t.Width())
		}
	}
	//
	return false
}
