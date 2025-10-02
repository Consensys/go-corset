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
package view

import "github.com/consensys/go-corset/pkg/util/field"

// TraceView provides an abstract view of a trace.  As such, it may not include
// all modules found in a trace; rather it may include some subset matching a
// predicate, or only those which are publically visible, etc.
type TraceView interface {
	// Determine number of modules in this view
	Width() uint
	// Access ith module view
	Module(uint) ModuleView
}

// ============================================================================
// Implementation
// ============================================================================

type traceView[F field.Element[F]] struct {
	modules []moduleView[F]
}

// Module implementation for TraceView interface
func (p *traceView[F]) Module(ith uint) ModuleView {
	return &p.modules[ith]
}

// Width implementation for TraceView interface
func (p *traceView[F]) Width() uint {
	return uint(len(p.modules))
}
