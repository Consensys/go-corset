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

	sc "github.com/consensys/go-corset/pkg/schema"
	test_util "github.com/consensys/go-corset/pkg/test/util"
)

func Test_Asm_Add(t *testing.T) {
	test_util.CheckWithFields(t, false, "asm/bench/add", sc.BLS12_377, sc.KOALABEAR_16)
}

// func Test_Asm_Exp(t *testing.T) {
// 	test_util.Check(t, false, "asm/bench/exp")
// }

func Test_Asm_Gas(t *testing.T) {
	test_util.CheckWithFields(t, false, "asm/bench/gas", sc.BLS12_377, sc.GF_8209, sc.GF_251)
}

// func Test_Asm_Shf(t *testing.T) {
// 	test_util.Check(t, false, "asm/bench/shf")
// }

func Test_Asm_Trm(t *testing.T) {
	test_util.CheckWithFields(t, false, "asm/bench/trm", sc.BLS12_377, sc.KOALABEAR_16)
}

// Field Element Out-Of-Bounds
func Test_Asm_Wcp(t *testing.T) {
	test_util.CheckWithFields(t, false, "asm/bench/wcp", sc.BLS12_377, sc.KOALABEAR_16)
}
