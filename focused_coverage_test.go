package xjson

import (
	"testing"
)

func TestFocusedCoverageBoost(t *testing.T) {
	t.Run("String_method_edge_cases", func(t *testing.T) {
		// Test Result.String with different scenarios to reach more branches

		// Test with complex unmarshalable data to trigger fmt.Sprintf path
		jsonData := `{"test": "value"}`
		doc, err := ParseString(jsonData)
		if err != nil {
			t.Fatalf("Failed to parse JSON: %v", err)
		}

		result := doc.Query("test")
		str, err := result.String()
		if err != nil {
			t.Errorf("String() should work: %v", err)
		}
		if str != "value" {
			t.Errorf("Expected 'value', got '%s'", str)
		}

		// Test String() with array results
		arrayData := `[1, 2, 3]`
		doc, err = ParseString(arrayData)
		if err != nil {
			t.Fatalf("Failed to parse array JSON: %v", err)
		}

		result = doc.Query("")
		_, err = result.String()
		if err != nil {
			t.Errorf("String() should work for arrays: %v", err)
		}
	})

	t.Run("IsArray_edge_cases", func(t *testing.T) {
		// Test different IsArray scenarios

		// Empty array
		doc, _ := ParseString(`[]`)
		result := doc.Query("")
		if !result.IsArray() {
			t.Error("Empty array should be detected as array")
		}

		// Non-array
		doc, _ = ParseString(`"string"`)
		result = doc.Query("")
		if result.IsArray() {
			t.Error("String should not be detected as array")
		}

		// Object
		doc, _ = ParseString(`{"key": "value"}`)
		result = doc.Query("")
		if result.IsArray() {
			t.Error("Object should not be detected as array")
		}

		// Number
		doc, _ = ParseString(`42`)
		result = doc.Query("")
		if result.IsArray() {
			t.Error("Number should not be detected as array")
		}

		// Boolean
		doc, _ = ParseString(`true`)
		result = doc.Query("")
		if result.IsArray() {
			t.Error("Boolean should not be detected as array")
		}

		// Null
		doc, _ = ParseString(`null`)
		result = doc.Query("")
		if result.IsArray() {
			t.Error("Null should not be detected as array")
		}
	})

	t.Run("Get_method_edge_cases", func(t *testing.T) {
		jsonData := `{
			"nested": {
				"value": "test"
			},
			"simple": "value"
		}`

		doc, err := ParseString(jsonData)
		if err != nil {
			t.Fatalf("Failed to parse JSON: %v", err)
		}

		// Test Get with simple path
		result := doc.Query("nested")
		deepResult := result.Get("value")
		if !deepResult.Exists() {
			t.Error("Get should find nested value")
		}

		// Test Get with non-existent path
		nonExistentResult := result.Get("nonexistent")
		if nonExistentResult.Exists() {
			t.Error("Get should return non-existent result for invalid path")
		}

		// Test Get with empty path
		emptyResult := result.Get("")
		if !emptyResult.Exists() {
			t.Error("Get with empty path should return the current result")
		}
	})

	t.Run("Count_method_edge_cases", func(t *testing.T) {
		// Test Count with different scenarios

		// Array
		doc, _ := ParseString(`[1, 2, 3, 4, 5]`)
		result := doc.Query("")
		count := result.Count()
		if count != 5 {
			t.Errorf("Expected count 5 for array, got %d", count)
		}

		// Object
		doc, _ = ParseString(`{"a": 1, "b": 2, "c": 3}`)
		result = doc.Query("")
		count = result.Count()
		if count != 3 {
			t.Errorf("Expected count 3 for object, got %d", count)
		}

		// Empty array
		doc, _ = ParseString(`[]`)
		result = doc.Query("")
		count = result.Count()
		if count != 0 {
			t.Errorf("Expected count 0 for empty array, got %d", count)
		}

		// Empty object
		doc, _ = ParseString(`{}`)
		result = doc.Query("")
		count = result.Count()
		if count != 0 {
			t.Errorf("Expected count 0 for empty object, got %d", count)
		}

		// Single value
		doc, _ = ParseString(`"single"`)
		result = doc.Query("")
		count = result.Count()
		if count != 1 {
			t.Errorf("Expected count 1 for single value, got %d", count)
		}

		// Null
		doc, _ = ParseString(`null`)
		result = doc.Query("")
		count = result.Count()
		if count != 0 {
			t.Errorf("Expected count 0 for null, got %d", count)
		}

		// Non-existent
		doc, _ = ParseString(`{"a": 1}`)
		result = doc.Query("nonexistent")
		count = result.Count()
		if count != 0 {
			t.Errorf("Expected count 0 for non-existent, got %d", count)
		}
	})

	t.Run("ForEach_edge_cases", func(t *testing.T) {
		// Test ForEach with different break conditions

		jsonData := `[1, 2, 3, 4, 5]`
		doc, _ := ParseString(jsonData)
		result := doc.Query("")

		// Test early break
		count := 0
		result.ForEach(func(index int, value IResult) bool {
			count++
			return index < 2 // Break after 3 iterations
		})

		if count != 3 {
			t.Errorf("Expected 3 iterations with early break, got %d", count)
		}

		// Test full iteration
		count = 0
		result.ForEach(func(index int, value IResult) bool {
			count++
			return true // Continue all iterations
		})

		if count != 5 {
			t.Errorf("Expected 5 iterations for full array, got %d", count)
		}

		// Test ForEach on empty result
		emptyResult := doc.Query("nonexistent")
		count = 0
		emptyResult.ForEach(func(index int, value IResult) bool {
			count++
			return true
		})

		if count != 0 {
			t.Errorf("Expected 0 iterations for empty result, got %d", count)
		}

		// Test ForEach on single value
		doc, _ = ParseString(`"single"`)
		result = doc.Query("")
		count = 0
		result.ForEach(func(index int, value IResult) bool {
			count++
			return true
		})

		if count != 1 {
			t.Errorf("Expected 1 iteration for single value, got %d", count)
		}
	})

	t.Run("Bool_method_edge_cases", func(t *testing.T) {
		// Test Bool method with various data types

		// Boolean true
		doc, _ := ParseString(`true`)
		result := doc.Query("")
		boolVal, err := result.Bool()
		if err != nil {
			t.Errorf("Bool() should work for boolean: %v", err)
		}
		if !boolVal {
			t.Error("Expected true, got false")
		}

		// Boolean false
		doc, _ = ParseString(`false`)
		result = doc.Query("")
		boolVal, err = result.Bool()
		if err != nil {
			t.Errorf("Bool() should work for boolean: %v", err)
		}
		if boolVal {
			t.Error("Expected false, got true")
		}

		// Number 0 (should be false)
		doc, _ = ParseString(`0`)
		result = doc.Query("")
		boolVal, err = result.Bool()
		if err != nil {
			t.Errorf("Bool() should work for number: %v", err)
		}
		if boolVal {
			t.Error("Expected false for 0, got true")
		}

		// Number non-zero (should be true)
		doc, _ = ParseString(`42`)
		result = doc.Query("")
		boolVal, err = result.Bool()
		if err != nil {
			t.Errorf("Bool() should work for number: %v", err)
		}
		if !boolVal {
			t.Error("Expected true for non-zero number, got false")
		}

		// String (should error)
		doc, _ = ParseString(`"string"`)
		result = doc.Query("")
		_, err = result.Bool()
		if err == nil {
			t.Error("Bool() should error for string")
		}

		// Array (should error)
		doc, _ = ParseString(`[1, 2, 3]`)
		result = doc.Query("")
		_, err = result.Bool()
		if err == nil {
			t.Error("Bool() should error for array")
		}

		// Object (should error)
		doc, _ = ParseString(`{"key": "value"}`)
		result = doc.Query("")
		_, err = result.Bool()
		if err == nil {
			t.Error("Bool() should error for object")
		}
	})

	t.Run("Float_method_edge_cases", func(t *testing.T) {
		// Test Float method with various number types

		// Integer
		doc, _ := ParseString(`42`)
		result := doc.Query("")
		floatVal, err := result.Float()
		if err != nil {
			t.Errorf("Float() should work for integer: %v", err)
		}
		if floatVal != 42.0 {
			t.Errorf("Expected 42.0, got %f", floatVal)
		}

		// Float
		doc, _ = ParseString(`3.14`)
		result = doc.Query("")
		floatVal, err = result.Float()
		if err != nil {
			t.Errorf("Float() should work for float: %v", err)
		}
		if floatVal != 3.14 {
			t.Errorf("Expected 3.14, got %f", floatVal)
		}

		// Boolean true (should be 1.0)
		doc, _ = ParseString(`true`)
		result = doc.Query("")
		floatVal, err = result.Float()
		if err != nil {
			t.Errorf("Float() should work for boolean: %v", err)
		}
		if floatVal != 1.0 {
			t.Errorf("Expected 1.0 for true, got %f", floatVal)
		}

		// Boolean false (should be 0.0)
		doc, _ = ParseString(`false`)
		result = doc.Query("")
		floatVal, err = result.Float()
		if err != nil {
			t.Errorf("Float() should work for boolean: %v", err)
		}
		if floatVal != 0.0 {
			t.Errorf("Expected 0.0 for false, got %f", floatVal)
		}

		// String (should error)
		doc, _ = ParseString(`"not a number"`)
		result = doc.Query("")
		_, err = result.Float()
		if err == nil {
			t.Error("Float() should error for non-numeric string")
		}
	})
}
