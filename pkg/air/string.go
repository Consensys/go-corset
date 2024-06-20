package air

import (
	"fmt"
)

func (e *ColumnAccess) String() string {
	if e.Shift == 0 {
		return fmt.Sprintf("#%d", e.Column)
	}

	return fmt.Sprintf("(shift #%d %d)", e.Column, e.Shift)
}

func (e *Constant) String() string {
	return e.Value.String()
}

func (e *Add) String() string {
	return naryString("+", e.Args)
}

func (e *Sub) String() string {
	return naryString("-", e.Args)
}

func (e *Mul) String() string {
	return naryString("*", e.Args)
}

func naryString(operator string, exprs []Expr) string {
	// This should be generalised and moved into common?
	var rs string

	for _, e := range exprs {
		es := e.String()
		rs = fmt.Sprintf("%s %s", rs, es)
	}

	return fmt.Sprintf("(%s%s)", operator, rs)
}
