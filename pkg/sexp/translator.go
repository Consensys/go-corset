package sexp

import (
	"errors"
	"fmt"
)

// SymbolRule is a symbol generator is responsible for converting a terminating
// expression (i.e. a symbol) into an expression type T.  For
// example, a number or a column access.
type SymbolRule[T comparable] func(string) (T, error)

// ListRule is a list translator is responsible converting a list with a given
// sequence of zero or more arguments into an expression type T.
// Observe that the arguments are already translated into the correct
// form.
type ListRule[T comparable] func([]SExp) (T, error)

// BinaryRule is a binary translator is a wrapper for translating lists which must
// have exactly two symbol arguments.  The wrapper takes care of
// ensuring sufficient arguments are given, etc.
type BinaryRule[T comparable] func(string, string) (T, error)

// RecursiveRule is a recursive translator is a wrapper for translating lists whose
// elements can be built by recursively reusing the enclosing
// translator.
type RecursiveRule[T comparable] func([]T) (T, error)

// ===================================================================
// Parser
// ===================================================================

// Translator is a generic mechanism for translating S-Expressions into a structured
// form.
type Translator[T comparable] struct {
	lists   map[string]ListRule[T]
	symbols []SymbolRule[T]
}

// NewTranslator constructs a new Translator instance.
func NewTranslator[T comparable]() *Translator[T] {
	return &Translator[T]{
		lists:   make(map[string]ListRule[T]),
		symbols: make([]SymbolRule[T], 0),
	}
}

// ===================================================================
// Public
// ===================================================================

// ParseAndTranslate a given string into a given structured representation T
// using an appropriately configured.
func (p *Translator[T]) ParseAndTranslate(s string) (T, error) {
	// Parse string into S-expression form
	e, err := Parse(s)
	if err != nil {
		var empty T
		return empty, err
	}

	// Process S-expression into AIR expression.
	return translateSExp(p, e)
}

// Translate a given string into a given structured representation T
// using an appropriately configured.
func (p *Translator[T]) Translate(sexp SExp) (T, error) {
	// Process S-expression into target expression
	return translateSExp(p, sexp)
}

// AddRecursiveRule adds a new list translator to this expression translator.
func (p *Translator[T]) AddRecursiveRule(name string, t RecursiveRule[T]) {
	// Construct a recursive list translator as a wrapper around a generic list translator.
	p.lists[name] = func(elements []SExp) (T, error) {
		var (
			empty T
			err   error
		)
		// Translate arguments
		args := make([]T, len(elements)-1)
		for i, s := range elements[1:] {
			args[i], err = translateSExp(p, s)
			if err != nil {
				return empty, err
			}
		}

		return t(args)
	}
}

// AddBinaryRule .
func (p *Translator[T]) AddBinaryRule(name string, t BinaryRule[T]) {
	p.lists[name] = func(elements []SExp) (T, error) {
		var empty T

		if len(elements) != 3 {
			msg := fmt.Sprintf("Incorrect number of arguments: {%d}", len(elements)-1)
			return empty, errors.New(msg)
		}

		lhs, ok1 := any(elements[1]).(*Symbol)
		rhs, ok2 := any(elements[2]).(*Symbol)

		if ok1 && ok2 {
			return t(lhs.Value, rhs.Value)
		}

		msg := fmt.Sprintf("Binary list malformed (%t,%t)", ok1, ok2)

		return empty, errors.New(msg)
	}
}

// AddSymbolRule adds a new symbol translator to this expression translator.
func (p *Translator[T]) AddSymbolRule(t SymbolRule[T]) {
	p.symbols = append(p.symbols, t)
}

// ===================================================================
// Private
// ===================================================================

// Translate an S-Expression into an IR expression.  Observe that
// this can still fail in the event that the given S-Expression does
// not describe a well-formed AIR expression.
func translateSExp[T comparable](p *Translator[T], s SExp) (T, error) {
	switch e := s.(type) {
	case *List:
		return translateSExpList[T](p, e.Elements)
	case *Symbol:
		for i := 0; i != len(p.symbols); i++ {
			ir, err := (p.symbols[i])(e.Value)
			if err == nil {
				return ir, err
			}
		}
	}

	panic("invalid S-Expression")
}

// Translate a list of S-Expressions into a unary, binary or n-ary
// expression of some kind.  This type of expression is determined by
// the first element of the list.  The remaining elements are treated
// as arguments which are first recursively translated.
func translateSExpList[T comparable](p *Translator[T], elements []SExp) (T, error) {
	var empty T
	// Sanity check this list makes sense
	if len(elements) == 0 || !elements[0].IsSymbol() {
		return empty, errors.New("invalid List")
	}
	// Extract expression name
	name := (elements[0].(*Symbol)).Value
	// Lookup appropriate translator
	t := p.lists[name]
	// Check whether we found one.
	if t != nil {
		return (t)(elements)
	}

	// Default fall back
	return empty, errors.New("unknown list encountered")
}
