package engine

import (
	"encoding/json"
	"fmt"

	"github.com/474420502/xjson/internal/core"
)

// arrayNode represents a JSON array.
type arrayNode struct {
	baseNode
	children []core.Node
}

// newArrayNode creates a new array node.
// Note: children are parsed lazily.
func newArrayNode(raw []byte, start, end int, parent core.Node, funcs *map[string]core.UnaryPathFunc) *arrayNode {
	return &arrayNode{
		baseNode: newBaseNode(raw, start, end, parent, funcs),
	}
}

func (n *arrayNode) Type() core.NodeType {
	return core.Array
}

func (n *arrayNode) Query(path string) core.Node {
	if n.err != nil {
		return newInvalidNode(n.err)
	}
	return applySimpleQuery(n, path)
}

// ensureParsed parses the children of the array if they haven't been parsed yet.
func (n *arrayNode) ensureParsed() {
	if n.err != nil || n.children != nil {
		return
	}

	p := &parser{
		raw:   n.raw,
		pos:   n.start + 1, // Skip '['
		funcs: n.funcs,
	}

	n.children = make([]core.Node, 0)

	for p.pos < n.end-1 {
		p.skipWhitespace()
		if p.raw[p.pos] == ']' {
			break
		}

		// Value
		valueNode := p.parseValue(n)
		if valueNode.Error() != nil {
			n.setError(valueNode.Error())
			return
		}
		// Set parent of the child node
		if c, ok := valueNode.(*objectNode); ok {
			c.parent = n
		} else if c, ok := valueNode.(*arrayNode); ok {
			c.parent = n
		} else if c, ok := valueNode.(*stringNode); ok {
			c.parent = n
		} else if c, ok := valueNode.(*numberNode); ok {
			c.parent = n
		} else if c, ok := valueNode.(*boolNode); ok {
			c.parent = n
		} else if c, ok := valueNode.(*nullNode); ok {
			c.parent = n
		}
		n.children = append(n.children, valueNode)

		p.skipWhitespace()
		if p.pos < n.end-1 && p.raw[p.pos] == ',' {
			p.pos++
			p.skipWhitespace()
		} else {
			break // No comma, should be end of array
		}
	}
	p.skipWhitespace()
	if p.pos > n.end-1 {
		n.setError(fmt.Errorf("array not properly terminated, expecting ] at pos %d", n.end-1))
	}
}

func (n *arrayNode) Index(i int) core.Node {
	if n.err != nil {
		return n
	}
	n.ensureParsed()

	if i < 0 {
		i = len(n.children) + i
	}

	if i < 0 || i >= len(n.children) {
		return newInvalidNode(fmt.Errorf("index out of bounds: %d", i))
	}
	return n.children[i]
}

func (n *arrayNode) ForEach(fn func(keyOrIndex interface{}, value core.Node)) {
	if n.err != nil {
		return
	}
	n.ensureParsed()
	for i, v := range n.children {
		fn(i, v)
	}
}

func (n *arrayNode) Len() int {
	if n.err != nil {
		return 0
	}
	n.ensureParsed()
	return len(n.children)
}

func (n *arrayNode) Array() []core.Node {
	if n.err != nil {
		return nil
	}
	n.ensureParsed()
	return n.children
}

func (n *arrayNode) MustArray() []core.Node {
	if n.err != nil {
		panic(n.err)
	}
	n.ensureParsed()
	return n.children
}

// Keys for array returns nil (arrays don't have string keys).
func (n *arrayNode) Keys() []string { return nil }

func (n *arrayNode) Interface() interface{} {
	if n.err != nil {
		return nil
	}
	n.ensureParsed()
	s := make([]interface{}, len(n.children))
	for i, v := range n.children {
		s[i] = v.Interface()
	}
	return s
}

func (n *arrayNode) String() string {
	n.ensureParsed()
	b, err := json.Marshal(n.Interface())
	if err != nil {
		return n.baseNode.String()
	}
	return string(b)
}

func (n *arrayNode) SetValue(value interface{}) core.Node {
	if n.err != nil {
		return n
	}
	newChildren, ok := value.([]core.Node)
	if !ok {
		n.setError(fmt.Errorf("SetValue on array requires a []core.Node"))
		return n
	}
	n.children = newChildren
	return n
}

func (n *arrayNode) RegisterFunc(name string, fn core.UnaryPathFunc) core.Node {
	if n.err != nil {
		return newInvalidNode(n.err)
	}
	if n.funcs == nil || *n.funcs == nil {
		m := make(map[string]core.UnaryPathFunc)
		n.funcs = &m
	}
	(*n.funcs)[name] = fn
	return n
}

func (n *arrayNode) RemoveFunc(name string) core.Node {
	if n.err != nil {
		return newInvalidNode(n.err)
	}
	if n.funcs != nil && *n.funcs != nil {
		delete(*n.funcs, name)
	}
	return n
}

func (n *arrayNode) Filter(fn core.PredicateFunc) core.Node {
	if n.err != nil {
		return newInvalidNode(n.err)
	}
	n.ensureParsed()
	out := make([]core.Node, 0, len(n.children))
	for _, c := range n.children {
		if fn(c) {
			out = append(out, c)
		}
	}
	return &arrayNode{baseNode: n.baseNode, children: out}
}

func (n *arrayNode) Map(fn core.TransformFunc) core.Node {
	if n.err != nil {
		return newInvalidNode(n.err)
	}
	n.ensureParsed()
	out := make([]core.Node, 0, len(n.children))
	for _, c := range n.children {
		v := fn(c)
		nn, err := NewNodeFromInterface(v)
		if err != nil {
			return newInvalidNode(err)
		}
		out = append(out, nn)
	}
	return &arrayNode{baseNode: n.baseNode, children: out}
}

func (n *arrayNode) Append(value interface{}) core.Node {
	if n.err != nil {
		return n
	}
	n.ensureParsed()
	nn, err := NewNodeFromInterface(value)
	if err != nil {
		return newInvalidNode(err)
	}
	// Set parent of the new child
	if c, ok := nn.(*objectNode); ok {
		c.parent = n
	} else if c, ok := nn.(*arrayNode); ok {
		c.parent = n
	} else if c, ok := nn.(*stringNode); ok {
		c.parent = n
	} else if c, ok := nn.(*numberNode); ok {
		c.parent = n
	} else if c, ok := nn.(*boolNode); ok {
		c.parent = n
	} else if c, ok := nn.(*nullNode); ok {
		c.parent = n
	}
	n.children = append(n.children, nn)
	return n
}

func (n *arrayNode) Strings() []string {
	if n.err != nil {
		return nil
	}
	n.ensureParsed()
	res := make([]string, 0, len(n.children))
	for _, c := range n.children {
		if s, ok := c.RawString(); ok {
			res = append(res, s)
		} else {
			res = append(res, c.String())
		}
	}
	return res
}

func (n *arrayNode) Contains(value string) bool {
	if n.err != nil {
		return false
	}
	n.ensureParsed()
	for _, c := range n.children {
		if s, ok := c.RawString(); ok && s == value {
			return true
		}
		if c.String() == value {
			return true
		}
	}
	return false
}

func (n *arrayNode) CallFunc(name string) core.Node {
	if n.err != nil {
		return newInvalidNode(n.err)
	}
	if n.funcs == nil || *n.funcs == nil {
		return newInvalidNode(fmt.Errorf("func not found: %s", name))
	}
	if fn, ok := (*n.funcs)[name]; ok && fn != nil {
		return fn(n)
	}
	return newInvalidNode(fmt.Errorf("func not found: %s", name))
}

func (n *arrayNode) Apply(fn core.PathFunc) core.Node {
	if n.err != nil {
		return newInvalidNode(n.err)
	}
	switch f := fn.(type) {
	case core.UnaryPathFunc:
		return f(n)
	case core.PredicateFunc:
		return n.Filter(f)
	case core.TransformFunc:
		return n.Map(f)
	default:
		return newInvalidNode(fmt.Errorf("unsupported function type"))
	}
}

// GetParent returns the parent node
func (n *arrayNode) GetParent() core.Node {
	return n.parent
}
