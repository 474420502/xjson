package parser

import (
	"fmt"
	"reflect"
	"testing"
)

func TestParseSimplePath(t *testing.T) {
	query := "store.book"
	p := NewParser(query)
	q, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if len(q.Steps) != 2 {
		t.Errorf("Expected 2 steps, got %d", len(q.Steps))
	}
	if q.Steps[0].Name != "store" || q.Steps[1].Name != "book" {
		t.Errorf("Unexpected step names: %v", []string{q.Steps[0].Name, q.Steps[1].Name})
	}
}

func TestParseWildcardAndRecursive(t *testing.T) {
	query := "store.*.price"
	p := NewParser(query)
	q, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	// store, *, price => 3 steps
	if len(q.Steps) != 3 {
		t.Errorf("Expected 3 steps, got %d", len(q.Steps))
	}
	if q.Steps[1].Type != StepWildcard {
		t.Errorf("Expected wildcard step, got %v", q.Steps[1].Type)
	}
}

func TestParseFilterExpression(t *testing.T) {
	query := "items[?(@.price>=10)]"
	p := NewParser(query)
	q, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if len(q.Steps) != 1 {
		t.Errorf("Expected 1 step, got %d", len(q.Steps))
	}
	step := q.Steps[0]
	if step.Name != "items" {
		t.Errorf("Expected step name items, got %s", step.Name)
	}
	if len(step.Predicates) != 1 {
		t.Errorf("Expected 1 predicate, got %d", len(step.Predicates))
	}
}

func TestParseErrors(t *testing.T) {
	cases := []string{
		"store.books[",
		"store.books[?(@.price > )]",
		"store.books[1:2:3]",
	}
	for _, c := range cases {
		t.Run(c, func(t *testing.T) {
			p := NewParser(c)
			_, err := p.Parse()
			if err == nil {
				t.Errorf("Expected error for %s, got nil", c)
			}
		})
	}
}

// Additional comprehensive parser tests

func TestParseComplexPaths(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		expectSteps int
		expectError bool
	}{
		{
			name:        "deep nested path",
			query:       "user.profile.address.city",
			expectSteps: 4,
		},
		{
			name:        "array access with index",
			query:       "items[0].name",
			expectSteps: 2, // items[0] and name
		},
		{
			name:        "array slice",
			query:       "items[1:3]",
			expectSteps: 1,
		},
		{
			name:        "recursive descent",
			query:       "store..price",
			expectSteps: 2,
		},
		{
			name:        "wildcard at root",
			query:       "*",
			expectSteps: 1,
		},
		{
			name:        "multiple wildcards",
			query:       "*.*.name",
			expectSteps: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(tt.query)
			q, err := p.Parse()

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(q.Steps) != tt.expectSteps {
				t.Errorf("Expected %d steps, got %d", tt.expectSteps, len(q.Steps))
			}
		})
	}
}

func TestParseIndexOperations(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		wantIdx int
		wantOk  bool
	}{
		{
			name:    "single index",
			query:   "items[5]",
			wantIdx: 5,
			wantOk:  true,
		},
		{
			name:    "negative index",
			query:   "items[-1]",
			wantIdx: -1,
			wantOk:  true,
		},
		{
			name:   "zero index",
			query:  "items[0]",
			wantOk: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(tt.query)
			q, err := p.Parse()
			if err != nil {
				t.Errorf("Parse error: %v", err)
				return
			}

			if len(q.Steps) != 1 {
				t.Errorf("Expected 1 step, got %d", len(q.Steps))
				return
			}

			step := q.Steps[0]
			if len(step.Predicates) == 0 {
				t.Error("Expected predicate for index operation")
				return
			}

			predicate := step.Predicates[0]
			if tt.wantOk && predicate.Type == PredicateIndex && predicate.Index != tt.wantIdx {
				t.Errorf("Expected index %d, got %d", tt.wantIdx, predicate.Index)
			}
		})
	}
}

func TestParseSliceOperations(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		wantStart *int
		wantEnd   *int
	}{
		{
			name:      "full slice",
			query:     "items[1:3]",
			wantStart: intPtr(1),
			wantEnd:   intPtr(3),
		},
		{
			name:    "slice from start",
			query:   "items[:3]",
			wantEnd: intPtr(3),
		},
		{
			name:      "slice to end",
			query:     "items[1:]",
			wantStart: intPtr(1),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(tt.query)
			q, err := p.Parse()
			if err != nil {
				t.Errorf("Parse error: %v", err)
				return
			}

			if len(q.Steps) != 1 {
				t.Errorf("Expected 1 step, got %d", len(q.Steps))
				return
			}

			step := q.Steps[0]
			if len(step.Predicates) == 0 {
				t.Error("Expected predicate for slice operation")
				return
			}

			predicate := step.Predicates[0]
			if predicate.Type != PredicateSlice {
				t.Error("Expected slice predicate")
				return
			}

			// Compare slice start and end values directly
			startMatch := (tt.wantStart == nil && predicate.Start == 0) ||
				(tt.wantStart != nil && predicate.Start == *tt.wantStart)

			// For open-ended slices, End is set to -1
			var endMatch bool
			if tt.wantEnd == nil {
				// For open-ended slices like "items[1:]", End should be -1
				endMatch = predicate.End == -1 || predicate.End == 0
			} else {
				endMatch = predicate.End == *tt.wantEnd
			}

			if !startMatch {
				t.Errorf("Expected start %v, got %d", ptrToString(tt.wantStart), predicate.Start)
			}

			if !endMatch {
				t.Errorf("Expected end %v, got %d", ptrToString(tt.wantEnd), predicate.End)
			}
		})
	}
}

func TestParseStepTypes(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		stepIdx  int
		wantType StepType
	}{
		{
			name:     "child step",
			query:    "store.book",
			stepIdx:  0,
			wantType: StepChild,
		},
		{
			name:     "wildcard step",
			query:    "store.*",
			stepIdx:  1,
			wantType: StepWildcard,
		},
		{
			name:     "descendant step",
			query:    "store..price",
			stepIdx:  1,
			wantType: StepDescendant,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(tt.query)
			q, err := p.Parse()
			if err != nil {
				t.Errorf("Parse error: %v", err)
				return
			}

			if tt.stepIdx >= len(q.Steps) {
				t.Errorf("Step index %d out of bounds for %d steps", tt.stepIdx, len(q.Steps))
				return
			}

			step := q.Steps[tt.stepIdx]
			if step.Type != tt.wantType {
				t.Errorf("Expected step type %v, got %v", tt.wantType, step.Type)
			}
		})
	}
}

func TestEmptyAndRootQueries(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		expectError bool
	}{
		{
			name:  "root query",
			query: "$",
		},
		{
			name:  "empty query",
			query: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(tt.query)
			_, err := p.Parse()

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// Helper functions
func intPtr(i int) *int {
	return &i
}

func ptrToString(ptr *int) string {
	if ptr == nil {
		return "nil"
	}
	return fmt.Sprintf("%d", *ptr)
}

// Test the Lexer functionality to cover readString, readNumber, readIdentifier
func TestLexerTokenization(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []TokenType
	}{
		{
			name:     "string literal",
			input:    `"hello world"`,
			expected: []TokenType{TokenString, TokenEOF},
		},
		{
			name:     "number literal",
			input:    "123.456",
			expected: []TokenType{TokenNumber, TokenEOF},
		},
		{
			name:     "identifier",
			input:    "property",
			expected: []TokenType{TokenIdent, TokenEOF},
		},
		{
			name:     "complex query",
			input:    `store.book[0].title`,
			expected: []TokenType{TokenIdent, TokenDot, TokenIdent, TokenLeftBracket, TokenNumber, TokenRightBracket, TokenDot, TokenIdent, TokenEOF},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			var tokens []TokenType

			for {
				token := lexer.NextToken()
				tokens = append(tokens, token.Type)
				if token.Type == TokenEOF {
					break
				}
			}

			if !reflect.DeepEqual(tokens, tt.expected) {
				t.Errorf("Expected tokens %v, got %v", tt.expected, tokens)
			}
		})
	}
}

// Test primary expression parsing (12.1% coverage)
func TestParsePrimaryExpression(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		hasError bool
	}{
		{
			name:  "string literal",
			input: `"test"`,
		},
		{
			name:  "number literal",
			input: "42",
		},
		{
			name:  "boolean true",
			input: "true",
		},
		{
			name:  "boolean false",
			input: "false",
		},
		{
			name:  "path expression",
			input: "@.price",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a parser in filter context
			parser := NewParser(`books[?(` + tt.input + `)]`)
			_, err := parser.Parse()

			if tt.hasError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// Test more complex expression parsing
func TestParseComplexExpressions(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		hasError bool
	}{
		{
			name:  "logical AND",
			query: `books[?(@.price > 10 && @.category == "fiction")]`,
		},
		{
			name:  "logical OR",
			query: `books[?(@.price < 5 || @.price > 20)]`,
		},
		{
			name:  "comparison operators",
			query: `books[?(@.rating >= 4.5)]`,
		},
		{
			name:  "nested path access",
			query: `store.book[0].author.name`,
		},
		{
			name:     "invalid expression",
			query:    `books[?(@.price >)]`,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(tt.query)
			_, err := parser.Parse()

			if tt.hasError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}
