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
)

func Test_Counter(t *testing.T) {
	Check(t, true, "bench/counter")
}

func Test_ByteDecomp(t *testing.T) {
	Check(t, true, "bench/byte_decomposition")
}

func Test_BitDecomp(t *testing.T) {
	Check(t, true, "bench/bit_decomposition")
}

func Test_ByteSorting(t *testing.T) {
	Check(t, true, "bench/byte_sorting")
}

func Test_WordSorting(t *testing.T) {
	Check(t, true, "bench/word_sorting")
}

func Test_Multiplier(t *testing.T) {
	Check(t, false, "multiplier")
}

func Test_Memory(t *testing.T) {
	Check(t, true, "bench/memory")
}

func Test_Adder(t *testing.T) {
	Check(t, true, "bench/adder")
}

func TestSlow_Fields(t *testing.T) {
	Check(t, true, "bench/fields")
}

func TestSlow_Add(t *testing.T) {
	Check(t, true, "bench/add")
}

func TestSlow_BinStatic(t *testing.T) {
	Check(t, true, "bench/bin-static")
}

func TestSlow_Bin(t *testing.T) {
	Check(t, true, "bench/bin")
}

func TestSlow_Wcp(t *testing.T) {
	Check(t, true, "bench/wcp")
}

func TestSlow_Mxp(t *testing.T) {
	Check(t, true, "bench/mxp")
}

func TestSlow_Shf(t *testing.T) {
	Check(t, true, "bench/shf")
}

func TestSlow_Euc(t *testing.T) {
	Check(t, true, "bench/euc")
}

func TestSlow_Oob(t *testing.T) {
	Check(t, true, "bench/oob")
}

func TestSlow_Stp(t *testing.T) {
	Check(t, true, "bench/stp")
}

func TestSlow_Mmio(t *testing.T) {
	Check(t, true, "bench/mmio")
}

// func TestSlow_Rom(t *testing.T) {
// 	Check(t, true, "bench/rom")
// }

func TestSlow_Gas(t *testing.T) {
	Check(t, true, "bench/gas")
}

func TestSlow_Exp(t *testing.T) {
	Check(t, true, "bench/exp")
}

func TestSlow_Mul(t *testing.T) {
	Check(t, true, "bench/mul")
}

func TestSlow_Mod(t *testing.T) {
	Check(t, true, "bench/mod")
}

func Test_TicTacToe(t *testing.T) {
	Check(t, true, "bench/tic_tac_toe")
}
