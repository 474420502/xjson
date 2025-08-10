package engine

import (
	"fmt"

	"github.com/474420502/xjson/internal/core"
)

// Parse is the entry point for the engine package. It creates a new parser
// and starts parsing the raw data.
func Parse(data []byte) (core.Node, error) {
	p := newParser(data)
	return p.Parse()
}

// NewNodeFromInterface creates a new node from a Go interface.
// This is useful for creating nodes programmatically in tests or applications.
func NewNodeFromInterface(v interface{}) (core.Node, error) {
	// TODO: This function will need to recursively build the node tree
	// based on the type of `v` (map, slice, string, etc.).
	// For now, return an invalid node.
	return newInvalidNode(fmt.Errorf("NewNodeFromInterface is not implemented")), nil
}
