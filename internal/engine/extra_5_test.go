package engine

import (
	"testing"

	"github.com/474420502/xjson/internal/core"
)

func TestInvalidNodeForEach(t *testing.T) {
	// Test ForEach method on invalidNode which still has 0% coverage
	jsonData := []byte(`{"valid": "json"}`)
	root, err := Parse(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Create an invalid node by querying a non-existent path
	invalidNode := root.Query("/nonexistent")

	// Test ForEach on invalid node (should not panic and should not call the function)
	called := false
	invalidNode.ForEach(func(keyOrIndex interface{}, value core.Node) {
		called = true
	})

	if called {
		t.Error("Expected ForEach not to call the function on an invalid node")
	}
}