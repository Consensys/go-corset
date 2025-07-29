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

func Test_Bench_Counter(t *testing.T) {
	test_util.Check(t, true, "bench/counter")
}

func Test_Bench_ByteDecomp(t *testing.T) {
	test_util.Check(t, true, "bench/byte_decomposition")
}

func Test_Bench_BitDecomp(t *testing.T) {
	test_util.Check(t, true, "bench/bit_decomposition")
}

func Test_Bench_BitShift(t *testing.T) {
	test_util.Check(t, true, "bench/bit_shift")
}

func Test_Bench_ByteSorting(t *testing.T) {
	test_util.Check(t, true, "bench/byte_sorting")
}

func Test_Bench_WordSorting(t *testing.T) {
	test_util.Check(t, true, "bench/word_sorting")
}

func Test_Bench_Multiplier(t *testing.T) {
	test_util.Check(t, false, "bench/multiplier")
}

func Test_Bench_Memory(t *testing.T) {
	test_util.Check(t, true, "bench/memory")
}

func Test_Bench_Adder(t *testing.T) {
	test_util.Check(t, true, "bench/adder")
}

func Test_Bench_Fields(t *testing.T) {
	test_util.Check(t, true, "bench/fields")
}

func Test_Bench_Add(t *testing.T) {
	test_util.Check(t, true, "bench/add")
}

func Test_Bench_BinStatic(t *testing.T) {
	test_util.Check(t, true, "bench/bin-static")
}

func Test_Bench_Bin(t *testing.T) {
	test_util.Check(t, true, "bench/bin")
}

func Test_Bench_Wcp(t *testing.T) {
	test_util.Check(t, true, "bench/wcp")
}

func Test_Bench_Mxp(t *testing.T) {
	test_util.Check(t, true, "bench/mxp")
}

func Test_Bench_Shf(t *testing.T) {
	test_util.Check(t, true, "bench/shf")
}

func Test_Bench_Euc(t *testing.T) {
	test_util.Check(t, true, "bench/euc")
}

func Test_Bench_Oob(t *testing.T) {
	test_util.Check(t, true, "bench/oob")
}

func Test_Bench_Stp(t *testing.T) {
	test_util.Check(t, true, "bench/stp")
}

func Test_Bench_Mmio(t *testing.T) {
	test_util.Check(t, true, "bench/mmio")
}

// func Test_Bench_Rom(t *testing.T) {
// 	test_util.Check(t, true, "bench/rom")
// }

func Test_Bench_Gas(t *testing.T) {
	test_util.Check(t, true, "bench/gas")
}

func Test_Bench_Exp(t *testing.T) {
	test_util.Check(t, true, "bench/exp")
}

func Test_Bench_Mul(t *testing.T) {
	test_util.Check(t, true, "bench/mul")
}

func Test_Bench_Mod(t *testing.T) {
	test_util.Check(t, true, "bench/mod")
}

func Test_Bench_TicTacToe(t *testing.T) {
	test_util.Check(t, true, "bench/tic_tac_toe")
}
