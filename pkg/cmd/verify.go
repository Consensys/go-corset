package cmd

import (
	"fmt"
	"os"

	"github.com/consensys/go-corset/pkg/cmd/verify/picus"
	"github.com/consensys/go-corset/pkg/ir/air"
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
// into a constraint verification backend. The current translator supports translating AIR and
// MIR schemas.
func runVerifyCmd[F field.Element[F]](cmd *cobra.Command, args []string) {
	mirEnable := GetFlag(cmd, "mir")

	backend := GetString(cmd, "tool")
	if backend != "picus" {
		fmt.Fprintf(os.Stderr, "%v", fmt.Errorf("expected `backend` = \"picus\". Found %s", backend))
		os.Exit(1)
	}

	schemas := getSchemaStack[F](cmd, SCHEMA_DEFAULT_MIR, args...).Build()
	for _, schema := range schemas.ConcreteSchemas() {
		switch v := schema.(type) {
		case mir.Schema[F]:
			// only translate mir schema if explicitly specified
			if mirEnable {
				picusLowering := picus.NewMirPicusTranslator(v)
				picusProgram := picusLowering.Translate()

				if _, err := picusProgram.WriteTo(os.Stdout); err != nil {
					fmt.Fprintf(os.Stderr, "Error writing out Picus program: %v", err)
					os.Exit(1)
				}
			}
		case air.Schema[F]:
			picusLowering := picus.NewAirPicusTranslator(v)
			picusProgram := picusLowering.Translate()

			if _, err := picusProgram.WriteTo(os.Stdout); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing out Picus program: %v", err)
				os.Exit(1)
			}
		}
	}
}

//nolint:errcheck
func init() {
	rootCmd.AddCommand(genVerifyCmd)
	genVerifyCmd.Flags().StringP("tool", "t", "picus", "specify output file(s).")
}
