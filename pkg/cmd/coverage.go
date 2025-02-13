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

	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/bit"
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
		//
		stdlib := !GetFlag(cmd, "no-stdlib")
		debug := GetFlag(cmd, "debug")
		legacy := GetFlag(cmd, "legacy")
		filter := regexFilter(GetString(cmd, "filter"))
		// Parse constraints
		binfile := ReadConstraintFiles(stdlib, debug, legacy, args[1:])
		// Parse coverage file
		coverage := readCoverageReport(args[0], binfile)
		//
		hirSchema := &binfile.Schema
		mirSchema := hirSchema.LowerToMir()
		airSchema := mirSchema.LowerToAir()
		//
		printCoverage(coverage[0], airSchema, filter)
		printCoverage(coverage[1], mirSchema, filter)
		printCoverage(coverage[2], hirSchema, filter)
	},
}

func regexFilter(filter string) func(string, string) bool {
	regex, err := regexp.Compile(filter)
	//
	if err != nil {
		fmt.Printf("invalid filter: %s", err)
		os.Exit(0)
	}
	//
	return func(m string, n string) bool {
		name := fmt.Sprintf("%s.%s", m, n)
		return regex.MatchString(name)
	}
}

func printCoverage(coverage sc.CoverageMap, schema sc.Schema, filter func(string, string) bool) {
	// Determine how many modules there are
	nModules := schema.Modules().Count()
	//
	if !coverage.IsEmpty() {
		var n = uint(len(coverageSummarisers))
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
			rs := generateModuleCoverage(mid, coverage, schema, filter)
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
}

func generateModuleCoverage(mid uint, coverage sc.CoverageMap, schema sc.Schema,
	filter func(string, string) bool) [][]string {
	//
	var (
		n     = uint(len(coverageSummarisers))
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

	for iter := coverage.KeysOf(mid).Iter(); iter.HasNext(); {
		// Determine constraint name
		name := iter.Next()
		// Filter out columns
		if filter(mod.Name, name) {
			// Construct row
			vals := make([]string, n+2)
			if len(rows) == 0 {
				vals[0] = title
			}
			//
			vals[1] = name
			// Apply summarisers
			for i, fn := range coverageSummarisers {
				vals[i+2] = fn.summary(mid, name, coverage, schema)
			}
			//
			rows = append(rows, vals)
		}
	}
	//
	return rows
}

type coverageSummariser struct {
	name    string
	summary func(uint, string, sc.CoverageMap, sc.Schema) string
}

var coverageSummarisers []coverageSummariser = []coverageSummariser{
	{"Coverage", constraintCoverageSummariser},
	{"Branches", constraintBranchesSummariser},
}

func constraintCoverageSummariser(mid uint, name string, coverage sc.CoverageMap, schema sc.Schema) string {
	var aggregated bit.Set
	// Extract available coverage data
	bitsets := coverage.CoverageOf(mid, name)
	//
	for _, b := range bitsets {
		// Aggregated coverage for this case
		aggregated.Union(b)
	}
	// Done
	return fmt.Sprintf("%d", aggregated.Count())
}

func constraintBranchesSummariser(mid uint, name string, coverage sc.CoverageMap, schema sc.Schema) string {
	var branches uint = 1
	// Extract available coverage data
	bitsets := coverage.CoverageOf(mid, name)
	//
	for i := range bitsets {
		// Lookup actual constraint
		if c := findConstraint(mid, name, uint(i), schema); c != nil {
			branches *= c.Branches()
		} else {
			module := schema.Modules().Nth(mid)
			log.Errorf("unknown constraint \"%s.%s#%d\" in coverage report", module, name, i)
		}
	}
	// Done
	return fmt.Sprintf("%d", branches)
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
	coverageCmd.Flags().StringP("filter", "f", "", "regex constraint filter")
}
