package hir

import (
	"fmt"
)

func (e *ColumnAccess) String() string {
	if e.Shift == 0 {
		return fmt.Sprintf("#%d", e.Column)
	}

	return fmt.Sprintf("(shift %d %d)", e.Column, e.Shift)
}

func (e *Constant) String() string {
	return e.Val.String()
}

func (e *Add) String() string {
	return naryString("+", e.Args)
}

func (e *List) String() string {
	return naryString("begin", e.Args)
}

func (e *Sub) String() string {
	return naryString("-", e.Args)
}

func (e *Mul) String() string {
	return naryString("*", e.Args)
}

func (e *Exp) String() string {
	return fmt.Sprintf("(^ %s %d)", e.Arg, e.Pow)
}

func (e *Normalise) String() string {
	return fmt.Sprintf("(~ %s)", e.Arg)
}

func (e *IfZero) String() string {
	if e.FalseBranch == nil {
		return fmt.Sprintf("(if %s %s)", e.Condition, e.TrueBranch)
	} else if e.TrueBranch == nil {
		return fmt.Sprintf("(ifnot %s %s)", e.Condition, e.FalseBranch)
	}

	return fmt.Sprintf("(if %s %s %s)", e.Condition, e.TrueBranch, e.FalseBranch)
}

func naryString(operator string, exprs []Expr) string {
	// This should be generalised and moved into common?
	rs := ""

	for _, e := range exprs {
		es := e.String()
		rs = fmt.Sprintf("%s %s", rs, es)
	}

	return fmt.Sprintf("(%s%s)", operator, rs)
}
