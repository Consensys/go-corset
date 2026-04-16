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
package compiler_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/consensys/go-corset/pkg/util/source"
	"github.com/consensys/go-corset/pkg/zkc/compiler"
)

// TestGlobInclude_Match verifies that a glob pattern in an include directive
// loads all matching files and compiles successfully.
func TestGlobInclude_Match(t *testing.T) {
	dir := t.TempDir()
	subdir := filepath.Join(dir, "sub")

	writeFile(t, filepath.Join(subdir, "a.zkc"), `fn a(x:u8) -> (r:u8) { r = x }`)
	writeFile(t, filepath.Join(subdir, "b.zkc"), `fn b(x:u8) -> (r:u8) { r = x }`)

	// Both a and b must be loaded for this to compile.
	mainContent := "include \"sub/*.zkc\"\nfn main(x:u8) -> (r:u8) { r = a(b(x)) }"
	mainPath := filepath.Join(dir, "main.zkc")
	writeFile(t, mainPath, mainContent)

	sf := *source.NewSourceFile(mainPath, []byte(mainContent))
	_, _, errs := compiler.Compile(sf)

	if len(errs) != 0 {
		t.Fatalf("expected no errors, got %d: %v", len(errs), errs)
	}
}

// TestGlobInclude_NoMatch verifies that a glob pattern that matches no files
// is reported as a syntax error.
func TestGlobInclude_NoMatch(t *testing.T) {
	dir := t.TempDir()

	mainContent := `include "nodir/*.zkc"`
	mainPath := filepath.Join(dir, "main.zkc")
	writeFile(t, mainPath, mainContent)

	sf := *source.NewSourceFile(mainPath, []byte(mainContent))
	_, _, errs := compiler.Compile(sf)

	requireErrorContaining(t, errs, "no files matched glob pattern")
}

// TestGlobInclude_BadPattern verifies that a malformed glob pattern is
// reported as a syntax error.
func TestGlobInclude_BadPattern(t *testing.T) {
	dir := t.TempDir()

	// An unclosed '[' is a malformed glob in Go's filepath.Glob.
	mainContent := `include "[invalid*.zkc"`
	mainPath := filepath.Join(dir, "main.zkc")
	writeFile(t, mainPath, mainContent)

	sf := *source.NewSourceFile(mainPath, []byte(mainContent))
	_, _, errs := compiler.Compile(sf)

	requireErrorContaining(t, errs, "syntax error in pattern")
}

// TestGlobInclude_Dedup verifies that a file already loaded by an explicit
// include is not loaded a second time when it is also matched by a subsequent
// glob pattern.
func TestGlobInclude_Dedup(t *testing.T) {
	dir := t.TempDir()
	subdir := filepath.Join(dir, "sub")

	// helper.zkc defines fn helper — loading it twice would cause a
	// "duplicate declaration" error, so a single error-free compile proves
	// deduplication is working.
	writeFile(t, filepath.Join(subdir, "helper.zkc"), `fn helper(x:u8) -> (r:u8) { r = x }`)

	mainContent := "include \"sub/helper.zkc\"\ninclude \"sub/*.zkc\"\nfn main(x:u8) -> (r:u8) { r = helper(x) }"
	mainPath := filepath.Join(dir, "main.zkc")
	writeFile(t, mainPath, mainContent)

	sf := *source.NewSourceFile(mainPath, []byte(mainContent))
	_, _, errs := compiler.Compile(sf)

	if len(errs) != 0 {
		t.Fatalf("expected no errors (dedup should prevent double-loading), got %d: %v", len(errs), errs)
	}
}

// ============================================================================
// Helpers
// ============================================================================

func writeFile(t *testing.T, path, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
}

func requireErrorContaining(t *testing.T, errs []source.SyntaxError, substr string) {
	t.Helper()

	if len(errs) == 0 {
		t.Fatalf("expected an error containing %q, got none", substr)
	}

	for _, e := range errs {
		if strings.Contains(e.Message(), substr) {
			return
		}
	}

	t.Fatalf("expected an error containing %q; actual errors: %v", substr, errs)
}
