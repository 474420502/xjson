package engine

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMustMethods(t *testing.T) {
	objNode := NewObjectNode(map[string]Node{}, "", nil)
	arrNode := NewArrayNode([]Node{}, "", nil)
	strNode := NewStringNode("hello", "", nil)
	numNode := NewNumberNode(123, "", nil)
	boolNode := NewBoolNode(true, "", nil)

	assert.Panics(t, func() { objNode.MustString() })
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

func TestAppend(t *testing.T) {
	arrNode := NewArrayNode([]Node{}, "", nil)
	arrNode.Append("hello")
	arrNode.Append(123)

	assert.Equal(t, 2, arrNode.Len())
	assert.Equal(t, "hello", arrNode.Index(0).String())
	assert.Equal(t, float64(123), arrNode.Index(1).Float())

	// Test Append on non-array
	objNode := NewObjectNode(map[string]Node{}, "", nil)
	objNode.Append("test")
	assert.Error(t, objNode.Error())
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
