package sexp

import (
	"errors"
	"fmt"
)

// Parse a given string into an S-expression, or return an error if the string
// is malformed.
func Parse(s string) (SExp, error) {
	p := &Parser{s, 0}
	// Parse the input
	sExp, err := p.Parse()
	// Sanity check everything was parsed
	if err == nil && p.index != len(p.text) {
		return nil, fmt.Errorf("unexpected string remainder: %s", p.text[p.index:])
	}

	return sExp, err
}

// Parser represents a parser in the process of parsing a given string into one
// or more S-expressions.
type Parser struct {
	// Text being parsed
	text string
	// Determine current position within text
	index int
}

// NewParser constructs a new instance of Parser
func NewParser(text string) *Parser {
	return &Parser{
		text:  text,
		index: 0,
	}
}

// Parse a given string into an S-Expression, or produce an error.
func (p *Parser) Parse() (SExp, error) {
	token := p.Next()

	if token == "" {
		return nil, nil
	} else if token == ")" {
		return nil, errors.New("unexpected end-of-list")
	} else if token == "(" {
		var elements []SExp

		for p.Lookahead(0) != ")" {
			// Parse next element
			element, err := p.Parse()
			if err != nil {
				return nil, err
			} else if element == nil {
				return nil, errors.New("unexpected end-of-file")
			}
			// Continue around!
			elements = append(elements, element)
		}
		// Consume right-brace
		p.Next()
		// Done
		return &List{elements}, nil
	}

	return &Symbol{token}, nil
}

// Next extracts the next token from a given string.
func (p *Parser) Next() string {
	index := p.index

	if index == len(p.text) {
		return ""
	}

	switch p.text[index] {
	case '(', ')':
		// List begin / end
		p.index = p.index + 1
		return p.text[index:p.index]
	case ' ', '\t', '\n':
		// Whitespace
		p.index = p.index + 1
		return p.Next()
	case ';':
		// Comment
		return p.parseComment()
	}
	// Symbol
	return p.parseSymbol()
}

// Lookahead and see what punctuation is next.
func (p *Parser) Lookahead(i int) string {
	// Compute actual position within text
	pos := i + p.index
	// Check what's there
	if len(p.text) > pos {
		switch p.text[pos] {
		case '(', ')', ';':
			return p.text[pos : pos+1]
		case ' ', '\n':
			return p.Lookahead(i + 1)
		default:
			return ""
		}
	}

	return ""
}

func (p *Parser) parseSymbol() string {
	// Parse token
	i := len(p.text)

	for j := p.index; j < i; j++ {
		c := p.text[j]
		if c == ')' || c == ' ' || c == '\n' || c == '\t' {
			i = j
			break
		}
	}
	// Reached end of token
	token := p.text[p.index:i]
	p.index = i

	return token
}

func (p *Parser) parseComment() string {
	// Parse token
	i := len(p.text)

	for j := p.index; j < i; j++ {
		c := p.text[j]
		if c == '\n' {
			i = j
			break
		}
	}
	// Skip comment
	p.index = i
	// Look for next token
	return p.Next()
}
