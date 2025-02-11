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
package binfile

import (
	"strconv"
	"strings"
)

// ============================================================================
// Column Ref
// ============================================================================

// Handle represents a module / column naming pair.
type Handle struct {
	module string
	column string
}

func asHandle(handle string) Handle {
	split := strings.Split(handle, ".")
	//
	if split[0] == "<prelude>" {
		return Handle{"", split[1]}
	}
	// Easy
	return Handle{split[0], split[1]}
}

func asColumn(handle string) uint {
	split := strings.Split(handle, "#")
	column, err := strconv.Atoi(split[1])
	// Error check
	if err != nil {
		panic(err.Error())
	}

	return uint(column)
}

func asColumns(handles []string) []uint {
	cols := make([]uint, len(handles))
	for i := 0; i < len(cols); i++ {
		cols[i] = asColumn(handles[i])
	}

	return cols
}
