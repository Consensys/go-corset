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
	"math/big"
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
	// Attempt to find module
	mid, ok := schema.Modules().Find(func(m sc.Module) bool { return m.Name == mod.Name })
	// Sanity check we found it
	if !ok {
		panic(fmt.Sprintf("unable to find module %s", mod.Name))
	}
	//
	generateJavaClassHeader(classname, builder)
	// construct builder for within the claass
	generateJavaModuleHeader(builder.Indent())
	generateJavaModuleRegisterFields(mid, schema, builder.Indent())
	generateJavaModuleHeaders(mid, schema, builder.Indent())
	generateJavaModuleConstructor(classname, mid, schema, builder.Indent())
	generateJavaModuleSize(builder.Indent())
	generateJavaModuleColumnSetters(mid, schema, builder.Indent())
	// Generate any submodules
	for _, submod := range mod.Submodules {
		if !submod.Virtual {
			generateJavaModule(submod.Name, submod, schema, builder.Indent())
		}
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

func generateJavaModuleRegisterFields(mid uint, schema *hir.Schema, builder indentBuilder) {
	for iter := schema.InputColumns(); iter.HasNext(); {
		column := iter.Next()
		// Check whether this is part of our module
		if column.Context.Module() == mid {
			// Yes, it is.
			builder.WriteIndentedString("private final MappedByteBuffer ", column.Name, ";\n")
		}
	}
	//
	builder.WriteString("\n")
}

func generateJavaModuleHeaders(mid uint, schema *hir.Schema, builder indentBuilder) {
	innerBuilder := builder.Indent()
	//
	builder.WriteIndentedString("static List<ColumnHeader> headers(int length) {\n")
	innerBuilder.WriteIndentedString("List<ColumnHeader> headers = new ArrayList<>();\n")
	//
	for iter := schema.InputColumns(); iter.HasNext(); {
		column := iter.Next()
		// Check whether this is part of our module
		if column.Context.Module() == mid {
			width := fmt.Sprintf("%d", column.DataType.ByteWidth())
			// Yes, it is.
			innerBuilder.WriteIndentedString("headers.add(new ColumnHeader(\"", column.Name, "\", ", width, ", length));\n")
		}
	}
	//
	innerBuilder.WriteIndentedString("return headers;\n")
	builder.WriteIndentedString("}\n\n")
}

func generateJavaModuleConstructor(classname string, mid uint, schema *hir.Schema, builder indentBuilder) {
	index := 0
	innerBuilder := builder.Indent()
	//
	builder.WriteIndentedString("public ", classname, "(List<MappedByteBuffer> buffers) {\n")
	//
	for iter := schema.InputColumns(); iter.HasNext(); {
		column := iter.Next()
		// Check whether this is part of our module
		if column.Context.Module() == mid {
			indexStr := fmt.Sprintf("%d", index)
			// Yes, it is.
			innerBuilder.WriteIndentedString("this.", column.Name, " = buffers.get(", indexStr, ");\n")
			//
			index++
		}
	}
	//
	builder.WriteIndentedString("}\n\n")
}

func generateJavaModuleSize(builder indentBuilder) {
	innerBuilder := builder.Indent()
	//
	builder.WriteIndentedString("public int size() {\n")
	innerBuilder.WriteIndentedString("if(!filled.isEmpty()) {\n")
	innerBuilder.WriteIndentedString(
		"   throw new RuntimeException(\"Cannot measure a trace with a non-validated row.\");\n")
	innerBuilder.WriteIndentedString("}\n")
	innerBuilder.WriteIndentedString("\n")
	innerBuilder.WriteIndentedString("return this.currentLine;\n")
	builder.WriteIndentedString("}\n\n")
}

func generateJavaModuleColumnSetters(mid uint, schema *hir.Schema, builder indentBuilder) {
	index := uint(0)
	//
	for iter := schema.InputColumns(); iter.HasNext(); {
		column := iter.Next()
		// Check whether this is part of our module
		if column.Context.Module() == mid {
			generateJavaModuleColumnSetter(index, column.Name, column.DataType.BitWidth(), builder)
			//
			index++
		}
	}
}

func generateJavaModuleColumnSetter(index uint, name string, bitwidth uint, builder indentBuilder) {
	indexStr := fmt.Sprintf("%d", index)
	typeStr := getJavaType(bitwidth)
	i1Builder := builder.Indent()
	i2Builder := i1Builder.Indent()
	//
	builder.WriteIndentedString("public int ", name, "(final ", typeStr, " val) {\n")
	i1Builder.WriteIndentedString("if(filled.get(", indexStr, ")) {\n")
	i2Builder.WriteIndentedString("throw new IllegalStateException(\"", name, " already set\")\n")
	i1Builder.WriteIndentedString("} else {\n")
	i2Builder.WriteIndentedString("filled.set(", indexStr, ");\n")
	i1Builder.WriteIndentedString("}\n\n")
	//
	switch {
	case bitwidth == 1:
		i1Builder.WriteIndentedString(name, ".put((byte) (val ? 1 : 0));\n")
	case bitwidth <= 8:
		i1Builder.WriteIndentedString(name, ".put(val.toByte());\n")
	case bitwidth <= 63:
		generateJavaModuleLongPutter(name, bitwidth, i1Builder)
	default:
		i1Builder.WriteIndentedString("???")
	}
	//
	builder.WriteIndentedString("}\n\n")
}

func generateJavaModuleLongPutter(name string, bitwidth uint, builder indentBuilder) {
	i1Builder := builder.Indent()
	builder.WriteIndentedString("if(val >= ", maxValueStr(bitwidth), "L) {\n")
	i1Builder.WriteIndentedString("throw new IllegalArgumentException(\"", name+", has invalid value (\" + val + \")\")\n")
	builder.WriteIndentedString("}\n")
	//
	for bitwidth > 0 {
		bitwidth -= 8
		builder.WriteIndentedString(name, ".put((byte) (val >> ", fmt.Sprintf("%d", bitwidth), "));\n")
	}
}

func maxValueStr(bitwidth uint) string {
	val := big.NewInt(2)
	val.Exp(val, big.NewInt(int64(bitwidth)), nil)
	//
	return val.String()
}

func getJavaType(bitwidth uint) string {
	switch {
	case bitwidth == 1:
		return "boolean"
	case bitwidth <= 8:
		return "UnsignedByte"
	case bitwidth <= 63:
		return "long"
	default:
		return "Bytes"
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

func (p *indentBuilder) WriteString(raw string) {
	p.builder.WriteString(raw)
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
