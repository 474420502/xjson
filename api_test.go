package xjson

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

func TestParseInvalidJSON(t *testing.T) {
	testCases := []string{
		`{"name": "John",`,  // Incomplete object
		`{"name": }`,        // Missing value
		`{name: "John"}`,    // Missing quotes on key
		`{"name": "John"`,   // Missing closing brace
		`[1, 2, 3`,          // Incomplete array
		`{"a": 1, "b": }`,   // Missing value in object
		`{"a": 1,, "b": 2}`, // Double comma
		`null null`,         // Multiple values
		`undefined`,         // Invalid literal
		`{"a": 01}`,         // Invalid number with leading zero
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("invalid_json_%s", tc), func(t *testing.T) {
			doc, err := ParseString(tc)
			if err == nil {
				t.Errorf("Expected error for invalid JSON: %s", tc)
			}
			if doc != nil && doc.IsValid() {
				t.Errorf("Document should not be valid for invalid JSON: %s", tc)
			}
		})
	}
}

func TestParseValidJSON(t *testing.T) {
	testCases := []string{
		`{}`,
		`[]`,
		`null`,
		`true`,
		`false`,
		`"string"`,
		`123`,
		`123.456`,
		`{"name": "John", "age": 30}`,
		`[1, 2, 3, {"nested": true}]`,
		`{"complex": {"nested": {"deep": [1, 2, {"even": "deeper"}]}}}`,
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("valid_json_%s", tc), func(t *testing.T) {
			doc, err := ParseString(tc)
			if err != nil {
				t.Errorf("Unexpected error for valid JSON: %s, error: %v", tc, err)
			}
			if doc == nil || !doc.IsValid() {
				t.Errorf("Document should be valid for valid JSON: %s", tc)
			}
			if doc.IsMaterialized() {
				t.Errorf("Document should not be materialized initially: %s", tc)
			}
		})
	}
}

func TestQueryInvalidPath(t *testing.T) {
	doc, _ := ParseString(`{"name": "John", "age": 30}`)

	// Test invalid XPath syntax (this depends on your parser implementation)
	invalidPaths := []string{
		"[",       // Unmatched bracket
		"[?(",     // Incomplete filter
		"[1:2:3]", // Invalid slice
		"$..[?(",  // Incomplete recursive filter
	}

	for _, path := range invalidPaths {
		t.Run(fmt.Sprintf("invalid_path_%s", path), func(t *testing.T) {
			result := doc.Query(path)
			// Should return an empty result rather than panic
			if result.Exists() {
				t.Errorf("Expected empty result for invalid path: %s", path)
			}
		})
	}
}

func TestResultTypeConversions(t *testing.T) {
	testJSON := `{
		"string_val": "hello",
		"int_val": 42,
		"float_val": 3.14,
		"bool_val": true,
		"null_val": null,
		"numeric_string": "123",
		"bool_string": "true",
		"array_val": [1, 2, 3],
		"object_val": {"nested": "value"}
	}`

	doc, err := ParseString(testJSON)
	if err != nil {
		t.Fatalf("ParseString error: %v", err)
	}

	// Test successful conversions
	t.Run("string_conversion", func(t *testing.T) {
		result := doc.Query("string_val")
		str, err := result.String()
		if err != nil {
			t.Errorf("String() error: %v", err)
		}
		if str != "hello" {
			t.Errorf("Expected 'hello', got '%s'", str)
		}

		mustStr := result.MustString()
		if mustStr != "hello" {
			t.Errorf("MustString() expected 'hello', got '%s'", mustStr)
		}
	})

	t.Run("int_conversion", func(t *testing.T) {
		result := doc.Query("int_val")
		intVal, err := result.Int()
		if err != nil {
			t.Errorf("Int() error: %v", err)
		}
		if intVal != 42 {
			t.Errorf("Expected 42, got %d", intVal)
		}

		mustInt := result.MustInt()
		if mustInt != 42 {
			t.Errorf("MustInt() expected 42, got %d", mustInt)
		}
	})

	t.Run("float_conversion", func(t *testing.T) {
		result := doc.Query("float_val")
		floatVal, err := result.Float()
		if err != nil {
			t.Errorf("Float() error: %v", err)
		}
		if floatVal != 3.14 {
			t.Errorf("Expected 3.14, got %f", floatVal)
		}

		mustFloat := result.MustFloat()
		if mustFloat != 3.14 {
			t.Errorf("MustFloat() expected 3.14, got %f", mustFloat)
		}
	})

	t.Run("bool_conversion", func(t *testing.T) {
		result := doc.Query("bool_val")
		boolVal, err := result.Bool()
		if err != nil {
			t.Errorf("Bool() error: %v", err)
		}
		if !boolVal {
			t.Errorf("Expected true, got %t", boolVal)
		}

		mustBool := result.MustBool()
		if !mustBool {
			t.Errorf("MustBool() expected true, got %t", mustBool)
		}
	})

	// Test type mismatch errors
	t.Run("string_to_int_error", func(t *testing.T) {
		result := doc.Query("string_val")
		_, err := result.Int()
		if err == nil {
			t.Error("Expected error converting string to int")
		}
	})

	t.Run("array_to_string_conversion", func(t *testing.T) {
		result := doc.Query("array_val")
		str, err := result.String()
		if err != nil {
			t.Errorf("String() should work for arrays via JSON marshal: %v", err)
		}
		if str == "" {
			t.Error("String() should return non-empty JSON for array")
		}
	})

	t.Run("object_to_int_error", func(t *testing.T) {
		result := doc.Query("object_val")
		_, err := result.Int()
		if err == nil {
			t.Error("Expected error converting object to int")
		}
	})

	// Test null handling
	t.Run("null_conversions", func(t *testing.T) {
		result := doc.Query("null_val")
		if !result.IsNull() {
			t.Error("Expected IsNull() to return true for null value")
		}

		// Null values should convert to empty string, not error
		str, err := result.String()
		if err != nil {
			t.Errorf("String() on null should not error: %v", err)
		}
		if str != "" {
			t.Errorf("String() on null should return empty string, got: %s", str)
		}

		// Null values should return type mismatch for int
		_, err = result.Int()
		if err == nil {
			t.Error("Expected error converting null to int")
		}
	})
}

func TestMustPanicBehavior(t *testing.T) {
	doc, _ := ParseString(`{"string_val": "hello", "array_val": [1, 2, 3]}`)

	// Test MustInt panic
	t.Run("must_int_panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected MustInt to panic on string value")
			}
		}()
		result := doc.Query("string_val")
		result.MustInt()
	})

	// Test MustString on array - should not panic since arrays convert to JSON
	t.Run("must_string_on_array", func(t *testing.T) {
		result := doc.Query("array_val")
		str := result.MustString()
		if str == "" {
			t.Error("MustString should return JSON representation of array")
		}
	})

	// Test MustFloat panic
	t.Run("must_float_panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected MustFloat to panic on string value")
			}
		}()
		result := doc.Query("string_val")
		result.MustFloat()
	})

	// Test MustBool - string values should now cause panic (type safety)
	t.Run("must_bool_on_string", func(t *testing.T) {
		result := doc.Query("string_val")
		didPanic := false
		func() {
			defer func() {
				if recover() != nil {
					didPanic = true
				}
			}()
			result.MustBool()
		}()
		if !didPanic {
			t.Error("MustBool should panic on string values due to type mismatch")
		}
	})
}

func TestResultUtilityMethods(t *testing.T) {
	testJSON := `{
		"array": [1, 2, 3, {"nested": "value"}],
		"object": {"a": 1, "b": 2, "c": {"deep": true}},
		"empty_array": [],
		"empty_object": {},
		"null_value": null,
		"string_value": "test"
	}`

	doc, _ := ParseString(testJSON)

	t.Run("exists_method", func(t *testing.T) {
		if !doc.Query("array").Exists() {
			t.Error("array should exist")
		}
		if doc.Query("nonexistent").Exists() {
			t.Error("nonexistent should not exist")
		}
		if !doc.Query("null_value").Exists() {
			t.Error("null_value should exist (but be null)")
		}
	})

	t.Run("is_array_method", func(t *testing.T) {
		if !doc.Query("array").IsArray() {
			t.Error("array should be identified as array")
		}
		if doc.Query("object").IsArray() {
			t.Error("object should not be identified as array")
		}
		if doc.Query("string_value").IsArray() {
			t.Error("string should not be identified as array")
		}
	})

	t.Run("is_object_method", func(t *testing.T) {
		if !doc.Query("object").IsObject() {
			t.Error("object should be identified as object")
		}
		if doc.Query("array").IsObject() {
			t.Error("array should not be identified as object")
		}
		if doc.Query("string_value").IsObject() {
			t.Error("string should not be identified as object")
		}
	})

	t.Run("count_method", func(t *testing.T) {
		arrayResult := doc.Query("array")
		if arrayResult.Count() != 4 {
			t.Errorf("array count should be 4, got %d", arrayResult.Count())
		}

		objectResult := doc.Query("object")
		// For objects, Count() returns 1 (the number of matches), not object keys
		objectCount := objectResult.Count()
		if objectCount != 3 {
			t.Errorf("object match count should be 3, got %d", objectCount)
		}

		// For scalar values, count should be 0 for empty results or 1 for found
		stringResult := doc.Query("string_value")
		if stringResult.Count() != 1 {
			t.Errorf("string result count should be 1, got %d", stringResult.Count())
		}

		emptyResult := doc.Query("nonexistent")
		if emptyResult.Count() != 0 {
			t.Errorf("empty result count should be 0, got %d", emptyResult.Count())
		}
	})

	t.Run("keys_method", func(t *testing.T) {
		keys := doc.Query("object").Keys()
		expectedKeys := []string{"a", "b", "c"}
		if len(keys) != len(expectedKeys) {
			t.Errorf("Expected %d keys, got %d", len(expectedKeys), len(keys))
		}

		// Convert to map for easier comparison
		keyMap := make(map[string]bool)
		for _, key := range keys {
			keyMap[key] = true
		}
		for _, expected := range expectedKeys {
			if !keyMap[expected] {
				t.Errorf("Expected key '%s' not found", expected)
			}
		}

		// Non-object should return empty keys
		stringKeys := doc.Query("string_value").Keys()
		if len(stringKeys) != 0 {
			t.Errorf("String value should have 0 keys, got %d", len(stringKeys))
		}
	})

	t.Run("first_last_methods", func(t *testing.T) {
		arrayResult := doc.Query("array")

		// First() and Last() work on the matches themselves, not array contents
		// Since array query returns a single match (the array), first and last are the same
		first := arrayResult.First()
		if !first.Exists() {
			t.Error("First element should exist")
		}
		if !first.IsArray() {
			t.Error("First result should be the array itself")
		}

		last := arrayResult.Last()
		if !last.Exists() {
			t.Error("Last element should exist")
		}
		if !last.IsArray() {
			t.Error("Last result should be the array itself")
		}

		// For array contents, we need to use array access methods
		firstElement := arrayResult.Index(0)
		if !firstElement.Exists() {
			t.Error("First array element should exist")
		}
		firstInt, err := firstElement.Int()
		if err != nil {
			t.Errorf("First element should be convertible to int: %v", err)
		}
		if firstInt != 1 {
			t.Errorf("First element should be 1, got %d", firstInt)
		}

		lastElement := arrayResult.Index(3)
		if !lastElement.Exists() {
			t.Error("Last array element should exist")
		}
		if !lastElement.IsObject() {
			t.Error("Last array element should be an object")
		}

		// Empty result
		emptyResult := doc.Query("nonexistent")
		if emptyResult.First().Exists() {
			t.Error("First of empty result should not exist")
		}
		if emptyResult.Last().Exists() {
			t.Error("Last of empty result should not exist")
		}
	})
}

func TestIterationMethods(t *testing.T) {
	testJSON := `{
		"numbers": [1, 2, 3, 4, 5],
		"objects": [
			{"name": "Alice", "age": 30},
			{"name": "Bob", "age": 25},
			{"name": "Charlie", "age": 35}
		]
	}`

	doc, _ := ParseString(testJSON)

	t.Run("foreach_method", func(t *testing.T) {
		numbersResult := doc.Query("numbers")
		sum := 0
		numbersResult.ForEach(func(index int, value IResult) bool {
			val, _ := value.Int()
			sum += val
			return true // continue iteration
		})

		if sum != 15 {
			t.Errorf("Expected sum of 15, got %d", sum)
		}

		// Test early termination
		count := 0
		numbersResult.ForEach(func(index int, value IResult) bool {
			count++
			return index < 2 // stop after index 2
		})

		if count != 3 {
			t.Errorf("Expected count of 3, got %d", count)
		}
	})

	t.Run("map_method", func(t *testing.T) {
		numbersResult := doc.Query("numbers")
		doubled := numbersResult.Map(func(index int, value IResult) interface{} {
			val, _ := value.Int()
			return val * 2
		})

		expected := []interface{}{2, 4, 6, 8, 10}
		if !reflect.DeepEqual(doubled, expected) {
			t.Errorf("Expected %v, got %v", expected, doubled)
		}
	})

	t.Run("filter_method", func(t *testing.T) {
		numbersResult := doc.Query("numbers")

		// Filter on matches level, not array content level
		// Since we have one match (the array), filtering works differently
		filtered := numbersResult.Filter(func(index int, value IResult) bool {
			// This filters the matches, not the array elements
			// Since we only have one match (the array itself), this will either keep or remove it
			return true // Keep the array match
		})

		if filtered.Count() != 5 { // The array has 5 elements
			t.Errorf("Expected filtered array to have 5 elements, got %d", filtered.Count())
		}

		// Filter that removes the match entirely
		filteredOut := numbersResult.Filter(func(index int, value IResult) bool {
			return false // Remove the array match
		})

		if filteredOut.Count() != 0 {
			t.Errorf("Expected filtered-out result to be empty, got %d", filteredOut.Count())
		}

		// To filter array contents, we would need to use ForEach or Map to process elements
		var filteredElements []interface{}
		numbersResult.ForEach(func(index int, value IResult) bool {
			val, _ := value.Int()
			if val > 3 {
				filteredElements = append(filteredElements, val)
			}
			return true
		})

		if len(filteredElements) != 2 {
			t.Errorf("Expected 2 elements > 3, got %d", len(filteredElements))
		}
	})

	t.Run("iteration_on_empty_result", func(t *testing.T) {
		emptyResult := doc.Query("nonexistent")

		// ForEach on empty should not execute callback
		executed := false
		emptyResult.ForEach(func(index int, value IResult) bool {
			executed = true
			return true
		})
		if executed {
			t.Error("ForEach callback should not execute on empty result")
		}

		// Map on empty should return empty slice
		mapped := emptyResult.Map(func(index int, value IResult) interface{} {
			return "mapped"
		})
		if len(mapped) != 0 {
			t.Errorf("Map on empty result should return empty slice, got %v", mapped)
		}

		// Filter on empty should return empty result
		filtered := emptyResult.Filter(func(index int, value IResult) bool {
			return true
		})
		if filtered.Exists() {
			t.Error("Filter on empty result should return empty result")
		}
	})
}

func TestDocumentSerialization(t *testing.T) {
	originalJSON := `{"name": "John", "age": 30, "items": [1, 2, 3]}`

	doc, err := ParseString(originalJSON)
	if err != nil {
		t.Fatalf("ParseString error: %v", err)
	}

	// Test Bytes() before materialization
	bytes, err := doc.Bytes()
	if err != nil {
		t.Errorf("Bytes() error: %v", err)
	}

	// Parse the result and compare
	resultDoc, err := ParseString(string(bytes))
	if err != nil {
		t.Errorf("Failed to parse serialized bytes: %v", err)
	}

	// Compare key values
	if resultDoc.Query("name").MustString() != "John" {
		t.Error("Name mismatch in serialized result")
	}
	if resultDoc.Query("age").MustInt() != 30 {
		t.Error("Age mismatch in serialized result")
	}

	// Test String() method
	str, err := doc.String()
	if err != nil {
		t.Errorf("String() error: %v", err)
	}
	if str != string(bytes) {
		t.Error("String() should return same as Bytes()")
	}

	// Test after materialization
	err = doc.Set("name", "Jane")
	if err != nil {
		t.Errorf("Set error: %v", err)
	}

	// Verify materialization
	if !doc.IsMaterialized() {
		t.Error("Document should be materialized after Set")
	}

	// Test serialization after materialization
	bytes2, err := doc.Bytes()
	if err != nil {
		t.Errorf("Bytes() error after materialization: %v", err)
	}

	resultDoc2, err := ParseString(string(bytes2))
	if err != nil {
		t.Errorf("Failed to parse materialized bytes: %v", err)
	}

	if resultDoc2.Query("name").MustString() != "Jane" {
		t.Error("Name should be updated in materialized result")
	}
}

func TestErrorAfterInvalidDocument(t *testing.T) {
	// Create document with invalid JSON
	doc, err := ParseString(`{"invalid": }`)
	if err == nil {
		t.Fatal("Expected error for invalid JSON")
	}

	// All operations on invalid document should return errors
	result := doc.Query("any")
	if result.Exists() {
		t.Error("Query on invalid document should return empty result")
	}

	setErr := doc.Set("key", "value")
	if setErr == nil {
		t.Error("Set on invalid document should return error")
	}

	deleteErr := doc.Delete("key")
	if deleteErr == nil {
		t.Error("Delete on invalid document should return error")
	}

	_, bytesErr := doc.Bytes()
	if bytesErr == nil {
		t.Error("Bytes on invalid document should return error")
	}

	_, stringErr := doc.String()
	if stringErr == nil {
		t.Error("String on invalid document should return error")
	}
}

// Test Int64 and MustInt64 functions (0% coverage)
func TestInt64Methods(t *testing.T) {
	jsonData := `{
		"smallInt": 42,
		"largeInt": 1234567890,
		"float": 3.14,
		"string": "123",
		"invalidString": "abc",
		"bool": true,
		"null": null
	}`

	doc, err := ParseString(jsonData)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	// Test Int64() method
	t.Run("Int64_valid_small_int", func(t *testing.T) {
		result := doc.Query("smallInt")
		if !result.Exists() {
			t.Error("smallInt should exist")
			return
		}

		val, err := result.Int64()
		if err != nil {
			t.Errorf("Int64() error: %v", err)
			return
		}

		if val != 42 {
			t.Errorf("Expected 42, got %d", val)
		}
	})

	t.Run("Int64_valid_large_int", func(t *testing.T) {
		result := doc.Query("largeInt")
		if !result.Exists() {
			t.Error("largeInt should exist")
			return
		}

		val, err := result.Int64()
		if err != nil {
			t.Errorf("Int64() error: %v", err)
			return
		}

		if val != 1234567890 {
			t.Errorf("Expected 1234567890, got %d", val)
		}
	})

	t.Run("Int64_from_string", func(t *testing.T) {
		result := doc.Query("string")
		if !result.Exists() {
			t.Error("string should exist")
			return
		}

		val, err := result.Int64()
		if err != nil {
			t.Errorf("Int64() error: %v", err)
			return
		}

		if val != 123 {
			t.Errorf("Expected 123, got %d", val)
		}
	})

	t.Run("Int64_invalid_string", func(t *testing.T) {
		result := doc.Query("invalidString")
		if !result.Exists() {
			t.Error("invalidString should exist")
			return
		}

		_, err := result.Int64()
		if err == nil {
			t.Error("Expected error for invalid string")
		}
	})

	t.Run("Int64_from_bool", func(t *testing.T) {
		result := doc.Query("bool")
		if !result.Exists() {
			t.Error("bool should exist")
			return
		}

		_, err := result.Int64()
		if err == nil {
			t.Error("Expected error for bool type")
		}
	})

	t.Run("Int64_non_existent", func(t *testing.T) {
		result := doc.Query("nonExistent")

		_, err := result.Int64()
		if err == nil {
			t.Error("Expected error for non-existent value")
		}
	})

	// Test MustInt64() method
	t.Run("MustInt64_valid", func(t *testing.T) {
		result := doc.Query("smallInt")

		val := result.MustInt64()
		if val != 42 {
			t.Errorf("Expected 42, got %d", val)
		}
	})

	t.Run("MustInt64_invalid_should_panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("MustInt64 should panic on invalid value")
			}
		}()

		result := doc.Query("invalidString")
		result.MustInt64()
	})
}

// Test additional low-coverage functions
func TestLowCoverageFunctions(t *testing.T) {
	jsonData := `{
		"data": [1, 2, 3],
		"nested": {"value": 42}
	}`

	doc, err := ParseString(jsonData)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	// Test Result.Bytes() method (60% coverage)
	t.Run("Result_Bytes", func(t *testing.T) {
		result := doc.Query("data")
		bytes, err := result.Bytes()

		if err != nil {
			t.Errorf("Bytes() error: %v", err)
			return
		}

		if bytes == nil {
			t.Error("Bytes() should return non-nil for valid result")
		}

		// Test with string result
		expected := "[1,2,3]"
		if string(bytes) != expected {
			t.Errorf("Expected %s, got %s", expected, string(bytes))
		}
	})

	// Test Document.materialize function (63.6% coverage)
	t.Run("Document_materialize", func(t *testing.T) {
		// Force materialization by modifying document
		err := doc.Set("newField", "newValue")
		if err != nil {
			t.Errorf("Set error: %v", err)
		}

		// Check if document is materialized
		if !doc.IsMaterialized() {
			t.Error("Document should be materialized after Set operation")
		}
	})
}

// 补充低覆盖率和边界测试
func TestResultAndDocumentEdgeCases(t *testing.T) {
	// 空对象、空数组、嵌套对象、嵌套数组、null、类型不匹配
	doc, _ := ParseString(`{"empty_obj":{}, "empty_arr":[], "nested":{"a":1}, "arr":[{"b":2}], "null":null}`)

	t.Run("EmptyObject", func(t *testing.T) {
		res := doc.Query("empty_obj")
		if !res.IsObject() {
			t.Error("empty_obj 应为对象")
		}
		if res.Count() != 0 {
			t.Errorf("empty_obj count 应为1, got %d", res.Count())
		}
		if len(res.Keys()) != 0 {
			t.Errorf("empty_obj keys 应为0, got %d", len(res.Keys()))
		}
	})

	t.Run("EmptyArray", func(t *testing.T) {
		res := doc.Query("empty_arr")
		if !res.IsArray() {
			t.Error("empty_arr 应为数组")
		}
		if res.Count() != 0 {
			t.Errorf("empty_arr count 应为0, got %d", res.Count())
		}
		// 空数组的 First/Last 可能返回空数组本身
		first := res.First()
		if !first.IsArray() || first.Count() != 0 {
			t.Error("First of empty array 应为空数组")
		}
		last := res.Last()
		if !last.IsArray() || last.Count() != 0 {
			t.Error("Last of empty array 应为空数组")
		}
	})

	t.Run("NestedObject", func(t *testing.T) {
		res := doc.Query("nested")
		if !res.IsObject() {
			t.Error("nested 应为对象")
		}
		if len(res.Keys()) != 1 || res.Keys()[0] != "a" {
			t.Error("nested keys 应为[a]")
		}
	})

	t.Run("NestedArray", func(t *testing.T) {
		res := doc.Query("arr")
		if !res.IsArray() {
			t.Error("arr 应为数组")
		}
		first := res.Index(0)
		if !first.Exists() {
			t.Error("arr[0] 应存在")
		}
		s := first.MustString()
		var m map[string]interface{}
		err := json.Unmarshal([]byte(s), &m)
		if err != nil {
			t.Errorf("arr[0] unmarshal 失败: %v", err)
		}
		if m["b"] != float64(2) {
			t.Errorf("arr[0].b 应为2, got %v", m["b"])
		}
	})

	t.Run("NullValue", func(t *testing.T) {
		res := doc.Query("null")
		if !res.IsNull() {
			t.Error("null 应为null")
		}
		if res.Exists() != true {
			t.Error("null 应存在")
		}
		if _, err := res.Int(); err == nil {
			t.Error("null 转 int 应报错")
		}
	})

	t.Run("NonExistent", func(t *testing.T) {
		res := doc.Query("notfound")
		if res.Exists() {
			t.Error("notfound 不应存在")
		}
		if res.IsNull() {
			t.Error("notfound 不应为null")
		}
		if res.IsArray() {
			t.Error("notfound 不应为array")
		}
		if res.IsObject() {
			t.Error("notfound 不应为object")
		}
		if res.Count() != 0 {
			t.Error("notfound count 应为0")
		}
		if len(res.Keys()) != 0 {
			t.Error("notfound keys 应为0")
		}
		if res.Raw() != nil {
			t.Error("notfound raw 应为nil")
		}
		if b, _ := res.Bytes(); b != nil {
			t.Error("notfound bytes 应为nil")
		}
	})

	// Document.Set/Delete/Bytes/String/IsMaterialized/IsValid 的异常和边界
	doc2, _ := ParseString(`{"a":1}`)
	t.Run("SetAndDelete", func(t *testing.T) {
		err := doc2.Set("b", 2)
		if err != nil {
			t.Error("Set b 失败")
		}
		if doc2.Query("b").MustInt() != 2 {
			t.Error("b 应为2")
		}
		err = doc2.Delete("a")
		if err != nil {
			t.Error("Delete a 失败")
		}
		if doc2.Query("a").Exists() {
			t.Error("a 应被删除")
		}
	})

	t.Run("BytesStringMaterialized", func(t *testing.T) {
		_, err := doc2.Bytes()
		if err != nil {
			t.Error("Bytes 应无错")
		}
		_, err = doc2.String()
		if err != nil {
			t.Error("String 应无错")
		}
		if !doc2.IsMaterialized() {
			t.Error("应已物化")
		}
		if !doc2.IsValid() {
			t.Error("应为有效文档")
		}
	})

	// panic 分支
	t.Run("MustIntPanic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("MustInt 应 panic")
			}
		}()
		_ = doc.Query("empty_obj").MustInt()
	})
}

func TestResultIndexEdgeCases(t *testing.T) {
	doc, _ := ParseString(`{"arr":[1,2,3],"obj":{"a":1}}`)
	arr := doc.Query("arr")
	obj := doc.Query("obj")
	// 越界
	if _, err := arr.Index(10).Int(); err == nil {
		t.Error("Index 越界应报错")
	}
	// 负数越界
	if _, err := arr.Index(-10).Int(); err == nil {
		t.Error("负数 Index 越界应报错")
	}
	// 非数组
	if _, err := obj.Index(0).Int(); err == nil {
		t.Error("非数组 Index 应报错")
	}
}

func TestSetDeleteInvalidDoc(t *testing.T) {
	doc, _ := ParseString(`{"a":}`)
	if err := doc.Set("b", 1); err == nil {
		t.Error("无效文档 Set 应报错")
	}
	if err := doc.Delete("a"); err == nil {
		t.Error("无效文档 Delete 应报错")
	}
}

func TestForEachBreak(t *testing.T) {
	doc, _ := ParseString(`{"arr":[1,2,3]}`)
	arr := doc.Query("arr")
	count := 0
	arr.ForEach(func(i int, v IResult) bool {
		count++
		return false // 只执行一次
	})
	if count != 1 {
		t.Error("ForEach 应能提前 break")
	}
}
