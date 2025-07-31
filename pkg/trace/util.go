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
	"fmt"

	bls12_377 "github.com/consensys/go-corset/pkg/util/field/bls12-377"
)

// QualifiedColumnName returns the fully qualified name of a given column.
func QualifiedColumnName(module string, column string) string {
	if module == "" {
		return column
	}

	return fmt.Sprintf("%s.%s", module, column)
}

// NumberOfColumns returns the total number of all columns in the given trace.
func NumberOfColumns(tr Trace[bls12_377.Element]) uint {
	var count = uint(0)
	//
	for i := range tr.Width() {
		ith := tr.Module(i)
		count += ith.Width()
	}
	//
	return count
}
