package cmd

import (
	"fmt"
	"os"
	"path"

	"github.com/consensys/go-corset/pkg/binfile"
	"github.com/consensys/go-corset/pkg/hir"
	"github.com/consensys/go-corset/pkg/table"
	"github.com/spf13/cobra"
)

// computeCmd represents the compute command
var checkCmd = &cobra.Command{
	Use:   "check [flags] trace_file constraint_file",
	Short: "Check a given trace against a set of constraints.",
	Long: `Check a given trace against a set of constraints.
	Traces can be given either as JSON or binary lt files.
	Constraints can be given either as lisp or bin files.`,
	Run: func(cmd *cobra.Command, args []string) {
		var trace *table.ArrayTrace
		var hirSchema *hir.Schema

		if len(args) != 2 {
			fmt.Println(cmd.UsageString())
			os.Exit(1)
		}
		raw, err := cmd.Flags().GetBool("raw")
		if err != nil {
			fmt.Println(err)
			os.Exit(2)
		}
		// Parse trace
		trace = readTraceFile(args[0])
		// Parse constraints
		hirSchema = readSchemaFile(args[1])
		// Go!
		checkTraceWithLowering(trace, hirSchema, raw)
	},
}

// Parse a trace file using a parser based on the extension of the filename.
func readTraceFile(filename string) *table.ArrayTrace {
	bytes, err := os.ReadFile(filename)
	if err == nil {
		// Check file extension
		ext := path.Ext(filename)
		//
		switch ext {
		case ".json":
			trace, err := table.ParseJsonTrace(bytes)
			if err == nil {
				return trace
			}
		case ".lt":
			panic("Support for lt trace files not implemented (yet).")
		default:
			err = fmt.Errorf("Unknown trace file format: %s", ext)
		}
	}
	// Handle error
	fmt.Println(err)
	os.Exit(2)
	// unreachable
	return nil
}

// Parse a constraints schema file using a parser based on the extension of the
// filename.
func readSchemaFile(filename string) *hir.Schema {
	bytes, err := os.ReadFile(filename)
	if err == nil {
		// Check file extension
		ext := path.Ext(filename)
		//
		switch ext {
		case ".lisp":
			schema, err := hir.ParseSchemaSExp(string(bytes))
			if err == nil {
				return schema
			}
		case ".bin":
			schema, err := binfile.HirSchemaFromJson(bytes)
			if err == nil {
				return schema
			}
		default:
			err = fmt.Errorf("Unknown trace file format: %s\n", ext)
		}
	}
	// Handle error
	fmt.Println(err)
	os.Exit(2)
	// unreachable
	return nil
}

// Check a given trace is consistently accepted (or rejected) at the different
// IR levels.
func checkTraceWithLowering(tr *table.ArrayTrace, hirSchema *hir.Schema, raw bool) {
	// Lower HIR => MIR
	mirSchema := hirSchema.LowerToMir()
	// Lower MIR => AIR
	airSchema := mirSchema.LowerToAir()
	//
	errHIR := checkTrace(tr, hirSchema, raw)
	errMIR := checkTrace(tr, mirSchema, raw)
	errAIR := checkTrace(tr, airSchema, raw)
	//
	if errHIR != nil || errMIR != nil || errAIR != nil {
		strHIR := errHIR.Error()
		strMIR := errMIR.Error()
		strAIR := errAIR.Error()
		// At least one error encountered.
		if strHIR == strMIR && strMIR == strAIR {
			fmt.Println(errHIR)
		} else {
			reportError(errHIR, "HIR")
			reportError(errMIR, "MIR")
			reportError(errAIR, "AIR")
		}

		os.Exit(1)
	}
}

func checkTrace(tr *table.ArrayTrace, schema table.TraceSchema, raw bool) error {
	if !raw {
		// Clone to prevent interefence with subsequent checks
		tr = tr.Clone()
		// Expand trace
		if err := schema.ExpandTrace(tr); err != nil {
			return err
		}
	}
	// Check whether accepted or not.
	return schema.Accepts(tr)
}

func reportError(err error, ir string) {
	if err != nil {
		fmt.Printf("%s: %s\n", ir, err)
	} else {
		fmt.Printf("Trace should have been rejected at %s level.\n", ir)
	}
}

func init() {
	rootCmd.AddCommand(checkCmd)
	checkCmd.Flags().Bool("raw", false, "assume input trace already expanded")
}
