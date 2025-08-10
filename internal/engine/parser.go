package engine

import (
	"fmt"

	"github.com/474420502/xjson/internal/core"
)

type parser struct {
	raw   []byte
	pos   int
	funcs *map[string]core.UnaryPathFunc
}

func newParser(data []byte) *parser {
	// Initialize funcs map for the entire parsing session
	funcs := make(map[string]core.UnaryPathFunc)
	return &parser{
		raw:   data,
		pos:   0,
		funcs: &funcs,
	}
}

// Parse is the main entry point for the internal parser.
func (p *parser) Parse() (core.Node, error) {
	p.skipWhitespace()
	if p.pos >= len(p.raw) {
		return nil, fmt.Errorf("empty json")
	}

	value := p.parseValue(nil)
	if value.Error() != nil {
		return nil, value.Error()
	}

	p.skipWhitespace()
	if p.pos < len(p.raw) {
		return nil, fmt.Errorf("trailing data found at position %d", p.pos)
	}

	return value, nil
}

func (p *parser) parseValue(parent core.Node) core.Node {
	p.skipWhitespace()

	if p.pos >= len(p.raw) {
		return newInvalidNode(fmt.Errorf("unexpected end of json"))
	}

	switch p.raw[p.pos] {
	case '{':
		return p.parseObject(parent)
	case '[':
		return p.parseArray(parent)
	case '"':
		return p.parseString(parent)
	case 't', 'f':
		return p.parseBool(parent)
	case 'n':
		return p.parseNull(parent)
	case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return p.parseNumber(parent)
	default:
		return newInvalidNode(fmt.Errorf("invalid character '%c' looking for beginning of value", p.raw[p.pos]))
	}
}

func (p *parser) parseObject(parent core.Node) core.Node {
	start := p.pos
	p.pos++ // consume '{'
	level := 1
	for p.pos < len(p.raw) && level > 0 {
		switch p.raw[p.pos] {
		case '{':
			level++
		case '}':
			level--
		case '"': // Skip strings
			p.pos++
			for p.pos < len(p.raw) {
				if p.raw[p.pos] == '\\' {
					p.pos++
				} else if p.raw[p.pos] == '"' {
					break
				}
				p.pos++
			}
		}
		p.pos++
	}

	if level != 0 {
		return newInvalidNode(fmt.Errorf("unbalanced curly braces"))
	}
	return newObjectNode(p.raw, start, p.pos, parent, p.funcs)
}

func (p *parser) parseArray(parent core.Node) core.Node {
	start := p.pos
	p.pos++ // consume '['
	level := 1
	for p.pos < len(p.raw) && level > 0 {
		switch p.raw[p.pos] {
		case '[':
			level++
		case ']':
			level--
		case '"': // Skip strings
			p.pos++
			for p.pos < len(p.raw) {
				if p.raw[p.pos] == '\\' {
					p.pos++
				} else if p.raw[p.pos] == '"' {
					break
				}
				p.pos++
			}
		}
		p.pos++
	}
	if level != 0 {
		return newInvalidNode(fmt.Errorf("unbalanced square brackets"))
	}
	return newArrayNode(p.raw, start, p.pos, parent, p.funcs)
}

func (p *parser) parseString(parent core.Node) core.Node {
	p.pos++ // consume opening quote
	start := p.pos
	var val []byte
	for p.pos < len(p.raw) {
		b := p.raw[p.pos]
		if b == '\\' {
			p.pos++
			if p.pos >= len(p.raw) {
				return newInvalidNode(fmt.Errorf("unexpected end of input after backslash"))
			}
			// This is a simplified unescape. A full implementation would handle \uXXXX, etc.
			val = append(val, p.raw[p.pos])
			p.pos++
		} else if b == '"' {
			end := p.pos
			p.pos++ // consume closing quote
			// The `value` field should contain the unescaped string.
			// The `raw` slice still points to the original data with escape sequences.
			return &stringNode{
				baseNode: newBaseNode(p.raw, start, end, parent, p.funcs),
				value:    string(val),
			}
		} else {
			val = append(val, b)
			p.pos++
		}
	}
	return newInvalidNode(fmt.Errorf("unterminated string"))
}

func (p *parser) parseNumber(parent core.Node) core.Node {
	start := p.pos
	for p.pos < len(p.raw) {
		b := p.raw[p.pos]
		if (b >= '0' && b <= '9') || b == '.' || b == 'e' || b == 'E' || b == '-' || b == '+' {
			p.pos++
		} else {
			break
		}
	}
	end := p.pos
	return &numberNode{baseNode: newBaseNode(p.raw, start, end, parent, p.funcs)}
}

func (p *parser) parseBool(parent core.Node) core.Node {
	start := p.pos
	var val bool
	if p.pos+4 <= len(p.raw) && string(p.raw[p.pos:p.pos+4]) == "true" {
		p.pos += 4
		val = true
	} else if p.pos+5 <= len(p.raw) && string(p.raw[p.pos:p.pos+5]) == "false" {
		p.pos += 5
		val = false
	} else {
		return newInvalidNode(fmt.Errorf("invalid boolean literal at pos %d", p.pos))
	}
	return &boolNode{
		baseNode: newBaseNode(p.raw, start, p.pos, parent, p.funcs),
		value:    val,
	}
}

func (p *parser) parseNull(parent core.Node) core.Node {
	start := p.pos
	if p.pos+4 > len(p.raw) || string(p.raw[p.pos:p.pos+4]) != "null" {
		return newInvalidNode(fmt.Errorf("invalid null literal at pos %d", p.pos))
	}
	p.pos += 4
	return &nullNode{baseNode: newBaseNode(p.raw, start, p.pos, parent, p.funcs)}
}

func (p *parser) skipWhitespace() {
	for p.pos < len(p.raw) {
		switch p.raw[p.pos] {
		case ' ', '\t', '\n', '\r':
			p.pos++
		default:
			return
		}
	}
}
