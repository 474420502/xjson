package engine

import (
	"fmt"

	"github.com/474420502/xjson/internal/core"
)

func Parse(data []byte) (core.Node, error) {
	return ParseWithFuncs(data, nil)
}

func ParseWithFuncs(data []byte, funcs *map[string]core.UnaryPathFunc) (core.Node, error) {
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
