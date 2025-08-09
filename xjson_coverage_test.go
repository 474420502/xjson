package xjson

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestArrayNodeCoverageInXJSON(t *testing.T) {
	node, err := NewParser(`[1, "hello", true]`)
	assert.NoError(t, err)

	// Test type assertion panics and zero-value returns
	assert.Panics(t, func() { node.MustString() })
	assert.Equal(t, float64(0), node.Float())
	assert.Panics(t, func() { node.MustFloat() })
	assert.Equal(t, int64(0), node.Int())
	assert.Panics(t, func() { node.MustInt() })
	assert.False(t, node.Bool())
	assert.Panics(t, func() { node.MustBool() })
	assert.Equal(t, time.Time{}, node.Time())
	assert.Panics(t, func() { node.MustTime() })

	// Test Get on array, should be invalid
	invalidGet := node.Get("somekey")
	assert.Error(t, invalidGet.Error())
	assert.Equal(t, InvalidNode, invalidGet.Type())

	// Test RawFloat and RawString on array
	f, ok := node.RawFloat()
	assert.Equal(t, float64(0), f)
	assert.False(t, ok)

	s, ok := node.RawString()
	assert.Equal(t, "", s)
	assert.False(t, ok)

	// Test AsMap and MustAsMap on array
	assert.Nil(t, node.AsMap())
	assert.Panics(t, func() { node.MustAsMap() })

	// Test Func and Apply on array
	node.Func("test", func(n Node) Node { return n })
	node.Apply(func(n Node) bool { return true })

	// Test Query on array with valid and invalid paths
	qNode, err := NewParser(`[{"name":"test"}]`)
	assert.NoError(t, err)

	queried := qNode.Query("[0].name")
	assert.NoError(t, queried.Error())
	assert.Equal(t, "test", queried.String())

	invalidQueried := qNode.Query("[invalid query")
	assert.Error(t, invalidQueried.Error())
	assert.Equal(t, InvalidNode, invalidQueried.Type())
}

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
