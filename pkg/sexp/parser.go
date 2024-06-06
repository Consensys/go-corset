package sexp

// Parse a given string into an S-expression, or return an error if the string
// is malformed.
func Parse(s string) (SExp, error) {
	p := NewParser(s)
	// Parse the input
	sExp, err := p.Parse()
	// Sanity check everything was parsed
	if err == nil && p.index != len(p.text) {
		return nil, p.error("unexpected remainder")
	}

	return sExp, err
}

// ParseAll parses a given string into zero or more S-expressions, whilst
// returning an error if the string is malformed.
func ParseAll(s string) ([]SExp, error) {
	terms := make([]SExp, 0)
	p := NewParser(s)
	// Parse the input
	for {
		term, err := p.Parse()
		// Sanity check everything was parsed
		if err != nil {
			return terms, err
		} else if term == nil {
			// EOF reached
			return terms, nil
		}

		terms = append(terms, term)
	}
}

// Parser represents a parser in the process of parsing a given string into one
// or more S-expressions.
type Parser struct {
	// Text being parsed
	text []rune
	// Determine current position within text
	index int
	//
}

// NewParser constructs a new instance of Parser
func NewParser(text string) *Parser {
	return &Parser{
		text:  []rune(text),
		index: 0,
	}
}

// Parse a given string into an S-Expression, or produce an error.
func (p *Parser) Parse() (SExp, error) {
	token := p.Next()

	if token == nil {
		return nil, nil
	} else if len(token) == 1 && token[0] == ')' {
		p.index-- // backup
		return nil, p.error("unexpected end-of-list")
	} else if len(token) == 1 && token[0] == '(' {
		var elements []SExp

		for c := p.Lookahead(0); c == nil || *c != ')'; c = p.Lookahead(0) {
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
		// Consume right-brace
		p.Next()
		// Done
		return &List{elements}, nil
	}

	return &Symbol{string(token)}, nil
}

// Next extracts the next token from a given string.
func (p *Parser) Next() []rune {
	index := p.index

	if index == len(p.text) {
		return nil
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
func (p *Parser) Lookahead(i int) *rune {
	// Compute actual position within text
	pos := i + p.index
	// Check what's there
	if len(p.text) > pos {
		switch p.text[pos] {
		case '(', ')', ';':
			return &p.text[pos]
		case ' ', '\n':
			return p.Lookahead(i + 1)
		default:
			return nil
		}
	}

	return nil
}

func (p *Parser) parseSymbol() []rune {
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

func (p *Parser) parseComment() []rune {
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

// Construct a parser error at the current position in the input stream.
func (p *Parser) error(msg string) *SyntaxError {
	span := NewSpan(p.index, p.index+1)
	return &SyntaxError{span, msg}
}
