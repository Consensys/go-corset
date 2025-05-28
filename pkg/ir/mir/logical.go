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
package mir

import "github.com/consensys/go-corset/pkg/ir"

// LogicalTerm represents the fundamental for logical expressions in the MIR
// representation.
type LogicalTerm interface {
	ir.LogicalTerm[LogicalTerm]
}

// Logical captures the notion of a logical expression at the MIR level.  This
// is really just for convenience more than anything.
type Logical = ir.Logical[LogicalTerm]

// BOTTOM represents the empty or unused logical expression.
var BOTTOM Logical
