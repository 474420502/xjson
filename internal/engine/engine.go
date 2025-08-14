package engine

import (
	"fmt"
	"github.com/474420502/xjson/internal/core"
)

// MustParse parses the JSON data and returns a fully parsed Node tree.
// All nodes in the tree are immediately parsed.
func MustParse(data []byte) (core.Node, error) {
	return MustParseWithFuncs(data, nil)
}

// MustParseWithFuncs parses the JSON data with custom functions and returns a fully parsed Node tree.
func MustParseWithFuncs(data []byte, funcs *map[string]core.UnaryPathFunc) (core.Node, error) {
	if funcs == nil {
		funcs = &map[string]core.UnaryPathFunc{}
	}
	p := newParser(data, funcs)
	node, err := p.Parse()
	if err != nil {
		return nil, err
	}
	if err := node.Error(); err != nil {
		return nil, err
	}
	return node, nil
}

// Parse parses the JSON data and returns a Node tree with lazy parsing.
// Nodes are parsed on-demand when accessed.
func Parse(data []byte) (core.Node, error) {
	return ParseWithFuncs(data, nil)
}

// ParseWithFuncs parses the JSON data with custom functions and returns a Node tree with lazy parsing.
// Nodes are parsed on-demand when accessed.
func ParseWithFuncs(data []byte, funcs *map[string]core.UnaryPathFunc) (core.Node, error) {
	if funcs == nil {
		funcs = &map[string]core.UnaryPathFunc{}
	}
	
	// Check first non-whitespace character to determine root node type
	firstChar := getFirstNonWhitespaceChar(data)
	
	// Create appropriate root node with the raw data but don't parse it yet
	// The parsing will happen on-demand when nodes are accessed
	var node core.Node
	switch firstChar {
	case '{':
		node = NewObjectNode(nil, data, funcs)
	case '[':
		node = NewArrayNode(nil, data, funcs)
	default:
		// For non-object/array root values, parse immediately
		p := newParser(data, funcs)
		return p.Parse()
	}
	
	return node, nil
}

// getFirstNonWhitespaceChar returns the first non-whitespace character in the data
func getFirstNonWhitespaceChar(data []byte) byte {
	for _, b := range data {
		switch b {
		case ' ', '\t', '\n', '\r':
			continue
		default:
			return b
		}
	}
	return 0
}

// Ensure all node types implement the Node interface.
var _ core.Node = (*objectNode)(nil)
var _ core.Node = (*arrayNode)(nil)
var _ core.Node = (*stringNode)(nil)
var _ core.Node = (*numberNode)(nil)
var _ core.Node = (*boolNode)(nil)
var _ core.Node = (*nullNode)(nil)
var _ core.Node = (*invalidNode)(nil)

func Traverse(n core.Node, path string) core.Node {
	// TODO: To be implemented based on the query module
	return newInvalidNode(fmt.Errorf("not implemented yet"))
}