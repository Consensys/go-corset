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
	"math/big"
	"os"

	"github.com/consensys/go-corset/pkg/asm"
	cmd_util "github.com/consensys/go-corset/pkg/cmd/util"
	"github.com/consensys/go-corset/pkg/corset"
	"github.com/consensys/go-corset/pkg/ir/mir"
	"github.com/consensys/go-corset/pkg/ir/picus"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/field/bls12_377"
	"github.com/consensys/go-corset/pkg/util/field/gf251"
	"github.com/consensys/go-corset/pkg/util/field/gf8209"
	"github.com/consensys/go-corset/pkg/util/field/koalabear"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var genPicusCmd = &cobra.Command{
	Use:   "picus [flags]",
	Short: "compile constraints into Picus' constraint language.",
	Long:  `Compile a given set of constraint file(s) into a single program in PCL (Picus Constraint Language) for Picus to check determinism.`,
	Run: func(cmd *cobra.Command, args []string) {
		runFieldAgnosticCmd(cmd, args, picusCmds)
	},
}

// Available instances
var picusCmds = []FieldAgnosticCmd{
	{sc.GF_251, runGenPicusCmd[gf251.Element]},
	{sc.GF_8209, runGenPicusCmd[gf8209.Element]},
	{sc.KOALABEAR_16, runGenPicusCmd[koalabear.Element]},
	{sc.BLS12_377, runGenPicusCmd[bls12_377.Element]},
}

func modulusOf[F field.Element[F]]() *big.Int {
	var z F
	return z.Modulus()
}

// The `picus` command takes as input a constraint file and translates the constraints
// into PCL. The current translator only supports translating `mir` files so that is hardcoded
// but can be changed to handle other targets
func runGenPicusCmd[F field.Element[F]](cmd *cobra.Command, args []string) {
	// Configure log level
	if GetFlag(cmd, "verbose") {
		log.SetLevel(log.DebugLevel)
	}

	mirSchema := getMirSchema[F](cmd, args)
	picusProgram := picus.NewProgram[F](modulusOf[F]())
	picusLowering := mir.NewPicusTranslator(mirSchema, picusProgram)
	picusLowering.Translate()
	picusProgram.WriteTo(os.Stdout)
}

// Gets an MIR schema from the input files. Most of this is borrowed from `GetSchemaStack`
func getMirSchema[F field.Element[F]](cmd *cobra.Command, args []string) mir.Schema[F] {
	var (
		mirEnable    = GetFlag(cmd, "mir")
		field        = GetString(cmd, "field")
		asmConfig    asm.LoweringConfig
		corsetConfig corset.CompilationConfig
		asmProgram   asm.MixedMacroProgram[bls12_377.Element]
		uasmProgram  asm.MixedMicroProgram[bls12_377.Element]
		mirSchema    mir.Schema[F]
	)
	if !mirEnable {
		fmt.Fprintf(os.Stderr, "%v", fmt.Errorf("-picus expects -mir flag"))
		os.Exit(1)
	}
	fieldConfig := sc.GetFieldConfig(field)
	asmConfig.Vectorize = GetFlag(cmd, "vectorize")
	asmConfig.Field = *fieldConfig
	// Initial corset compilation configuration
	corsetConfig.Stdlib = !GetFlag(cmd, "no-stdlib")
	corsetConfig.Debug = GetFlag(cmd, "debug")
	corsetConfig.Legacy = GetFlag(cmd, "legacy")
	corsetConfig.EnforceTypes = GetFlag(cmd, "enforce-types")
	binFile := cmd_util.ReadConstraintFiles(corsetConfig, asmConfig, args)
	asmProgram = binFile.Schema
	// Lower to mixed micro schema
	uasmProgram = asm.LowerMixedMacroProgram(asmConfig.Vectorize, asmProgram)
	// Apply register splitting for field agnosticity
	mirSchema, _ = asm.Concretize[bls12_377.Element, F](*fieldConfig, uasmProgram)
	return mirSchema
}

//nolint:errcheck
func init() {
	rootCmd.AddCommand(genPicusCmd)
}
