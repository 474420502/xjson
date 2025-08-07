package xjson

import (
	"testing"
)

func TestXPathCoverage(t *testing.T) {
	jsonData := `{
		"a": {
			"b": [
				{"c": 1},
				{"c": 2},
				{"d": 3}
			]
		},
		"x": [
			{"y": {"z": 1}},
			{"y": {"z": 2}}
		]
	}`

	doc, err := ParseString(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Test nested filters on an array, which should not match but trigger a warning.
	result := doc.Query("//b[@c=2]")
	if result.Exists() {
		t.Error("Expected nested filter on an array itself to find no match")
	}

	// Correctly query the items within the array.
	result = doc.Query("//b/*[@c=2]")
	if !result.Exists() {
		t.Error("Expected nested filter to find a match when querying items")
	}
	t.Error(result.Value())
	c, _ := result.Get("c").Int()
	if c != 2 {
		t.Errorf("Expected c=2, got %d", c)
	}

	// Test wildcard with filters
	result = doc.Query("/a/b/*[c=1]")
	if !result.Exists() {
		t.Error("Expected wildcard with filter to find a match")
	}
	c, _ = result.Get("c").Int()
	if c != 1 {
		t.Errorf("Expected c=1, got %d", c)
	}
}
