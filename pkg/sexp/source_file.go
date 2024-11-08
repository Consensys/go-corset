package sexp

import (
	"fmt"
)

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
func (s *SourceFile) Parse() (SExp, *SourceMap[SExp], error) {
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
func (s *SourceFile) ParseAll() ([]SExp, *SourceMap[SExp], error) {
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
