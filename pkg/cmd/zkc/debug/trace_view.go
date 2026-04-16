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
package debug

import "github.com/consensys/go-corset/pkg/util/termio"

// TraceView adapts a given execution trace to act as a TableSource provider.
type TraceView struct {
}

// ColumnWidth implementation for TableSource interface.
func (p *TraceView) ColumnWidth(col uint) uint {
	return 10
}

// Dimensions implementation for TableSource interface.
func (p *TraceView) Dimensions() (uint, uint) {
	return 5, 2
}

// CellAt implementation for TableSource interface.
func (p *TraceView) CellAt(col, row uint) termio.FormattedText {
	if col >= 5 || row >= 2 {
		return termio.NewText("")
	}
	//
	return termio.NewText("Test")
}
