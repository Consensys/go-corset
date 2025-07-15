package cmd

import (
	"fmt"
	"os"
	"slices"
	"sort"

	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util/collection/hash"
	"github.com/consensys/go-corset/pkg/util/collection/set"
	"github.com/consensys/go-corset/pkg/util/collection/word"
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
		// Extract column names
		trace1cols := extractColumnNames(tracefile1.Columns)
		trace2cols := extractColumnNames(tracefile2.Columns)
		// Sanity check
		if !slices.Equal(trace1cols, trace2cols) {
			var common set.SortedSet[string]
			//
			common.InsertSorted(&trace1cols)
			common.Intersect(&trace2cols)
			//
			reportExtraColumns(filename1, trace1cols, common)
			reportExtraColumns(filename2, trace2cols, common)
			tracefile1.Columns = filterCommonColumns(tracefile1.Columns, common)
			tracefile2.Columns = filterCommonColumns(tracefile2.Columns, common)
		}
		//
		errors := parallelDiff(tracefile1.Columns, tracefile2.Columns)
		// report any differences
		for _, err := range errors {
			fmt.Println(err)
		}
	},
}

func extractColumnNames(columns []trace.BigEndianColumn) set.SortedSet[string] {
	var names set.SortedSet[string]
	//
	for _, c := range columns {
		names.Insert(c.QualifiedName())
	}
	//
	return names
}

func reportExtraColumns(name string, columns []string, common set.SortedSet[string]) {
	for _, c := range columns {
		if !common.Contains(c) {
			fmt.Printf("column %s only in trace %s\n", c, name)
		}
	}
}

func filterCommonColumns(columns []trace.BigEndianColumn, common set.SortedSet[string]) []trace.BigEndianColumn {
	var ncolumns []trace.BigEndianColumn
	//
	for _, c := range columns {
		if common.Contains(c.QualifiedName()) {
			ncolumns = append(ncolumns, c)
		}
	}
	//
	return ncolumns
}

func parallelDiff(columns1 []trace.BigEndianColumn, columns2 []trace.BigEndianColumn) []error {
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

func diffColumns(index int, columns1 []trace.BigEndianColumn, columns2 []trace.BigEndianColumn) []error {
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
		count1, ok1 := set1.Get(val)
		count2, ok2 := set2.Get(val)
		//
		if ok1 != ok2 || count1 != count2 {
			err := fmt.Errorf("column %s, element %s occurs %d or %d times", name, val.String(), count1, count2)
			errors = append(errors, err)
		}
	}
	//
	return errors
}

func identify_vals(lhs hash.Map[word.BigEndian, uint], rhs hash.Map[word.BigEndian, uint]) []word.BigEndian {
	seen := hash.NewSet[word.BigEndian](0)
	vals := make([]word.BigEndian, 0)
	// lhs
	for iter := lhs.KeyValues(); iter.HasNext(); {
		ith := iter.Next()
		val, count := ith.Split()
		//
		if !seen.Contains(val) && count > 0 {
			vals = append(vals, val)
			seen.Insert(val)
		}
	}
	// rhs
	for iter := rhs.KeyValues(); iter.HasNext(); {
		ith := iter.Next()
		val, count := ith.Split()
		//
		if !seen.Contains(val) && count > 0 {
			vals = append(vals, val)
			seen.Insert(val)
		}
	}
	// Sort items so the output is easier to understand
	sort.Slice(vals, func(i, j int) bool {
		return vals[i].Cmp(vals[j]) < 0
	})
	//
	return vals
}

func summarise(data word.BigEndianArray) hash.Map[word.BigEndian, uint] {
	summary := *hash.NewMap[word.BigEndian, uint](data.Len())
	//
	for i := uint(0); i < data.Len(); i++ {
		ith := data.Get(i)
		if v, ok := summary.Get(ith); ok {
			summary.Insert(ith, v+1)
		} else {
			summary.Insert(ith, 1)
		}
	}
	//
	return summary
}

func findColumn(name string, columns []trace.BigEndianColumn) *trace.BigEndianColumn {
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
