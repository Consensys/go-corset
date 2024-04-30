package cmd

import (
	"fmt"
	"os"

	"github.com/consensys/go-corset/pkg/binfile"
	"github.com/spf13/cobra"
)

// computeCmd represents the compute command
var computeCmd = &cobra.Command{
	Use:   "compute",
	Short: "Given a set of constraints and a trace file, fill the computed columns.",
	Long:  `Given a set of constraints and a trace file, fill the computed columns.`,
	Run: func(cmd *cobra.Command, args []string) {
		file := args[0]
		fmt.Printf("Reading JSON bin file: %s\n", file)

		bytes, err := os.ReadFile(file)
		if err != nil {
			panic(err)
		}

		// Parse binary file into JSON form
		cs, _ := binfile.ConstraintSetFromJson(bytes)

		// Translate JSON form into MIR
		for i := 0; i < len(cs.Constraints); i++ {
			ith := cs.Constraints[i]
			mir := ith.Vanishes.Expr.ToMir()
			fmt.Println(mir)
		}
	},
}

func init() {
	rootCmd.AddCommand(computeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// computeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// computeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
