package sexp

import (
	"fmt"
)

// SymbolRule is a symbol generator is responsible for converting a terminating
// expression (i.e. a symbol) into an expression type T.  For
// example, a number or a column access.
type SymbolRule[T comparable] func(string) (T, bool, error)

// ListRule is a list translator is responsible converting a list with a given
// sequence of zero or more arguments into an expression type T.
// Observe that the arguments are already translated into the correct
// form.
type ListRule[T comparable] func(*List) (T, error)

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
	// Maps S-Expressions to their spans in the original source file.  This is
	// used to build the new source map.
	old_srcmap *SourceMap[SExp]
	// Maps translated expressions to their spans in the original source file.
	// This is constructed using the old source map.
	new_srcmap *SourceMap[T]
}

// NewTranslator constructs a new Translator instance.
func NewTranslator[T comparable](srcmap *SourceMap[SExp]) *Translator[T] {
	return &Translator[T]{
		lists:      make(map[string]ListRule[T]),
		symbols:    make([]SymbolRule[T], 0),
		old_srcmap: srcmap,
		new_srcmap: NewSourceMap[T](srcmap.text),
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
	p.lists[name] = func(l *List) (T, error) {
		var (
			empty T
			err   error
		)
		// Translate arguments
		args := make([]T, len(l.Elements)-1)
		for i, s := range l.Elements[1:] {
			args[i], err = translateSExp(p, s)
			// Handle error
			if err != nil {
				return empty, err
			}
		}
		// Apply constructor
		term, err := t(args)
		// Check for error
		if err == nil {
			return term, nil
		}
		// Add span information
		return empty, p.SyntaxError(l, err.Error())
	}
}

// AddBinaryRule .
func (p *Translator[T]) AddBinaryRule(name string, t BinaryRule[T]) {
	var empty T
	//
	p.lists[name] = func(l *List) (T, error) {
		if len(l.Elements) != 3 {
			// Should be unreachable.
			return empty, p.SyntaxError(l, "Incorrect number of arguments")
		}

		lhs, ok1 := any(l.Elements[1]).(*Symbol)
		rhs, ok2 := any(l.Elements[2]).(*Symbol)

		var msg string

		if ok1 && ok2 {
			term, err := t(lhs.Value, rhs.Value)
			if err == nil {
				return term, nil
			}
			// Adorn error
			msg = err.Error()
		} else {
			msg = fmt.Sprintf("Binary list malformed (%t,%t)", ok1, ok2)
		}
		// error
		return empty, p.SyntaxError(l, msg)
	}
}

// AddSymbolRule adds a new symbol translator to this expression translator.
func (p *Translator[T]) AddSymbolRule(t SymbolRule[T]) {
	p.symbols = append(p.symbols, t)
}

// SyntaxError constructs a suitable syntax error for a given S-Expression.
func (p *Translator[T]) SyntaxError(s SExp, msg string) error {
	// Get span of enclosing list
	span := p.old_srcmap.Get(s)
	// This should be unreachable.
	return NewSyntaxError(span, msg)
}

// ===================================================================
// Private
// ===================================================================

// Translate an S-Expression into an IR expression.  Observe that
// this can still fail in the event that the given S-Expression does
// not describe a well-formed IR expression.
func translateSExp[T comparable](p *Translator[T], s SExp) (T, error) {
	var empty T

	switch e := s.(type) {
	case *List:
		return translateSExpList[T](p, e)
	case *Symbol:
		for i := 0; i != len(p.symbols); i++ {
			ir, ok, err := (p.symbols[i])(e.Value)
			if ok && err != nil {
				// Transform into syntax error
				return empty, p.SyntaxError(s, err.Error())
			} else if ok {
				return ir, nil
			}
		}
	}
	// This should be unreachable.
	return empty, p.SyntaxError(s, "invalid s-expression")
}

// Translate a list of S-Expressions into a unary, binary or n-ary
// expression of some kind.  This type of expression is determined by
// the first element of the list.  The remaining elements are treated
// as arguments which are first recursively translated.
func translateSExpList[T comparable](p *Translator[T], l *List) (T, error) {
	var empty T
	// Sanity check this list makes sense
	if len(l.Elements) == 0 || l.Elements[0].AsSymbol() == nil {
		return empty, p.SyntaxError(l, "invalid list")
	}
	// Extract expression name
	name := (l.Elements[0].(*Symbol)).Value
	// Lookup appropriate translator
	t := p.lists[name]
	// Check whether we found one.
	if t != nil {
		return (t)(l)
	}
	// Default fall back
	return empty, p.SyntaxError(l, "unknown list encountered")
}
