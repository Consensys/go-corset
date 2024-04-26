package sexp

import (
	"errors"
	"fmt"
)

/// A symbol generator is responsible for converting a terminating
/// expression (i.e. a symbol) into an expression type T.  For
/// example, a number or a column access.
type SymbolRule[T comparable] func(string)(T,error)

// A list translator is reponsible converting a list with a given
// sequence of zero or more arguments into an expression type T.
// Observe that the arguments are already translated into the correct
// form.
type ListRule[T comparable] func([]SExp)(T,error)

// A binary translator is a wrapper for translating lists which must
// have exactly two symbol arguments.  The wrapper takes care of
// ensuring sufficient arguments are given, etc.
type BinaryRule[T comparable] func(string,string)(T,error)

// A recursive translator is a wrapper for translating lists whose
// elements can be built by recursively reusing the enclosing
// translator.
type RecursiveRule[T comparable] func([]T)(T,error)

// ===================================================================
// Parser
// ===================================================================

// A generic mechanism for translating S-Expressions into a structured
// form.
type Translator[T comparable] struct {
	lists map[string]ListRule[T]
	symbols []SymbolRule[T]
}

func NewTranslator[T comparable]() Translator[T] {
	var ep Translator[T]
	ep.lists = make(map[string]ListRule[T])
	ep.symbols = make([]SymbolRule[T],0)
	return ep
}

// ===================================================================
// Public
// ===================================================================

// Translate a given string into a given structured representation T
// using an appropriately configured.
func (p *Translator[T]) ParseAndTranslate(s string) (T,error) {
	// Parse string into S-expression form
	e,err := Parse(s)
	if err != nil {
		var empty T
		return empty,err
	}
	// Process S-expression into AIR expression
	return translateSExp(p, e)
}

// Translate a given string into a given structured representation T
// using an appropriately configured.
func (p *Translator[T]) Translate(sexp SExp) (T,error) {
	// Process S-expression into target expression
	return translateSExp(p, sexp)
}

// Add a new list translator to this expression translator
func (p *Translator[T]) AddRecursiveRule(name string, t RecursiveRule[T]) {
	// Construct a recursive list translator as a wrapper around a generic list translator.
	p.lists[name] = func(elements []SExp) (T,error) {
		var empty T
		var err error
		// Translate arguments
		args := make([]T,len(elements)-1)
		for i,s := range elements[1:] {
			args[i],err = translateSExp(p,s)
			if err != nil { return empty,err }
		}
		return t(args)
	}
}

func (p *Translator[T]) AddBinaryRule(name string, t BinaryRule[T]) {
	p.lists[name] = func(elements []SExp) (T,error) {
		var empty T
		if len(elements) == 3 {
			lhs,ok1 := any(elements[1]).(*Symbol)
			rhs,ok2 := any(elements[2]).(*Symbol)
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

// Add a new symbol translator to this expression translater.
func (p *Translator[T]) AddSymbolRule(t SymbolRule[T]) {
	p.symbols = append(p.symbols,t)
}

// ===================================================================
// Private
// ===================================================================

// Translate an S-Expression into an IR expression.  Observe that
// this can still fail in the event that the given S-Expression does
// not describe a well-formed AIR expression.
func translateSExp[T comparable](p *Translator[T], s SExp) (T,error) {
	switch e := s.(type) {
	case *List:
		return translateSExpList[T](p, e.Elements)
	case *Symbol:
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
func translateSExpList[T comparable](p *Translator[T], elements []SExp) (T,error) {
	var empty T
	// Sanity check this list makes sense
	if len(elements) == 0 || !elements[0].IsSymbol() {
		return empty,errors.New("Invalid List")
	}
	// Extract expression name
	name := (elements[0].(*Symbol)).Value
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
