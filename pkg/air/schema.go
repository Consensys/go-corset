package air

import (
	"github.com/consensys/go-corset/pkg/table"
)

// Schema for AIR constraints and columns.
type Schema = table.Schema[Column, Constraint]
