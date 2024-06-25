package constraint

import (
	"errors"
	"fmt"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
)

// PermutationConstraint declares a constraint that one column is a permutation
// of another.
type PermutationConstraint struct {
	// The target columns
	targets []uint
	// The source columns
	sources []uint
}

// NewPermutationConstraint creates a new permutation
func NewPermutationConstraint(targets []uint, sources []uint) *PermutationConstraint {
	if len(targets) != len(sources) {
		panic("differeng number of target / source permutation columns")
	}

	return &PermutationConstraint{targets, sources}
}

// RequiredSpillage returns the minimum amount of spillage required to ensure
// valid traces are accepted in the presence of arbitrary padding.
func (p *PermutationConstraint) RequiredSpillage() uint {
	return uint(0)
}

// Accepts checks whether a permutation holds between the source and
// target columns.
func (p *PermutationConstraint) Accepts(tr trace.Trace) error {
	// Slice out data
	src := sliceColumns(p.sources, tr)
	dst := sliceColumns(p.targets, tr)
	// Sanity check whether column exists
	if !util.ArePermutationOf(dst, src) {
		msg := fmt.Sprintf("Target columns (%v) not permutation of source columns ({%v})",
			p.targets, p.sources)
		return errors.New(msg)
	}
	// Success
	return nil
}

func (p *PermutationConstraint) String() string {
	targets := ""
	sources := ""

	for i, s := range p.targets {
		if i != 0 {
			targets += " "
		}

		targets += fmt.Sprintf("%d", s)
	}

	for i, s := range p.sources {
		if i != 0 {
			sources += " "
		}

		sources += fmt.Sprintf("%d", s)
	}

	return fmt.Sprintf("(permutation (%s) (%s))", targets, sources)
}

func sliceColumns(columns []uint, tr trace.Trace) [][]*fr.Element {
	// Allocate return array
	cols := make([][]*fr.Element, len(columns))
	// Slice out the data
	for i, n := range columns {
		nth := tr.ColumnByIndex(n)
		cols[i] = nth.Data()
	}
	// Done
	return cols
}
