package query

import (
	"fmt"
	"strconv"
)

// Parser holds the state of the query parser.
type Parser struct {
	input  string
	errors []string
}

// NewParser creates a new parser for the given query.
func NewParser(query string) *Parser {
	return &Parser{input: query, errors: []string{}}
}

// Errors returns any errors that occurred during parsing.
func (p *Parser) Errors() []string {
	return p.errors
}

// Parse parses the query and returns a list of query tokens.
func (p *Parser) Parse() ([]QueryToken, error) {
	tokens := make([]QueryToken, 0, 8)
	for i := 0; i < len(p.input); {
		switch p.input[i] {
		case ' ', '\t', '\n', '\r':
			i++
		case '/':
			if i+1 < len(p.input) && p.input[i+1] == '/' {
				i += 2
				name, next, err := parseIdentifierSegment(p.input, i)
				if err != nil {
					return nil, err
				}
				if name == "" {
					return nil, fmt.Errorf("expected key after '//' ")
				}
				tokens = append(tokens, QueryToken{Type: OpRecursiveKey, Value: name})
				i = next
				continue
			}
			i++
		case '.':
			if i+1 >= len(p.input) || p.input[i+1] != '.' {
				return nil, fmt.Errorf("unexpected '.' at position %d", i)
			}
			next := i + 2
			if next < len(p.input) {
				switch p.input[next] {
				case '/', '[', ' ', '\t', '\n', '\r':
				default:
					return nil, fmt.Errorf("invalid parent navigation near position %d", i)
				}
			}
			tokens = append(tokens, QueryToken{Type: OpParent})
			i = next
		case '[':
			token, next, err := parseBracketExpression(p.input, i)
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, token)
			i = next
		case '*':
			tokens = append(tokens, QueryToken{Type: OpWildcard})
			i++
		default:
			segment, next, err := parseIdentifierSegment(p.input, i)
			if err != nil {
				return nil, err
			}
			if segment == "" {
				return nil, fmt.Errorf("unexpected token at position %d", i)
			}
			if idx, ok := tryParseInt(segment); ok {
				tokens = append(tokens, QueryToken{Type: OpIndex, Value: idx})
			} else if isIdentifier(segment) {
				tokens = append(tokens, QueryToken{Type: OpKey, Value: segment})
			} else {
				return nil, fmt.Errorf("invalid path segment %q", segment)
			}
			i = next
		}
	}
	return tokens, nil
}

func parseBracketExpression(input string, start int) (QueryToken, int, error) {
	i := start + 1
	if i >= len(input) {
		return QueryToken{}, 0, fmt.Errorf("unterminated bracket expression")
	}

	switch input[i] {
	case '@':
		name, next, err := parseIdentifierSegment(input, i+1)
		if err != nil {
			return QueryToken{}, 0, err
		}
		if !isIdentifier(name) {
			return QueryToken{}, 0, fmt.Errorf("invalid function name %q", name)
		}
		if next >= len(input) || input[next] != ']' {
			return QueryToken{}, 0, fmt.Errorf("expected ']' after function call")
		}
		return QueryToken{Type: OpFunc, Value: name}, next + 1, nil
	case '*':
		if i+1 >= len(input) || input[i+1] != ']' {
			return QueryToken{}, 0, fmt.Errorf("expected ']' after wildcard")
		}
		return QueryToken{Type: OpWildcard}, i + 2, nil
	case '\'', '"':
		key, next, err := parseQuotedKey(input, i)
		if err != nil {
			return QueryToken{}, 0, err
		}
		if next >= len(input) || input[next] != ']' {
			return QueryToken{}, 0, fmt.Errorf("expected ']' after quoted key")
		}
		return QueryToken{Type: OpKey, Value: key}, next + 1, nil
	default:
		colon := -1
		end := i
		for end < len(input) && input[end] != ']' {
			if input[end] == ':' && colon == -1 {
				colon = end
			}
			end++
		}
		if end >= len(input) {
			return QueryToken{}, 0, fmt.Errorf("unterminated bracket expression")
		}
		content := input[i:end]
		if colon == -1 {
			idx, ok := tryParseInt(content)
			if !ok {
				return QueryToken{}, 0, fmt.Errorf("invalid index %q", content)
			}
			return QueryToken{Type: OpIndex, Value: idx}, end + 1, nil
		}

		left := content[:colon-i]
		right := content[colon-i+1:]
		startIdx := 0
		if left != "" {
			parsed, ok := tryParseInt(left)
			if !ok {
				return QueryToken{}, 0, fmt.Errorf("invalid slice start %q", left)
			}
			startIdx = parsed
		}
		endIdx := -1
		if right != "" {
			parsed, ok := tryParseInt(right)
			if !ok {
				return QueryToken{}, 0, fmt.Errorf("invalid slice end %q", right)
			}
			endIdx = parsed
		}
		return QueryToken{Type: OpSlice, Value: [2]int{startIdx, endIdx}}, end + 1, nil
	}
}

func parseIdentifierSegment(input string, start int) (string, int, error) {
	i := start
	for i < len(input) {
		switch input[i] {
		case '/', '[', ']', '.', ' ', '\t', '\n', '\r':
			return input[start:i], i, nil
		default:
			i++
		}
	}
	return input[start:i], i, nil
}

func parseQuotedKey(input string, start int) (string, int, error) {
	quote := input[start]
	i := start + 1
	buf := make([]byte, 0, 16)
	for i < len(input) {
		if input[i] == '\\' {
			if i+1 >= len(input) {
				return "", 0, fmt.Errorf("unterminated escape in quoted key")
			}
			buf = append(buf, input[i+1])
			i += 2
			continue
		}
		if input[i] == quote {
			return string(buf), i + 1, nil
		}
		buf = append(buf, input[i])
		i++
	}
	return "", 0, fmt.Errorf("unterminated quoted key")
}

func tryParseInt(s string) (int, bool) {
	if s == "" {
		return 0, false
	}
	for i := 0; i < len(s); i++ {
		if s[i] == '-' && i == 0 {
			continue
		}
		if s[i] < '0' || s[i] > '9' {
			return 0, false
		}
	}
	v, err := strconv.Atoi(s)
	return v, err == nil
}

func isIdentifier(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_' {
			continue
		}
		if i > 0 && c >= '0' && c <= '9' {
			continue
		}
		return false
	}
	return true
}
