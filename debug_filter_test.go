package xjson

import (
	"fmt"
	"testing"
)

func TestDebugFilterExpressions(t *testing.T) {
	jsonData := `{
		"store": {
			"book": [
				{
					"category": "reference",
					"author": "Nigel Rees",
					"title": "Sayings of the Century",
					"price": 8.95,
					"inStock": true
				},
				{
					"category": "fiction",
					"author": "Evelyn Waugh",
					"title": "Sword of Honour",
					"price": 12.99,
					"inStock": false
				},
				{
					"category": "fiction",
					"author": "Herman Melville",
					"title": "Moby Dick",
					"isbn": "0-553-21311-3",
					"price": 8.99,
					"inStock": true
				}
			]
		}
	}`

	doc, err := ParseString(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Test basic filters first
	result := doc.Query("store.book[?(@.price < 10)]")
	fmt.Printf("Books with price < 10: %d\n", result.Count())
	result.ForEach(func(index int, value IResult) bool {
		title := value.Get("title").MustString()
		price := value.Get("price").MustFloat()
		inStock := value.Get("inStock").MustBool()
		fmt.Printf("  %s: $%.2f, inStock: %t\n", title, price, inStock)
		return true
	})

	// Test inStock filter
	result = doc.Query("store.book[?(@.inStock == true)]")
	fmt.Printf("\nBooks in stock: %d\n", result.Count())
	result.ForEach(func(index int, value IResult) bool {
		title := value.Get("title").MustString()
		price := value.Get("price").MustFloat()
		inStock := value.Get("inStock").MustBool()
		fmt.Printf("  %s: $%.2f, inStock: %t\n", title, price, inStock)
		return true
	})

	// Test AND filter
	result = doc.Query("store.book[?(@.price < 10 && @.inStock == true)]")
	fmt.Printf("\nBooks with price < 10 AND in stock: %d\n", result.Count())
	result.ForEach(func(index int, value IResult) bool {
		title := value.Get("title").MustString()
		price := value.Get("price").MustFloat()
		inStock := value.Get("inStock").MustBool()
		fmt.Printf("  %s: $%.2f, inStock: %t\n", title, price, inStock)
		return true
	})
}
