package xjson

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"reflect"
	"testing"
)

func init() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
}

func TestXJSONCoverage(t *testing.T) {
	// Comprehensive JSON for testing various features
	jsonData := `{
		"string": "hello",
		"int": 42,
		"float": 3.14,
		"bool_true": true,
		"bool_false": false,
		"null": null,
		"array": [1, "two", 3.0],
		"object": {
			"nested_key": "nested_value"
		}
	}`
	doc, err := ParseString(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Test Must functions
	if doc.Query("/string").MustString() != "hello" {
		t.Error("MustString failed")
	}
	if doc.Query("/int").MustInt() != 42 {
		t.Error("MustInt failed")
	}
	if doc.Query("/int").MustInt64() != 42 {
		t.Error("MustInt64 failed")
	}
	if doc.Query("/float").MustFloat() != 3.14 {
		t.Error("MustFloat failed")
	}
	if !doc.Query("/bool_true").MustBool() {
		t.Error("MustBool failed for true")
	}

	// Test Value and Values
	val := doc.Query("/string").Value()
	if val.(string) != "hello" {
		t.Errorf("Value() expected 'hello', got %v", val)
	}
	vals := doc.Query("/array/*").Values()
	if len(vals) != 3 {
		t.Fatalf("Values() expected 3 items, got %d", len(vals))
	}
	if vals[0].(float64) != 1 || vals[1].(string) != "two" || vals[2].(float64) != 3.0 {
		t.Errorf("Values() returned incorrect data: %v", vals)
	}

	// Test Map and Filter
	arrResult := doc.Query("/array")
	mapped, err := arrResult.Map(func(idx int, item IResult) (interface{}, error) {
		if item.IsNumber() {
			return item.Float()
		}
		return item.String()
	})
	if err != nil {
		t.Fatalf("Map() failed: %v", err)
	}
	expectedMap := []interface{}{1.0, "two", 3.0}
	if !reflect.DeepEqual(mapped, expectedMap) {
		t.Errorf("Map() expected %v, got %v", expectedMap, mapped)
	}

	filtered, err := arrResult.Filter(func(idx int, item IResult) (bool, error) {
		return item.IsNumber(), nil
	})
	if err != nil {
		t.Fatalf("Filter() failed: %v", err)
	}
	if filtered.Count() != 2 {
		t.Errorf("Filter() expected 2 items, got %d", filtered.Count())
	}
}
