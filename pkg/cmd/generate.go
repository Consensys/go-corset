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
			generateJavaModule(toPascalCase(submod.Name), submod, schema, builder.Indent())
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
	//
	for iter := schema.InputColumns(); iter.HasNext(); {
		column := iter.Next()
		// Check whether this is part of our module
		if column.Context.Module() == mid {
			// Yes, include register
			if count == 0 {
				builder.WriteIndentedString("List<ColumnHeader> headers(int length) {\n")
				i1Builder.WriteIndentedString("List<ColumnHeader> headers = new ArrayList<>();\n")
			}
			//
			width := fmt.Sprintf("%d", column.DataType.ByteWidth())
			name := fmt.Sprintf("%s.%s", mod.Name, column.Name)
			i1Builder.WriteIndentedString("headers.add(new ColumnHeader(\"", name, "\",", width, ",length));\n")
			//
			count++
		}
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
			maxInt      = big.NewInt(math.MaxInt32)
			javaType    string
			constructor string
			fieldName   string = strings.ReplaceAll(constant.Name, "-", "_")
		)
		// Determine suitable Java type
		switch {
		case constant.Value.Sign() < 0:
			// TODO: for now, we always skip negative constants since it is
			// entirely unclear how they should be interpreted.
			continue
		case constant.Value.Cmp(maxInt) <= 0:
			constructor = fmt.Sprintf("0x%s", constant.Value.Text(16))
			javaType = "int"
		case constant.Value.IsInt64():
			constructor = fmt.Sprintf("0x%sL", constant.Value.Text(16))
			javaType = "long"
		default:
			constructor = fmt.Sprintf("new BigInteger(\"%s\")", constant.Value.String())
			javaType = "BigInteger"
		}
		//
		builder.WriteIndentedString("public static final ", javaType, " ", fieldName, " = ", constructor, ";\n")
	}
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
	builder.WriteIndentedString(fmt.Sprintf("for(int i=bs.size(); i<bs.size(); i++) { %s.put(bs.get(i)); }\n", fieldName))
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
 * Please DO NOT ATTEMPT TO MODIFY this code directly.
 */
`

const javaImports string = `
import java.io.IOException;
import java.io.RandomAccessFile;
import java.math.BigInteger;
import java.nio.MappedByteBuffer;
import java.nio.channels.FileChannel;
import java.util.ArrayList;
import java.util.BitSet;
import java.util.List;

import net.consensys.linea.zktracer.types.UnsignedByte;
import org.apache.tuweni.bytes.Bytes;
`

const javaTraceOf string = `
   /**
    * Construct a new trace which will be written to a given file.
    *
    * @param file File into which the trace will be written.  Observe any previous contents of this file will be lost.
    * @return {} object to use for writing column data.
    *
    * @throws IOException If an I/O error occurs.
    */
   public static {} of(RandomAccessFile file, ColumnHeader[] headers) throws IOException {
      long headerSize = determineHeaderSize(headers);
      long dataSize = determineHeaderSize(headers);
      file.setLength(headerSize + dataSize);
      // Write header
      writeHeader(file,headers,headerSize);
      // Initialise buffers
      MappedByteBuffer[] buffers = initialiseByteBuffers(file,headers,headerSize);
      // Done
      return new {}(buffers);
   }

   /**
    * Precompute the size of the trace file in order to memory map the buffers.
    *
    * @param headers Set of headers for the columns being written.
    * @return Number of bytes requires for the trace file header.
    */
   private static long determineHeaderSize(ColumnHeader[] headers) {
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
   private static long determineDataSize(ColumnHeader[] headers) {
      long nBytes = 0;

      for (ColumnHeader header : headers) {
         nBytes += header.length * header.bytesPerElement;
      }

      return nBytes;
   }

   /**
    * Write header information for the trace file.
    * @param file Trace file being written.
    * @param headers Column headers.
    * @param headerSize Overall size of the header.
    */
   private static void writeHeader(RandomAccessFile file, ColumnHeader[] headers, long headerSize) throws IOException {
      final var header = file.getChannel().map(FileChannel.MapMode.READ_WRITE, 0, headerSize);
      // Write column count as uint32
      header.putInt(headers.length);
      // Write column headers one-by-one
      for(ColumnHeader h : headers) {
         header.putShort((short) h.name.length());
         header.put(h.name.getBytes());
         header.put((byte) h.bytesPerElement);
         header.putInt((int) h.length);
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
   public record ColumnHeader(String name, long bytesPerElement, long length) { }
`

//nolint:errcheck
func init() {
	rootCmd.AddCommand(generateCmd)
	generateCmd.Flags().StringP("output", "o", "Trace.java", "specify output file.")
	generateCmd.Flags().StringP("package", "p", "", "specify Java package.")
}
