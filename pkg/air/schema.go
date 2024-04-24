package air

import (
	"github.com/consensys/go-corset/pkg/trace"
)

type Schema = trace.Schema[Column,Constraint]

type Column = trace.Column
