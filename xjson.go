package xjson

import (
	"github.com/474420502/xjson/internal/core"
	"github.com/474420502/xjson/internal/engine"
)

// Public type aliases
type Node = core.Node
type NodeType = core.NodeType

// Functional type aliases for public API
type PathFunc = core.PathFunc
type UnaryPathFunc = core.UnaryPathFunc
type PredicateFunc = core.PredicateFunc
type TransformFunc = core.TransformFunc

const (
	ObjectNode  = core.ObjectNode
	ArrayNode   = core.ArrayNode
	StringNode  = core.StringNode
	NumberNode  = core.NumberNode
	BoolNode    = core.BoolNode
	NullNode    = core.NullNode
	InvalidNode = core.InvalidNode
)

// Parse takes a JSON string and returns the root node of the parsed structure.
// The returned Node can be used to navigate and manipulate the JSON data.
func Parse(data string) (Node, error) {
	return engine.ParseJSONToNode(data)
}

// ParseBytes is a convenience wrapper around Parse for byte slices.
func ParseBytes(data []byte) (Node, error) {
	return Parse(string(data))
}

// NewNodeFromInterface creates a new Node from a Go interface{}.
// This is useful for building nodes programmatically.
func NewNodeFromInterface(value interface{}) (Node, error) {
	return engine.NewNodeFromInterface(value, "", nil)
}
