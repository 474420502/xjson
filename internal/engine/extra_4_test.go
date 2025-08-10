package engine

import (
	"testing"

	"github.com/474420502/xjson/internal/core"
)

func TestInvalidNodeParentMethod(t *testing.T) {
	// Test the Parent method of invalidNode which has 0% coverage
	jsonData := []byte(`{"valid": "json"}`)
	root, err := Parse(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Create an invalid node by querying a non-existent path
	invalidNode := root.Query("/nonexistent")

	// Test Parent method
	parent := invalidNode.Parent()
	if parent != nil && parent.IsValid() {
		// Parent of an invalid node could be valid (the root in this case)
		// but the invalid node itself should remain invalid
		t.Log("Parent of invalid node:", parent.Path())
	}
}

func TestInvalidNodeForEachMethod(t *testing.T) {
	// Test the ForEach method of invalidNode which has 0% coverage
	jsonData := []byte(`{"valid": "json"}`)
	root, err := Parse(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Create an invalid node by querying a non-existent path
	invalidNode := root.Query("/nonexistent")

	// Test ForEach on invalid node (should not panic or call the function)
	invalidNode.ForEach(func(keyOrIndex interface{}, value core.Node) {
		t.Error("ForEach should not be called on invalid node")
	})
}

func TestInvalidNodeMustMethods(t *testing.T) {
	// Test the Must* methods of invalidNode which have 0% coverage
	jsonData := []byte(`{"valid": "json"}`)
	root, err := Parse(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Create an invalid node by querying a non-existent path
	invalidNode := root.Query("/nonexistent")

	// Test MustString method panics
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected invalidNode.MustString() to panic")
			}
		}()
		invalidNode.MustString()
	}()

	// Test MustFloat method panics
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected invalidNode.MustFloat() to panic")
			}
		}()
		invalidNode.MustFloat()
	}()

	// Test MustInt method panics
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected invalidNode.MustInt() to panic")
			}
		}()
		invalidNode.MustInt()
	}()

	// Test MustBool method panics
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected invalidNode.MustBool() to panic")
			}
		}()
		invalidNode.MustBool()
	}()

	// Test MustTime method panics
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected invalidNode.MustTime() to panic")
			}
		}()
		invalidNode.MustTime()
	}()

	// Test MustArray method panics
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected invalidNode.MustArray() to panic")
			}
		}()
		invalidNode.MustArray()
	}()

	// Test MustAsMap method panics
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected invalidNode.MustAsMap() to panic")
			}
		}()
		invalidNode.MustAsMap()
	}()
}

func TestBaseNodeForEachMethod(t *testing.T) {
	// Test the ForEach method of baseNode which has 0% coverage
	var n core.Node = &baseNode{}

	// Test ForEach on baseNode (should call the function once with nil key and the node itself as value)
	var called bool
	var receivedKey interface{}
	var receivedValue core.Node

	n.ForEach(func(keyOrIndex interface{}, value core.Node) {
		called = true
		receivedKey = keyOrIndex
		receivedValue = value
	})

	if !called {
		t.Error("Expected ForEach to call the function")
	}

	if receivedKey != nil {
		t.Errorf("Expected key to be nil, got %v", receivedKey)
	}

	if receivedValue != n {
		t.Errorf("Expected value to be the node itself, got %v", receivedValue)
	}
}

func TestNewNodeFromInterface(t *testing.T) {
	// Test the NewNodeFromInterface function which has low coverage
	funcs := &map[string]core.UnaryPathFunc{}

	// Test with nil
	nilNode := NewNodeFromInterface(nil, nil, funcs)
	if nilNode.Type() != core.Null {
		t.Errorf("Expected nil to create a null node, got type %v", nilNode.Type())
	}

	// Test with string
	stringNode := NewNodeFromInterface(nil, "test", funcs)
	if stringNode.Type() != core.String {
		t.Errorf("Expected string to create a string node, got type %v", stringNode.Type())
	}
	if stringNode.String() != "test" {
		t.Errorf("Expected string node to have value \"test\", got %s", stringNode.String())
	}

	// Test with int
	intNode := NewNodeFromInterface(nil, 42, funcs)
	if intNode.Type() != core.Number {
		t.Errorf("Expected int to create a number node, got type %v", intNode.Type())
	}
	if intNode.Int() != 42 {
		t.Errorf("Expected number node to have value 42, got %d", intNode.Int())
	}

	// Test with float64
	floatNode := NewNodeFromInterface(nil, 3.14, funcs)
	if floatNode.Type() != core.Number {
		t.Errorf("Expected float64 to create a number node, got type %v", floatNode.Type())
	}
	if floatNode.Float() != 3.14 {
		t.Errorf("Expected number node to have value 3.14, got %f", floatNode.Float())
	}

	// Test with bool
	boolNode := NewNodeFromInterface(nil, true, funcs)
	if boolNode.Type() != core.Bool {
		t.Errorf("Expected bool to create a bool node, got type %v", boolNode.Type())
	}
	if !boolNode.Bool() {
		t.Errorf("Expected bool node to have value true, got %t", boolNode.Bool())
	}

	// Test with []interface{}
	arrayData := []interface{}{"a", "b", "c"}
	arrayNode := NewNodeFromInterface(nil, arrayData, funcs)
	if arrayNode.Type() != core.Array {
		t.Errorf("Expected []interface{} to create an array node, got type %v", arrayNode.Type())
	}
	if arrayNode.Len() != 3 {
		t.Errorf("Expected array node to have length 3, got %d", arrayNode.Len())
	}

	// Test with map[string]interface{}
	mapData := map[string]interface{}{"key1": "value1", "key2": "value2"}
	mapNode := NewNodeFromInterface(nil, mapData, funcs)
	if mapNode.Type() != core.Object {
		t.Errorf("Expected map[string]interface{} to create an object node, got type %v", mapNode.Type())
	}
	if mapNode.Len() != 2 {
		t.Errorf("Expected object node to have length 2, got %d", mapNode.Len())
	}
}

func TestUnescapeFunction(t *testing.T) {
	// Test the unescape function which has low coverage
	testCases := []struct {
		input    string
		expected string
	}{
		{`\"`, `"`},
		{`\\`, `\`},
		{`\/`, `/`},
		{`\b`, "\b"},
		{`\f`, "\f"},
		{`\n`, "\n"},
		{`\r`, "\r"},
		{`\t`, "\t"},
		{`\u0041`, "A"},     // Unicode escape
		{`\u0000`, "\u0000"}, // Null character
	}

	for _, tc := range testCases {
		result, err := unescape([]byte(tc.input))
		if err != nil {
			t.Errorf("Unexpected error for input %s: %v", tc.input, err)
		}
		if string(result) != tc.expected {
			t.Errorf("Expected unescape(%s) to be %s, got %s", tc.input, tc.expected, string(result))
		}
	}

	// Test error case with invalid unicode
	_, err := unescape([]byte(`\uGGGG`)) // Invalid hex digits
	if err == nil {
		t.Error("Expected error for invalid unicode escape sequence")
	}
}