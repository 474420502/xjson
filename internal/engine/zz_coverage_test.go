package engine

import (
	"fmt"
	"testing"
	"time"

	"github.com/474420502/xjson/internal/core"
	"github.com/stretchr/testify/assert"
)

// TestInvalidNode_Coverage ensures all methods on an invalid node are covered.
func TestInvalidNode_Coverage(t *testing.T) {
	err := fmt.Errorf("test error")
	invalid := NewInvalidNode("/path", err)

	assert.False(t, invalid.IsValid())
	assert.Equal(t, err, invalid.Error())
	assert.Equal(t, core.InvalidNode, invalid.Type())
	assert.Equal(t, "/path", invalid.Path())
	assert.Equal(t, "invalid", invalid.Raw())

	// Test all panic methods
	assert.Panics(t, func() { invalid.MustString() })
	assert.Panics(t, func() { invalid.MustFloat() })
	assert.Panics(t, func() { invalid.MustInt() })
	assert.Panics(t, func() { invalid.MustBool() })
	assert.Panics(t, func() { invalid.MustTime() })
	assert.Panics(t, func() { invalid.MustArray() })

	// Test zero-value returning methods
	assert.Equal(t, "", invalid.String())
	assert.Equal(t, 0.0, invalid.Float())
	assert.Equal(t, int64(0), invalid.Int())
	assert.False(t, invalid.Bool())
	assert.True(t, invalid.Time().IsZero())
	assert.Nil(t, invalid.Array())
	assert.Nil(t, invalid.Interface())
	assert.Equal(t, 0, invalid.Len())
	invalid.ForEach(func(keyOrIndex interface{}, value core.Node) {
		t.Fatal("should not iterate over invalid node")
	})

	// Chainable methods should return the same invalid node
	assert.Equal(t, invalid, invalid.Get("key"))
	assert.Equal(t, invalid, invalid.Index(0))
	assert.Equal(t, invalid, invalid.Query("a.b"))
	assert.Equal(t, invalid, invalid.RegisterFunc("f", nil))
	assert.Equal(t, invalid, invalid.CallFunc("f"))
	assert.Equal(t, invalid, invalid.RemoveFunc("f"))
	assert.Equal(t, invalid, invalid.Filter(nil))
	assert.Equal(t, invalid, invalid.Map(nil))
	assert.Equal(t, invalid, invalid.Set("k", "v"))
	assert.Equal(t, invalid, invalid.Append("v"))

	// Raw value methods
	f, ok := invalid.RawFloat()
	assert.False(t, ok)
	assert.Equal(t, 0.0, f)
	s, ok := invalid.RawString()
	assert.False(t, ok)
	assert.Equal(t, "", s)

	assert.Nil(t, invalid.Strings())
	assert.False(t, invalid.Contains("a"))
}

// TestObjectNode_Coverage covers edge cases and panics for objectNode.
func TestObjectNode_Coverage(t *testing.T) {
	newNode := func() core.Node {
		node, err := ParseJSONToNode(`{"a": 1, "b": "str"}`)
		assert.NoError(t, err)
		return node
	}

	// Type assertion panics
	assert.Panics(t, func() { newNode().MustFloat() })
	assert.Panics(t, func() { newNode().MustInt() })
	assert.Panics(t, func() { newNode().MustBool() })
	assert.Panics(t, func() { newNode().MustTime() })
	assert.Panics(t, func() { newNode().MustArray() })

	// Zero-value returns
	nodeForZero := newNode()
	assert.Equal(t, 0.0, nodeForZero.Float())
	assert.Equal(t, int64(0), nodeForZero.Int())
	assert.False(t, nodeForZero.Bool())
	assert.True(t, nodeForZero.Time().IsZero())
	assert.Nil(t, nodeForZero.Array())

	// Indexing on object is an error
	indexNode := newNode().Index(0)
	assert.Error(t, indexNode.Error())
	assert.False(t, indexNode.IsValid())

	// Raw value methods
	nodeForRaw := newNode()
	_, ok := nodeForRaw.RawFloat()
	assert.False(t, ok)
	_, ok = nodeForRaw.RawString()
	assert.False(t, ok)

	// Contains/Strings are no-ops
	nodeForContains := newNode()
	assert.False(t, nodeForContains.Contains("str"))
	assert.Nil(t, nodeForContains.Strings())

	// Append is an error
	appendNode := newNode().Append("v")
	assert.Error(t, appendNode.Error())
	assert.False(t, appendNode.IsValid())

	// Filter/Map on values
	nodeForFilter := newNode()
	arr := nodeForFilter.Filter(func(n core.Node) bool { return n.Type() == core.NumberNode })
	assert.NoError(t, arr.Error())
	assert.True(t, arr.IsValid())
	assert.Equal(t, 1, arr.Len())

	nodeForMap := newNode()
	arr = nodeForMap.Map(func(n core.Node) interface{} { return int(n.Type()) })
	if arr.Error() != nil {
		t.Error(arr.Error())
	} else {
		assert.True(t, arr.IsValid())
	}

	// The Map function on an object node applies to its values and returns an array node.
	// The test JSON has 2 values ("a": 1 and "b": "str"), so the result should have 2 elements.
	assert.Equal(t, 2, arr.Len())

	// ForEach on object node
	count := 0
	newNode().ForEach(func(key interface{}, value core.Node) {
		count++
		// Also assert the types to be safe
		assert.IsType(t, "", key)
		assert.NotNil(t, value)
	})
	assert.Equal(t, 2, count)

	// Set with a value that causes NewNodeFromInterface to fail
	nodeForSetFail := newNode()
	nodeForSetFail = nodeForSetFail.Set("c", make(chan int))
	assert.Error(t, nodeForSetFail.Error())
}

// TestArrayNode_Coverage covers edge cases and panics for arrayNode.
func TestArrayNode_Coverage(t *testing.T) {
	node, err := ParseJSONToNode(`[1, "str", true]`)
	assert.NoError(t, err)

	// Type assertion panics
	assert.Panics(t, func() { node.MustString() })
	assert.Panics(t, func() { node.MustFloat() })
	assert.Panics(t, func() { node.MustInt() })
	assert.Panics(t, func() { node.MustBool() })
	assert.Panics(t, func() { node.MustTime() })

	// Zero-value returns
	assert.Equal(t, `[1,"str",true]`, node.String())
	assert.Equal(t, 0.0, node.Float())
	assert.Equal(t, int64(0), node.Int())
	assert.False(t, node.Bool())
	assert.True(t, node.Time().IsZero())

	// Get on array is an error
	getNode := node.Get("key")
	assert.Error(t, getNode.Error())
	assert.False(t, getNode.IsValid())

	// Raw value methods
	_, ok := node.RawFloat()
	assert.False(t, ok)
	_, ok = node.RawString()
	assert.False(t, ok)

	// Set on array of non-objects is an error
	setNode := node.Set("key", "v")
	assert.Error(t, setNode.Error())
	assert.False(t, setNode.IsValid())

	// Strings on mixed-type array is an error
	assert.Nil(t, node.Strings())
	assert.Error(t, node.Error())
}

// TestStringNode_Coverage covers edge cases and panics for stringNode.
func TestStringNode_Coverage(t *testing.T) {
	newNode := func() core.Node {
		node, _ := ParseJSONToNode(`"hello"`)
		return node
	}
	nodeTime, _ := ParseJSONToNode(fmt.Sprintf(`"%s"`, time.Now().Format(time.RFC3339Nano)))

	// Type assertion panics
	assert.Panics(t, func() { newNode().MustFloat() })
	assert.Panics(t, func() { newNode().MustInt() })
	assert.Panics(t, func() { newNode().MustBool() })
	assert.Panics(t, func() { newNode().MustArray() })
	assert.Panics(t, func() { newNode().MustTime() }) // Not a valid time format

	// Zero-value returns
	nodeForZero := newNode()
	assert.Equal(t, 0.0, nodeForZero.Float())
	assert.Equal(t, int64(0), nodeForZero.Int())
	assert.False(t, nodeForZero.Bool())
	assert.True(t, nodeForZero.Time().IsZero())
	assert.Nil(t, nodeForZero.Array())

	// Invalid operations
	assert.Error(t, newNode().Get("key").Error())
	assert.Error(t, newNode().Index(0).Error())
	assert.Error(t, newNode().Query("a").Error())

	// Filter and Map on string node with nil function should return invalid node
	assert.Error(t, newNode().Filter(nil).Error())
	assert.Error(t, newNode().Map(nil).Error())

	assert.Error(t, newNode().Set("k", "v").Error())
	assert.Error(t, newNode().Append("v").Error())

	// Raw float
	_, ok := newNode().RawFloat()
	assert.False(t, ok)

	// Strings returns a slice with the string
	assert.Equal(t, []string{"hello"}, newNode().Strings())

	// MustTime should succeed for a valid time string
	assert.NotPanics(t, func() { nodeTime.MustTime() })
}

// TestNumberNode_Coverage covers edge cases and panics for numberNode.
func TestNumberNode_Coverage(t *testing.T) {
	node, _ := ParseJSONToNode("123.45")

	// Type assertion panics
	assert.Panics(t, func() { node.MustString() })
	assert.Panics(t, func() { node.MustBool() })
	assert.Panics(t, func() { node.MustTime() })
	assert.Panics(t, func() { node.MustArray() })

	// Zero-value returns
	assert.False(t, node.Bool())
	assert.True(t, node.Time().IsZero())
	assert.Nil(t, node.Array())

	// Invalid operations
	assert.Error(t, node.Get("key").Error())
	assert.Error(t, node.Index(0).Error())
	assert.Error(t, node.Query("a").Error())
	assert.Error(t, node.Filter(nil).Error())
	assert.Error(t, node.Map(nil).Error())
	assert.Error(t, node.Set("k", "v").Error())
	assert.Error(t, node.Append("v").Error())

	// Raw string
	_, ok := node.RawString()
	assert.False(t, ok)

	// Contains/Strings are no-ops
	assert.False(t, node.Contains("123"))
	assert.Nil(t, node.Strings())
}

// TestBoolNode_Coverage covers edge cases and panics for boolNode.
func TestBoolNode_Coverage(t *testing.T) {
	node, _ := ParseJSONToNode("true")

	// Type assertion panics
	assert.Panics(t, func() { node.MustString() })
	assert.Panics(t, func() { node.MustFloat() })
	assert.Panics(t, func() { node.MustInt() })
	assert.Panics(t, func() { node.MustTime() })
	assert.Panics(t, func() { node.MustArray() })

	// Invalid operations
	assert.Error(t, node.Get("key").Error())
	assert.Error(t, node.Index(0).Error())
	assert.Error(t, node.Query("a").Error())
	assert.Error(t, node.Filter(nil).Error())
	assert.Error(t, node.Map(nil).Error())
	assert.Error(t, node.Set("k", "v").Error())
	assert.Error(t, node.Append("v").Error())

	// Raw values
	_, ok := node.RawFloat()
	assert.False(t, ok)
	_, ok = node.RawString()
	assert.False(t, ok)

	// Contains/Strings are no-ops
	assert.False(t, node.Contains("true"))
	assert.Nil(t, node.Strings())
}

// TestNullNode_Coverage covers edge cases and panics for nullNode.
func TestNullNode_Coverage(t *testing.T) {
	node, _ := ParseJSONToNode("null")

	// Type assertion panics
	assert.Panics(t, func() { node.MustString() })
	assert.Panics(t, func() { node.MustFloat() })
	assert.Panics(t, func() { node.MustInt() })
	assert.Panics(t, func() { node.MustBool() })
	assert.Panics(t, func() { node.MustTime() })
	assert.Panics(t, func() { node.MustArray() })

	// Invalid operations
	assert.Error(t, node.Get("key").Error())
	assert.Error(t, node.Index(0).Error())
	assert.Error(t, node.Query("a").Error())
	assert.Error(t, node.Filter(nil).Error())
	assert.Error(t, node.Map(nil).Error())
	assert.Error(t, node.Set("k", "v").Error())
	assert.Error(t, node.Append("v").Error())

	// Raw values
	_, ok := node.RawFloat()
	assert.False(t, ok)
	_, ok = node.RawString()
	assert.False(t, ok)

	// Contains/Strings are no-ops
	assert.False(t, node.Contains("null"))
	assert.Nil(t, node.Strings())
}

// TestNewNodeFromInterface_Coverage covers unsupported types.
func TestNewNodeFromInterface_Coverage(t *testing.T) {
	// Unsupported type
	ch := make(chan int)
	_, err := NewNodeFromInterface(ch, "", nil)
	assert.Error(t, err)

	// Error within nested map
	m := map[string]interface{}{"bad": ch}
	_, err = NewNodeFromInterface(m, "", nil)
	assert.Error(t, err)

	// Error within nested slice
	s := []interface{}{ch}
	_, err = NewNodeFromInterface(s, "", nil)
	assert.Error(t, err)
}

// TestBuildArrayNode_Coverage covers the internal buildArrayNode function.
func TestBuildArrayNode_Coverage(t *testing.T) {
	// This function is internal, but we can test its effects via parsing.
	data := []interface{}{float64(1), "test", true, nil, map[string]interface{}{"a": "b"}, []interface{}{float64(2)}}
	node, err := NewNodeFromInterface(data, "", nil)
	assert.NoError(t, err)
	assert.Equal(t, core.ArrayNode, node.Type())
	assert.Equal(t, 6, node.Len())
	assert.Equal(t, int64(1), node.Index(0).Int())
	assert.Equal(t, "test", node.Index(1).String())
	assert.True(t, node.Index(2).Bool())
	assert.Equal(t, core.NullNode, node.Index(3).Type())
	assert.Equal(t, "b", node.Index(4).Get("a").String())
	assert.Equal(t, int64(2), node.Index(5).Index(0).Int())
}

// TestQuery_Coverage covers query parser and evaluator errors.
func TestQuery_Coverage(t *testing.T) {
	node, _ := ParseJSONToNode("{}")
	// Invalid query syntax
	res := node.Query("[invalid")
	assert.False(t, res.IsValid())
	assert.Error(t, res.Error())

	// Function that returns an invalid node
	node.RegisterFunc("fail", func(n core.Node) core.Node {
		return NewInvalidNode(n.Path(), fmt.Errorf("func failed"))
	})
	res = node.Query("/[@fail]")
	assert.False(t, res.IsValid())
	assert.Error(t, res.Error())
}
