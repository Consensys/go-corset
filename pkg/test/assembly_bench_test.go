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
	"github.com/consensys/go-corset/pkg/test/util"
)

func Test_AsmBench_Add(t *testing.T) {
	util.CheckWithFields(t, false, "asm/bench/add", util.ASM_MAX_PADDING, sc.BLS12_377, sc.KOALABEAR_16)
}

func Test_AsmBench_Exp(t *testing.T) {
	util.CheckWithFields(t, false, "asm/bench/exp", util.ASM_MAX_PADDING, sc.BLS12_377)
}

func Test_AsmBench_Gas(t *testing.T) {
	util.CheckWithFields(t, false, "asm/bench/gas", util.ASM_MAX_PADDING, sc.BLS12_377, sc.KOALABEAR_16)
}

func Test_AsmBench_Shf(t *testing.T) {
	util.Check(t, false, "asm/bench/shf")
}

func Test_AsmBench_Stp(t *testing.T) {
	util.Check(t, false, "asm/bench/stp")
}

func Test_AsmBench_Trm(t *testing.T) {
	util.CheckWithFields(t, false, "asm/bench/trm", util.ASM_MAX_PADDING, sc.BLS12_377)
}

func Test_AsmBench_Bin(t *testing.T) {
	util.CheckWithFields(t, false, "asm/bench/bin", util.ASM_MAX_PADDING, sc.BLS12_377, sc.KOALABEAR_16)
}

// Field Element Out-Of-Bounds
func Test_AsmBench_Wcp(t *testing.T) {
	util.CheckWithFields(t, false, "asm/bench/wcp", util.ASM_MAX_PADDING, sc.BLS12_377)
}
