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
	"github.com/consensys/go-corset/pkg/util/source"
	"github.com/consensys/go-corset/pkg/zkc/compiler"
	"github.com/consensys/go-corset/pkg/zkc/compiler/parser"
	"go.lsp.dev/protocol"
)

// DefinitionFor compiles the given document and returns the source location of
// the top-level declaration named by the identifier under the cursor at pos. It
// returns nil when no definition is available (e.g. the cursor is on a keyword,
// the name is not a top-level declaration, or the document cannot be compiled).
func DefinitionFor(uri protocol.URI, text string, pos protocol.Position) ([]protocol.Location, error) {
	srcfile := source.NewSourceFile(uri.Filename(), []byte(text))
	program, srcmaps, _ := compiler.Compile(*srcfile)

	// Convert LSP cursor position to a rune offset in the source file.
	offset := posToOffset(*srcfile, pos)

	// Lex the document to find the identifier token under the cursor.
	tokens, _ := parser.Lex(*srcfile, false)

	tok, ok := tokenAtOffset(tokens, offset)
	if !ok || tok.Kind != parser.IDENTIFIER {
		return nil, nil
	}

	// Extract the identifier text.
	contents := srcfile.Contents()
	name := string(contents[tok.Span.Start():tok.Span.End()])

	// Search top-level declarations by name and return the first match.
	for _, d := range program.Components() {
		if d.Name() != name {
			continue
		}

		defFile, span, found := srcmaps.Lookup(d)
		if !found {
			continue
		}

		rng := spanToRange(defFile, span)
		defURI := protocol.URI("file://" + defFile.Filename())

		return []protocol.Location{{URI: defURI, Range: rng}}, nil
	}

	return nil, nil
}
