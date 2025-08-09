package xjson

import (
	"testing"
)

func TestXJsonNewNodeFromInterface(t *testing.T) {
	node, err := NewNodeFromInterface(map[string]interface{}{"key": "value"})
	if err != nil {
		t.Fatalf("NewNodeFromInterface failed: %v", err)
	}
	if node.Type() != ObjectNode {
		t.Errorf("Expected ObjectNode, got %v", node.Type())
	}
	if node.Get("key").String() != "value" {
		t.Errorf("Expected object content to be correct, got %v", node.Get("key").String())
	}
}
