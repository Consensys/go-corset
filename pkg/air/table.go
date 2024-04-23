package air

import (
	"github.com/consensys/go-corset/pkg/trace"
)

type Table = trace.Table[Column,Constraint]

type Column = trace.Column
