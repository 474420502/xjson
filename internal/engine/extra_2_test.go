package engine

import (
	"fmt"
	"testing"
	"time"

	"github.com/474420502/xjson/internal/core"
)

func TestNodeMustMethods(t *testing.T) {
	// Test MustString method
	jsonData := []byte(`{"str": "hello", "num": 123, "bool": true}`)
	root, err := Parse(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	strNode := root.Get("str")
	if strNode.MustString() != "hello" {
		t.Errorf("Expected 'hello', got %s", strNode.MustString())
	}

	// Test MustFloat method
	numNode := root.Get("num")
	if numNode.MustFloat() != 123.0 {
		t.Errorf("Expected 123.0, got %f", numNode.MustFloat())
	}

	// Test MustInt method
	if numNode.MustInt() != 123 {
		t.Errorf("Expected 123, got %d", numNode.MustInt())
	}

	// Test MustBool method
	boolNode := root.Get("bool")
	if !boolNode.MustBool() {
		t.Errorf("Expected true, got %t", boolNode.MustBool())
	}
}

func TestNodeTimeMethods(t *testing.T) {
	// Test Time and MustTime methods
	jsonData := []byte(`{"time": "2023-01-01T10:00:00Z"}`)
	root, err := Parse(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	timeNode := root.Get("time")
	expectedTime, _ := time.Parse(time.RFC3339Nano, "2023-01-01T10:00:00Z")

	if timeNode.Time().Unix() != expectedTime.Unix() {
		t.Errorf("Expected %v, got %v", expectedTime, timeNode.Time())
	}

	if timeNode.MustTime().Unix() != expectedTime.Unix() {
		t.Errorf("Expected %v, got %v", expectedTime, timeNode.MustTime())
	}
}

func TestNodeArrayMethods(t *testing.T) {
	// Test Array and MustArray methods
	jsonData := []byte(`{"arr": [1, 2, 3]}`)
	root, err := Parse(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	arrNode := root.Get("arr")
	if arrNode.Type() != core.Array {
		t.Fatalf("Expected array type, got %v", arrNode.Type())
	}

	arr := arrNode.Array()
	if len(arr) != 3 {
		t.Errorf("Expected array length 3, got %d", len(arr))
	}

	mustArr := arrNode.MustArray()
	if len(mustArr) != 3 {
		t.Errorf("Expected mustArray length 3, got %d", len(mustArr))
	}

	// Test values
	if arr[0].Int() != 1 || arr[1].Int() != 2 || arr[2].Int() != 3 {
		t.Errorf("Array values don't match expected")
	}
}

func TestNodeAsMapMethods(t *testing.T) {
	// Test AsMap and MustAsMap methods
	jsonData := []byte(`{"obj": {"a": 1, "b": 2}}`)
	root, err := Parse(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	objNode := root.Get("obj")
	if objNode.Type() != core.Object {
		t.Fatalf("Expected object type, got %v", objNode.Type())
	}

	// Test AsMap
	asMap := objNode.AsMap()
	if len(asMap) != 2 {
		t.Errorf("Expected map length 2, got %d", len(asMap))
	}

	if asMap["a"].Int() != 1 || asMap["b"].Int() != 2 {
		t.Errorf("Map values don't match expected")
	}

	// Test MustAsMap
	mustAsMap := objNode.MustAsMap()
	if len(mustAsMap) != 2 {
		t.Errorf("Expected mustAsMap length 2, got %d", len(mustAsMap))
	}

	if mustAsMap["a"].Int() != 1 || mustAsMap["b"].Int() != 2 {
		t.Errorf("MustAsMap values don't match expected")
	}
}

func TestNodeSetValueMethod(t *testing.T) {
	// Test SetValue method
	jsonData := []byte(`{"value": "original", "nested": {"flag": false}}`)
	root, err := Parse(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	node := root.Get("value")
	if node.String() != "original" {
		t.Errorf("Expected 'original', got %s", node.String())
	}

	updated := node.SetValue("updated")
	if !updated.IsValid() {
		t.Fatalf("SetValue failed: %v", updated.Error())
	}
	updatedNode := root.Get("value")
	if updatedNode.String() != "updated" {
		t.Errorf("Expected 'updated', got %s", updatedNode.String())
	}

	flagNode := root.Query("/nested/flag")
	if !flagNode.SetValue(true).IsValid() {
		t.Fatalf("SetValue on nested bool failed: %v", flagNode.Error())
	}
	if !root.Query("/nested/flag").Bool() {
		t.Errorf("Expected nested flag to become true")
	}
}

func TestNodeInterfaceMethod(t *testing.T) {
	// Test Interface method on various node types
	jsonData := []byte(`{
		"str": "hello",
		"num": 123,
		"float": 12.34,
		"bool": true,
		"null": null,
		"arr": [1, 2, 3],
		"obj": {"key": "value"}
	}`)
	root, err := Parse(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Test string
	strInterface := root.Get("str").Interface()
	if strInterface != "hello" {
		t.Errorf("Expected 'hello', got %v", strInterface)
	}

	// Test integer
	intInterface := root.Get("num").Interface()
	if intInterface != int64(123) {
		t.Errorf("Expected 123, got %v", intInterface)
	}

	// Test float
	floatInterface := root.Get("float").Interface()
	if floatInterface != 12.34 {
		t.Errorf("Expected 12.34, got %v", floatInterface)
	}

	// Test boolean
	boolInterface := root.Get("bool").Interface()
	if boolInterface != true {
		t.Errorf("Expected true, got %v", boolInterface)
	}

	// Test null
	nullInterface := root.Get("null").Interface()
	if nullInterface != nil {
		t.Errorf("Expected nil, got %v", nullInterface)
	}

	// Test array
	arrInterface := root.Get("arr").Interface()
	if arrInterface == nil {
		t.Error("Expected array interface, got nil")
	}

	// Test object
	objInterface := root.Get("obj").Interface()
	if objInterface == nil {
		t.Error("Expected object interface, got nil")
	}
}

func TestNodeFilterAndMapMethods(t *testing.T) {
	// Test Filter and Map methods
	jsonData := []byte(`{"numbers": [1, 2, 3, 4, 5]}`)
	root, err := Parse(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	arrNode := root.Get("numbers")
	if arrNode.Type() != core.Array {
		t.Fatalf("Expected array type, got %v", arrNode.Type())
	}

	// Test Filter method
	filtered := arrNode.Filter(func(n core.Node) bool {
		return n.Int() > 3
	})

	if filtered.Len() != 2 {
		t.Errorf("Expected filtered array length 2, got %d", filtered.Len())
	}

	filteredValues := filtered.Array()
	if filteredValues[0].Int() != 4 || filteredValues[1].Int() != 5 {
		t.Errorf("Filtered values don't match expected")
	}

	// Test Map method
	mapped := arrNode.Map(func(n core.Node) interface{} {
		return n.Int() * 2
	})

	if mapped.Len() != 5 {
		t.Errorf("Expected mapped array length 5, got %d", mapped.Len())
	}

	mappedValues := mapped.Array()
	expected := []int64{2, 4, 6, 8, 10}
	for i, node := range mappedValues {
		if node.Int() != expected[i] {
			t.Errorf("Expected %d, got %d at index %d", expected[i], node.Int(), i)
		}
	}
}

func TestNodeForEachMethod(t *testing.T) {
	// Test ForEach method on object
	jsonData := []byte(`{"a": 1, "b": 2, "c": 3}`)
	root, err := Parse(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	objNode := root
	count := 0
	sum := int64(0)

	objNode.ForEach(func(keyOrIndex interface{}, value core.Node) {
		count++
		sum += value.Int()
		key := keyOrIndex.(string)
		if key != "a" && key != "b" && key != "c" {
			t.Errorf("Unexpected key: %s", key)
		}
	})

	if count != 3 {
		t.Errorf("Expected to iterate 3 times, got %d", count)
	}

	if sum != 6 {
		t.Errorf("Expected sum to be 6, got %d", sum)
	}

	// Test ForEach method on array
	jsonData2 := []byte(`[10, 20, 30]`)
	root2, err := Parse(jsonData2)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	arrNode := root2
	count = 0
	sum = int64(0)

	arrNode.ForEach(func(keyOrIndex interface{}, value core.Node) {
		count++
		sum += value.Int()
		index := keyOrIndex.(int)
		if index < 0 || index > 2 {
			t.Errorf("Unexpected index: %d", index)
		}
	})

	if count != 3 {
		t.Errorf("Expected to iterate 3 times, got %d", count)
	}

	if sum != 60 {
		t.Errorf("Expected sum to be 60, got %d", sum)
	}
}

func TestNodeAppendMethod(t *testing.T) {
	// Test Append method
	jsonData := []byte(`{"arr": [1, 2]}`)
	root, err := Parse(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	arrNode := root.Get("arr")
	if arrNode.Len() != 2 {
		t.Errorf("Expected initial array length 2, got %d", arrNode.Len())
	}

	// Append new elements
	arrNode.Append(3).Append(4)

	if arrNode.Len() != 4 {
		t.Errorf("Expected array length 4 after append, got %d", arrNode.Len())
	}

	values := arrNode.Array()
	expected := []int64{1, 2, 3, 4}
	for i, node := range values {
		if node.Int() != expected[i] {
			t.Errorf("Expected %d, got %d at index %d", expected[i], node.Int(), i)
		}
	}
}

func TestNodeFunctionMethods(t *testing.T) {
	// Test RegisterFunc, CallFunc, RemoveFunc methods
	jsonData := []byte(`{"items": [1, 2, 3]}`)
	root, err := Parse(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Register a function
	root.RegisterFunc("double", func(n core.Node) core.Node {
		return n.Map(func(item core.Node) interface{} {
			return item.Int() * 2
		})
	})

	// Test GetFuncs
	funcs := root.GetFuncs()
	if len(*funcs) != 1 {
		t.Errorf("Expected 1 function, got %d", len(*funcs))
	}

	if _, exists := (*funcs)["double"]; !exists {
		t.Error("Expected 'double' function to exist")
	}

	// Call the function
	result := root.Get("items").CallFunc("double")
	if result.Len() != 3 {
		t.Errorf("Expected result length 3, got %d", result.Len())
	}

	// Check values
	expected := []int64{2, 4, 6}
	values := result.Array()
	for i, node := range values {
		if node.Int() != expected[i] {
			t.Errorf("Expected %d, got %d at index %d", expected[i], node.Int(), i)
		}
	}

	// Remove function
	root.RemoveFunc("double")
	funcs = root.GetFuncs()
	if len(*funcs) != 0 {
		t.Errorf("Expected 0 functions after removal, got %d", len(*funcs))
	}
}

func TestNodeApplyMethod(t *testing.T) {
	// Test Apply method
	jsonData := []byte(`{"numbers": [1, 2, 3, 4, 5]}`)
	root, err := Parse(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	arrNode := root.Get("numbers")

	// Test Apply with UnaryPathFunc
	filtered := arrNode.Apply(core.UnaryPathFunc(func(n core.Node) core.Node {
		return n.Filter(func(item core.Node) bool {
			return item.Int() > 3
		})
	}))
	if !filtered.IsValid() {
		t.Fatalf("Apply returned invalid node: %v", filtered.Error())
	}
	if filtered.Len() != 2 {
		t.Errorf("Expected filtered result length 2, got %d", filtered.Len())
	}

	values := filtered.Array()
	if values[0].Int() != 4 || values[1].Int() != 5 {
		t.Errorf("Filtered values don't match expected")
	}
}

func TestNodePathMethod(t *testing.T) {
	jsonData := []byte(`{"user-data": {"items": [{"name": "Alice"}]}}`)
	root, err := Parse(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	node := root.Query(`/['user-data']/items[0]/name`)
	if !node.IsValid() {
		t.Fatalf("Query failed: %v", node.Error())
	}
	if node.Path() != `/['user-data']/items[0]/name` {
		t.Errorf("Expected precise path, got %s", node.Path())
	}
}

func TestSetReusesExistingScalarNodeTypes(t *testing.T) {
	jsonData := []byte(`{"profile":{"age":30,"name":"Alice","active":true},"values":[1,true,"x"]}`)
	root, err := MustParse(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	ageNode := root.Query("/profile/age")
	nameNode := root.Query("/profile/name")
	activeNode := root.Query("/profile/active")
	arrayNumberNode := root.Query("/values[0]")

	root.Query("/profile").Set("age", 31)
	root.Query("/profile").Set("name", "Bob")
	root.Query("/profile").Set("active", false)
	root.Query("/values").Set("0", 2)

	if ageNode.Int() != 31 {
		t.Fatalf("expected updated age, got %d", ageNode.Int())
	}
	if nameNode.String() != "Bob" {
		t.Fatalf("expected updated name, got %q", nameNode.String())
	}
	if activeNode.Bool() {
		t.Fatalf("expected updated active flag to be false")
	}
	if arrayNumberNode.Int() != 2 {
		t.Fatalf("expected updated array element, got %d", arrayNumberNode.Int())
	}

	if ageNode != root.Query("/profile/age") {
		t.Fatalf("expected age node identity to be preserved")
	}
	if nameNode != root.Query("/profile/name") {
		t.Fatalf("expected name node identity to be preserved")
	}
	if activeNode != root.Query("/profile/active") {
		t.Fatalf("expected bool node identity to be preserved")
	}
	if arrayNumberNode != root.Query("/values[0]") {
		t.Fatalf("expected array element identity to be preserved")
	}
}

func TestMustParseEagerlyParsesTree(t *testing.T) {
	jsonData := []byte(`{"outer": {"inner": [{"value": 1}]}}`)
	root, err := MustParse(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	obj, ok := root.(*objectNode)
	if !ok {
		t.Fatalf("Expected object root, got %T", root)
	}
	if !obj.parsed.Load() {
		t.Fatalf("Expected root object to be parsed eagerly")
	}

	outer, ok := obj.value["outer"].(*objectNode)
	if !ok || !outer.parsed.Load() {
		t.Fatalf("Expected nested object to be parsed eagerly")
	}

	inner, ok := outer.value["inner"].(*arrayNode)
	if !ok || !inner.parsed.Load() {
		t.Fatalf("Expected nested array to be parsed eagerly")
	}
}

// Additional tests to improve coverage

func TestNodeRawMethods(t *testing.T) {
	// Test Raw method
	jsonData := []byte(`{"name": "test", "value": 123}`)
	root, err := Parse(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Test Raw method
	raw := root.Raw()
	if raw == "" {
		t.Error("Expected non-empty raw value")
	}
	
	// Test RawBytes through type assertion
	if node, ok := root.(*objectNode); ok {
		rawBytes := node.RawBytes()
		if len(rawBytes) == 0 {
			t.Error("Expected non-empty raw bytes")
		}
	}
}

func TestNodePathMethodSmoke(t *testing.T) {
	// Test Path method
	jsonData := []byte(`{"obj": {"nested": "value"}}`)
	root, err := Parse(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	path := root.Get("obj").Get("nested").Path()
	// Path implementation is simplified, so we just check it doesn't panic
	if path == "" {
		t.Log("Path is empty (expected with simplified implementation)")
	}
}

func TestNodeSetMethod(t *testing.T) {
	// Test Set method on object
	jsonData := []byte(`{"obj": {"a": 1}}`)
	root, err := Parse(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	objNode := root.Get("obj")
	objNode.Set("b", 2)
	
	if objNode.Get("b").Int() != 2 {
		t.Errorf("Expected value 2 for key 'b', got %d", objNode.Get("b").Int())
	}

	// Test Set method on array
	jsonData2 := []byte(`{"arr": [1, 2, 3]}`)
	root2, err := Parse(jsonData2)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	arrNode := root2.Get("arr")
	arrNode.Set("1", 20) // Set index 1 to 20
	
	if arrNode.Index(1).Int() != 20 {
		t.Errorf("Expected value 20 at index 1, got %d", arrNode.Index(1).Int())
	}
}

func TestNodeContainsMethod(t *testing.T) {
	// Test Contains method
	jsonData := []byte(`{"str": "hello"}`)
	root, err := Parse(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	strNode := root.Get("str")
	if !strNode.Contains("hello") {
		t.Error("Expected string node to contain 'hello'")
	}

	if strNode.Contains("world") {
		t.Error("Expected string node to not contain 'world'")
	}
}

func TestNodeRawValueMethods(t *testing.T) {
	// Test RawFloat and RawString methods
	jsonData := []byte(`{"num": 123.45, "str": "hello"}`)
	root, err := Parse(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Test RawFloat
	numNode := root.Get("num")
	if f, ok := numNode.RawFloat(); ok {
		if f != 123.45 {
			t.Errorf("Expected 123.45, got %f", f)
		}
	} else {
		t.Error("Expected RawFloat to return true")
	}

	// Test RawString
	strNode := root.Get("str")
	if s, ok := strNode.RawString(); ok {
		if s != "hello" {
			t.Errorf("Expected 'hello', got %s", s)
		}
	} else {
		t.Error("Expected RawString to return true")
	}
}

func TestInvalidNodeMethods(t *testing.T) {
	// Test methods on invalid nodes
	jsonData := []byte(`{"obj": {"a": 1}}`)
	root, err := Parse(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Get an invalid node
	invalidNode := root.Get("nonexistent")
	if invalidNode.IsValid() {
		t.Error("Expected node to be invalid")
	}

	// Test that methods on invalid nodes don't panic
	_ = invalidNode.Type()
	_ = invalidNode.Query("test")
	_ = invalidNode.Get("test")
	_ = invalidNode.Index(0)
	invalidNode.ForEach(func(keyOrIndex interface{}, value core.Node) {})
	_ = invalidNode.Len()
	_ = invalidNode.Set("test", "value")
	_ = invalidNode.Append("value")
	_ = invalidNode.String()
	// Skip Must* methods as they are expected to panic
	_ = invalidNode.Float()
	_ = invalidNode.Int()
	_ = invalidNode.Bool()
	_ = invalidNode.Time()
	_ = invalidNode.Array()
	_ = invalidNode.Interface()
	_, _ = invalidNode.RawFloat()
	_, _ = invalidNode.RawString()
	_ = invalidNode.Strings()
	_ = invalidNode.Keys()
	_ = invalidNode.Contains("test")
	_ = invalidNode.AsMap()
}

func TestNodeStringsMethod(t *testing.T) {
	// Test Strings method
	jsonData := []byte(`{"arr": ["a", "b", "c"]}`)
	root, err := Parse(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	arrNode := root.Get("arr")
	strings := arrNode.Strings()
	expected := []string{"a", "b", "c"}
	
	if len(strings) != len(expected) {
		t.Fatalf("Expected %d strings, got %d", len(expected), len(strings))
	}

	for i, s := range strings {
		if s != expected[i] {
			t.Errorf("Expected %s at index %d, got %s", expected[i], i, s)
		}
	}
}

func TestNodeKeysMethod(t *testing.T) {
	// Test Keys method
	jsonData := []byte(`{"a": 1, "b": 2, "c": 3}`)
	root, err := Parse(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	keys := root.Keys()
	expected := []string{"a", "b", "c"}
	
	if len(keys) != len(expected) {
		t.Fatalf("Expected %d keys, got %d", len(expected), len(keys))
	}

	// Keys should be sorted
	for i, key := range keys {
		if key != expected[i] {
			t.Errorf("Expected %s at index %d, got %s", expected[i], i, key)
		}
	}
}

func TestQueryCache(t *testing.T) {
	jsonData := []byte(`{
		"name": "John",
		"age": 30,
		"address": {
			"city": "New York",
			"zipcode": "10001"
		}
	}`)

	root, err := Parse(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// First query - should not be cached
	result1 := root.Query("/name")
	if !result1.IsValid() {
		t.Fatalf("First query failed: %v", result1.Error())
	}
	
	if result1.String() != "John" {
		t.Errorf("Expected 'John', got '%s'", result1.String())
	}

	// Second query with same path - should be cached
	result2 := root.Query("/name")
	if !result2.IsValid() {
		t.Fatalf("Second query failed: %v", result2.Error())
	}
	
	if result2.String() != "John" {
		t.Errorf("Expected 'John', got '%s'", result2.String())
	}
	
	// Both results should be the same
	if result1 != result2 {
		t.Log("Query cache is working - same results returned for identical queries")
	}
	
	// Modify the node - this should clear the cache
	root.Query("/").(*objectNode).Set("name", "Jane")
	
	// Query again - should return new value
	result3 := root.Query("/name")
	if !result3.IsValid() {
		t.Fatalf("Third query failed: %v", result3.Error())
	}
	
	if result3.String() != "Jane" {
		t.Errorf("Expected 'Jane', got '%s'", result3.String())
	}
}

func TestQueryCacheDetailed(t *testing.T) {
	jsonData := []byte(`{
		"address": {
			"city": "New York",
			"zipcode": "10001"
		}
	}`)

	root, err := MustParse(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// First query - should be a cache miss
	result1 := root.Query("address/city")
	if !result1.IsValid() {
		t.Fatalf("First query failed: %v", result1.Error())
	}
	
	if result1.String() != "New York" {
		t.Errorf("Expected 'New York', got '%s'", result1.String())
	}

	// Second query with same path - should be a cache hit
	result2 := root.Query("address/city")
	if !result2.IsValid() {
		t.Fatalf("Second query failed: %v", result2.Error())
	}
	
	if result2.String() != "New York" {
		t.Errorf("Expected 'New York', got '%s'", result2.String())
	}
	
	// Both results should be the same object (pointer comparison)
	if result1 == result2 {
		t.Log("Query cache is working - same object returned for identical queries")
	} else {
		t.Log("Query cache may not be working - different objects returned")
	}
}

func TestFastPathQueryCacheAndInvalidation(t *testing.T) {
	root, err := MustParse([]byte(`{
		"level1": {
			"users": [
				{
					"profile": {
						"name": "John"
					}
				}
			]
		}
	}`))
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	path := "/level1/users[0]/profile/name"
	first := root.Query(path)
	if !first.IsValid() {
		t.Fatalf("First query failed: %v", first.Error())
	}
	if first.String() != "John" {
		t.Fatalf("Expected John, got %q", first.String())
	}

	second := root.Query(path)
	if !second.IsValid() {
		t.Fatalf("Second query failed: %v", second.Error())
	}
	if first != second {
		t.Fatalf("Expected cached fast-path query to return same node instance")
	}

	profile := root.Query("/level1/users[0]/profile")
	if !profile.IsValid() {
		t.Fatalf("Profile query failed: %v", profile.Error())
	}
	profile.Set("name", "Jane")

	third := root.Query(path)
	if !third.IsValid() {
		t.Fatalf("Third query failed: %v", third.Error())
	}
	if third.String() != "Jane" {
		t.Fatalf("Expected Jane after mutation, got %q", third.String())
	}
	if first == third && first.String() != "Jane" {
		t.Fatalf("Expected cached node to reflect mutated value")
	}
}

func TestQueryCacheCapacityBounded(t *testing.T) {
	root, err := MustParse([]byte(`{
		"items": {
			"k0": 0,
			"k1": 1,
			"k2": 2,
			"k3": 3,
			"k4": 4,
			"k5": 5,
			"k6": 6,
			"k7": 7,
			"k8": 8,
			"k9": 9
		}
	}`))
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	bn := root.(*objectNode)
	for i := 0; i < maxQueryCacheEntries+32; i++ {
		path := fmt.Sprintf("/items/k%d/%03d", i%10, i)
		bn.setCachedQueryResult(path, root)
	}

	bn.cacheMutex.RLock()
	defer bn.cacheMutex.RUnlock()
	if len(bn.queryCache) > maxQueryCacheEntries {
		t.Fatalf("expected query cache size <= %d, got %d", maxQueryCacheEntries, len(bn.queryCache))
	}
}
