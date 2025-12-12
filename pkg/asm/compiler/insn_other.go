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
package compiler

import "github.com/consensys/go-corset/pkg/asm/io/micro"

func (p *StateTranslator[F, T, E, M]) translateInOut(cc uint, codes []micro.Code) E {
	var code = codes[cc].(*micro.InOut)
	// In/Out codes are really nops from the perspective of compilation.  Their
	// primary purposes is to assist trace expansion.
	//
	// NOTE: we have to pretend that we've written registers here, otherwise
	// forwarding will not be enabled.
	p.WriteRegisters(code.RegistersWritten())
	//
	return p.translateCode(cc+1, codes)
}

func (p *StateTranslator[F, T, E, M]) translateJmp(cc uint, codes []micro.Code) E {
	var code = codes[cc].(*micro.Jmp)
	//
	return p.Goto(code.Target)
}

func (p *StateTranslator[F, T, E, M]) translateRet() E {
	return p.Terminate()
}

// Translate this instruction into low-level constraints.
func (p *StateTranslator[F, T, E, M]) translateDivision(cc uint, codes []micro.Code) E {
	var (
		code     = codes[cc].(*micro.Division)
		quot     = p.WriteRegister(code.Quotient)
		rem      = p.WriteRegister(code.Remainder)
		dividend = p.translateMicroExpr(code.Dividend)
		divisor  = p.translateMicroExpr(code.Divisor)
	)
	// FIXME: missing constraint to ensure remainder < divisor
	//
	// Construct equation by converting division into multiplication
	eqn := quot.Multiply(divisor).Add(rem).Equals(dividend)
	// Continue
	return eqn.And(p.translateCode(cc+1, codes))
}

func (p *StateTranslator[F, T, E, M]) translateMicroExpr(e micro.Expr) E {
	if e.HasSecond() {
		var c = e.Second()
		return BigNumber[T, E](&c)
	}
	//
	return p.ReadRegister(e.First())
}
