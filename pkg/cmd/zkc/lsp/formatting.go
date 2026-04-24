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
package lsp

import (
	"bytes"
	"strings"

	"github.com/consensys/go-corset/pkg/util/source"
	"github.com/consensys/go-corset/pkg/zkc/compiler/format"
	"github.com/consensys/go-corset/pkg/zkc/compiler/parser"
	"go.lsp.dev/protocol"
)

// OnTypeFormattingFor returns a TextEdit correcting the indentation of the line
// at pos.Line after the user typed ch ("}" or "\n"). Returns nil when the line
// is already correctly indented or pos.Line is out of range.
func OnTypeFormattingFor(
	uri protocol.URI,
	text string,
	pos protocol.Position,
	ch string,
	opts protocol.FormattingOptions,
) ([]protocol.TextEdit, error) {
	src := source.NewSourceFile(uri.Filename(), []byte(text))

	offset := posToOffset(*src, pos)
	targetLevel := braceDepthAtOffset(*src, offset)
	indentStr := buildIndentString(opts, targetLevel)

	lines := src.Lines()

	lineIdx := int(pos.Line)
	if lineIdx >= len(lines) {
		// The target line is an empty line at the end of a newline-terminated file.
		// Lines() only creates an entry when there is content after the last '\n',
		// so this empty trailing line is not represented — but the position is valid.
		if indentStr == "" {
			return nil, nil
		}

		editRange := protocol.Range{
			Start: protocol.Position{Line: pos.Line, Character: 0},
			End:   protocol.Position{Line: pos.Line, Character: 0},
		}

		return []protocol.TextEdit{{Range: editRange, NewText: indentStr}}, nil
	}

	line := lines[lineIdx]
	lineText := line.String()
	existingLen := leadingWhitespaceLen(lineText)

	if existingLen == len(indentStr) && lineText[:existingLen] == indentStr {
		return nil, nil
	}

	editRange := protocol.Range{
		Start: protocol.Position{Line: pos.Line, Character: 0},
		End:   protocol.Position{Line: pos.Line, Character: uint32(existingLen)},
	}

	return []protocol.TextEdit{{Range: editRange, NewText: indentStr}}, nil
}

// FormattingFor formats the given document text and returns a single TextEdit
// that replaces the entire document with its canonical form. Returns nil (no
// edits) when the document has parse errors or is already correctly formatted.
func FormattingFor(uri protocol.URI, text string) ([]protocol.TextEdit, error) {
	var (
		// temporary buffer for writing output
		buf bytes.Buffer
		// source file representation
		src = source.NewSourceFile(uri.Filename(), []byte(text))
		// construct default formatter
		formatter, _ = format.NewFormatter(&buf, src)
	)
	// apply formatting
	if err := formatter.Format(); err != nil {
		return nil, err
	}

	formatted := buf.String()

	if formatted == text {
		return nil, nil
	}

	// Span covering the whole document; spanToRange handles coordinate encoding.
	wholeDoc := source.NewSpan(0, len(src.Contents()))
	docRange := spanToRange(*src, wholeDoc)

	return []protocol.TextEdit{{
		Range:   docRange,
		NewText: formatted,
	}}, nil
}

// braceDepthAtOffset lexes src (no whitespace, no comments) and returns the
// brace nesting depth counting only tokens whose Span.Start() < offset.
// When called with the post-insertion cursor offset this gives:
//   - For "}": depth after the "}" = correct indent level for the "}" line.
//   - For "\n": depth at the start of the new line = correct indent for that line.
func braceDepthAtOffset(src source.File, offset int) uint {
	var depth uint

	for _, tok := range parser.Lex(src, false, false) {
		if tok.Span.Start() >= offset {
			break
		}

		switch tok.Kind {
		case parser.LCURLY:
			depth++
		case parser.RCURLY:
			if depth > 0 {
				depth--
			}
		}
	}

	return depth
}

// buildIndentString returns the whitespace string for the given nesting level
// using the client's formatting options. Uses tabs when InsertSpaces is false;
// falls back to DEFAULT_INDENTATION spaces when TabSize is zero.
func buildIndentString(opts protocol.FormattingOptions, level uint) string {
	if level == 0 {
		return ""
	}

	if !opts.InsertSpaces {
		return strings.Repeat("\t", int(level))
	}

	tabSize := uint(opts.TabSize)
	if tabSize == 0 {
		tabSize = format.DEFAULT_INDENTATION
	}

	return strings.Repeat(" ", int(level*tabSize))
}

// leadingWhitespaceLen returns the number of leading space or tab runes in s.
func leadingWhitespaceLen(s string) int {
	for i, ch := range s {
		if ch != ' ' && ch != '\t' {
			return i
		}
	}

	return len(s)
}
