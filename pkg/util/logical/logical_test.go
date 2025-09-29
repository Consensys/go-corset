package logical

import (
	"cmp"
	"math/big"
	"testing"

	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/source/bexp"
)

func Test_Prop_01(t *testing.T) {
	testEquivalent(t, "x==1", "x==1")
}

func Test_Prop_02(t *testing.T) {
	testEquivalent(t, "x==y", "y==x")
}

func Test_Prop_03(t *testing.T) {
	testEquivalent(t, "x!=x", "⊥")
}

func Test_Prop_04(t *testing.T) {
	testEquivalent(t, "x==x", "⊤")
}

// Conjunctions

func Test_Prop_10(t *testing.T) {
	testEquivalent(t, "⊥ ∧ ⊥", "⊥")
}

func Test_Prop_11(t *testing.T) {
	testEquivalent(t, "⊥ ∧ ⊤", "⊥")
}

func Test_Prop_12(t *testing.T) {
	testEquivalent(t, "⊤ ∧ ⊥", "⊥")
}

func Test_Prop_13(t *testing.T) {
	testEquivalent(t, "⊤ ∧ ⊤", "⊤")
}

func Test_Prop_14(t *testing.T) {
	testEquivalent(t, "x==y ∧ ⊤", "x==y")
}

func Test_Prop_15(t *testing.T) {
	testEquivalent(t, "⊤ ∧ x==y", "x==y")
}

func Test_Prop_16(t *testing.T) {
	testEquivalent(t, "x==y ∧ ⊥", "⊥")
}

func Test_Prop_17(t *testing.T) {
	testEquivalent(t, "⊥ ∧ x==y", "⊥")
}

func Test_Prop_18(t *testing.T) {
	testEquivalent(t, "x==y ∧ x==y", "x==y")
}

func Test_Prop_19(t *testing.T) {
	testEquivalent(t, "x==y ∧ x≠y", "⊥")
}

func Test_Prop_20(t *testing.T) {
	testEquivalent(t, "x==1 ∧ x==2", "⊥")
}

func Test_Prop_21(t *testing.T) {
	testEquivalent(t, "x==1 ∧ x!=2", "x==1")
}

func Test_Prop_22(t *testing.T) {
	testEquivalent(t, "x==1 ∧ x!=y", "x==1 ∧ x!=y")
}

func Test_Prop_23(t *testing.T) {
	testEquivalent(t, "x==y ∧ x!=1", "x==y ∧ x!=1")
}

func Test_Prop_24(t *testing.T) {
	testEquivalent(t, "x==y ∧ x!=z", "x==y ∧ x!=z")
}

// Disjunctions

func Test_Prop_50(t *testing.T) {
	testEquivalent(t, "⊥ ∨ ⊥", "⊥")
}

func Test_Prop_51(t *testing.T) {
	testEquivalent(t, "⊥ ∨ ⊤", "⊤")
}

func Test_Prop_52(t *testing.T) {
	testEquivalent(t, "⊤ ∨ ⊥", "⊤")
}

func Test_Prop_53(t *testing.T) {
	testEquivalent(t, "⊤ ∨ ⊤", "⊤")
}

func Test_Prop_54(t *testing.T) {
	testEquivalent(t, "x==y ∨ ⊤", "⊤")
}

func Test_Prop_55(t *testing.T) {
	testEquivalent(t, "⊤ ∨ x==y", "⊤")
}

func Test_Prop_56(t *testing.T) {
	testEquivalent(t, "x==y ∨ ⊥", "x==y")
}

func Test_Prop_57(t *testing.T) {
	testEquivalent(t, "⊥ ∨ x==y", "x==y")
}

func Test_Prop_58(t *testing.T) {
	testEquivalent(t, "x==y ∨ x==y", "x==y")
}

// func Test_Prop_59(t *testing.T) {
// 	testEquivalent(t, "x==y ∨ x≠y", "⊤")
// }

// ============================================================================
// Framework
// ============================================================================

func testEquivalent(t *testing.T, left, right string) {
	lhs := Parse(t, left)
	rhs := Parse(t, right)
	//
	if !lhs.Equals(rhs) {
		t.Errorf("not equivalent: %s=>%s but %s=>%s", left, lhs.String(id), right, rhs.String(id))
	}
}

// ============================================================================
// Parser
// ============================================================================

func Parse(t *testing.T, input string) Proposition[Var, Equality[Var]] {
	var env = func(string) bool { return true }
	// Parse input
	term, errs := bexp.Parse[LogicalTerm](input, env)
	// Sanity check errors
	if len(errs) > 0 {
		t.Errorf("internal failure: %s", errs[0].Message())
		t.FailNow()
	}
	//
	return term.prop
}

func id(x Var) string {
	return x.name
}

// Var is a wrapper around a string
type Var struct {
	name string
}

// Cmp
func (p Var) Cmp(o Var) int {
	return cmp.Compare(p.name, o.name)
}

func (p Var) String(func(string) string) string {
	return p.name
}

type LogicalTerm struct {
	expr util.Union[string, big.Int]
	prop Proposition[Var, Equality[Var]]
}

func (p LogicalTerm) Variable(v string) LogicalTerm {
	return LogicalTerm{expr: util.Union1[string, big.Int](v)}
}

func (p LogicalTerm) Number(v big.Int) LogicalTerm {
	return LogicalTerm{expr: util.Union2[string](v)}
}

func (p LogicalTerm) Or(terms ...LogicalTerm) LogicalTerm {
	term := p.prop
	//
	for _, t := range terms {
		term = term.Or(t.prop)
	}
	//
	return LogicalTerm{prop: term}
}

func (p LogicalTerm) And(terms ...LogicalTerm) LogicalTerm {
	term := p.prop
	//
	for _, t := range terms {
		term = term.And(t.prop)
	}
	//
	return LogicalTerm{prop: term}
}

func (p LogicalTerm) Truth(val bool) LogicalTerm {
	return LogicalTerm{prop: Truth[Var, Equality[Var]](val)}
}

func (p LogicalTerm) Equals(o LogicalTerm) LogicalTerm {
	var (
		atom Equality[Var]
		lhs  = Var{p.expr.First()}
	)
	// Parse rhs
	if o.expr.HasFirst() {
		// var = var
		atom = Equals(lhs, Var{o.expr.First()})
	} else {
		atom = EqualsConst(lhs, o.expr.Second())
	}
	//
	return LogicalTerm{prop: NewProposition(atom)}
}

func (p LogicalTerm) NotEquals(o LogicalTerm) LogicalTerm {
	var (
		atom Equality[Var]
		lhs  = Var{p.expr.First()}
	)
	// Parse rhs
	if o.expr.HasFirst() {
		// var = var
		atom = NotEquals(lhs, Var{o.expr.First()})
	} else {
		atom = NotEqualsConst(lhs, o.expr.Second())
	}
	//
	return LogicalTerm{prop: NewProposition(atom)}
}

func (p LogicalTerm) LessThan(LogicalTerm) LogicalTerm {
	panic("unsupported operation")
}

func (p LogicalTerm) LessThanEquals(LogicalTerm) LogicalTerm {
	panic("unsupported operation")
}

// Arithmetic
func (p LogicalTerm) Add(...LogicalTerm) LogicalTerm {
	panic("unsupported operation")
}

func (p LogicalTerm) Mul(...LogicalTerm) LogicalTerm {
	panic("unsupported operation")
}

func (p LogicalTerm) Sub(...LogicalTerm) LogicalTerm {
	panic("unsupported operation")
}
