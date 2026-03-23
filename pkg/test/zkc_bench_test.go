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

	"github.com/consensys/go-corset/pkg/test/util"
)

// ===================================================================
// Benchmark Tests
// ===================================================================

func Test_ZkcBench_Fnv1aHash(t *testing.T) {
	checkZkcBench(t, "zkc/bench/fnv1a_hash")
}

func Test_ZkcBench_Sort(t *testing.T) {
	checkZkcBench(t, "zkc/bench/sort")
}

func Test_ZkcBench_SgnExtend(t *testing.T) {
	checkZkcBench(t, "zkc/bench/sgn_extension_u32_u64")
}

func Test_ZkcBench_Lo32(t *testing.T) {
	checkZkcBench(t, "zkc/bench/lo_32")
}

func Test_ZkcBench_Hi32(t *testing.T) {
	checkZkcBench(t, "zkc/bench/hi_32")
}

func Test_ZkcBench_Mul32(t *testing.T) {
	checkZkcBench(t, "zkc/bench/mul_32")
}

func Test_ZkcBench_Mulh32(t *testing.T) {
	checkZkcBench(t, "zkc/bench/mulh_32")
}

func Test_ZkcBench_Mulhu32(t *testing.T) {
	checkZkcBench(t, "zkc/bench/mulhu_32")
}

func Test_ZkcBench_Mulhsu32(t *testing.T) {
	checkZkcBench(t, "zkc/bench/mulhsu_32")
}

func Test_ZkcBench_LongDivisionU32(t *testing.T) {
	checkZkcBench(t, "zkc/bench/long_division_u32")
}

func Test_ZkcBench_DivuRemu32(t *testing.T) {
	checkZkcBench(t, "zkc/bench/long_division_u32")
	// we use the same testing file as the one for
	// LongDivisionU32 and check if we want q or r
}

func Test_ZkcBench_DivRem32(t *testing.T) {
	checkZkcBench(t, "zkc/bench/div_rem_32")
}

// ===================================================================
// Test Helpers
// ===================================================================

func checkZkcBench(t *testing.T, test string) {
	util.CheckValid(t, test, "zkc", util.CompileZkc)
}
