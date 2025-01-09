package cmd

import (
	"github.com/spf13/cobra"
)

var compileCmd = &cobra.Command{
	Use:   "compile [flags] constraint_file(s)",
	Short: "compile constraints into a binary package.",
	Long: `Compile a given set of constraint file(s) into a single binary package which can
	 be subsequently used without requiring a full compilation step.`,
	Run: func(cmd *cobra.Command, args []string) {
		stdlib := !GetFlag(cmd, "no-stdlib")
		debug := GetFlag(cmd, "debug")
		legacy := GetFlag(cmd, "legacy")
		output := GetString(cmd, "output")
		// Parse constraints
		hirSchema := readSchema(stdlib, debug, legacy, args)
		// Serialise as a gob file.
		writeHirSchema(hirSchema, legacy, output)
	},
}

//nolint:errcheck
func init() {
	rootCmd.AddCommand(compileCmd)
	compileCmd.Flags().Bool("debug", false, "enable debugging constraints")
	compileCmd.Flags().StringP("output", "o", "a.bin", "specify output file.")
	compileCmd.MarkFlagRequired("output")
}
