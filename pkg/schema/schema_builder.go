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
package schema

type SchemaBuilder[C Constraint] struct {
	modmap      map[string]uint
	modules     []ModuleBuilder
	constraints []C
}

func NewSchemaBuilder[C Constraint]() SchemaBuilder[C] {
	panic("todo")
}

func (p *SchemaBuilder[C]) NewModule(name string) uint {
	panic("todo")
}

func (p *SchemaBuilder[C]) Module(mid uint) *ModuleBuilder {
	return &p.modules[mid]
}

func (p *SchemaBuilder[C]) ModuleOf(name string) *ModuleBuilder {
	panic("todo")
}

func (p *SchemaBuilder[C]) AddConstraint(constraint C) {
	panic("todo")
}

func (p *SchemaBuilder[C]) Build() Schema[C] {
	panic("todo")
}

type ModuleBuilder struct {
	name    string
	columns []Column
}

func (p *ModuleBuilder) NewColumn(column Column) uint {
	panic("todo")
}
