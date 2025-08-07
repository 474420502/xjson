// Package parser provides XPath-like query parsing capabilities for xjson.
// It converts XPath-like query strings into executable query plans that can be
// executed against JSON data efficiently.
package parser

import (
	"fmt"
	"strconv"
)

// SyntaxProfile 定义了查询语言的语法符号
type SyntaxProfile struct {
	ChildSeparator      byte   // e.g., '/' or '.'
	DescendantSeparator string // e.g., "//" or ".."
	Wildcard            byte   // e.g., '*'
	// 可扩展更多符号
}

// DefaultXPathSyntax 提供了一个默认的类XPath语法配置
var DefaultXPathSyntax = SyntaxProfile{
	ChildSeparator:      '/',
	DescendantSeparator: "//",
	Wildcard:            '*',
}

// DotNotationSyntax 提供了一个基于点符号的备用语法
var DotNotationSyntax = SyntaxProfile{
	ChildSeparator:      '.',
	DescendantSeparator: "..",
	Wildcard:            '*',
}

// TokenType represents the type of a token in the query
type TokenType int

const (
	TokenEOF TokenType = iota
	TokenDot
	TokenSlash
	TokenDoubleSlash
	TokenStar
	TokenIdent
	TokenLeftBracket
	TokenRightBracket
	TokenLeftParen
	TokenRightParen
	TokenNumber
	TokenString
	TokenColon
	TokenComma

	// Operators
	TokenEQ  // ==
	TokenNE  // !=
	TokenLT  // <
	TokenLE  // <=
	TokenGT  // >
	TokenGE  // >=
	TokenAnd // and
	TokenOr  // or
	TokenNot // !
	TokenAt  // @

	// Special functions
	TokenExists
	TokenIncludes

	// Boolean literals
	TokenTrue
	TokenFalse
)

// Token represents a single token in the query
type Token struct {
	Type  TokenType
	Value string
	Pos   int
}

// Lexer tokenizes XPath-like query strings
type Lexer struct {
	input string
	pos   int
	start int
	len   int
}

// NewLexer creates a new lexer for the given input
func NewLexer(input string) *Lexer {
	return &Lexer{
		input: input,
		pos:   0,
		start: 0,
		len:   len(input),
	}
}

// NextToken returns the next token from the input
func (l *Lexer) NextToken() Token {
	l.skipWhitespace()

	if l.pos >= l.len {
		return Token{Type: TokenEOF, Pos: l.pos}
	}

	l.start = l.pos
	ch := l.current()

	switch ch {
	case '.':
		if l.peek() == '.' {
			l.advance()
			l.advance()
			return Token{Type: TokenDoubleSlash, Value: "..", Pos: l.start}
		}
		l.advance()
		return Token{Type: TokenDot, Value: ".", Pos: l.start}

	case '/':
		if l.peek() == '/' {
			l.advance()
			l.advance()
			return Token{Type: TokenDoubleSlash, Value: "//", Pos: l.start}
		}
		l.advance()
		return Token{Type: TokenSlash, Value: "/", Pos: l.start}

	case '@':
		l.advance()
		return Token{Type: TokenAt, Value: "@", Pos: l.start}

	case '*':
		l.advance()
		return Token{Type: TokenStar, Value: "*", Pos: l.start}

	case '[':
		l.advance()
		return Token{Type: TokenLeftBracket, Value: "[", Pos: l.start}

	case ']':
		l.advance()
		return Token{Type: TokenRightBracket, Value: "]", Pos: l.start}

	case '(':
		l.advance()
		return Token{Type: TokenLeftParen, Value: "(", Pos: l.start}

	case ')':
		l.advance()
		return Token{Type: TokenRightParen, Value: ")", Pos: l.start}

	case ':':
		l.advance()
		return Token{Type: TokenColon, Value: ":", Pos: l.start}

	case ',':
		l.advance()
		return Token{Type: TokenComma, Value: ",", Pos: l.start}

	case '"', '\'':
		return l.readString()

	case '=':
		if l.peek() == '=' {
			l.advance()
			l.advance()
			return Token{Type: TokenEQ, Value: "==", Pos: l.start}
		}
		// Allow single '=' for equality as well for simplified syntax
		l.advance()
		return Token{Type: TokenEQ, Value: "==", Pos: l.start}

	case '!':
		if l.peek() == '=' {
			l.advance()
			l.advance()
			return Token{Type: TokenNE, Value: "!=", Pos: l.start}
		}
		l.advance()
		return Token{Type: TokenNot, Value: "!", Pos: l.start}

	case '<':
		if l.peek() == '=' {
			l.advance()
			l.advance()
			return Token{Type: TokenLE, Value: "<=", Pos: l.start}
		}
		l.advance()
		return Token{Type: TokenLT, Value: "<", Pos: l.start}

	case '>':
		if l.peek() == '=' {
			l.advance()
			l.advance()
			return Token{Type: TokenGE, Value: ">=", Pos: l.start}
		}
		l.advance()
		return Token{Type: TokenGT, Value: ">", Pos: l.start}

	}

	if isDigit(ch) || ch == '-' {
		return l.readNumber()
	}

	if isAlpha(ch) || ch == '_' {
		return l.readIdentifier()
	}

	// Unknown character, skip it
	l.advance()
	return l.NextToken()
}

// Query represents a parsed XPath-like query
type Query struct {
	Steps []Step
}

// Step represents a single step in the query path
type Step struct {
	Type       StepType
	Axis       AxisType
	Name       string
	Predicates []Predicate
}

// StepType represents the type of a query step
type StepType int

const (
	StepChild StepType = iota
	StepDescendant
	StepWildcard
	StepRoot
)

// AxisType represents the axis of a query step
type AxisType int

const (
	AxisChild AxisType = iota
	AxisDescendant
)

// Predicate represents a filter condition
type Predicate struct {
	Type       PredicateType
	Expression *Expression
	Index      int
	Start      int
	End        int
}

// PredicateType represents the type of predicate
type PredicateType int

const (
	PredicateIndex PredicateType = iota
	PredicateSlice
	PredicateExpression
	PredicateWildcard
)

// Expression represents a filter expression
type Expression struct {
	Type        ExpressionType
	Left        *Expression
	Right       *Expression
	Operator    string
	Path        []string
	Value       interface{}
	Function    string
	IsAttribute bool // True if the path refers to an attribute
}

// ExpressionType represents the type of expression
type ExpressionType int

const (
	ExpressionBinary ExpressionType = iota
	ExpressionUnary
	ExpressionPath
	ExpressionLiteral
	ExpressionFunction
)

// Parser parses XPath-like queries into executable query plans
type Parser struct {
	lexer   *Lexer
	current Token
	peek    Token
}

// NewParser creates a new parser for the given query string
func NewParser(query string) *Parser {
	lexer := NewLexer(query)
	p := &Parser{lexer: lexer}
	p.nextToken() // Load first token
	p.nextToken() // Load second token
	return p
}

// Parse parses the query and returns a Query object
func (p *Parser) Parse() (*Query, error) {
	query := &Query{}

	// Handle root path starting with $ or /
	if p.current.Type == TokenSlash || p.current.Value == "$" {
		p.nextToken()
	}

	for p.current.Type != TokenEOF {
		step, err := p.parseStep()
		if err != nil {
			return nil, err
		}
		query.Steps = append(query.Steps, step)
	}

	return query, nil
}

// parseStep parses a single step in the query
func (p *Parser) parseStep() (Step, error) {
	step := Step{}
	initialPos := p.current.Pos

	// Handle recursive descent
	if p.current.Type == TokenDoubleSlash {
		step.Type = StepDescendant
		step.Axis = AxisDescendant
		p.nextToken()
	} else if p.current.Type == TokenSlash {
		step.Type = StepChild
		step.Axis = AxisChild
		p.nextToken()
	} else if p.current.Type == TokenDot {
		step.Type = StepChild
		step.Axis = AxisChild
		p.nextToken()
	}

	// Handle wildcard
	if p.current.Type == TokenStar {
		step.Type = StepWildcard
		step.Name = "*"
		p.nextToken()
	} else if p.current.Type == TokenIdent {
		step.Name = p.current.Value
		p.nextToken()
	} else if p.current.Type == TokenString {
		step.Name = p.current.Value
		p.nextToken()
	}

	// Handle predicates
	for p.current.Type == TokenLeftBracket {
		predicate, err := p.parsePredicate()
		if err != nil {
			return step, err
		}
		step.Predicates = append(step.Predicates, predicate)
	}

	// If we haven't consumed any tokens and we're not at the end, it's an error.
	if p.current.Pos == initialPos && p.current.Type != TokenEOF {
		return step, fmt.Errorf("unexpected token '%s' at position %d", p.current.Value, p.current.Pos)
	}

	// After a step name, if the next token is not a predicate, assume a child separator.
	if p.current.Type != TokenEOF && p.current.Type != TokenLeftBracket {
		step.Type = StepChild
		step.Axis = AxisChild
	}

	return step, nil
}

// parsePredicate parses a predicate (filter condition)
func (p *Parser) parsePredicate() (Predicate, error) {
	predicate := Predicate{}

	if p.current.Type != TokenLeftBracket {
		return predicate, fmt.Errorf("expected '[' at position %d", p.current.Pos)
	}
	p.nextToken()

	// Handle script expression `[?(...)]`
	if p.current.Type == TokenIdent && p.current.Value == "?" {
		p.nextToken() // Consume '?'
		if p.current.Type != TokenLeftParen {
			return predicate, fmt.Errorf("expected '(' after '?' in predicate at position %d", p.current.Pos)
		}
		p.nextToken() // Consume '('

		predicate.Type = PredicateExpression
		expr, err := p.parseExpression()
		if err != nil {
			return predicate, err
		}
		predicate.Expression = expr

		if p.current.Type != TokenRightParen {
			return predicate, fmt.Errorf("expected ')' at position %d", p.current.Pos)
		}
		p.nextToken() // Consume ')'

	} else {
		// Handle different predicate types
		switch p.current.Type {
		case TokenStar:
			// This case might be invalid in standard XPath, but we keep it for now.
			predicate.Type = PredicateWildcard
			p.nextToken()

		case TokenNumber:
			// In XPath, a number inside a predicate is a positional index.
			// e.g., /books[1]
			start, err := strconv.Atoi(p.current.Value)
			if err != nil {
				return predicate, fmt.Errorf("invalid number in predicate: %s", p.current.Value)
			}
			p.nextToken()

			// Unlike JSONPath, XPath slices are not standard. We'll treat numbers as indices.
			predicate.Type = PredicateIndex
			predicate.Index = start // XPath is 1-based, engine needs to adjust

		default:
			// This is now the standard path for XPath expressions like [@attribute='value']
			predicate.Type = PredicateExpression
			expr, err := p.parseExpression()
			if err != nil {
				return predicate, err
			}
			predicate.Expression = expr
		}
	}

	if p.current.Type != TokenRightBracket {
		return predicate, fmt.Errorf("expected ']' at position %d", p.current.Pos)
	}
	p.nextToken()

	return predicate, nil
}

// parseExpression parses a filter expression
func (p *Parser) parseExpression() (*Expression, error) {
	return p.parseOrExpression()
}

// parseOrExpression parses logical OR expressions
func (p *Parser) parseOrExpression() (*Expression, error) {
	left, err := p.parseAndExpression()
	if err != nil {
		return nil, err
	}

	for p.current.Type == TokenOr {
		op := p.current.Value
		p.nextToken()

		right, err := p.parseAndExpression()
		if err != nil {
			return nil, err
		}

		left = &Expression{
			Type:     ExpressionBinary,
			Left:     left,
			Right:    right,
			Operator: op,
		}
	}

	return left, nil
}

// parseAndExpression parses logical AND expressions
func (p *Parser) parseAndExpression() (*Expression, error) {
	left, err := p.parseComparisonExpression()
	if err != nil {
		return nil, err
	}

	for p.current.Type == TokenAnd {
		op := p.current.Value
		p.nextToken()

		right, err := p.parseComparisonExpression()
		if err != nil {
			return nil, err
		}

		left = &Expression{
			Type:     ExpressionBinary,
			Left:     left,
			Right:    right,
			Operator: op,
		}
	}

	return left, nil
}

// parseComparisonExpression parses comparison expressions
func (p *Parser) parseComparisonExpression() (*Expression, error) {
	left, err := p.parsePrimaryExpression()
	if err != nil {
		return nil, err
	}

	switch p.current.Type {
	case TokenEQ, TokenNE, TokenLT, TokenLE, TokenGT, TokenGE:
		op := p.current.Value
		p.nextToken()

		right, err := p.parsePrimaryExpression()
		if err != nil {
			return nil, err
		}

		return &Expression{
			Type:     ExpressionBinary,
			Left:     left,
			Right:    right,
			Operator: op,
		}, nil
	}

	return left, nil
}

// parsePrimaryExpression parses primary expressions (paths, literals, functions)
func (p *Parser) parsePrimaryExpression() (*Expression, error) {
	switch p.current.Type {
	case TokenIdent, TokenAt:
		// Path expression starting with an identifier or '@'
		return p.parsePathExpression()

	case TokenString:
		value := p.current.Value
		p.nextToken()
		return &Expression{
			Type:  ExpressionLiteral,
			Value: value,
		}, nil

	case TokenNumber:
		value := p.current.Value
		p.nextToken()

		// Try to parse as integer first
		if intVal, err := strconv.Atoi(value); err == nil {
			return &Expression{
				Type:  ExpressionLiteral,
				Value: intVal,
			}, nil
		}

		// Try to parse as float
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			return &Expression{
				Type:  ExpressionLiteral,
				Value: floatVal,
			}, nil
		}

		return nil, fmt.Errorf("invalid number: %s", value)

	case TokenTrue:
		p.nextToken()
		return &Expression{
			Type:  ExpressionLiteral,
			Value: true,
		}, nil

	case TokenFalse:
		p.nextToken()
		return &Expression{
			Type:  ExpressionLiteral,
			Value: false,
		}, nil

	case TokenNot:
		p.nextToken()
		expr, err := p.parsePrimaryExpression()
		if err != nil {
			return nil, err
		}

		return &Expression{
			Type:     ExpressionUnary,
			Left:     expr,
			Operator: "!",
		}, nil

	case TokenExists:
		// exists() function
		p.nextToken()
		if p.current.Type != TokenLeftParen {
			return nil, fmt.Errorf("expected '(' after 'exists' at position %d", p.current.Pos)
		}
		p.nextToken()

		// Parse path argument
		pathExpr, err := p.parsePathExpression()
		if err != nil {
			return nil, err
		}

		if p.current.Type != TokenRightParen {
			return nil, fmt.Errorf("expected ')' at position %d", p.current.Pos)
		}
		p.nextToken()

		return &Expression{
			Type:     ExpressionFunction,
			Function: "exists",
			Left:     pathExpr,
		}, nil

	case TokenIncludes:
		// includes() function
		p.nextToken()
		if p.current.Type != TokenLeftParen {
			return nil, fmt.Errorf("expected '(' after 'includes' at position %d", p.current.Pos)
		}
		p.nextToken()

		// Parse first argument (array/path)
		arrayExpr, err := p.parsePathExpression()
		if err != nil {
			return nil, err
		}

		if p.current.Type != TokenComma {
			return nil, fmt.Errorf("expected ',' in includes() at position %d", p.current.Pos)
		}
		p.nextToken()

		// Parse second argument (value to search for)
		valueExpr, err := p.parsePrimaryExpression()
		if err != nil {
			return nil, err
		}

		if p.current.Type != TokenRightParen {
			return nil, fmt.Errorf("expected ')' at position %d", p.current.Pos)
		}
		p.nextToken()

		return &Expression{
			Type:     ExpressionFunction,
			Function: "includes",
			Left:     arrayExpr,
			Right:    valueExpr,
		}, nil

	case TokenLeftParen:
		p.nextToken()
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}

		if p.current.Type != TokenRightParen {
			return nil, fmt.Errorf("expected ')' at position %d", p.current.Pos)
		}
		p.nextToken()

		return expr, nil
	}

	return nil, fmt.Errorf("unexpected token %s at position %d", p.current.Value, p.current.Pos)
}

// parsePathExpression parses a path expression.
// In XPath filters, paths can start with an attribute signifier '@' or be a relative path.
func (p *Parser) parsePathExpression() (*Expression, error) {
	var path []string
	isAttribute := false

	if p.current.Type == TokenAt {
		isAttribute = true
		p.nextToken() // Consume '@'
		if p.current.Type != TokenIdent {
			return nil, fmt.Errorf("expected attribute name after '@' at position %d", p.current.Pos)
		}
		path = append(path, p.current.Value)
		p.nextToken() // Consume attribute name
	} else if p.current.Type == TokenIdent {
		// This is a relative path from the context node.
		path = append(path, p.current.Value)
		p.nextToken()
	} else {
		return nil, fmt.Errorf("expected identifier or attribute at position %d", p.current.Pos)
	}

	// Chain path segments with '/' or '.'
	for p.current.Type == TokenSlash || p.current.Type == TokenDot {
		p.nextToken()

		if p.current.Type == TokenIdent {
			path = append(path, p.current.Value)
			p.nextToken()
		} else {
			return nil, fmt.Errorf("expected identifier after separator at position %d", p.current.Pos)
		}
	}

	expr := &Expression{
		Type:        ExpressionPath,
		Path:        path,
		IsAttribute: isAttribute,
	}

	return expr, nil
}

// Helper methods for the lexer

func (l *Lexer) current() byte {
	if l.pos >= l.len {
		return 0
	}
	return l.input[l.pos]
}

func (l *Lexer) peek() byte {
	if l.pos+1 >= l.len {
		return 0
	}
	return l.input[l.pos+1]
}

func (l *Lexer) advance() {
	l.pos++
}

func (l *Lexer) skipWhitespace() {
	for l.pos < l.len && isWhitespace(l.current()) {
		l.pos++
	}
}

func (l *Lexer) readString() Token {
	quote := l.current()
	l.advance() // Skip opening quote

	start := l.pos

	for l.pos < l.len && l.current() != quote {
		if l.current() == '\\' {
			l.advance() // Skip escape character
		}
		l.advance()
	}

	value := l.input[start:l.pos]

	if l.pos < l.len {
		l.advance() // Skip closing quote
	}

	return Token{
		Type:  TokenString,
		Value: value,
		Pos:   l.start,
	}
}

func (l *Lexer) readNumber() Token {
	start := l.pos

	if l.current() == '-' {
		l.advance()
	}

	for l.pos < l.len && isDigit(l.current()) {
		l.advance()
	}

	if l.pos < l.len && l.current() == '.' {
		l.advance()
		for l.pos < l.len && isDigit(l.current()) {
			l.advance()
		}
	}

	if l.pos < l.len && (l.current() == 'e' || l.current() == 'E') {
		l.advance()
		if l.pos < l.len && (l.current() == '+' || l.current() == '-') {
			l.advance()
		}
		for l.pos < l.len && isDigit(l.current()) {
			l.advance()
		}
	}

	return Token{
		Type:  TokenNumber,
		Value: l.input[start:l.pos],
		Pos:   l.start,
	}
}

func (l *Lexer) readIdentifier() Token {
	start := l.pos

	for l.pos < l.len && (isAlpha(l.current()) || isDigit(l.current()) || l.current() == '_') {
		l.advance()
	}

	value := l.input[start:l.pos]
	tokenType := TokenIdent

	// Check for keywords
	switch value {
	case "exists":
		tokenType = TokenExists
	case "includes":
		tokenType = TokenIncludes
	case "true":
		tokenType = TokenTrue
	case "false":
		tokenType = TokenFalse
	case "and":
		tokenType = TokenAnd
	case "or":
		tokenType = TokenOr
	}

	return Token{
		Type:  tokenType,
		Value: value,
		Pos:   l.start,
	}
}

func (p *Parser) nextToken() {
	p.current = p.peek
	p.peek = p.lexer.NextToken()
}

// Helper functions

func isWhitespace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}

func isAlpha(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}
