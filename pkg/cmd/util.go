// Copyright Consensys Software Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with
// the License. You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/consensys/go-corset/pkg/binfile"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/trace/json"
	"github.com/consensys/go-corset/pkg/trace/lt"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/typed"
	"github.com/spf13/cobra"
)

// GetFlag gets an expected flag, or panic if an error arises.
func GetFlag(cmd *cobra.Command, flag string) bool {
	r, err := cmd.Flags().GetBool(flag)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}

	return r
}

// GetInt gets an expectedsigned integer, or panic if an error arises.
func GetInt(cmd *cobra.Command, flag string) int {
	r, err := cmd.Flags().GetInt(flag)
	if err != nil {
		fmt.Println(err)
		os.Exit(3)
	}

	return r
}

// GetUint gets an expected unsigned integer, or panic if an error arises.
func GetUint(cmd *cobra.Command, flag string) uint {
	r, err := cmd.Flags().GetUint(flag)
	if err != nil {
		fmt.Println(err)
		os.Exit(4)
	}

	return r
}

// GetString gets an expected string, or panic if an error arises.
func GetString(cmd *cobra.Command, flag string) string {
	r, err := cmd.Flags().GetString(flag)
	if err != nil {
		fmt.Println(err)
		os.Exit(4)
	}

	return r
}

// GetStringArray gets an expected string array, or panic if an error arises.
func GetStringArray(cmd *cobra.Command, flag string) []string {
	r, err := cmd.Flags().GetStringArray(flag)
	if err != nil {
		fmt.Println(err)
		os.Exit(4)
	}

	return r
}

// GetIntArray gets an expected int array, or panic if an error arises.
func GetIntArray(cmd *cobra.Command, flag string) []int {
	tmp, err := cmd.Flags().GetStringArray(flag)
	if err != nil {
		fmt.Println(err)
		os.Exit(4)
	}
	//
	r := make([]int, len(tmp))
	//
	for i, str := range tmp {
		ith, err := strconv.ParseInt(str, 16, 8)
		// Error check
		if err != nil {
			panic(err.Error())
		}
		//
		r[i] = int(ith)
	}
	//
	return r
}

// func writeCoverageReport(filename string, binfile *binfile.BinaryFile, coverage [3]sc.CoverageMap,
// 	config mir.OptimisationConfig) {
// 	//
// 	var (
// 		air                 = coverage[0]
// 		mir                 = coverage[1]
// 		hir                 = coverage[2]
// 		data map[string]any = make(map[string]any)
// 	)
// 	// Lower schemas
// 	hirSchema := &binfile.Schema
// 	mirSchema := hirSchema.LowerToMir()
// 	airSchema := mirSchema.LowerToAir(config)
// 	// Add AIR data (if applicable)
// 	if !air.IsEmpty() {
// 		data["air"] = air.ToJson(airSchema)
// 	}
// 	// Add MIR data (if applicable)
// 	if !mir.IsEmpty() {
// 		data["mir"] = mir.ToJson(mirSchema)
// 	}
// 	// Add HIR data (if applicable)
// 	if !hir.IsEmpty() {
// 		data["hir"] = hir.ToJson(hirSchema)
// 	}
// 	// write to disk
// 	jsonString, err := enc_json.Marshal(data)
// 	//
// 	if err != nil {
// 		fmt.Println(err)
// 		os.Exit(5)
// 	} else if err := os.WriteFile(filename, jsonString, 0644); err != nil {
// 		fmt.Println(err)
// 		os.Exit(6)
// 	}
// }

//func readCoverageReport(filename string,binfile *binfile.BinaryFile,config mir.OptimisationConfig) [3]sc.CoverageMap {
// 	var (
// 		report map[string]map[string][]uint
// 		air    sc.CoverageMap
// 		mir    sc.CoverageMap
// 		hir    sc.CoverageMap
// 	)
// 	// Read data file
// 	bytes, err := os.ReadFile(filename)
// 	// Lower schemas
// 	hirSchema := &binfile.Schema
// 	mirSchema := hirSchema.LowerToMir()
// 	airSchema := mirSchema.LowerToAir(config)
// 	// Check success
// 	if err == nil {
// 		if err = enc_json.Unmarshal(bytes, &report); err == nil {
// 			// Read air section
// 			if section, ok := report["air"]; ok {
// 				air = readCoverageReportSection(section, airSchema)
// 			}
// 			// Read mir section
// 			if section, ok := report["mir"]; ok {
// 				mir = readCoverageReportSection(section, mirSchema)
// 			}
// 			// Read hir section
// 			if section, ok := report["hir"]; ok {
// 				hir = readCoverageReportSection(section, hirSchema)
// 			}
// 			// Done
// 			return [3]sc.CoverageMap{air, mir, hir}
// 		}
// 	}
// 	// Handle error
// 	fmt.Println(err)
// 	os.Exit(4)
// 	// unreachable
// 	return [3]sc.CoverageMap{air, mir, hir}
// }

// func readCoverageReportSection(section map[string][]uint, schema sc.Schema) sc.CoverageMap {
// 	report := sc.NewBranchCoverage()
// 	//
// 	for k, vals := range section {
// 		var (
// 			covered            bit.Set
// 			mid, name, casenum = splitConstraintName(k, schema)
// 		)
// 		// Insert all elements
// 		covered.InsertAll(vals...)
// 		// Done
// 		report.Record(mid, name, casenum, covered)
// 	}
// 	//
// 	return report
// }

// func splitConstraintName(name string, schema sc.Schema) (uint, string, uint) {
// 	mid, name := splitConstraintModuleName(name, schema)
// 	name, casenum := splitConstraintNameNum(name)
// 	// Done
// 	return mid, name, casenum
// }

// func splitConstraintModuleName(name string, schema sc.Schema) (uint, string) {
// 	var (
// 		err    error
// 		splits = strings.Split(name, ".")
// 	)
// 	//
// 	switch len(splits) {
// 	case 1:
// 		return 0, name
// 	case 2:
// 		// Lookup the module identifier for the given module name
// 		if mid, ok := schema.Modules().Find(func(m sc.Module) bool { return m.Name == splits[0] }); ok {
// 			return mid, splits[1]
// 		}
// 		// error
// 		err = fmt.Errorf("unknown module %s in coverage report", splits[0])
// 	default:
// 		err = fmt.Errorf("unknown constraint %s in coverage report", name)
// 	}
// 	// Handle error
// 	fmt.Println(err)
// 	os.Exit(4)
// 	// unreachable
// 	return 0, ""
// }
// func splitConstraintNameNum(name string) (string, uint) {
// 	var (
// 		err    error
// 		splits = strings.Split(name, "#")
// 	)
// 	//
// 	switch len(splits) {
// 	case 1:
// 		return name, 0
// 	case 2:
// 		var num int
// 		// Lookup the module identifier for the given module name
// 		if num, err = strconv.Atoi(splits[1]); err == nil && num >= 0 {
// 			return splits[0], uint(num)
// 		}
// 		// error
// 		err = fmt.Errorf("unknown module %s in coverage report", splits[0])
// 	default:
// 		err = fmt.Errorf("unknown constraint %s in coverage report", name)
// 	}
// 	// Handle error
// 	fmt.Println(err)
// 	os.Exit(4)
// 	// unreachable
// 	return "", 0
// }

func writeBatchedTracesFile(filename string, traces ...[]trace.RawColumn) {
	var buf bytes.Buffer
	// Check file extension
	if len(traces) == 1 {
		writeTraceFile(filename, lt.NewTraceFile(nil, traces[0]))
		return
	}
	// Always write JSON in batched mode
	for _, trace := range traces {
		js := json.ToJsonString(trace)
		buf.WriteString(js)
		buf.WriteString("\n")
	}
	// Write file
	if err := os.WriteFile(filename, buf.Bytes(), 0644); err != nil {
		// Handle error
		fmt.Println(err)
		os.Exit(4)
	}
}

// Write a given trace file to disk
func writeTraceFile(filename string, tracefile *lt.TraceFile) {
	var err error

	var bytes []byte
	// Check file extension
	ext := path.Ext(filename)
	//
	switch ext {
	case ".json":
		js := json.ToJsonString(tracefile.Columns)
		//
		if err = os.WriteFile(filename, []byte(js), 0644); err == nil {
			return
		}
	case ".lt":
		bytes, err = tracefile.MarshalBinary()
		//
		if err == nil {
			if err = os.WriteFile(filename, bytes, 0644); err == nil {
				return
			}
		}
	default:
		err = fmt.Errorf("unknown trace file format: %s", ext)
	}
	// Handle error
	fmt.Println(err)
	os.Exit(4)
}

// ReadTraceFile reads a trace file (either binary lt or json), and parses it
// into an array of raw columns.  The determination of what kind of trace file
// (i.e. binary or json) is based on the extension.
func ReadTraceFile(filename string) *lt.TraceFile {
	var columns []trace.RawColumn
	// Read data file
	bytes, err := os.ReadFile(filename)
	// Check success
	if err == nil {
		// Check file extension
		ext := path.Ext(filename)
		//
		switch ext {
		case ".json":
			columns, err = json.FromBytes(bytes)
			if err == nil {
				return lt.NewTraceFile(nil, columns)
			}
		case ".lt":
			// Check for legacy format
			if !lt.IsTraceFile(bytes) {
				// legacy format
				columns, err = lt.FromBytesLegacy(bytes)
				if err == nil {
					return lt.NewTraceFile(nil, columns)
				}
			} else {
				// versioned format
				var tracefile lt.TraceFile
				//
				if err = tracefile.UnmarshalBinary(bytes); err == nil {
					return &tracefile
				}
			}
			//
		default:
			err = fmt.Errorf("unknown trace file format: %s", ext)
		}
	}
	// Handle error
	fmt.Println(err)
	os.Exit(2)
	// unreachable
	return nil
}

// ReadBatchedTraceFile reads a file containing zero or more traces expressed as
// JSON, where each trace is on a separate line.
func ReadBatchedTraceFile(filename string) [][]trace.RawColumn {
	lines := util.ReadInputFile(filename)
	traces := make([][]trace.RawColumn, 0)
	// Read constraints line by line
	for i, line := range lines {
		// Parse input line as JSON
		if line != "" && !strings.HasPrefix(line, ";;") {
			tr, err := json.FromBytes([]byte(line))
			if err != nil {
				msg := fmt.Sprintf("%s:%d: %s", filename, i+1, err)
				panic(msg)
			}

			traces = append(traces, tr)
		}
	}

	return traces
}

// WriteBinaryFile writes a binary file (e.g. zkevm.bin) to disk using the given
// binfile versioning defined in the binfile package.
//
//nolint:errcheck
func WriteBinaryFile(binfile *binfile.BinaryFile, filename string) {
	var (
		bytes []byte
		err   error
	)
	// Encode binary file as bytes
	if bytes, err = binfile.MarshalBinary(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	// Write file
	if err := os.WriteFile(filename, bytes, 0644); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func maxHeightColumns(cols []trace.RawColumn) uint {
	h := uint(0)
	// Iterate over modules
	for _, col := range cols {
		h = max(h, col.Data.Len())
	}
	// Done
	return h
}

func printTypedMetadata(indent uint, metadata typed.Map) {
	for _, k := range metadata.Keys() {
		printIndent(indent)
		//
		if val, ok := metadata.String(k); ok {
			fmt.Printf("%s: %s\n", k, val)
		} else if val, ok := metadata.Map(k); ok {
			fmt.Printf("%s:\n", k)
			printTypedMetadata(indent+1, val)
		} else if metadata.Nil(k) {
			fmt.Printf("%s: (nil)\n", k)
		} else {
			fmt.Printf("%s: ???\n", k)
		}
	}
}

func printIndent(indent uint) {
	for i := uint(0); i < indent; i++ {
		fmt.Print("\t")
	}
}
