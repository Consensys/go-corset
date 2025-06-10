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
)

// MaxHeight determines the maximum height of any column in the trace.  This is
// useful in some scenarios for bounding the number of rows for any column.
// This is done by computing the maximum height of any module.
func MaxHeight(tr Trace) uint {
	h := uint(0)
	// Iterate over modules
	for i := uint(0); i < tr.Width(); i++ {
		m := tr.Module(i)
		// Iterate over columns
		for c := uint(0); c < m.Width(); c++ {
			h = max(h, tr.Height(m.Column(c).Context()))
		}
	}
	// Done
	return h
}

// QualifiedColumnName returns the fully qualified name of a given column.
func QualifiedColumnName(module string, column string) string {
	if module == "" {
		return column
	}

	return fmt.Sprintf("%s.%s", module, column)
}
