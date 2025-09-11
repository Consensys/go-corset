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
package test

import (
	"testing"

	test_util "github.com/consensys/go-corset/pkg/test/util"
)

func Test_Asm_Byte(t *testing.T) {
	test_util.Check(t, false, "asm/util/byte")
}
func Test_Asm_FillBytes(t *testing.T) {
	test_util.Check(t, false, "asm/util/fill_bytes")
}

func Test_Asm_Log2(t *testing.T) {
	test_util.Check(t, false, "asm/util/log2")
}

func Test_Asm_Log256(t *testing.T) {
	test_util.Check(t, false, "asm/util/log256")
}
func Test_Asm_Min(t *testing.T) {
	test_util.Check(t, false, "asm/util/min")
}

func Test_Asm_SetByte(t *testing.T) {
	test_util.Check(t, false, "asm/util/set_byte")
}
