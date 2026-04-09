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
	"strings"

	"github.com/consensys/go-corset/pkg/zkc/vm/word"
)

// EscapeFormattedText takes a string and escapes any characters which need to
// be escaped in order to be printed.
func EscapeFormattedText(text string) string {
	text = strings.ReplaceAll(text, "\\", "\\\\")
	text = strings.ReplaceAll(text, "\n", "\\n")
	text = strings.ReplaceAll(text, "\t", "\\t")
	//
	return strings.ReplaceAll(text, "\r", "\\r")
}

const (
	// FORMAT_NONE indicates an empty format string
	FORMAT_NONE uint = iota
	// FORMAT_DEC indicates to format in decimal
	FORMAT_DEC
	// FORMAT_HEX indicates to format in hexadecimal
	FORMAT_HEX
	// FORMAT_BIN indicates to format in hexadecimal
	FORMAT_BIN
)

// Format simply encodes the set of permitted formatting strings in a printf
// statement, such as "%d", "%x", etc.
type Format struct {
	Code uint
}

// EMPTY_FORMAT indicates no formatted argument is required.
var EMPTY_FORMAT = Format{FORMAT_NONE}

// DecimalFormat constructs a new decimal format.
func DecimalFormat() Format {
	return Format{FORMAT_DEC}
}

// HexFormat constructs a new hexadecimal format.
func HexFormat() Format {
	return Format{FORMAT_HEX}
}

// BinFormat constructs a new hexadecimal format.
func BinFormat() Format {
	return Format{FORMAT_BIN}
}

// HasFormat checks whether this actually represents a format, or is empty.
func (p Format) HasFormat() bool {
	return p.Code != FORMAT_NONE
}

func (p Format) String() string {
	switch p.Code {
	case FORMAT_DEC:
		return "%d"
	case FORMAT_HEX:
		return "%x"
	case FORMAT_BIN:
		return "%b"
	}
	//
	panic("invalid format")
}

// FormatWord applies a given format to a given word to generate a formatted string.
func FormatWord[W word.Word[W]](format Format, word W) string {
	switch format.Code {
	case FORMAT_DEC:
		return word.Text(10)
	case FORMAT_HEX:
		return "0x" + word.Text(16)
	case FORMAT_BIN:
		return "0b" + word.Text(2)
	}
	//
	panic("invalid format")
}
