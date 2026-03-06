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
	"fmt"
	"strings"

	"github.com/consensys/go-corset/pkg/util/file"
	"github.com/consensys/go-corset/pkg/util/source"
	"github.com/consensys/go-corset/pkg/zkc/compiler"
	"github.com/consensys/go-corset/pkg/zkc/util"
)

// TestCase simply packages together a filename, a line number and the
// corresponding test on that line.  This is primiarly useful for error
// reporting when a test fails.
type TestCase struct {
	// name of enclosing file
	filename string
	// line in the file reprensented by this test
	line uint
	// indicates whether this test is expected to pass or fail.
	expected bool
	// input/output data of test
	data map[string][]byte
}

// CompileMachine compiles one or more zkc source files into a base machine for
// executing tests with.
func CompileMachine(srcfiles ...source.File) []source.SyntaxError {
	_, _, errors := compiler.Compile(srcfiles...)
	//
	return errors
}

// CompileZkc compiles a single zkc source file, potentially producing errors.
func CompileZkc(srcfile source.File) []source.SyntaxError {
	_, _, errors := compiler.Compile(srcfile)
	//
	return errors
}

// ReadTestsFile reads a file containing zero or more tests expressed as JSON,
// where each test is on a separate line.
func ReadTestsFile(cfg TestConfig, test string) []TestCase {
	var (
		// Construct test filename
		filename = fmt.Sprintf("%s/%s.%s", TestDir, test, cfg.extension)
		// Read input file
		lines = file.ReadInputFileAsLines(filename)
		tests []TestCase
	)
	// Read constraints line by line
	for i, line := range lines {
		// Parse input line as JSON
		if line != "" && !strings.HasPrefix(line, ";;") {
			// Read inputs / outputs
			inputs, err := util.ParseJsonInputFile([]byte(line))
			//
			if err != nil {
				msg := fmt.Sprintf("%s:%d: %s", filename, i+1, err)
				panic(msg)
			}

			tests = append(tests, TestCase{filename, uint(i + 1), cfg.expected, inputs})
		}
	}

	return tests
}
