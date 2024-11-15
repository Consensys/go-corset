package sexp

import "fmt"

// Span represents a contiguous slice of the original string.  Instead of
// representing this as a string slice, however, it is useful to retain the
// physical indices.  This allows us to do certain things, such as determine the
// enclosing line, etc.
type Span struct {
	// The first character of this span in the original string.
	start int
	// One past the final character of this span in the original string.
	end int
}

// NewSpan constructs a new span whilst checking the internal invariants are
// maintained.
func NewSpan(start int, end int) Span {
	if start > end {
		panic("invalid span")
	}

	return Span{start, end}
}

// Start returns the starting index of this span in the original string.
func (p *Span) Start() int {
	return p.start
}

// End returns one past the last index of this span in the original string.
func (p *Span) End() int {
	return p.end
}

// Length returns the number of characters covered by this span in the original
// string.
func (p *Span) Length() int {
	return p.end - p.start
}

// SourceMaps provides a mechanism for mapping terms from an AST to multiple
// source files.
type SourceMaps[T comparable] struct {
	// Arrray of known source maps.
	maps []SourceMap[T]
}

// NewSourceMaps constructs an (initially empty) set of source maps.  The
// intention is that this is populated as each file is parsed.
func NewSourceMaps[T comparable]() *SourceMaps[T] {
	return &SourceMaps[T]{[]SourceMap[T]{}}
}

// SyntaxError constructs a syntax error for a given node contained within one
// of the source files managed by this set of source maps.
func (p *SourceMaps[T]) SyntaxError(node T, msg string) *SyntaxError {
	for _, m := range p.maps {
		if m.Has(node) {
			span := m.Get(node)
			return m.srcfile.SyntaxError(span, msg)
		}
	}
	// If we get here, then it means the node on which the error occurrs is not
	// present in any of the source maps.  This should not be possible, provided
	// the parser is implemented correctly.
	panic("missing mapping for source node")
}

// SyntaxErrors is really just a helper that construct a syntax error and then
// places it into an array of size one.  This is helpful for situations where
// sets of syntax errors are being passed around.
func (p *SourceMaps[T]) SyntaxErrors(node T, msg string) []SyntaxError {
	err := p.SyntaxError(node, msg)
	return []SyntaxError{*err}
}

// Join a given source map into this set of source maps.  The effect of this is
// that nodes recorded in the given source map can be accessed from this set.
func (p *SourceMaps[T]) Join(srcmap *SourceMap[T]) {
	p.maps = append(p.maps, *srcmap)
}

// SourceMap maps terms from an AST to slices of their originating string.  This
// is important for error handling when we wish to highlight exactly where, in
// the original source file, a given error has arisen.
//
// This provides various useful functions to aid reporting syntax errors, such
// as identifying the enclosing line for a given span, etc.
type SourceMap[T comparable] struct {
	// Maps a given AST object to a span in the original string.
	mapping map[T]Span
	// Enclosing source file
	srcfile SourceFile
}

// NewSourceMap constructs an initially empty source map for a given string.
func NewSourceMap[T comparable](srcfile SourceFile) *SourceMap[T] {
	mapping := make(map[T]Span)
	return &SourceMap[T]{mapping, srcfile}
}

// Source returns the underlying source file on which this map operates.
func (p *SourceMap[T]) Source() SourceFile {
	return p.srcfile
}

// Put registers a new AST item with a given span.  Note, if the item exists
// already, then it will panic.
func (p *SourceMap[T]) Put(item T, span Span) {
	if _, ok := p.mapping[item]; ok {
		panic(fmt.Sprintf("source map key already exists: %s", any(item)))
	}
	// Assign it
	p.mapping[item] = span
}

// Has checks whether a given item is contained within this source map.
func (p *SourceMap[T]) Has(item T) bool {
	_, ok := p.mapping[item]
	return ok
}

// Get determines the span associated with a given AST item extract from the
// original text.  Note, if the item is not registered with this source map,
// then it will panic.
func (p *SourceMap[T]) Get(item T) Span {
	if s, ok := p.mapping[item]; ok {
		return s
	}

	panic(fmt.Sprintf("invalid source map key: %s", any(item)))
}
