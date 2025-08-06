package xjson

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestCoverageBoost(t *testing.T) {
	t.Run("MustString_panic_on_error", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("MustString should have panicked on error")
			}
		}()
		result := &Result{err: ErrNotFound}
		result.MustString()
	})

	t.Run("Query_with_various_paths", func(t *testing.T) {
		jsonData := `{"a":{"b":1}, "arr":[{"x":10}, {"y":20}], "obj":{"nested":"value"}}`
		doc, err := ParseString(jsonData)
		if err != nil {
			t.Fatal(err)
		}

		// Simple path
		if doc.Query("/a/b").MustInt() != 1 {
			t.Error("Simple path /a/b failed")
		}

		// Array path
		if doc.Query("/arr[0]/x").MustInt() != 10 {
			t.Error("Array path /arr[0]/x failed")
		}

		// Non-existent
		if doc.Query("/a/c").Exists() {
			t.Error("Querying non-existent path /a/c should not exist")
		}

		// Test direct object access in result
		result := doc.Query("/obj")
		if !result.IsObject() {
			t.Error("Querying /obj should return an object")
		}
		if result.Get("nested").MustString() != "value" {
			t.Error("Get() on result object failed")
		}
	})

	t.Run("Result_Raw_and_Bytes", func(t *testing.T) {
		jsonData := `{"key": "value", "num": 123}`
		doc, err := ParseString(jsonData)
		if err != nil {
			t.Fatal(err)
		}

		// Raw on object
		raw := doc.Query("/").Raw()
		if _, ok := raw.(map[string]interface{}); !ok {
			t.Errorf("Raw() should return a map, got %T", raw)
		}

		// Raw on primitive
		raw = doc.Query("/key").Raw()
		if val, ok := raw.(string); !ok || val != "value" {
			t.Errorf("Raw() on string failed, got %v", raw)
		}

		// Bytes on object
		bytes, err := doc.Query("/").Bytes()
		if err != nil {
			t.Errorf("Bytes() failed: %v", err)
		}
		var m map[string]interface{}
		if json.Unmarshal(bytes, &m) != nil || m["key"] != "value" {
			t.Errorf("Bytes() did not return correct JSON: %s", string(bytes))
		}
	})

	t.Run("Result_Filter", func(t *testing.T) {
		jsonData := `[1, 2, 3, 4, 5]`
		doc, err := ParseString(jsonData)
		if err != nil {
			t.Fatal(err)
		}

		result := doc.Query("/")
		filtered := result.Filter(func(index int, value IResult) bool {
			return value.MustInt() > 3
		})

		if filtered.Count() != 2 {
			t.Errorf("Filter should return 2 items, got %d", filtered.Count())
		}
		expected := []interface{}{4.0, 5.0} // Numbers are decoded as float64
		if !reflect.DeepEqual(filtered.Raw(), expected) {
			t.Errorf("Filter result is incorrect. Got %v", filtered.Raw())
		}
	})
}
