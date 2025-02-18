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
	"math"
	"os"
	"regexp"

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
		filter := defaultFilter()
		//
		stdlib := !GetFlag(cmd, "no-stdlib")
		debug := GetFlag(cmd, "debug")
		legacy := GetFlag(cmd, "legacy")
		expand := GetFlag(cmd, "expand")
		module := GetFlag(cmd, "module")
		filter = regexFilter(filter, GetString(cmd, "filter"))
		// Parse constraints
		binfile := ReadConstraintFiles(stdlib, debug, legacy, args[1:])
		// Parse coverage file
		coverage := readCoverageReport(args[0], binfile)
		//
		hirSchema := &binfile.Schema
		mirSchema := hirSchema.LowerToMir()
		airSchema := mirSchema.LowerToAir()
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
		//
		printCoverage(mode, coverage[0], airSchema, filter)
		printCoverage(mode, coverage[1], mirSchema, filter)
		printCoverage(mode, coverage[2], hirSchema, filter)
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

// Filter defines the type of a constraint filter.
type Filter func(uint, string, sc.CoverageMap, sc.Schema) bool

func defaultFilter() Filter {
	// The default filter eliminates any constraints which have only a single
	// branch, as these simply dilute the outcome.
	return func(mid uint, name string, cov sc.CoverageMap, schema sc.Schema) bool {
		return determineConstraintBranches(mid, name, cov, schema) != 1
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

func printCoverage(mode uint, coverage []sc.CoverageMap, schema sc.Schema, filter Filter) {
	// Determine how many modules there are
	nModules := schema.Modules().Count()
	//
	var n = uint(len(coverage) * len(coverageSummarisers))
	//
	row := make([]string, n+2)
	//
	for i, s := range coverageSummarisers {
		row[i+2] = s.name
	}
	//
	rows := [][]string{row}
	//
	for mid := uint(0); mid < nModules; mid++ {
		rs := generateCoverage(mode, mid, coverage, schema, filter)
		rows = append(rows, rs...)
	}
	// Print matching entries
	tbl := util.NewTablePrinter(n+2, uint(len(rows)))
	//
	for i, row := range rows {
		tbl.SetRow(uint(i), row...)
	}
	//
	tbl.SetMaxWidth(1, 64)
	//
	tbl.Print()
}

func generateCoverage(mode uint, mid uint, coverage []sc.CoverageMap, schema sc.Schema, filter Filter) [][]string {
	//
	var (
		rows  [][]string
		title string
	)
	// Determine how many rows
	mod := schema.Modules().Nth(mid)
	// Print module header
	if mod.Name != "" {
		title = mod.Name
	} else {
		title = "<prelude>"
	}
	//
	if mode == MODULE_MODE {
		crows := generateModuleCoverage(mid, title, coverage, schema, filter)
		rows = append(rows, crows...)
	} else {
		for iter := schema.Constraints(); iter.HasNext(); {
			// Determine constraint name
			ith := iter.Next()
			name, _ := ith.Name()
			// Filter out columns
			if ith.Contexts()[0].Module() == mid && filter(mid, name, coverage, schema) {
				// Construct row
				crows := generateConstraintCoverage(mode, mid, title, name, coverage, schema)
				//
				rows = append(rows, crows...)
			}
		}
	}
	//
	return rows
}

func generateModuleCoverage(mid uint, mod string, coverage []sc.CoverageMap,
	schema sc.Schema, filter Filter) [][]string {
	//
	var (
		n    = len(coverageSummarisers)
		vals = make([]string, (len(coverage)*n)+2)
	)
	//
	vals[0] = mod
	// Apply summarisers
	for i, cov := range coverage {
		for j, fn := range coverageSummarisers {
			vals[(i*n)+j+2] = fn.format(fn.module(mid, cov, schema, filter))
		}
	}
	// Done
	return [][]string{vals}
}

func generateConstraintCoverage(mode uint, mid uint, mod string, name string, coverage []sc.CoverageMap,
	schema sc.Schema) [][]string {
	//
	switch mode {
	case CONSTRAINT_MODE:
		return generateUnexpandedConstraintCoverage(mid, mod, name, coverage, schema)
	case EXPANDED_MODE:
		return generateExpandedConstraintCoverage(mid, mod, name, coverage, schema)
	}
	//
	panic("unreachable")
}

func generateUnexpandedConstraintCoverage(mid uint, mod string, name string, coverage []sc.CoverageMap,
	schema sc.Schema) [][]string {
	//
	var (
		n    = len(coverageSummarisers)
		vals = make([]string, (len(coverage)*n)+2)
	)
	//
	vals[0], vals[1] = mod, name
	// Apply summarisers
	for i, cov := range coverage {
		for j, fn := range coverageSummarisers {
			vals[(i*n)+j+2] = fn.format(fn.constraint(mid, name, cov, schema))
		}
	}
	// Done
	return [][]string{vals}
}

func generateExpandedConstraintCoverage(mid uint, mod string, name string, coverage []sc.CoverageMap,
	schema sc.Schema) [][]string {
	//
	var (
		n      = len(coverageSummarisers)
		ncases = uint(len(coverage.CoverageOf(mid, name)))
		vals   [][]string
	)
	//
	for i := uint(0); i < ncases; i++ {
		row := make([]string, (len(coverage)*n)+2)
		row[0], row[1] = mod, fmt.Sprintf("%s#%d", name, i)
		for j, cov := range coverage {
			// Apply summarisers
			for k, fn := range coverageSummarisers {
				row[(j*n)+k+2] = fn.format(fn.expanded(mid, name, i, cov, schema))
			}
			//
			vals = append(vals, row)
		}
	}
	// Done
	return vals
}

type constraintSummariser struct {
	name       string
	module     func(uint, sc.CoverageMap, sc.Schema, Filter) float64
	constraint func(uint, string, sc.CoverageMap, sc.Schema) float64
	expanded   func(uint, string, uint, sc.CoverageMap, sc.Schema) float64
	format     func(float64) string
}

var coverageSummarisers []constraintSummariser = []constraintSummariser{
	{"Coverage", moduleCoverageSummariser, constraintCoverageSummariser, expandedCoverageSummariser, uintFormatter},
	{"Branches", moduleBranchesSummariser, constraintBranchesSummariser, expandedBranchesSummariser, uintFormatter},
	{"Percentage", modulePercentSummariser, constraintPercentSummariser, expandedPercentSummariser, percentFormatter},
}

func uintFormatter(val float64) string {
	return fmt.Sprintf("%.0f", val)
}

func percentFormatter(val float64) string {
	if math.IsNaN(val) {
		return "-"
	}

	return fmt.Sprintf("%.1f%%", val)
}

// ============================================================================
// Module Summarisers
// ============================================================================

func moduleCoverageSummariser(mid uint, coverage sc.CoverageMap, schema sc.Schema, filter Filter) float64 {
	total := float64(0)
	//
	for iter := coverage.KeysOf(mid).Iter(); iter.HasNext(); {
		// Determine constraint name
		name := iter.Next()
		// Apply filter
		if filter(mid, name, coverage, schema) {
			total += constraintCoverageSummariser(mid, name, coverage, schema)
		}
	}
	//
	return total
}

func moduleBranchesSummariser(mid uint, coverage sc.CoverageMap, schema sc.Schema, filter Filter) float64 {
	total := float64(0)
	//
	for iter := coverage.KeysOf(mid).Iter(); iter.HasNext(); {
		// Determine constraint name
		name := iter.Next()
		// Apply filter
		if filter(mid, name, coverage, schema) {
			total += constraintBranchesSummariser(mid, name, coverage, schema)
		}
	}
	//
	return total
}

func modulePercentSummariser(mid uint, coverage sc.CoverageMap, schema sc.Schema, filter Filter) float64 {
	val := moduleCoverageSummariser(mid, coverage, schema, filter)
	total := moduleBranchesSummariser(mid, coverage, schema, filter)
	percent := float64(val*100) / float64(total)
	//
	return percent
}

// ============================================================================
// Constraint Summarisers
// ============================================================================

func constraintCoverageSummariser(mid uint, name string, coverage sc.CoverageMap, schema sc.Schema) float64 {
	var total uint
	// Extract available coverage data
	bitsets := coverage.CoverageOf(mid, name)
	//
	for _, b := range bitsets {
		// Aggregated coverage for this case
		total += b.Count()
	}
	// Done
	return float64(total)
}

func constraintBranchesSummariser(mid uint, name string, coverage sc.CoverageMap, schema sc.Schema) float64 {
	var branches uint = 0
	// Extract available coverage data
	bitsets := coverage.CoverageOf(mid, name)
	//
	for i := range bitsets {
		// Lookup actual constraint
		if c := findConstraint(mid, name, uint(i), schema); c != nil {
			branches += c.Branches()
		} else {
			module := schema.Modules().Nth(mid)
			log.Errorf("unknown constraint \"%s.%s#%d\" in coverage report", module, name, i)
		}
	}
	// Done
	return float64(branches)
}

func constraintPercentSummariser(mid uint, name string, coverage sc.CoverageMap, schema sc.Schema) float64 {
	val := constraintCoverageSummariser(mid, name, coverage, schema)
	total := constraintBranchesSummariser(mid, name, coverage, schema)
	percent := float64(val*100) / float64(total)
	//
	return percent
}

// ============================================================================
// Expanded Summarisers
// ============================================================================

func expandedCoverageSummariser(mid uint, name string, casenum uint, coverage sc.CoverageMap,
	schema sc.Schema) float64 {
	//
	return float64(coverage.CoverageOf(mid, name)[casenum].Count())
}

func expandedBranchesSummariser(mid uint, name string, casenum uint, coverage sc.CoverageMap,
	schema sc.Schema) float64 {
	//
	if c := findConstraint(mid, name, casenum, schema); c != nil {
		return float64(c.Branches())
	} else {
		module := schema.Modules().Nth(mid)
		log.Errorf("unknown constraint \"%s.%s#%d\" in coverage report", module, name, casenum)
	}
	// Done
	return 0
}

func expandedPercentSummariser(mid uint, name string, casenum uint, coverage sc.CoverageMap, schema sc.Schema) float64 {
	val := expandedCoverageSummariser(mid, name, casenum, coverage, schema)
	total := expandedBranchesSummariser(mid, name, casenum, coverage, schema)
	percent := float64(val*100) / float64(total)
	//
	return percent
}

func determineConstraintBranches(mid uint, name string, coverage sc.CoverageMap, schema sc.Schema) uint {
	var branches uint = 0
	// Extract available coverage data
	bitsets := coverage.CoverageOf(mid, name)
	//
	for i := range bitsets {
		// Lookup actual constraint
		c := findConstraint(mid, name, uint(i), schema)
		if c == nil {
			module := schema.Modules().Nth(mid)
			panic(fmt.Sprintf("unknown constraint \"%s.%s#%d\" in coverage report", module, name, i))
		}
		//
		branches += c.Branches()
	}
	// Done
	return branches
}

func findConstraint(mid uint, name string, casenum uint, schema sc.Schema) sc.Constraint {
	// Identify constraint
	index, ok := schema.Constraints().Find(func(c sc.Constraint) bool {
		n, m := c.Name()
		return c.Contexts()[0].Module() == mid && n == name && m == casenum
	})
	// Check whether we found it (or not)
	if ok {
		return schema.Constraints().Nth(index)
	}
	// Nope, failed to find it
	return nil
}

//nolint:errcheck
func init() {
	rootCmd.AddCommand(coverageCmd)
	coverageCmd.Flags().Bool("debug", false, "enable debugging constraints")
	coverageCmd.Flags().BoolP("module", "m", false, "show module summaries")
	coverageCmd.Flags().BoolP("expand", "e", false, "show expanded constraints")
	coverageCmd.Flags().StringP("filter", "f", "", "regex constraint filter")
}
