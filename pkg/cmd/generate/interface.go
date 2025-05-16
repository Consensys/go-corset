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

	"github.com/consensys/go-corset/pkg/corset"
)

// JavaTraceInterface generates a suitable interface capturing the given schema,
// as outlined in the source map.
func JavaTraceInterface(filename string, pkgname string, srcmap *corset.SourceMap) (string, error) {
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
	generateInterfaceContents(classname, srcmap.Root, indentBuilder{0, &builder})

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
}

func generateInterfaceContents(className string, mod corset.SourceModule, builder indentBuilder) {
	builder.WriteIndentedString("public interface ", className, " {\n")
	//
	generateInterfaceColumnSetters(className, mod, builder.Indent())
	// Generate any submodules
	for _, submod := range mod.Submodules {
		if !submod.Virtual {
			generateInterfaceContents(toPascalCase(submod.Name), submod, builder.Indent())
		} else {
			generateInterfaceColumnSetters(className, submod, builder.Indent())
		}
	}
	//
	builder.WriteIndentedString("}\n")
}

func generateInterfaceColumnSetters(className string, mod corset.SourceModule,
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
}
