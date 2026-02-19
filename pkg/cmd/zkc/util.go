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
package zkc

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path"
	"strings"

	"github.com/consensys/go-corset/pkg/util/file"
	"github.com/consensys/go-corset/pkg/util/source"
	"github.com/consensys/go-corset/pkg/zkc/compiler"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast"
	log "github.com/sirupsen/logrus"
)

// ParseInputFile parses a given input file (which is currently assumed to be
// JSON).  An input file contains the input bytes for each ROM in the given
// program.
func ParseInputFile(filename string) map[string][]byte {
	// Read input file
	filename, fileBytes, err := file.ReadAndUncompress(filename)
	//
	if err == nil {
		ext := path.Ext(filename)
		//
		switch ext {
		case ".json":
			return parseJsonInputFile(fileBytes)
		default:
			err = fmt.Errorf("unknown trace file format: %s", ext)
		}
	}
	// Handle error
	fmt.Println(err)
	os.Exit(2)
	// unreachable
	return nil
}

func parseJsonInputFile(bytes []byte) map[string][]byte {
	var (
		rawData map[string]string
		data    map[string][]byte
		err     error
	)
	// Unmarshall data
	if err = json.Unmarshal(bytes, &rawData); err == nil {
		// Parse data
		data = make(map[string][]byte)
		// Initialise data
		for k, v := range rawData {
			if strings.HasPrefix(v, "0x") {
				data[k], err = hex.DecodeString(v[2:])
			} else {
				var val big.Int
				val.SetString(v, 10)
				data[k] = val.Bytes()
			}
			//
			if err != nil {
				break
			}
		}
	}
	//
	//
	if err == nil {
		return data
	}
	// Handle error
	fmt.Println(err)
	os.Exit(2)
	// unreachable
	return nil
}

// CompileSourceFiles accepts a set of source files and compiles them into a
// program.  This can result, for example, in one or more syntax errors, etc.
func CompileSourceFiles(filenames []string) ast.Program {
	//
	var (
		errors   []source.SyntaxError
		srcfiles = make([]source.File, len(filenames))
	)
	// Read each file
	for i, n := range filenames {
		log.Debug(fmt.Sprintf("including source file %s", n))
		// Read source file
		bytes, err := os.ReadFile(n)
		// Sanity check for errors
		if err != nil {
			fmt.Println(err)
			os.Exit(3)
		}
		//
		srcfiles[i] = *source.NewSourceFile(n, bytes)
	}
	// Compile source files
	macroProgram, _, errors := compiler.Compile(srcfiles...)
	// Check for errors
	if len(errors) != 0 {
		// Report errors
		for _, err := range errors {
			printSyntaxError(&err)
		}
		// Fail
		os.Exit(4)
	}
	// Done
	return macroProgram
}

// Print a syntax error with appropriate highlighting.
func printSyntaxError(err *source.SyntaxError) {
	span := err.Span()
	line := err.FirstEnclosingLine()
	lineOffset := span.Start() - line.Start()
	// Calculate length (ensures don't overflow line)
	length := min(line.Length()-lineOffset, span.Length())
	// Print error + line number
	fmt.Printf("%s:%d:%d-%d %s\n", err.SourceFile().Filename(),
		line.Number(), 1+lineOffset, 1+lineOffset+length, err.Message())
	// Print separator line
	fmt.Println()
	// Print line
	fmt.Println(line.String())
	// Print indent (todo: account for tabs)
	fmt.Print(strings.Repeat(" ", lineOffset))
	// Print highlight
	fmt.Println(strings.Repeat("^", length))
}
