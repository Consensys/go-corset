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
	enc_json "encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/consensys/go-corset/pkg/binfile"
	legacy_binfile "github.com/consensys/go-corset/pkg/binfile/legacy"
	"github.com/consensys/go-corset/pkg/corset"
	"github.com/consensys/go-corset/pkg/hir"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/trace/json"
	"github.com/consensys/go-corset/pkg/trace/lt"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/bit"
	"github.com/consensys/go-corset/pkg/util/sexp"
	log "github.com/sirupsen/logrus"
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

func writeCoverageReport(filename string, binfile *binfile.BinaryFile, coverage [3]sc.CoverageMap) {
	var (
		air                 = coverage[0]
		mir                 = coverage[1]
		hir                 = coverage[2]
		data map[string]any = make(map[string]any)
	)
	// Lower schemas
	hirSchema := &binfile.Schema
	mirSchema := hirSchema.LowerToMir()
	airSchema := mirSchema.LowerToAir()
	// Add AIR data (if applicable)
	if !air.IsEmpty() {
		data["air"] = air.ToJson(airSchema)
	}
	// Add MIR data (if applicable)
	if !mir.IsEmpty() {
		data["mir"] = mir.ToJson(mirSchema)
	}
	// Add HIR data (if applicable)
	if !hir.IsEmpty() {
		data["hir"] = hir.ToJson(hirSchema)
	}
	// write to disk
	jsonString, err := enc_json.Marshal(data)
	//
	if err != nil {
		fmt.Println(err)
		os.Exit(5)
	} else if err := os.WriteFile(filename, jsonString, 0644); err != nil {
		fmt.Println(err)
		os.Exit(6)
	}
}

func readCoverageReport(filename string, binfile *binfile.BinaryFile) [3]sc.CoverageMap {
	var (
		report map[string]map[string][]uint
		air    sc.CoverageMap
		mir    sc.CoverageMap
		hir    sc.CoverageMap
	)
	// Read data file
	bytes, err := os.ReadFile(filename)
	// Lower schemas
	hirSchema := &binfile.Schema
	mirSchema := hirSchema.LowerToMir()
	airSchema := mirSchema.LowerToAir()
	// Check success
	if err == nil {
		if err = enc_json.Unmarshal(bytes, &report); err == nil {
			// Read air section
			if section, ok := report["air"]; ok {
				air = readCoverageReportSection(section, airSchema)
			}
			// Read mir section
			if section, ok := report["mir"]; ok {
				mir = readCoverageReportSection(section, mirSchema)
			}
			// Read hir section
			if section, ok := report["hir"]; ok {
				hir = readCoverageReportSection(section, hirSchema)
			}
			// Done
			return [3]sc.CoverageMap{air, mir, hir}
		}
	}
	// Handle error
	fmt.Println(err)
	os.Exit(4)
	// unreachable
	return [3]sc.CoverageMap{air, mir, hir}
}

func readCoverageReportSection(section map[string][]uint, schema sc.Schema) sc.CoverageMap {
	report := sc.NewBranchCoverage()
	//
	for k, vals := range section {
		var (
			covered            bit.Set
			mid, name, casenum = splitConstraintName(k, schema)
		)
		// Insert all elements
		covered.InsertAll(vals...)
		// Done
		report.Record(mid, name, casenum, covered)
	}
	//
	return report
}

func splitConstraintName(name string, schema sc.Schema) (uint, string, uint) {
	mid, name := splitConstraintModuleName(name, schema)
	name, casenum := splitConstraintNameNum(name)
	// Done
	return mid, name, casenum
}

func splitConstraintModuleName(name string, schema sc.Schema) (uint, string) {
	var (
		err    error
		splits = strings.Split(name, ".")
	)
	//
	switch len(splits) {
	case 1:
		return 0, name
	case 2:
		// Lookup the module identifier for the given module name
		if mid, ok := schema.Modules().Find(func(m sc.Module) bool { return m.Name == splits[0] }); ok {
			return mid, splits[1]
		}
		// error
		err = fmt.Errorf("unknown module %s in coverage report", splits[0])
	default:
		err = fmt.Errorf("unknown constraint %s in coverage report", name)
	}
	// Handle error
	fmt.Println(err)
	os.Exit(4)
	// unreachable
	return 0, ""
}
func splitConstraintNameNum(name string) (string, uint) {
	var (
		err    error
		splits = strings.Split(name, "#")
	)
	//
	switch len(splits) {
	case 1:
		return name, 0
	case 2:
		var num int
		// Lookup the module identifier for the given module name
		if num, err = strconv.Atoi(splits[1]); err == nil && num >= 0 {
			return splits[0], uint(num)
		}
		// error
		err = fmt.Errorf("unknown module %s in coverage report", splits[0])
	default:
		err = fmt.Errorf("unknown constraint %s in coverage report", name)
	}
	// Handle error
	fmt.Println(err)
	os.Exit(4)
	// unreachable
	return "", 0
}

// Write a given trace file to disk
func writeTraceFile(filename string, columns []trace.RawColumn) {
	var err error

	var bytes []byte
	// Check file extension
	ext := path.Ext(filename)
	//
	switch ext {
	case ".json":
		js := json.ToJsonString(columns)
		//
		if err = os.WriteFile(filename, []byte(js), 0644); err == nil {
			return
		}
	case ".lt":
		bytes, err = lt.ToBytes(columns)
		//
		if err == nil {
			if err = os.WriteFile(filename, bytes, 0644); err == nil {
				return
			}
		}
	default:
		err = fmt.Errorf("Unknown trace file format: %s", ext)
	}
	// Handle error
	fmt.Println(err)
	os.Exit(4)
}

// ReadTraceFile reads a trace file (either binary lt or json), and parses it
// into an array of raw columns.  The determination of what kind of trace file
// (i.e. binary or json) is based on the extension.
func ReadTraceFile(filename string) []trace.RawColumn {
	var tr []trace.RawColumn
	// Read data file
	bytes, err := os.ReadFile(filename)
	// Check success
	if err == nil {
		// Check file extension
		ext := path.Ext(filename)
		//
		switch ext {
		case ".json":
			tr, err = json.FromBytes(bytes)
			if err == nil {
				return tr
			}
		case ".lt":
			tr, err = lt.FromBytes(bytes)
			if err == nil {
				return tr
			}
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
func WriteBinaryFile(binfile *binfile.BinaryFile, legacy bool, filename string) {
	var (
		bytes []byte
		err   error
	)
	// Sanity checks
	if legacy {
		// Currently, there is no support for this.
		fmt.Println("legacy binary format not supported for writing")
	}
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

// ReadConstraintFiles provides a generic interface for reading constraint files
// in one of two ways.  If a single file is provided with the "bin" extension
// then this is treated as a binfile (e.g. zkevm.bin).  Otherwise, the files are
// assumed to be source (i.e. lisp) files and are read in and then compiled into
// a binfile.  NOTES: (1) when reading a binfile, the legacy format can be
// explicitly specified (though it is also detected automatically so this is
// largely redundant now); (2) when source files are provided, they can be
// compiled with (or without) the standard library.  Generally speaking, you
// want to compile with the standard library.  However, some internal tests are
// run without including the standard library to minimise the surface area.
func ReadConstraintFiles(stdlib bool, debug bool, legacy bool, filenames []string) *binfile.BinaryFile {
	var err error
	//
	if len(filenames) == 0 {
		fmt.Println("source or binary constraint(s) file required.")
		os.Exit(5)
	} else if len(filenames) == 1 && path.Ext(filenames[0]) == ".bin" {
		// Single (binary) file supplied
		return ReadBinaryFile(legacy, filenames[0])
	}
	// Recursively expand any directories given in the list of filenames.
	if filenames, err = expandSourceFiles(filenames); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// Must be source files
	return CompileSourceFiles(stdlib, debug, filenames)
}

// ReadBinaryFile reads a binfile which includes the metadata bytes, along with
// the schema, and any included attributes.  The legacy format can be explicitly
// requested, though this function will now automatically detect whether it is a
// legeacy or non-legacy binfile.
func ReadBinaryFile(legacy bool, filename string) *binfile.BinaryFile {
	var binf binfile.BinaryFile
	// Read schema file
	data, err := os.ReadFile(filename)
	// Handle errors
	if err == nil && (legacy || !binfile.IsBinaryFile(data)) {
		var schema *hir.Schema
		// Read the binary file
		schema, err = legacy_binfile.HirSchemaFromJson(data)
		//
		binf.Schema = *schema
	} else if err == nil {
		err = binf.UnmarshalBinary(data)
	}
	// Return if no errors
	if err == nil {
		return &binf
	}
	// Handle error & exit
	fmt.Println(err)
	os.Exit(2)
	// unreachable
	return nil
}

// CompileSourceFiles accepts a set of source files and compiles them into a
// single schema.  This can result, for example, in a syntax error, etc.  This
// can be done with (or without) including the standard library, and also with
// (or without) debug constraints.
func CompileSourceFiles(stdlib bool, debug bool, filenames []string) *binfile.BinaryFile {
	srcfiles := make([]*sexp.SourceFile, len(filenames))
	// Read each file
	for i, n := range filenames {
		log.Debug(fmt.Sprintf("including source file %s", n))
		// Read source file
		bytes, err := os.ReadFile(n)
		// Sanity check for errors
		if err != nil {
			fmt.Println(err)
			os.Exit(3)
		}
		//
		srcfiles[i] = sexp.NewSourceFile(n, bytes)
	}
	// Parse and compile source files
	binf, errs := corset.CompileSourceFiles(stdlib, debug, srcfiles)
	// Check for any errors
	if len(errs) == 0 {
		return binf
	}
	// Report errors
	for _, err := range errs {
		printSyntaxError(&err)
	}
	// Fail
	os.Exit(4)
	// unreachable
	return nil
}

// Look through the list of filenames and identify any which are directories.
// Those are then recursively expanded.
func expandSourceFiles(filenames []string) ([]string, error) {
	var expandedFilenames []string
	//
	for _, f := range filenames {
		// Lookup information on the given file.
		if info, err := os.Stat(f); err != nil {
			// Something is wrong with one of the files provided, therefore
			// terminate with an error.
			return nil, err
		} else if info.IsDir() {
			// This a directory, so read its contents
			if contents, err := expandDirectory(f); err != nil {
				return nil, err
			} else {
				expandedFilenames = append(expandedFilenames, contents...)
			}
		} else {
			// This is a single file
			expandedFilenames = append(expandedFilenames, f)
		}
	}
	//
	return expandedFilenames, nil
}

// Recursively search through a given directory looking for any lisp files.
func expandDirectory(dirname string) ([]string, error) {
	var filenames []string
	// Recursively walk the given directory.
	err := filepath.Walk(dirname, func(filename string, info os.FileInfo, err error) error {
		if !info.IsDir() && path.Ext(filename) == ".lisp" {
			filenames = append(filenames, filename)
		} else if !info.IsDir() && path.Ext(filename) == ".lispX" {
			log.Info(fmt.Sprintf("ignoring file %s", filename))
		}
		// Continue.
		return nil
	})
	// Done
	return filenames, err
}

// Print a syntax error with appropriate highlighting.
func printSyntaxError(err *sexp.SyntaxError) {
	span := err.Span()
	line := err.FirstEnclosingLine()
	lineOffset := span.Start() - line.Start()
	// Calculate length (ensures don't overflow line)
	length := min(line.Length()-lineOffset, span.Length())
	// Print error + line number
	fmt.Printf("%s:%d:%d-%d %s\n", err.SourceFile().Filename(),
		line.Number(), 1+lineOffset, 1+lineOffset+length, err.Message())
	// Print separator line
	fmt.Println()
	// Print line
	fmt.Println(line.String())
	// Print indent (todo: account for tabs)
	fmt.Print(strings.Repeat(" ", lineOffset))
	// Print highlight
	fmt.Println(strings.Repeat("^", length))
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
