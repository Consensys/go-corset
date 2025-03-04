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
	"math"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/consensys/go-corset/pkg/binfile"
	"github.com/consensys/go-corset/pkg/corset"
	"github.com/consensys/go-corset/pkg/hir"
	"github.com/consensys/go-corset/pkg/schema"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/util/collection/typed"
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
	metadata, err := binfile.Header.GetMetaData()
	// Error check
	if err != nil {
		return "", err
	}
	// begin generation
	generateJavaHeader(pkgname, metadata, &builder)
	generateJavaModule(classname, srcmap.Root, metadata, &binfile.Schema, indentBuilder{0, &builder})
	//
	return builder.String(), nil
}

func generateJavaHeader(pkgname string, metadata typed.Map, builder *strings.Builder) {
	builder.WriteString(license)
	// Write package line
	if pkgname != "" {
		builder.WriteString(fmt.Sprintf("package %s;\n", pkgname))
	}
	//
	builder.WriteString(javaImports)
	builder.WriteString(javaWarning)
	//
	if !metadata.IsEmpty() {
		// Write embedded metadata for the record.
		builder.WriteString(" * <p>Embedded Metata</p>\n")
		builder.WriteString(" * <ul>\n")
		//
		for _, k := range metadata.Keys() {
			builder.WriteString(" * <li>")
			builder.WriteString(k)
			builder.WriteString(": ")

			if v, ok := metadata.String(k); ok {
				builder.WriteString(v)
			} else {
				// NOTE: for now, we don't support nested maps.  This could be added
				// relatively straightforwardly.
				builder.WriteString("???")
			}

			builder.WriteString("</li>\n")
		}
		//
		builder.WriteString(" * </ul>\n")
	}
	//
	builder.WriteString(" */\n")
}

func generateJavaModule(className string, mod corset.SourceModule, metadata typed.Map, schema *hir.Schema,
	builder indentBuilder) {
	//
	var nFields uint
	// Attempt to find module
	mid, ok := schema.Modules().Find(func(m sc.Module) bool { return m.Name == mod.Name })
	// Sanity check we found it
	if !ok {
		panic(fmt.Sprintf("unable to find module %s", mod.Name))
	}
	// Generate what we need
	generateJavaClassHeader(mod.Name == "", className, builder)
	generateJavaModuleConstants(mod.Constants, builder.Indent())
	generateJavaModuleSubmoduleFields(mod.Submodules, builder.Indent())
	//
	if mod.Name == "" {
		generateJavaModuleMetadata(metadata, builder.Indent())
	}
	//
	generateJavaModuleHeaders(mid, mod, schema, builder.Indent())
	//
	if nFields = generateJavaModuleRegisterFields(mid, schema, builder.Indent()); nFields > 0 {
		generateJavaModuleHeader(builder.Indent())
	}
	//
	generateJavaModuleConstructor(className, mid, mod, schema, builder.Indent())
	generateJavaModuleColumnSetters(className, mod, schema, builder.Indent())

	if nFields > 0 {
		generateJavaModuleSize(builder.Indent())
		generateJavaModuleValidateRow(className, mid, mod, schema, builder.Indent())
		generateJavaModuleFillAndValidateRow(className, mid, schema, builder.Indent())
	}
	// Generate any submodules
	for _, submod := range mod.Submodules {
		if !submod.Virtual {
			generateJavaModule(toPascalCase(submod.Name), submod, metadata, schema, builder.Indent())
		} else {
			generateJavaModuleColumnSetters(className, submod, schema, builder.Indent())
		}
	}
	//
	if mod.Name == "" {
		// Write out constructor function.
		builder.WriteIndentedString(strings.ReplaceAll(javaTraceOf, "{}", className))
	}
	//
	generateJavaClassFooter(builder)
}

func generateJavaClassHeader(root bool, classname string, builder indentBuilder) {
	if root {
		builder.WriteIndentedString("public class ", classname, " {\n")
	} else {
		builder.WriteIndentedString("public static class ", classname, " {\n")
	}
}

func generateJavaClassFooter(builder indentBuilder) {
	builder.WriteIndentedString("}\n")
}

func generateJavaModuleHeaders(mid uint, mod corset.SourceModule, schema *hir.Schema, builder indentBuilder) {
	i1Builder := builder.Indent()
	// Count of created registers
	count := uint(0)
	register := uint(0)
	//
	for iter := schema.InputColumns(); iter.HasNext(); {
		column := iter.Next()
		// Check whether this is part of our module
		if column.Context.Module() == mid {
			// Yes, include register
			if count == 0 {
				builder.WriteIndentedString("public static List<ColumnHeader> headers(int length) {\n")
				i1Builder.WriteIndentedString("List<ColumnHeader> headers = new ArrayList<>();\n")
			}
			//
			width := fmt.Sprintf("%d", column.DataType.ByteWidth())
			name := fmt.Sprintf("%s.%s", mod.Name, column.Name)
			regStr := fmt.Sprintf("%d", register)
			i1Builder.WriteIndentedString(
				"headers.add(new ColumnHeader(\"", name, "\",", regStr, ",", width, ",length));\n")
			//
			count++
		}
		//
		register++
	}
	//
	if count > 0 {
		i1Builder.WriteIndentedString("return headers;\n")
		builder.WriteIndentedString("}\n\n")
	}
}

func generateJavaModuleHeader(builder indentBuilder) {
	builder.WriteIndentedString("private final BitSet filled = new BitSet();\n")
	builder.WriteIndentedString("private int currentLine = 0;\n\n")
}

func generateJavaModuleConstants(constants []corset.SourceConstant, builder indentBuilder) {
	for _, constant := range constants {
		var (
			javaType    string
			constructor string
			fieldName   string = strings.ReplaceAll(constant.Name, "-", "_")
		)
		// Determine suitable Java type
		if constant.Value.Sign() < 0 {
			// TODO: for now, we always skip negative constants since it is
			// entirely unclear how they should be interpreted.
			continue
		} else if constant.DataType != nil {
			constructor, javaType = translateJavaType(constant.DataType, constant.Value)
		} else {
			constructor, javaType = inferJavaType(constant.Value)
		}
		//
		builder.WriteIndentedString("public static final ", javaType, " ", fieldName, " = ", constructor, ";\n")
	}
}

func translateJavaType(datatype schema.Type, value big.Int) (string, string) {
	var (
		constructor string
		javaType    string
	)
	//
	if datatype.AsUint() == nil {
		// default fall back
		constructor = fmt.Sprintf("new BigInteger(\"%s\")", value.String())
		javaType = "BigInteger"
	} else {
		bitwidth := datatype.AsUint().BitWidth()
		//
		switch {
		case bitwidth < 32:
			// NOTE: cannot embed arbitrary unsigned 32bit constant into a Java
			// "int" because this type is signed.
			constructor = fmt.Sprintf("0x%s", value.Text(16))
			javaType = "int"
		case bitwidth < 64:
			// NOTE: cannot embed arbitrary unsigned 64bit constant into a Java
			// "long" because this type is signed.
			constructor = fmt.Sprintf("0x%sL", value.Text(16))
			javaType = "long"
		default:
			constructor = fmt.Sprintf("new BigInteger(\"%s\")", value.String())
			javaType = "BigInteger"
		}
	}
	//
	return constructor, javaType
}

func inferJavaType(value big.Int) (string, string) {
	var (
		maxInt      = big.NewInt(math.MaxInt32)
		constructor string
		javaType    string
	)
	//
	switch {
	case value.Cmp(maxInt) <= 0:
		constructor = fmt.Sprintf("0x%s", value.Text(16))
		javaType = "int"
	case value.IsInt64():
		constructor = fmt.Sprintf("0x%sL", value.Text(16))
		javaType = "long"
	default:
		constructor = fmt.Sprintf("new BigInteger(\"%s\")", value.String())
		javaType = "BigInteger"
	}
	//
	return constructor, javaType
}

func generateJavaModuleRegisterFields(mid uint, schema *hir.Schema, builder indentBuilder) uint {
	register := uint(0)
	// Count of created registers
	count := uint(0)
	//
	for iter := schema.InputColumns(); iter.HasNext(); {
		column := iter.Next()
		// Check whether this is part of our module
		if column.Context.Module() == mid {
			// Yes, include register
			if count == 0 {
				builder.WriteIndentedString("// Registers\n")
			}
			// Determine suitable name for field
			fieldName := toRegisterName(register, column.Name)
			//
			builder.WriteIndentedString("private final MappedByteBuffer ", fieldName, ";\n")
			// increase count
			count++
		}
		//
		register++
	}
	//
	if count > 0 {
		builder.WriteString("\n")
	}
	//
	return count
}

func generateJavaModuleSubmoduleFields(submodules []corset.SourceModule, builder indentBuilder) {
	first := true
	//
	for _, m := range submodules {
		// Only consider non-virtual modules (for now)
		if !m.Virtual {
			className := toPascalCase(m.Name)
			// Determine suitable name for field
			fieldName := toCamelCase(m.Name)
			// Start submodules section
			if first {
				builder.WriteIndentedString("// Submodules\n")
			}
			// Yes, it is.
			builder.WriteIndentedString("public final ", className, " ", fieldName, ";\n")
			//
			first = false
		}
	}
	//
	if !first {
		builder.WriteString("\n")
	}
}

func generateJavaModuleMetadata(metadata typed.Map, builder indentBuilder) {
	// Write field declaration
	builder.WriteIndentedString("public static Map<String,Object> metadata() {\n")
	// Initialise map using Java static initialiser
	if !metadata.IsEmpty() {
		i1Builder := builder.Indent()
		i1Builder.WriteIndentedString("Map<String,Object> metadata = new HashMap<>();\n")
		i1Builder.WriteIndentedString("Map<String,String> constraints = new HashMap<>();\n")

		for _, k := range metadata.Keys() {
			val, ok := metadata.String(k)

			if !ok {
				// NOTE: for now, we don't support nested maps.  This could be added
				// relatively straightforwardly.
				panic("nested metadata not currently supported")
			}
			//
			i1Builder.WriteIndentedString("constraints.put(\"", k, "\",\"", val, "\");\n")
		}
		//
		i1Builder.WriteIndentedString("metadata.put(\"constraints\",constraints);\n")
		i1Builder.WriteIndentedString("return metadata;\n")
		builder.WriteIndentedString("}\n\n")
	}
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
		// Only support non-virtual modules for now
		if !m.Virtual {
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
		var methodName string = column.Name
		//
		if !column.Computed {
			if mod.Virtual {
				methodName = toCamelCase(fmt.Sprintf("p_%s_%s", mod.Name, methodName))
			}
			//
			generateJavaModuleColumnSetter(className, methodName, column, schema, builder)
		}
	}
}

func generateJavaModuleColumnSetter(className string, methodName string, col corset.SourceColumn, schema *hir.Schema,
	builder indentBuilder) {
	//
	methodName = toCamelCase(methodName)
	bitwidth := col.DataType.BitWidth()
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
		i1Builder.WriteIndentedString(fieldName, ".put((byte) val);\n")
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
	// Legacy case for bytes
	if bitwidth == 8 {
		generateJavaModuleLegacyColumnSetter(className, methodName, builder)
	}
}

// legacy setter to support UnsignedByte.
func generateJavaModuleLegacyColumnSetter(className string, methodName string, builder indentBuilder) {
	i1Builder := builder.Indent()
	builder.WriteIndentedString("public ", className, " ", methodName, "(final UnsignedByte val) {\n")
	i1Builder.WriteIndentedString("return ", methodName, "(val.toByte());\n")
	builder.WriteIndentedString("}\n\n")
}

func generateJavaModuleLongPutter(columnName, fieldName string, bitwidth uint, builder indentBuilder) {
	n := byteWidth(bitwidth)
	i1Builder := builder.Indent()
	builder.WriteIndentedString("if(val < 0 || val >= ", maxValueStr(bitwidth), "L) {\n")
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
	builder.WriteIndentedString(fmt.Sprintf("for(int i=0; i<bs.size(); i++) { %s.put(bs.get(i)); }\n", fieldName))
}

func generateJavaModuleValidateRow(className string, mid uint, mod corset.SourceModule, schema *hir.Schema,
	builder indentBuilder) {
	//
	i1Builder := builder.Indent()
	i2Builder := i1Builder.Indent()
	register := uint(0)
	//
	builder.WriteIndentedString("public ", className, " validateRow() {\n")
	//
	for iter := schema.InputColumns(); iter.HasNext(); {
		column := iter.Next()
		// Check whether this is part of our module
		if column.Context.Module() == mid {
			name := fmt.Sprintf("%s.%s", mod.Name, column.Name)
			regstr := fmt.Sprintf("%d", register)
			// Yes, include register
			i1Builder.WriteIndentedString("if(!filled.get(", regstr, ")) {\n")
			i2Builder.WriteIndentedString("throw new IllegalStateException(\"", name, " has not been filled.\");\n")
			i1Builder.WriteIndentedString("}\n")
		}
		//
		register++
	}
	//
	i1Builder.WriteIndentedString("this.filled.clear();\n")
	i1Builder.WriteIndentedString("this.currentLine++;\n")
	i1Builder.WriteIndentedString("return this;\n")
	builder.WriteIndentedString("}\n\n")
}

func generateJavaModuleFillAndValidateRow(className string, mid uint, schema *hir.Schema, builder indentBuilder) {
	//
	i1Builder := builder.Indent()
	i2Builder := i1Builder.Indent()
	register := uint(0)
	//
	builder.WriteIndentedString("public ", className, " fillAndValidateRow() {\n")
	//
	for iter := schema.InputColumns(); iter.HasNext(); {
		column := iter.Next()
		// Check whether this is part of our module
		if column.Context.Module() == mid {
			name := toRegisterName(register, column.Name)
			regstr := fmt.Sprintf("%d", register)
			width := fmt.Sprintf("%d", column.DataType.ByteWidth())
			// Yes, include register
			i1Builder.WriteIndentedString("if(!filled.get(", regstr, ")) {\n")
			i2Builder.WriteIndentedString(name, ".position(", name, ".position() + ", width, ");\n")
			i1Builder.WriteIndentedString("}\n")
		}
		//
		register++
	}
	//
	i1Builder.WriteIndentedString("this.filled.clear();\n")
	i1Builder.WriteIndentedString("this.currentLine++;\n")
	i1Builder.WriteIndentedString("return this;\n")
	builder.WriteIndentedString("}\n\n")
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
	//
	for i, w := range splitWords(name) {
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

func splitWords(name string) []string {
	var (
		words []string
	)
	//
	for _, w1 := range strings.Split(name, "_") {
		for _, w2 := range strings.Split(w1, "-") {
			words = append(words, splitCaseChange(w2)...)
		}
	}
	//
	return words
}

func splitCaseChange(word string) []string {
	var (
		runes = []rune(word)
		words []string
		last  bool = true
		start int
	)
	//
	for i, r := range runes {
		ith := unicode.IsUpper(r)
		if !last && ith {
			// case change
			words = append(words, string(runes[start:i]))
			start = i
		}

		last = ith
	}
	// Append whatever is left
	words = append(words, string(runes[start:]))
	//
	return words
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

const license string = `/*
 * Copyright Consensys Software Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with
 * the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 *
 * SPDX-License-Identifier: Apache-2.0
 */
 `

const javaWarning string = `
/**
 * WARNING: This code is generated automatically.
 *
 * <p>Any modifications to this code may be overwritten and could lead to unexpected behavior.
 * Please DO NOT ATTEMPT TO MODIFY this code directly</p>.
 *
`

const javaImports string = `
import java.io.IOException;
import java.io.RandomAccessFile;
import java.math.BigInteger;
import java.nio.ByteBuffer;
import java.nio.MappedByteBuffer;
import java.nio.channels.FileChannel;
import java.util.ArrayList;
import java.util.BitSet;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

import net.consensys.linea.zktracer.types.UnsignedByte;
import org.apache.tuweni.bytes.Bytes;
`

// nolint
const javaTraceOf string = `
   /**
    * Construct a new trace which will be written to a given file.
    *
    * @param file File into which the trace will be written.  Observe any previous contents of this file will be lost.
    * @return Trace object to use for writing column data.
    *
    * @throws IOException If an I/O error occurs.
    */
   public static Trace of(RandomAccessFile file, List<ColumnHeader> rawHeaders, byte[] metadata) throws IOException {
      // Construct trace file header bytes
      byte[] header = constructTraceFileHeader(metadata);
      // Align headers according to register indices.
      ColumnHeader[] columnHeaders = alignHeaders(rawHeaders);
      // Determine file size
      long headerSize = determineColumnHeadersSize(columnHeaders) + header.length;
      long dataSize = determineColumnDataSize(columnHeaders);
      file.setLength(headerSize + dataSize);
      // Write headers
      writeHeaders(file,header,columnHeaders,headerSize);
      // Initialise buffers
      MappedByteBuffer[] buffers = initialiseByteBuffers(file,columnHeaders,headerSize);
      // Done
      return new Trace(buffers);
   }

  /**
   * Construct trace file header containing the given metadata bytes.
   *
   * @param metadata Metadata bytes to be embedded in the trace file.
   *
   * @return bytes making up the header.
   */
   private static byte[] constructTraceFileHeader(byte[] metadata) {
     ByteBuffer buffer = ByteBuffer.allocate(16 + metadata.length);
	 // File identifier
     buffer.put(new byte[]{'z','k','t','r','a','c','e','r'});
     // Major version
     buffer.putShort((short) 1);
     // Minor version
     buffer.putShort((short) 0);
     // Metadata length
     buffer.putInt(metadata.length);
     // Metadata
     buffer.put(metadata);
     // Done
     return buffer.array();
   }

  /**
   * Align headers ensures that the order in which columns are seen matches the order found in the trace schema.
   *
   * @param headers The headers to be aligned.
   * @return The aligned headers.
   */
   private static ColumnHeader[] alignHeaders(List<ColumnHeader> headers) {
     ColumnHeader[] alignedHeaders = new ColumnHeader[headers.size()];
     //
     for(ColumnHeader header : headers) {
       alignedHeaders[header.register] = header;
     }
     //
     return alignedHeaders;
   }

   /**
    * Precompute the size of the trace file in order to memory map the buffers.
    *
    * @param headers Set of headers for the columns being written.
    * @return Number of bytes requires for the trace file header.
    */
   private static long determineColumnHeadersSize(ColumnHeader[] headers) {
      long nBytes = 4; // column count

      for (ColumnHeader header : headers) {
        nBytes += 2; // name length
        nBytes += header.name.length();
        nBytes += 1; // byte per element
        nBytes += 4; // element count
      }

      return nBytes;
   }

   /**
    * Precompute the size of the trace file in order to memory map the buffers.
    *
    * @param headers Set of headers for the columns being written.
    * @return Number of bytes required for storing all column data, excluding the header.
    */
   private static long determineColumnDataSize(ColumnHeader[] headers) {
      long nBytes = 0;

      for (ColumnHeader header : headers) {
         nBytes += header.length * header.bytesPerElement;
      }

      return nBytes;
   }

   /**
    * Write header information for the trace file.
    *
    * @param file Trace file being written.
    * @param header Trace file header
    * @param headers Column headers.
    * @param size Overall size of the header.
    */
   private static void writeHeaders(RandomAccessFile file, byte[] header, ColumnHeader[] headers, long size) throws IOException {
      final var buffer = file.getChannel().map(FileChannel.MapMode.READ_WRITE, 0, size);
      // Write trace file header
      buffer.put(header);
      // Write column count as uint32
      buffer.putInt(headers.length);
      // Write column headers one-by-one
      for(ColumnHeader h : headers) {
         buffer.putShort((short) h.name.length());
         buffer.put(h.name.getBytes());
         buffer.put((byte) h.bytesPerElement);
         buffer.putInt((int) h.length);
      }
   }

   /**
    * Initialise one memory mapped byte buffer for each column to be written in the trace.
    * @param headers Set of headers for the columns being written.
    * @param headerSize Space required at start of trace file for header.
    * @return Buffer array with one entry per header.
    */
   private static MappedByteBuffer[] initialiseByteBuffers(RandomAccessFile file, ColumnHeader[] headers,
    long headerSize) throws IOException {
      MappedByteBuffer[] buffers = new MappedByteBuffer[headers.length];
      long offset = headerSize;
      for(int i=0;i<headers.length;i++) {
         // Determine size (in bytes) required to store all elements of this column.
         long length = headers[i].length * headers[i].bytesPerElement;
         // Preallocate space for this column.
         buffers[i] = file.getChannel().map(FileChannel.MapMode.READ_WRITE, offset, length);
         //
         offset += length;
      }
      return buffers;
   }

   /**
    * ColumnHeader contains information about a given column in the resulting trace file.
    *
    * @param name Name of the column, as found in the trace file.
    * @param bytesPerElement Bytes required for each element in the column.
    */
   public record ColumnHeader(String name, int register, long bytesPerElement, long length) { }
`

//nolint:errcheck
func init() {
	rootCmd.AddCommand(generateCmd)
	generateCmd.Flags().StringP("output", "o", "Trace.java", "specify output file.")
	generateCmd.Flags().StringP("package", "p", "", "specify Java package.")
}
