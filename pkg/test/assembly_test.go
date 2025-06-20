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

import "testing"

// Recusion
//
// func Test_Asm_Byte(t *testing.T) {
// 	Check(t, false, "asm/byte")
// }

func Test_Asm_Counter(t *testing.T) {
	Check(t, false, "asm/counter")
}

func Test_Asm_FastPow(t *testing.T) {
	Check(t, false, "asm/fast_pow")
}

func Test_Asm_Max(t *testing.T) {
	Check(t, false, "asm/max")
}

// Recusion
//
// func Test_Asm_RecPow(t *testing.T) {
// 	Check(t, false, "asm/rec_pow")
// }

// Recusion
//
// func Test_Asm_Shift(t *testing.T) {
// 	Check(t, false, "asm/shift")
// }

func Test_Asm_SlowPow(t *testing.T) {
	Check(t, false, "asm/slow_pow")
}

// Field Element Out-Of-Bounds
//
// func Test_Asm_Wcp(t *testing.T) {
// 	Check(t, false, "asm/wcp")
// }
