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
	"github.com/consensys/go-corset/pkg/util/collection/pool"
	"github.com/consensys/go-corset/pkg/util/collection/typed"
	"github.com/consensys/go-corset/pkg/util/file"
	"github.com/consensys/go-corset/pkg/util/word"
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

func writeBatchedTracesFile(filename string, traces ...lt.TraceFile) {
	var buf bytes.Buffer
	// Check file extension
	if len(traces) == 1 {
		writeTraceFile(filename, traces[0])
		return
	}
	// Always write JSON in batched mode
	for _, trace := range traces {
		js := json.ToJsonString(trace.Columns)
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
func writeTraceFile(filename string, tracefile lt.TraceFile) {
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
	case ".lt", ".ltv2":
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
func ReadTraceFile(filename string) lt.TraceFile {
	var (
		stats     = util.NewPerfStats()
		columns   []trace.RawColumn[word.BigEndian]
		pool      pool.LocalHeap[word.BigEndian]
		tracefile lt.TraceFile
	)
	// Read data file
	data, err := os.ReadFile(filename)
	// Check success
	if err == nil {
		// Check file extension
		ext := path.Ext(filename)
		//
		switch ext {
		case ".json":
			pool, columns, err = json.FromBytes(data)
			if err == nil {
				tracefile = lt.NewTraceFile(nil, pool, columns)
			}
		case ".lt", ".ltv2":
			// Check for legacy format
			if !lt.IsTraceFile(data) {
				// legacy format
				pool, columns, err = lt.FromBytesLegacy(data)
				if err == nil {
					tracefile = lt.NewTraceFile(nil, pool, columns)
				}
			} else {
				err = tracefile.UnmarshalBinary(data)
			}
			//
		default:
			err = fmt.Errorf("unknown trace file format: %s", ext)
		}
	}
	//
	stats.Log("Reading trace file")
	//
	if err == nil {
		return tracefile
	}
	// Handle error
	fmt.Println(err)
	os.Exit(2)
	// unreachable
	return lt.TraceFile{}
}

// ReadBatchedTraceFile reads a file containing zero or more traces expressed as
// JSON, where each trace is on a separate line.
func ReadBatchedTraceFile(filename string) []lt.TraceFile {
	var (
		stats  = util.NewPerfStats()
		lines  = file.ReadInputFile(filename)
		traces = make([]lt.TraceFile, 0)
	)
	// Read constraints line by line
	for i, line := range lines {
		// Parse input line as JSON
		if line != "" && !strings.HasPrefix(line, ";;") {
			pool, cols, err := json.FromBytes([]byte(line))
			if err != nil {
				msg := fmt.Sprintf("%s:%d: %s", filename, i+1, err)
				panic(msg)
			}

			traces = append(traces, lt.NewTraceFile(nil, pool, cols))
		}
	}
	//
	stats.Log("Reading trace file")
	//
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

func maxHeightColumns(cols []trace.RawColumn[word.BigEndian]) uint {
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
