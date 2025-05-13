package cmd

import (
	"fmt"
	"os"
	"sort"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/spf13/cobra"
)

// traceDiffCmd represents the trace command for manipulating traces.
var traceDiffCmd = &cobra.Command{
	Use:   "diff [flags] trace_file trace_file",
	Short: "Show differences between two trace files.",
	Long: `Reports differences between two trace files,
	which is useful when the trace files are expected to be identical.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 2 {
			fmt.Println(cmd.UsageString())
			os.Exit(1)
		}
		// Extract trace file to diagnose
		filename1 := args[0]
		filename2 := args[1]
		// Read trace files
		tracefile1 := ReadTraceFile(filename1)
		tracefile2 := ReadTraceFile(filename2)
		// Sanity check
		if len(tracefile1.Columns) != len(tracefile2.Columns) {
			fmt.Printf("differing number of columns (%d v %d)", len(tracefile1.Columns), len(tracefile2.Columns))
			os.Exit(2)
		}
		//
		errors := parallelDiff(tracefile1.Columns, tracefile2.Columns)
		// report any differences
		for _, err := range errors {
			fmt.Println(err)
		}
	},
}

func parallelDiff(columns1 []trace.RawColumn, columns2 []trace.RawColumn) []error {
	errors := make([]error, 0)
	ncols := len(columns1)
	// Look through all STAMP columns searching for 0s at the end.
	c := make(chan []error, ncols)
	// Dispatch go-routines
	for i := 0; i < ncols; i++ {
		go func(i int) {
			c <- diffColumns(i, columns1, columns2)
		}(i)
	}
	// Bring back together
	for i := uint(0); i < uint(ncols); i++ {
		// Read packaged result from channel
		res := <-c
		// Record any differences
		errors = append(errors, res...)
	}
	//
	return errors
}

func diffColumns(index int, columns1 []trace.RawColumn, columns2 []trace.RawColumn) []error {
	errors := make([]error, 0)
	name := columns1[index].QualifiedName()
	data1 := columns1[index].Data
	data2 := findColumn(name, columns2).Data
	// Sanity check
	if data2 == nil {
		return errors
	}
	//
	set1 := summarise(data1)
	set2 := summarise(data2)
	// Determine set of unique values
	vals := identify_vals(set1, set2)
	// Print differences
	for _, val := range vals {
		count1, ok1 := set1[val]
		count2, ok2 := set2[val]
		//
		if ok1 != ok2 || count1 != count2 {
			err := fmt.Errorf("column %s, element %s occurs %d or %d times", name, val.String(), count1, count2)
			errors = append(errors, err)
		}
	}
	//
	return errors
}

func identify_vals(lhs map[fr.Element]uint, rhs map[fr.Element]uint) []fr.Element {
	seen := make(map[fr.Element]bool)
	vals := make([]fr.Element, 0)
	// lhs
	for val, count := range lhs {
		if c, ok := seen[val]; (!ok || !c) && count > 0 {
			vals = append(vals, val)
			seen[val] = true
		}
	}
	// rhs
	for val, count := range rhs {
		if c, ok := seen[val]; (!ok || !c) && count > 0 {
			vals = append(vals, val)
			seen[val] = true
		}
	}
	// Sort items so the output is easier to understand
	sort.Slice(vals, func(i, j int) bool {
		return vals[i].Cmp(&vals[j]) < 0
	})
	//
	return vals
}

func summarise(data field.FrArray) map[fr.Element]uint {
	summary := make(map[fr.Element]uint)
	//
	for i := uint(0); i < data.Len(); i++ {
		ith := data.Get(i)
		if v, ok := summary[ith]; ok {
			summary[ith] = v + 1
		} else {
			summary[ith] = 1
		}
	}
	//
	return summary
}

func findColumn(name string, columns []trace.RawColumn) *trace.RawColumn {
	for _, c := range columns {
		if c.QualifiedName() == name {
			return &c
		}
	}
	//
	fmt.Printf("WARNING: missing column %s", name)
	//
	return nil
}

func init() {
	traceCmd.AddCommand(traceDiffCmd)
}
