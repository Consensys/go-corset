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
	"bytes"
	"fmt"
	"os"

	"github.com/consensys/go-corset/pkg/util/source"
	"github.com/consensys/go-corset/pkg/zkc/compiler/format"
	"github.com/consensys/go-corset/pkg/zkc/compiler/parser"
	"github.com/spf13/cobra"
)

var formatCmd = &cobra.Command{
	Use:     "format [flags] file1.zkc file2.zkc ...",
	Aliases: []string{"fmt"},
	Short:   "Format zkc source files.",
	Long:    `Format (pretty-print) a given set of zkc source file(s) in a canonical style.`,
	Run: func(cmd *cobra.Command, args []string) {
		check := GetFlag(cmd, "check")
		different := false

		for _, filename := range args {
			if runFormatFile(filename, check) {
				different = true
			}
		}

		if check && different {
			os.Exit(1)
		}
	},
}

// runFormatFile formats a single file, returning true if the file differs from
// the formatted output (relevant only when check=true).
func runFormatFile(filename string, check bool) bool {
	original, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println(err)
		os.Exit(3)
	}

	src := source.NewSourceFile(filename, original)
	file, errs := parser.Parse(src)

	if len(errs) > 0 {
		for _, e := range errs {
			printSyntaxError(&e)
		}

		os.Exit(4)
	}

	var buf bytes.Buffer

	if err := format.Format(&buf, file, *src); err != nil {
		fmt.Println(err)
		os.Exit(3)
	}

	formatted := buf.Bytes()

	if check {
		if !bytes.Equal(original, formatted) {
			fmt.Println(filename)

			return true
		}

		return false
	}

	if err := os.WriteFile(filename, formatted, 0600); err != nil {
		fmt.Println(err)
		os.Exit(3)
	}

	return false
}

// ============================================================================
// Misc
// ============================================================================

//nolint:errcheck
func init() {
	rootCmd.AddCommand(formatCmd)
	formatCmd.Flags().Bool("check", false, "report files that differ without rewriting")
}
