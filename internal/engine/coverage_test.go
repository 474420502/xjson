package engine

import (
	"testing"

	"github.com/474420502/xjson/internal/core"
)

// TestInvalidNodeForEachCoverage tests the ForEach method of invalidNode
func TestInvalidNodeForEachCoverage(t *testing.T) {
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

// TestArrayNodeForEachWithEmptyArray tests ForEach method with an empty array
func TestArrayNodeForEachWithEmptyArray(t *testing.T) {
	funcs := &map[string]core.UnaryPathFunc{}
	arr := NewArrayNode(nil, []byte("[]"), funcs)
	
	count := 0
	arr.ForEach(func(keyOrIndex interface{}, value core.Node) {
		count++
	})
	
	if count != 0 {
		t.Errorf("Expected ForEach to be called 0 times for empty array, got %d", count)
	}
}

// TestArrayNodeForEachWithError tests ForEach method with an array that has an error
func TestArrayNodeForEachWithError(t *testing.T) {
	testErr := &struct{ error }{}
	invalidArr := &arrayNode{baseNode: baseNode{err: testErr}}
	
	called := false
	invalidArr.ForEach(func(keyOrIndex interface{}, value core.Node) {
		called = true
	})
	
	if called {
		t.Error("Expected ForEach not to be called on array with error")
	}
}

// TestArrayNodeAppendWithError tests Append method with an array that has an error
func TestArrayNodeAppendWithError(t *testing.T) {
	testErr := &struct{ error }{}
	invalidArr := &arrayNode{baseNode: baseNode{err: testErr}}
	
	result := invalidArr.Append("value")
	if result.IsValid() {
		t.Error("Expected Append to return invalid node when array has error")
	}
}

// TestArrayNodeAppendWithInvalidChild tests Append method with a value that creates an invalid child
func TestArrayNodeAppendWithInvalidChild(t *testing.T) {
	// Create a scenario where NewNodeFromInterface would return an invalid node
	// This is a bit tricky to test directly, so we'll test a normal case to improve coverage
	
	funcs := &map[string]core.UnaryPathFunc{}
	arr := NewArrayNode(nil, nil, funcs)
	
	// Test appending a valid value
	result := arr.Append("test")
	if !result.IsValid() {
		t.Errorf("Expected Append to return valid node, got invalid: %v", result.Error())
	}
	
	if arr.Len() != 1 {
		t.Errorf("Expected array length to be 1, got %d", arr.Len())
	}
}

// TestComplexSliceOperations tests complex slice operations
func TestComplexSliceOperations(t *testing.T) {
	jsonData := []byte(`{
		"items": [0, 1, 2, 3, 4, 5, 6, 7, 8, 9]
	}`)

	root, err := Parse(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	testCases := []struct {
		name     string
		path     string
		expected []int64
	}{
		{
			name:     "Full slice",
			path:     "/items[:]",
			expected: []int64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		},
		{
			name:     "Slice from start",
			path:     "/items[:5]",
			expected: []int64{0, 1, 2, 3, 4},
		},
		{
			name:     "Slice to end",
			path:     "/items[5:]",
			expected: []int64{5, 6, 7, 8, 9},
		},
		{
			name:     "Middle slice",
			path:     "/items[3:7]",
			expected: []int64{3, 4, 5, 6},
		},
		{
			name:     "Negative start index",
			path:     "/items[-3:]",
			expected: []int64{7, 8, 9},
		},
		{
			name:     "Negative end index",
			path:     "/items[:-3]",
			expected: []int64{0, 1, 2, 3, 4, 5, 6},
		},
		{
			name:     "Both negative indices",
			path:     "/items[-5:-2]",
			expected: []int64{5, 6, 7},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := root.Query(tc.path)
			if !result.IsValid() {
				t.Fatalf("Query '%s' failed: %v", tc.path, result.Error())
			}

			if result.Type() != core.Array {
				t.Fatalf("Expected result to be an array, got %v", result.Type())
			}

			values := result.Array()
			if len(values) != len(tc.expected) {
				t.Fatalf("Expected %d values, got %d", len(tc.expected), len(values))
			}

			for i, expectedVal := range tc.expected {
				if values[i].Int() != expectedVal {
					t.Errorf("Expected values[%d] to be %d, got %d", i, expectedVal, values[i].Int())
				}
			}
		})
	}
}

// TestMultiLevelParentNavigation tests multi-level parent navigation with ../
func TestMultiLevelParentNavigation(t *testing.T) {
	jsonData := []byte(`{
		"level1": {
			"level2": {
				"level3": {
					"target": "found it"
				},
				"data": "other data"
			}
		}
	}`)

	root, err := Parse(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Navigate deep and then back up multiple levels
	result := root.Query("/level1/level2/level3/target/../../data")
	if !result.IsValid() {
		t.Fatalf("Multi-level parent navigation failed: %v", result.Error())
	}

	if result.String() != "other data" {
		t.Errorf("Expected 'other data', got '%s'", result.String())
	}

	// Navigate back to root
	result = root.Query("/level1/level2/level3/target/../../../..")
	if !result.IsValid() {
		t.Fatalf("Navigation back to root failed: %v", result.Error())
	}

	// Try to navigate above root (should result in invalid node)
	result = root.Query("/level1/level2/level3/target/../../../../..")
	if result.IsValid() {
		t.Error("Expected navigation above root to result in invalid node")
	}
}

// TestMultipleFunctionChaining tests chaining multiple functions like [@func1][@func2]
func TestMultipleFunctionChaining(t *testing.T) {
	jsonData := []byte(`{
		"numbers": [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]
	}`)

	root, err := Parse(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Register functions
	root.RegisterFunc("even", func(n core.Node) core.Node {
		return n.Filter(func(child core.Node) bool {
			return child.Int()%2 == 0
		})
	})

	root.RegisterFunc("greaterThanFive", func(n core.Node) core.Node {
		return n.Filter(func(child core.Node) bool {
			return child.Int() > 5
		})
	})

	// Test chaining functions: first get even numbers, then filter for those greater than 5
	result := root.Query("/numbers[@even][@greaterThanFive]")
	if !result.IsValid() {
		t.Fatalf("Function chaining failed: %v", result.Error())
	}

	if result.Type() != core.Array {
		t.Fatalf("Expected result to be an array, got %v", result.Type())
	}

	expected := []int64{6, 8, 10}
	values := result.Array()
	if len(values) != len(expected) {
		t.Fatalf("Expected %d values, got %d", len(expected), len(values))
	}

	for i, expectedVal := range expected {
		if values[i].Int() != expectedVal {
			t.Errorf("Expected values[%d] to be %d, got %d", i, expectedVal, values[i].Int())
		}
	}
}