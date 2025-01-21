package sexp

import (
	"fmt"
)

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

// SourceFile represents a given source file (typically stored on disk).
type SourceFile struct {
	// File name for this source file.
	filename string
	// Contents of this file.
	contents []rune
}

// NewSourceFile constructs a new source file from a given byte array.
func NewSourceFile(filename string, bytes []byte) *SourceFile {
	// Convert bytes into runes for easier parsing
	contents := []rune(string(bytes))
	return &SourceFile{filename, contents}
}

// Filename returns the filename associated with this source file.
func (s *SourceFile) Filename() string {
	return s.filename
}

// Contents returns the contents of this source file.
func (s *SourceFile) Contents() []rune {
	return s.contents
}

// Parse a given string into an S-expression, or return an error if the string
// is malformed.  A source map is also returned for debugging purposes.
func (s *SourceFile) Parse() (SExp, *SourceMap[SExp], *SyntaxError) {
	p := NewParser(s)
	// Parse the input
	sExp, err := p.Parse()
	// Sanity check everything was parsed
	if err == nil && p.index != len(p.text) {
		return nil, nil, p.error("unexpected remainder")
	}
	// Done
	return sExp, p.SourceMap(), err
}

// ParseAll converts a given string into zero or more S-expressions, or returns
// an error if the string is malformed.  A source map is also returned for
// debugging purposes.  The key distinction from Parse is that this function
// continues parsing after the first S-expression is encountered.
func (s *SourceFile) ParseAll() ([]SExp, *SourceMap[SExp], *SyntaxError) {
	p := NewParser(s)
	//
	terms := make([]SExp, 0)
	// Parse the input
	for {
		term, err := p.Parse()
		// Sanity check everything was parsed
		if err != nil {
			return terms, p.srcmap, err
		} else if term == nil {
			// EOF reached
			return terms, p.srcmap, nil
		}

		terms = append(terms, term)
	}
}

// SyntaxError constructs a syntax error over a given span of this file with a
// given message.
func (s *SourceFile) SyntaxError(span Span, msg string) *SyntaxError {
	return &SyntaxError{s, span, msg}
}

// FindFirstEnclosingLine determines the first line  in this source file which
// encloses the start of a span.  Observe that, if the position is beyond the
// bounds of the source file then the last physical line is returned.  Also,
// the returned line is not guaranteed to enclose the entire span, as these can
// cross multiple lines.
func (s *SourceFile) FindFirstEnclosingLine(span Span) Line {
	// Index identifies the current position within the original text.
	index := span.start
	// Num records the line number, counting from 1.
	num := 1
	// Start records the starting offset of the current line.
	start := 0
	// Find the line.
	for i := 0; i < len(s.contents); i++ {
		if i == index {
			end := findEndOfLine(index, s.contents)
			return Line{s.contents, Span{start, end}, num}
		} else if s.contents[i] == '\n' {
			num++
			start = i + 1
		}
	}
	//
	return Line{s.contents, Span{start, len(s.contents)}, num}
}

// SyntaxError is a structured error which retains the index into the original
// string where an error occurred, along with an error message.
type SyntaxError struct {
	srcfile *SourceFile
	// Byte index into string being parsed where error arose.
	span Span
	// Error message being reported
	msg string
}

// SourceFile returns the underlying source file that this syntax error covers.
func (p *SyntaxError) SourceFile() *SourceFile {
	return p.srcfile
}

// Span returns the span of the original text on which this error is reported.
func (p *SyntaxError) Span() Span {
	return p.span
}

// Message returns the message to be reported.
func (p *SyntaxError) Message() string {
	return p.msg
}

// Error implements the error interface.
func (p *SyntaxError) Error() string {
	return fmt.Sprintf("%d:%d:%s", p.span.Start(), p.span.End(), p.Message())
}

// FirstEnclosingLine determines the first line in this source file to which
// this error is associated. Observe that, if the position is beyond the bounds
// of the source file then the last physical line is returned.  Also, the
// returned line is not guaranteed to enclose the entire span, as these can
// cross multiple lines.
func (p *SyntaxError) FirstEnclosingLine() Line {
	return p.srcfile.FindFirstEnclosingLine(p.span)
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