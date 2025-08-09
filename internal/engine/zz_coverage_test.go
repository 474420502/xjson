package engine

import (
	"errors"
	"testing"
	"time"

	"github.com/474420502/xjson/internal/core"
	"github.com/stretchr/testify/assert"
)

func TestArrayNodeZeroCoverage(t *testing.T) {
	node, err := ParseJSONToNode(`[1, "hello", true]`)
	assert.NoError(t, err)

	arrNode := node.(*arrayNode)

	// Test type assertion panics and zero-value returns
	assert.Panics(t, func() { arrNode.MustString() })
	assert.Panics(t, func() { arrNode.MustFloat() })
	assert.Equal(t, int64(0), arrNode.Int())
	assert.Panics(t, func() { arrNode.MustInt() })
	assert.False(t, arrNode.Bool())
	assert.Panics(t, func() { arrNode.MustBool() })
	assert.Equal(t, time.Time{}, arrNode.Time())
	assert.Panics(t, func() { arrNode.MustTime() })

	// Test RegisterFunc, Apply and Query
	arrNode.RegisterFunc("test", func(n core.Node) core.Node { return n })
	arrNode.Apply(func(n core.Node) bool { return true })
	arrNode.Query("[0]")
}

func TestBoolNodeZeroCoverage(t *testing.T) {
	node, err := ParseJSONToNode(`true`)
	assert.NoError(t, err)
	boolNode := node.(*boolNode)

	// Cover Get, Index, Query
	assert.Error(t, boolNode.Get("any").Error())
	assert.Error(t, boolNode.Index(0).Error())
	assert.Error(t, boolNode.Query("any").Error())

	// Cover RegisterFunc and Apply
	boolNode.RegisterFunc("test", func(n core.Node) core.Node { return n })
	boolNode.Apply(func(n core.Node) bool { return true })
}

func TestNullNodeZeroCoverage(t *testing.T) {
	node, err := ParseJSONToNode(`null`)
	assert.NoError(t, err)
	nullNode := node.(*nullNode)

	// Cover Get, Index, Query
	assert.Error(t, nullNode.Get("any").Error())
	assert.Error(t, nullNode.Index(0).Error())
	assert.Error(t, nullNode.Query("any").Error())

	// Cover RegisterFunc and Apply
	nullNode.RegisterFunc("test", func(n core.Node) core.Node { return n })
	nullNode.Apply(func(n core.Node) bool { return true })
}

func TestNumberNodeZeroCoverage(t *testing.T) {
	node, err := ParseJSONToNode(`123.45`)
	assert.NoError(t, err)
	numberNode := node.(*numberNode)

	// Cover Get, Index, Query
	assert.Error(t, numberNode.Get("any").Error())
	assert.Error(t, numberNode.Index(0).Error())
	assert.Error(t, numberNode.Query("any").Error())

	// Cover RegisterFunc and Apply
	numberNode.RegisterFunc("test", func(n core.Node) core.Node { return n })
	numberNode.Apply(func(n core.Node) bool { return true })
}

func TestStringNodeZeroCoverage(t *testing.T) {
	node, err := ParseJSONToNode(`"hello"`)
	assert.NoError(t, err)
	stringNode := node.(*stringNode)

	// Cover Get, Index, Query
	assert.Error(t, stringNode.Get("any").Error())
	assert.Error(t, stringNode.Index(0).Error())
	assert.Error(t, stringNode.Query("any").Error())

	// Cover RegisterFunc and Apply
	stringNode.RegisterFunc("test", func(n core.Node) core.Node { return n })
	stringNode.Apply(func(n core.Node) bool { return true })
}

func TestObjectNodeZeroCoverage(t *testing.T) {
	node, err := ParseJSONToNode(`{"key":"value"}`)
	assert.NoError(t, err)
	objectNode := node.(*objectNode)

	// Cover type assertion panics and zero-value returns
	assert.Panics(t, func() { objectNode.MustFloat() })
	assert.Equal(t, float64(0), objectNode.Float())
	assert.Panics(t, func() { objectNode.MustInt() })
	assert.Equal(t, int64(0), objectNode.Int())
	assert.Panics(t, func() { objectNode.MustBool() })
	assert.False(t, objectNode.Bool())
	assert.Panics(t, func() { objectNode.MustTime() })
	assert.Equal(t, time.Time{}, objectNode.Time())
	assert.Panics(t, func() { objectNode.MustArray() })
	assert.Nil(t, objectNode.Array())

	// Cover RegisterFunc, Apply, Filter and Map
	objectNode.RegisterFunc("test", func(n core.Node) core.Node { return n })
	objectNode.Apply(func(n core.Node) bool { return true })
	objectNode.Filter(func(n core.Node) bool { return true })
	objectNode.Map(func(n core.Node) interface{} { return n })
}

func TestBaseNodeForEach(t *testing.T) {
	bn := &baseNode{}
	var called bool
	bn.ForEach(func(key interface{}, value core.Node) {
		called = true
	})
	assert.False(t, called)
}

func TestInvalidNodeCoverageForZz(t *testing.T) {
	err := errors.New("test error")
	node := NewInvalidNode("root", err)
	invalid := node.(*invalidNode)

	// Cover RegisterFunc and Apply
	assert.Same(t, invalid, invalid.RegisterFunc("any", nil))
	assert.Same(t, invalid, invalid.Apply(nil))

	// Cover other methods for completeness, although they are mostly panicking or returning self
	assert.Equal(t, core.InvalidNode, invalid.Type())
	assert.Same(t, invalid, invalid.Get("key"))
	assert.Same(t, invalid, invalid.Index(0))
	assert.Same(t, invalid, invalid.Query("path"))
	invalid.ForEach(nil) // Should not panic
	assert.Equal(t, 0, invalid.Len())
	assert.Equal(t, "", invalid.String())
	assert.PanicsWithError(t, "test error", func() { invalid.MustString() })
	assert.Equal(t, float64(0), invalid.Float())
	assert.PanicsWithError(t, "test error", func() { invalid.MustFloat() })
	assert.Equal(t, int64(0), invalid.Int())
	assert.PanicsWithError(t, "test error", func() { invalid.MustInt() })
	assert.False(t, invalid.Bool())
	assert.PanicsWithError(t, "test error", func() { invalid.MustBool() })
	assert.NotNil(t, invalid.Time())
	assert.PanicsWithError(t, "test error", func() { invalid.MustTime() })
	assert.Nil(t, invalid.Array())
	assert.PanicsWithError(t, "test error", func() { invalid.MustArray() })
	assert.Nil(t, invalid.Interface())

	assert.Same(t, invalid, invalid.RegisterFunc("any", nil))
	assert.Same(t, invalid, invalid.CallFunc("any"))
	assert.Same(t, invalid, invalid.RemoveFunc("any"))

	assert.Same(t, invalid, invalid.Filter(nil))
	assert.Same(t, invalid, invalid.Map(nil))
	assert.Same(t, invalid, invalid.Set("key", "val"))
	assert.Same(t, invalid, invalid.Append("val"))
	assert.Equal(t, "invalid", invalid.Raw())

	_, ok := invalid.RawFloat()
	assert.False(t, ok)
	_, ok = invalid.RawString()
	assert.False(t, ok)

	assert.False(t, invalid.Contains("val"))
	assert.Nil(t, invalid.Strings())
	assert.Nil(t, invalid.AsMap())
	assert.PanicsWithError(t, "test error", func() { invalid.MustAsMap() })

}
