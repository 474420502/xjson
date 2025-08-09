package engine

import (
	"errors"
	"fmt"
	"testing"

	"github.com/474420502/xjson/internal/core"
)

// This file contains a collection of small tests to cover remaining functions
// that were missed by other, more feature-focused tests.

func TestRemainingFuncAndApplyCoverage(t *testing.T) {
	// Setup a simple function to use
	dummyFunc := func(n core.Node) core.Node { return n }

	// Test cases for various node types
	testNodes := []core.Node{
		NewBoolNode(true, "bool", nil),
		NewNullNode("null", nil),
		NewStringNode("hello", "string", nil),
		NewObjectNode(map[string]core.Node{}, "object", nil),
		NewInvalidNode("invalid", errors.New("test error")),
		NewNumberNode(123, "number", nil),
	}

	for _, node := range testNodes {
		nodeName := fmt.Sprintf("%T", node)
		t.Run(nodeName+"_Func", func(t *testing.T) {
			// Get the concrete type to call Func, which is deprecated and not on the interface
			switch n := node.(type) {
			case *boolNode:
				n.Func("test", dummyFunc)
			case *nullNode:
				n.Func("test", dummyFunc)
			case *stringNode:
				n.Func("test", dummyFunc)
			case *objectNode:
				n.Func("test", dummyFunc)
			case *invalidNode:
				n.Func("test", dummyFunc)
			case *numberNode:
				n.Func("test", dummyFunc)
			}
		})

		t.Run(nodeName+"_Apply", func(t *testing.T) {
			// Apply is on the Node interface, no need for type switch
			node.Apply(core.PredicateFunc(func(n core.Node) bool { return true }))
		})
	}
}

// Custom type to trigger json.Marshal errors
type marshalerError struct{}

func (m marshalerError) MarshalJSON() ([]byte, error) {
	return nil, errors.New("marshal error")
}

// Custom node type whose Interface() method returns a type that fails to marshal
type jsonErrorNode struct {
	baseNode
}

func (n *jsonErrorNode) Interface() interface{} {
	return marshalerError{}
}
func (n *jsonErrorNode) Type() core.NodeType { return core.StringNode } // Lie about type to satisfy interface

func TestRawAndStringErrorPaths(t *testing.T) {
	// 1. Test n.raw != nil path for objectNode.Raw()
	parsedNode, err := ParseJSONToNode(`{"a":1}`)
	if err != nil {
		t.Fatal(err)
	}
	if parsedNode.Raw() != `{"a":1}` {
		t.Errorf("Expected raw string from parsed node, got %s", parsedNode.Raw())
	}

	// 2. Test json.Marshal error path for objectNode.Raw() and arrayNode.String()
	errorNode := &jsonErrorNode{}
	objWithErr := NewObjectNode(map[string]core.Node{"key": errorNode}, "test", nil)
	arrWithErr := NewArrayNode([]core.Node{errorNode}, "test", nil)

	if raw := objWithErr.Raw(); raw != "" {
		t.Errorf("Expected empty string from Raw() on marshal error, got %s", raw)
	}
	if objWithErr.Error() == nil {
		t.Error("Expected error to be set on object node after marshal error")
	}

	if str := arrWithErr.String(); str != "" {
		t.Errorf("Expected empty string from String() on marshal error, got %s", str)
	}
	if arrWithErr.Error() == nil {
		t.Error("Expected error to be set on array node after marshal error")
	}

	// 3. Test n.err != nil path
	objWithErr.Get("nonexistent") // This sets an error on the node
	if raw := objWithErr.Raw(); raw != "" {
		t.Errorf("Expected empty string from Raw() on a node with pre-existing error, got %s", raw)
	}
}
