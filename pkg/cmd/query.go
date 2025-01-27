package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var queryCmd = &cobra.Command{
	Use:   "query [flags] binary_file",
	Short: "Query information from a binary package.",
	Long:  `Query specific information from the binary package.`,
	Run: func(cmd *cobra.Command, args []string) {
		field := GetFlag(cmd, "field-columns")
		fieldType := GetString(cmd, "field-type")
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
				if colType.String() == fieldType {
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
	queryCmd.Flags().String("field-type", "u128", "specify field type. Default is field type u128 for BLS12-377.")
}
