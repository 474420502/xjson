package query

import (
	"fmt"
	"strconv"
	"strings"
)

// Parser holds the state of the query parser.
type Parser struct {
	s      *Scanner
	cur    Token
	peek   Token
	errors []string
}

// NewParser creates a new parser for the given query.
func NewParser(query string) *Parser {
	p := &Parser{
		s:      NewScanner(strings.NewReader(query)),
		errors: []string{},
	}
	p.nextToken()
	p.nextToken()
	return p
}

// Errors returns any errors that occurred during parsing.
func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) nextToken() {
	p.cur = p.peek
	p.peek = p.s.Scan()
}

// Parse parses the query and returns a list of query tokens.
func (p *Parser) Parse() ([]QueryToken, error) {
	var tokens []QueryToken

	for p.cur.Type != TokenEOF {
		var token QueryToken
		var err error

		switch p.cur.Type {
		case TokenSlash:
			token, err = p.parsePathSegment()
		case TokenSlash2:
			token, err = p.parseRecursiveDescent()
		case TokenLBRACKET:
			token, err = p.parseBracketExpression()
		case TokenDOT2:
			token = QueryToken{Type: OpParent}
		default:
			err = fmt.Errorf("unexpected token: %v", p.cur)
		}

		if err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
		p.nextToken()
	}

	return tokens, nil
}

func (p *Parser) parsePathSegment() (QueryToken, error) {
	p.nextToken() // Consume '/'
	if p.cur.Type != TokenIdentity {
		return QueryToken{}, fmt.Errorf("expected identity after '/', got %v", p.cur)
	}
	return QueryToken{Type: OpKey, Value: p.cur.Literal}, nil
}

func (p *Parser) parseRecursiveDescent() (QueryToken, error) {
	p.nextToken() // Consume '//'
	if p.cur.Type != TokenIdentity {
		return QueryToken{}, fmt.Errorf("expected identity after '//', got %v", p.cur)
	}
	return QueryToken{Type: OpRecursiveKey, Value: p.cur.Literal}, nil
}

func (p *Parser) parseBracketExpression() (QueryToken, error) {
	p.nextToken() // Consume '['

	if p.cur.Type == TokenCurrent {
		p.nextToken() // Consume '@'
		if p.cur.Type != TokenIdentity {
			return QueryToken{}, fmt.Errorf("expected function name after '@', got %v", p.cur.Type)
		}
		name := p.cur.Literal
		p.nextToken() // Consume function name
		if p.cur.Type != TokenRBRACKET {
			return QueryToken{}, fmt.Errorf("expected ']' after function call, got %v", p.cur.Type)
		}
		return QueryToken{Type: OpFunc, Value: name}, nil
	}

	if p.cur.Type == TokenColon { // [:idx]
		p.nextToken() // consume ':'
		end, err := strconv.Atoi(p.cur.Literal)
		if err != nil {
			return QueryToken{}, fmt.Errorf("invalid slice end index: %w", err)
		}
		p.nextToken()
		if p.cur.Type != TokenRBRACKET {
			return QueryToken{}, fmt.Errorf("expected ']' after slice, got %v", p.cur.Type)
		}
		return QueryToken{Type: OpSlice, Value: [2]int{0, end}}, nil
	}

	if p.cur.Type == TokenIdentity { // [idx] or [start:end] or [start:]
		start, err := strconv.Atoi(p.cur.Literal)
		if err != nil {
			return QueryToken{}, fmt.Errorf("invalid index or slice start: %w", err)
		}
		p.nextToken()

		if p.cur.Type == TokenRBRACKET { // [idx]
			return QueryToken{Type: OpIndex, Value: start}, nil
		}

		if p.cur.Type == TokenColon {
			p.nextToken()                    // consume ':'
			if p.cur.Type == TokenRBRACKET { // [start:]
				return QueryToken{Type: OpSlice, Value: [2]int{start, -1}}, nil
			}
			end, err := strconv.Atoi(p.cur.Literal)
			if err != nil {
				return QueryToken{}, fmt.Errorf("invalid slice end index: %w", err)
			}
			p.nextToken()
			if p.cur.Type != TokenRBRACKET {
				return QueryToken{}, fmt.Errorf("expected ']' after slice, got %v", p.cur.Type)
			}
			return QueryToken{Type: OpSlice, Value: [2]int{start, end}}, nil
		}

		return QueryToken{}, fmt.Errorf("unexpected token in bracket expression: %v", p.cur)
	}

	return QueryToken{}, fmt.Errorf("invalid token in bracket expression: %v", p.cur.Type)
}
