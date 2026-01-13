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
package io

import "github.com/consensys/go-corset/pkg/schema/register"

// Unit defines a distinct entity within the system, such as a function, or a
// read-only memory or a static reference table.  Units contain registers, some
// of which may be marked as inputs/outputs and others as internal, etc.
type Unit[T Instruction] interface {
	register.Map

	// IsPublic determines whether or not this is an "externally visible" unit
	// or not.  The differences between internal and external units is small.
	// Specifically, internal units are not visible in the generated trace
	// interface; likewise, they are hidden by default in the inspector.
	IsPublic() bool

	// IsSynthetic units are generated during compilation, rather than being
	// provided by the user.  At this time, units can never be synthetic
	IsSynthetic() bool
}
