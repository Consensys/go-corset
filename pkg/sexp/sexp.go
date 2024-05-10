package sexp

// SExp is an S-Expression is either a List of zero or more S-Expressions, or
// a Symbol.
type SExp interface {
	// AsList checks whether this S-Expression is a list and, if
	// so, returns it.  Otherwise, it returns nil.
	AsList() *List
	// AsSymbol checks whether this S-Expression is a symbol and,
	// if so, returns it.  Otherwise, it returns nil.
	AsSymbol() *Symbol
	// String generates a string representation.
	String() string
}

// ===================================================================
// List
// ===================================================================

// List represents a list of zero or more S-Expressions.
type List struct {
	Elements []SExp
}

// NOTE: This is used for compile time type checking if the given type
// satisfies the given interface.
var _ SExp = (*List)(nil)

// AsList returns the given list.
func (l *List) AsList() *List { return l }

// AsSymbol returns nil for a list.
func (l *List) AsSymbol() *Symbol { return nil }

// Len gets the number of elements in this list.
func (l *List) Len() int { return len(l.Elements) }

// Get the ith element of this list
func (l *List) Get(i int) SExp { return l.Elements[i] }

func (l *List) String() string {
	var s = "("

	for i := 0; i < len(l.Elements); i++ {
		if i != 0 {
			s += ","
		}

		s += l.Elements[i].String()
	}

	s += ")"

	return s
}

// MatchSymbols matches a list which starts with at least n symbols, of which the
// first m match the given strings.
func (l *List) MatchSymbols(n int, symbols ...string) bool {
	if len(l.Elements) < n || len(symbols) > n {
		return false
	}

	for i := 0; i < len(symbols); i++ {
		switch ith := l.Elements[i].(type) {
		case *Symbol:
			if ith.Value != symbols[i] {
				return false
			}
		default:
			return false
		}
	}

	return true
}

// ===================================================================
// Symbol
// ===================================================================

// Symbol represents a terminating symbol.
type Symbol struct {
	Value string
}

// NOTE: This is used for compile time type checking if the given type
// satisfies the given interface.
var _ SExp = (*Symbol)(nil)

// AsList returns nil for a symbol.
func (s *Symbol) AsList() *List { return nil }

// AsSymbol returns the given symbol
func (s *Symbol) AsSymbol() *Symbol { return s }

func (s *Symbol) String() string { return s.Value }
