package engine

import (
	"fmt"
	"testing"

	"github.com/474420502/xjson/internal/core"
)

func TestArrayNodeQuery(t *testing.T) {
	arrNode := NewArrayNode([]core.Node{
		NewNumberNode(float64(10), "", nil),
		NewStringNode("hello", "", nil),
	}, "root", nil)

	// Test valid query
	result := arrNode.Query("[0]")
	if result.Type() != core.NumberNode || result.Int() != 10 {
		t.Errorf("Expected number 10, but got %v", result.Interface())
	}

	// Test invalid query
	result = arrNode.Query("[unmatched")
	if result.Type() != core.InvalidNode {
		t.Errorf("Expected InvalidNode for invalid query, but got %v", result.Type())
	}

	// Test out of bounds
	result = arrNode.Query("[2]")
	if result.Type() != core.InvalidNode {
		t.Errorf("Expected InvalidNode for out-of-bounds query, but got %v", result.Type())
	}
}

func TestArrayNodeFunctionRegistrationAndCall(t *testing.T) {
	arrNode := NewArrayNode([]core.Node{NewNumberNode(1, "", nil)}, "root", nil)

	// Register a function that operates on the array node itself
	arrNode.(*arrayNode).RegisterFunc("double", func(n core.Node) core.Node {
		if n.Type() != core.ArrayNode {
			return NewInvalidNode(n.Path(), fmt.Errorf("double only applicable to arrays"))
		}
		var results []core.Node
		for _, child := range n.Array() {
			if child.Type() == core.NumberNode {
				results = append(results, NewNumberNode(float64(child.Int()*2), "", nil))
			} else {
				results = append(results, child) // Keep non-numbers as is
			}
		}
		return NewArrayNode(results, n.Path(), n.GetFuncs())
	})

	// Call the registered function.
	resultNode := arrNode.(*arrayNode).CallFunc("double")
	if resultNode.Type() != core.ArrayNode {
		t.Fatalf("Expected ArrayNode, got %v", resultNode.Type())
	}

	resultArr := resultNode.Array()
	if len(resultArr) != 1 || resultArr[0].Int() != 2 {
		t.Errorf("Expected [2], got %v", resultNode.Interface())
	}

	// Test calling a non-existent function
	invalidResult := arrNode.(*arrayNode).CallFunc("nonexistent")
	if invalidResult.Type() != core.InvalidNode {
		t.Errorf("Expected InvalidNode for non-existent function call, but got %v", invalidResult.Type())
	}
}

func TestArrayNodeApply(t *testing.T) {
	arrNode := NewArrayNode([]core.Node{
		NewNumberNode(1, "", nil),
		NewNumberNode(2, "", nil),
		NewNumberNode(3, "", nil),
	}, "root", nil)

	// Test with PredicateFunc (Filter)
	filterFunc := func(n core.Node) bool {
		return n.Int() > 1
	}
	filteredNode := arrNode.(*arrayNode).Apply(core.PredicateFunc(filterFunc))
	if filteredNode.Type() != core.ArrayNode {
		t.Fatalf("Filter should return ArrayNode, got %v", filteredNode.Type())
	}
	filteredArr := filteredNode.Array()
	if len(filteredArr) != 2 || filteredArr[0].Int() != 2 || filteredArr[1].Int() != 3 {
		t.Errorf("Filter failed. Expected nodes for [2, 3], got %v", filteredNode.Interface())
	}

	// Test with TransformFunc (Map)
	transformFunc := func(n core.Node) interface{} {
		return n.Int() * 2
	}
	mappedNode := arrNode.(*arrayNode).Apply(core.TransformFunc(transformFunc))
	if mappedNode.Type() != core.ArrayNode {
		t.Fatalf("Map should return ArrayNode, got %v", mappedNode.Type())
	}
	mappedArr := mappedNode.Array()
	if len(mappedArr) != 3 || mappedArr[0].Int() != 2 || mappedArr[1].Int() != 4 || mappedArr[2].Int() != 6 {
		t.Errorf("Map failed. Expected nodes for [2, 4, 6], got %v", mappedNode.Interface())
	}

	// Test with unsupported function type
	unsupportedFunc := func() {}
	invalidNode := arrNode.(*arrayNode).Apply(unsupportedFunc)
	if invalidNode.Type() != core.InvalidNode {
		t.Errorf("Expected InvalidNode for unsupported function type, but got %v", invalidNode.Type())
	}
}

// Deprecated: But we still need to test it.
func TestArrayNodeFunc(t *testing.T) {
	arrNode := NewArrayNode([]core.Node{NewNumberNode(1, "", nil)}, "root", nil)

	// Register a function using the deprecated method
	arrNode.(*arrayNode).Func("triple", func(n core.Node) core.Node {
		if n.Type() != core.ArrayNode {
			return NewInvalidNode(n.Path(), fmt.Errorf("triple only applicable to arrays"))
		}
		var results []core.Node
		for _, child := range n.Array() {
			if child.Type() == core.NumberNode {
				results = append(results, NewNumberNode(float64(child.Int()*3), "", nil))
			} else {
				results = append(results, child)
			}
		}
		return NewArrayNode(results, n.Path(), n.GetFuncs())
	})

	// Call the registered function
	resultNode := arrNode.(*arrayNode).CallFunc("triple")

	if resultNode.Type() != core.ArrayNode {
		t.Fatalf("Expected ArrayNode, got %v", resultNode.Type())
	}
	resultArr := resultNode.Array()
	if len(resultArr) != 1 || resultArr[0].Int() != 3 {
		t.Errorf("Expected [3], got %v", resultNode.Interface())
	}
}
