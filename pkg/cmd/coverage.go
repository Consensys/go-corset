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
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/consensys/go-corset/pkg/binfile"
	cov "github.com/consensys/go-corset/pkg/cmd/coverage"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var coverageCmd = &cobra.Command{
	Use:   "coverage [flags] coverage_file constraint_file(s)",
	Short: "query coverage data generated for a given set of constraints.",
	Long:  `Provides mechanisms for investigating the coverage data generated for a given set of constraints and traces.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			fmt.Println(cmd.UsageString())
			os.Exit(1)
		}
		// Configure log level
		if GetFlag(cmd, "verbose") {
			log.SetLevel(log.DebugLevel)
		}
		//filter := defaultFilter()
		//
		stdlib := !GetFlag(cmd, "no-stdlib")
		debug := GetFlag(cmd, "debug")
		legacy := GetFlag(cmd, "legacy")
		expand := GetFlag(cmd, "expand")
		module := GetFlag(cmd, "module")
		//filter = regexFilter(filter, GetString(cmd, "filter"))
		//
		json, others := splitArgs(args)
		// Parse constraints
		binfile := ReadConstraintFiles(stdlib, debug, legacy, others)
		// Parse coverage file
		coverage := readCoverageReports(json, binfile)
		//
		hirSchema := &binfile.Schema
		mirSchema := hirSchema.LowerToMir()
		//airSchema := mirSchema.LowerToAir()
		// Calculate mode
		mode := CONSTRAINT_MODE
		if module && expand {
			fmt.Println("--module and --expand are incompatible")
			os.Exit(2)
		} else if GetString(cmd, "filter") != "" && module {
			fmt.Println("--module and --filter are (currently) incompatible")
			os.Exit(2)
		} else if module {
			mode = MODULE_MODE
		} else if expand {
			mode = EXPANDED_MODE
		}
		// Determine metrics to print
		calcs := cov.DEFAULT_CALCS
		// Determine relevant set of constraint identifiers
		ids := determineConstraintIds(mode, mirSchema)
		// Build the coverage reports
		mirReports := buildReports(calcs, coverage[1], mirSchema)
		//printCoverage(ids, calcs, []sc.CoverageMap{coverage[0]}, airSchema)
		printCoverage(ids, calcs, mirReports, mirSchema)
		//printCoverage(ids, calcs, []sc.CoverageMap{coverage[2]}, hirSchema)
	},
}

const (
	// MODULE_MODE produces module summaries
	MODULE_MODE = uint(0)
	// CONSTRAINT_MODE produces constraint summaries
	CONSTRAINT_MODE = uint(1)
	// EXPANDED_MODE produces expanded constraint summaries
	EXPANDED_MODE = uint(2)
)

func splitArgs(args []string) ([]string, []string) {
	for i, arg := range args {
		if !strings.HasSuffix(arg, ".json") {
			return args[0:i], args[i:]
		}
	}
	//
	return args, nil
}

func determineConstraintIds(_ uint, schema sc.Schema) []cov.ConstraintId {
	var ids []cov.ConstraintId
	//
	for iter := schema.Constraints(); iter.HasNext(); {
		ith := iter.Next()
		mid := ith.Contexts()[0].Module()
		name, num := ith.Name()
		ids = append(ids, cov.ConstraintId{mid, name, num})
	}
	//
	return ids
}

func buildReports(calcs []cov.ColumnCalc, coverage []sc.CoverageMap, schema sc.Schema) []cov.Report {
	reports := make([]cov.Report, len(coverage))
	//
	for i, c := range coverage {
		ith := cov.NewReport(calcs, c, schema)
		reports[i] = *ith
	}
	//
	return reports
}

func readCoverageReports(filenames []string, binf *binfile.BinaryFile) [3][]sc.CoverageMap {
	var maps [3][]sc.CoverageMap
	//
	maps[0] = make([]sc.CoverageMap, len(filenames))
	maps[1] = make([]sc.CoverageMap, len(filenames))
	maps[2] = make([]sc.CoverageMap, len(filenames))
	//
	for i, n := range filenames {
		tmp := readCoverageReport(n, binf)
		maps[0][i] = tmp[0]
		maps[1][i] = tmp[1]
		maps[2][i] = tmp[2]
	}
	//
	return maps
}

// Filter defines the type of a constraint filter.
type Filter func(uint, string, sc.CoverageMap, sc.Schema) bool

func defaultFilter() Filter {
	// The default filter eliminates any constraints which have only a single
	// branch, as these simply dilute the outcome.
	return func(mid uint, name string, cov sc.CoverageMap, schema sc.Schema) bool {
		return true
	}
}

func regexFilter(filter Filter, regexStr string) Filter {
	if regexStr == "" {
		return filter
	}
	//
	regex, err := regexp.Compile(regexStr)
	//
	if err != nil {
		fmt.Printf("invalid filter: %s", err)
		os.Exit(0)
	}
	//
	return and(filter, func(mid uint, name string, _ sc.CoverageMap, schema sc.Schema) bool {
		modName := schema.Modules().Nth(mid).Name
		name = fmt.Sprintf("%s.%s", modName, name)
		//
		return regex.MatchString(name)
	})
}

func and(lhs Filter, rhs Filter) Filter {
	return func(mid uint, name string, cov sc.CoverageMap, schema sc.Schema) bool {
		return lhs(mid, name, cov, schema) && rhs(mid, name, cov, schema)
	}
}

func printCoverage(ids []cov.ConstraintId, calcs []cov.ColumnCalc, coverage []cov.Report, schema sc.Schema) {
	var (
		// Determine number of calculated columns per map
		n = len(calcs)
		// Total number of calculated columns
		m = uint(len(coverage) * n)
	)
	// Make column titles
	titles := make([]string, m+3)
	// Configure titles
	for i := range coverage {
		for j, s := range calcs {
			titles[(i*n)+j+3] = s.Name
		}
	}
	// Initialise row
	rows := [][]string{titles}
	//
	for _, id := range ids {
		row := make([]string, 3)
		// initialise row title
		row[0] = "module"
		row[1] = id.Name
		row[2] = fmt.Sprintf("%d", id.Case)
		// Build up reports
		for _, c := range coverage {
			row = append(row, c.Row(id)...)
		}
		//
		rows = append(rows, row)
	}
	// Print matching entries
	tbl := util.NewTablePrinter(m+3, uint(len(rows)))
	//
	for i, row := range rows {
		tbl.SetRow(uint(i), row...)
	}
	//
	tbl.SetMaxWidth(1, 64)
	//
	tbl.Print()
}

//nolint:errcheck
func init() {
	rootCmd.AddCommand(coverageCmd)
	coverageCmd.Flags().Bool("debug", false, "enable debugging constraints")
	coverageCmd.Flags().BoolP("module", "m", false, "show module summaries")
	coverageCmd.Flags().BoolP("expand", "e", false, "show expanded constraints")
	coverageCmd.Flags().StringP("filter", "f", "", "regex constraint filter")
}
