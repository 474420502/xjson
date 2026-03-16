package query

import (
	"strings"
	"testing"
)

func TestParserSupportsDocumentedSyntax(t *testing.T) {
	testCases := []struct {
		name  string
		path  string
		check func(t *testing.T, tokens []QueryToken)
	}{
		{
			name: "quoted special keys",
			path: `/store/['special.keys']/['user.profile']/name`,
			check: func(t *testing.T, tokens []QueryToken) {
				if len(tokens) != 4 {
					t.Fatalf("expected 4 tokens, got %d", len(tokens))
				}
				if tokens[1].Type != OpKey || tokens[1].Value != "special.keys" {
					t.Fatalf("unexpected second token: %#v", tokens[1])
				}
				if tokens[2].Type != OpKey || tokens[2].Value != "user.profile" {
					t.Fatalf("unexpected third token: %#v", tokens[2])
				}
			},
		},
		{
			name: "negative index",
			path: `/items[-1]`,
			check: func(t *testing.T, tokens []QueryToken) {
				if len(tokens) != 2 || tokens[1].Type != OpIndex || tokens[1].Value != -1 {
					t.Fatalf("unexpected tokens: %#v", tokens)
				}
			},
		},
		{
			name: "slice syntax",
			path: `/items[-3:]`,
			check: func(t *testing.T, tokens []QueryToken) {
				if len(tokens) != 2 || tokens[1].Type != OpSlice {
					t.Fatalf("unexpected tokens: %#v", tokens)
				}
				sl, ok := tokens[1].Value.([2]int)
				if !ok || sl[0] != -3 || sl[1] != -1 {
					t.Fatalf("unexpected slice token value: %#v", tokens[1].Value)
				}
			},
		},
		{
			name: "wildcard function and parent",
			path: `/users[*][@active]/../meta`,
			check: func(t *testing.T, tokens []QueryToken) {
				if len(tokens) != 5 {
					t.Fatalf("expected 5 tokens, got %d", len(tokens))
				}
				if tokens[1].Type != OpWildcard || tokens[2].Type != OpFunc || tokens[3].Type != OpParent {
					t.Fatalf("unexpected token sequence: %#v", tokens)
				}
			},
		},
		{
			name: "recursive descent",
			path: `//price`,
			check: func(t *testing.T, tokens []QueryToken) {
				if len(tokens) != 1 || tokens[0].Type != OpRecursiveKey || tokens[0].Value != "price" {
					t.Fatalf("unexpected tokens: %#v", tokens)
				}
			},
		},
		{
			name: "empty quoted key",
			path: `/['']/name`,
			check: func(t *testing.T, tokens []QueryToken) {
				if len(tokens) != 2 {
					t.Fatalf("expected 2 tokens, got %d", len(tokens))
				}
				if tokens[0].Type != OpKey || tokens[0].Value != "" {
					t.Fatalf("unexpected first token: %#v", tokens[0])
				}
			},
		},
		{
			name: "escaped quoted key",
			path: `/['a\'b\\c']/name`,
			check: func(t *testing.T, tokens []QueryToken) {
				if len(tokens) != 2 {
					t.Fatalf("expected 2 tokens, got %d", len(tokens))
				}
				if tokens[0].Type != OpKey || tokens[0].Value != "a'b\\c" {
					t.Fatalf("unexpected first token value: %#v", tokens[0])
				}
			},
		},
		{
			name: "repeated parent navigation",
			path: `/books[0]/../../meta`,
			check: func(t *testing.T, tokens []QueryToken) {
				if len(tokens) != 5 {
					t.Fatalf("expected 5 tokens, got %d", len(tokens))
				}
				if tokens[2].Type != OpParent || tokens[3].Type != OpParent {
					t.Fatalf("unexpected parent sequence: %#v", tokens)
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parser := NewParser(tc.path)
			tokens, err := parser.Parse()
			if err != nil {
				t.Fatalf("parse failed: %v", err)
			}
			tc.check(t, tokens)
		})
	}
}

func TestParserRejectsInvalidSyntax(t *testing.T) {
	testCases := []struct {
		path       string
		errContain string
	}{
		{path: `/a/..b`, errContain: "invalid parent navigation"},
		{path: `/a/b[a]`, errContain: "invalid index"},
		{path: `['key`, errContain: "unterminated quoted key"},
		{path: `/a@func`, errContain: "invalid path segment"},
		{path: `//`, errContain: "expected key after '//'"},
		{path: `/[@1bad]`, errContain: "invalid function name"},
	}

	for _, tc := range testCases {
		t.Run(tc.path, func(t *testing.T) {
			_, err := NewParser(tc.path).Parse()
			if err == nil {
				t.Fatalf("expected parse error for %q", tc.path)
			}
			if !strings.Contains(err.Error(), tc.errContain) {
				t.Fatalf("expected error containing %q, got %q", tc.errContain, err.Error())
			}
		})
	}
}