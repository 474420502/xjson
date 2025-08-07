package xjson

import (
	"github.com/474420502/xjson/internal/core" // Import core types
	"github.com/474420502/xjson/internal/engine/parser"
)

type Node = core.Node         // Use type alias for Node
type NodeType = core.NodeType // Use type alias for NodeType

const (
	ObjectNode  = core.ObjectNode
	ArrayNode   = core.ArrayNode
	StringNode  = core.StringNode
	NumberNode  = core.NumberNode
	BoolNode    = core.BoolNode
	NullNode    = core.NullNode
	InvalidNode = core.InvalidNode
)

type nodeWrapper struct {
	core.Node // Embed core.Node
}

func (w *nodeWrapper) Type() NodeType {
	return NodeType(w.Node.Type())
}

func (w *nodeWrapper) Get(key string) core.Node {
	return &nodeWrapper{w.Node.Get(key)}
}

func (w *nodeWrapper) Index(i int) core.Node {
	return &nodeWrapper{w.Node.Index(i)}
}

func (w *nodeWrapper) Query(path string) core.Node {
	return &nodeWrapper{w.Node.Query(path)}
}

func (w *nodeWrapper) ForEach(iterator func(keyOrIndex interface{}, value core.Node)) {
	w.Node.ForEach(func(keyOrIndex interface{}, value core.Node) {
		iterator(keyOrIndex, &nodeWrapper{value})
	})
}

func (w *nodeWrapper) Array() []core.Node {
	engineNodes := w.Node.Array()
	nodes := make([]core.Node, len(engineNodes))
	for i, n := range engineNodes {
		nodes[i] = &nodeWrapper{n}
	}
	return nodes
}

func (w *nodeWrapper) MustArray() []core.Node {
	return w.Array()
}

func (w *nodeWrapper) Filter(fn func(core.Node) bool) core.Node {
	return &nodeWrapper{w.Node.Filter(func(n core.Node) bool {
		return fn(&nodeWrapper{n})
	})}
}

func (w *nodeWrapper) Map(fn func(core.Node) interface{}) core.Node {
	return &nodeWrapper{w.Node.Map(func(n core.Node) interface{} {
		return fn(&nodeWrapper{n})
	})}
}

func (w *nodeWrapper) Set(key string, value interface{}) core.Node {
	w.Node.Set(key, value)
	return w
}

func (w *nodeWrapper) Append(value interface{}) core.Node {
	w.Node.Append(value)
	return w
}

func (w *nodeWrapper) RawFloat() (float64, bool) {
	return w.Node.RawFloat()
}

func (w *nodeWrapper) RawString() (string, bool) {
	return w.Node.RawString()
}

func (w *nodeWrapper) Strings() []string {
	return w.Node.Strings()
}

func (w *nodeWrapper) Contains(value string) bool {
	return w.Node.Contains(value)
}

func (w *nodeWrapper) Func(name string, fn func(core.Node) core.Node) core.Node {
	w.Node.Func(name, func(n core.Node) core.Node {
		return fn(&nodeWrapper{n}).(*nodeWrapper).Node
	})
	return w
}

func (w *nodeWrapper) CallFunc(name string) core.Node {
	return &nodeWrapper{w.Node.CallFunc(name)}
}

func (w *nodeWrapper) RemoveFunc(name string) core.Node {
	w.Node.RemoveFunc(name)
	return w
}

func (w *nodeWrapper) GetFuncs() *map[string]func(core.Node) core.Node {
	return w.Node.GetFuncs()
}

// Parse takes a JSON string and returns the root node of the parsed structure.
// It is the main entry point for the xjson library.
func Parse(data string) (core.Node, error) {
	node, err := parser.ParseJSONToNode(data) // Call the new parser function
	if err != nil {
		return nil, err
	}
	return &nodeWrapper{node}, nil
}

// ParseBytes is a convenience wrapper around Parse for byte slices.
func ParseBytes(data []byte) (core.Node, error) {
	return Parse(string(data))
}
