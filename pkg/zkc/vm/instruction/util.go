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
package instruction

import (
	"fmt"
	"strings"

	"github.com/consensys/go-corset/pkg/schema/register"
)

// RegistersToString returns a string representation for zero or more registers
// separated by a comma.
func registersToString(regs []register.Id, env register.Map) string {
	var builder strings.Builder
	//
	for i := 0; i < len(regs); i++ {
		var rid = regs[i]
		//
		if i != 0 {
			builder.WriteString(", ")
		}
		//
		builder.WriteString(env.Register(rid).Name())
	}
	//
	return builder.String()
}

func expressionToString[W any](op string, regs []register.Id, constant W, env register.Map) string {
	var builder strings.Builder
	//
	for i := 0; i < len(regs); i++ {
		var rid = regs[i]
		//
		builder.WriteString(env.Register(rid).Name())
		builder.WriteString("")
		builder.WriteString(op)
		builder.WriteString("")
	}
	//
	builder.WriteString(fmt.Sprintf("%v", constant))
	//
	return builder.String()
}
