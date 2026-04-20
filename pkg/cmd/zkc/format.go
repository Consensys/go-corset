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
	"github.com/spf13/cobra"
)

var formatCmd = &cobra.Command{
	Use:     "format [flags] file1.zkc file2.zkc ...",
	Aliases: []string{"fmt"},
	Short:   "Format zkc source files.",
	Long:    `Format (pretty-print) a given set of zkc source file(s) in a canonical style.`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: implement formatting
	},
}

// ============================================================================
// Misc
// ============================================================================

//nolint:errcheck
func init() {
	rootCmd.AddCommand(formatCmd)
}
