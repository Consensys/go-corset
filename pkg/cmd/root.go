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
	"runtime/debug"

	"github.com/consensys/go-corset/pkg/asm"
	cmd_util "github.com/consensys/go-corset/pkg/cmd/util"
	"github.com/consensys/go-corset/pkg/corset"
	"github.com/consensys/go-corset/pkg/ir"
	"github.com/consensys/go-corset/pkg/ir/mir"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/util/field/bls12_377"
	"github.com/spf13/cobra"
)

// Version is filled when building with make, but *not* when installing via "go
// install".
var Version string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "go-corset",
	Short: "A compiler for the Corset language.",
	Long:  "A compiler (and general toolbox) for the Corset language.",
	Run: func(cmd *cobra.Command, args []string) {
		if GetFlag(cmd, "version") {
			fmt.Print("go-corset ")
			if Version != "" {
				// Built via "make"
				fmt.Printf("%s", Version)
			} else if info, ok := debug.ReadBuildInfo(); ok {
				// Built via "go install"
				fmt.Printf("%s", info.Main.Version)
			} else {
				// Unknown, perhaps "go run"
				fmt.Printf("(unknown version)")
			}
			fmt.Println()
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// SCHEMA_OPTIONAL indicates a schema is optional
const SCHEMA_OPTIONAL = uint(0)

// SCHEMA_DEFAULT_MIR indicates a schema must be indicated on the command line,
// and that the default is for the stack to be lowered to the MIR level.
const SCHEMA_DEFAULT_MIR = uint(1)

// SCHEMA_DEFAULT_AIR indicates a schema must be indicated on the command line,
// and that the default is for the stack to be lowered to the AIR level.
const SCHEMA_DEFAULT_AIR = uint(2)

func getSchemaStack(cmd *cobra.Command, mode uint, filenames ...string) *cmd_util.SchemaStack[bls12_377.Element] {
	var (
		schemaStack  cmd_util.SchemaStack[bls12_377.Element]
		corsetConfig corset.CompilationConfig
		asmConfig    asm.LoweringConfig
		field        = GetString(cmd, "field")
		mirEnable    = GetFlag(cmd, "mir")
		airEnable    = GetFlag(cmd, "air")
		asmEnable    = GetFlag(cmd, "asm")
		uasmEnable   = GetFlag(cmd, "uasm")
		optimisation = GetUint(cmd, "opt")
		externs      = GetStringArray(cmd, "set")
		//
		parallel  = !GetFlag(cmd, "sequential")
		batchSize = GetUint(cmd, "batch")
		defensive = GetFlag(cmd, "defensive")
		expand    = !GetFlag(cmd, "raw")
		validate  = GetFlag(cmd, "validate")
	)
	// Field configuration
	fieldConfig := schema.GetFieldConfig(field)
	// Sanity check
	if fieldConfig == nil {
		fmt.Printf("unknown prime field \"%s\"\n", field)
		os.Exit(3)
	}
	// Apply field overrides
	if cmd.Flags().Lookup("field-width").Changed {
		fieldConfig.FieldBandWidth = GetUint(cmd, "field-width")
	}

	if cmd.Flags().Lookup("register-width").Changed {
		fieldConfig.RegisterWidth = GetUint(cmd, "register-width")
	}
	// Initial corset compilation configuration
	corsetConfig.Stdlib = !GetFlag(cmd, "no-stdlib")
	corsetConfig.Debug = GetFlag(cmd, "debug")
	corsetConfig.Legacy = GetFlag(cmd, "legacy")
	// Assembly lowering config
	asmConfig.Vectorize = GetFlag(cmd, "vectorize")
	asmConfig.Field = *fieldConfig
	//
	// Sanity check MIR optimisation level
	if optimisation >= uint(len(mir.OPTIMISATION_LEVELS)) {
		fmt.Printf("invalid optimisation level %d\n", optimisation)
		os.Exit(2)
	}
	// If no IR was specified, set a default
	if !airEnable && !mirEnable && !uasmEnable && !asmEnable {
		switch mode {
		case SCHEMA_DEFAULT_MIR:
			mirEnable = true
		case SCHEMA_DEFAULT_AIR:
			airEnable = true
		}
	}
	// Construct trace builder
	builder := ir.NewTraceBuilder[bls12_377.Element]().
		WithValidation(validate).
		WithDefensivePadding(defensive).
		WithExpansion(expand).
		WithParallelism(parallel).
		WithBatchSize(batchSize)
	// Configure the stack
	schemaStack.WithAssemblyConfig(asmConfig)
	schemaStack.WithCorsetConfig(corsetConfig)
	schemaStack.WithOptimisationConfig(mir.OPTIMISATION_LEVELS[optimisation])
	schemaStack.WithConstantDefinitions(externs)
	//
	if asmEnable {
		schemaStack.WithLayer(cmd_util.MACRO_ASM_LAYER)
	}
	//
	if uasmEnable {
		schemaStack.WithLayer(cmd_util.MICRO_ASM_LAYER)
	}
	//
	if mirEnable {
		schemaStack.WithLayer(cmd_util.MIR_LAYER)
	}
	//
	if airEnable {
		schemaStack.WithLayer(cmd_util.AIR_LAYER)
	}
	// Read / compile given source files.
	if mode != SCHEMA_OPTIONAL || len(filenames) > 0 {
		schemaStack.Read(filenames...)
	} else {
		// In this situation, we cannot perform trace expansion.
		builder = builder.WithExpansion(false)
	}
	// Configure builder
	schemaStack.WithTraceBuilder(builder)
	// Done
	return &schemaStack
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.Flags().Bool("version", false, "Report version of this executable")
	// Corset compilation config
	rootCmd.PersistentFlags().Bool("debug", false, "enable debugging constraints")
	rootCmd.PersistentFlags().Bool("legacy", true, "use legacy register allocator")
	rootCmd.PersistentFlags().Bool("no-stdlib", false, "prevent standard library from being included")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "increase logging verbosity")
	rootCmd.PersistentFlags().UintP("opt", "O", 1, "set optimisation level")
	// Assembly lowering config
	rootCmd.PersistentFlags().Bool("vectorize", true, "Apply instruction vectorization")
	rootCmd.PersistentFlags().String("field", "BLS12_377", "prime field to use throughout")
	rootCmd.PersistentFlags().Uint("field-width", 252, "maximum usable bitwidth of underlying field element")
	rootCmd.PersistentFlags().Uint("register-width", 160, "maximum bitwidth for registers")
	// Schema stack
	rootCmd.PersistentFlags().Bool("air", false, "include constraints at AIR level")
	rootCmd.PersistentFlags().Bool("asm", false, "include constraints at ASM level")
	rootCmd.PersistentFlags().Bool("mir", false, "include constraints at MIR level")
	rootCmd.PersistentFlags().Bool("uasm", false, "include constraints at micro ASM level")
	// Trace expansion
	rootCmd.PersistentFlags().Bool("raw", false, "assume input trace already expanded")
	rootCmd.PersistentFlags().Bool("sequential", false, "perform sequential trace expansion")
	rootCmd.PersistentFlags().Bool("defensive", true, "defensively pad modules")
	rootCmd.PersistentFlags().Bool("validate", true, "apply trace validation")
	rootCmd.PersistentFlags().UintP("batch", "b", 1024, "specify batch size for constraint checking")
	// Misc
	rootCmd.PersistentFlags().StringArrayP("set", "S", []string{}, "set value of externalised constant.")
}
