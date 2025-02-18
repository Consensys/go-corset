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
package coverage

import (
	"fmt"
	"os"
	"regexp"

	sc "github.com/consensys/go-corset/pkg/schema"
)

// Filter defines the type of a constraint filter.
type Filter func(sc.Constraint, sc.Schema) bool

// DefaultFilter is a very simple filter which accepts everything.
func DefaultFilter() Filter {
	// The default filter eliminates any constraints which have only a single
	// branch, as these simply dilute the outcome.
	return func(_ sc.Constraint, _ sc.Schema) bool {
		return true
	}
}

// UnitBranchFilter filters out all constraints which only have a single branch
// (i.e. because these simply dilute the overall results).
func UnitBranchFilter(filter Filter) Filter {
	return and(filter, func(c sc.Constraint, schema sc.Schema) bool {
		return c.Branches() != 1
	})
}

// RegexFilter is a another simple filter which applies a regex to the full
// constraint name.
func RegexFilter(filter Filter, regexStr string) Filter {
	if regexStr == "" {
		return filter
	}
	//
	regex, err := regexp.Compile(regexStr)
	//
	if err != nil {
		fmt.Printf("invalid filter: %s", err)
		os.Exit(0)
	}
	//
	return and(filter, func(c sc.Constraint, schema sc.Schema) bool {
		mid := c.Contexts()[0].Module()
		name, num := c.Name()
		// Determine module name
		modName := schema.Modules().Nth(mid).Name
		// Construct qualified name for constraint
		name = fmt.Sprintf("%s.%s#%d", modName, name, num)
		// See whether it matches, or not.
		return regex.MatchString(name)
	})
}

func and(lhs Filter, rhs Filter) Filter {
	return func(c sc.Constraint, schema sc.Schema) bool {
		return lhs(c, schema) && rhs(c, schema)
	}
}
