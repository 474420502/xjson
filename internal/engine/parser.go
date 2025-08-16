package engine

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/474420502/xjson/internal/core"
)

type parser struct {
	data  []byte
	pos   int
	funcs *map[string]core.UnaryPathFunc
}

func newParser(data []byte, funcs *map[string]core.UnaryPathFunc) *parser {
	return &parser{data: data, funcs: funcs}
}

func (p *parser) Parse() (core.Node, error) {
	p.skipWhitespace()
	if p.pos >= len(p.data) {
		return nil, fmt.Errorf("empty json")
	}
	n := p.doParse(nil)
	if !n.IsValid() {
		return nil, n.Error()
	}
	return n, nil
}

func (p *parser) parseValue(parent core.Node) core.Node {
	p.skipWhitespace()
	if p.pos >= len(p.data) {
		return newInvalidNode(fmt.Errorf("unexpected end of json"))
	}
	return p.doParse(parent)
}

func (p *parser) doParse(parent core.Node) core.Node {
	switch p.data[p.pos] {
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
	}
	return newInvalidNode(fmt.Errorf("invalid character '%c' looking for beginning of value", p.data[p.pos]))
}

func (p *parser) parseObject(parent core.Node) core.Node {
	start := p.pos
	p.pos++ // skip '{'

	// Find the end of the object without parsing content
	braceCount := 1
	inString := false
	escaped := false

	for p.pos < len(p.data) && braceCount > 0 {
		c := p.data[p.pos]

		if inString {
			if escaped {
				escaped = false
			} else if c == '\\' {
				escaped = true
			} else if c == '"' {
				inString = false
			}
		} else {
			switch c {
			case '"':
				inString = true
			case '{':
				braceCount++
			case '}':
				braceCount--
			}
		}
		p.pos++
	}

	if braceCount > 0 {
		return newInvalidNode(fmt.Errorf("unterminated object"))
	}

	raw := p.data[start:p.pos]
	node := NewObjectNode(parent, raw, p.funcs).(*objectNode)
	node.start = 0
	node.end = len(raw)
	// Mark as not yet parsed for lazy evaluation
	node.value = nil
	node.isDirty = false
	return node
}

// parseObjectFull parses object content immediately (for lazyParse)
func (p *parser) parseObjectFull(parent core.Node) core.Node {
	start := p.pos
	p.pos++ // skip '{'
	p.skipWhitespace()

	node := NewObjectNode(parent, nil, p.funcs).(*objectNode)
	node.isDirty = true

	for p.pos < len(p.data) {
		if p.data[p.pos] == '}' {
			p.pos++
			node.raw = p.data[start:p.pos]
			node.start = 0
			node.end = len(node.raw)
			return node
		}

		keyNode := p.parseString(node)
		if !keyNode.IsValid() {
			return keyNode
		}
		key, _ := keyNode.RawString()

		p.skipWhitespace()
		if p.pos >= len(p.data) || p.data[p.pos] != ':' {
			return newInvalidNode(fmt.Errorf("missing ':' after object key"))
		}
		p.pos++ // skip ':'

		valueNode := p.parseValue(node)
		if !valueNode.IsValid() {
			return valueNode
		}

		if node.value == nil {
			node.value = make(map[string]core.Node)
		}
		node.value[key] = valueNode

		p.skipWhitespace()
		if p.data[p.pos] == '}' {
			p.pos++
			node.raw = p.data[start:p.pos]
			node.start = 0
			node.end = len(node.raw)
			return node
		}

		if p.data[p.pos] != ',' {
			return newInvalidNode(fmt.Errorf("missing ',' after object value"))
		}
		p.pos++ // skip ','
		p.skipWhitespace()
	}

	return newInvalidNode(fmt.Errorf("unterminated object"))
}

func (p *parser) parseArray(parent core.Node) core.Node {
	start := p.pos
	p.pos++ // skip '['
	p.skipWhitespace()

	node := NewArrayNode(parent, nil, p.funcs).(*arrayNode)
	node.isDirty = true

	for p.pos < len(p.data) {
		if p.data[p.pos] == ']' {
			p.pos++
			node.raw = p.data[start:p.pos]
			node.start = 0
			node.end = len(node.raw)
			return node
		}

		valueNode := p.parseValue(node)
		if !valueNode.IsValid() {
			return valueNode
		}
		node.value = append(node.value, valueNode)

		p.skipWhitespace()
		if p.data[p.pos] == ']' {
			p.pos++
			node.raw = p.data[start:p.pos]
			node.start = 0
			node.end = len(node.raw)
			return node
		}

		if p.data[p.pos] != ',' {
			return newInvalidNode(fmt.Errorf("missing ',' after array value"))
		}
		p.pos++ // skip ','
		p.skipWhitespace()
	}
	return newInvalidNode(fmt.Errorf("unterminated array"))
}

func (p *parser) parseString(parent core.Node) core.Node {
	start := p.pos
	p.pos++ // skip '"'
	end := -1
	for i := p.pos; i < len(p.data); i++ {
		if p.data[i] == '"' {
			isEscaped := i > 0 && p.data[i-1] == '\\'
			if !isEscaped {
				end = i
				break
			}
		}
	}

	if end == -1 {
		return newInvalidNode(fmt.Errorf("unterminated string"))
	}
	p.pos = end + 1
	raw := p.data[start:p.pos]
	// Check if unescape is needed
	needsUnescape := bytes.IndexByte(p.data[start+1:end], '\\') != -1
	// start/end for unquoted region relative to raw slice
	// raw[0] == '"', so unquoted starts at 1 and ends at len(raw)-1
	node := NewRawStringNode(parent, raw, 1, len(raw)-1, needsUnescape, p.funcs).(*stringNode)
	return node
}

func (p *parser) parseNumber(parent core.Node) core.Node {
	start := p.pos
	for p.pos < len(p.data) {
		c := p.data[p.pos]
		if (c < '0' || c > '9') && c != '.' && c != 'e' && c != 'E' && c != '+' && c != '-' {
			break
		}
		p.pos++
	}
	raw := p.data[start:p.pos]
	n := NewNumberNode(parent, raw, p.funcs).(*numberNode)
	n.start = 0
	n.end = len(raw)
	return n
}

func (p *parser) parseBool(parent core.Node) core.Node {
	if bytes.HasPrefix(p.data[p.pos:], []byte("true")) {
		raw := p.data[p.pos : p.pos+4]
		p.pos += 4
		node := NewBoolNode(parent, true, p.funcs).(*boolNode)
		node.raw = raw
		node.start = 0
		node.end = len(raw)
		return node
	}
	if bytes.HasPrefix(p.data[p.pos:], []byte("false")) {
		raw := p.data[p.pos : p.pos+5]
		p.pos += 5
		node := NewBoolNode(parent, false, p.funcs).(*boolNode)
		node.raw = raw
		node.start = 0
		node.end = len(raw)
		return node
	}
	return newInvalidNode(fmt.Errorf("invalid boolean"))
}

func (p *parser) parseNull(parent core.Node) core.Node {
	if bytes.HasPrefix(p.data[p.pos:], []byte("null")) {
		raw := p.data[p.pos : p.pos+4]
		p.pos += 4
		node := NewNullNode(parent, p.funcs).(*nullNode)
		node.raw = raw
		node.start = 0
		node.end = len(raw)
		return node
	}
	return newInvalidNode(fmt.Errorf("invalid null"))
}

func (p *parser) skipWhitespace() {
	for p.pos < len(p.data) {
		c := p.data[p.pos]
		if c == ' ' || c == '\n' || c == '\r' || c == '\t' {
			p.pos++
		} else {
			break
		}
	}
}

func unescape(data []byte) ([]byte, error) {
	if bytes.IndexByte(data, '\\') == -1 {
		return data, nil
	}
	var buf bytes.Buffer
	i := 0
	for i < len(data) {
		if data[i] == '\\' {
			i++
			if i >= len(data) {
				return nil, fmt.Errorf("invalid escape sequence at end of string")
			}
			switch data[i] {
			case '"', '\\', '/':
				buf.WriteByte(data[i])
			case 'b':
				buf.WriteByte('\b')
			case 'f':
				buf.WriteByte('\f')
			case 'n':
				buf.WriteByte('\n')
			case 'r':
				buf.WriteByte('\r')
			case 't':
				buf.WriteByte('\t')
			case 'u':
				if i+4 >= len(data) {
					return nil, fmt.Errorf("invalid unicode escape sequence: not enough digits")
				}
				val, err := strconv.ParseInt(string(data[i+1:i+5]), 16, 32)
				if err != nil {
					return nil, fmt.Errorf("invalid unicode escape sequence: %w", err)
				}
				buf.WriteRune(rune(val))
				i += 4
			default:
				return nil, fmt.Errorf("invalid escape character: %c", data[i])
			}
		} else {
			buf.WriteByte(data[i])
		}
		i++
	}
	return buf.Bytes(), nil
}
