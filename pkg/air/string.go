package air

import (
	"fmt"
)

func (e *ColumnAccess) String() string {
	if e.Shift == 0 {
		return e.Column
	}

	return fmt.Sprintf("(shift %s %d)", e.Column, e.Shift)
}

func (e *Constant) String() string {
	return e.Value.String()
}

func (e *Add) String() string {
	return NaryString("+", e.Arguments)
}

func (e *Sub) String() string {
	return NaryString("-", e.Arguments)
}

func (e *Mul) String() string {
	return NaryString("*", e.Arguments)
}

func (e *Inverse) String() string {
	return fmt.Sprintf("(inv %s)", e.Expr)
}

func NaryString(operator string, exprs []Expr) string {
	// This should be generalised and moved into common?
	var rs string

	for _, e := range exprs {
		es := e.String()
		rs = fmt.Sprintf("%s %s", rs, es)
	}

	return fmt.Sprintf("(%s%s)", operator, rs)
}
