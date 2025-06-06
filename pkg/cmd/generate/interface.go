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
package generate

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/consensys/go-corset/pkg/binfile"
	"github.com/consensys/go-corset/pkg/corset"
)

// JavaTraceInterface generates a suitable interface capturing the given schema,
// as outlined in the source map.
func JavaTraceInterface(filename string, pkgname string, super string, isRoot bool,
	binfiles []binfile.BinaryFile) (string, error) {
	//
	var root corset.SourceModule
	// Intersect roots to determine set of common functionality.
	for i, bf := range binfiles {
		// Extract source map (which we assume is present)
		srcmap, _ := binfile.GetAttribute[*corset.SourceMap](&bf)
		//
		if i == 0 {
			root = srcmap.Root
		} else {
			root = *intersectModules(root, srcmap.Root)
		}
	}
	// Finally, generate the interface
	return generateInterface(filename, pkgname, super, isRoot, root)
}

func generateInterface(filename string, pkgname string, super string, isRoot bool,
	root corset.SourceModule) (string, error) {
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
	generateInterfaceHeader(pkgname, &builder)
	generateInterfaceContents(classname, super, isRoot, root, indentBuilder{0, &builder})

	return builder.String(), nil
}

func generateInterfaceHeader(pkgname string, builder *strings.Builder) {
	builder.WriteString(license)
	// Write package line
	if pkgname != "" {
		fmt.Fprintf(builder, "package %s;\n", pkgname)
	}
	//
	builder.WriteString(javaImports)
	builder.WriteString(javaWarning)
	//
	builder.WriteString(" */\n")
}

func generateInterfaceContents(className string, super string, isRoot bool, mod corset.SourceModule,
	builder indentBuilder) {
	//
	builder.WriteIndentedString("public interface ", className)
	//
	if super != "" {
		builder.WriteString(" extends ")
		builder.WriteString(super)
	}
	//
	builder.WriteString(" {\n")
	//
	generateInterfaceConstants(mod.Constants, builder.Indent())
	generateInterfaceSubmoduleAccessors(mod.Submodules, builder.Indent())
	generateInterfaceHeaders(builder.Indent())
	generateInterfaceColumnSetters(className, mod, builder.Indent())
	generateInterfaceValidateRow(className, builder.Indent())
	// Generate any submodules
	for _, submod := range mod.Submodules {
		if !submod.Virtual {
			subclass, subinterface := determineSubNames(submod.Name, super)
			generateInterfaceContents(subclass, subinterface, false, submod, builder.Indent())
		} else {
			generateInterfaceColumnSetters(className, submod, builder.Indent())
		}
	}
	//
	if mod.Name == "" {
		if isRoot {
			builder.WriteString(javaColumnHeader)
		}
		//
		builder.WriteString(javaAddMetadataSignature)
		builder.WriteString(javaOpenSignature)
	}
	//
	builder.WriteIndentedString("}\n")
}

func determineSubNames(modName, superName string) (string, string) {
	modName = toPascalCase(modName)
	if superName != "" {
		superName = fmt.Sprintf("%s.%s", superName, modName)
	}

	return modName, superName
}

func generateInterfaceSubmoduleAccessors(submodules []corset.SourceModule, builder indentBuilder) {
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
			builder.WriteIndentedString("public ", className, " ", fieldName, "();\n")
			//
			first = false
		}
	}
	//
	if !first {
		builder.WriteString("\n")
	}
}

func generateInterfaceHeaders(builder indentBuilder) {
	builder.WriteIndentedString("public List<ColumnHeader> headers(int length);\n")
}

func generateInterfaceConstants(constants []corset.SourceConstant, builder indentBuilder) {
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
		builder.WriteIndentedString("public final ", javaType, " ", fieldName, " = ", constructor, ";\n")
	}
	//
	builder.WriteIndentedString("public int spillage();\n")
}

func generateInterfaceColumnSetters(className string, mod corset.SourceModule,
	builder indentBuilder) {
	//
	for _, column := range mod.Columns {
		var methodName string = column.Name
		//
		if !column.Computed && !column.Internal {
			if mod.Virtual {
				methodName = toCamelCase(fmt.Sprintf("p_%s_%s", mod.Name, methodName))
			}
			//
			generateInterfaceColumnSetter(className, methodName, column, builder)
		}
	}
}

func generateInterfaceColumnSetter(className string, methodName string, col corset.SourceColumn,
	builder indentBuilder) {
	//
	methodName = toCamelCase(methodName)
	bitwidth := col.DataType.BitWidth()
	typeStr := getJavaType(bitwidth)
	//
	builder.WriteIndentedString("public ", className, " ", methodName, "(final ", typeStr, " val);\n")
	// Legacy case for bytes
	if bitwidth == 8 {
		builder.WriteIndentedString("public ", className, " ", methodName, "(final UnsignedByte val);\n")
	}
}

func generateInterfaceValidateRow(className string, builder indentBuilder) {
	//
	builder.WriteIndentedString("public ", className, " validateRow();\n")
	builder.WriteIndentedString("public ", className, " fillAndValidateRow();\n")
}
