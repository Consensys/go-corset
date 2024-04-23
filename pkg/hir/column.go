package hir

import (
	"github.com/consensys/go-corset/pkg/mir"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
)

type Column interface {
	// hir.Column is-a Column
	trace.Column
	// Lower this column to an MirColumn
	LowerTo() mir.Column
}

type DataColumn trace.DataColumn

func NewDataColumn(name string, data []*fr.Element) *DataColumn {
	return (*DataColumn)(trace.NewDataColumn(name,data))
}

func (c *DataColumn) Name() string {
	return (*trace.DataColumn)(c).Name()
}

func (c *DataColumn) MinHeight() int {
	return (*trace.DataColumn)(c).MinHeight()
}

func (c *DataColumn) Get(row int, tr trace.Trace) (*fr.Element,error) {
	return (*trace.DataColumn)(c).Get(row,tr)
}

func (c *DataColumn) LowerTo() mir.Column {
	// FIXME: this is only temporary
	return (mir.Column)(c)
}
