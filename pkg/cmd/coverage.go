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
	"strings"

	"github.com/consensys/go-corset/pkg/binfile"
	cov "github.com/consensys/go-corset/pkg/cmd/coverage"
	"github.com/consensys/go-corset/pkg/corset"
	"github.com/consensys/go-corset/pkg/mir"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/set"
	"github.com/consensys/go-corset/pkg/util/termio"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var coverageCmd = &cobra.Command{
	Use:   "coverage [flags] coverage_file constraint_file(s)",
	Short: "query coverage data generated for a given set of constraints.",
	Long:  `Provides mechanisms for investigating the coverage data generated for a given set of constraints and traces.`,
	Run: func(cmd *cobra.Command, args []string) {
		var (
			cfg          coverageConfig
			corsetConfig corset.CompilationConfig
		)
		//
		if len(args) < 2 {
			fmt.Println(cmd.UsageString())
			os.Exit(1)
		}
		// Configure log level
		if GetFlag(cmd, "verbose") {
			log.SetLevel(log.DebugLevel)
		}
		optimisation := GetUint(cmd, "opt")
		// Set optimisation level
		if optimisation >= uint(len(mir.OPTIMISATION_LEVELS)) {
			fmt.Printf("invalid optimisation level %d\n", optimisation)
			os.Exit(2)
		}
		optConfig := mir.OPTIMISATION_LEVELS[optimisation]
		//
		filter := cov.DefaultFilter()
		//
		corsetConfig.Stdlib = !GetFlag(cmd, "no-stdlib")
		corsetConfig.Debug = GetFlag(cmd, "debug")
		corsetConfig.Legacy = GetFlag(cmd, "legacy")
		corsetConfig.StrictMode = GetFlag(cmd, "strict")
		cfg.diff = GetFlag(cmd, "diff")
		expand := GetFlag(cmd, "expand")
		module := GetFlag(cmd, "module")
		includes := GetStringArray(cmd, "include")
		cfg.titles = GetStringArray(cmd, "titles")
		// Apply unit branch filter
		filter = cov.UnitBranchFilter(filter)
		// Apply regex filter
		cfg.filter = cov.RegexFilter(filter, GetString(cmd, "filter"))
		//
		json, others := splitArgs(args)
		// Parse constraints
		binfile := ReadConstraintFiles(corsetConfig, others)
		// Parse coverage file
		coverage := readCoverageReports(json, binfile, optConfig)
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
		cfg.calcs = determineIncludedCalcs(cov.DEFAULT_CALCS, includes)
		// Determine relevant set of constraint identifiers
		cfg.groups, cfg.depth = determineConstraintGroups(mode, mirSchema)
		//printCoverage(ids, calcs, []sc.CoverageMap{coverage[0]}, airSchema)
		printCoverage(cfg, coverage[1], mirSchema)
		//printCoverage(ids, calcs, []sc.CoverageMap{coverage[2]}, hirSchema)
	},
}

type coverageConfig struct {
	depth uint
	// Titles to use for each report
	titles []string
	// Determines how constraints are grouped (e.g. by module, etc)
	groups []cov.ConstraintGroup
	// Filter to use for selecting constraints.
	filter cov.Filter
	// Determines which metrics to show (e.g. coverage only, or actually branch
	// counts, etc)
	calcs []cov.ColumnCalc
	// Determines whether coverage diff mode is enabled.
	diff bool
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

func determineIncludedCalcs(calcs []cov.ColumnCalc, includes []string) []cov.ColumnCalc {
	var included []cov.ColumnCalc
	// Handle case where no includes provided
	if includes == nil {
		return calcs
	}
	// Otherwise construct calcs to be included
	for _, inc := range includes {
		// Look for matching calc
		index := util.FindMatching(calcs, func(calc cov.ColumnCalc) bool {
			return inc == calc.Name
		})
		// Check whether is included (or not)
		if index == math.MaxUint {
			fmt.Printf("unknown metric \"%s\"\n", inc)
			os.Exit(4)
		}
		// Append calc
		included = append(included, calcs[index])
	}
	//
	return included
}

func determineConstraintGroups(mode uint, schema sc.Schema) ([]cov.ConstraintGroup, uint) {
	switch mode {
	case MODULE_MODE:
		return determineModuleGroups(schema), 1
	case CONSTRAINT_MODE:
		return determineUnexpandedGroups(schema), 2
	case EXPANDED_MODE:
		return determineExpandedGroups(schema), 3
	}
	//
	panic("unreachable")
}

func determineModuleGroups(schema sc.Schema) []cov.ConstraintGroup {
	var groups []cov.ConstraintGroup
	// Determine how many modules
	n := schema.Modules().Count()
	//
	for i := uint(0); i < n; i++ {
		groups = append(groups, cov.NewModuleGroup(i))
	}
	//
	return groups
}

func determineUnexpandedGroups(schema sc.Schema) []cov.ConstraintGroup {
	var groups []cov.ConstraintGroup
	// Determine how many modules
	n := schema.Modules().Count()
	//
	for mid := uint(0); mid < n; mid++ {
		names := set.NewSortedSet[string]()
		// Construct set of unique names
		for iter := schema.Constraints(); iter.HasNext(); {
			ith := iter.Next()
			if ith.Contexts()[0].Module() == mid {
				name, _ := ith.Name()
				names.Insert(name)
			}
		}
		// Construct group for each unique name
		for _, n := range *names {
			groups = append(groups, cov.NewConstraintGroup(mid, n))
		}
	}
	//
	return groups
}

func determineExpandedGroups(schema sc.Schema) []cov.ConstraintGroup {
	var groups []cov.ConstraintGroup
	//
	for iter := schema.Constraints(); iter.HasNext(); {
		ith := iter.Next()
		mid := ith.Contexts()[0].Module()
		name, num := ith.Name()
		groups = append(groups, cov.NewIndividualConstraintGroup(mid, name, num))
	}
	//
	return groups
}

func readCoverageReports(filenames []string, binf *binfile.BinaryFile,
	optConfig mir.OptimisationConfig) [3][]sc.CoverageMap {
	//
	var maps [3][]sc.CoverageMap
	//
	maps[0] = make([]sc.CoverageMap, len(filenames))
	maps[1] = make([]sc.CoverageMap, len(filenames))
	maps[2] = make([]sc.CoverageMap, len(filenames))
	//
	for i, n := range filenames {
		tmp := readCoverageReport(n, binf, optConfig)
		maps[0][i] = tmp[0]
		maps[1][i] = tmp[1]
		maps[2][i] = tmp[2]
	}
	//
	return maps
}

func printCoverage(cfg coverageConfig,
	// Distinct coverage reports to show side-by-side
	coverages []sc.CoverageMap,
	// Schema which defines what constraints are available, etc.
	schema sc.Schema) {
	//
	var (
		// Determine number of calculated columns per map
		n = len(cfg.calcs)
		// Total number of calculated columns
		m = uint(len(coverages) * n)
	)
	// Initialise row
	var rows [][]string
	//
	if len(cfg.titles) > 0 {
		reportTitles := make([]string, m+cfg.depth)
		//
		for i := range coverages {
			for j := range cfg.calcs {
				offset := uint((i * n) + j)
				//
				if i < len(cfg.titles) {
					reportTitles[offset+cfg.depth] = cfg.titles[i]
				}
			}
		}
		// Append row
		rows = append(rows, reportTitles)
	}
	// Configure report & calc titles
	calcTitles := make([]string, m+cfg.depth)
	//
	for i := range coverages {
		for j, s := range cfg.calcs {
			offset := uint((i * n) + j)
			calcTitles[offset+cfg.depth] = s.Name
		}
	}
	//
	rows = append(rows, calcTitles)
	//
	for _, grp := range cfg.groups {
		// Determine constraints to summarise on this row.
		constraints := grp.Select(schema, cfg.filter)
		// Only generate row if there are matching constraints
		if len(constraints) > 0 {
			row := make([]string, cfg.depth)
			// Initialise row title
			row[0] = schema.Modules().Nth(grp.ModuleId).Name
			// Initialise name column
			if cfg.depth >= 2 {
				row[1] = grp.Name
			}
			// Initialise case column
			if cfg.depth >= 3 {
				row[2] = fmt.Sprintf("%d", grp.Case)
			}
			// Build up reports
			for _, coverage := range coverages {
				// determine columns for this coverage map
				crow := coverageRow(constraints, cfg.calcs, coverage, schema)
				//
				row = append(row, crow...)
			}
			//
			rows = append(rows, row)
		}
	}
	// Print matching entries
	tbl := util.NewTablePrinter(m+cfg.depth, uint(len(rows)))
	//
	for i, row := range rows {
		tbl.SetRow(uint(i), row...)
	}
	//
	setTitleColours(tbl, cfg, coverages)
	//
	if cfg.diff {
		setDiffColours(tbl, cfg, coverages)
	}
	//
	tbl.SetMaxWidth(1, 64)
	//
	tbl.Print()
}

func coverageRow(constraints []sc.Constraint, calcs []cov.ColumnCalc, cov sc.CoverageMap, schema sc.Schema) []string {
	row := make([]string, len(calcs))
	//
	for i, calc := range calcs {
		value := calc.Constructor(constraints, cov, schema)
		row[i] = value.String()
	}
	// Done
	return row
}

func setTitleColours(tbl *util.TablePrinter, cfg coverageConfig, covs []sc.CoverageMap) {
	escape := termio.NewAnsiEscape().FgColour(termio.TERM_BLUE).Build()
	n := uint(1)
	// Check for report titles
	if len(cfg.titles) > 0 {
		n++
	}
	// Constraint groups
	for i := n; i < tbl.Height(); i++ {
		for j := uint(0); j < cfg.depth; j++ {
			tbl.SetEscape(j, i, escape)
		}
	}
	// Calcs
	for i := uint(0); i < n; i++ {
		for j := uint(0); j < uint(len(cfg.calcs)*(len(covs))); j++ {
			tbl.SetEscape(j+1, i, escape)
		}
	}
}

func setDiffColours(tbl *util.TablePrinter, cfg coverageConfig, covs []sc.CoverageMap) {
	n := uint(1)
	// Check for report titles
	if len(cfg.titles) > 0 {
		n++
	}
	//
	escape := termio.NewAnsiEscape().Fg256Colour(102).Build()
	white := termio.BoldAnsiEscape().FgColour(termio.TERM_YELLOW).Build()
	// Set all columns to hidden
	for i := n; i < tbl.Height(); i++ {
		for j := uint(0); j < uint(len(covs)*len(cfg.calcs)); j++ {
			tbl.SetEscape(cfg.depth+j, i, escape)
		}
	}
	//
	for i := uint(1); i < tbl.Height(); i++ {
		for j := 1; j < len(covs); j++ {
			for k := uint(0); k < uint(len(cfg.calcs)); k++ {
				cur := cfg.depth + k + uint(j*len(cfg.calcs))
				prev := cfg.depth + k + uint((j-1)*len(cfg.calcs))
				//
				if tbl.Get(prev, i) != tbl.Get(cur, i) {
					tbl.SetEscape(prev, i, white)
					tbl.SetEscape(cur, i, white)
				}
			}
		}
	}
}

//nolint:errcheck
func init() {
	rootCmd.AddCommand(coverageCmd)
	coverageCmd.Flags().Bool("debug", false, "enable debugging constraints")
	coverageCmd.Flags().Bool("diff", false, "highlight differences between coverage reports")
	coverageCmd.Flags().BoolP("module", "m", false, "show module summaries")
	coverageCmd.Flags().BoolP("expand", "e", false, "show expanded constraints")
	coverageCmd.Flags().StringP("filter", "f", "", "regex constraint filter")
	coverageCmd.Flags().StringArrayP("include", "i", []string{"covered", "branches", "coverage"},
		"specify information to include in report")
	coverageCmd.Flags().StringArrayP("titles", "t", nil,
		"specify report titles")
}
