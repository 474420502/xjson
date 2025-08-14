package engine

import (
	"strings"
	"testing"

	"github.com/474420502/xjson/internal/core"
)

func TestArrayAppendAndString(t *testing.T) {
	// Test data with nested structure
	jsonData := []byte(`{"a": []}`)

	root, err := Parse(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	arr := root.Query("/a")
	if !arr.IsValid() {
		t.Fatalf("Query failed: %v", arr.Error())
	}

	if arr.Type() != core.Array {
		t.Fatalf("Expected array type, got %v", arr.Type())
	}

	arr.Append(1).Append(2)
	if arr.Error() != nil {
		t.Fatalf("Append failed: %v", arr.Error())
	}

	if arr.Len() != 2 {
		t.Fatalf("Expected length 2, got %d", arr.Len())
	}

	// ensure printable
	s := root.String()
	if s == "" {
		t.Fatal("Expected non-empty string representation")
	}

	// Check if the appended elements are in the string representation
	// The string should contain "a":[1,2] or similar
	if !strings.Contains(s, "[1,2]") && !strings.Contains(s, "[1, 2]") && !strings.Contains(s, "[1, 2") {
		t.Errorf("Expected string to contain appended elements, got %s", s)
	}
}

func TestKeysSorted(t *testing.T) {
	// Test data with unsorted keys
	jsonData := []byte(`{"z":1,"a":2,"b":3,"c":4}`)

	root, err := Parse(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	obj, ok := root.(*objectNode)
	if !ok {
		t.Fatal("Expected objectNode")
	}

	keys := obj.Keys()
	expectedKeys := []string{"a", "b", "c", "z"}

	if len(keys) != len(expectedKeys) {
		t.Fatalf("Expected %d keys, got %d", len(expectedKeys), len(keys))
	}

	for i, key := range keys {
		if key != expectedKeys[i] {
			t.Errorf("Expected key %s at position %d, got %s", expectedKeys[i], i, key)
		}
	}
}

func TestNumberAndStringRawAccess(t *testing.T) {
	// Test data with numbers and strings
	jsonData := []byte(`{"n":12.34,"i":5,"s":"hello"}`)

	root, err := Parse(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Test float number
	n := root.Query("/n")
	if !n.IsValid() {
		t.Fatalf("Query failed: %v", n.Error())
	}

	f, ok := n.RawFloat()
	if !ok {
		t.Fatal("Expected RawFloat to return true")
	}

	const epsilon = 1e-9
	if f < 12.34-epsilon || f > 12.34+epsilon {
		t.Errorf("Expected 12.34, got %f", f)
	}

	// Test integer number
	i := root.Query("/i")
	if !i.IsValid() {
		t.Fatalf("Query failed: %v", i.Error())
	}

	fi, ok := i.RawFloat()
	if !ok {
		t.Fatal("Expected RawFloat to return true")
	}

	if fi < 5.0-epsilon || fi > 5.0+epsilon {
		t.Errorf("Expected 5.0, got %f", fi)
	}

	// Test string
	s := root.Query("/s")
	if !s.IsValid() {
		t.Fatalf("Query failed: %v", s.Error())
	}

	str, ok := s.RawString()
	if !ok {
		t.Fatal("Expected RawString to return true")
	}

	expectedStr := "hello"
	if str != expectedStr {
		t.Errorf("Expected %s, got %s", expectedStr, str)
	}
}

func TestStringsHelper(t *testing.T) {
	// Test data with string array
	jsonData := []byte(`{"arr":["a","b","c"]}`)

	root, err := Parse(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	arr := root.Query("/arr")
	if !arr.IsValid() {
		t.Fatalf("Query failed: %v", arr.Error())
	}

	ss := arr.Strings()
	expectedStrings := []string{"a", "b", "c"}

	if len(ss) != len(expectedStrings) {
		t.Fatalf("Expected %d strings, got %d", len(expectedStrings), len(ss))
	}

	for i, s := range ss {
		if s != expectedStrings[i] {
			t.Errorf("Expected string %s at position %d, got %s", expectedStrings[i], i, s)
		}
	}
}

func TestEmptyArrayAndObject(t *testing.T) {
	// Test data with empty array and object
	jsonData := []byte(`{"emptyArr":[],"emptyObj":{}}`)

	root, err := Parse(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Test empty array
	emptyArr := root.Query("/emptyArr")
	if !emptyArr.IsValid() {
		t.Fatalf("Query failed: %v", emptyArr.Error())
	}

	if emptyArr.Type() != core.Array {
		t.Fatalf("Expected array type, got %v", emptyArr.Type())
	}

	if emptyArr.Len() != 0 {
		t.Errorf("Expected empty array, got length %d", emptyArr.Len())
	}

	// Test empty object
	emptyObj := root.Query("/emptyObj")
	if !emptyObj.IsValid() {
		t.Fatalf("Query failed: %v", emptyObj.Error())
	}

	if emptyObj.Type() != core.Object {
		t.Fatalf("Expected object type, got %v", emptyObj.Type())
	}

	if emptyObj.Len() != 0 {
		t.Errorf("Expected empty object, got length %d", emptyObj.Len())
	}
}

func TestArrayIndexAccess(t *testing.T) {
	// Test data with array elements
	jsonData := []byte(`{"arr":[10,20,30]}`)

	root, err := Parse(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Test valid index access
	first := root.Query("/arr[0]")
	if !first.IsValid() {
		t.Fatalf("Query failed: %v", first.Error())
	}

	if first.Type() != core.Number {
		t.Fatalf("Expected number type, got %v", first.Type())
	}

	if first.Int() != 10 {
		t.Errorf("Expected value 10, got %d", first.Int())
	}

	// Test last element access with negative index
	last := root.Query("/arr[-1]")
	if !last.IsValid() {
		t.Fatalf("Query failed: %v", last.Error())
	}

	if last.Int() != 30 {
		t.Errorf("Expected value 30, got %d", last.Int())
	}

	// Test out of bounds access
	outOfBounds := root.Query("/arr[5]")
	if outOfBounds.IsValid() {
		t.Error("Expected invalid result for out of bounds access")
	}
}

func TestBooleanAndNullValues(t *testing.T) {
	// Test data with boolean and null values
	jsonData := []byte(`{"trueVal":true,"falseVal":false,"nullVal":null}`)

	root, err := Parse(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Test true value
	trueVal := root.Query("/trueVal")
	if !trueVal.IsValid() {
		t.Fatalf("Query failed: %v", trueVal.Error())
	}

	if trueVal.Type() != core.Bool {
		t.Fatalf("Expected bool type, got %v", trueVal.Type())
	}

	if !trueVal.Bool() {
		t.Error("Expected true, got false")
	}

	// Test false value
	falseVal := root.Query("/falseVal")
	if !falseVal.IsValid() {
		t.Fatalf("Query failed: %v", falseVal.Error())
	}

	if falseVal.Bool() {
		t.Error("Expected false, got true")
	}

	// Test null value
	nullVal := root.Query("/nullVal")
	if !nullVal.IsValid() {
		t.Fatalf("Query failed: %v", nullVal.Error())
	}

	if nullVal.Type() != core.Null {
		t.Fatalf("Expected null type, got %v", nullVal.Type())
	}
}
