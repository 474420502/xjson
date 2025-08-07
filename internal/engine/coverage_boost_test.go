package engine

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStringsAndContainsMethods(t *testing.T) {
	t.Run("Strings method on arrayNode", func(t *testing.T) {
		arrNode := NewArrayNode([]Node{
			NewStringNode("apple", "", nil),
			NewStringNode("banana", "", nil),
			NewStringNode("cherry", "", nil),
		}, "", nil)

		strings := arrNode.Strings()
		assert.ElementsMatch(t, []string{"apple", "banana", "cherry"}, strings)

		// Test empty array
		emptyArrNode := NewArrayNode([]Node{}, "", nil)
		assert.Empty(t, emptyArrNode.Strings())

		// Test array with non-string elements
		mixedArrNode := NewArrayNode([]Node{
			NewStringNode("apple", "", nil),
			NewNumberNode(42, "", nil),
			NewBoolNode(true, "", nil),
		}, "", nil)
		// When array contains non-string elements, Strings should return nil and set error
		strings = mixedArrNode.Strings()
		assert.Nil(t, strings)
		assert.False(t, mixedArrNode.IsValid())
	})

	t.Run("Strings method on stringNode", func(t *testing.T) {
		strNode := NewStringNode("hello", "", nil)
		assert.Equal(t, []string{"hello"}, strNode.Strings())

		// Test empty string
		emptyStrNode := NewStringNode("", "", nil)
		assert.Equal(t, []string{""}, emptyStrNode.Strings())
	})

	t.Run("Strings method on other node types", func(t *testing.T) {
		numNode := NewNumberNode(123, "", nil)
		assert.Nil(t, numNode.Strings())

		boolNode := NewBoolNode(true, "", nil)
		assert.Nil(t, boolNode.Strings())

		nullNode := NewNullNode("", nil)
		assert.Nil(t, nullNode.Strings())
	})

	t.Run("Contains method on stringNode", func(t *testing.T) {
		strNode := NewStringNode("hello world", "", nil)
		assert.True(t, strNode.Contains("hello"))
		assert.True(t, strNode.Contains("world"))
		assert.False(t, strNode.Contains("universe"))

		// Test empty substring
		assert.True(t, strNode.Contains(""))
		// Test case sensitivity
		assert.False(t, strNode.Contains("HELLO"))
	})

	t.Run("Contains method on arrayNode", func(t *testing.T) {
		arrNode := NewArrayNode([]Node{
			NewStringNode("apple", "", nil),
			NewStringNode("banana", "", nil),
			NewStringNode("cherry", "", nil),
		}, "", nil)

		assert.True(t, arrNode.Contains("apple"))
		assert.True(t, arrNode.Contains("banana"))
		assert.False(t, arrNode.Contains("orange"))

		// Test with non-string elements
		mixedArrNode := NewArrayNode([]Node{
			NewStringNode("apple", "", nil),
			NewNumberNode(42, "", nil),
			NewBoolNode(true, "", nil),
		}, "", nil)
		assert.True(t, mixedArrNode.Contains("apple"))
		assert.False(t, mixedArrNode.Contains("42"))
		assert.False(t, mixedArrNode.Contains("true"))
	})

	t.Run("Contains method on other node types", func(t *testing.T) {
		numNode := NewNumberNode(123, "", nil)
		assert.False(t, numNode.Contains("anything"))

		boolNode := NewBoolNode(true, "", nil)
		assert.False(t, boolNode.Contains("anything"))

		nullNode := NewNullNode("", nil)
		assert.False(t, nullNode.Contains("anything"))
	})
}

func TestNodeWrapperMethods(t *testing.T) {
	// Test ForEach method on different node types
	t.Run("ForEach method", func(t *testing.T) {
		// Test with object node
		objNode := NewObjectNode(map[string]Node{
			"key1": NewStringNode("value1", "", nil),
			"key2": NewStringNode("value2", "", nil),
		}, "", nil)

		count := 0
		keys := []string{}
		values := []string{}

		objNode.ForEach(func(keyOrIndex interface{}, value Node) {
			count++
			if key, ok := keyOrIndex.(string); ok {
				keys = append(keys, key)
			}
			values = append(values, value.String())
		})

		assert.Equal(t, 2, count)
		assert.ElementsMatch(t, []string{"key1", "key2"}, keys)
		assert.ElementsMatch(t, []string{"value1", "value2"}, values)

		// Test with array node
		arrNode := NewArrayNode([]Node{
			NewStringNode("item1", "", nil),
			NewStringNode("item2", "", nil),
		}, "", nil)

		count = 0
		indices := []int{}
		values = []string{}

		arrNode.ForEach(func(keyOrIndex interface{}, value Node) {
			count++
			if idx, ok := keyOrIndex.(int); ok {
				indices = append(indices, idx)
			}
			values = append(values, value.String())
		})

		assert.Equal(t, 2, count)
		assert.ElementsMatch(t, []int{0, 1}, indices)
		assert.ElementsMatch(t, []string{"item1", "item2"}, values)

		// Test with scalar nodes (should not call the function)
		strNode := NewStringNode("test", "", nil)
		count = 0
		strNode.ForEach(func(keyOrIndex interface{}, value Node) {
			count++
		})
		assert.Equal(t, 0, count)
	})

	// Test Len method on different node types
	t.Run("Len method", func(t *testing.T) {
		// Test with object node - this should return the number of keys
		objNode := NewObjectNode(map[string]Node{
			"key1": NewStringNode("value1", "", nil),
			"key2": NewStringNode("value2", "", nil),
		}, "", nil)
		assert.Equal(t, 2, objNode.Len())

		// Test with array node
		arrNode := NewArrayNode([]Node{
			NewStringNode("item1", "", nil),
			NewStringNode("item2", "", nil),
			NewStringNode("item3", "", nil),
		}, "", nil)
		assert.Equal(t, 3, arrNode.Len())

		// Test with scalar nodes
		strNode := NewStringNode("test", "", nil)
		// String nodes return the length of their string value
		assert.Equal(t, 4, strNode.Len())

		numNode := NewNumberNode(42, "", nil)
		// Number nodes return 0 for Len()
		assert.Equal(t, 0, numNode.Len())
	})

	// Test Type method on different node types
	t.Run("Type method", func(t *testing.T) {
		objNode := NewObjectNode(map[string]Node{}, "", nil)
		assert.Equal(t, ObjectNode, objNode.Type())

		arrNode := NewArrayNode([]Node{}, "", nil)
		assert.Equal(t, ArrayNode, arrNode.Type())

		strNode := NewStringNode("test", "", nil)
		assert.Equal(t, StringNode, strNode.Type())

		numNode := NewNumberNode(42, "", nil)
		assert.Equal(t, NumberNode, numNode.Type())

		boolNode := NewBoolNode(true, "", nil)
		assert.Equal(t, BoolNode, boolNode.Type())

		nullNode := NewNullNode("", nil)
		assert.Equal(t, NullNode, nullNode.Type())
	})
}

func TestStringNodeAdditionalMethods(t *testing.T) {
	t.Run("String and MustString methods", func(t *testing.T) {
		strNode := NewStringNode("test value", "", nil)
		assert.Equal(t, "test value", strNode.String())
		assert.Equal(t, "test value", strNode.MustString())

		// Test with error
		invalidNode := NewInvalidNode("", errors.New("test error"))
		assert.Equal(t, "", invalidNode.String())
		assert.Panics(t, func() {
			invalidNode.MustString()
		})
	})

	t.Run("Bool conversion", func(t *testing.T) {
		// Test valid boolean strings
		trueNode := NewStringNode("true", "", nil)
		falseNode := NewStringNode("false", "", nil)
		// String nodes always return false for Bool()
		assert.False(t, trueNode.Bool())
		assert.False(t, falseNode.Bool())

		// Test invalid boolean string
		invalidNode := NewStringNode("not a boolean", "", nil)
		assert.False(t, invalidNode.Bool())
	})

	t.Run("Time conversion", func(t *testing.T) {
		// Test valid time string
		timeStr := "2023-01-01T00:00:00Z"
		timeNode := NewStringNode(timeStr, "", nil)
		expectedTime, _ := time.Parse(time.RFC3339, timeStr)
		assert.Equal(t, expectedTime, timeNode.Time())

		// Test invalid time string
		invalidTimeNode := NewStringNode("not a time", "", nil)
		assert.True(t, invalidTimeNode.Time().IsZero())
	})

	t.Run("Float and Int conversion", func(t *testing.T) {
		// Test valid number strings
		floatNode := NewStringNode("3.14", "", nil)
		intNode := NewStringNode("42", "", nil)
		// String nodes return 0 for numeric values
		assert.Equal(t, float64(0), floatNode.Float())
		assert.Equal(t, int64(0), intNode.Int())

		// Test invalid number strings
		invalidNode := NewStringNode("not a number", "", nil)
		assert.Equal(t, float64(0), invalidNode.Float())
		assert.Equal(t, int64(0), invalidNode.Int())
	})

	t.Run("Array method", func(t *testing.T) {
		strNode := NewStringNode("test", "", nil)
		// String nodes return nil for Array()
		result := strNode.Array()
		assert.Nil(t, result)
	})

	t.Run("RawString method", func(t *testing.T) {
		strNode := NewStringNode("test value", "", nil)
		value, ok := strNode.RawString()
		assert.True(t, ok)
		assert.Equal(t, "test value", value)

		// Test with error
		invalidNode := NewInvalidNode("", errors.New("test error"))
		_, ok = invalidNode.RawString()
		assert.False(t, ok)
	})

	t.Run("Contains and Strings methods", func(t *testing.T) {
		strNode := NewStringNode("hello world", "", nil)
		// String nodes should return their value in Strings
		assert.Equal(t, []string{"hello world"}, strNode.Strings())
		// Contains should check if the string value contains the substring
		assert.True(t, strNode.Contains("hello"))
		assert.True(t, strNode.Contains("world"))
		assert.False(t, strNode.Contains("universe"))
	})
}

func TestNumberNodeAdditionalMethods(t *testing.T) {
	t.Run("String and MustString methods", func(t *testing.T) {
		// According to the changes, MustString should now panic for number nodes
		numNode := NewNumberNode(3.14, "", nil)
		assert.Equal(t, "3.14", numNode.String())
		assert.Panics(t, func() {
			numNode.MustString()
		})

		// Test with error
		invalidNode := NewInvalidNode("", errors.New("test error"))
		assert.Equal(t, "", invalidNode.String())
		assert.Panics(t, func() {
			invalidNode.MustString()
		})
	})

	t.Run("Bool conversion", func(t *testing.T) {
		zeroNode := NewNumberNode(0, "", nil)
		nonZeroNode := NewNumberNode(1, "", nil)
		// Number nodes always return false for Bool()
		assert.False(t, zeroNode.Bool())
		assert.False(t, nonZeroNode.Bool())

		// Test with error
		invalidNode := NewInvalidNode("", errors.New("test error"))
		assert.False(t, invalidNode.Bool())
	})

	t.Run("Time conversion", func(t *testing.T) {
		// Number nodes should return zero time
		numNode := NewNumberNode(1234567890, "", nil)
		assert.True(t, numNode.Time().IsZero())

		// Test with error
		invalidNode := NewInvalidNode("", errors.New("test error"))
		assert.True(t, invalidNode.Time().IsZero())
	})

	t.Run("Interface method", func(t *testing.T) {
		numNode := NewNumberNode(3.14, "", nil)
		assert.Equal(t, 3.14, numNode.Interface())

		// Test with error
		invalidNode := NewInvalidNode("", errors.New("test error"))
		assert.Nil(t, invalidNode.Interface())
	})

	t.Run("RawFloat method", func(t *testing.T) {
		numNode := NewNumberNode(3.14, "", nil)
		value, ok := numNode.RawFloat()
		assert.True(t, ok)
		assert.Equal(t, 3.14, value)

		// Test with error
		invalidNode := NewInvalidNode("", errors.New("test error"))
		_, ok = invalidNode.RawFloat()
		assert.False(t, ok)
	})

	t.Run("Contains and Strings methods", func(t *testing.T) {
		numNode := NewNumberNode(42, "", nil)
		// These methods should be no-ops for number nodes
		assert.False(t, numNode.Contains("anything"))
		assert.Nil(t, numNode.Strings())
	})
}

func TestBoolNodeAdditionalMethods(t *testing.T) {
	t.Run("String and MustString methods", func(t *testing.T) {
		trueNode := NewBoolNode(true, "", nil)
		falseNode := NewBoolNode(false, "", nil)
		assert.Equal(t, "true", trueNode.String())
		assert.Equal(t, "false", falseNode.String())
		// MustString should panic for bool nodes
		assert.Panics(t, func() {
			trueNode.MustString()
		})

		// Test with error
		invalidNode := NewInvalidNode("", errors.New("test error"))
		assert.Equal(t, "", invalidNode.String())
		assert.Panics(t, func() {
			invalidNode.MustString()
		})
	})

	t.Run("Float, Int, and Bool methods", func(t *testing.T) {
		trueNode := NewBoolNode(true, "", nil)
		falseNode := NewBoolNode(false, "", nil)
		assert.Equal(t, float64(0), trueNode.Float())
		assert.Equal(t, int64(0), trueNode.Int())
		assert.True(t, trueNode.Bool())
		assert.False(t, falseNode.Bool())

		// Test Must methods with panic
		assert.Panics(t, func() {
			trueNode.MustFloat()
		})
		assert.Panics(t, func() {
			trueNode.MustInt()
		})

		// Test with error
		invalidNode := NewInvalidNode("", errors.New("test error"))
		assert.Equal(t, float64(0), invalidNode.Float())
		assert.Equal(t, int64(0), invalidNode.Int())
		assert.False(t, invalidNode.Bool())
	})

	t.Run("Time conversion", func(t *testing.T) {
		// Bool nodes should return zero time
		boolNode := NewBoolNode(true, "", nil)
		assert.True(t, boolNode.Time().IsZero())

		// Test with error
		invalidNode := NewInvalidNode("", errors.New("test error"))
		assert.True(t, invalidNode.Time().IsZero())
	})

	t.Run("Interface method", func(t *testing.T) {
		trueNode := NewBoolNode(true, "", nil)
		falseNode := NewBoolNode(false, "", nil)
		assert.Equal(t, true, trueNode.Interface())
		assert.Equal(t, false, falseNode.Interface())

		// Test with error
		invalidNode := NewInvalidNode("", errors.New("test error"))
		assert.Nil(t, invalidNode.Interface())
	})

	t.Run("Contains and Strings methods", func(t *testing.T) {
		boolNode := NewBoolNode(true, "", nil)
		// These methods should be no-ops for bool nodes
		assert.False(t, boolNode.Contains("anything"))
		assert.Nil(t, boolNode.Strings())
	})
}

func TestNullNodeAdditionalMethods(t *testing.T) {
	t.Run("String and MustString methods", func(t *testing.T) {
		nullNode := NewNullNode("", nil)
		assert.Equal(t, "null", nullNode.String())
		// MustString should panic for null nodes
		assert.Panics(t, func() {
			nullNode.MustString()
		})

		// Test with error
		invalidNode := NewInvalidNode("", errors.New("test error"))
		assert.Equal(t, "", invalidNode.String())
		assert.Panics(t, func() {
			invalidNode.MustString()
		})
	})

	t.Run("Type conversion methods", func(t *testing.T) {
		nullNode := NewNullNode("", nil)
		assert.Equal(t, float64(0), nullNode.Float())
		assert.Equal(t, int64(0), nullNode.Int())
		assert.False(t, nullNode.Bool())
		assert.True(t, nullNode.Time().IsZero())

		// Test Must methods with panic
		assert.Panics(t, func() {
			nullNode.MustFloat()
		})
		assert.Panics(t, func() {
			nullNode.MustInt()
		})
		assert.Panics(t, func() {
			nullNode.MustBool()
		})
		assert.Panics(t, func() {
			nullNode.MustTime()
		})

		// Test with error
		invalidNode := NewInvalidNode("", errors.New("test error"))
		assert.Equal(t, float64(0), invalidNode.Float())
		assert.Equal(t, int64(0), invalidNode.Int())
		assert.False(t, invalidNode.Bool())
		assert.True(t, invalidNode.Time().IsZero())
	})

	t.Run("Interface method", func(t *testing.T) {
		nullNode := NewNullNode("", nil)
		assert.Nil(t, nullNode.Interface())

		// Test with error
		invalidNode := NewInvalidNode("", errors.New("test error"))
		assert.Nil(t, invalidNode.Interface())
	})

	t.Run("Contains and Strings methods", func(t *testing.T) {
		nullNode := NewNullNode("", nil)
		// These methods should be no-ops for null nodes
		assert.False(t, nullNode.Contains("anything"))
		assert.Nil(t, nullNode.Strings())
	})
}

func TestFunctionMethods(t *testing.T) {
	t.Run("Func, CallFunc, RemoveFunc methods", func(t *testing.T) {
		// Create a node with functions map
		funcs := make(map[string]func(Node) Node)
		objNode := NewObjectNode(map[string]Node{}, "", &funcs)

		// Register a function
		doubleFunc := func(n Node) Node {
			return NewNumberNode(n.Float()*2, "", &funcs)
		}
		result := objNode.Func("double", doubleFunc)
		assert.Equal(t, objNode, result) // Should return self

		// Call the function
		numNode := NewNumberNode(21, "", &funcs)
		result = numNode.CallFunc("double")
		assert.Equal(t, float64(42), result.Float())

		// Remove the function
		objNode.RemoveFunc("double")
		result = numNode.CallFunc("double")
		assert.False(t, result.IsValid())

		// Test with invalid node
		invalidNode := NewInvalidNode("", errors.New("test error"))
		invalidNode.Func("test", func(n Node) Node { return n })
		result = invalidNode.CallFunc("test")
		assert.False(t, result.IsValid())
		invalidNode.RemoveFunc("test")
	})

	t.Run("GetFuncs method", func(t *testing.T) {
		funcs := make(map[string]func(Node) Node)
		objNode := NewObjectNode(map[string]Node{}, "", &funcs)

		// Initially should have empty funcs
		retrievedFuncs := objNode.GetFuncs()
		assert.NotNil(t, retrievedFuncs)
		assert.Equal(t, 0, len(*retrievedFuncs))

		// Add a function
		objNode.Func("test", func(n Node) Node { return n })

		// Should now have one function
		retrievedFuncs = objNode.GetFuncs()
		assert.NotNil(t, retrievedFuncs)
		assert.Equal(t, 1, len(*retrievedFuncs))
		assert.Contains(t, *retrievedFuncs, "test")

		// Test with invalid node
		invalidNode := NewInvalidNode("", errors.New("test error"))
		assert.Nil(t, invalidNode.GetFuncs())
	})
}

func TestRawMethods(t *testing.T) {
	t.Run("Object node Raw method", func(t *testing.T) {
		objNode := NewObjectNode(map[string]Node{
			"key": NewStringNode("value", "", nil),
		}, "", nil)
		// We can't predict the exact JSON format due to key ordering, so just check it's not empty
		assert.NotEqual(t, "", objNode.Raw())

		// Test with error
		invalidNode := NewInvalidNode("", errors.New("test error"))
		assert.Equal(t, "", invalidNode.Raw())
	})

	t.Run("Array node Raw method", func(t *testing.T) {
		arrNode := NewArrayNode([]Node{
			NewStringNode("item1", "", nil),
			NewNumberNode(42, "", nil),
		}, "", nil)
		// Just check it's not empty
		assert.NotEqual(t, "", arrNode.Raw())

		// Test with error
		invalidNode := NewInvalidNode("", errors.New("test error"))
		assert.Equal(t, "", invalidNode.Raw())
	})

	t.Run("String node Raw method", func(t *testing.T) {
		strNode := NewStringNode("test value", "", nil)
		assert.Equal(t, `"test value"`, strNode.Raw())

		// Test with error
		invalidNode := NewInvalidNode("", errors.New("test error"))
		assert.Equal(t, "", invalidNode.Raw())
	})

	t.Run("Number node Raw method", func(t *testing.T) {
		numNode := NewNumberNode(3.14, "", nil)
		assert.Equal(t, "3.14", numNode.Raw())

		// Test with error
		invalidNode := NewInvalidNode("", errors.New("test error"))
		assert.Equal(t, "", invalidNode.Raw())
	})

	t.Run("Bool node Raw method", func(t *testing.T) {
		trueNode := NewBoolNode(true, "", nil)
		falseNode := NewBoolNode(false, "", nil)
		assert.Equal(t, "true", trueNode.Raw())
		assert.Equal(t, "false", falseNode.Raw())

		// Test with error
		invalidNode := NewInvalidNode("", errors.New("test error"))
		assert.Equal(t, "", invalidNode.Raw())
	})

	t.Run("Null node Raw method", func(t *testing.T) {
		nullNode := NewNullNode("", nil)
		assert.Equal(t, "null", nullNode.Raw())
	})
}

func TestNewNodeFromInterfaceErrors(t *testing.T) {
	t.Run("Unsupported type", func(t *testing.T) {
		// Test with channel (unsupported type)
		ch := make(chan int)
		_, err := NewNodeFromInterface(ch, "", nil)
		assert.Error(t, err)

		// Test with function (unsupported type)
		fn := func() {}
		_, err = NewNodeFromInterface(fn, "", nil)
		assert.Error(t, err)
	})

	t.Run("Error within nested map", func(t *testing.T) {
		ch := make(chan int)
		m := map[string]interface{}{"bad": ch}
		_, err := NewNodeFromInterface(m, "", nil)
		assert.Error(t, err)
	})

	t.Run("Error within nested slice", func(t *testing.T) {
		ch := make(chan int)
		s := []interface{}{ch}
		_, err := NewNodeFromInterface(s, "", nil)
		assert.Error(t, err)
	})
}

func TestMustMethodsAdditional(t *testing.T) {
	objNode := NewObjectNode(map[string]Node{}, "", nil)
	arrNode := NewArrayNode([]Node{}, "", nil)
	strNode := NewStringNode("hello", "", nil)
	numNode := NewNumberNode(123, "", nil)
	boolNode := NewBoolNode(true, "", nil)

	// objNode.MustString should not panic, it should return a JSON representation
	assert.NotPanics(t, func() { objNode.MustString() })
	assert.Equal(t, "hello", strNode.MustString())

	assert.Panics(t, func() { strNode.MustFloat() })
	assert.Equal(t, float64(123), numNode.MustFloat())

	assert.Panics(t, func() { strNode.MustInt() })
	assert.Equal(t, int64(123), numNode.MustInt())

	assert.Panics(t, func() { strNode.MustBool() })
	// Note: boolNode.MustBool() should return true, but strNode.MustBool() should panic
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

func TestAppendMethod(t *testing.T) {
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

func TestTimeStringMethods(t *testing.T) {
	timeStr := "2024-01-01T15:04:05Z"
	timeNode := NewStringNode(timeStr, "", nil)
	parsedTime, _ := time.Parse(time.RFC3339, timeStr)
	assert.Equal(t, parsedTime, timeNode.MustTime())

	// Error case
	badTimeNode := NewStringNode("not-a-time", "", nil)
	assert.True(t, badTimeNode.Time().IsZero())
	assert.Error(t, badTimeNode.Error())
}

func TestMiscCoverageAdditional(t *testing.T) {
	// Cover some zero-return cases for non-applicable types
	objNode := NewObjectNode(map[string]Node{}, "", nil)
	assert.Equal(t, int64(0), objNode.Int())
	assert.False(t, objNode.Bool())
	assert.True(t, objNode.Time().IsZero())

	numNode := NewNumberNode(1, "", nil)
	assert.False(t, numNode.Bool())
	assert.True(t, numNode.Time().IsZero())
}

func TestForEachOnDifferentNodeTypes(t *testing.T) {
	t.Run("Object node ForEach", func(t *testing.T) {
		objNode := NewObjectNode(map[string]Node{
			"key1": NewStringNode("value1", "", nil),
			"key2": NewNumberNode(42, "", nil),
		}, "", nil)

		count := 0
		keys := []string{}
		values := []string{}

		objNode.ForEach(func(keyOrIndex interface{}, value Node) {
			count++
			if key, ok := keyOrIndex.(string); ok {
				keys = append(keys, key)
			}
			values = append(values, value.String())
		})

		assert.Equal(t, 2, count)
		assert.ElementsMatch(t, []string{"key1", "key2"}, keys)
		assert.ElementsMatch(t, []string{"value1", "42"}, values)
	})

	t.Run("Array node ForEach", func(t *testing.T) {
		arrNode := NewArrayNode([]Node{
			NewStringNode("item1", "", nil),
			NewStringNode("item2", "", nil),
		}, "", nil)

		count := 0
		indices := []int{}
		values := []string{}

		arrNode.ForEach(func(keyOrIndex interface{}, value Node) {
			count++
			if idx, ok := keyOrIndex.(int); ok {
				indices = append(indices, idx)
			}
			values = append(values, value.String())
		})

		assert.Equal(t, 2, count)
		assert.ElementsMatch(t, []int{0, 1}, indices)
		assert.ElementsMatch(t, []string{"item1", "item2"}, values)
	})

	t.Run("Scalar nodes ForEach", func(t *testing.T) {
		// Test with scalar nodes (should not call the function)
		strNode := NewStringNode("test", "", nil)
		count := 0
		strNode.ForEach(func(keyOrIndex interface{}, value Node) {
			count++
		})
		assert.Equal(t, 0, count)

		numNode := NewNumberNode(42, "", nil)
		count = 0
		numNode.ForEach(func(keyOrIndex interface{}, value Node) {
			count++
		})
		assert.Equal(t, 0, count)
	})
}

func TestLenMethod(t *testing.T) {
	t.Run("Object node Len", func(t *testing.T) {
		objNode := NewObjectNode(map[string]Node{
			"key1": NewStringNode("value1", "", nil),
			"key2": NewNumberNode(42, "", nil),
		}, "", nil)
		assert.Equal(t, 2, objNode.Len())
	})

	t.Run("Array node Len", func(t *testing.T) {
		arrNode := NewArrayNode([]Node{
			NewStringNode("item1", "", nil),
			NewStringNode("item2", "", nil),
		}, "", nil)
		assert.Equal(t, 2, arrNode.Len())
	})

	t.Run("String node Len", func(t *testing.T) {
		strNode := NewStringNode("test", "", nil)
		assert.Equal(t, 4, strNode.Len())
	})

	t.Run("Other node types Len", func(t *testing.T) {
		numNode := NewNumberNode(42, "", nil)
		assert.Equal(t, 0, numNode.Len())

		boolNode := NewBoolNode(true, "", nil)
		assert.Equal(t, 0, boolNode.Len())

		nullNode := NewNullNode("", nil)
		assert.Equal(t, 0, nullNode.Len())
	})
}

func TestLowCoverageMethods(t *testing.T) {
	t.Run("Raw method coverage", func(t *testing.T) {
		// Test Raw method on different node types to improve coverage
		objNode := NewObjectNode(map[string]Node{
			"key": NewStringNode("value", "", nil),
		}, "", nil)
		raw := objNode.Raw()
		assert.NotEmpty(t, raw)
		assert.Contains(t, raw, `"key"`)
		assert.Contains(t, raw, `"value"`)

		arrNode := NewArrayNode([]Node{
			NewStringNode("item1", "", nil),
			NewNumberNode(42, "", nil),
		}, "", nil)
		raw = arrNode.Raw()
		assert.NotEmpty(t, raw)
		assert.Contains(t, raw, `"item1"`)
		assert.Contains(t, raw, `42`)

		// Test Raw on invalid node
		invalidNode := NewInvalidNode("", errors.New("test error"))
		raw = invalidNode.Raw()
		assert.Equal(t, "", raw)
	})

	t.Run("Interface method coverage", func(t *testing.T) {
		// Test Interface method on different node types
		objNode := NewObjectNode(map[string]Node{
			"key": NewStringNode("value", "", nil),
		}, "", nil)
		iface := objNode.Interface()
		assert.NotNil(t, iface)

		// Check that it's a map
		ifaceMap, ok := iface.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "value", ifaceMap["key"])

		// Test Interface on invalid node
		invalidNode := NewInvalidNode("", errors.New("test error"))
		iface = invalidNode.Interface()
		assert.Nil(t, iface)
	})

	t.Run("ForEach method coverage", func(t *testing.T) {
		// Test ForEach on array node
		arrNode := NewArrayNode([]Node{
			NewStringNode("item1", "", nil),
			NewStringNode("item2", "", nil),
		}, "", nil)

		count := 0
		arrNode.ForEach(func(keyOrIndex interface{}, value Node) {
			count++
			// keyOrIndex should be int for array nodes
			if idx, ok := keyOrIndex.(int); ok {
				assert.GreaterOrEqual(t, idx, 0)
				assert.Less(t, idx, 2)
			} else {
				t.Errorf("Expected int key, got %T", keyOrIndex)
			}
		})
		assert.Equal(t, 2, count)

		// Test ForEach on scalar nodes (should not be called)
		strNode := NewStringNode("test", "", nil)
		count = 0
		strNode.ForEach(func(keyOrIndex interface{}, value Node) {
			count++
		})
		assert.Equal(t, 0, count)
	})

	t.Run("Len method coverage", func(t *testing.T) {
		// Test Len on different node types
		objNode := NewObjectNode(map[string]Node{
			"key1": NewStringNode("value1", "", nil),
			"key2": NewStringNode("value2", "", nil),
		}, "", nil)
		assert.Equal(t, 2, objNode.Len())

		arrNode := NewArrayNode([]Node{
			NewStringNode("item1", "", nil),
			NewStringNode("item2", "", nil),
			NewStringNode("item3", "", nil),
		}, "", nil)
		assert.Equal(t, 3, arrNode.Len())

		strNode := NewStringNode("hello", "", nil)
		assert.Equal(t, 5, strNode.Len()) // Length of the string

		numNode := NewNumberNode(123, "", nil)
		assert.Equal(t, 0, numNode.Len()) // Numbers have length 0
	})
}

func TestFunctionRegistrationMethods(t *testing.T) {
	// Test Func, CallFunc, RemoveFunc methods on different node types
	t.Run("Func method", func(t *testing.T) {
		objNode := NewObjectNode(map[string]Node{}, "", nil)

		// Register a function
		result := objNode.Func("testFunc", func(n Node) Node {
			return NewStringNode("test", "", nil)
		})

		// Should return the same node
		assert.Equal(t, objNode, result)
	})

	t.Run("CallFunc method", func(t *testing.T) {
		funcs := make(map[string]func(Node) Node)
		objNode := NewObjectNode(map[string]Node{}, "", &funcs)

		// Register a function
		objNode.Func("double", func(n Node) Node {
			return NewNumberNode(n.Float()*2, "", &funcs)
		})

		// Call the function
		numNode := NewNumberNode(21, "", &funcs)
		result := numNode.CallFunc("double")
		assert.True(t, result.IsValid())
		assert.Equal(t, float64(42), result.Float())

		// Call non-existent function
		result = numNode.CallFunc("nonexistent")
		assert.False(t, result.IsValid())
	})

	t.Run("RemoveFunc method", func(t *testing.T) {
		funcs := make(map[string]func(Node) Node)
		objNode := NewObjectNode(map[string]Node{}, "", &funcs)

		// Register a function
		objNode.Func("testFunc", func(n Node) Node {
			return NewStringNode("test", "", &funcs)
		})

		// Remove the function
		objNode.RemoveFunc("testFunc")

		// Try to call the removed function
		result := objNode.CallFunc("testFunc")
		assert.False(t, result.IsValid())
	})

	t.Run("Function methods on invalid node", func(t *testing.T) {
		invalidNode := NewInvalidNode("", errors.New("test error"))

		// These should not panic and should return the node itself
		result := invalidNode.Func("test", func(n Node) Node { return n })
		assert.Equal(t, invalidNode, result)

		result = invalidNode.RemoveFunc("test")
		assert.Equal(t, invalidNode, result)

		result = invalidNode.CallFunc("test")
		assert.Equal(t, invalidNode, result)
	})
}

func TestArrayMethodOnDifferentNodes(t *testing.T) {
	t.Run("Null node Array", func(t *testing.T) {
		nullNode := NewNullNode("", nil)
		result := nullNode.Array()
		assert.Nil(t, result)
	})

	t.Run("Bool node Array", func(t *testing.T) {
		boolNode := NewBoolNode(true, "", nil)
		result := boolNode.Array()
		assert.Nil(t, result)
	})

	t.Run("Number node Array", func(t *testing.T) {
		numNode := NewNumberNode(42, "", nil)
		result := numNode.Array()
		assert.Nil(t, result)
	})
}

func TestZeroCoverageMethods(t *testing.T) {
	t.Run("Query method on array node", func(t *testing.T) {
		// Test Query method on arrayNode (which has 0% coverage)
		arrNode := NewArrayNode([]Node{
			NewStringNode("item1", "", nil),
			NewStringNode("item2", "", nil),
		}, "", nil)

		// Query for a valid index
		result := arrNode.Query("[0]")
		assert.True(t, result.IsValid())
		assert.Equal(t, "item1", result.String())

		// Query for an invalid index
		result = arrNode.Query("[5]")
		assert.False(t, result.IsValid())
	})

	t.Run("ForEach method on different node types", func(t *testing.T) {
		// Test ForEach on stringNode
		strNode := NewStringNode("test", "", nil)
		count := 0
		strNode.ForEach(func(keyOrIndex interface{}, value Node) {
			count++
		})
		assert.Equal(t, 0, count) // Should not be called for scalar nodes

		// Test ForEach on numberNode
		numNode := NewNumberNode(42, "", nil)
		count = 0
		numNode.ForEach(func(keyOrIndex interface{}, value Node) {
			count++
		})
		assert.Equal(t, 0, count) // Should not be called for scalar nodes

		// Test ForEach on boolNode
		boolNode := NewBoolNode(true, "", nil)
		count = 0
		boolNode.ForEach(func(keyOrIndex interface{}, value Node) {
			count++
		})
		assert.Equal(t, 0, count) // Should not be called for scalar nodes

		// Test ForEach on nullNode
		nullNode := NewNullNode("", nil)
		count = 0
		nullNode.ForEach(func(keyOrIndex interface{}, value Node) {
			count++
		})
		assert.Equal(t, 0, count) // Should not be called for scalar nodes
	})

	t.Run("Contains method on different node types", func(t *testing.T) {
		// Test Contains on stringNode
		strNode := NewStringNode("hello world", "", nil)
		assert.True(t, strNode.Contains("hello"))
		assert.True(t, strNode.Contains("world"))
		assert.False(t, strNode.Contains("universe"))

		// Test Contains on numberNode (should always return false)
		numNode := NewNumberNode(42, "", nil)
		assert.False(t, numNode.Contains("anything"))

		// Test Contains on boolNode (should always return false)
		boolNode := NewBoolNode(true, "", nil)
		assert.False(t, boolNode.Contains("anything"))

		// Test Contains on nullNode (should always return false)
		nullNode := NewNullNode("", nil)
		assert.False(t, nullNode.Contains("anything"))
	})

	t.Run("Filter and Map methods", func(t *testing.T) {
		// Test Filter method on arrayNode
		arrNode := NewArrayNode([]Node{
			NewNumberNode(1, "", nil),
			NewNumberNode(2, "", nil),
			NewNumberNode(3, "", nil),
			NewNumberNode(4, "", nil),
		}, "", nil)

		// Filter even numbers
		filtered := arrNode.Filter(func(n Node) bool {
			return int(n.Float())%2 == 0
		})
		assert.Equal(t, 2, filtered.Len())
		assert.Equal(t, float64(2), filtered.Index(0).Float())
		assert.Equal(t, float64(4), filtered.Index(1).Float())

		// Test Map method on arrayNode
		mapped := arrNode.Map(func(n Node) interface{} {
			return n.Float() * 2
		})
		assert.Equal(t, 4, mapped.Len())
		assert.Equal(t, float64(2), mapped.Index(0).Float())
		assert.Equal(t, float64(4), mapped.Index(1).Float())
		assert.Equal(t, float64(6), mapped.Index(2).Float())
		assert.Equal(t, float64(8), mapped.Index(3).Float())
	})
}

func TestForEachCoverage(t *testing.T) {
	t.Run("ForEach on objectNode", func(t *testing.T) {
		objNode := NewObjectNode(map[string]Node{
			"key1": NewStringNode("value1", "", nil),
			"key2": NewNumberNode(42, "", nil),
		}, "", nil)

		count := 0
		keys := []string{}
		values := []string{}

		objNode.ForEach(func(keyOrIndex interface{}, value Node) {
			count++
			if key, ok := keyOrIndex.(string); ok {
				keys = append(keys, key)
			}
			values = append(values, value.String())
		})

		assert.Equal(t, 2, count)
		assert.ElementsMatch(t, []string{"key1", "key2"}, keys)
		assert.ElementsMatch(t, []string{"value1", "42"}, values)
	})

	t.Run("ForEach on arrayNode", func(t *testing.T) {
		arrNode := NewArrayNode([]Node{
			NewStringNode("item1", "", nil),
			NewStringNode("item2", "", nil),
		}, "", nil)

		count := 0
		indices := []int{}
		values := []string{}

		arrNode.ForEach(func(keyOrIndex interface{}, value Node) {
			count++
			if idx, ok := keyOrIndex.(int); ok {
				indices = append(indices, idx)
			}
			values = append(values, value.String())
		})

		assert.Equal(t, 2, count)
		assert.ElementsMatch(t, []int{0, 1}, indices)
		assert.ElementsMatch(t, []string{"item1", "item2"}, values)
	})

	t.Run("ForEach on scalar nodes", func(t *testing.T) {
		// Test ForEach on stringNode (should not be called)
		strNode := NewStringNode("test", "", nil)
		count := 0
		strNode.ForEach(func(keyOrIndex interface{}, value Node) {
			count++
		})
		assert.Equal(t, 0, count)

		// Test ForEach on numberNode (should not be called)
		numNode := NewNumberNode(42, "", nil)
		count = 0
		numNode.ForEach(func(keyOrIndex interface{}, value Node) {
			count++
		})
		assert.Equal(t, 0, count)

		// Test ForEach on boolNode (should not be called)
		boolNode := NewBoolNode(true, "", nil)
		count = 0
		boolNode.ForEach(func(keyOrIndex interface{}, value Node) {
			count++
		})
		assert.Equal(t, 0, count)

		// Test ForEach on nullNode (should not be called)
		nullNode := NewNullNode("", nil)
		count = 0
		nullNode.ForEach(func(keyOrIndex interface{}, value Node) {
			count++
		})
		assert.Equal(t, 0, count)
	})
}

func TestFuncCallRemoveMethods(t *testing.T) {
	t.Run("Func, CallFunc, RemoveFunc on objectNode", func(t *testing.T) {
		funcs := make(map[string]func(Node) Node)
		objNode := NewObjectNode(map[string]Node{}, "", &funcs)

		// Register a function
		result := objNode.Func("testFunc", func(n Node) Node {
			return NewStringNode("test", "", &funcs)
		})
		assert.Equal(t, objNode, result)

		// Call the function
		called := objNode.CallFunc("testFunc")
		assert.True(t, called.IsValid())
		assert.Equal(t, "test", called.String())

		// Remove the function
		result = objNode.RemoveFunc("testFunc")
		assert.Equal(t, objNode, result)

		// Try to call the removed function
		called = objNode.CallFunc("testFunc")
		assert.False(t, called.IsValid())
	})

	t.Run("Func, CallFunc, RemoveFunc on arrayNode", func(t *testing.T) {
		funcs := make(map[string]func(Node) Node)
		arrNode := NewArrayNode([]Node{}, "", &funcs)

		// Register a function
		result := arrNode.Func("double", func(n Node) Node {
			return NewNumberNode(n.Float()*2, "", &funcs)
		})
		assert.Equal(t, arrNode, result)

		// Call the function
		numNode := NewNumberNode(21, "", &funcs)
		called := numNode.CallFunc("double")
		assert.True(t, called.IsValid())
		assert.Equal(t, float64(42), called.Float())

		// Remove the function
		result = arrNode.RemoveFunc("double")
		assert.Equal(t, arrNode, result)
	})

	t.Run("Func, CallFunc, RemoveFunc on scalar nodes", func(t *testing.T) {
		funcs := make(map[string]func(Node) Node)
		strNode := NewStringNode("test", "", &funcs)

		// These should work since we passed a funcs map
		result := strNode.Func("testFunc", func(n Node) Node { return n })
		assert.Equal(t, strNode, result)

		result = strNode.RemoveFunc("testFunc")
		assert.Equal(t, strNode, result)

		result = strNode.CallFunc("testFunc")
		// This should return an invalid node since the function doesn't exist
		assert.NotEqual(t, strNode, result)
		assert.False(t, result.IsValid())
	})
}

func TestRemainingZeroCoverageMethods(t *testing.T) {
	t.Run("Contains method on arrayNode", func(t *testing.T) {
		// Test Contains method on arrayNode (currently 0% coverage)
		arrNode := NewArrayNode([]Node{
			NewStringNode("item1", "", nil),
			NewStringNode("item2", "", nil),
		}, "", nil)

		// Contains should work on array nodes
		assert.True(t, arrNode.Contains("item1"))
		assert.True(t, arrNode.Contains("item2"))
		assert.False(t, arrNode.Contains("item3"))
	})

	t.Run("AsMap methods", func(t *testing.T) {
		// These methods are only available on internal node types, not the public interface
		// We can't directly test them through the public interface, but we can test
		// the behavior through other methods that use them internally
		objNode := NewObjectNode(map[string]Node{
			"key1": NewStringNode("value1", "", nil),
			"key2": NewNumberNode(42, "", nil),
		}, "", nil)

		// Test that we can access the values as a map through other methods
		assert.Equal(t, "value1", objNode.Get("key1").String())
		assert.Equal(t, float64(42), objNode.Get("key2").Float())
	})

	t.Run("CallFunc on various nodes", func(t *testing.T) {
		// Test CallFunc on stringNode with no function map
		strNode := NewStringNode("test", "", nil)
		result := strNode.CallFunc("nonexistent")
		assert.False(t, result.IsValid())

		// Test CallFunc on numberNode with no function map
		numNode := NewNumberNode(42, "", nil)
		result = numNode.CallFunc("nonexistent")
		assert.False(t, result.IsValid())
	})

	t.Run("Func and RemoveFunc on various nodes", func(t *testing.T) {
		// Test Func and RemoveFunc on boolNode
		funcs := make(map[string]func(Node) Node)
		boolNode := NewBoolNode(true, "", &funcs)

		result := boolNode.Func("testFunc", func(n Node) Node {
			return NewStringNode("test", "", &funcs)
		})
		assert.Equal(t, boolNode, result)

		result = boolNode.RemoveFunc("testFunc")
		assert.Equal(t, boolNode, result)

		// Test Func and RemoveFunc on nullNode
		nullNode := NewNullNode("", &funcs)

		result = nullNode.Func("testFunc", func(n Node) Node {
			return NewStringNode("test", "", &funcs)
		})
		assert.Equal(t, nullNode, result)

		result = nullNode.RemoveFunc("testFunc")
		assert.Equal(t, nullNode, result)
	})
}

func TestZeroCoverageMethodsPart2(t *testing.T) {
	t.Run("ForEach on different node types", func(t *testing.T) {
		// Test ForEach on objectNode
		objNode := NewObjectNode(map[string]Node{
			"key1": NewStringNode("value1", "", nil),
			"key2": NewNumberNode(42, "", nil),
		}, "", nil)

		count := 0
		objNode.ForEach(func(keyOrIndex interface{}, value Node) {
			count++
		})
		assert.Equal(t, 2, count)

		// Test ForEach on arrayNode
		arrNode := NewArrayNode([]Node{
			NewStringNode("item1", "", nil),
			NewStringNode("item2", "", nil),
		}, "", nil)

		count = 0
		arrNode.ForEach(func(keyOrIndex interface{}, value Node) {
			count++
		})
		assert.Equal(t, 2, count)

		// Test ForEach on scalar nodes (should not be called)
		strNode := NewStringNode("test", "", nil)
		count = 0
		strNode.ForEach(func(keyOrIndex interface{}, value Node) {
			count++
		})
		assert.Equal(t, 0, count)

		numNode := NewNumberNode(42, "", nil)
		count = 0
		numNode.ForEach(func(keyOrIndex interface{}, value Node) {
			count++
		})
		assert.Equal(t, 0, count)
	})

	t.Run("Contains method on arrayNode", func(t *testing.T) {
		// Test Contains method on arrayNode
		arrNode := NewArrayNode([]Node{
			NewStringNode("item1", "", nil),
			NewStringNode("item2", "", nil),
		}, "", nil)

		assert.True(t, arrNode.Contains("item1"))
		assert.True(t, arrNode.Contains("item2"))
		assert.False(t, arrNode.Contains("item3"))
	})

	t.Run("Func, CallFunc, RemoveFunc on various nodes", func(t *testing.T) {
		// Test on boolNode
		funcs := make(map[string]func(Node) Node)
		boolNode := NewBoolNode(true, "", &funcs)

		result := boolNode.Func("testFunc", func(n Node) Node {
			return NewStringNode("test", "", &funcs)
		})
		assert.Equal(t, boolNode, result)

		result = boolNode.RemoveFunc("testFunc")
		assert.Equal(t, boolNode, result)

		// Test on nullNode
		nullNode := NewNullNode("", &funcs)

		result = nullNode.Func("testFunc", func(n Node) Node {
			return NewStringNode("test", "", &funcs)
		})
		assert.Equal(t, nullNode, result)

		result = nullNode.RemoveFunc("testFunc")
		assert.Equal(t, nullNode, result)
	})
}

// Add specific tests for methods that are at 0% to make sure we cover all code paths
func TestSpecificZeroCoverageMethods(t *testing.T) {
	t.Run("CallFunc on various node types with no funcs map", func(t *testing.T) {
		// Test CallFunc on stringNode with no function map
		strNode := NewStringNode("test", "", nil)
		result := strNode.CallFunc("nonexistent")
		assert.False(t, result.IsValid())

		// Test CallFunc on numberNode with no function map
		numNode := NewNumberNode(42, "", nil)
		result = numNode.CallFunc("nonexistent")
		assert.False(t, result.IsValid())

		// Test CallFunc on boolNode with no function map
		boolNode := NewBoolNode(true, "", nil)
		result = boolNode.CallFunc("nonexistent")
		assert.False(t, result.IsValid())

		// Test CallFunc on nullNode with no function map
		nullNode := NewNullNode("", nil)
		result = nullNode.CallFunc("nonexistent")
		assert.False(t, result.IsValid())
	})

	t.Run("AsMap methods", func(t *testing.T) {
		// These methods are internal and not exposed in the public interface
		// We can't directly test them but we can test behavior that uses them internally

		// Test objectNode AsMap behavior through other methods
		objNode := NewObjectNode(map[string]Node{
			"key1": NewStringNode("value1", "", nil),
			"key2": NewNumberNode(42, "", nil),
		}, "", nil)

		// Verify we can access the data
		assert.Equal(t, "value1", objNode.Get("key1").String())
		assert.Equal(t, float64(42), objNode.Get("key2").Float())
	})
}

func TestForEachOnAllNodeTypes(t *testing.T) {
	// Test ForEach on all node types to improve coverage

	t.Run("ForEach on objectNode", func(t *testing.T) {
		objNode := NewObjectNode(map[string]Node{
			"key1": NewStringNode("value1", "", nil),
			"key2": NewNumberNode(42, "", nil),
		}, "", nil)

		count := 0
		keys := make([]string, 0)
		values := make([]string, 0)

		objNode.ForEach(func(keyOrIndex interface{}, value Node) {
			count++
			if key, ok := keyOrIndex.(string); ok {
				keys = append(keys, key)
			}
			values = append(values, value.String())
		})

		assert.Equal(t, 2, count)
		assert.ElementsMatch(t, []string{"key1", "key2"}, keys)
		assert.ElementsMatch(t, []string{"value1", "42"}, values)
	})

	t.Run("ForEach on arrayNode", func(t *testing.T) {
		arrNode := NewArrayNode([]Node{
			NewStringNode("item1", "", nil),
			NewStringNode("item2", "", nil),
		}, "", nil)

		count := 0
		indices := make([]int, 0)
		values := make([]string, 0)

		arrNode.ForEach(func(keyOrIndex interface{}, value Node) {
			count++
			if idx, ok := keyOrIndex.(int); ok {
				indices = append(indices, idx)
			}
			values = append(values, value.String())
		})

		assert.Equal(t, 2, count)
		assert.ElementsMatch(t, []int{0, 1}, indices)
		assert.ElementsMatch(t, []string{"item1", "item2"}, values)
	})

	t.Run("ForEach on stringNode", func(t *testing.T) {
		strNode := NewStringNode("test", "", nil)
		count := 0

		// ForEach should not be called on scalar nodes
		strNode.ForEach(func(keyOrIndex interface{}, value Node) {
			count++
		})

		assert.Equal(t, 0, count)
	})

	t.Run("ForEach on numberNode", func(t *testing.T) {
		numNode := NewNumberNode(42, "", nil)
		count := 0

		// ForEach should not be called on scalar nodes
		numNode.ForEach(func(keyOrIndex interface{}, value Node) {
			count++
		})

		assert.Equal(t, 0, count)
	})

	t.Run("ForEach on boolNode", func(t *testing.T) {
		boolNode := NewBoolNode(true, "", nil)
		count := 0

		// ForEach should not be called on scalar nodes
		boolNode.ForEach(func(keyOrIndex interface{}, value Node) {
			count++
		})

		assert.Equal(t, 0, count)
	})

	t.Run("ForEach on nullNode", func(t *testing.T) {
		nullNode := NewNullNode("", nil)
		count := 0

		// ForEach should not be called on scalar nodes
		nullNode.ForEach(func(keyOrIndex interface{}, value Node) {
			count++
		})

		assert.Equal(t, 0, count)
	})
}

func TestCallFuncOnAllNodeTypes(t *testing.T) {
	t.Run("CallFunc on stringNode without funcs map", func(t *testing.T) {
		strNode := NewStringNode("test", "", nil)
		result := strNode.CallFunc("nonexistent")
		assert.False(t, result.IsValid())
	})

	t.Run("CallFunc on numberNode without funcs map", func(t *testing.T) {
		numNode := NewNumberNode(42, "", nil)
		result := numNode.CallFunc("nonexistent")
		assert.False(t, result.IsValid())
	})

	t.Run("CallFunc on boolNode without funcs map", func(t *testing.T) {
		boolNode := NewBoolNode(true, "", nil)
		result := boolNode.CallFunc("nonexistent")
		assert.False(t, result.IsValid())
	})

	t.Run("CallFunc on nullNode without funcs map", func(t *testing.T) {
		nullNode := NewNullNode("", nil)
		result := nullNode.CallFunc("nonexistent")
		assert.False(t, result.IsValid())
	})
}

func TestNodeSpecificCallFuncMethods(t *testing.T) {
	t.Run("CallFunc on arrayNode with valid function", func(t *testing.T) {
		funcs := make(map[string]func(Node) Node)
		arrNode := NewArrayNode([]Node{
			NewStringNode("item1", "", nil),
			NewStringNode("item2", "", nil),
		}, "", &funcs)

		// Register a function
		arrNode.Func("count", func(n Node) Node {
			return NewNumberNode(float64(n.Len()), "", &funcs)
		})

		// Call the function
		result := arrNode.CallFunc("count")
		assert.True(t, result.IsValid())
		assert.Equal(t, float64(2), result.Float())
	})

	t.Run("CallFunc on arrayNode with invalid function", func(t *testing.T) {
		funcs := make(map[string]func(Node) Node)
		arrNode := NewArrayNode([]Node{
			NewStringNode("item1", "", nil),
			NewStringNode("item2", "", nil),
		}, "", &funcs)

		// Call a non-existent function
		result := arrNode.CallFunc("nonexistent")
		assert.False(t, result.IsValid())
	})

	t.Run("CallFunc on stringNode with valid function", func(t *testing.T) {
		funcs := make(map[string]func(Node) Node)
		strNode := NewStringNode("test", "", &funcs)

		// Register a function
		strNode.Func("length", func(n Node) Node {
			return NewNumberNode(float64(len(n.String())), "", &funcs)
		})

		// Call the function
		result := strNode.CallFunc("length")
		assert.True(t, result.IsValid())
		assert.Equal(t, float64(4), result.Float())
	})

	t.Run("CallFunc on stringNode with invalid function", func(t *testing.T) {
		funcs := make(map[string]func(Node) Node)
		strNode := NewStringNode("test", "", &funcs)

		// Call a non-existent function
		result := strNode.CallFunc("nonexistent")
		assert.False(t, result.IsValid())
	})

	t.Run("CallFunc on numberNode with valid function", func(t *testing.T) {
		funcs := make(map[string]func(Node) Node)
		numNode := NewNumberNode(5, "", &funcs)

		// Register a function
		numNode.Func("square", func(n Node) Node {
			val := n.Float()
			return NewNumberNode(val*val, "", &funcs)
		})

		// Call the function
		result := numNode.CallFunc("square")
		assert.True(t, result.IsValid())
		assert.Equal(t, float64(25), result.Float())
	})

	t.Run("CallFunc on numberNode with invalid function", func(t *testing.T) {
		funcs := make(map[string]func(Node) Node)
		numNode := NewNumberNode(5, "", &funcs)

		// Call a non-existent function
		result := numNode.CallFunc("nonexistent")
		assert.False(t, result.IsValid())
	})
}

func TestRemainingForEachMethods(t *testing.T) {
	t.Run("ForEach on objectNode specifically", func(t *testing.T) {
		objNode := NewObjectNode(map[string]Node{
			"name": NewStringNode("John", "", nil),
			"age":  NewNumberNode(30, "", nil),
		}, "", nil)

		var keys []string
		var values []string
		count := 0

		objNode.ForEach(func(keyOrIndex interface{}, value Node) {
			count++
			if key, ok := keyOrIndex.(string); ok {
				keys = append(keys, key)
			}
			values = append(values, value.String())
		})

		assert.Equal(t, 2, count)
		assert.ElementsMatch(t, []string{"name", "age"}, keys)
		assert.ElementsMatch(t, []string{"John", "30"}, values)
	})

	t.Run("ForEach on arrayNode specifically", func(t *testing.T) {
		arrNode := NewArrayNode([]Node{
			NewStringNode("first", "", nil),
			NewStringNode("second", "", nil),
		}, "", nil)

		var indices []int
		var values []string
		count := 0

		arrNode.ForEach(func(keyOrIndex interface{}, value Node) {
			count++
			if idx, ok := keyOrIndex.(int); ok {
				indices = append(indices, idx)
			}
			values = append(values, value.String())
		})

		assert.Equal(t, 2, count)
		assert.ElementsMatch(t, []int{0, 1}, indices)
		assert.ElementsMatch(t, []string{"first", "second"}, values)
	})

	t.Run("ForEach on stringNode specifically", func(t *testing.T) {
		strNode := NewStringNode("hello", "", nil)
		count := 0

		strNode.ForEach(func(keyOrIndex interface{}, value Node) {
			count++
		})

		assert.Equal(t, 0, count)
	})

	t.Run("ForEach on numberNode specifically", func(t *testing.T) {
		numNode := NewNumberNode(42, "", nil)
		count := 0

		numNode.ForEach(func(keyOrIndex interface{}, value Node) {
			count++
		})

		assert.Equal(t, 0, count)
	})

	t.Run("ForEach on boolNode specifically", func(t *testing.T) {
		boolNode := NewBoolNode(true, "", nil)
		count := 0

		boolNode.ForEach(func(keyOrIndex interface{}, value Node) {
			count++
		})

		assert.Equal(t, 0, count)
	})

	t.Run("ForEach on nullNode specifically", func(t *testing.T) {
		nullNode := NewNullNode("", nil)
		count := 0

		nullNode.ForEach(func(keyOrIndex interface{}, value Node) {
			count++
		})

		assert.Equal(t, 0, count)
	})
}

func TestZeroCoverageMethodsFinal(t *testing.T) {
	t.Run("AsMap and MustAsMap methods", func(t *testing.T) {
		// These are internal methods not exposed in the public interface
		// We can't directly call them, but we can test behavior that uses them internally

		// Test objectNode AsMap behavior through other methods
		objNode := NewObjectNode(map[string]Node{
			"key1": NewStringNode("value1", "", nil),
			"key2": NewNumberNode(42, "", nil),
		}, "", nil)

		// Verify we can access the data (this indirectly uses AsMap)
		assert.Equal(t, "value1", objNode.Get("key1").String())
		assert.Equal(t, float64(42), objNode.Get("key2").Float())

		// Test that non-object nodes return nil for map-like operations
		strNode := NewStringNode("test", "", nil)
		assert.Nil(t, strNode.Array())

		numNode := NewNumberNode(123, "", nil)
		assert.Nil(t, numNode.Array())

		boolNode := NewBoolNode(true, "", nil)
		assert.Nil(t, boolNode.Array())

		nullNode := NewNullNode("", nil)
		assert.Nil(t, nullNode.Array())
	})

	t.Run("RemoveFunc on various node types", func(t *testing.T) {
		// Test RemoveFunc on numberNode
		funcs := make(map[string]func(Node) Node)
		numNode := NewNumberNode(42, "", &funcs)

		// Add a function
		result := numNode.Func("square", func(n Node) Node {
			return NewNumberNode(n.Float()*n.Float(), "", &funcs)
		})
		assert.Equal(t, numNode, result)

		// Remove the function
		result = numNode.RemoveFunc("square")
		assert.Equal(t, numNode, result)

		// Try to call the removed function
		called := numNode.CallFunc("square")
		assert.False(t, called.IsValid())

		// Test RemoveFunc on boolNode
		boolFuncs := make(map[string]func(Node) Node)
		boolNode := NewBoolNode(true, "", &boolFuncs)

		result = boolNode.Func("invert", func(n Node) Node {
			return NewBoolNode(!n.Bool(), "", &boolFuncs)
		})
		assert.Equal(t, boolNode, result)

		result = boolNode.RemoveFunc("invert")
		assert.Equal(t, boolNode, result)

		// Test RemoveFunc on nullNode
		nullFuncs := make(map[string]func(Node) Node)
		nullNode := NewNullNode("", &nullFuncs)

		result = nullNode.Func("alwaysTrue", func(n Node) Node {
			return NewBoolNode(true, "", &nullFuncs)
		})
		assert.Equal(t, nullNode, result)

		result = nullNode.RemoveFunc("alwaysTrue")
		assert.Equal(t, nullNode, result)
	})

	t.Run("ForEach on all scalar node types", func(t *testing.T) {
		// Test ForEach on all scalar node types to ensure we hit those code paths

		// stringNode ForEach
		strNode := NewStringNode("test", "", nil)
		count := 0
		strNode.ForEach(func(keyOrIndex interface{}, value Node) {
			count++
		})
		assert.Equal(t, 0, count)

		// numberNode ForEach
		numNode := NewNumberNode(123, "", nil)
		count = 0
		numNode.ForEach(func(keyOrIndex interface{}, value Node) {
			count++
		})
		assert.Equal(t, 0, count)

		// boolNode ForEach
		boolNode := NewBoolNode(true, "", nil)
		count = 0
		boolNode.ForEach(func(keyOrIndex interface{}, value Node) {
			count++
		})
		assert.Equal(t, 0, count)

		// nullNode ForEach
		nullNode := NewNullNode("", nil)
		count = 0
		nullNode.ForEach(func(keyOrIndex interface{}, value Node) {
			count++
		})
		assert.Equal(t, 0, count)
	})
}

func TestAllNodeTypesForEach(t *testing.T) {
	t.Run("ForEach on objectNode", func(t *testing.T) {
		objNode := NewObjectNode(map[string]Node{
			"name": NewStringNode("John", "", nil),
			"age":  NewNumberNode(30, "", nil),
		}, "", nil)

		count := 0
		objNode.ForEach(func(keyOrIndex interface{}, value Node) {
			count++
			// Verify that key is a string for object nodes
			_, ok := keyOrIndex.(string)
			assert.True(t, ok)
		})
		assert.Equal(t, 2, count)
	})

	t.Run("ForEach on arrayNode", func(t *testing.T) {
		arrNode := NewArrayNode([]Node{
			NewStringNode("first", "", nil),
			NewStringNode("second", "", nil),
		}, "", nil)

		count := 0
		arrNode.ForEach(func(keyOrIndex interface{}, value Node) {
			count++
			// Verify that key is an int for array nodes
			_, ok := keyOrIndex.(int)
			assert.True(t, ok)
		})
		assert.Equal(t, 2, count)
	})

	t.Run("ForEach on stringNode", func(t *testing.T) {
		strNode := NewStringNode("hello", "", nil)
		count := 0
		strNode.ForEach(func(keyOrIndex interface{}, value Node) {
			count++
		})
		assert.Equal(t, 0, count)
	})

	t.Run("ForEach on numberNode", func(t *testing.T) {
		numNode := NewNumberNode(42, "", nil)
		count := 0
		numNode.ForEach(func(keyOrIndex interface{}, value Node) {
			count++
		})
		assert.Equal(t, 0, count)
	})

	t.Run("ForEach on boolNode", func(t *testing.T) {
		boolNode := NewBoolNode(true, "", nil)
		count := 0
		boolNode.ForEach(func(keyOrIndex interface{}, value Node) {
			count++
		})
		assert.Equal(t, 0, count)
	})

	t.Run("ForEach on nullNode", func(t *testing.T) {
		nullNode := NewNullNode("", nil)
		count := 0
		nullNode.ForEach(func(keyOrIndex interface{}, value Node) {
			count++
		})
		assert.Equal(t, 0, count)
	})

	t.Run("ForEach on invalidNode", func(t *testing.T) {
		invalidNode := NewInvalidNode("", errors.New("test error"))
		count := 0
		invalidNode.ForEach(func(keyOrIndex interface{}, value Node) {
			count++
		})
		assert.Equal(t, 0, count)
	})
}
