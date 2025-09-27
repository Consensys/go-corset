package cmd

import (
	"fmt"
	"os"

	"github.com/consensys/go-corset/pkg/asm"
	cmd_util "github.com/consensys/go-corset/pkg/cmd/util"
	"github.com/consensys/go-corset/pkg/cmd/verify/picus"
	"github.com/consensys/go-corset/pkg/corset"
	"github.com/consensys/go-corset/pkg/ir/mir"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/field/bls12_377"
	"github.com/consensys/go-corset/pkg/util/field/gf251"
	"github.com/consensys/go-corset/pkg/util/field/gf8209"
	"github.com/consensys/go-corset/pkg/util/field/koalabear"
	"github.com/spf13/cobra"
)

var genVerifyCmd = &cobra.Command{
	Use:   "verify [flags]",
	Short: "Verify constraints using [tool] verifier.",
	Long:  `Extract constraints into a form required by target verification tool`,
	Run: func(cmd *cobra.Command, args []string) {
		runFieldAgnosticCmd(cmd, args, verifyCmd)
	},
}

// Available instances
var verifyCmd = []FieldAgnosticCmd{
	{sc.GF_251, runVerifyCmd[gf251.Element]},
	{sc.GF_8209, runVerifyCmd[gf8209.Element]},
	{sc.KOALABEAR_16, runVerifyCmd[koalabear.Element]},
	{sc.BLS12_377, runVerifyCmd[bls12_377.Element]},
}

// The `verify` command takes as input a constraint file and translates the constraints
// into a constraint verification backend. The current translator only supports translating `mir`
// files to a Picus backend.
func runVerifyCmd[F field.Element[F]](cmd *cobra.Command, args []string) {
	// Configure log level
	backend := GetString(cmd, "backend")
	if backend == "picus" {
		mirSchema := getMirSchema[F](cmd, args)
		picusLowering := picus.NewPicusTranslator(mirSchema)
		picusProgram := picusLowering.Translate()
		if _, err := picusProgram.WriteTo(os.Stdout); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing out Picus program: %v", err)
		}
	}
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
		fmt.Fprintf(os.Stderr, "%v", fmt.Errorf("-verify expects -mir flag"))
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
	rootCmd.AddCommand(genVerifyCmd)
	genVerifyCmd.Flags().StringP("backend", "b", "picus", "specify output file(s).")
}
