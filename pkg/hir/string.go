package hir

import (
	"fmt"
)

func (e *ColumnAccess) String() string {
	if e.Shift == 0 {
		return e.Column
	} else {
		return fmt.Sprintf("(shift %s %d)",e.Column,e.Shift)
	}
}

func (e *Constant) String() string {
	return e.Val.String()
}

func (e *Add) String() string {
	return NaryString("+",e.Args)
}

func (e *List) String() string {
	return NaryString("begin",e.Args)
}

func (e *Sub) String() string {
	return NaryString("-",e.Args)
}

func (e *Mul) String() string {
	return NaryString("*",e.Args)
}

func (e *Normalise) String() string {
	return fmt.Sprintf("~(%s)",e.Arg)
}

func (e *IfZero) String() string {
	if e.FalseBranch == nil {
		return fmt.Sprintf("(if %s %s)",e.Condition,e.TrueBranch)
	} else if e.TrueBranch == nil {
		return fmt.Sprintf("(if %s _ %s)",e.Condition,e.FalseBranch)
	} else {
		return fmt.Sprintf("(if %s %s %s)",e.Condition,e.TrueBranch,e.FalseBranch)
	}
}

func NaryString(operator string, exprs []Expr) string {
	// This should be generalised and moved into common?
	rs := ""
	for _,e := range exprs {
		es := e.String()
		rs = fmt.Sprintf("%s %s",rs,es)
	}
	return fmt.Sprintf("(%s%s)",operator,rs)
}
