package mir

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/consensys/go-corset/pkg/ir/picus"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/util/field"
)

// `PicusTranslator` captures any state needed to lower an MIR schema to a program in PCL (Picus Constraint Language).
// Core structure is borrowed from `lower.go` which converts MIR to AIR
type PicusTranslator[F field.Element[F]] struct {
	// Modules we are lowering from
	mirSchema Schema[F]
	// Picus program we are building
	picusProgram *picus.Program[F]
	// Tracks the current path conditions during lowering. This is necessary because of the
	// ⊥ symbol which is a constraint indicating that the path condition at the location is false
	pathConditions []picus.Formula[F]
}

// Constructor
func NewPicusTranslator[F field.Element[F]](mirSchema Schema[F], picusProgram *picus.Program[F]) *PicusTranslator[F] {
	return &PicusTranslator[F]{
		mirSchema:      mirSchema,
		picusProgram:   picusProgram,
		pathConditions: make([]picus.Formula[F], 0),
	}
}

// Translate the provided MIR Schema to a Picus Program.
func (p *PicusTranslator[F]) Translate() {
	// Perform a 1-1 translation between MIR modules and Picus modules
	var i uint
	for i = 0; i < p.mirSchema.Width(); i++ {
		p.TranslateModule(i)
	}
}

// Compiles an MIR Module to a Picus Module.
// i denotes the index of the MIR module in the schema
func (p *PicusTranslator[F]) TranslateModule(i uint) {
	// get the MIR module
	mirModule := p.mirSchema.Module(i)
	// initialize the corresponding PCL module
	picusModule := p.picusProgram.AddModule(mirModule.Name())
	// register inputs and outputs from MIR inputs/outputs
	for _, register := range mirModule.Registers() {
		picusVar := picus.V[F](register.Name)
		if register.IsInput() {
			picusModule.AddInput(picusVar)
		} else if register.IsOutput() {
			picusModule.AddOutput(picusVar)
		}
	}

	// build PCL constraints from MIR constraints
	for iter := mirModule.Constraints(); iter.HasNext(); {
		constraint := iter.Next().(Constraint[F])
		p.translateConstraint(constraint, picusModule, mirModule)
	}
}

// Translates MIR constraints into PCL constraints. The built constraints are implicitly added to `picusModule`
func (p *PicusTranslator[F]) translateConstraint(c Constraint[F], picusModule *picus.Module[F], mirModule schema.Module[F]) {
	// Check what kind of constraint we have
	switch v := c.constraint.(type) {
	case RangeConstraint[F]:
		p.translateRangeConstraint(v, picusModule, mirModule)
	case VanishingConstraint[F]:
		p.translateVanishing(v, picusModule, mirModule)
	default:
		panic(fmt.Sprintf("Unhandled constraint: %s", c.constraint))
	}
}

// Translates a MIR range constraint `r` to a PCL less than constraint.
func (p *PicusTranslator[F]) translateRangeConstraint(r RangeConstraint[F], picusModule *picus.Module[F], mirModule schema.Module[F]) {
	expr := p.lowerTerm(r.Expr, mirModule)
	// 1. Get the `big.Int` representation of the max unisgned value for a given bitwidth.
	// 2. Create a field element from the big integer.
	// 3. Construct a PCL constant from the field element.
	upperBound := picus.C(field.BigInt[F](*MaxValueBig(int(r.Bitwidth))))
	// Add (assert (<= `expr` `upperBound`))
	picusModule.AddLeqConstraint(expr, upperBound)
}

// Translates an MIR vanishing constraint into one or more PCL constraints.
func (p *PicusTranslator[F]) translateVanishing(v VanishingConstraint[F], picusModule *picus.Module[F], mirModule schema.Module[F]) {
	// TODO: Ask how the translation would work for a specific row i.e `v.Domain` is not empty
	if v.Domain.HasValue() {
		panic("row specific constraints are not supported!")
	}
	// We translate the logical term to a collection of Picus constraints
	picusModule.Constraints = append(picusModule.Constraints, p.logicalTermToConstraints(v.Constraint, mirModule)...)
}

// Translate a MIR logical term to one or more PCL constraints.
// Most operations have a one to one translation but for others it makes more sense to break them up into multiple Picus constraints
// For example, a vanishing constraint that is a conjunction of terms can be split into two Picus constraints.
// Picus doesn't allow ITEs to be combined with connectives i.e (/\ (= x y) /\ (if ...)) so an ITE needs to map
// to a PCL ITE constraint
func (p *PicusTranslator[F]) logicalTermToConstraints(t LogicalTerm[F], mirModule schema.Module[F]) []picus.Constraint[F] {
	switch e := t.(type) {
	case *Conjunct[F]:
		return p.conjunctToConstraints(e, mirModule)
	case *Equal[F]:
		return p.equalToConstraints(e, mirModule)
	case *Ite[F]:
		return p.iteToConstraints(e, mirModule)
	case *Disjunct[F]:
		return p.disjunctToConstraints(e, mirModule)
	default:
		panic(fmt.Sprintf("Unhandled term: %v", e))
	}
}

// Translates an MIR if-then-else constraint into a PCL ite constraint
func (p *PicusTranslator[F]) iteToConstraints(ite *Ite[F], mirModule schema.Module[F]) []picus.Constraint[F] {
	// Generate a Picus formula that corresponds to the ite guard
	picusCondition := p.logicalTermToFormula(ite.Condition, mirModule)
	// Convert the `ite.TrueBranch` and add the guard to our path conditions
	p.pathConditions = append(p.pathConditions, picusCondition)
	constraintsTrueBranch := p.logicalTermToConstraints(ite.TrueBranch, mirModule)
	p.pathConditions = p.pathConditions[:len(p.pathConditions)-1]
	p.pathConditions = append(p.pathConditions, picus.NewNot[F](picusCondition))
	constraintsFalseBranch := p.logicalTermToConstraints(ite.FalseBranch, mirModule)
	p.pathConditions = p.pathConditions[:len(p.pathConditions)-1]
	return []picus.Constraint[F]{
		picus.NewIfElse(picusCondition, constraintsTrueBranch, constraintsFalseBranch),
	}
}

// Translate an equality MIR constraint to a PCL equality constraint
func (p *PicusTranslator[F]) equalToConstraints(eq *Equal[F], mirModule schema.Module[F]) []picus.Constraint[F] {
	return []picus.Constraint[F]{picus.NewPicusConstraint[F](p.convertEqual(eq, mirModule))}
}

// Translates an MIR conjunction (/\ m_1 .. m_n) to a collection of Picus constraints [p_1, ..., p_n]
// Right now Picus works better if the constraints are split into different assertions so we do that here.
func (p *PicusTranslator[F]) conjunctToConstraints(conj *Conjunct[F], mirModule schema.Module[F]) []picus.Constraint[F] {
	picusConstraints := make([]picus.Constraint[F], 0)
	for _, conjunct := range conj.Args {
		picusConstraints = append(picusConstraints, p.logicalTermToConstraints(conjunct, mirModule)...)
	}
	return picusConstraints
}

// Translates an MIR disjunction (V m_1 ... m_n) to to a Picus disjunction. There is a special case that needs to be handled for
// an empty disjunction which corresponds to ⊥. This code assumes that constraint only appears in `if-then-else`'s.
func (p *PicusTranslator[F]) disjunctToConstraints(disj *Disjunct[F], mirModule schema.Module[F]) []picus.Constraint[F] {
	if len(disj.Args) == 0 {
		if len(p.pathConditions) == 0 {
			panic("Empty disjunct should correspond to impossible path condition. As such, path conditions cannot be empty")
		}

		return []picus.Constraint[F]{
			picus.NewPicusConstraint[F](picus.NewNot[F](picus.FoldAnd(p.pathConditions))),
		}
	}
	picusConstraints := make([]picus.Constraint[F], 1)
	picusConstraints[0] = picus.NewPicusConstraint[F](p.convertDisjunct(disj, mirModule))
	return picusConstraints
}

// Converts an MIR LogicalTerm to a Picus LogicalTerm
func (p *PicusTranslator[F]) logicalTermToFormula(t LogicalTerm[F], mirModule schema.Module[F]) picus.Formula[F] {
	switch e := t.(type) {
	case *Conjunct[F]:
		return p.convertConjunct(e, mirModule)
	case *Disjunct[F]:
		return p.convertDisjunct(e, mirModule)
	case *Equal[F]:
		return p.convertEqual(e, mirModule)
	case *NotEqual[F]:
		return p.converNotEqual(e, mirModule)
	default:
		panic(fmt.Sprintf("Unhandled term: %v", e))
	}
}

// Converts an MIR conjunction to a PCL conjunction formula.
func (p *PicusTranslator[F]) convertConjunct(conj *Conjunct[F], mirModule schema.Module[F]) picus.Formula[F] {
	return picus.FoldAnd(p.logicalTermsToFormulas(conj.Args, mirModule))
}

// Converts an MIR disjunction to a PCL conjunction formula.
func (p *PicusTranslator[F]) convertDisjunct(conj *Disjunct[F], mirModule schema.Module[F]) picus.Formula[F] {
	return picus.FoldOr(p.logicalTermsToFormulas(conj.Args, mirModule))
}

// Converts an MIR equals to a PCL equals formula
func (p *PicusTranslator[F]) convertEqual(eq *Equal[F], mirModule schema.Module[F]) picus.Formula[F] {
	lhs := p.lowerTerm(eq.Lhs, mirModule)
	rhs := p.lowerTerm(eq.Rhs, mirModule)
	return picus.NewEq[F](lhs, rhs)
}

// Converts an MIR not equals to a PCL not equals formula
func (p *PicusTranslator[F]) converNotEqual(neq *NotEqual[F], mirModule schema.Module[F]) picus.Formula[F] {
	lhs := p.lowerTerm(neq.Lhs, mirModule)
	rhs := p.lowerTerm(neq.Rhs, mirModule)
	return picus.NewNeq[F](lhs, rhs)
}

// Converts an array of MIR terms to an array of PCL terms
func (p *PicusTranslator[F]) logicalTermsToFormulas(terms []LogicalTerm[F], mirModule schema.Module[F]) []picus.Formula[F] {
	nterms := make([]picus.Formula[F], 0)

	for i := range len(terms) {
		nterms = append(nterms, p.logicalTermToFormula(terms[i], mirModule))
	}

	return nterms
}

// Converts an MIR term to a PCL expression
func (p *PicusTranslator[F]) lowerTerm(t Term[F], module schema.Module[F]) picus.Expr[F] {
	switch e := t.(type) {
	case *Add[F]:
		args := p.lowerTerms(e.Args, module)
		return picus.FoldBinaryE(picus.Add, args)
	case *Constant[F]:
		return picus.C(e.Value)
	case *RegisterAccess[F]:
		name := module.Register(e.Register).Name
		if strings.Contains(name, " ") {
			name = fmt.Sprintf("\"%s\"", name)
		}
		if e.Shift != 0 {
			name = fmt.Sprintf("%s_%d", name, e.Shift)
		}
		return picus.V[F](name)
	case *Mul[F]:
		args := p.lowerTerms(e.Args, module)
		return picus.FoldBinaryE(picus.Mul, args)
	case *Sub[F]:
		args := p.lowerTerms(e.Args, module)
		return picus.FoldBinaryE(picus.Sub, args)
	default:
		panic(fmt.Sprintf("unknown MIR expression \"%v\"", e))
	}
}

// Lower a set of zero or more MIR expressions.
func (p *PicusTranslator[F]) lowerTerms(exprs []Term[F], mirModule schema.Module[F]) []picus.Expr[F] {
	nexprs := make([]picus.Expr[F], len(exprs))

	for i := range len(exprs) {
		nexprs[i] = p.lowerTerm(exprs[i], mirModule)
	}

	return nexprs
}

// MaxValueBig returns (1<<bitwidth) - 1 as a `*big.Int`.
func MaxValueBig(bitwidth int) *big.Int {
	if bitwidth < 0 {
		panic("bitwidth must be non-negative")
	}
	if bitwidth == 0 {
		return new(big.Int) // 0
	}
	one := big.NewInt(1)
	return new(big.Int).Sub(new(big.Int).Lsh(one, uint(bitwidth)), one)
}
