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
package air

import "github.com/consensys/go-corset/pkg/schema"

// Constraint captures the essence of a constraint at the AIR level.
type Constraint interface {
	schema.Constraint
	// Air marks the constraints as been valid for the AIR representation.
	Air()
}

// Module captures the essence of a module at the AIR level.  Specifically, it
// is limited to only those constraint forms permitted at the AIR level.
type Module = schema.Table[Constraint]

// Schema captures the essence of an arithmetisation at the AIR level.
// Specifically, it is limited to only those constraint forms permitted at the
// AIR level.
type Schema = schema.UniformSchema[Module]
