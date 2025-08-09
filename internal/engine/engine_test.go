package engine

import (
	"errors"
	"testing"

	"github.com/474420502/xjson/internal/core"
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
	funcs := make(map[string]func(core.Node) core.Node)
	obj := NewObjectNode(
		map[string]core.Node{
			"key1": NewStringNode("value1", ".key1", &funcs),
		},
		"",
		&funcs,
	)

	// Test successful Get
	child := obj.Get("key1")
	assert.NoError(t, child.Error())
	assert.True(t, child.IsValid())
	assert.Equal(t, "value1", child.String())

	// Test failed Get
	child = obj.Get("nonexistent")
	assert.Error(t, child.Error())
	assert.False(t, child.IsValid())
	assert.Equal(t, ErrNotFound, child.Error())
}

func TestIndex(t *testing.T) {
	funcs := make(map[string]func(core.Node) core.Node)
	arr := NewArrayNode(
		[]core.Node{
			NewStringNode("value0", "[0]", &funcs),
		},
		"",
		&funcs,
	)

	// Test successful Index
	child := arr.Index(0)
	assert.NoError(t, child.Error())
	assert.True(t, child.IsValid())
	assert.Equal(t, "value0", child.String())

	// Test failed Index (out of bounds)
	child = arr.Index(1)
	assert.Error(t, child.Error())
	assert.False(t, child.IsValid())
	assert.Equal(t, ErrIndexOutOfBounds, child.Error())
}

func TestInvalidOperations(t *testing.T) {
	funcs := make(map[string]func(core.Node) core.Node)
	arr := NewArrayNode(nil, "", &funcs)
	obj := NewObjectNode(nil, "", &funcs)

	// Test Get on array
	assert.Equal(t, ErrTypeAssertion, arr.Get("key").Error())

	// Test Index on object
	assert.Equal(t, ErrTypeAssertion, obj.Index(0).Error())
}

func TestQuery(t *testing.T) {
	funcs := make(map[string]func(core.Node) core.Node)
	root := NewObjectNode(
		map[string]core.Node{
			"a": NewObjectNode(
				map[string]core.Node{
					"b": NewArrayNode(
						[]core.Node{
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
	assert.NoError(t, result.Error())
	assert.True(t, result.IsValid())
	assert.Equal(t, "c", result.String())

	// Test nonexistent path
	result = root.Query("a.b[1]")
	assert.Error(t, result.Error())
	assert.False(t, result.IsValid())
}

func TestEvaluateQueryEdgeCases(t *testing.T) {
	t.Run("FlattenWithInvalidNodes", func(t *testing.T) {
		// Test case where flattenIfNestedArrays encounters an invalid node
		nestedArray := NewArrayNode([]core.Node{
			NewArrayNode([]core.Node{
				NewStringNode("a", "", nil),
				NewInvalidNode("", errors.New("invalid node")),
			}, "", nil),
			NewStringNode("b", "", nil),
		}, "", nil)
		ops, _ := ParseQuery("*")
		result := EvaluateQuery(nestedArray, ops)
		assert.NoError(t, result.Error())
		assert.True(t, result.IsValid())
		assert.Equal(t, 2, result.Len(), "Should contain 2 valid nodes after flattening")
		if assert.NoError(t, result.Index(0).Error()) {
			assert.Equal(t, "a", result.Index(0).String())
		}
		if assert.NoError(t, result.Index(1).Error()) {
			assert.Equal(t, "b", result.Index(1).String())
		}
	})

	t.Run("GetOnArrayWithSomeInvalid", func(t *testing.T) {
		// Test Get on an array where some elements don't have the key
		arrayNode := NewArrayNode([]core.Node{
			NewObjectNode(map[string]core.Node{"key": NewStringNode("v1", "", nil)}, "", nil),
			NewObjectNode(map[string]core.Node{"other": NewStringNode("v2", "", nil)}, "", nil),
			NewObjectNode(map[string]core.Node{"key": NewStringNode("v3", "", nil)}, "", nil),
		}, "", nil)
		ops, _ := ParseQuery("key")
		result := EvaluateQuery(arrayNode, ops)
		assert.NoError(t, result.Error())
		assert.True(t, result.IsValid())
		assert.Equal(t, 2, result.Len(), "Should only return nodes that had the key")
		if assert.NoError(t, result.Index(0).Error()) {
			assert.Equal(t, "v1", result.Index(0).String())
		}
		if assert.NoError(t, result.Index(1).Error()) {
			assert.Equal(t, "v3", result.Index(1).String())
		}
	})

	t.Run("ParseInvalidIndex", func(t *testing.T) {
		_, err := ParseQuery("[abc]")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid array index")
	})
}

func TestEvaluateQueryCoverage(t *testing.T) {
	funcs := make(map[string]func(core.Node) core.Node)

	t.Run("WildcardOnNonObjectOrArray", func(t *testing.T) {
		node := NewStringNode("test", "", &funcs)
		ops, _ := ParseQuery("*")
		result := EvaluateQuery(node, ops)
		assert.Error(t, result.Error())
		assert.False(t, result.IsValid())
	})

	t.Run("WildcardOnArrayWithoutNestedArrays", func(t *testing.T) {
		node := NewArrayNode([]core.Node{NewStringNode("a", "", &funcs)}, "", &funcs)
		ops, _ := ParseQuery("*")
		result := EvaluateQuery(node, ops)
		assert.NoError(t, result.Error())
		assert.True(t, result.IsValid())
		assert.Equal(t, core.ArrayNode, result.Type())
	})

	t.Run("OperationOnInvalidNode", func(t *testing.T) {
		node := NewInvalidNode("", errors.New("invalid"))
		ops, _ := ParseQuery("a")
		result := EvaluateQuery(node, ops)
		assert.Error(t, result.Error())
		assert.False(t, result.IsValid())
	})

	t.Run("FlattenOnNonArray", func(t *testing.T) {
		localFuncs := make(map[string]func(core.Node) core.Node)
		localFuncs["someFunc"] = func(n core.Node) core.Node { return n }
		node := NewObjectNode(nil, "", &localFuncs)
		ops, _ := ParseQuery("[@someFunc]")
		result := EvaluateQuery(node, ops)
		assert.NoError(t, result.Error())
		assert.True(t, result.IsValid())
	})

	t.Run("FlattenOnArrayWithNoNestedArrays", func(t *testing.T) {
		localFuncs := make(map[string]func(core.Node) core.Node)
		localFuncs["someFunc"] = func(n core.Node) core.Node { return n }
		node := NewArrayNode([]core.Node{NewStringNode("a", "", &localFuncs)}, "", &localFuncs)
		ops, _ := ParseQuery("[@someFunc]")
		result := EvaluateQuery(node, ops)
		assert.NoError(t, result.Error())
		assert.True(t, result.IsValid())
	})

	t.Run("MalformedPath", func(t *testing.T) {
		_, err := ParseQuery("a[b")
		assert.Error(t, err)
		_, err = ParseQuery("a[0]extra")
		assert.Error(t, err)
	})
}
