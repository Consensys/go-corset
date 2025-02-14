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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/consensys/go-corset/pkg/binfile"
	"github.com/consensys/go-corset/pkg/corset"
	"github.com/consensys/go-corset/pkg/hir"
	sc "github.com/consensys/go-corset/pkg/schema"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate [flags] constraint_file(s)",
	Short: "generate suitable Java class(es) for integration.",
	Long:  `Generate suitable Java class(es) for integration with a Java-based tracer generator.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Configure log level
		if GetFlag(cmd, "verbose") {
			log.SetLevel(log.DebugLevel)
		}
		//
		stdlib := !GetFlag(cmd, "no-stdlib")
		filename := GetString(cmd, "output")
		pkgname := GetString(cmd, "package")
		// Parse constraints
		binf := ReadConstraintFiles(stdlib, false, false, args)
		// Sanity check debug information is available.
		srcmap, srcmap_ok := binfile.GetAttribute[*corset.SourceMap](binf)
		//
		if !srcmap_ok {
			fmt.Printf("constraints file(s) \"%s\" missing source map", args[1])
		}
		// Generate appropriate Java source
		source, err := generateJavaIntegration(filename, pkgname, srcmap, binf)
		// check for errors / write out file.
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		} else if err := os.WriteFile(filename, []byte(source), 0644); err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	},
}

// Generate a suitable tracefile integration.
func generateJavaIntegration(filename string, pkgname string, srcmap *corset.SourceMap,
	binfile *binfile.BinaryFile) (string, error) {
	//
	var builder strings.Builder
	// Extract base of filename
	basename := filepath.Base(filename)
	// Sanity check a request is made to generate a java source file.
	if !strings.HasSuffix(basename, ".java") {
		return "", errors.New("invalid Java classname")
	}
	// Strip suffix to determine classname
	classname := strings.TrimSuffix(basename, ".java")
	// begin generation
	generateJavaHeader(pkgname, &builder)
	generateJavaModule(classname, srcmap.Root, &binfile.Schema, indentBuilder{0, &builder})
	//
	return builder.String(), nil
}

func generateJavaHeader(pkgname string, builder *strings.Builder) {
	builder.WriteString(license)
	// Write package line
	if pkgname != "" {
		builder.WriteString(fmt.Sprintf("package %s;\n", pkgname))
	}
	//
	builder.WriteString(javaImports)
	builder.WriteString(javaWarning)
}

func generateJavaModule(classname string, mod corset.SourceModule, schema *hir.Schema, builder indentBuilder) {
	generateJavaClassHeader(classname, builder)
	// construct builder for within the claass
	generateJavaModuleHeader(builder.Indent())
	generateJavaRegisterDeclarations(mod, schema, builder.Indent())
	// Generate any submodules
	for _, submod := range mod.Submodules {
		generateJavaModule(submod.Name, submod, schema, builder.Indent())
	}
	//
	generateJavaClassFooter(builder)
}

func generateJavaClassHeader(classname string, builder indentBuilder) {
	builder.WriteIndentedString("public class ", classname, " {\n")
}

func generateJavaClassFooter(builder indentBuilder) {
	builder.WriteIndentedString("}\n")
}

func generateJavaModuleHeader(builder indentBuilder) {
	builder.WriteIndentedString("private final BitSet filled = new BitSet();\n")
	builder.WriteIndentedString("private int currentLine = 0;\n\n")
}

func generateJavaRegisterDeclarations(mod corset.SourceModule, schema *hir.Schema, builder indentBuilder) {
	mid, ok := schema.Modules().Find(func(m sc.Module) bool { return m.Name == mod.Name })
	//
	if !ok {
		panic(fmt.Sprintf("unable to find module %s", mod.Name))
	}
	//
	for iter := schema.InputColumns(); iter.HasNext(); {
		column := iter.Next()
		// Check whether this is part of our module
		if column.Context.Module() == mid {
			// Yes, it is.
			builder.WriteIndentedString("private final MappedByteBuffer ", column.Name, ";\n")
		}
	}
}

// A string builder which supports indentation.
type indentBuilder struct {
	indent  uint
	builder *strings.Builder
}

func (p *indentBuilder) Indent() indentBuilder {
	return indentBuilder{p.indent + 1, p.builder}
}

func (p *indentBuilder) WriteIndentedString(pieces ...string) {
	p.WriteIndent()
	//
	for _, s := range pieces {
		p.builder.WriteString(s)
	}
}

func (p *indentBuilder) WriteIndent() {
	for i := uint(0); i < p.indent; i++ {
		p.builder.WriteString("   ")
	}
}

const license string = `// Copyright Consensys Software Inc.
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
`

const javaWarning string = `
/**
 * WARNING: This code is generated automatically.
 *
 * <p>Any modifications to this code may be overwritten and could lead to unexpected behavior.
 * Please DO NOT ATTEMPT TO MODIFY this code directly.
 */
`

const javaImports string = `
import java.math.BigInteger;
import java.nio.MappedByteBuffer;
import java.util.ArrayList;
import java.util.BitSet;
import java.util.List;

import net.consensys.linea.zktracer.ColumnHeader;
import net.consensys.linea.zktracer.types.UnsignedByte;
import org.apache.tuweni.bytes.Bytes;
`

//nolint:errcheck
func init() {
	rootCmd.AddCommand(generateCmd)
	generateCmd.Flags().StringP("output", "o", "Trace.java", "specify output file.")
	generateCmd.Flags().StringP("package", "p", "", "specify Java package.")
}
