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

// ASM_MAX_PADDING determines the maximum amount of padding to use when testing.
// Specifically, every trace is tested with varying amounts of padding upto this
// value.  NOTE: assembly modules don't need to be tested for higher padding
// values, since they only ever do unit shifts.
const ASM_MAX_PADDING uint = 2

func Test_AsmBench_Add(t *testing.T) {
	test_util.CheckWithFields(t, false, "asm/bench/add", ASM_MAX_PADDING, sc.BLS12_377, sc.KOALABEAR_16)
}

func Test_AsmBench_Exp(t *testing.T) {
	test_util.CheckWithFields(t, false, "asm/bench/exp", ASM_MAX_PADDING, sc.BLS12_377)
}

func Test_AsmBench_Gas(t *testing.T) {
	test_util.CheckWithFields(t, false, "asm/bench/gas", ASM_MAX_PADDING, sc.BLS12_377, sc.KOALABEAR_16)
}

func Test_AsmBench_Shf(t *testing.T) {
	test_util.Check(t, false, "asm/bench/shf")
}

func Test_AsmBench_Stp(t *testing.T) {
	test_util.Check(t, false, "asm/bench/stp")
}

func Test_AsmBench_Trm(t *testing.T) {
	test_util.CheckWithFields(t, false, "asm/bench/trm", ASM_MAX_PADDING, sc.BLS12_377, sc.KOALABEAR_16)
}

// Field Element Out-Of-Bounds
func Test_AsmBench_Wcp(t *testing.T) {
	test_util.CheckWithFields(t, false, "asm/bench/wcp", ASM_MAX_PADDING, sc.BLS12_377, sc.KOALABEAR_16)
}
