package sexp

import (
	"fmt"
)

// SyntaxError is a structured error which retains the index into the original
// string where an error occurred, along with an error message.
type SyntaxError struct {
	// Byte index into string being parsed where error arose.
	span Span
	// Error message being reported
	msg string
}

// NewSyntaxError simply constructs a new syntax error.
func NewSyntaxError(span Span, msg string) *SyntaxError {
	return &SyntaxError{span, msg}
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
