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

	"github.com/consensys/go-corset/pkg/mir"
	sc "github.com/consensys/go-corset/pkg/schema"
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
		// Parse constraints
		binfile := ReadConstraintFiles(stdlib, debug, legacy, args[1:])
		// Parse coverage file
		coverage := readCoverageReport(args[0], binfile)
		//
		hirSchema := &binfile.Schema
		mirSchema := hirSchema.LowerToMir()
		airSchema := mirSchema.LowerToAir()
		//
		printCoverage(coverage[0], airSchema)
		printCoverage(coverage[1], mirSchema)
		printCoverage(coverage[2], hirSchema)
	},
}

func printCoverage(coverage sc.CoverageMap, schema sc.Schema) {
	// Determine how many modules there are
	nModules := schema.Modules().Count()
	//
	if !coverage.IsEmpty() {
		for mid := uint(0); mid < nModules; mid++ {
			printModuleCoverage(mid, coverage, schema)
		}
	}
}

func printModuleCoverage(mid uint, coverage sc.CoverageMap, schema sc.Schema) {
	mod := schema.Modules().Nth(mid)
	// Print module header
	if mod.Name != "" {
		fmt.Printf("Module %s: \n", mod.Name)
	} else {
		fmt.Printf("Prelude:\n")
	}
	// Print matching entries
	for iter := coverage.KeysOf(mid).Iter(); iter.HasNext(); {
		// Determine constraint name
		name := iter.Next()
		//
		printConstraintCoverage(mid, name, coverage, schema)
	}
	//
	fmt.Println()
}

func printConstraintCoverage(mid uint, name string, coverage sc.CoverageMap, schema sc.Schema) {
	// Find constraint in question
	constraint := findConstraint(mid, name, schema)
	// Extract coverage for this constraint
	covered := coverage.CoverageOf(mid, name)
	//
	if constraint != nil {
		// HACK
		if vc, ok := constraint.(mir.VanishingConstraint); ok {
			total := vc.Constraint.Branches()
			fmt.Printf("\t%s: %d / %d\n", name, covered.Count(), total)
		}
	} else {
		// print out data
		fmt.Printf("\t%s: %d [MISSING]\n", name, covered.Count())
	}
}

func findConstraint(mid uint, name string, schema sc.Schema) sc.Constraint {
	// Identify constraint
	index, ok := schema.Constraints().Find(func(c sc.Constraint) bool {
		return c.Contexts()[0].Module() == mid && c.Name() == name
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
}
