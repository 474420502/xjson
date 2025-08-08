package engine

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNodeCoverageImprovement(t *testing.T) {
	// Test AsMap method for objectNode
	t.Run("ObjectNode_AsMap", func(t *testing.T) {
		funcs := make(map[string]func(Node) Node)
		objNode := NewObjectNode(
			map[string]Node{
				"key1": NewStringNode("value1", ".key1", &funcs),
				"key2": NewNumberNode(42, ".key2", &funcs),
			},
			"",
			&funcs,
		)

		// Test AsMap method
		asMap := objNode.(interface{ AsMap() map[string]Node }).AsMap()
		assert.NotNil(t, asMap)
		assert.Equal(t, 2, len(asMap))
		assert.Equal(t, "value1", asMap["key1"].String())
		assert.Equal(t, float64(42), asMap["key2"].Float())

		// Test MustAsMap method
		mustAsMap := objNode.(interface{ MustAsMap() map[string]Node }).MustAsMap()
		assert.NotNil(t, mustAsMap)
		assert.Equal(t, 2, len(mustAsMap))
		assert.Equal(t, "value1", mustAsMap["key1"].String())
		assert.Equal(t, float64(42), mustAsMap["key2"].Float())

		// Test AsMap and MustAsMap on invalid node
		invalidNode := NewInvalidNode("test", errors.New("test error"))
		assert.Nil(t, invalidNode.(interface{ AsMap() map[string]Node }).AsMap())

		assert.Panics(t, func() {
			invalidNode.(interface{ MustAsMap() map[string]Node }).MustAsMap()
		})
	})

	// Test AsMap and MustAsMap for non-object nodes
	t.Run("NonObjectNode_AsMap", func(t *testing.T) {
		// Test array node
		arrNode := NewArrayNode([]Node{}, "", &map[string]func(Node) Node{})
		assert.Nil(t, arrNode.(interface{ AsMap() map[string]Node }).AsMap())

		assert.Panics(t, func() {
			arrNode.(interface{ MustAsMap() map[string]Node }).MustAsMap()
		})

		// Test string node
		strNode := NewStringNode("test", "", &map[string]func(Node) Node{})
		assert.Nil(t, strNode.(interface{ AsMap() map[string]Node }).AsMap())

		assert.Panics(t, func() {
			strNode.(interface{ MustAsMap() map[string]Node }).MustAsMap()
		})

		// Test number node
		numNode := NewNumberNode(42, "", &map[string]func(Node) Node{})
		assert.Nil(t, numNode.(interface{ AsMap() map[string]Node }).AsMap())

		assert.Panics(t, func() {
			numNode.(interface{ MustAsMap() map[string]Node }).MustAsMap()
		})

		// Test bool node
		boolNode := NewBoolNode(true, "", &map[string]func(Node) Node{})
		assert.Nil(t, boolNode.(interface{ AsMap() map[string]Node }).AsMap())

		assert.Panics(t, func() {
			boolNode.(interface{ MustAsMap() map[string]Node }).MustAsMap()
		})

		// Test null node
		nullNode := NewNullNode("", &map[string]func(Node) Node{})
		assert.Nil(t, nullNode.(interface{ AsMap() map[string]Node }).AsMap())

		assert.Panics(t, func() {
			nullNode.(interface{ MustAsMap() map[string]Node }).MustAsMap()
		})
	})

	// Test ForEach for all node types
	t.Run("ForEach_AllNodeTypes", func(t *testing.T) {
		// Test ForEach on array node
		funcs := make(map[string]func(Node) Node)
		arrNode := NewArrayNode(
			[]Node{
				NewStringNode("item1", "[0]", &funcs),
				NewStringNode("item2", "[1]", &funcs),
			},
			"",
			&funcs,
		)

		count := 0
		var items []string
		arrNode.ForEach(func(keyOrIndex interface{}, value Node) {
			count++
			index, ok := keyOrIndex.(int)
			assert.True(t, ok)
			assert.Equal(t, index, count-1)
			items = append(items, value.String())
		})
		assert.Equal(t, 2, count)
		assert.Equal(t, []string{"item1", "item2"}, items)

		// Test ForEach on string node (should not iterate)
		strNode := NewStringNode("test", "", &map[string]func(Node) Node{})
		strNode.ForEach(func(keyOrIndex interface{}, value Node) {
			t.Errorf("ForEach should not be called on string node")
		})

		// Test ForEach on number node (should not iterate)
		numNode := NewNumberNode(42, "", &map[string]func(Node) Node{})
		numNode.ForEach(func(keyOrIndex interface{}, value Node) {
			t.Errorf("ForEach should not be called on number node")
		})

		// Test ForEach on bool node (should not iterate)
		boolNode := NewBoolNode(true, "", &map[string]func(Node) Node{})
		boolNode.ForEach(func(keyOrIndex interface{}, value Node) {
			t.Errorf("ForEach should not be called on bool node")
		})

		// Test ForEach on null node (should not iterate)
		nullNode := NewNullNode("", &map[string]func(Node) Node{})
		nullNode.ForEach(func(keyOrIndex interface{}, value Node) {
			t.Errorf("ForEach should not be called on null node")
		})
	})

	// Test Len for all node types
	t.Run("Len_AllNodeTypes", func(t *testing.T) {
		// Test Len on object node
		objNode := NewObjectNode(
			map[string]Node{
				"key1": NewStringNode("value1", ".key1", &map[string]func(Node) Node{}),
				"key2": NewStringNode("value2", ".key2", &map[string]func(Node) Node{}),
			},
			"",
			&map[string]func(Node) Node{},
		)
		assert.Equal(t, 2, objNode.Len())

		// Test Len on array node
		arrNode := NewArrayNode(
			[]Node{
				NewStringNode("item1", "[0]", &map[string]func(Node) Node{}),
				NewStringNode("item2", "[1]", &map[string]func(Node) Node{}),
				NewStringNode("item3", "[2]", &map[string]func(Node) Node{}),
			},
			"",
			&map[string]func(Node) Node{},
		)
		assert.Equal(t, 3, arrNode.Len())

		// Test Len on string node
		strNode := NewStringNode("hello", "", &map[string]func(Node) Node{})
		assert.Equal(t, 5, strNode.Len()) // Length of the string

		// Test Len on number node
		numNode := NewNumberNode(42, "", &map[string]func(Node) Node{})
		assert.Equal(t, 0, numNode.Len())

		// Test Len on bool node
		boolNode := NewBoolNode(true, "", &map[string]func(Node) Node{})
		assert.Equal(t, 0, boolNode.Len())

		// Test Len on null node
		nullNode := NewNullNode("", &map[string]func(Node) Node{})
		assert.Equal(t, 0, nullNode.Len())
	})

	// Test String/MustString methods for all node types
	t.Run("StringMethods_AllNodeTypes", func(t *testing.T) {
		// Test string node
		strNode := NewStringNode("hello", "", &map[string]func(Node) Node{})
		assert.Equal(t, "hello", strNode.String())
		assert.Equal(t, "hello", strNode.MustString())

		// Test number node
		numNode := NewNumberNode(42.5, "", &map[string]func(Node) Node{})
		assert.Equal(t, "42.5", numNode.String())

		assert.Panics(t, func() {
			numNode.MustString()
		})

		// Test bool node
		boolNode := NewBoolNode(true, "", &map[string]func(Node) Node{})
		assert.Equal(t, "true", boolNode.String())

		assert.Panics(t, func() {
			boolNode.MustString()
		})

		// Test null node
		nullNode := NewNullNode("", &map[string]func(Node) Node{})
		assert.Equal(t, "null", nullNode.String())

		assert.Panics(t, func() {
			nullNode.MustString()
		})
	})

	// Test numeric methods for all node types
	t.Run("NumericMethods_AllNodeTypes", func(t *testing.T) {
		// Test number node
		numNode := NewNumberNode(42.7, "", &map[string]func(Node) Node{})
		assert.Equal(t, 42.7, numNode.Float())
		assert.Equal(t, 42.7, numNode.MustFloat())
		assert.Equal(t, int64(42), numNode.Int())
		assert.Equal(t, int64(42), numNode.MustInt())

		// Test string node
		strNode := NewStringNode("hello", "", &map[string]func(Node) Node{})
		assert.Equal(t, float64(0), strNode.Float())

		assert.Panics(t, func() {
			strNode.MustFloat()
		})

		assert.Equal(t, int64(0), strNode.Int())

		assert.Panics(t, func() {
			strNode.MustInt()
		})

		// Test bool node
		boolNode := NewBoolNode(true, "", &map[string]func(Node) Node{})
		assert.Equal(t, float64(0), boolNode.Float())

		assert.Panics(t, func() {
			boolNode.MustFloat()
		})

		assert.Equal(t, int64(0), boolNode.Int())

		assert.Panics(t, func() {
			boolNode.MustInt()
		})
	})

	// Test boolean methods for all node types
	t.Run("BooleanMethods_AllNodeTypes", func(t *testing.T) {
		// Test bool node
		boolNode := NewBoolNode(true, "", &map[string]func(Node) Node{})
		assert.Equal(t, true, boolNode.Bool())
		assert.Equal(t, true, boolNode.MustBool())

		// Test string node
		strNode := NewStringNode("hello", "", &map[string]func(Node) Node{})
		assert.Equal(t, false, strNode.Bool())

		assert.Panics(t, func() {
			strNode.MustBool()
		})

		// Test number node
		numNode := NewNumberNode(42.7, "", &map[string]func(Node) Node{})
		assert.Equal(t, false, numNode.Bool())

		assert.Panics(t, func() {
			numNode.MustBool()
		})
	})

	// Test time methods for all node types
	t.Run("TimeMethods_AllNodeTypes", func(t *testing.T) {
		// Test string node with valid time
		timeStr := "2024-01-02T15:04:05Z"
		strNode := NewStringNode(timeStr, "", &map[string]func(Node) Node{})
		expectedTime, _ := time.Parse(time.RFC3339, timeStr)
		assert.Equal(t, expectedTime, strNode.Time())
		assert.Equal(t, expectedTime, strNode.MustTime())

		// Test string node with invalid time
		invalidTimeStr := "invalid-time"
		invalidStrNode := NewStringNode(invalidTimeStr, "", &map[string]func(Node) Node{})
		assert.True(t, invalidStrNode.Time().IsZero())

		assert.Panics(t, func() {
			invalidStrNode.MustTime()
		})

		// Test number node
		numNode := NewNumberNode(42.7, "", &map[string]func(Node) Node{})
		assert.True(t, numNode.Time().IsZero())

		assert.Panics(t, func() {
			numNode.MustTime()
		})
	})

	// Test array methods for all node types
	t.Run("ArrayMethods_AllNodeTypes", func(t *testing.T) {
		// Test array node
		arrNode := NewArrayNode(
			[]Node{
				NewStringNode("item1", "[0]", &map[string]func(Node) Node{}),
				NewStringNode("item2", "[1]", &map[string]func(Node) Node{}),
			},
			"",
			&map[string]func(Node) Node{},
		)
		array := arrNode.Array()
		assert.Equal(t, 2, len(array))
		assert.Equal(t, "item1", array[0].String())
		assert.Equal(t, "item2", array[1].String())

		mustArray := arrNode.MustArray()
		assert.Equal(t, 2, len(mustArray))
		assert.Equal(t, "item1", mustArray[0].String())
		assert.Equal(t, "item2", mustArray[1].String())

		// Test string node
		strNode := NewStringNode("hello", "", &map[string]func(Node) Node{})
		assert.Nil(t, strNode.Array())

		assert.Panics(t, func() {
			strNode.MustArray()
		})
	})

	// Test Raw methods for all node types
	t.Run("RawMethods_AllNodeTypes", func(t *testing.T) {
		// Test string node
		strNode := NewStringNode("hello", "", &map[string]func(Node) Node{})
		assert.Equal(t, `"hello"`, strNode.Raw())

		// Test number node
		numNode := NewNumberNode(42.5, "", &map[string]func(Node) Node{})
		assert.Equal(t, "42.5", numNode.Raw())

		// Test bool node
		boolNode := NewBoolNode(true, "", &map[string]func(Node) Node{})
		assert.Equal(t, "true", boolNode.Raw())

		// Test null node
		nullNode := NewNullNode("", &map[string]func(Node) Node{})
		assert.Equal(t, "null", nullNode.Raw())
	})

	// Test Contains method for all node types
	t.Run("ContainsMethod_AllNodeTypes", func(t *testing.T) {
		// Test string node
		strNode := NewStringNode("hello world", "", &map[string]func(Node) Node{})
		assert.True(t, strNode.Contains("world"))
		assert.False(t, strNode.Contains("universe"))

		// Test array node with strings
		arrNode := NewArrayNode(
			[]Node{
				NewStringNode("item1", "[0]", &map[string]func(Node) Node{}),
				NewStringNode("item2", "[1]", &map[string]func(Node) Node{}),
			},
			"",
			&map[string]func(Node) Node{},
		)
		assert.True(t, arrNode.Contains("item1"))
		assert.False(t, arrNode.Contains("item3"))

		// Test array node with non-strings
		mixedArrNode := NewArrayNode(
			[]Node{
				NewStringNode("item1", "[0]", &map[string]func(Node) Node{}),
				NewNumberNode(42, "[1]", &map[string]func(Node) Node{}),
			},
			"",
			&map[string]func(Node) Node{},
		)
		// Behavior: Contains should return true if any string element matches, regardless of mixed types
		assert.True(t, mixedArrNode.Contains("item1"))

		// Test number node
		numNode := NewNumberNode(42, "", &map[string]func(Node) Node{})
		assert.False(t, numNode.Contains("42"))

		// Test bool node
		boolNode := NewBoolNode(true, "", &map[string]func(Node) Node{})
		assert.False(t, boolNode.Contains("true"))

		// Test null node
		nullNode := NewNullNode("", &map[string]func(Node) Node{})
		assert.False(t, nullNode.Contains("null"))

		// Test object node
		objNode := NewObjectNode(map[string]Node{}, "", &map[string]func(Node) Node{})
		assert.False(t, objNode.Contains("test"))
	})

}

func TestNodeFunctionMethods(t *testing.T) {
	t.Run("FuncMethod_AllNodeTypes", func(t *testing.T) {
		// Test Func on object node
		objNode := NewObjectNode(
			map[string]Node{
				"key": NewStringNode("value", ".key", &map[string]func(Node) Node{}),
			},
			"",
			&map[string]func(Node) Node{},
		)
		testFunc := func(n Node) Node { return n }
		result := objNode.Func("test", testFunc)
		assert.Equal(t, objNode, result)

		// Test CallFunc on object node
		called := objNode.CallFunc("test")
		assert.True(t, called.IsValid())

		// Test RemoveFunc on object node
		removed := objNode.RemoveFunc("test")
		assert.Equal(t, objNode, removed)

		// Test CallFunc after removal
		afterRemoved := objNode.CallFunc("test")
		assert.False(t, afterRemoved.IsValid())
		assert.Contains(t, afterRemoved.Error().Error(), "function test not found")

		// Test Func on array node
		arrNode := NewArrayNode([]Node{}, "", &map[string]func(Node) Node{})
		arrResult := arrNode.Func("test", testFunc)
		assert.Equal(t, arrNode, arrResult)

		// Test Func on string node
		strNode := NewStringNode("test", "", &map[string]func(Node) Node{})
		strResult := strNode.Func("test", testFunc)
		assert.Equal(t, strNode, strResult)

		// Test Func on number node
		numNode := NewNumberNode(42, "", &map[string]func(Node) Node{})
		numResult := numNode.Func("test", testFunc)
		assert.Equal(t, numNode, numResult)

		// Test Func on bool node
		boolNode := NewBoolNode(true, "", &map[string]func(Node) Node{})
		boolResult := boolNode.Func("test", testFunc)
		assert.Equal(t, boolNode, boolResult)

		// Test Func on null node
		nullNode := NewNullNode("", &map[string]func(Node) Node{})
		nullResult := nullNode.Func("test", testFunc)
		assert.Equal(t, nullNode, nullResult)
	})

	t.Run("CallFunc_InvalidFunction", func(t *testing.T) {
		objNode := NewObjectNode(
			map[string]Node{
				"key": NewStringNode("value", ".key", &map[string]func(Node) Node{}),
			},
			"",
			&map[string]func(Node) Node{},
		)

		// Test CallFunc with non-existent function
		result := objNode.CallFunc("nonexistent")
		assert.False(t, result.IsValid())
		assert.Contains(t, result.Error().Error(), "function nonexistent not found")
	})
}

func TestStreamingOperations(t *testing.T) {
	t.Run("Filter_Map_ArrayNode", func(t *testing.T) {
		// Create an array
		arrNode := NewArrayNode(
			[]Node{
				NewNumberNode(1, "[0]", &map[string]func(Node) Node{}),
				NewNumberNode(2, "[1]", &map[string]func(Node) Node{}),
				NewNumberNode(3, "[2]", &map[string]func(Node) Node{}),
			},
			"",
			&map[string]func(Node) Node{},
		)

		// Test Filter on array node
		filtered := arrNode.Filter(func(n Node) bool {
			return n.Float() > 1
		})
		assert.True(t, filtered.IsValid())
		assert.Equal(t, 2, filtered.Len())

		// Test Map on array node
		mapped := arrNode.Map(func(n Node) interface{} {
			return n.Float() * 2
		})
		assert.True(t, mapped.IsValid())
		assert.Equal(t, 3, mapped.Len())

		results := make([]float64, mapped.Len())
		for i := 0; i < mapped.Len(); i++ {
			results[i] = mapped.Index(i).Float()
		}
		assert.Contains(t, results, 2.0)
		assert.Contains(t, results, 4.0)
		assert.Contains(t, results, 6.0)
	})

	t.Run("Filter_Map_NonCollectionNodes", func(t *testing.T) {
		// Test Filter on string node (should return invalid)
		strNode := NewStringNode("test", "", &map[string]func(Node) Node{})
		filteredStr := strNode.Filter(func(n Node) bool {
			return true
		})
		assert.False(t, filteredStr.IsValid())
		assert.Equal(t, ErrTypeAssertion, filteredStr.Error())

		// Test Map on string node (should return invalid)
		mappedStr := strNode.Map(func(n Node) interface{} {
			return nil
		})
		assert.False(t, mappedStr.IsValid())
		assert.Equal(t, ErrTypeAssertion, mappedStr.Error())

		// Test Filter on number node (should return invalid)
		numNode := NewNumberNode(42, "", &map[string]func(Node) Node{})
		filteredNum := numNode.Filter(func(n Node) bool {
			return true
		})
		assert.False(t, filteredNum.IsValid())
		assert.Equal(t, ErrTypeAssertion, filteredNum.Error())

		// Test Map on number node (should return invalid)
		mappedNum := numNode.Map(func(n Node) interface{} {
			return nil
		})
		assert.False(t, mappedNum.IsValid())
		assert.Equal(t, ErrTypeAssertion, mappedNum.Error())
	})
}

func TestMutationMethods(t *testing.T) {
	t.Run("Set_ArrayNode", func(t *testing.T) {
		// Create an array of objects
		arrNode := NewArrayNode(
			[]Node{
				NewObjectNode(
					map[string]Node{
						"name": NewStringNode("Alice", "[0].name", &map[string]func(Node) Node{}),
					},
					"[0]",
					&map[string]func(Node) Node{},
				),
				NewObjectNode(
					map[string]Node{
						"name": NewStringNode("Bob", "[1].name", &map[string]func(Node) Node{}),
					},
					"[1]",
					&map[string]func(Node) Node{},
				),
			},
			"",
			&map[string]func(Node) Node{},
		)

		// Test Set on array of objects
		result := arrNode.Set("age", 25)
		assert.True(t, result.IsValid())
		assert.Equal(t, float64(25), arrNode.Index(0).Get("age").Float())
		assert.Equal(t, float64(25), arrNode.Index(1).Get("age").Float())

		// Test Set on array with non-objects (should fail)
		nonObjArrNode := NewArrayNode(
			[]Node{
				NewStringNode("item1", "[0]", &map[string]func(Node) Node{}),
				NewStringNode("item2", "[1]", &map[string]func(Node) Node{}),
			},
			"",
			&map[string]func(Node) Node{},
		)

		invalidResult := nonObjArrNode.Set("key", "value")
		assert.False(t, invalidResult.IsValid())
		assert.Equal(t, ErrTypeAssertion, invalidResult.Error())
	})

	t.Run("Append_ArrayNode", func(t *testing.T) {
		// Create an array
		arrNode := NewArrayNode(
			[]Node{
				NewStringNode("item1", "[0]", &map[string]func(Node) Node{}),
			},
			"",
			&map[string]func(Node) Node{},
		)

		// Test Append
		result := arrNode.Append("item2")
		assert.True(t, result.IsValid())
		assert.Equal(t, 2, arrNode.Len())
		assert.Equal(t, "item2", arrNode.Index(1).String())

		// Test Append on non-array (should fail)
		strNode := NewStringNode("test", "", &map[string]func(Node) Node{})
		invalidResult := strNode.Append("value")
		assert.False(t, invalidResult.IsValid())
		assert.Equal(t, ErrTypeAssertion, invalidResult.Error())
	})
}

func TestNewNodeFromInterface(t *testing.T) {
	t.Run("NewNodeFromInterface_Coverage", func(t *testing.T) {
		// Test map[string]int
		mapInt := map[string]int{"key": 42}
		node, err := NewNodeFromInterface(mapInt, "", &map[string]func(Node) Node{})
		assert.NoError(t, err)
		assert.True(t, node.IsValid())
		assert.Equal(t, float64(42), node.Get("key").Float())

		// Test map[string]int64
		mapInt64 := map[string]int64{"key": 42}
		node, err = NewNodeFromInterface(mapInt64, "", &map[string]func(Node) Node{})
		assert.NoError(t, err)
		assert.True(t, node.IsValid())
		assert.Equal(t, float64(42), node.Get("key").Float())

		// Test map[string]float64
		mapFloat64 := map[string]float64{"key": 42.5}
		node, err = NewNodeFromInterface(mapFloat64, "", &map[string]func(Node) Node{})
		assert.NoError(t, err)
		assert.True(t, node.IsValid())
		assert.Equal(t, float64(42.5), node.Get("key").Float())

		// Test map[string]string
		mapString := map[string]string{"key": "value"}
		node, err = NewNodeFromInterface(mapString, "", &map[string]func(Node) Node{})
		assert.NoError(t, err)
		assert.True(t, node.IsValid())
		assert.Equal(t, "value", node.Get("key").String())

		// Test map[string]bool
		mapBool := map[string]bool{"key": true}
		node, err = NewNodeFromInterface(mapBool, "", &map[string]func(Node) Node{})
		assert.NoError(t, err)
		assert.True(t, node.IsValid())
		assert.Equal(t, true, node.Get("key").Bool())

		// Test unsupported type
		unsupported := make(chan int)
		node, err = NewNodeFromInterface(unsupported, "", &map[string]func(Node) Node{})
		assert.Error(t, err)
		assert.Nil(t, node)
	})
}

func TestStringsMethod(t *testing.T) {
	t.Run("Strings_ArrayNode", func(t *testing.T) {
		// Test array of strings
		arrNode := NewArrayNode(
			[]Node{
				NewStringNode("item1", "[0]", &map[string]func(Node) Node{}),
				NewStringNode("item2", "[1]", &map[string]func(Node) Node{}),
			},
			"",
			&map[string]func(Node) Node{},
		)

		strings := arrNode.Strings()
		assert.Equal(t, []string{"item1", "item2"}, strings)

		// Test array with non-strings
		mixedArrNode := NewArrayNode(
			[]Node{
				NewStringNode("item1", "[0]", &map[string]func(Node) Node{}),
				NewNumberNode(42, "[1]", &map[string]func(Node) Node{}),
			},
			"",
			&map[string]func(Node) Node{},
		)

		strings = mixedArrNode.Strings()
		assert.Nil(t, strings)
	})

	t.Run("Strings_StringNode", func(t *testing.T) {
		strNode := NewStringNode("hello", "", &map[string]func(Node) Node{})
		strings := strNode.Strings()
		assert.Equal(t, []string{"hello"}, strings)
	})

	t.Run("Strings_NonArrayNodes", func(t *testing.T) {
		// Test number node
		numNode := NewNumberNode(42, "", &map[string]func(Node) Node{})
		strings := numNode.Strings()
		assert.Nil(t, strings)

		// Test bool node
		boolNode := NewBoolNode(true, "", &map[string]func(Node) Node{})
		strings = boolNode.Strings()
		assert.Nil(t, strings)

		// Test null node
		nullNode := NewNullNode("", &map[string]func(Node) Node{})
		strings = nullNode.Strings()
		assert.Nil(t, strings)
	})
}
