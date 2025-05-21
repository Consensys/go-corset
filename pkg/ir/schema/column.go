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
package schema

import (
	"fmt"

	"github.com/consensys/go-corset/pkg/trace"
)

// ============================================================================

const (
	INPUT_COLUMN    = uint8(0)
	COMPUTED_COLUMN = uint8(1)
)

// Column represents a specific column in the schema that, ultimately, will
// correspond 1:1 with a column in the trace.
type Column struct {
	// Evaluation Context of this column.
	Context trace.Context
	// Returns the Name of this column
	Name string
	// Returns the expected type of data in this column
	DataType Type
	// Determines what type of column we have.
	Kind uint8
}

// NewColumn constructs a new column
func NewColumn(context trace.Context, name string, datatype Type, kind uint8) Column {
	return Column{context, name, datatype, kind}
}

// QualifiedName returns the fully qualified name of this column
func (p Column) QualifiedName(mod Module) string {
	if mod.Name() != "" {
		return fmt.Sprintf("%s:%s", mod.Name, p.Name)
	}
	//
	return p.Name
}

func (p Column) String() string {
	return fmt.Sprintf("%s:%s", p.Name, p.DataType.String())
}
