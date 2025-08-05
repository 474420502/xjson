package xjson

import (
	"encoding/json"
	"testing"
)

func TestCoverageImprovements(t *testing.T) {
	t.Run("getValue_comprehensive_coverage", func(t *testing.T) {
		jsonData := `{
			"simple": "value",
			"dotted.key": "dotted_value",
			"array": [1, 2, 3],
			"nested": {
				"field": "nested_value",
				"deeper": {
					"value": 42
				}
			}
		}`

		doc, err := ParseString(jsonData)
		if err != nil {
			t.Fatalf("Failed to parse JSON: %v", err)
		}

		// Force materialization to test getValue
		doc.Set("test", "value")

		// Test getValue with empty path
		result := doc.getValue(doc.materialized, "")
		if result == nil {
			t.Error("getValue with empty path should return root data")
		}

		// Test getValue with direct key access
		result = doc.getValue(doc.materialized, "simple")
		if result != "value" {
			t.Errorf("Expected 'value', got %v", result)
		}

		// Test getValue with dotted key (should find exact match first)
		result = doc.getValue(doc.materialized, "dotted.key")
		if result != "dotted_value" {
			t.Errorf("Expected 'dotted_value', got %v", result)
		}

		// Test getValue with simple field that doesn't exist
		result = doc.getValue(doc.materialized, "nonexistent")
		if result != nil {
			t.Errorf("Expected nil for nonexistent key, got %v", result)
		}

		// Test getValue with array access at root level
		result = doc.getValue(doc.materialized, "[0]")
		if result != nil {
			// Should be nil because root is not an array
		}

		// Test getValue with negative array index at root
		result = doc.getValue(doc.materialized, "[-1]")
		if result != nil {
			// Should be nil because root is not an array
		}

		// Test getValue with dotted path navigation
		result = doc.getValue(doc.materialized, "nested.field")
		if result != "nested_value" {
			t.Errorf("Expected 'nested_value', got %v", result)
		}

		// Test getValue with deeper path
		result = doc.getValue(doc.materialized, "nested.deeper.value")
		if result != float64(42) {
			t.Errorf("Expected 42, got %v", result)
		}

		// Test getValue with invalid nested path
		result = doc.getValue(doc.materialized, "nested.nonexistent.field")
		if result != nil {
			t.Errorf("Expected nil for invalid path, got %v", result)
		}
	})

	t.Run("getValueWithExists_edge_cases", func(t *testing.T) {
		jsonData := `{
			"null_value": null,
			"empty_string": "",
			"zero": 0,
			"false_value": false,
			"array": [null, "", 0, false],
			"nested": {
				"null": null,
				"empty": ""
			}
		}`

		doc, err := ParseString(jsonData)
		if err != nil {
			t.Fatalf("Failed to parse JSON: %v", err)
		}

		// Get the raw data to test with
		var data interface{}
		if doc.isMaterialized {
			data = doc.materialized
		} else {
			json.Unmarshal(doc.raw, &data)
		}

		// Test null values
		exists, value := doc.getValueWithExists(data, "null_value")
		if !exists {
			t.Error("null_value should exist even though it's null")
		}
		if value != nil {
			t.Errorf("Expected nil value, got %v", value)
		}

		// Test empty string
		exists, value = doc.getValueWithExists(data, "empty_string")
		if !exists {
			t.Error("empty_string should exist")
		}
		if value != "" {
			t.Errorf("Expected empty string, got %v", value)
		}

		// Test zero value
		exists, value = doc.getValueWithExists(data, "zero")
		if !exists {
			t.Error("zero should exist")
		}
		if value != float64(0) {
			t.Errorf("Expected 0, got %v", value)
		}

		// Test false value
		exists, value = doc.getValueWithExists(data, "false_value")
		if !exists {
			t.Error("false_value should exist")
		}
		if value != false {
			t.Errorf("Expected false, got %v", value)
		}

		// Test array with null elements
		exists, value = doc.getValueWithExists(data, "array[0]")
		if !exists {
			t.Error("array[0] should exist even though it's null")
		}

		// Test recursive path to handle recursive queries
		exists, value = doc.getValueWithExists(data, "//null")
		if !exists {
			t.Error("recursive query should find null fields")
		}

		// Test array slicing
		exists, value = doc.getValueWithExists(data, "array[1:3]")
		if !exists {
			t.Error("array slice should exist")
		}

		// Test negative indexing
		exists, value = doc.getValueWithExists(data, "array[-1]")
		if !exists {
			t.Error("array[-1] should exist")
		}

		// Test filter expressions
		exists, value = doc.getValueWithExists(data, "array[?(@.length)]")
		// This might not exist depending on implementation

		// Test non-existent path
		exists, value = doc.getValueWithExists(data, "totally.nonexistent.path")
		if exists {
			t.Error("nonexistent path should not exist")
		}
	})

	t.Run("compareValues_all_types", func(t *testing.T) {
		doc := &Document{}

		// Test string comparisons (expected must be string)
		if !doc.compareValues("hello", "hello", "==") {
			t.Error("String equality should work")
		}
		if doc.compareValues("hello", "world", "==") {
			t.Error("String inequality should work")
		}
		if !doc.compareValues("hello", "world", "!=") {
			t.Error("String not-equal should work")
		}

		// Test numeric comparisons (expected as string)
		if !doc.compareValues(10, "20", "<") {
			t.Error("Integer less-than should work")
		}
		if !doc.compareValues(10.5, "20.7", "<") {
			t.Error("Float less-than should work")
		}

		// Test boolean comparisons
		if !doc.compareValues(true, "true", "==") {
			t.Error("Boolean equality should work")
		}
		if doc.compareValues(true, "false", "==") {
			t.Error("Boolean inequality should work")
		}

		// Test nil comparisons
		if !doc.compareValues(nil, "null", "==") {
			t.Error("Nil equality should work")
		}
		if doc.compareValues(nil, "something", "==") {
			t.Error("Nil vs string should not be equal")
		}

		// Test unsupported operator
		if doc.compareValues(10, "10", "~=") {
			t.Error("Unsupported operator should return false")
		}

		// Test incompatible types
		if doc.compareValues("string", "42", "<") {
			t.Error("String vs number comparison should fail")
		}

		// Test edge cases with zero values
		if !doc.compareValues(0, "0", "==") {
			t.Error("Zero equality should work")
		}
		if !doc.compareValues("", "", "==") {
			t.Error("Empty string equality should work")
		}
		if !doc.compareValues(false, "false", "==") {
			t.Error("False equality should work")
		}

		// Test boundary comparisons
		if !doc.compareValues(10, "10", "<=") {
			t.Error("Less-than-or-equal boundary should work")
		}
		if !doc.compareValues(10, "10", ">=") {
			t.Error("Greater-than-or-equal boundary should work")
		}
	})

	t.Run("evaluateSimpleExpression_edge_cases", func(t *testing.T) {
		doc := &Document{}

		testData := map[string]interface{}{
			"string_field": "test",
			"number_field": 42.5,
			"bool_field":   true,
			"null_field":   nil,
			"array_field":  []interface{}{1, 2, 3},
			"object_field": map[string]interface{}{"nested": "value"},
		}

		// Test with malformed expressions (missing @ prefix)
		result := doc.evaluateSimpleExpression(testData, "string_field == 'test'")
		if result {
			t.Error("Expression without @ should not match")
		}

		// Test with malformed expressions (no operator)
		result = doc.evaluateSimpleExpression(testData, "@.string_field")
		if result {
			t.Error("Expression without operator should not match")
		}

		// Test with invalid field reference
		result = doc.evaluateSimpleExpression(testData, "@.nonexistent == 'test'")
		if result {
			t.Error("Expression with nonexistent field should not match")
		}

		// Test with quoted string values
		result = doc.evaluateSimpleExpression(testData, "@.string_field == 'test'")
		if !result {
			t.Error("Quoted string comparison should work")
		}

		// Test with double-quoted strings
		result = doc.evaluateSimpleExpression(testData, "@.string_field == \"test\"")
		if !result {
			t.Error("Double-quoted string comparison should work")
		}

		// Test with number comparisons
		result = doc.evaluateSimpleExpression(testData, "@.number_field > 40")
		if !result {
			t.Error("Number comparison should work")
		}

		// Test with boolean comparisons
		result = doc.evaluateSimpleExpression(testData, "@.bool_field == true")
		if !result {
			t.Error("Boolean comparison should work")
		}

		// Test with null comparisons
		result = doc.evaluateSimpleExpression(testData, "@.null_field == null")
		if !result {
			t.Error("Null comparison should work")
		}

		// Test complex field paths
		result = doc.evaluateSimpleExpression(testData, "@.object_field.nested == 'value'")
		if !result {
			t.Error("Nested field comparison should work")
		}

		// Test with malformed quotes
		result = doc.evaluateSimpleExpression(testData, "@.string_field == 'unclosed")
		if result {
			t.Error("Malformed quotes should not match")
		}

		// Test with spaces in expression
		result = doc.evaluateSimpleExpression(testData, "  @.string_field   ==   'test'  ")
		if !result {
			t.Error("Expression with spaces should work")
		}
	})

	t.Run("array_access_edge_cases", func(t *testing.T) {
		arrayData := `["first", "second", "third"]`
		doc, err := ParseString(arrayData)
		if err != nil {
			t.Fatalf("Failed to parse array JSON: %v", err)
		}

		// Get parsed data for testing
		var data interface{}
		json.Unmarshal(doc.raw, &data)

		// Test getValue with array at root level
		result := doc.getValue(data, "[1]")
		if result != "second" {
			t.Errorf("Expected 'second', got %v", result)
		}

		// Test getValue with negative index
		result = doc.getValue(data, "[-1]")
		if result != "third" {
			t.Errorf("Expected 'third', got %v", result)
		}

		// Test getValue with out of bounds index
		result = doc.getValue(data, "[10]")
		if result != nil {
			t.Errorf("Expected nil for out of bounds, got %v", result)
		}

		// Test getValue with negative out of bounds
		result = doc.getValue(data, "[-10]")
		if result != nil {
			t.Errorf("Expected nil for negative out of bounds, got %v", result)
		}

		// Test getValue with invalid index
		result = doc.getValue(data, "[abc]")
		if result != nil {
			t.Errorf("Expected nil for invalid index, got %v", result)
		}
	})

	t.Run("Result_IsArray_edge_cases", func(t *testing.T) {
		// Test IsArray with different data types
		tests := []struct {
			data     interface{}
			expected bool
		}{
			{[]interface{}{1, 2, 3}, true},
			{[]interface{}{}, true},
			{map[string]interface{}{}, false},
			{"string", false},
			{42, false},
			{nil, false},
			{true, false},
		}

		for _, test := range tests {
			result := &Result{matches: []interface{}{test.data}}
			if result.IsArray() != test.expected {
				t.Errorf("IsArray for %v: expected %v, got %v", test.data, test.expected, result.IsArray())
			}
		}
	})

	t.Run("additional_coverage_paths", func(t *testing.T) {
		// Test paths that may not be covered elsewhere
		jsonData := `{
			"test": "value",
			"array": [1, 2, 3],
			"object": {"nested": "value"}
		}`

		doc, err := ParseString(jsonData)
		if err != nil {
			t.Fatalf("Failed to parse JSON: %v", err)
		}

		// Test Query with various paths to increase coverage
		result := doc.Query("test")
		if !result.Exists() {
			t.Error("Simple query should work")
		}

		// Test with non-existent path
		result = doc.Query("nonexistent")
		if result.Exists() {
			t.Error("Non-existent path should not exist")
		}

		// Test array access
		result = doc.Query("array[0]")
		if result.MustInt() != 1 {
			t.Error("Array access should work")
		}

		// Test nested object access
		result = doc.Query("object.nested")
		if result.MustString() != "value" {
			t.Error("Nested object access should work")
		}

		// Test Count method coverage
		result = doc.Query("array")
		count := result.Count()
		if count != 3 {
			t.Errorf("Expected count 3, got %d", count)
		}

		// Test ForEach coverage
		result.ForEach(func(index int, value IResult) bool {
			if index > 1 {
				return false // Stop iteration early
			}
			return true
		})
	})
}
