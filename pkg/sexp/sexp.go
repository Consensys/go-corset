package sexp

// An S-Expression is either a List of zero or more S-Expressions, or
// a Symbol.
type SExp interface {
	// Check whether this S-Expression is a list.
	IsList() bool
	// Check whether this S-Expression is a symbol.
	IsSymbol() bool
	// Generate string
	String() string
}

// ===================================================================
// List
// ===================================================================

// Represents a list of zero or more S-Expressions.
type List struct { Elements []SExp }
// A list is a list
func (*List) IsList() bool { return true }
// A list is not a symbol
func (*List) IsSymbol() bool { return false }
// Get the number of elements in this list
func (l *List) Len() int { return len(l.Elements) }
//
func (l *List) String() string {
	var s string = "("
	for i := 0; i < len(l.Elements); i++ {
		if i != 0 { s += "," }
		s += l.Elements[i].String()
	}
	s += ")"
	return s
}

/// Matches a list which starts with at least n symbols, of which the
/// first m match the given strings.
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

// A terminating symbol.
type Symbol struct { Value string }
// A Symbol is not a List
func (*Symbol) IsList() bool { return false }
// A Sybmol is a Symbol
func (*Symbol) IsSymbol() bool { return true }
//
func (p *Symbol) String() string { return p.Value }
