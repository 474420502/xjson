package engine

import (
	"testing"

	"github.com/474420502/xjson/internal/core"
)

func TestBaseNodeDefaultMethods(t *testing.T) {
	// Test default implementations of baseNode methods that should be overridden
	var n core.Node = &baseNode{}

	// Test default Len method
	if n.Len() != 1 {
		t.Errorf("Expected baseNode.Len() to return 1, got %d", n.Len())
	}

	// Test default Get method
	getResult := n.Get("key")
	if getResult.IsValid() {
		t.Error("Expected baseNode.Get() to return invalid node")
	}

	// Test default Index method
	indexResult := n.Index(0)
	if indexResult.IsValid() {
		t.Error("Expected baseNode.Index() to return invalid node")
	}

	// Test default Set method
	setResult := n.Set("key", "value")
	if setResult.IsValid() {
		t.Error("Expected baseNode.Set() to return invalid node")
	}

	// Test default Append method
	appendResult := n.Append("value")
	if appendResult.IsValid() {
		t.Error("Expected baseNode.Append() to return invalid node")
	}

	// Test default Filter method
	filterResult := n.Filter(func(node core.Node) bool { return true })
	if filterResult.IsValid() {
		t.Error("Expected baseNode.Filter() to return invalid node")
	}

	// Test default Map method
	mapResult := n.Map(func(node core.Node) interface{} { return nil })
	if mapResult.IsValid() {
		t.Error("Expected baseNode.Map() to return invalid node")
	}

	// Test default SetValue method
	setValueResult := n.SetValue("value")
	if setValueResult.IsValid() {
		t.Error("Expected baseNode.SetValue() to return invalid node")
	}

	// Test default Apply method
	applyResult := n.Apply(func(n core.Node) core.Node { return n })
	if applyResult.IsValid() {
		t.Error("Expected baseNode.Apply() to return invalid node")
	}

	// Test default type conversion methods
	if n.String() != "" {
		t.Errorf("Expected baseNode.String() to return empty string, got %s", n.String())
	}

	// Test Must* methods panic as expected
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected baseNode.MustString() to panic")
			}
		}()
		n.MustString()
	}()

	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected baseNode.MustFloat() to panic")
			}
		}()
		n.MustFloat()
	}()

	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected baseNode.MustInt() to panic")
			}
		}()
		n.MustInt()
	}()

	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected baseNode.MustBool() to panic")
			}
		}()
		n.MustBool()
	}()

	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected baseNode.MustTime() to panic")
			}
		}()
		n.MustTime()
	}()

	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected baseNode.MustArray() to panic")
			}
		}()
		n.MustArray()
	}()

	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected baseNode.MustAsMap() to panic")
			}
		}()
		n.MustAsMap()
	}()

	// Test default value methods
	if n.Float() != 0 {
		t.Errorf("Expected baseNode.Float() to return 0, got %f", n.Float())
	}

	if n.Int() != 0 {
		t.Errorf("Expected baseNode.Int() to return 0, got %d", n.Int())
	}

	if n.Bool() != false {
		t.Errorf("Expected baseNode.Bool() to return false, got %t", n.Bool())
	}

	if !n.Time().IsZero() {
		t.Errorf("Expected baseNode.Time() to return zero time, got %v", n.Time())
	}

	if n.Array() != nil {
		t.Error("Expected baseNode.Array() to return nil")
	}

	if n.Interface() != nil {
		t.Error("Expected baseNode.Interface() to return nil")
	}

	// Test RawFloat and RawString default implementations
	if f, ok := n.RawFloat(); ok && f != 0 {
		t.Errorf("Expected baseNode.RawFloat() to return (0, false), got (%f, %t)", f, ok)
	}

	if s, ok := n.RawString(); !ok || s != "" {
		t.Errorf("Expected baseNode.RawString() to return (\"\", true), got (%s, %t)", s, ok)
	}

	// Test Strings default implementation
	strings := n.Strings()
	if len(strings) != 1 || strings[0] != "" {
		t.Errorf("Expected baseNode.Strings() to return [\"\"], got %v", strings)
	}

	if n.Keys() != nil {
		t.Error("Expected baseNode.Keys() to return nil")
	}

	if !n.Contains("") {
		t.Error("Expected baseNode.Contains(\"\") to return true")
	}

	if n.AsMap() != nil {
		t.Error("Expected baseNode.AsMap() to return nil")
	}
}


func TestSimpleNodeStringMethods(t *testing.T) {
	// Test boolNode and nullNode String methods which have 0% coverage
	funcs := &map[string]core.UnaryPathFunc{}
	trueNode := NewBoolNode(nil, true, funcs)
	falseNode := NewBoolNode(nil, false, funcs)
	nullNode := NewNullNode(nil, funcs)

	if trueNode.String() != "true" {
		t.Errorf("Expected trueNode.String() to return \"true\", got %s", trueNode.String())
	}

	if falseNode.String() != "false" {
		t.Errorf("Expected falseNode.String() to return \"false\", got %s", falseNode.String())
	}

	if nullNode.String() != "null" {
		t.Errorf("Expected nullNode.String() to return \"null\", got %s", nullNode.String())
	}
}

func TestBaseNodeSetError(t *testing.T) {
	// Test the setError method which has 0% coverage
	jsonData := []byte(`{"test": "value"}`)
	root, err := Parse(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Get the base node
	node, ok := root.(*objectNode)
	if !ok {
		t.Fatal("Expected root to be an objectNode")
	}

	// Test setting error when none exists
	testErr := &testError{"test error"}
	node.setError(testErr)

	// The error should be set
	if node.Error() != testErr {
		t.Errorf("Expected node.Error() to return testErr, got %v", node.Error())
	}

	// Test that setting another error doesn't overwrite the first one
	anotherErr := &testError{"another error"}
	node.setError(anotherErr)

	// The error should still be the first one
	if node.Error() != testErr {
		t.Errorf("Expected node.Error() to still return testErr, got %v", node.Error())
	}
}

func TestTraverseFunction(t *testing.T) {
	// Test the Traverse function which has 0% coverage
	jsonData := []byte(`{"test": "value"}`)
	root, err := Parse(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Currently Traverse is not implemented and should return an invalid node
	result := Traverse(root, "/test")
	if result.IsValid() {
		t.Error("Expected Traverse to return invalid node as it's not implemented")
	}
}

func TestObjectAndArrayAddChild(t *testing.T) {
	// Test the addChild methods which have 0% coverage
	// Create object and array nodes
	funcs := &map[string]core.UnaryPathFunc{}
	obj := NewObjectNode(nil, nil, funcs)
	arr := NewArrayNode(nil, nil, funcs)

	// Test object addChild
	stringNode := NewStringNode(nil, "test value", funcs)
	obj.(*objectNode).addChild("key", stringNode)

	// Verify the child was added
	if obj.Len() != 1 {
		t.Errorf("Expected object length to be 1, got %d", obj.Len())
	}

	value := obj.Get("key")
	if !value.IsValid() {
		t.Error("Expected to find added key in object")
	}

	if value.String() != "test value" {
		t.Errorf("Expected value to be \"test value\", got %s", value.String())
	}

	// Test array addChild
	numberNode := NewNumberNode(nil, []byte("42"), funcs)
	arr.(*arrayNode).addChild(numberNode)

	// Verify the child was added
	if arr.Len() != 1 {
		t.Errorf("Expected array length to be 1, got %d", arr.Len())
	}

	element := arr.Index(0)
	if !element.IsValid() {
		t.Error("Expected to find added element in array")
	}

	if element.Int() != 42 {
		t.Errorf("Expected element to be 42, got %d", element.Int())
	}
}

func TestObjectAndArrayLazyParseErrorHandling(t *testing.T) {
	// Test error handling in lazyParse methods
	// We'll create nodes with invalid raw data to trigger parsing errors

	// Create an object node with invalid raw data
	funcs := &map[string]core.UnaryPathFunc{}
	invalidObjRaw := []byte(`{invalid json}`)
	obj := &objectNode{
		baseNode: baseNode{
			raw:   invalidObjRaw,
			funcs: funcs,
		},
	}

	// Trigger lazyParse by calling a method that requires parsing
	obj.Len()

	// Should have an error
	if obj.Error() == nil {
		t.Error("Expected object with invalid raw data to have an error after parsing")
	}

	// Create an array node with invalid raw data
	invalidArrRaw := []byte(`[invalid json]`)
	arr := &arrayNode{
		baseNode: baseNode{
			raw:   invalidArrRaw,
			funcs: funcs,
		},
	}

	// Trigger lazyParse by calling a method that requires parsing
	arr.Len()

	// Should have an error
	if arr.Error() == nil {
		t.Error("Expected array with invalid raw data to have an error after parsing")
	}
}

// Helper type for testing setError
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
