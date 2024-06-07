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

// Line provides information about a given line within the original string.
// This includes the line number (counting from 1), and the span of the line
// within the original string.
type Line struct {
	// Original text
	text []rune
	// Span within original text of this line.
	span Span
	// Line number of this line (counting from 1).
	number int
}

// Get the string representing this line.
func (p *Line) String() string {
	// Extract runes representing line
	runes := p.text[p.span.start:p.span.end]
	// Convert into string
	return string(runes)
}

// Number gets the line number of this line, where the first line in a string
// has line number 1.
func (p *Line) Number() int {
	return p.number
}

// Start returns the starting index of this line in the original string.
func (p *Line) Start() int {
	return p.span.start
}

// Length returns the number of characters in this line.
func (p *Line) Length() int {
	return p.span.Length()
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
	// Original string
	text []rune
}

// NewSourceMap constructs an initially empty source map for a given string.
func NewSourceMap[T comparable](text []rune) *SourceMap[T] {
	mapping := make(map[T]Span)
	return &SourceMap[T]{mapping, text}
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

// Get determines the span associated with a given AST item extract from the
// original text.  Note, if the item is not registered with this source map,
// then it will panic.
func (p *SourceMap[T]) Get(item T) Span {
	if s, ok := p.mapping[item]; ok {
		return s
	}

	panic(fmt.Sprintf("invalid source map key: %s", any(item)))
}

// FindFirstEnclosingLine determines the first line which encloses the start of
// a span.  Observe that, if the position is beyond the bounds of the source
// string then the last physical line is returned.  Also, the returned line is
// not guaranteed to enclose the entire span, as these can cross multiple lines.
func (p *SourceMap[T]) FindFirstEnclosingLine(span Span) Line {
	// Index identifies the current position within the original text.
	index := span.start
	// Num records the line number, counting from 1.
	num := 1
	// Start records the starting offset of the current line.
	start := 0
	// Find the line.
	for i := 0; i < len(p.text); i++ {
		if i == index {
			end := findEndOfLine(index, p.text)
			return Line{p.text, Span{start, end}, num}
		} else if p.text[i] == '\n' {
			num++
			start = i + 1
		}
	}
	//
	return Line{p.text, Span{start, len(p.text)}, num}
}

// Find the end of the enclosing line
func findEndOfLine(index int, text []rune) int {
	for i := index; i < len(text); i++ {
		if text[i] == '\n' {
			return i
		}
	}
	// No end in sight!
	return len(text)
}
