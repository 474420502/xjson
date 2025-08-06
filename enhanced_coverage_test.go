package xjson

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestAdvancedCoverageTargets(t *testing.T) {
	jsonData := `{
		"items": [1, 2, 3, 4, 5],
		"complex": {
			"array": ["a", "b", "c"],
			"nested": {
				"deep": {
					"value": "found"
				}
			}
		},
		"filter_test": [
			{"name": "item1", "value": 10},
			{"name": "item2", "value": 20}
		]
	}`

	doc, err := ParseString(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	t.Run("Query_deeper_coverage", func(t *testing.T) {
		// Test array slicing
		result := doc.Query("/items[1:4]")
		if !result.Exists() || result.Count() != 3 {
			t.Errorf("Array slice [1:4] failed, count: %d", result.Count())
		}

		// Test slice with only start index
		result = doc.Query("/items[2:]")
		if !result.Exists() || result.Count() != 3 {
			t.Errorf("Array slice [2:] failed, count: %d", result.Count())
		}

		// Test slice with only end index
		result = doc.Query("/items[:3]")
		if !result.Exists() || result.Count() != 3 {
			t.Errorf("Array slice [:3] failed, count: %d", result.Count())
		}

		// Test filter with complex conditions
		result = doc.Query("/filter_test[?(@.value > 15)]")
		if !result.Exists() || result.Count() != 1 {
			t.Errorf("Filter should find 1 matching item, found %d", result.Count())
		}
		if name, _ := result.Get("name").String(); name != "item2" {
			t.Errorf("Expected filter to find item2, but it did not")
		}

		// Test deeply nested paths
		result = doc.Query("/complex/nested/deep/value")
		if !result.Exists() || result.MustString() != "found" {
			t.Error("Deep nested path should exist and have correct value")
		}

		// Test path that goes through array
		result = doc.Query("/complex/array[1]")
		if !result.Exists() || result.MustString() != "b" {
			t.Error("Path through array should exist and have correct value")
		}

		// Test invalid path that tries to index non-array
		result = doc.Query("/complex/nested[0]")
		if result.Exists() {
			t.Error("Indexing non-array should fail")
		}
	})

	t.Run("Result_String_coverage", func(t *testing.T) {
		// Test with error
		var dummy interface{}
		result := &Result{err: json.Unmarshal([]byte("invalid"), &dummy)}
		_, err := result.String()
		if err == nil {
			t.Error("String() should return error when result has error")
		}

		// Test with empty matches
		result = &Result{matches: []interface{}{}}
		_, err = result.String()
		if err == nil {
			t.Error("String() on empty result should return ErrNotFound")
		}

		// Test with single match
		result = &Result{matches: []interface{}{"hello"}}
		str, err := result.String()
		if err != nil {
			t.Errorf("String() should not error on single match: %v", err)
		}
		if str != "hello" {
			t.Errorf("Expected 'hello', got '%s'", str)
		}

		// Test with multiple matches (should only stringify the first one)
		result = &Result{matches: []interface{}{"a", "b", "c"}}
		str, err = result.String()
		if err != nil {
			t.Errorf("String() should not error on multiple matches: %v", err)
		}
		if str != "a" {
			t.Errorf("String() on multiple results should return the first element, expected 'a', got %s", str)
		}

		// Test with complex object
		complexObj := map[string]interface{}{"name": "test", "value": 42}
		result = &Result{matches: []interface{}{complexObj}}
		str, err = result.String()
		if err != nil {
			t.Errorf("String() should not error on object: %v", err)
		}
		// Compare after unmarshalling to ignore key order differences
		var res map[string]interface{}
		json.Unmarshal([]byte(str), &res)
		if !reflect.DeepEqual(complexObj, res) {
			t.Errorf("Expected marshalled object to be equivalent, got %s", str)
		}
	})

	t.Run("IsArray_more_coverage", func(t *testing.T) {
		// Test IsArray with nil matches
		result := &Result{matches: nil}
		if result.IsArray() {
			t.Error("IsArray should return false for nil matches")
		}

		// Test IsArray with empty matches
		result = &Result{matches: []interface{}{}}
		if result.IsArray() {
			t.Error("IsArray should return false for empty matches")
		}

		// Test IsArray with multiple matches
		result = &Result{matches: []interface{}{1, 2, 3}}
		if !result.IsArray() {
			t.Error("IsArray should return true for multiple matches")
		}

		// Test IsArray with single non-array match
		result = &Result{matches: []interface{}{"string"}}
		if result.IsArray() {
			t.Error("IsArray should return false for single non-array match")
		}

		// Test IsArray with single array match
		result = &Result{matches: []interface{}{[]interface{}{1, 2, 3}}}
		if !result.IsArray() {
			t.Error("IsArray should return true for single array match")
		}
	})
}
