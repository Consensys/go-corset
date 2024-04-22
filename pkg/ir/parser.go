package ir

import (
	"errors"
	"fmt"
	"github.com/consensys/go-corset/pkg/sexp"
)

/// A symbol generator is responsible for converting a terminating
/// expression (i.e. a symbol) into an expression type T.  For
/// example, a number or a column access.
type SExpSymbolTranslator[T comparable] func(string)(T,error)

// A list generator is reponsible converting a list with a given
// sequence of zero or more arguments into an expression type T.
// Observe that the arguments are already translated into the correct
// form.
type SExpListTranslator[T comparable] func([]sexp.SExp)(T,error)

type SExpBinaryListTranslator[T comparable] func(string,string)(T,error)

type SExpRecursiveListTranslator[T comparable] func([]T)(T,error)

// A generic mechanism for translating S-Expressions into the various
// IR forms.
type IrParser[T comparable] struct {
	lists map[string]SExpListTranslator[T]
	symbols []SExpSymbolTranslator[T]
}

func NewIrParser[T comparable]() IrParser[T] {
	var ep IrParser[T]
	ep.lists = make(map[string]SExpListTranslator[T])
	ep.symbols = make([]SExpSymbolTranslator[T],0)
	return ep
}

// Add a new list translator to this expression parser.
func AddRecursiveListTranslator[T comparable](p *IrParser[T], name string, t SExpRecursiveListTranslator[T]) {
	// Construct a recursive list translator as a wrapper around a generic list translator.
	p.lists[name] = func(elements []sexp.SExp) (T,error) {
		var empty T
		var err error
		// Translate arguments
		args := make([]T,len(elements)-1)
		for i,s := range elements[1:] {
			args[i],err = parseSExp(p,s)
			if err != nil { return empty,err }
		}
		return t(args)
	}
}

func AddBinaryListTranslator[T comparable](p *IrParser[T], name string, t SExpBinaryListTranslator[T]) {
	p.lists[name] = func(elements []sexp.SExp) (T,error) {
		var empty T
		if len(elements) == 3 {
			lhs,ok1 := any(elements[1]).(*sexp.Symbol)
			rhs,ok2 := any(elements[2]).(*sexp.Symbol)
			if ok1 && ok2 {
				return t(lhs.Value,rhs.Value)
			} else {
				msg := fmt.Sprintf("Binary list malformed (%t,%t)",ok1,ok2)
				return empty,errors.New(msg)
			}
		} else {
			msg := fmt.Sprintf("Incorrect number of arguments: {%d}",len(elements)-1)
			return empty,errors.New(msg)
		}
	}
}

// Add a new symbol translator to this expression parser.
func AddSymbolTranslator[T comparable](p *IrParser[T], t SExpSymbolTranslator[T]) {
	p.symbols = append(p.symbols,t)
}

// Parse a given string into a given intermediate representation T
// using an appropriately configured IrParser.
func Parse[T comparable](p IrParser[T], s string)(T,error) {
	// Parse string into S-expression form
	e,err := sexp.Parse(s)
	if err != nil {
		var empty T
		return empty,err
	}
	// Process S-expression into AIR expression
	return parseSExp(&p, e)
}

// Translate an S-Expression into an IR expression.  Observe that
// this can still fail in the event that the given S-Expression does
// not describe a well-formed AIR expression.
func parseSExp[T comparable](p *IrParser[T], s sexp.SExp) (T,error) {
	switch e := s.(type) {
	case *sexp.List:
		return parseSExpList[T](p, e.Elements)
	case *sexp.Symbol:
		for i := 0; i!=len(p.symbols); i++ {
			ir,err := (p.symbols[i])(e.Value)
			if err == nil { return ir,err }
		}
	}
	panic("invalid S-Expression")
}

// Translate a list of S-Expressions into a unary, binary or n-ary
// expression of some kind.  This type of expression is determined by
// the first element of the list.  The remaining elements are treated
// as arguments which are first recursively translated.
func parseSExpList[T comparable](p *IrParser[T], elements []sexp.SExp) (T,error) {
	var empty T
	// Sanity check this list makes sense
	if len(elements) == 0 || !elements[0].IsSymbol() {
		return empty,errors.New("Invalid sexp.List")
	}
	// Extract expression name
	name := (elements[0].(*sexp.Symbol)).Value
	// Lookup appropriate translator
	t := p.lists[name]
	// Check whether we found one.
	if t != nil {
		return (t)(elements)
	} else {
		// Default fall back
		return empty, errors.New("unknown list encountered")
	}
}
