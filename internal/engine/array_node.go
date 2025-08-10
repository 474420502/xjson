package engine

import (
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
	newFuncs := make(map[string]core.UnaryPathFunc)
	if n.funcs != nil && *n.funcs != nil {
		for k, v := range *n.funcs {
			newFuncs[k] = v
		}
	}
	newFuncs[name] = fn
	n.funcs = &newFuncs
	return n
}

func (n *arrayNode) RemoveFunc(name string) core.Node {
	if n.err != nil {
		return newInvalidNode(n.err)
	}
	if n.funcs != nil && *n.funcs != nil {
		newFuncs := make(map[string]core.UnaryPathFunc)
		for k, v := range *n.funcs {
			if k != name {
				newFuncs[k] = v
			}
		}
		n.funcs = &newFuncs
	}
	return n
}
