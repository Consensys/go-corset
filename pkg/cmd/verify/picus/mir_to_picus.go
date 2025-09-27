package picus

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/consensys/go-corset/pkg/cmd/verify/picus/pcl"
	"github.com/consensys/go-corset/pkg/ir/mir"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/util/field"
)

// `PicusTranslator` captures any state needed to lower an MIR schema to a program in PCL (Picus Constraint Language).
// Core structure is borrowed from `lower.go` which converts MIR to AIR
type PicusTranslator[F field.Element[F]] struct {
	// Modules we are lowering from
	mirSchema mir.Schema[F]
	// Picus program we are building
	picusProgram *pcl.Program[F]
	// Tracks the current path conditions during lowering. This is necessary because of the
	// ⊥ symbol which is a constraint indicating that the path condition at the location is false
	pathConditions []pcl.Formula[F]
}

func modulusOf[F field.Element[F]]() *big.Int {
	var z F
	return z.Modulus()
}

// Constructor
func NewPicusTranslator[F field.Element[F]](mirSchema mir.Schema[F]) *PicusTranslator[F] {
	return &PicusTranslator[F]{
		mirSchema:      mirSchema,
		picusProgram:   pcl.NewProgram[F](modulusOf[F]()),
		pathConditions: make([]pcl.Formula[F], 0),
	}
}

// Translate the provided MIR Schema to a Picus Program.
func (p *PicusTranslator[F]) Translate() *pcl.Program[F] {
	// Perform a 1-1 translation between MIR modules and Picus modules
	var i uint
	for i = 0; i < p.mirSchema.Width(); i++ {
		p.TranslateModule(i)
	}
	return p.picusProgram
}

// Compiles an MIR Module to a Picus Module.
// i denotes the index of the MIR module in the schema
func (p *PicusTranslator[F]) TranslateModule(i uint) {
	// get the MIR module
	mirModule := p.mirSchema.Module(i)
	if mirModule.IsSynthetic() {
		return
	}
	fmt.Printf("mir module %s\n", mirModule.Name())
	// initialize the corresponding PCL module
	picusModule := p.picusProgram.AddModule(mirModule.Name())
	// register inputs and outputs from MIR inputs/outputs
	for _, register := range mirModule.Registers() {
		picusVar := pcl.V[F](register.Name)
		if register.IsInput() {
			picusModule.AddInput(picusVar)
		} else if register.IsOutput() {
			picusModule.AddOutput(picusVar)
		}
	}

	// build PCL constraints from MIR constraints
	for iter := mirModule.Constraints(); iter.HasNext(); {
		constraint := iter.Next().(mir.Constraint[F])
		p.translateConstraint(constraint, picusModule, mirModule)
	}
}

// Translates MIR constraints into PCL constraints. The built constraints are implicitly added to `picusModule`
func (p *PicusTranslator[F]) translateConstraint(c mir.Constraint[F], picusModule *pcl.Module[F], mirModule schema.Module[F]) {
	// Check what kind of constraint we have
	switch v := c.Unwrap().(type) {
	case mir.RangeConstraint[F]:
		p.translateRangeConstraint(v, picusModule, mirModule)
	case mir.VanishingConstraint[F]:
		p.translateVanishing(v, picusModule, mirModule)
	default:
		panic(fmt.Sprintf("Unhandled constraint: %s", c.Unwrap()))
	}
}

// Translates a MIR range constraint `r` to a PCL less than constraint.
func (p *PicusTranslator[F]) translateRangeConstraint(r mir.RangeConstraint[F], picusModule *pcl.Module[F], mirModule schema.Module[F]) {
	expr := p.lowerTerm(r.Expr, mirModule)
	// 1. Get the `big.Int` representation of the max unisgned value for a given bitwidth.
	// 2. Create a field element from the big integer.
	// 3. Construct a PCL constant from the field element.
	upperBound := pcl.C(field.BigInt[F](*MaxValueBig(int(r.Bitwidth))))
	// Add (assert (<= `expr` `upperBound`))
	picusModule.AddLeqConstraint(expr, upperBound)
}

// Translates an MIR vanishing constraint into one or more PCL constraints.
func (p *PicusTranslator[F]) translateVanishing(v mir.VanishingConstraint[F], picusModule *pcl.Module[F], mirModule schema.Module[F]) {
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
func (p *PicusTranslator[F]) logicalTermToConstraints(t mir.LogicalTerm[F], mirModule schema.Module[F]) []pcl.Constraint[F] {
	switch e := t.(type) {
	case *mir.Conjunct[F]:
		return p.conjunctToConstraints(e, mirModule)
	case *mir.Equal[F]:
		return p.equalToConstraints(e, mirModule)
	case *mir.Ite[F]:
		return p.iteToConstraints(e, mirModule)
	case *mir.Disjunct[F]:
		return p.disjunctToConstraints(e, mirModule)
	default:
		panic(fmt.Sprintf("Unhandled term: %v", e))
	}
}

// Translates an MIR if-then-else constraint into a PCL ite constraint.
func (p *PicusTranslator[F]) iteToConstraints(ite *mir.Ite[F], mirModule schema.Module[F]) []pcl.Constraint[F] {
	// Generate a Picus formula that corresponds to the ite guard.
	picusCondition := p.logicalTermToFormula(ite.Condition, mirModule)
	// Convert the `ite.TrueBranch` and add the guard to our path conditions.
	p.pathConditions = append(p.pathConditions, picusCondition)
	constraintsTrueBranch := p.logicalTermToConstraints(ite.TrueBranch, mirModule)
	// Drop the constraint from the current path condition.
	p.pathConditions = p.pathConditions[:len(p.pathConditions)-1]
	// Add the negation of the guard to the path conditions to process the False branch.
	p.pathConditions = append(p.pathConditions, pcl.NewNot[F](picusCondition))
	// Convert the `ite.FalseBranch`.
	constraintsFalseBranch := p.logicalTermToConstraints(ite.FalseBranch, mirModule)
	// Drop the negation of the guard from path condition
	p.pathConditions = p.pathConditions[:len(p.pathConditions)-1]
	return []pcl.Constraint[F]{
		pcl.NewIfElse(picusCondition, constraintsTrueBranch, constraintsFalseBranch),
	}
}

// Translate an equality MIR constraint to a PCL equality constraint
func (p *PicusTranslator[F]) equalToConstraints(eq *mir.Equal[F], mirModule schema.Module[F]) []pcl.Constraint[F] {
	return []pcl.Constraint[F]{pcl.NewPicusConstraint[F](p.convertEqual(eq, mirModule))}
}

// Translates an MIR conjunction (/\ m_1 .. m_n) to a collection of Picus constraints [p_1, ..., p_n]
// Right now Picus works better if the constraints are split into different assertions so we do that here.
func (p *PicusTranslator[F]) conjunctToConstraints(conj *mir.Conjunct[F], mirModule schema.Module[F]) []pcl.Constraint[F] {
	picusConstraints := make([]pcl.Constraint[F], 0)
	for _, conjunct := range conj.Args {
		picusConstraints = append(picusConstraints, p.logicalTermToConstraints(conjunct, mirModule)...)
	}
	return picusConstraints
}

// Translates an MIR disjunction (V m_1 ... m_n) to to a Picus disjunction. There is a special case that needs to be handled for
// an empty disjunction which corresponds to ⊥. This code assumes that constraint only appears in `if-then-else`'s.
func (p *PicusTranslator[F]) disjunctToConstraints(disj *mir.Disjunct[F], mirModule schema.Module[F]) []pcl.Constraint[F] {
	if len(disj.Args) == 0 {
		if len(p.pathConditions) == 0 {
			panic("Empty disjunct should correspond to impossible path condition. As such, path conditions cannot be empty")
		}

		return []pcl.Constraint[F]{
			pcl.NewPicusConstraint[F](pcl.NewNot[F](pcl.FoldAnd(p.pathConditions))),
		}
	}
	picusConstraints := make([]pcl.Constraint[F], 1)
	picusConstraints[0] = pcl.NewPicusConstraint[F](p.convertDisjunct(disj, mirModule))
	return picusConstraints
}

// Converts an MIR LogicalTerm to a Picus LogicalTerm
func (p *PicusTranslator[F]) logicalTermToFormula(t mir.LogicalTerm[F], mirModule schema.Module[F]) pcl.Formula[F] {
	switch e := t.(type) {
	case *mir.Conjunct[F]:
		return p.convertConjunct(e, mirModule)
	case *mir.Disjunct[F]:
		return p.convertDisjunct(e, mirModule)
	case *mir.Equal[F]:
		return p.convertEqual(e, mirModule)
	case *mir.NotEqual[F]:
		return p.converNotEqual(e, mirModule)
	default:
		panic(fmt.Sprintf("Unhandled term: %v", e))
	}
}

// Converts an MIR conjunction to a PCL conjunction formula.
func (p *PicusTranslator[F]) convertConjunct(conj *mir.Conjunct[F], mirModule schema.Module[F]) pcl.Formula[F] {
	return pcl.FoldAnd(p.logicalTermsToFormulas(conj.Args, mirModule))
}

// Converts an MIR disjunction to a PCL conjunction formula.
func (p *PicusTranslator[F]) convertDisjunct(conj *mir.Disjunct[F], mirModule schema.Module[F]) pcl.Formula[F] {
	return pcl.FoldOr(p.logicalTermsToFormulas(conj.Args, mirModule))
}

// Converts an MIR equals to a PCL equals formula
func (p *PicusTranslator[F]) convertEqual(eq *mir.Equal[F], mirModule schema.Module[F]) pcl.Formula[F] {
	lhs := p.lowerTerm(eq.Lhs, mirModule)
	rhs := p.lowerTerm(eq.Rhs, mirModule)
	return pcl.NewEq[F](lhs, rhs)
}

// Converts an MIR not equals to a PCL not equals formula
func (p *PicusTranslator[F]) converNotEqual(neq *mir.NotEqual[F], mirModule schema.Module[F]) pcl.Formula[F] {
	lhs := p.lowerTerm(neq.Lhs, mirModule)
	rhs := p.lowerTerm(neq.Rhs, mirModule)
	return pcl.NewNeq[F](lhs, rhs)
}

// Converts an array of MIR terms to an array of PCL terms
func (p *PicusTranslator[F]) logicalTermsToFormulas(terms []mir.LogicalTerm[F], mirModule schema.Module[F]) []pcl.Formula[F] {
	nterms := make([]pcl.Formula[F], 0)

	for i := range len(terms) {
		nterms = append(nterms, p.logicalTermToFormula(terms[i], mirModule))
	}

	return nterms
}

// Converts an MIR term to a PCL expression
func (p *PicusTranslator[F]) lowerTerm(t mir.Term[F], module schema.Module[F]) pcl.Expr[F] {
	switch e := t.(type) {
	case *mir.Add[F]:
		args := p.lowerTerms(e.Args, module)
		return pcl.FoldBinaryE(pcl.Add, args)
	case *mir.Constant[F]:
		return pcl.C(e.Value)
	case *mir.RegisterAccess[F]:
		name := module.Register(e.Register).Name
		if strings.Contains(name, " ") {
			name = fmt.Sprintf("\"%s\"", name)
		}
		if e.Shift != 0 {
			name = fmt.Sprintf("%s_%d", name, e.Shift)
		}
		return pcl.V[F](name)
	case *mir.Mul[F]:
		args := p.lowerTerms(e.Args, module)
		return pcl.FoldBinaryE(pcl.Mul, args)
	case *mir.Sub[F]:
		args := p.lowerTerms(e.Args, module)
		return pcl.FoldBinaryE(pcl.Sub, args)
	default:
		panic(fmt.Sprintf("unknown MIR expression \"%v\"", e))
	}
}

// Lower a set of zero or more MIR expressions.
func (p *PicusTranslator[F]) lowerTerms(exprs []mir.Term[F], mirModule schema.Module[F]) []pcl.Expr[F] {
	nexprs := make([]pcl.Expr[F], len(exprs))

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
