package sexp

import (
	"unicode"
)

// Parser represents a parser in the process of parsing a given string into one
// or more S-expressions.
type Parser struct {
	// Source file being parsed
	srcfile *SourceFile
	// Cache (for simplicity)
	text []rune
	// Determine current position within text
	index int
	// Mapping from constructed S-Expressions to their spans in the original text.
	srcmap *SourceMap[SExp]
}

// NewParser constructs a new instance of Parser
func NewParser(srcfile *SourceFile) *Parser {
	// Construct initial parser.
	return &Parser{
		srcfile: srcfile,
		text:    srcfile.Contents(),
		index:   0,
		srcmap:  NewSourceMap[SExp](srcfile.Contents()),
	}
}

// SourceMap returns the internal source map constructing during parsing.  Using
// this one can determine, for each SExp, where in the original text it
// originated.  This is helpful, for example, when reporting syntax errors.
func (p *Parser) SourceMap() *SourceMap[SExp] {
	return p.srcmap
}

// Text returns the underlying text for this parser.
func (p *Parser) Text() []rune {
	return p.text
}

// Parse a given string into an S-Expression, or produce an error.
func (p *Parser) Parse() (SExp, *SyntaxError) {
	var term SExp
	// Skip over any whitespace.  This is import to get the correct starting
	// point for this term.
	p.SkipWhiteSpace()
	// Record start of this term
	start := p.index
	// Extract next token from the stream
	token := p.Next()

	if token == nil {
		return nil, nil
	} else if len(token) == 1 && token[0] == ')' {
		p.index-- // backup
		return nil, p.error("unexpected end-of-list")
	} else if len(token) == 1 && token[0] == '(' {
		elements, err := p.parseSequence(')')
		// Check for error
		if err != nil {
			return nil, err
		}
		// Done
		term = &List{elements}
	} else if len(token) == 1 && token[0] == '{' {
		elements, err := p.parseSequence('}')
		// Check for error
		if err != nil {
			return nil, err
		}
		// Done
		term = &Set{elements}
	} else {
		// Must be a symbol
		term = &Symbol{string(token)}
	}
	// Register item in source map
	p.srcmap.Put(term, NewSpan(start, p.index))
	// Done
	return term, nil
}

// Next extracts the next token from a given string.
func (p *Parser) Next() []rune {
	// Skip any whitespace and/or comments.
	p.SkipWhiteSpace()
	// Catch end-of-file
	if p.index == len(p.text) {
		return nil
	}
	// Check what we have
	switch p.text[p.index] {
	case '(', ')', '{', '}':
		// List/set begin / end
		p.index = p.index + 1
		return p.text[p.index-1 : p.index]
	}
	// Symbol
	return p.parseSymbol()
}

// SkipWhiteSpace skips over any whitespace, including comments.
func (p *Parser) SkipWhiteSpace() {
	for p.index < len(p.text) && (unicode.IsSpace(p.text[p.index]) || p.text[p.index] == ';') {
		// Skip comment
		if p.text[p.index] == ';' {
			i := len(p.text)
			//
			for j := p.index; j < i; j++ {
				c := p.text[j]
				if c == '\n' {
					i = j + 1
					break
				}
			}
			// Skip comment
			p.index = i
		} else {
			// skip space
			p.index++
		}
	}
}

// Lookahead and see what punctuation is next.
func (p *Parser) Lookahead(i int) *rune {
	// Compute actual position within text
	pos := i + p.index
	// Check what's there
	if len(p.text) > pos {
		r := p.text[pos]
		if r == '(' || r == ')' || r == '{' || r == '}' || r == ';' {
			return &r
		} else if unicode.IsSpace(r) {
			return p.Lookahead(i + 1)
		}
	}

	return nil
}

func (p *Parser) parseSymbol() []rune {
	// Parse token
	i := len(p.text)

	for j := p.index; j < i; j++ {
		c := p.text[j]
		if c == ')' || c == '}' || c == ' ' || c == '\n' || c == '\t' {
			i = j
			break
		}
	}
	// Reached end of token
	token := p.text[p.index:i]
	p.index = i

	return token
}

func (p *Parser) parseSequence(terminator rune) ([]SExp, *SyntaxError) {
	var elements []SExp

	for c := p.Lookahead(0); c == nil || *c != terminator; c = p.Lookahead(0) {
		// Parse next element
		element, err := p.Parse()
		if err != nil {
			return nil, err
		} else if element == nil {
			p.index-- // backup
			return nil, p.error("unexpected end-of-file")
		}
		// Continue around!
		elements = append(elements, element)
	}
	// Consume terminator
	p.Next()
	//
	return elements, nil
}

// Construct a parser error at the current position in the input stream.
func (p *Parser) error(msg string) *SyntaxError {
	span := NewSpan(p.index, p.index+1)
	return p.srcfile.SyntaxError(span, msg)
}
