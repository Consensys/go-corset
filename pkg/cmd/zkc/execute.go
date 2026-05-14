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
	"fmt"
	"os"

	"github.com/consensys/go-corset/pkg/cmd/corset"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/trace/lt"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/field/bls12_377"
	"github.com/consensys/go-corset/pkg/util/field/gf251"
	"github.com/consensys/go-corset/pkg/util/field/gf8209"
	"github.com/consensys/go-corset/pkg/util/field/koalabear"
	"github.com/consensys/go-corset/pkg/zkc/compiler/codegen"
	"github.com/consensys/go-corset/pkg/zkc/constraints"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var executeCmd = &cobra.Command{
	Use:     "execute [flags] input.json file1.zkc file2.zkc ...",
	Short:   "Execute a zkc program.",
	Long:    `Execute a zkc program to produce a set of outputs a from given a set of inputs.`,
	Aliases: []string{"exec"},
	Run: func(cmd *cobra.Command, args []string) {
		runFieldAgnosticCmd(cmd, args, executeCmds)
	},
}

// Available instances
var executeCmds = []FieldAgnosticCmd{
	{field.GF_251, runExecuteCmd[gf251.Element]},
	{field.GF_8209, runExecuteCmd[gf8209.Element]},
	{field.KOALABEAR_16, runExecuteCmd[koalabear.Element]},
	{field.BLS12_377, runExecuteCmd[bls12_377.Element]},
}

func runExecuteCmd[F field.Element[F]](cmd *cobra.Command, args []string, field field.Config) {
	var (
		errors []error
		// compiler buildConfig
		buildConfig = codegen.DEFAULT_CONFIG.
				Vectorize(GetFlag(cmd, "vectorize")).
				Field(field)
		//
		traceConfig = constraints.DEFAULT_TRACE_CONFIG
		// outputFile file for trace
		outputFile = GetString(cmd, "output")
		// check constraints
		check = GetFlag(cmd, "check")
		// identify whether tracing required or not.
		tracing = check || outputFile != ""
		//
		trace   trace.Trace[F]
		outputs map[string][]byte
	)
	//
	input := ParseInputFile(args[0])
	// Compile source files, or print errors
	binfile := BuildSourceFiles[F](buildConfig, field, args[1:]...)
	// =====================================================
	// Trace / Execute
	// =====================================================
	if tracing {
		trace, errors = binfile.Trace(input, traceConfig)
	} else {
		outputs, errors = binfile.Execute(input, 1024)
	}
	// =====================================================
	// Generate output
	// =====================================================
	if outputFile == "" {
		for name, bytes := range outputs {
			fmt.Printf("%s = 0x%s\n", name, hex.EncodeToString(bytes))
		}
	} else if outputFile != "" {
		// Construct trace file
		ltf := lt.FromRawTrace(nil, trace)
		// Write out trace file
		WriteTraceFile(outputFile, ltf)
	}
	// =====================================================
	// Check Constraints
	// =====================================================
	if check {
		checkConstraints(binfile, trace, traceConfig)
	}
	// =====================================================
	// Report Execution Failures
	// =====================================================
	if len(errors) > 0 {
		// Log errors
		for _, e := range errors {
			log.Error(fmt.Sprintf("%s (IR)", e))
		}
		//
		os.Exit(4)
	}
}

func checkConstraints[F field.Element[F]](binfile *constraints.BinaryFile[F], tr trace.Trace[F],
	cfg constraints.TraceConfig) {
	//
	var checkConfig corset.CheckConfig
	// Set sensible defaults (for now)
	checkConfig.Report = true
	checkConfig.ReportCellWidth = 32
	checkConfig.ReportTitleWidth = 40
	checkConfig.ReportPadding = 2
	checkConfig.ReportLimbs = true
	checkConfig.ReportComputed = true
	checkConfig.AnsiEscapes = true
	// Construct limbs map
	mapping := binfile.LimbsMap()
	// Run the check
	if failures := binfile.Check(tr, cfg); len(failures) > 0 {
		corset.ReportFailures("AIR", failures, tr, mapping, checkConfig)
	}
}

// ============================================================================
// Misc
// ============================================================================

//nolint:errcheck
func init() {
	rootCmd.AddCommand(executeCmd)
	executeCmd.Flags().StringP("output", "o", "", "specify output file for writing trace")
	executeCmd.Flags().BoolP("check", "c", false, "check generated trace against constraints")
}
