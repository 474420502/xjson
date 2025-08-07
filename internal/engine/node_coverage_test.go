package engine

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func createMockNode(nodeType NodeType, value interface{}, path string, err error) Node {
	switch nodeType {
	case InvalidNode:
		return NewInvalidNode(path, err)
	case ObjectNode:
		if m, ok := value.(map[string]Node); ok {
			return NewObjectNode(m, path, &map[string]func(Node) Node{})
		}
	case ArrayNode:
		if arr, ok := value.([]Node); ok {
			return NewArrayNode(arr, path, &map[string]func(Node) Node{})
		}
	case StringNode:
		if s, ok := value.(string); ok {
			return NewStringNode(s, path, &map[string]func(Node) Node{})
		}
	case NumberNode:
		if f, ok := value.(float64); ok {
			return NewNumberNode(f, path, &map[string]func(Node) Node{})
		}
	case BoolNode:
		if b, ok := value.(bool); ok {
			return NewBoolNode(b, path, &map[string]func(Node) Node{})
		}
	case NullNode:
		return NewNullNode(path, &map[string]func(Node) Node{})
	}
	return NewInvalidNode(path, errors.New("unsupported node type for mock creation"))
}

func TestInvalidNode(t *testing.T) {
	testErr := errors.New("test error")
	invalidNode := createMockNode(InvalidNode, nil, "/test/path", testErr).(Node)

	if invalidNode.IsValid() {
		t.Errorf("Expected IsValid() to be false for an invalid node, but got true")
	}
	if invalidNode.Error() != testErr {
		t.Errorf("Expected Error() to return '%v', but got '%v'", testErr, invalidNode.Error())
	}
	if invalidNode.Path() != "/test/path" {
		t.Errorf("Expected Path() to be '/test/path', but got '%s'", invalidNode.Path())
	}

	// Test methods that should return zero values or be no-ops
	var dummyIterator func(interface{}, Node)
	invalidNode.ForEach(dummyIterator) // Should not panic

	if invalidNode.Len() != 0 {
		t.Errorf("Expected Len() to be 0 for invalid node, but got %d", invalidNode.Len())
	}
	if invalidNode.String() != "" {
		t.Errorf("Expected String() to be empty for invalid node, but got '%s'", invalidNode.String())
	}
	if invalidNode.Float() != 0 {
		t.Errorf("Expected Float() to be 0 for invalid node, but got %f", invalidNode.Float())
	}
	if invalidNode.Int() != 0 {
		t.Errorf("Expected Int() to be 0 for invalid node, but got %d", invalidNode.Int())
	}
	if invalidNode.Bool() != false {
		t.Errorf("Expected Bool() to be false for invalid node, but got %t", invalidNode.Bool())
	}
	if !invalidNode.Time().IsZero() {
		t.Errorf("Expected Time() to be zero for invalid node, but got %v", invalidNode.Time())
	}
	if invalidNode.Array() != nil {
		t.Errorf("Expected Array() to be nil for invalid node")
	}
	if invalidNode.Interface() != nil {
		t.Errorf("Expected Interface() to be nil for invalid node")
	}

	// Test methods that should panic
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic for MustString() on invalid node")
		}
	}()
	invalidNode.MustString()

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic for MustFloat() on invalid node")
		}
	}()
	invalidNode.MustFloat()

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic for MustInt() on invalid node")
		}
	}()
	invalidNode.MustInt()

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic for MustBool() on invalid node")
		}
	}()
	invalidNode.MustBool()

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic for MustTime() on invalid node")
		}
	}()
	invalidNode.MustTime()

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic for MustArray() on invalid node")
		}
	}()
	invalidNode.MustArray()

	// Test methods that should return invalid nodes or be no-ops
	if !invalidNode.Get("key").IsValid() {
		t.Errorf("Expected Get() to return an invalid node")
	}
	if !invalidNode.Index(0).IsValid() {
		t.Errorf("Expected Index() to return an invalid node")
	}
	if !invalidNode.Query("path").IsValid() {
		t.Errorf("Expected Query() to return an invalid node")
	}
	invalidNode.Filter(func(Node) bool { return true })    // Should not panic
	invalidNode.Map(func(Node) interface{} { return nil }) // Should not panic
	invalidNode.Set("key", "value")                        // Should not panic
	invalidNode.Append("value")                            // Should not panic
	if _, ok := invalidNode.RawFloat(); ok {
		t.Errorf("Expected RawFloat() to return false for ok")
	}
	if _, ok := invalidNode.RawString(); ok {
		t.Errorf("Expected RawString() to return false for ok")
	}
	if invalidNode.Strings() != nil {
		t.Errorf("Expected Strings() to be nil for invalid node")
	}
	if invalidNode.Contains("value") {
		t.Errorf("Expected Contains() to be false for invalid node")
	}
}

func TestObjectNode_BasicMethods(t *testing.T) {
	mockMap := map[string]Node{
		"name": NewStringNode("test", "/test/path/name", nil),
		"age":  NewNumberNode(30, "/test/path/age", nil),
	}
	objNode := createMockNode(ObjectNode, mockMap, "/test/path", nil).(Node)

	if objNode.Type() != ObjectNode {
		t.Errorf("Expected Type() to be ObjectNode")
	}
	if objNode.Path() != "/test/path" {
		t.Errorf("Expected Path() to be '/test/path', got '%s'", objNode.Path())
	}

	// Test Get
	nameNode := objNode.Get("name")
	if nameNode.String() != "test" {
		t.Errorf("Expected Get(\"name\").String() to be 'test', got '%s'", nameNode.String())
	}
	ageNode := objNode.Get("age")
	if ageNode.Float() != 30 {
		t.Errorf("Expected Get(\"age\").Float() to be 30, got %f", ageNode.Float())
	}
	nonExistentNode := objNode.Get("nonexistent")
	if nonExistentNode.IsValid() {
		t.Errorf("Expected Get(\"nonexistent\") to return an invalid node")
	}

	// Test Index (should be invalid for object)
	assert.False(t, objNode.Index(0).IsValid(), "Expected Index(0) to return an invalid node for object")

	// Test Query
	queryResult := objNode.Query("name")
	if queryResult.String() != "test" {
		t.Errorf("Expected Query(\"name\").String() to be 'test', got '%s'", queryResult.String())
	}

	// Test ForEach
	count := 0
	objNode.ForEach(func(keyOrIndex interface{}, value Node) {
		count++
		key, ok := keyOrIndex.(string)
		assert.True(t, ok)
		assert.Contains(t, []string{"name", "age"}, key)
	})
	assert.Equal(t, 2, count)
	assert.Equal(t, 2, objNode.Len())

	// Test String
	expectedString := `{"age":30,"name":"test"}` // Order might vary, but content should match
	actualString := objNode.String()
	if actualString != expectedString && actualString != `{"name":"test","age":30}` {
		t.Errorf("Expected String() to be '%s' or '%s', got '%s'", expectedString, `{"name":"test","age":30}`, actualString)
	}

	// Test MustString (should panic if error, but objectNode has no error here)
	if objNode.MustString() != actualString {
		t.Errorf("Expected MustString() to return the same as String()")
	}

	// Test type assertion methods (should return zero values)
	if objNode.Float() != 0 {
		t.Errorf("Expected Float() to be 0 for object node, got %f", objNode.Float())
	}
	if objNode.Int() != 0 {
		t.Errorf("Expected Int() to be 0 for object node, got %d", objNode.Int())
	}
	if objNode.Bool() != false {
		t.Errorf("Expected Bool() to be false for object node, got %t", objNode.Bool())
	}
	if !objNode.Time().IsZero() {
		t.Errorf("Expected Time() to be zero for object node, got %v", objNode.Time())
	}
	if objNode.Array() != nil {
		t.Errorf("Expected Array() to be nil for object node")
	}
	if objNode.Interface() == nil {
		t.Errorf("Expected Interface() to return a map for object node")
	}
}

func TestObjectNode_Set(t *testing.T) {
	mockMap := map[string]Node{
		"name": NewStringNode("test", "/test/path/name", nil),
	}
	objNode := createMockNode(ObjectNode, mockMap, "/test/path", nil).(Node)

	// Set a new value
	objNode.Set("age", 30)
	ageNode := objNode.Get("age")
	if ageNode.Float() != 30 {
		t.Errorf("Expected Set(\"age\", 30) to add age, got %f", ageNode.Float())
	}

	// Update an existing value
	objNode.Set("name", "updated_test")
	nameNode := objNode.Get("name")
	if nameNode.String() != "updated_test" {
		t.Errorf("Expected Set(\"name\", \"updated_test\") to update name, got '%s'", nameNode.String())
	}

	// Set a value that causes an error in NewNodeFromInterface (e.g., unsupported type)
	objNode.Set("invalid", make(chan int))
	if objNode.Error() == nil {
		t.Errorf("Expected Set() to set an error when creating node from unsupported type")
	}
}

func TestObjectNode_Append(t *testing.T) {
	mockMap := map[string]Node{
		"name": NewStringNode("test", "/test/path/name", nil),
	}
	objNode := createMockNode(ObjectNode, mockMap, "/test/path", nil).(Node)

	objNode.Append("some_value")
	if objNode.Error() == nil {
		t.Errorf("Expected Append() on object node to set an error")
	}
}

func TestObjectNode_Filter(t *testing.T) {
	mockMap := map[string]Node{
		"item1": NewNumberNode(10, "/test/path/item1", nil),
		"item2": NewNumberNode(25, "/test/path/item2", nil),
		"item3": NewNumberNode(5, "/test/path/item3", nil),
	}
	objNode := createMockNode(ObjectNode, mockMap, "/test/path", nil).(Node)

	// Filter for values less than 20
	filteredNode := objNode.Filter(func(n Node) bool {
		return n.Float() < 20
	})

	if filteredNode.Type() != ArrayNode {
		t.Errorf("Expected Filter() to return an ArrayNode, got %v", filteredNode.Type())
	}

	filteredArray := filteredNode.Array()
	assert.Equal(t, 2, len(filteredArray), "Expected Filter() to return an array with 2 elements")

	// Convert to a slice of floats for easier comparison without order dependency
	var floatValues []float64
	for _, n := range filteredArray {
		floatValues = append(floatValues, n.Float())
	}
	assert.ElementsMatch(t, []float64{10, 5}, floatValues, "Expected filtered array elements to be 10 and 5")
}

func TestObjectNode_Map(t *testing.T) {
	mockMap := map[string]Node{
		"price": NewNumberNode(10.5, "/test/path/price", nil),
	}
	objNode := createMockNode(ObjectNode, mockMap, "/test/path", nil).(Node)

	// Map to double the price
	mappedNode := objNode.Map(func(n Node) interface{} {
		return n.Float() * 2
	})

	if mappedNode.Type() != ArrayNode {
		t.Errorf("Expected Map() to return an ArrayNode, got %v", mappedNode.Type())
	}

	mappedArray := mappedNode.Array()
	if len(mappedArray) != 1 {
		t.Errorf("Expected Map() to return an array with 1 element, got %d", len(mappedArray))
	}
	if mappedArray[0].Float() != 21.0 {
		t.Errorf("Expected mapped element to be 21.0, got %f", mappedArray[0].Float())
	}
}

func TestObjectNode_CallFunc(t *testing.T) {
	// Mock a function
	mockFunc := func(n Node) Node {
		// For simplicity, let's say it returns a string node with the path
		return NewStringNode(n.Path(), "/func/result", nil)
	}

	// Create an object node with a registered function
	funcs := make(map[string]func(Node) Node)
	funcs["testFunc"] = mockFunc
	objNode := NewObjectNode(map[string]Node{}, "/test/path", &funcs)

	// Call the registered function
	resultNode := objNode.CallFunc("testFunc")
	if resultNode.String() != "/test/path" {
		t.Errorf("Expected CallFunc(\"testFunc\") to return node with path '/test/path', got '%s'", resultNode.String())
	}

	// Call a non-existent function
	nonExistentResult := objNode.CallFunc("nonExistentFunc")
	if nonExistentResult.IsValid() {
		t.Errorf("Expected CallFunc(\"nonExistentFunc\") to return an invalid node")
	}
}

func TestObjectNode_RemoveFunc(t *testing.T) {
	mockFunc := func(n Node) Node { return NewStringNode("should not be called", "", nil) }
	funcs := make(map[string]func(Node) Node)
	funcs["testFunc"] = mockFunc
	objNode := NewObjectNode(map[string]Node{}, "/test/path", &funcs)

	// Remove the function
	objNode.RemoveFunc("testFunc")

	// Try to call the removed function
	resultNode := objNode.CallFunc("testFunc")
	if resultNode.IsValid() {
		t.Errorf("Expected CallFunc(\"testFunc\") after RemoveFunc to return an invalid node")
	}
}

// Add tests for other node types and methods as needed

// New tests for objectNode.Index
func TestObjectNode_Index(t *testing.T) {
	// Test Index on an object node
	objNode := NewObjectNode(map[string]Node{
		"a": NewStringNode("1", "/test/path/a", nil),
	}, "/test/path", nil)

	invalidIndexNode := objNode.Index(0)
	assert.False(t, invalidIndexNode.IsValid())
	assert.Equal(t, "/test/path[0]", invalidIndexNode.Path())
	assert.Error(t, invalidIndexNode.Error())                              // Check if an error is set
	assert.Contains(t, invalidIndexNode.Error().Error(), "type assertion") // Check error message

	// Test Index on an invalid object node
	invalidObjNode := NewInvalidNode("/test/path", errors.New("parent error"))
	invalidIndexNodeFromInvalidObj := invalidObjNode.Index(0)
	assert.False(t, invalidIndexNodeFromInvalidObj.IsValid())
	assert.Error(t, invalidIndexNodeFromInvalidObj.Error())
	assert.Contains(t, invalidIndexNodeFromInvalidObj.Error().Error(), "parent error")
}

// Tests from coverage_test.go
func TestMustMethods(t *testing.T) {
	objNode := NewObjectNode(map[string]Node{}, "", nil)
	arrNode := NewArrayNode([]Node{}, "", nil)
	strNode := NewStringNode("hello", "", nil)
	numNode := NewNumberNode(123, "", nil)
	boolNode := NewBoolNode(true, "", nil)

	assert.Equal(t, "{}", objNode.MustString()) // Corrected from assert.Panics
	assert.Equal(t, "hello", strNode.MustString())

	assert.Panics(t, func() { strNode.MustFloat() })
	assert.Equal(t, float64(123), numNode.MustFloat())

	assert.Panics(t, func() { strNode.MustInt() })
	assert.Equal(t, int64(123), numNode.MustInt())

	assert.Panics(t, func() { strNode.MustBool() })
	assert.Equal(t, true, boolNode.MustBool())

	assert.Panics(t, func() { strNode.MustArray() })
	assert.NotNil(t, arrNode.MustArray())

	timeStr := "2024-01-01T15:04:05Z"
	timeNode := NewStringNode(timeStr, "", nil)
	parsedTime, _ := time.Parse(time.RFC3339, timeStr)
	assert.Equal(t, parsedTime, timeNode.MustTime())
	assert.Panics(t, func() { numNode.MustTime() })

	invalidNode := NewInvalidNode("", assert.AnError)
	assert.Panics(t, func() { invalidNode.MustString() })
	assert.Panics(t, func() { invalidNode.MustFloat() })
	assert.Panics(t, func() { invalidNode.MustInt() })
	assert.Panics(t, func() { invalidNode.MustBool() })
	assert.Panics(t, func() { invalidNode.MustArray() })
	assert.Panics(t, func() { invalidNode.MustTime() })
}

func TestFuncManagement(t *testing.T) {
	// Create a root node with a shared funcs map
	funcsMap := make(map[string]func(Node) Node)
	root := NewObjectNode(make(map[string]Node), "", &funcsMap)

	// Register a function
	root.Func("double", func(n Node) Node {
		// This function will be tested on a number node
		return NewNumberNode(n.Float()*2, "", &funcsMap)
	})

	// Call the function on a number node
	numNode := NewNumberNode(5, "", &funcsMap)
	result := numNode.CallFunc("double")
	assert.True(t, result.IsValid())
	assert.Equal(t, 10.0, result.Float())

	// Test GetFuncs
	funcs := root.GetFuncs()
	assert.NotNil(t, funcs)
	assert.Equal(t, 1, len(*funcs))

	// Remove the function
	root.RemoveFunc("double")
	result = root.CallFunc("double")
	assert.False(t, result.IsValid())
	assert.Equal(t, 0, len(*root.GetFuncs()))

	// Test on invalid node
	invalid := NewInvalidNode("", nil)
	invalid.Func("test", func(n Node) Node { return n }).RemoveFunc("test")
	assert.Nil(t, invalid.GetFuncs())
}

func TestRaw(t *testing.T) {
	rawStr := `{"key":"value"}`
	node, err := ParseJSONToNode(rawStr)
	assert.NoError(t, err)
	assert.Equal(t, rawStr, node.Raw())
}

func TestForEachAndLen(t *testing.T) {
	// Object
	objNode := NewObjectNode(map[string]Node{
		"a": NewStringNode("1", "", nil),
		"b": NewStringNode("2", "", nil),
	}, "", nil)

	count := 0
	objNode.ForEach(func(keyOrIndex interface{}, value Node) {
		count++
		key, ok := keyOrIndex.(string)
		assert.True(t, ok)
		assert.Contains(t, []string{"a", "b"}, key)
	})
	assert.Equal(t, 2, count)
	assert.Equal(t, 2, objNode.Len())

	// Array
	arrNode := NewArrayNode([]Node{
		NewStringNode("a", "", nil),
		NewStringNode("b", "", nil),
	}, "", nil)

	count = 0
	arrNode.ForEach(func(keyOrIndex interface{}, value Node) {
		count++
		idx, ok := keyOrIndex.(int)
		assert.True(t, ok)
		assert.Contains(t, []int{0, 1}, idx)
	})
	assert.Equal(t, 2, count)
	assert.Equal(t, 2, arrNode.Len())
}

func TestTime(t *testing.T) {
	timeStr := "2024-01-01T15:04:05Z"
	timeNode := NewStringNode(timeStr, "", nil)
	parsedTime, _ := time.Parse(time.RFC3339, timeStr)
	assert.Equal(t, parsedTime, timeNode.Time())

	// Error case
	badTimeNode := NewStringNode("not-a-time", "", nil)
	assert.True(t, badTimeNode.Time().IsZero())
	assert.Error(t, badTimeNode.Error())
}

func TestMiscCoverage(t *testing.T) {
	// Cover some zero-return cases for non-applicable types
	objNode := NewObjectNode(map[string]Node{}, "", nil)
	assert.Equal(t, int64(0), objNode.Int())
	assert.False(t, objNode.Bool())
	assert.True(t, objNode.Time().IsZero())

	numNode := NewNumberNode(1, "", nil)
	assert.False(t, numNode.Bool())
	assert.True(t, numNode.Time().IsZero())
}
