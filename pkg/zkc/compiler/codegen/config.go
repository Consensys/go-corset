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
package codegen

// DEFAULT_CONFIG is the configuration used when no overrides are supplied.
// Vectorisation is enabled, which matches the behaviour expected by the
// downstream prover; callers wanting to disable individual passes (for
// debugging, for example) should derive a custom Config via the chainable
// setters below.
var DEFAULT_CONFIG = Config{vectorize: true}

// Config captures the tunable aspects of the ZkC code generator.  Instances
// are immutable: each setter (e.g. Vectorize) returns a new Config rather
// than mutating the receiver, so a Config can be safely shared between
// concurrent compilations.
type Config struct {
	// vectorize controls whether the codegen pipeline runs the
	// instruction-vectorisation pass in pkg/zkc/compiler/codegen/vectorize.go.
	// Vectorisation merges sequences of micro-instructions that have no
	// register conflicts into single (vector) macro-instructions, allowing
	// the prover to handle them in one step.  Disabling it produces a less
	// compact program but leaves the macro instruction stream identical to
	// the codegen output, which is useful when debugging the codegen or
	// inspecting the un-merged IR.
	vectorize bool
}

// Vectorize returns a copy of this Config in which the vectorisation pass is
// either enabled (flag=true) or disabled (flag=false).  The receiver is left
// unchanged, so this can be chained: codegen.DEFAULT_CONFIG.Vectorize(false).
func (p Config) Vectorize(flag bool) Config {
	var q = p
	//
	q.vectorize = flag
	//
	return q
}
