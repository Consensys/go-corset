package picus

import (
	"fmt"
	"io"
	"strings"

	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// WriteTo implements io.WriterTo for Program, emitting the program in Lisp form
// to w and returning the number of bytes written. Module iteration follows Go's
// map iteration order; if deterministic output is required, sort pp.Modules first.
func (pp *Program[F]) WriteTo(w io.Writer) (int64, error) {
	var total int64

	n, err := fmt.Fprintf(w, "(prime-number %v)\n", pp.Prime)
	total += int64(n)
	if err != nil {
		return total, err
	}

	for _, m := range pp.Modules {
		wn, err := m.WriteTo(w)
		total += wn
		if err != nil {
			return total, err
		}
	}
	return total, nil
}

// WriteTo implements io.WriterTo for Module, emitting the module in Lisp form
// (begin-module, inputs, outputs, constraints, end-module) using the provided
// S-expression formatter. It returns the number of bytes written.
func (m *Module[F]) WriteTo(w io.Writer) (int64, error) {
	formatter := sexp.NewFormatter(130)
	var total int64

	n, err := fmt.Fprintf(w, "(begin-module %s)\n", m.Name)
	total += int64(n)
	if err != nil {
		return total, err
	}

	// Inputs
	for _, in := range m.Inputs {
		wn, err := writeFormatted(w, formatter, sexp.NewList([]sexp.SExp{
			sexp.NewSymbol("input"),
			in.Lisp(),
		}))
		total += wn
		if err != nil {
			return total, err
		}
	}

	// Outputs
	for _, out := range m.Outputs {
		wn, err := writeFormatted(w, formatter, sexp.NewList([]sexp.SExp{
			sexp.NewSymbol("output"),
			out.Lisp(),
		}))
		total += wn
		if err != nil {
			return total, err
		}
	}

	// Constraints
	for _, c := range m.Constraints {
		wn, err := writeFormatted(w, formatter, c.Lisp())
		total += wn
		if err != nil {
			return total, err
		}
	}

	n, err = io.WriteString(w, "(end-module)\n")
	total += int64(n)
	return total, err
}

// writeFormatted formats the given S-expression (formatter includes a newline)
// and writes it to w, returning the number of bytes written.
func writeFormatted(w io.Writer, f *sexp.Formatter, s sexp.SExp) (int64, error) {
	str := f.Format(s) // includes newline
	n, err := io.WriteString(w, str)
	return int64(n), err
}

// Convenience: String uses WriteTo.
func (pp *Program[F]) String() string {
	var b strings.Builder
	_, _ = pp.WriteTo(&b)
	return b.String()
}
func (m *Module[F]) String() string {
	var b strings.Builder
	_, _ = m.WriteTo(&b)
	return b.String()
}
