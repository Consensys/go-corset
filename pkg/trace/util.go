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
	"strings"
)

// MaxHeight determines the maximum height of any column in the trace.  This is
// useful in some scenarios for bounding the number of rows for any column.
// This is done by computing the maximum height of any module.
func MaxHeight(tr Trace) uint {
	h := uint(0)
	// Iterate over modules
	for i := uint(0); i < tr.Width(); i++ {
		ctx := tr.Column(i).Context()
		h = max(h, tr.Height(ctx))
	}
	// Done
	return h
}

// QualifiedColumnNamesToCommaSeparatedString produces a suitable string for use
// in error messages from a list of one or more column identifies.
func QualifiedColumnNamesToCommaSeparatedString(columns []uint, trace Trace) string {
	var names strings.Builder

	for i, c := range columns {
		if i != 0 {
			names.WriteString(",")
		}

		names.WriteString(trace.Column(c).Name())
	}
	// Done
	return names.String()
}

// QualifiedColumnName returns the fully qualified name of a given column.
func QualifiedColumnName(module string, column string) string {
	if module == "" {
		return column
	}

	return fmt.Sprintf("%s.%s", module, column)
}
