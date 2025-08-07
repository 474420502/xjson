package engine

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorChaining(t *testing.T) {
	// Create an invalid node
	invalid := NewInvalidNode("root", errors.New("initial error"))

	// Chain multiple operations
	result := invalid.Get("key1").Index(0).Query("path")

	// Verify that the result is still invalid
	assert.False(t, result.IsValid())

	// Verify that the original error is preserved
	assert.EqualError(t, result.Error(), "initial error")
}

func TestGet(t *testing.T) {
	funcs := make(map[string]func(Node) Node)
	obj := NewObjectNode(
		map[string]Node{
			"key1": NewStringNode("value1", ".key1", &funcs),
		},
		"",
		&funcs,
	)

	// Test successful Get
	child := obj.Get("key1")
	assert.True(t, child.IsValid())
	assert.Equal(t, "value1", child.String())

	// Test failed Get
	child = obj.Get("nonexistent")
	assert.False(t, child.IsValid())
	assert.Equal(t, ErrNotFound, child.Error())
}

func TestIndex(t *testing.T) {
	funcs := make(map[string]func(Node) Node)
	arr := NewArrayNode(
		[]Node{
			NewStringNode("value0", "[0]", &funcs),
		},
		"",
		&funcs,
	)

	// Test successful Index
	child := arr.Index(0)
	assert.True(t, child.IsValid())
	assert.Equal(t, "value0", child.String())

	// Test failed Index (out of bounds)
	child = arr.Index(1)
	assert.False(t, child.IsValid())
	assert.Equal(t, ErrIndexOutOfBounds, child.Error())
}

func TestInvalidOperations(t *testing.T) {
	funcs := make(map[string]func(Node) Node)
	arr := NewArrayNode(nil, "", &funcs)
	obj := NewObjectNode(nil, "", &funcs)

	// Test Get on array
	assert.Equal(t, ErrTypeAssertion, arr.Get("key").Error())

	// Test Index on object
	assert.Equal(t, ErrTypeAssertion, obj.Index(0).Error())
}

func TestQuery(t *testing.T) {
	funcs := make(map[string]func(Node) Node)
	root := NewObjectNode(
		map[string]Node{
			"a": NewObjectNode(
				map[string]Node{
					"b": NewArrayNode(
						[]Node{
							NewStringNode("c", ".a.b[0]", &funcs),
						},
						".a.b",
						&funcs,
					),
				},
				".a",
				&funcs,
			),
		},
		"",
		&funcs,
	)

	// Test successful query
	result := root.Query("a.b[0]")
	assert.True(t, result.IsValid())
	assert.Equal(t, "c", result.String())

	// Test nonexistent path
	result = root.Query("a.b[1]")
	assert.False(t, result.IsValid())
}
