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
	"path"
	"path/filepath"
	"strings"

	"github.com/consensys/go-corset/pkg/cmd/generate"
	cmd_util "github.com/consensys/go-corset/pkg/cmd/util"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/util/field/bls12_377"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate [flags] constraint_file(s)",
	Short: "generate suitable Java class(es) for integration.",
	Long:  `Generate suitable Java class(es) for integration with a Java-based tracer generator.`,
	Run: func(cmd *cobra.Command, args []string) {
		var (
			source string
			err    error
			super  string
		)
		// Configure log level
		if GetFlag(cmd, "verbose") {
			log.SetLevel(log.DebugLevel)
		}
		//
		outputs := GetStringArray(cmd, "output")
		pkgname := GetString(cmd, "package")
		intrface := GetString(cmd, "interface")
		// Parse constraints
		files := splitConstraintSets(args)
		schemas := make([]cmd_util.SchemaStack[bls12_377.Element], len(files))
		//
		for i := range schemas {
			schemas[i] = *getSchemaStack[bls12_377.Element](cmd, SCHEMA_DEFAULT_AIR, files[i]...)
		}
		//
		if len(outputs) < len(schemas) {
			fmt.Println("insufficient output Java files specified.")
			os.Exit(2)
		}
		//
		if intrface != "" {
			// Attempt to write java interface
			source, err = generate.JavaTraceInterfaceUnion(intrface, pkgname, schemas)
			// check for errors
			checkError(err)
			// write out class file
			writeJavaFile(intrface, source)
			// Determine interface class name
			filename := filepath.Base(intrface)
			super = strings.TrimSuffix(filename, ".java")
		}
		//
		for i, stack := range schemas {
			var (
				filename = outputs[i]
				binf     = stack.BinaryFile()
			)
			// NOTE: assume defensive padding is enabled.
			spillage := determineSpillage(stack.LowestSchema(), true)
			// Generate appropriate Java source
			source, err = generate.JavaTraceClass(filename, pkgname, super, spillage, binf)
			// check for errors
			checkError(err)
			// write out class file
			writeJavaFile(filename, source)
		}
	},
}

func checkError(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func writeJavaFile(filename, source string) {
	if err := os.WriteFile(filename, []byte(source), 0644); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

// Attempt to figure out what the user intended from the given files.  Two easy
// cases: (1) all lisp files; (2) all binary files.  In the first case, we meant
// to generate a single binary file from the lisp files.  In the second case,
// well we just have multiple binary files.  If there's a mixture, it will abort
// for now.
func splitConstraintSets(filenames []string) [][]string {
	var (
		binfiles [][]string
		srcfiles []string
	)
	//
	for _, f := range filenames {
		if path.Ext(f) == ".lisp" {
			srcfiles = append(srcfiles, f)
		} else {
			binfiles = append(binfiles, []string{f})
		}
	}
	//
	if len(srcfiles) == 0 {
		return binfiles
	}
	//
	return append(binfiles, srcfiles)
}

// Determine spillage required for a given schema and optimisation configuration
// with (or without) defensive padding.
func determineSpillage(schema sc.AnySchema[bls12_377.Element], defensive bool) []uint {
	nModules := schema.Width()
	//
	spillage := make([]uint, nModules)
	// Iterate modules and print spillage
	for mid := uint(0); mid < nModules; mid++ {
		spillage[mid] = sc.RequiredPaddingRows(mid, defensive, schema)
	}
	//
	return spillage
}

//nolint:errcheck
func init() {
	rootCmd.AddCommand(generateCmd)
	generateCmd.Flags().StringArrayP("output", "o", nil, "specify output file(s).")
	generateCmd.Flags().StringP("interface", "i", "", "generate interface file.")
	generateCmd.Flags().StringP("package", "p", "", "specify Java package.")
}
