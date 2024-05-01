package sexp

// SExp is an S-Expression is either a List of zero or more S-Expressions, or
// a Symbol.
type SExp interface {
	// IsList checks whether this S-Expression is a list.
	IsList() bool
	// IsSymbol checks whether this S-Expression is a symbol.
	IsSymbol() bool
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

// IsList sets that is a list.
func (l *List) IsList() bool { return true }

// IsSymbol that a List is not a Symbol.
func (l *List) IsSymbol() bool { return false }

// Len gets the number of elements in this list.
func (l *List) Len() int { return len(l.Elements) }

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

// IsList sets that A Symbol is not a List.
func (s *Symbol) IsList() bool { return false }

// IsSymbol sets tha is a Symbol.
func (s *Symbol) IsSymbol() bool { return true }

func (s *Symbol) String() string { return s.Value }
