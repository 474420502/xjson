package engine

import (
	"testing"
	"time"

	"github.com/474420502/xjson/internal/core"
)

// TestInvalidNodeForEach tests the ForEach method of invalidNode
func TestInvalidNodeForEach2(t *testing.T) {
	invalidNode := &invalidNode{}

	// ForEach on invalidNode should not call the function
	called := false
	invalidNode.ForEach(func(keyOrIndex interface{}, value core.Node) {
		called = true
	})

	if called {
		t.Error("Expected ForEach not to be called on invalidNode")
	}
}

// TestTryParseInt tests the tryParseInt function
func TestTryParseInt(t *testing.T) {
	tests := []struct {
		input    string
		expected int
		ok       bool
	}{
		{"", 0, false},       // empty string
		{"-123", -123, true}, // negative number
		{"123", 123, true},   // positive number
		{"0", 0, true},       // zero
		{"-0", 0, true},      // negative zero
		{"abc", 0, false},    // non-numeric
		{"12a", 0, false},    // partially numeric
		{"-", 0, false},      // just minus sign
		{"12.3", 0, false},   // float number
	}

	for _, test := range tests {
		result, ok := tryParseInt(test.input)
		if ok != test.ok {
			t.Errorf("tryParseInt(%q) ok = %v, want %v", test.input, ok, test.ok)
		}
		if ok && result != test.expected {
			t.Errorf("tryParseInt(%q) = %d, want %d", test.input, result, test.expected)
		}
	}
}

// TestStringNodeTime tests the Time method of stringNode
func TestStringNodeTime(t *testing.T) {
	funcs := &map[string]core.UnaryPathFunc{}

	// Valid RFC3339 time string
	validTimeStr := "2023-01-01T00:00:00Z"
	node := NewStringNode(nil, validTimeStr, funcs)

	expectedTime, _ := time.Parse(time.RFC3339Nano, validTimeStr)
	actualTime := node.Time()

	if !actualTime.Equal(expectedTime) {
		t.Errorf("Time() = %v, want %v", actualTime, expectedTime)
	}

	// Invalid time string
	invalidTimeStr := "not-a-time"
	node = NewStringNode(nil, invalidTimeStr, funcs)

	timeVal := node.Time()
	if !timeVal.IsZero() {
		t.Errorf("Time() = %v, want zero time for invalid string", timeVal)
	}

	// Check that error was set
	if node.Error() == nil {
		t.Error("Expected error to be set for invalid time string")
	}
}

// TestArrayNodeAppend tests the Append method of arrayNode
func TestArrayNodeAppend(t *testing.T) {
	funcs := &map[string]core.UnaryPathFunc{}
	arr := NewArrayNode(nil, nil, funcs)

	// Test appending values
	values := []interface{}{"test", 42, true}
	for i, val := range values {
		result := arr.Append(val)
		if !result.IsValid() {
			t.Errorf("Append(%v) returned invalid node", val)
		}

		if arr.Len() != i+1 {
			t.Errorf("Len() = %d, want %d", arr.Len(), i+1)
		}
	}

	// Test appending to array with error
	testErr := &struct{ error }{}
	invalidArr := &arrayNode{baseNode: baseNode{err: testErr}}
	result := invalidArr.Append("value")
	if result.IsValid() {
		t.Error("Expected Append to return invalid node when array has error")
	}
}

// TestBaseNodeRemoveFunc tests the RemoveFunc method of baseNode
func TestBaseNodeRemoveFunc(t *testing.T) {
	// Test removing function from node with functions
	funcs := &map[string]core.UnaryPathFunc{
		"testFunc": func(node core.Node) core.Node { return node },
	}

	node := &baseNode{funcs: funcs}

	// Verify function exists
	if _, ok := (*node.funcs)["testFunc"]; !ok {
		t.Fatal("Expected testFunc to exist")
	}

	// Remove function
	result := node.RemoveFunc("testFunc")
	if !result.IsValid() {
		t.Error("RemoveFunc returned invalid node")
	}

	// Verify function was removed
	if _, ok := (*node.funcs)["testFunc"]; ok {
		t.Error("Expected testFunc to be removed")
	}

	// Test removing from node with error
	nodeWithErr := &baseNode{err: &testError2{"test error"}}
	result = nodeWithErr.RemoveFunc("nonexistent")
	if result.IsValid() {
		t.Error("Expected RemoveFunc to return invalid node when node has error")
	}

	// Test removing from node without functions
	emptyNode := &baseNode{}
	result = emptyNode.RemoveFunc("nonexistent")
	if !result.IsValid() {
		t.Error("RemoveFunc should work even when funcs is nil")
	}
}

// TestObjectNodeKeys tests the Keys method of objectNode
func TestObjectNodeKeys(t *testing.T) {
	// Test with valid object
	data := []byte(`{"c": "third", "a": "first", "b": "second"}`)
	obj, err := Parse(data)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	keys := obj.Keys()
	expectedKeys := []string{"a", "b", "c"} // Should be sorted

	if len(keys) != len(expectedKeys) {
		t.Errorf("Keys() returned %d keys, want %d", len(keys), len(expectedKeys))
	}

	for i, key := range expectedKeys {
		if keys[i] != key {
			t.Errorf("Keys()[%d] = %s, want %s", i, keys[i], key)
		}
	}

	// Test with object that has error
	testErr := &struct{ error }{}
	invalidObj := &objectNode{baseNode: baseNode{err: testErr}}
	keys = invalidObj.Keys()
	if keys != nil {
		t.Error("Expected Keys() to return nil when object has error")
	}
}

type testError2 struct {
	msg string
}

func (e *testError2) Error() string {
	return e.msg
}
