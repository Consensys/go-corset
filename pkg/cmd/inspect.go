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
	"math"
	"os"

	"github.com/consensys/go-corset/pkg/asm"
	"github.com/consensys/go-corset/pkg/binfile"
	"github.com/consensys/go-corset/pkg/cmd/inspector"
	"github.com/consensys/go-corset/pkg/cmd/view"
	"github.com/consensys/go-corset/pkg/corset"
	"github.com/consensys/go-corset/pkg/schema/module"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/field/bls12_377"
	"github.com/consensys/go-corset/pkg/util/field/gf251"
	"github.com/consensys/go-corset/pkg/util/field/gf8209"
	"github.com/consensys/go-corset/pkg/util/field/koalabear"
	"github.com/consensys/go-corset/pkg/util/termio"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var inspectCmd = &cobra.Command{
	Use:   "inspect [flags] trace_file constraint_file(s)",
	Short: "Inspect a trace file",
	Long:  `Inspect a trace file using an interactive (terminal-based) environment`,
	Run: func(cmd *cobra.Command, args []string) {
		runFieldAgnosticCmd(cmd, args, inspectCmds)
	},
}

// Available instances
var inspectCmds = []FieldAgnosticCmd{
	{field.GF_251, runInspectCmd[gf251.Element]},
	{field.GF_8209, runInspectCmd[gf8209.Element]},
	{field.KOALABEAR_16, runInspectCmd[koalabear.Element]},
	{field.BLS12_377, runInspectCmd[bls12_377.Element]},
}

func runInspectCmd[F field.Element[F]](cmd *cobra.Command, args []string) {
	var (
		errors []error
		trace  tr.Trace[F]
	)
	//
	if len(args) != 2 {
		fmt.Println(cmd.UsageString())
		os.Exit(1)
	}
	// Configure log level
	if GetFlag(cmd, "verbose") {
		log.SetLevel(log.DebugLevel)
	}
	// Read in constraint files
	stacker := *getSchemaStack[F](cmd, SCHEMA_DEFAULT_AIR, args[1:]...)
	stack := stacker.Build()
	//
	stats := util.NewPerfStats()
	// Parse constraints
	binf := stacker.BinaryFile()
	// Determine whether expansion is being performed
	expanding := stack.TraceBuilder().Expanding()
	// Sanity check debug information is available.
	srcmap, srcmap_ok := binfile.GetAttribute[*corset.SourceMap](binf)
	//
	if !srcmap_ok {
		fmt.Printf("binary file \"%s\" missing source map", args[1])
	} else if !stack.HasUniqueSchema() {
		fmt.Println("must specify exactly one of --air/mir/uasm/asm")
		os.Exit(2)
	}
	//
	stats.Log("Reading constraints file")
	// Parse trace file
	tracefile := ReadTraceFile(args[0])
	// Extract scheam
	schema := stack.UniqueConcreteSchema()
	//
	stats.Log("Reading trace file")
	//
	if expanding {
		// Apply trace propagation
		tracefile, errors = asm.Propagate(binf.Schema, tracefile, true)
	}
	// Apply trace expansion
	if len(errors) == 0 {
		trace, errors = stack.TraceBuilder().Build(schema, tracefile)
	}
	//
	if len(errors) == 0 {
		// Run the inspector.
		errors = inspect(stack.TraceBuilder().Mapping(), srcmap, trace)
	}
	// Sanity check what happened
	if len(errors) > 0 {
		for _, err := range errors {
			fmt.Println(err)
		}

		os.Exit(1)
	}
}

// Inspect a given trace using a given schema.
func inspect[F field.Element[F]](mapping module.LimbsMap, srcmap *corset.SourceMap, trace tr.Trace[F]) []error {
	// Construct inspector window
	inspector := construct(mapping, trace, srcmap)
	// Render inspector
	if err := inspector.Render(); err != nil {
		return []error{err}
	}
	//
	return inspector.Start()
}

func construct[F field.Element[F]](mapping module.LimbsMap, trace tr.Trace[F], srcmap *corset.SourceMap,
) *inspector.Inspector {
	//
	term, err := termio.NewTerminal()
	// Check whether successful
	if err == nil {
		window := view.NewBuilder[F](mapping).
			WithSourceMap(*srcmap).
			WithTitleWidth(math.MaxUint).
			WithFormatting(inspector.NewFormatter()).
			Build(trace)
		// Construct inspector state
		return inspector.NewInspector(term, window)
	}

	fmt.Println(error.Error(err))
	os.Exit(1)
	// Unreachable
	return nil
}

//nolint:errcheck
func init() {
	rootCmd.AddCommand(inspectCmd)
}
