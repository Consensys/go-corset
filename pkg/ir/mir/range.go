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
package mir

import (
	"fmt"
	"math/big"
	"reflect"

	"github.com/consensys/go-corset/pkg/ir/schema"
	"github.com/consensys/go-corset/pkg/util"
)

func rangeOfTerm(e Term, module schema.Module) *util.Interval {
	switch e := e.(type) {
	case *Add:
		return rangeOfAdd(e.Args, module)
	case *Cast:
		return rangeOfCast(e.Arg, e.Range(), module)
	case *Constant:
		var c big.Int
		// Extract big integer from field element
		e.Value.BigInt(&c)
		// Return as interval
		return util.NewInterval(&c, &c)
	case *ColumnAccess:
		return rangeOfColumnAccess(e.Column, module)
	case *Exp:
		bounds := rangeOfTerm(e.Arg, module)
		bounds.Exp(uint(e.Pow))
		//
		return bounds
	case *Mul:
		return rangeOfMul(e.Args, module)
	case *Norm:
		return util.NewInterval(big.NewInt(0), big.NewInt(1))
	case *Sub:
		return rangeOfSub(e.Args, module)
	default:
		name := reflect.TypeOf(e).Name()
		panic(fmt.Sprintf("unknown MIR expression \"%s\"", name))
	}
}

func rangeOfAdd(args []Term, module schema.Module) *util.Interval {
	var res util.Interval

	for i, arg := range args {
		ith := rangeOfTerm(arg, module)
		if i == 0 {
			res.Set(ith)
		} else {
			res.Add(ith)
		}
	}
	//
	return &res
}

func rangeOfCast(arg Term, cast *util.Interval, module schema.Module) *util.Interval {
	// Compute actual interval
	res := rangeOfTerm(arg, module)
	// Check whether is within (or not)
	if res.Within(cast) {
		return res
	}
	//
	return cast
}

func rangeOfColumnAccess(column uint, module schema.Module) *util.Interval {
	bound := big.NewInt(2)
	width := int64(module.Columns().Nth(column).DataType.BitWidth())
	bound.Exp(bound, big.NewInt(width), nil)
	// Subtract 1 because interval is inclusive.
	bound.Sub(bound, big.NewInt(1))
	// Done
	return util.NewInterval(big.NewInt(0), bound)
}

func rangeOfMul(args []Term, module schema.Module) *util.Interval {
	var res util.Interval

	for i, arg := range args {
		ith := rangeOfTerm(arg, module)
		if i == 0 {
			res.Set(ith)
		} else {
			res.Mul(ith)
		}
	}
	//
	return &res
}

func rangeOfSub(args []Term, module schema.Module) *util.Interval {
	var res util.Interval

	for i, arg := range args {
		ith := rangeOfTerm(arg, module)
		if i == 0 {
			res.Set(ith)
		} else {
			res.Sub(ith)
		}
	}
	//
	return &res
}
