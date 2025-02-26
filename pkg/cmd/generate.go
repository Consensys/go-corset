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

func generateJavaModule(className string, mod corset.SourceModule, schema *hir.Schema, builder indentBuilder) {
	var nFields uint
	//
	generateJavaClassHeader(className, builder)
	//
	generateJavaModuleSubmoduleFields(mod.Submodules, builder.Indent())
	//
	if !mod.Virtual {
		// Attempt to find module
		mid, ok := schema.Modules().Find(func(m sc.Module) bool { return m.Name == mod.Name })
		// Sanity check we found it
		if !ok {
			panic(fmt.Sprintf("unable to find module %s", mod.Name))
		}
		//
		if nFields = generateJavaModuleRegisterFields(mid, schema, builder.Indent()); nFields > 0 {
			generateJavaModuleHeader(builder.Indent())
		}
		//
		generateJavaModuleConstructor(className, mid, mod, schema, builder.Indent())

		if nFields > 0 {
			generateJavaModuleSize(builder.Indent())
		}

		generateJavaModuleColumnSetters(className, mod, schema, builder.Indent())
		generateJavaModuleValidateRow(className, builder.Indent())
		generateJavaModuleFillAndValidateRow(className, builder.Indent())
	} else {
		generateJavaModuleColumnSetters(className, mod, schema, builder.Indent())
	}
	// Generate any submodules
	for _, submod := range mod.Submodules {
		generateJavaModule(toPascalCase(submod.Name), submod, schema, builder.Indent())
	}
	//
	if mod.Name == "" {
		generateJavaClassStaticBuilder(className, builder.Indent())
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

func generateJavaModuleRegisterFields(mid uint, schema *hir.Schema, builder indentBuilder) uint {
	register := uint(0)
	// Count of created registers
	count := uint(0)
	//
	builder.WriteIndentedString("// Registers\n")
	//
	for iter := schema.InputColumns(); iter.HasNext(); {
		column := iter.Next()
		// Check whether this is part of our module
		if column.Context.Module() == mid {
			// Determine suitable name for field
			fieldName := toRegisterName(register, column.Name)
			// Yes, it is.
			builder.WriteIndentedString("private final MappedByteBuffer ", fieldName, ";\n")
			// increase count
			count++
		}
		//
		register++
	}
	//
	builder.WriteString("\n")
	//
	return count
}

func generateJavaModuleSubmoduleFields(submodules []corset.SourceModule, builder indentBuilder) {
	if len(submodules) > 0 {
		builder.WriteIndentedString("// Submodules\n")
		//
		for _, m := range submodules {
			className := toPascalCase(m.Name)
			// Determine suitable name for field
			fieldName := toCamelCase(m.Name)
			// Yes, it is.
			builder.WriteIndentedString("public final ", className, " ", fieldName, ";\n")
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

func generateJavaModuleConstructor(classname string, mid uint, mod corset.SourceModule,
	schema *hir.Schema, builder indentBuilder) {
	//
	register := uint(0)
	innerBuilder := builder.Indent()
	//
	builder.WriteIndentedString("private ", classname, "(MappedByteBuffer[] registers) {\n")
	innerBuilder.WriteIndentedString("// initialise register(s)\n")
	// Write register initialisers
	for iter := schema.InputColumns(); iter.HasNext(); {
		column := iter.Next()
		// Check whether this is part of our module
		if column.Context.Module() == mid {
			// Yes, it is.
			fieldName := toRegisterName(register, column.Name)
			registerStr := fmt.Sprintf("%d", register)
			innerBuilder.WriteIndentedString("this.", fieldName, " = registers[", registerStr, "];\n")
		}
		//
		register++
	}
	//
	innerBuilder.WriteIndentedString("// initialise submodule(s)\n")
	// Write submodule initialisers
	for _, m := range mod.Submodules {
		className := toPascalCase(m.Name)
		// Determine suitable name for field
		fieldName := toCamelCase(m.Name)
		// Yes, it is.
		if m.Virtual {
			innerBuilder.WriteIndentedString("this.", fieldName, " = new ", className, "();\n")
		} else {
			innerBuilder.WriteIndentedString("this.", fieldName, " = new ", className, "(registers);\n")
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

func generateJavaModuleColumnSetters(className string, mod corset.SourceModule, schema *hir.Schema,
	builder indentBuilder) {
	//
	for _, column := range mod.Columns {
		if !column.Computed {
			generateJavaModuleColumnSetter(className, column, schema, builder)
		}
	}
}

func generateJavaModuleColumnSetter(className string, col corset.SourceColumn, schema *hir.Schema,
	builder indentBuilder) {
	//
	bitwidth := col.DataType.BitWidth()
	methodName := toCamelCase(col.Name)
	fieldName := toRegisterName(col.Register, schema.Columns().Nth(col.Register).Name)
	//
	indexStr := fmt.Sprintf("%d", col.Register) // BROKEN
	typeStr := getJavaType(bitwidth)
	i1Builder := builder.Indent()
	i2Builder := i1Builder.Indent()
	//
	builder.WriteIndentedString("public ", className, " ", methodName, "(final ", typeStr, " val) {\n")
	i1Builder.WriteIndentedString("if(filled.get(", indexStr, ")) {\n")
	i2Builder.WriteIndentedString("throw new IllegalStateException(\"", col.Name, " already set\");\n")
	i1Builder.WriteIndentedString("} else {\n")
	i2Builder.WriteIndentedString("filled.set(", indexStr, ");\n")
	i1Builder.WriteIndentedString("}\n\n")
	//
	switch {
	case bitwidth == 1:
		i1Builder.WriteIndentedString(fieldName, ".put((byte) (val ? 1 : 0));\n")
	case bitwidth <= 8:
		i1Builder.WriteIndentedString(fieldName, ".put(val.toByte());\n")
	case bitwidth <= 63:
		generateJavaModuleLongPutter(col.Name, fieldName, bitwidth, i1Builder)
	default:
		generateJavaModuleBytesPutter(col.Name, fieldName, bitwidth, i1Builder)
	}
	//
	i1Builder.WriteIndentedString("\n")
	i1Builder.WriteIndentedString("return this;\n")
	// Done
	builder.WriteIndentedString("}\n\n")
}

func generateJavaModuleLongPutter(columnName, fieldName string, bitwidth uint, builder indentBuilder) {
	n := byteWidth(bitwidth)
	i1Builder := builder.Indent()
	builder.WriteIndentedString("if(val >= ", maxValueStr(bitwidth), "L) {\n")
	i1Builder.WriteIndentedString(
		"throw new IllegalArgumentException(\"", columnName+" has invalid value (\" + val + \")\");\n")
	builder.WriteIndentedString("}\n")
	//
	for i := int(n) - 1; i >= 0; i-- {
		shift := (i * 8)
		if shift == 0 {
			builder.WriteIndentedString(fieldName, ".put((byte) val);\n")
		} else {
			builder.WriteIndentedString(fieldName, ".put((byte) (val >> ", fmt.Sprintf("%d", shift), "));\n")
		}
	}
}

func generateJavaModuleBytesPutter(columnName, fieldName string, bitwidth uint, builder indentBuilder) {
	i1Builder := builder.Indent()
	n := byteWidth(bitwidth)
	//
	builder.WriteIndentedString("// Trim array to size\n")
	builder.WriteIndentedString("Bytes bs = val.trimLeadingZeros();\n")
	builder.WriteIndentedString("// Sanity check against expected width\n")
	builder.WriteIndentedString(fmt.Sprintf("if(bs.bitLength() > %d) {\n", bitwidth))
	i1Builder.WriteIndentedString(
		fmt.Sprintf("throw new IllegalArgumentException(\"%s has invalid width (\"+bs.bitLength()+\"bits)\");\n", columnName))
	builder.WriteIndentedString("}\n")
	builder.WriteIndentedString("// Write padding (if necessary)\n")
	builder.WriteIndentedString(fmt.Sprintf("for(int i=bs.size(); i<%d; i++) { %s.put((byte) 0); }\n", n, fieldName))
	builder.WriteIndentedString("// Write bytes\n")
	builder.WriteIndentedString(fmt.Sprintf("for(int i=bs.size(); i<bs.size(); i++) { %s.put(bs.get(i)); }\n", fieldName))
}

func generateJavaModuleValidateRow(className string, builder indentBuilder) {
	i1Builder := builder.Indent()
	//
	builder.WriteIndentedString("public ", className, " validateRow() {\n")
	i1Builder.WriteIndentedString("return this;\n")
	builder.WriteIndentedString("}\n\n")
}

func generateJavaModuleFillAndValidateRow(className string, builder indentBuilder) {
	i1Builder := builder.Indent()
	//
	builder.WriteIndentedString("public ", className, " fillAndValidateRow() {\n")
	i1Builder.WriteIndentedString("return this;\n")
	builder.WriteIndentedString("}\n\n")
}

func generateJavaClassStaticBuilder(className string, builder indentBuilder) {
	i1Builder := builder.Indent()
	//
	builder.WriteIndentedString("/**\n")
	builder.WriteIndentedString(" * Construct a new trace which will be written to a given file.\n")
	builder.WriteIndentedString(" **/\n")
	builder.WriteIndentedString("public static ", className, " of(RandomAccessFile file) throws IOException {\n")
	i1Builder.WriteIndentedString("return null;\n")
	builder.WriteIndentedString("}\n")
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

func toRegisterName(register uint, name string) string {
	return fmt.Sprintf("r%d_%s", register, toCamelCase(name))
}

// Capitalise each word
func toPascalCase(name string) string {
	return camelify(name, true)
}

// Capitalise each word, except first.
func toCamelCase(name string) string {
	var word string
	// Remove any invalid characters
	name = strings.ReplaceAll(name, "-", "")
	//
	for i, w := range strings.Split(name, "_") {
		if i == 0 {
			word = camelify(w, false)
		} else {
			word = fmt.Sprintf("%s%s", word, camelify(w, true))
		}
	}
	//
	return word
}

// Make all letters lowercase, and optionally capitalise the first letter.
func camelify(name string, first bool) string {
	letters := strings.Split(name, "")
	for i := range letters {
		if first && i == 0 {
			letters[i] = strings.ToUpper(letters[i])
		} else {
			letters[i] = strings.ToLower(letters[i])
		}
	}
	//
	return strings.Join(letters, "")
}

// Determine number of bytes the given bitwidth represents.
func byteWidth(bitwidth uint) uint {
	n := bitwidth / 8
	// roung up bitwidth if necessary
	if bitwidth%8 != 0 {
		return n + 1
	}
	//
	return n
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
import java.io.IOException;
import java.io.RandomAccessFile;
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
