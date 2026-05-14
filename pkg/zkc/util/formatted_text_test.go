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
package util

import (
	"math/big"
	"testing"
)

type bigWord big.Int

func (p *bigWord) Text(base int) string {
	return (*big.Int)(p).Text(base)
}

func newBigWord(n int64) *bigWord {
	return (*bigWord)(big.NewInt(n))
}

func TestFormatString(t *testing.T) {
	tests := []struct {
		format   Format
		expected string
	}{
		{DecimalFormat(), "%d"},
		{HexFormat(), "%x"},
		{BinFormat(), "%b"},
		{CharFormat(), "%c"},
		{Format{Code: FORMAT_HEX, Width: 8, ZeroPad: true}, "%08x"},
		{Format{Code: FORMAT_HEX, Width: 8}, "%8x"},
		{Format{Code: FORMAT_DEC, Width: 4, ZeroPad: true}, "%04d"},
		{Format{Code: FORMAT_DEC, Width: 4}, "%4d"},
		{Format{Code: FORMAT_BIN, Width: 16, ZeroPad: true}, "%016b"},
	}
	//
	for _, tc := range tests {
		if got := tc.format.String(); got != tc.expected {
			t.Errorf("Format.String() = %q, want %q", got, tc.expected)
		}
	}
}

func TestFormatWord(t *testing.T) {
	tests := []struct {
		format   Format
		value    int64
		expected string
	}{
		// No padding.
		{HexFormat(), 0x42, "42"},
		{DecimalFormat(), 42, "42"},
		{BinFormat(), 5, "101"},
		// Zero padding for %08x.
		{Format{Code: FORMAT_HEX, Width: 8, ZeroPad: true}, 0x42, "00000042"},
		{Format{Code: FORMAT_HEX, Width: 8, ZeroPad: true}, 0xabcdef01, "abcdef01"},
		// Space padding for %8x.
		{Format{Code: FORMAT_HEX, Width: 8}, 0x42, "      42"},
		// Width smaller than digits: leave unchanged.
		{Format{Code: FORMAT_HEX, Width: 2, ZeroPad: true}, 0xabcdef, "abcdef"},
		// Decimal padding.
		{Format{Code: FORMAT_DEC, Width: 4, ZeroPad: true}, 42, "0042"},
		{Format{Code: FORMAT_DEC, Width: 4}, 42, "  42"},
		// Binary padding.
		{Format{Code: FORMAT_BIN, Width: 8, ZeroPad: true}, 5, "00000101"},
		// Character: 'A' (0x41), 'a' (0x61), digit '7' (0x37).
		{CharFormat(), 0x41, "A"},
		{CharFormat(), 0x61, "a"},
		{CharFormat(), 0x37, "7"},
		// Non-printable bytes still round-trip exactly.
		{CharFormat(), 0x0a, "\n"},
	}
	//
	for _, tc := range tests {
		if got := FormatWord(tc.format, newBigWord(tc.value)); got != tc.expected {
			t.Errorf("FormatWord(%s, %d) = %q, want %q", tc.format.String(), tc.value, got, tc.expected)
		}
	}
}
