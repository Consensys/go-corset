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
package register

import "github.com/consensys/go-corset/pkg/trace"

// Ref abstracts a complete (i.e. global) register identifier.
type Ref = trace.ColumnRef

// NewRef constructs a new register reference from the given module and
// register identifiers.
func NewRef(mid trace.ModuleId, id Id) Ref {
	return trace.NewColumnRef(mid, id)
}
