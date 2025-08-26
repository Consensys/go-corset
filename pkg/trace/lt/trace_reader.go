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
package lt

import (
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util/word"
)

// FromBytes parses a byte array representing a given LTv2 trace file into a set
// of columns, or produces an error if the original file was malformed in some
// way. The input represents the original legacy format of trace files (i.e.
// without any additional header information prepended, etc).
func FromBytes(data []byte) (WordHeap, []trace.RawColumn[word.BigEndian], error) {
	panic("todo")
}
