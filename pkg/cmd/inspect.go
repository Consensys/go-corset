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
package cmd

import (
	"fmt"
	"os"

	"github.com/consensys/go-corset/pkg/binfile"
	"github.com/consensys/go-corset/pkg/cmd/inspector"
	"github.com/consensys/go-corset/pkg/corset"
	sc "github.com/consensys/go-corset/pkg/schema"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/termio"
	"github.com/spf13/cobra"
)

var inspectCmd = &cobra.Command{
	Use:   "inspect [flags] trace_file constraint_file(s)",
	Short: "Inspect a trace file",
	Long:  `Inspect a trace file using an interactive (terminal-based) environment`,
	Run: func(cmd *cobra.Command, args []string) {
		var corsetConfig corset.CompilationConfig
		//
		if len(args) != 2 {
			fmt.Println(cmd.UsageString())
			os.Exit(1)
		}
		defensive := GetFlag(cmd, "defensive")
		//
		corsetConfig.Stdlib = !GetFlag(cmd, "no-stdlib")
		corsetConfig.Legacy = GetFlag(cmd, "legacy")
		corsetConfig.StrictMode = GetFlag(cmd, "strict")
		//
		stats := util.NewPerfStats()
		// Parse constraints
		binf := ReadConstraintFiles(corsetConfig, args[1:])
		// Sanity check debug information is available.
		srcmap, srcmap_ok := binfile.GetAttribute[*corset.SourceMap](binf)
		//
		if !srcmap_ok {
			fmt.Printf("binary file \"%s\" missing source map", args[1])
		}
		//
		stats.Log("Reading constraints file")
		// Parse trace file
		tracefile := ReadTraceFile(args[0])
		//
		stats.Log("Reading trace file")
		//
		builder := sc.NewTraceBuilder(&binf.Schema).Expand(true).Defensive(defensive).Parallel(true)
		//
		trace, errors := builder.Build(tracefile.Columns)
		//
		if len(errors) == 0 {
			// Run the inspector.
			errors = inspect(&binf.Schema, srcmap, trace)
		}
		// Sanity check what happened
		if len(errors) > 0 {
			for _, err := range errors {
				fmt.Println(err)
			}
			os.Exit(1)
		}
	},
}

// Inspect a given trace using a given schema.
func inspect(schema sc.Schema, srcmap *corset.SourceMap, trace tr.Trace) []error {
	// Construct inspector window
	inspector := construct(schema, trace, srcmap)
	// Render inspector
	if err := inspector.Render(); err != nil {
		return []error{err}
	}
	//
	return inspector.Start()
}

func construct(schema sc.Schema, trace tr.Trace, srcmap *corset.SourceMap) *inspector.Inspector {
	term, err := termio.NewTerminal()
	// Check whether successful
	if err != nil {
		fmt.Println(error.Error(err))
		os.Exit(1)
	}
	// Construct inspector state
	return inspector.NewInspector(term, schema, trace, srcmap)
}

//nolint:errcheck
func init() {
	rootCmd.AddCommand(inspectCmd)
	inspectCmd.Flags().Bool("defensive", true, "enable / disable defensive padding")
}
