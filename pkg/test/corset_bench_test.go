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

func Test_Bench_Counter(t *testing.T) {
	Check(t, true, "corset/bench/counter")
}

func Test_Bench_ByteDecomp(t *testing.T) {
	Check(t, true, "corset/bench/byte_decomposition")
}

func Test_Bench_BitDecomp(t *testing.T) {
	Check(t, true, "corset/bench/bit_decomposition")
}

func Test_Bench_BitShift(t *testing.T) {
	Check(t, true, "corset/bench/bit_shift")
}

func Test_Bench_ByteSorting(t *testing.T) {
	Check(t, true, "corset/bench/byte_sorting")
}

func Test_Bench_WordSorting(t *testing.T) {
	Check(t, true, "corset/bench/word_sorting")
}

func Test_Bench_Multiplier(t *testing.T) {
	Check(t, false, "corset/bench/multiplier")
}

func Test_Bench_Memory16u8(t *testing.T) {
	Check(t, true, "corset/bench/memory16u8")
}

func Test_Bench_Memory32u32(t *testing.T) {
	util.Check(t, true, "corset/bench/memory32u32")
}

func Test_Bench_Memory32u64(t *testing.T) {
	util.Check(t, true, "corset/bench/memory32u64")
}
func Test_Bench_Adder(t *testing.T) {
	Check(t, true, "corset/bench/adder")
}

func Test_Bench_Fields(t *testing.T) {
	Check(t, true, "corset/bench/fields")
}

func Test_Bench_Add(t *testing.T) {
	util.Check(t, true, "corset/bench/add")
}

func Test_Bench_BinStatic(t *testing.T) {
	Check(t, true, "corset/bench/bin-static")
}

func Test_Bench_Bin(t *testing.T) {
	util.Check(t, true, "corset/bench/bin")
}

func Test_Bench_Wcp(t *testing.T) {
	util.Check(t, true, "corset/bench/wcp")
}

func Test_Bench_Mxp(t *testing.T) {
	util.Check(t, true, "corset/bench/mxp")
}

func Test_Bench_Shf(t *testing.T) {
	Check(t, true, "corset/bench/shf")
}

func Test_Bench_Euc(t *testing.T) {
	Check(t, true, "corset/bench/euc")
}

func Test_Bench_Oob(t *testing.T) {
	util.Check(t, true, "corset/bench/oob")
}

func Test_Bench_Stp(t *testing.T) {
	Check(t, true, "corset/bench/stp")
}

func Test_Bench_Mmio(t *testing.T) {
	util.Check(t, true, "corset/bench/mmio")
}

func Test_Bench_Rom(t *testing.T) {
	Check(t, true, "corset/bench/rom")
}

func Test_Bench_Gas(t *testing.T) {
	Check(t, true, "corset/bench/gas")
}

func Test_Bench_Exp(t *testing.T) {
	Check(t, true, "corset/bench/exp")
}

func Test_Bench_Mul(t *testing.T) {
	Check(t, true, "corset/bench/mul")
}

func Test_Bench_Mod(t *testing.T) {
	Check(t, true, "corset/bench/mod")
}

//#834
// func Test_Bench_TicTacToe(t *testing.T) {
// 	util.Check(t, true, "corset/bench/tic_tac_toe")
// }
