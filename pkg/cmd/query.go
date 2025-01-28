package cmd

import (
	"fmt"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/spf13/cobra"
)

var queryCmd = &cobra.Command{
	Use:   "query [flags] binary_file",
	Short: "Query information from a binary package.",
	Long:  `Query specific information from the binary package.`,
	Run: func(cmd *cobra.Command, args []string) {
		field := GetFlag(cmd, "field-columns")
		// Parse constraints
		hirSchema := readSchema(true, false, true, args[0:])
		if field {
			schemaCols := hirSchema.Columns()
			// Check each column
			for i := uint(0); i < schemaCols.Count(); i++ {
				scCol := schemaCols.Next()
				// Extract type for ith column
				colType := scCol.DataType
				// If type field
				if colType.String() == (&schema.FieldType{}).String() {
					fmt.Println(scCol.Name)
				}
			}
		}
	},
}

//nolint:errcheck
func init() {
	rootCmd.AddCommand(queryCmd)
	queryCmd.Flags().Bool("field-columns", false, "list column names of field type. ")
}
