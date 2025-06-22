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
package micro

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/consensys/go-corset/pkg/asm/io"
	"github.com/consensys/go-corset/pkg/schema"
)

func assignmentToString(dsts []io.RegisterId, srcs []io.RegisterId, constant big.Int, fn schema.Module,
	c big.Int, op string) string {
	//
	var (
		builder strings.Builder
		regs    = fn.Registers()
	)
	//
	builder.WriteString(io.RegistersReversedToString(dsts, regs))
	builder.WriteString(" = ")
	//
	for i, rid := range srcs {
		r := rid.Unwrap()
		//
		if i != 0 {
			builder.WriteString(op)
		}
		//
		if r < uint(len(regs)) {
			builder.WriteString(regs[r].Name)
		} else {
			builder.WriteString(fmt.Sprintf("?%d", r))
		}
	}
	//
	if len(srcs) == 0 || constant.Cmp(&c) != 0 {
		if len(srcs) > 0 {
			builder.WriteString(op)
		}
		//
		builder.WriteString("0x")
		builder.WriteString(constant.Text(16))
	}
	//
	return builder.String()
}
