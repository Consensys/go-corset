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
	"testing"
)

// CheckValid checks that a given source file compiles without any errors.
// nolint
func CheckValid(t *testing.T, test, ext string, compiler ErrorCompiler) {
	var filename = fmt.Sprintf("%s/%s.%s", TestDir, test, ext)
	// Enable testing each trace in parallel
	t.Parallel()
	//
	srcfile := readSourceFile(t, filename)
	// Compile source file and expect no errors
	if errors := compiler(*srcfile); len(errors) > 0 {
		msg := fmt.Sprintf("Error %s should have compiled\n", srcfile.Filename())
		for _, err := range errors {
			msg = fmt.Sprintf("%s  %s\n", msg, errorToString(err))
		}
		t.Fatal(msg)
	}
}
