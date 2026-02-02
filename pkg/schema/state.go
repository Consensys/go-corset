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

import (
	"math/big"

	"github.com/consensys/go-corset/pkg/schema/register"
)

type State interface {
	IsTerminal() bool
	Terminate()
	Goto(pc uint)
	// In(bus Bus)
	Outputs() []big.Int
	Internal() []big.Int
	Load(reg register.Id) *big.Int
	LoadN(registers []register.Id) []big.Int
	// Out(bus Bus)
	Terminated() bool
	Pc() uint
	Registers() []register.Register
	Store(reg register.Id, value big.Int)
	StoreAcross(value big.Int, registers ...register.Id)
	StoreN(registers []register.Id, values []big.Int)
	String() string
}
// Clone() State
// EmptyState(pc uint, registers []register.Register, io Map) State
//	NewState(state []big.Int, registers []register.Register, io Map) State
//	InitialState(inputs []big.Int, registers []register.Register, buses []Bus, iomap Map) State