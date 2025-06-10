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

	"github.com/consensys/go-corset/pkg/asm"
	"github.com/consensys/go-corset/pkg/binfile"
	"github.com/consensys/go-corset/pkg/cmd/generate"
	"github.com/consensys/go-corset/pkg/corset"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate [flags] constraint_file(s)",
	Short: "generate suitable Java class(es) for integration.",
	Long:  `Generate suitable Java class(es) for integration with a Java-based tracer generator.`,
	Run: func(cmd *cobra.Command, args []string) {
		var (
			corsetConfig corset.CompilationConfig
			source       string
			err          error
			binfiles     []binfile.BinaryFile
			super        string
		)
		// Configure log level
		if GetFlag(cmd, "verbose") {
			log.SetLevel(log.DebugLevel)
		}
		//
		corsetConfig.Stdlib = !GetFlag(cmd, "no-stdlib")
		corsetConfig.Legacy = GetFlag(cmd, "legacy")
		asmConfig := parseLoweringConfig(cmd)
		outputs := GetStringArray(cmd, "output")
		pkgname := GetString(cmd, "package")
		intrface := GetString(cmd, "interface")
		// Parse constraints
		binfiles = readConstraintSets(corsetConfig, asmConfig, args)
		//
		if len(outputs) < len(binfiles) {
			fmt.Println("insufficient output Java files specified.")
			os.Exit(2)
		}
		//
		if intrface != "" {
			// Attempt to write java interface
			source, err = generate.JavaTraceInterfaceUnion(intrface, pkgname, "", binfiles)
			// check for errors
			checkError(err)
			// write out class file
			writeJavaFile(intrface, source)
			// Determine interface class name
			filename := filepath.Base(intrface)
			super = strings.TrimSuffix(filename, ".java")
		}
		//
		for i, bf := range binfiles {
			filename := outputs[i]
			// NOTE: assume defensive padding is enabled.
			spillage := determineConservativeSpillage(true, &bf.Schema)
			// Generate appropriate Java source
			source, err = generate.JavaTraceClass(filename, pkgname, super, spillage, &bf)
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
func readConstraintSets(corsetCfg corset.CompilationConfig, asmCfg asm.LoweringConfig,
	filenames []string) []binfile.BinaryFile {
	var binfiles []binfile.BinaryFile = make([]binfile.BinaryFile, len(filenames))
	//
	for i, f := range filenames {
		if path.Ext(f) == ".lisp" {
			binf := ReadConstraintFiles(corsetCfg, asmCfg, filenames)
			return []binfile.BinaryFile{*binf}
		}
		//
		binfiles[i] = *ReadBinaryFile(f)
		// Check we have source mapping info.
		// Sanity check debug information is available.
		if _, srcmap_ok := binfile.GetAttribute[*corset.SourceMap](&binfiles[i]); !srcmap_ok {
			fmt.Printf("constraints file(s) \"%s\" missing source map", f)
			os.Exit(1)
		}
	}
	//
	return binfiles
}

//nolint:errcheck
func init() {
	rootCmd.AddCommand(generateCmd)
	generateCmd.Flags().StringArrayP("output", "o", nil, "specify output file(s).")
	generateCmd.Flags().StringP("interface", "i", "", "generate interface file.")
	generateCmd.Flags().StringP("package", "p", "", "specify Java package.")
}
