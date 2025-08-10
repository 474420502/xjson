package engine

import (
	"encoding/json"
	"fmt"

	"github.com/474420502/xjson/internal/core"
)

// objectNode represents a JSON object.
type objectNode struct {
	baseNode
	children map[string]core.Node
}

// newObjectNode creates a new object node.
// Note: children are parsed lazily.
func newObjectNode(raw []byte, start, end int, parent core.Node, funcs *map[string]core.UnaryPathFunc) *objectNode {
	return &objectNode{
		baseNode: newBaseNode(raw, start, end, parent, funcs),
	}
}

func (n *objectNode) Type() core.NodeType {
	return core.Object
}

func (n *objectNode) Query(path string) core.Node {
	if n.err != nil {
		return newInvalidNode(n.err)
	}
	return applySimpleQuery(n, path)
}

// ensureParsed parses the children of the object if they haven't been parsed yet.
func (n *objectNode) ensureParsed() {
	if n.err != nil || n.children != nil {
		return
	}

	p := &parser{
		raw:   n.raw,
		pos:   n.start + 1, // Skip '{'
		funcs: n.funcs,
	}

	n.children = make(map[string]core.Node)

	for p.pos < n.end-1 {
		p.skipWhitespace()

		// Key
		keyNode := p.parseString(n)
		if keyNode.Error() != nil {
			n.setError(keyNode.Error())
			return
		}

		p.skipWhitespace()

		// Colon
		if p.pos >= n.end-1 || p.raw[p.pos] != ':' {
			n.setError(fmt.Errorf("expecting : after object key at pos %d", p.pos))
			return
		}
		p.pos++ // consume ':'
		p.skipWhitespace()

		// Value
		valueNode := p.parseValue(n)
		if valueNode.Error() != nil {
			n.setError(valueNode.Error())
			return
		}
		n.children[keyNode.MustString()] = valueNode

		p.skipWhitespace()
		if p.pos < n.end-1 && p.raw[p.pos] == ',' {
			p.pos++
			p.skipWhitespace()
		} else {
			break // No comma, should be end of object
		}
	}

	p.skipWhitespace()
	if p.pos > n.end-1 {
		n.setError(fmt.Errorf("object not properly terminated, expecting } at pos %d", n.end-1))
	}
}

func (n *objectNode) Get(key string) core.Node {
	if n.err != nil {
		return newInvalidNode(n.err)
	}
	n.ensureParsed()

	child, ok := n.children[key]
	if !ok {
		return newInvalidNode(fmt.Errorf("key not found: %s", key))
	}
	// Set the parent of the child to this node
	if c, ok := child.(*objectNode); ok {
		c.parent = n
	} else if c, ok := child.(*arrayNode); ok {
		c.parent = n
	} else if c, ok := child.(*stringNode); ok {
		c.parent = n
	} else if c, ok := child.(*numberNode); ok {
		c.parent = n
	} else if c, ok := child.(*boolNode); ok {
		c.parent = n
	} else if c, ok := child.(*nullNode); ok {
		c.parent = n
	}
	return child
}

func (n *objectNode) ForEach(fn func(keyOrIndex interface{}, value core.Node)) {
	if n.err != nil {
		return
	}
	n.ensureParsed()
	for k, v := range n.children {
		fn(k, v)
	}
}

func (n *objectNode) Len() int {
	if n.err != nil {
		return 0
	}
	n.ensureParsed()
	return len(n.children)
}

func (n *objectNode) AsMap() map[string]core.Node {
	if n.err != nil {
		return nil
	}
	n.ensureParsed()
	return n.children
}

func (n *objectNode) MustAsMap() map[string]core.Node {
	if n.err != nil {
		panic(n.err)
	}
	n.ensureParsed()
	return n.children
}

// Keys returns the object's keys in arbitrary order.
func (n *objectNode) Keys() []string {
	if n.err != nil {
		return nil
	}
	n.ensureParsed()
	keys := make([]string, 0, len(n.children))
	for k := range n.children {
		keys = append(keys, k)
	}
	return keys
}

func (n *objectNode) Interface() interface{} {
	if n.err != nil {
		return nil
	}
	n.ensureParsed()
	m := make(map[string]interface{})
	for k, v := range n.children {
		m[k] = v.Interface()
	}
	return m
}

func (n *objectNode) String() string {
	n.ensureParsed()
	b, err := json.Marshal(n.Interface())
	if err != nil {
		return n.baseNode.String()
	}
	return string(b)
}

func (n *objectNode) SetValue(value interface{}) core.Node {
	if n.err != nil {
		return n
	}
	newChildren, ok := value.(map[string]core.Node)
	if !ok {
		n.setError(fmt.Errorf("SetValue on object requires a map[string]core.Node"))
		return n
	}
	n.children = newChildren
	return n
}

func (n *objectNode) Set(key string, value interface{}) core.Node {
	if n.err != nil {
		return n
	}
	n.ensureParsed()
	child, err := NewNodeFromInterface(value)
	if err != nil {
		return newInvalidNode(err)
	}
	if n.children == nil {
		n.children = make(map[string]core.Node)
	}
	// Set parent of the new child
	if c, ok := child.(*objectNode); ok {
		c.parent = n
	} else if c, ok := child.(*arrayNode); ok {
		c.parent = n
	} else if c, ok := child.(*stringNode); ok {
		c.parent = n
	} else if c, ok := child.(*numberNode); ok {
		c.parent = n
	} else if c, ok := child.(*boolNode); ok {
		c.parent = n
	} else if c, ok := child.(*nullNode); ok {
		c.parent = n
	}
	n.children[key] = child
	return n
}

func (n *objectNode) RegisterFunc(name string, fn core.UnaryPathFunc) core.Node {
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

func (n *objectNode) RemoveFunc(name string) core.Node {
	if n.err != nil {
		return newInvalidNode(n.err)
	}
	if n.funcs != nil && *n.funcs != nil {
		delete(*n.funcs, name)
	}
	return n
}

func (n *objectNode) CallFunc(name string) core.Node {
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

func (n *objectNode) Apply(fn core.PathFunc) core.Node {
	if n.err != nil {
		return newInvalidNode(n.err)
	}
	switch f := fn.(type) {
	case core.UnaryPathFunc:
		return f(n)
	case core.PredicateFunc:
		return newInvalidNode(fmt.Errorf("predicate apply requires array node"))
	case core.TransformFunc:
		return newInvalidNode(fmt.Errorf("transform apply requires array node"))
	default:
		return newInvalidNode(fmt.Errorf("unsupported function type"))
	}
}

// GetParent returns the parent node
func (n *objectNode) GetParent() core.Node {
	return n.parent
}
