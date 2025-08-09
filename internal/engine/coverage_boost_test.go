package engine

import (
	"errors"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/474420502/xjson/internal/core"
)

func assertPanics(t *testing.T, f func()) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	f()
}

func TestStringsAndContainsMethods(t *testing.T) {
	t.Run("Strings method on arrayNode", func(t *testing.T) {
		arrNode := NewArrayNode([]core.Node{
			NewStringNode("apple", "", nil),
			NewStringNode("banana", "", nil),
			NewStringNode("cherry", "", nil),
		}, "", nil)

		strings := arrNode.Strings()
		expected := []string{"apple", "banana", "cherry"}
		sort.Strings(strings)
		sort.Strings(expected)
		if !reflect.DeepEqual(expected, strings) {
			t.Errorf("expected %v, got %v", expected, strings)
		}

		// Test empty array
		emptyArrNode := NewArrayNode([]core.Node{}, "", nil)
		if len(emptyArrNode.Strings()) != 0 {
			t.Errorf("expected empty slice, got %v", emptyArrNode.Strings())
		}

		// Test array with non-string elements
		mixedArrNode := NewArrayNode([]core.Node{
			NewStringNode("apple", "", nil),
			NewNumberNode(42, "", nil),
			NewBoolNode(true, "", nil),
		}, "", nil)
		// When array contains non-string elements, Strings should return nil and set error
		strings = mixedArrNode.Strings()
		if strings != nil {
			t.Errorf("expected nil, got %v", strings)
		}
		if mixedArrNode.Error() == nil {
			t.Errorf("expected node to have an error")
		}
	})

	t.Run("Strings method on stringNode", func(t *testing.T) {
		strNode := NewStringNode("hello", "", nil)
		if !reflect.DeepEqual([]string{"hello"}, strNode.Strings()) {
			t.Errorf("expected %v, got %v", []string{"hello"}, strNode.Strings())
		}

		// Test empty string
		emptyStrNode := NewStringNode("", "", nil)
		if !reflect.DeepEqual([]string{""}, emptyStrNode.Strings()) {
			t.Errorf("expected %v, got %v", []string{""}, emptyStrNode.Strings())
		}
	})

	t.Run("Strings method on other node types", func(t *testing.T) {
		numNode := NewNumberNode(123, "", nil)
		if numNode.Strings() != nil {
			t.Errorf("expected nil, got %v", numNode.Strings())
		}

		boolNode := NewBoolNode(true, "", nil)
		if boolNode.Strings() != nil {
			t.Errorf("expected nil, got %v", boolNode.Strings())
		}

		nullNode := NewNullNode("", nil)
		if nullNode.Strings() != nil {
			t.Errorf("expected nil, got %v", nullNode.Strings())
		}
	})

	t.Run("Contains method on stringNode", func(t *testing.T) {
		strNode := NewStringNode("hello world", "", nil)
		if !strNode.Contains("hello") {
			t.Errorf("expected true for Contains 'hello'")
		}
		if !strNode.Contains("world") {
			t.Errorf("expected true for Contains 'world'")
		}
		if strNode.Contains("universe") {
			t.Errorf("expected false for Contains 'universe'")
		}

		// Test empty substring
		if !strNode.Contains("") {
			t.Errorf("expected true for Contains ''")
		}
		// Test case sensitivity
		if strNode.Contains("HELLO") {
			t.Errorf("expected false for Contains 'HELLO'")
		}
	})

	t.Run("Contains method on arrayNode", func(t *testing.T) {
		arrNode := NewArrayNode([]core.Node{
			NewStringNode("apple", "", nil),
			NewStringNode("banana", "", nil),
			NewStringNode("cherry", "", nil),
		}, "", nil)

		if !arrNode.Contains("apple") {
			t.Errorf("expected true for Contains 'apple'")
		}
		if !arrNode.Contains("banana") {
			t.Errorf("expected true for Contains 'banana'")
		}
		if arrNode.Contains("orange") {
			t.Errorf("expected false for Contains 'orange'")
		}

		// Test with non-string elements
		mixedArrNode := NewArrayNode([]core.Node{
			NewStringNode("apple", "", nil),
			NewNumberNode(42, "", nil),
			NewBoolNode(true, "", nil),
		}, "", nil)
		if !mixedArrNode.Contains("apple") {
			t.Errorf("expected true for Contains 'apple'")
		}
		if mixedArrNode.Contains("42") {
			t.Errorf("expected false for Contains '42'")
		}
		if mixedArrNode.Contains("true") {
			t.Errorf("expected false for Contains 'true'")
		}
	})

	t.Run("Contains method on other node types", func(t *testing.T) {
		numNode := NewNumberNode(123, "", nil)
		if numNode.Contains("anything") {
			t.Errorf("expected false for Contains on number node")
		}

		boolNode := NewBoolNode(true, "", nil)
		if boolNode.Contains("anything") {
			t.Errorf("expected false for Contains on bool node")
		}

		nullNode := NewNullNode("", nil)
		if nullNode.Contains("anything") {
			t.Errorf("expected false for Contains on null node")
		}
	})
}

func TestNodeWrapperMethods(t *testing.T) {
	// Test ForEach method on different node types
	t.Run("ForEach method", func(t *testing.T) {
		// Test with object node
		objNode := NewObjectNode(map[string]core.Node{
			"key1": NewStringNode("value1", "", nil),
			"key2": NewStringNode("value2", "", nil),
		}, "", nil)

		count := 0
		keys := []string{}
		values := []string{}

		objNode.ForEach(func(keyOrIndex interface{}, value core.Node) {
			count++
			if key, ok := keyOrIndex.(string); ok {
				keys = append(keys, key)
			}
			values = append(values, value.String())
		})

		if count != 2 {
			t.Errorf("expected count 2, got %d", count)
		}

		expectedKeys := []string{"key1", "key2"}
		sort.Strings(keys)
		sort.Strings(expectedKeys)
		if !reflect.DeepEqual(expectedKeys, keys) {
			t.Errorf("expected keys %v, got %v", expectedKeys, keys)
		}

		expectedValues := []string{"value1", "value2"}
		sort.Strings(values)
		sort.Strings(expectedValues)
		if !reflect.DeepEqual(expectedValues, values) {
			t.Errorf("expected values %v, got %v", expectedValues, values)
		}

		// Test with array node
		arrNode := NewArrayNode([]core.Node{
			NewStringNode("item1", "", nil),
			NewStringNode("item2", "", nil),
		}, "", nil)

		count = 0
		var indices []int
		values = []string{}

		arrNode.ForEach(func(keyOrIndex interface{}, value core.Node) {
			count++
			if idx, ok := keyOrIndex.(int); ok {
				indices = append(indices, idx)
			}
			values = append(values, value.String())
		})

		if count != 2 {
			t.Errorf("expected count 2, got %d", count)
		}

		expectedIndices := []int{0, 1}
		sort.Ints(indices)
		if !reflect.DeepEqual(expectedIndices, indices) {
			t.Errorf("expected indices %v, got %v", expectedIndices, indices)
		}
		expectedArrValues := []string{"item1", "item2"}
		sort.Strings(values)
		if !reflect.DeepEqual(expectedArrValues, values) {
			t.Errorf("expected values %v, got %v", expectedArrValues, values)
		}

		// Test with scalar nodes (should not call the function)
		strNode := NewStringNode("test", "", nil)
		count = 0
		strNode.ForEach(func(keyOrIndex interface{}, value core.Node) {
			count++
		})
		if count != 0 {
			t.Errorf("expected count 0, got %d", count)
		}
	})

	// Test Len method on different node types
	t.Run("Len method", func(t *testing.T) {
		// Test with object node - this should return the number of keys
		objNode := NewObjectNode(map[string]core.Node{
			"key1": NewStringNode("value1", "", nil),
			"key2": NewStringNode("value2", "", nil),
		}, "", nil)
		if objNode.Len() != 2 {
			t.Errorf("expected 2, got %d", objNode.Len())
		}

		// Test with array node
		arrNode := NewArrayNode([]core.Node{
			NewStringNode("item1", "", nil),
			NewStringNode("item2", "", nil),
			NewStringNode("item3", "", nil),
		}, "", nil)
		if arrNode.Len() != 3 {
			t.Errorf("expected 3, got %d", arrNode.Len())
		}

		// Test with scalar nodes
		strNode := NewStringNode("test", "", nil)
		// String nodes return the length of their string value
		if strNode.Len() != 4 {
			t.Errorf("expected 4, got %d", strNode.Len())
		}

		numNode := NewNumberNode(42, "", nil)
		// Number nodes return 0 for Len()
		if numNode.Len() != 0 {
			t.Errorf("expected 0, got %d", numNode.Len())
		}
	})

	// Test Type method on different node types
	t.Run("Type method", func(t *testing.T) {
		objNode := NewObjectNode(map[string]core.Node{}, "", nil)
		if objNode.Type() != core.ObjectNode {
			t.Errorf("expected %v, got %v", core.ObjectNode, objNode.Type())
		}

		arrNode := NewArrayNode([]core.Node{}, "", nil)
		if arrNode.Type() != core.ArrayNode {
			t.Errorf("expected %v, got %v", core.ArrayNode, arrNode.Type())
		}

		strNode := NewStringNode("test", "", nil)
		if strNode.Type() != core.StringNode {
			t.Errorf("expected %v, got %v", core.StringNode, strNode.Type())
		}

		numNode := NewNumberNode(42, "", nil)
		if numNode.Type() != core.NumberNode {
			t.Errorf("expected %v, got %v", core.NumberNode, numNode.Type())
		}

		boolNode := NewBoolNode(true, "", nil)
		if boolNode.Type() != core.BoolNode {
			t.Errorf("expected %v, got %v", core.BoolNode, boolNode.Type())
		}

		nullNode := NewNullNode("", nil)
		if nullNode.Type() != core.NullNode {
			t.Errorf("expected %v, got %v", core.NullNode, nullNode.Type())
		}
	})
}

func TestStringNodeAdditionalMethods(t *testing.T) {
	t.Run("String and MustString methods", func(t *testing.T) {
		strNode := NewStringNode("test value", "", nil)
		if strNode.String() != "test value" {
			t.Errorf("expected 'test value', got '%s'", strNode.String())
		}
		if strNode.MustString() != "test value" {
			t.Errorf("expected 'test value', got '%s'", strNode.MustString())
		}

		// Test with error
		invalidNode := NewInvalidNode("", errors.New("test error"))
		if invalidNode.String() != "" {
			t.Errorf("expected empty string, got '%s'", invalidNode.String())
		}
		assertPanics(t, func() {
			invalidNode.MustString()
		})
	})

	t.Run("Bool conversion", func(t *testing.T) {
		// Test valid boolean strings
		trueNode := NewStringNode("true", "", nil)
		falseNode := NewStringNode("false", "", nil)
		// String nodes always return false for Bool()
		if trueNode.Bool() {
			t.Errorf("expected false for Bool on string 'true'")
		}
		if falseNode.Bool() {
			t.Errorf("expected false for Bool on string 'false'")
		}

		// Test invalid boolean string
		invalidNode := NewStringNode("not a boolean", "", nil)
		if invalidNode.Bool() {
			t.Errorf("expected false for Bool on non-boolean string")
		}
	})

	t.Run("Time conversion", func(t *testing.T) {
		// Test valid time string
		timeStr := "2023-01-01T00:00:00Z"
		timeNode := NewStringNode(timeStr, "", nil)
		expectedTime, _ := time.Parse(time.RFC3339, timeStr)
		if !timeNode.Time().Equal(expectedTime) {
			t.Errorf("expected %v, got %v", expectedTime, timeNode.Time())
		}

		// Test invalid time string
		invalidTimeNode := NewStringNode("not a time", "", nil)
		if !invalidTimeNode.Time().IsZero() {
			t.Errorf("expected zero time, got %v", invalidTimeNode.Time())
		}
	})

	t.Run("Float and Int conversion", func(t *testing.T) {
		// Test valid number strings
		floatNode := NewStringNode("3.14", "", nil)
		intNode := NewStringNode("42", "", nil)
		// String nodes return 0 for numeric values
		if floatNode.Float() != 0 {
			t.Errorf("expected 0.0, got %f", floatNode.Float())
		}
		if intNode.Int() != 0 {
			t.Errorf("expected 0, got %d", intNode.Int())
		}

		// Test invalid number strings
		invalidNode := NewStringNode("not a number", "", nil)
		if invalidNode.Float() != 0 {
			t.Errorf("expected 0.0, got %f", invalidNode.Float())
		}
		if invalidNode.Int() != 0 {
			t.Errorf("expected 0, got %d", invalidNode.Int())
		}
	})

	t.Run("Array method", func(t *testing.T) {
		strNode := NewStringNode("test", "", nil)
		// String nodes return nil for Array()
		result := strNode.Array()
		if result != nil {
			t.Errorf("expected nil, got %v", result)
		}
	})

	t.Run("RawString method", func(t *testing.T) {
		strNode := NewStringNode("test value", "", nil)
		value, ok := strNode.RawString()
		if !ok {
			t.Errorf("expected true, got false")
		}
		if value != "test value" {
			t.Errorf("expected 'test value', got '%s'", value)
		}

		// Test with error
		invalidNode := NewInvalidNode("", errors.New("test error"))
		_, ok = invalidNode.RawString()
		if ok {
			t.Errorf("expected false, got true")
		}
	})

	t.Run("Contains and Strings methods", func(t *testing.T) {
		strNode := NewStringNode("hello world", "", nil)
		// String nodes should return their value in Strings
		if !reflect.DeepEqual([]string{"hello world"}, strNode.Strings()) {
			t.Errorf("expected %v, got %v", []string{"hello world"}, strNode.Strings())
		}
		// Contains should check if the string value contains the substring
		if !strNode.Contains("hello") {
			t.Errorf("expected true for Contains 'hello'")
		}
		if !strNode.Contains("world") {
			t.Errorf("expected true for Contains 'world'")
		}
		if strNode.Contains("universe") {
			t.Errorf("expected false for Contains 'universe'")
		}
	})
}

func TestNumberNodeAdditionalMethods(t *testing.T) {
	t.Run("String and MustString methods", func(t *testing.T) {
		// According to the changes, MustString should now panic for number nodes
		numNode := NewNumberNode(3.14, "", nil)
		if numNode.String() != "3.14" {
			t.Errorf("expected '3.14', got '%s'", numNode.String())
		}
		assertPanics(t, func() {
			numNode.MustString()
		})

		// Test with error
		invalidNode := NewInvalidNode("", errors.New("test error"))
		if invalidNode.String() != "" {
			t.Errorf("expected empty string, got '%s'", invalidNode.String())
		}
		assertPanics(t, func() {
			invalidNode.MustString()
		})
	})

	t.Run("Bool conversion", func(t *testing.T) {
		zeroNode := NewNumberNode(0, "", nil)
		nonZeroNode := NewNumberNode(1, "", nil)
		// Number nodes always return false for Bool()
		if zeroNode.Bool() {
			t.Errorf("expected false for Bool on 0")
		}
		if nonZeroNode.Bool() {
			t.Errorf("expected false for Bool on 1")
		}

		// Test with error
		invalidNode := NewInvalidNode("", errors.New("test error"))
		if invalidNode.Bool() {
			t.Errorf("expected false for Bool on invalid node")
		}
	})

	t.Run("Time conversion", func(t *testing.T) {
		// Number nodes should return zero time
		numNode := NewNumberNode(1234567890, "", nil)
		if !numNode.Time().IsZero() {
			t.Errorf("expected zero time, got %v", numNode.Time())
		}

		// Test with error
		invalidNode := NewInvalidNode("", errors.New("test error"))
		if !invalidNode.Time().IsZero() {
			t.Errorf("expected zero time, got %v", invalidNode.Time())
		}
	})

	t.Run("Interface method", func(t *testing.T) {
		numNode := NewNumberNode(3.14, "", nil)
		if numNode.Interface() != 3.14 {
			t.Errorf("expected 3.14, got %v", numNode.Interface())
		}

		// Test with error
		invalidNode := NewInvalidNode("", errors.New("test error"))
		if invalidNode.Interface() != nil {
			t.Errorf("expected nil, got %v", invalidNode.Interface())
		}
	})

	t.Run("RawFloat method", func(t *testing.T) {
		numNode := NewNumberNode(3.14, "", nil)
		value, ok := numNode.RawFloat()
		if !ok {
			t.Errorf("expected true, got false")
		}
		if value != 3.14 {
			t.Errorf("expected 3.14, got %f", value)
		}

		// Test with error
		invalidNode := NewInvalidNode("", errors.New("test error"))
		_, ok = invalidNode.RawFloat()
		if ok {
			t.Errorf("expected false, got true")
		}
	})

	t.Run("Contains and Strings methods", func(t *testing.T) {
		numNode := NewNumberNode(42, "", nil)
		// These methods should be no-ops for number nodes
		if numNode.Contains("anything") {
			t.Errorf("expected false for Contains on number node")
		}
		if numNode.Strings() != nil {
			t.Errorf("expected nil for Strings on number node")
		}
	})
}

func TestBoolNodeAdditionalMethods(t *testing.T) {
	t.Run("String and MustString methods", func(t *testing.T) {
		trueNode := NewBoolNode(true, "", nil)
		falseNode := NewBoolNode(false, "", nil)
		if trueNode.String() != "true" {
			t.Errorf("expected 'true', got '%s'", trueNode.String())
		}
		if falseNode.String() != "false" {
			t.Errorf("expected 'false', got '%s'", falseNode.String())
		}
		// MustString should panic for bool nodes
		assertPanics(t, func() {
			trueNode.MustString()
		})

		// Test with error
		invalidNode := NewInvalidNode("", errors.New("test error"))
		if invalidNode.String() != "" {
			t.Errorf("expected empty string, got '%s'", invalidNode.String())
		}
		assertPanics(t, func() {
			invalidNode.MustString()
		})
	})

	t.Run("Float, Int, and Bool methods", func(t *testing.T) {
		trueNode := NewBoolNode(true, "", nil)
		falseNode := NewBoolNode(false, "", nil)
		if trueNode.Float() != 0 {
			t.Errorf("expected 0.0, got %f", trueNode.Float())
		}
		if trueNode.Int() != 0 {
			t.Errorf("expected 0, got %d", trueNode.Int())
		}
		if !trueNode.Bool() {
			t.Errorf("expected true, got false")
		}
		if falseNode.Bool() {
			t.Errorf("expected false, got true")
		}

		// Test Must methods with panic
		assertPanics(t, func() {
			trueNode.MustFloat()
		})
		assertPanics(t, func() {
			trueNode.MustInt()
		})

		// Test with error
		invalidNode := NewInvalidNode("", errors.New("test error"))
		if invalidNode.Float() != 0 {
			t.Errorf("expected 0.0 on invalid node, got %f", invalidNode.Float())
		}
		if invalidNode.Int() != 0 {
			t.Errorf("expected 0 on invalid node, got %d", invalidNode.Int())
		}
		if invalidNode.Bool() {
			t.Errorf("expected false on invalid node, got true")
		}
	})

	t.Run("Time conversion", func(t *testing.T) {
		// Bool nodes should return zero time
		boolNode := NewBoolNode(true, "", nil)
		if !boolNode.Time().IsZero() {
			t.Errorf("expected zero time, got %v", boolNode.Time())
		}

		// Test with error
		invalidNode := NewInvalidNode("", errors.New("test error"))
		if !invalidNode.Time().IsZero() {
			t.Errorf("expected zero time on invalid node, got %v", invalidNode.Time())
		}
	})

	t.Run("Interface method", func(t *testing.T) {
		trueNode := NewBoolNode(true, "", nil)
		falseNode := NewBoolNode(false, "", nil)
		if trueNode.Interface() != true {
			t.Errorf("expected true, got %v", trueNode.Interface())
		}
		if falseNode.Interface() != false {
			t.Errorf("expected false, got %v", falseNode.Interface())
		}

		// Test with error
		invalidNode := NewInvalidNode("", errors.New("test error"))
		if invalidNode.Interface() != nil {
			t.Errorf("expected nil on invalid node, got %v", invalidNode.Interface())
		}
	})

	t.Run("Contains and Strings methods", func(t *testing.T) {
		boolNode := NewBoolNode(true, "", nil)
		// These methods should be no-ops for bool nodes
		if boolNode.Contains("anything") {
			t.Errorf("expected false for Contains on bool node")
		}
		if boolNode.Strings() != nil {
			t.Errorf("expected nil for Strings on bool node")
		}
	})
}

func TestNullNodeAdditionalMethods(t *testing.T) {
	t.Run("String and MustString methods", func(t *testing.T) {
		nullNode := NewNullNode("", nil)
		if nullNode.String() != "null" {
			t.Errorf("expected 'null', got '%s'", nullNode.String())
		}
		// MustString should panic for null nodes
		assertPanics(t, func() {
			nullNode.MustString()
		})

		// Test with error
		invalidNode := NewInvalidNode("", errors.New("test error"))
		if invalidNode.String() != "" {
			t.Errorf("expected empty string, got '%s'", invalidNode.String())
		}
		assertPanics(t, func() {
			invalidNode.MustString()
		})
	})

	t.Run("Type conversion methods", func(t *testing.T) {
		nullNode := NewNullNode("", nil)
		if nullNode.Float() != 0 {
			t.Errorf("expected 0.0, got %f", nullNode.Float())
		}
		if nullNode.Int() != 0 {
			t.Errorf("expected 0, got %d", nullNode.Int())
		}
		if nullNode.Bool() {
			t.Errorf("expected false for Bool on null node")
		}
		if !nullNode.Time().IsZero() {
			t.Errorf("expected zero time, got %v", nullNode.Time())
		}

		// Test Must methods with panic
		assertPanics(t, func() {
			nullNode.MustFloat()
		})
		assertPanics(t, func() {
			nullNode.MustInt()
		})
		assertPanics(t, func() {
			nullNode.MustBool()
		})
		assertPanics(t, func() {
			nullNode.MustTime()
		})

		// Test with error
		invalidNode := NewInvalidNode("", errors.New("test error"))
		if invalidNode.Float() != 0 {
			t.Errorf("expected 0.0 on invalid node, got %f", invalidNode.Float())
		}
		if invalidNode.Int() != 0 {
			t.Errorf("expected 0 on invalid node, got %d", invalidNode.Int())
		}
		if invalidNode.Bool() {
			t.Errorf("expected false on invalid node, got true")
		}
		if !invalidNode.Time().IsZero() {
			t.Errorf("expected zero time on invalid node, got %v", invalidNode.Time())
		}
	})

	t.Run("Interface method", func(t *testing.T) {
		nullNode := NewNullNode("", nil)
		if nullNode.Interface() != nil {
			t.Errorf("expected nil, got %v", nullNode.Interface())
		}

		// Test with error
		invalidNode := NewInvalidNode("", errors.New("test error"))
		if invalidNode.Interface() != nil {
			t.Errorf("expected nil on invalid node, got %v", invalidNode.Interface())
		}
	})

	t.Run("Contains and Strings methods", func(t *testing.T) {
		nullNode := NewNullNode("", nil)
		// These methods should be no-ops for null nodes
		if nullNode.Contains("anything") {
			t.Errorf("expected false for Contains on null node")
		}
		if nullNode.Strings() != nil {
			t.Errorf("expected nil for Strings on null node")
		}
	})
}
