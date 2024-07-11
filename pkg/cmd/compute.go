package cmd

import (
	"github.com/spf13/cobra"
)

// computeCmd represents the compute command
var computeCmd = &cobra.Command{
	Use:   "compute",
	Short: "Given a set of constraints and a trace file, fill the computed columns.",
	Long:  `Given a set of constraints and a trace file, fill the computed columns.`,
	Run: func(cmd *cobra.Command, args []string) {
		panic("todo")
	},
}

func init() {
	rootCmd.AddCommand(computeCmd)
}
