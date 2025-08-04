package xjson

import (
	"testing"
)

func TestSimpleBooleanFilter(t *testing.T) {
	jsonData := `{
		"products": [
			{"inStock": 1},
			{"inStock": 0}
		]
	}`

	doc, err := ParseString(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	result := doc.Query("products[?(@.inStock == true)]")
	if result.Count() != 1 {
		t.Error("Expected 1 product in stock, got", result.Count())
	}
}
