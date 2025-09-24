package picus

import (
	"math/big"

	"github.com/consensys/go-corset/pkg/util/field"
)

//====================
// Module / Program
//====================

type Module[F field.Element[F]] struct {
	Name        string
	Inputs      []Expr[F]
	Outputs     []Expr[F]
	Constraints []Constraint[F]
}

func NewModule[F field.Element[F]](name string) *Module[F] {
	return &Module[F]{
		Name:        name,
		Inputs:      make([]Expr[F], 0),
		Outputs:     make([]Expr[F], 0),
		Constraints: make([]Constraint[F], 0),
	}
}

func (pm *Module[F]) AddInput(input Expr[F]) {
	pm.Inputs = append(pm.Inputs, input)
}

func (pm *Module[F]) AddOutput(output Expr[F]) {
	pm.Outputs = append(pm.Outputs, output)
}

func (pm *Module[F]) AddLtConstraint(op1 Expr[F], op2 Expr[F]) {
	pm.Constraints = append(pm.Constraints, NewPicusConstraint[F](NewLt[F](op1, op2)))
}

func (pm *Module[F]) AddLeqConstraint(op1 Expr[F], op2 Expr[F]) {
	pm.Constraints = append(pm.Constraints, NewPicusConstraint[F](NewLeq[F](op1, op2)))
}

func (pm *Module[F]) AddGeqConstraint(op1 Expr[F], op2 Expr[F]) {
	pm.Constraints = append(pm.Constraints, NewPicusConstraint[F](NewGeq[F](op1, op2)))
}

func (pm *Module[F]) AddEqConstraint(op1 Expr[F], op2 Expr[F]) {
	pm.Constraints = append(pm.Constraints, NewPicusConstraint[F](NewEq[F](op1, op2)))
}

type Program[F field.Element[F]] struct {
	Prime   *big.Int // field modulus as element
	Modules map[string]*Module[F]
}

func NewProgram[F field.Element[F]](prime *big.Int) *Program[F] {
	return &Program[F]{
		Prime:   prime,
		Modules: make(map[string]*Module[F]),
	}
}

func (pp *Program[F]) AddModule(moduleName string) *Module[F] {
	picusModule := NewModule[F](moduleName)
	pp.Modules[moduleName] = picusModule
	return picusModule
}
