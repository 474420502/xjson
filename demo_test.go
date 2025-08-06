package xjson

import (
	"testing"
)

func TestNewFeatureDemo(t *testing.T) {
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
		},
		"company": {
			"name": "TechCorp",
			"employees": [
				{"name": "Alice", "salary": 75000, "department": "Engineering"},
				{"name": "Bob", "salary": 85000, "department": "Engineering"},
				{"name": "Charlie", "salary": 70000, "department": "Marketing"}
			]
		}
	}`

	doc, err := ParseString(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	t.Run("JSONPath_filter_expressions", func(t *testing.T) {
		// Test basic filter expression
		result := doc.Query("/store/book[price < 10]")
		count := result.Count()
		if count != 2 {
			t.Errorf("Expected 2 books with price < 10, got %d", count)
		}

		// Test boolean filter
		result = doc.Query("/store/book[inStock == true]")
		count = result.Count()
		if count != 2 {
			t.Errorf("Expected 2 books in stock, got %d", count)
		}

		// Test string filter
		result = doc.Query("/store/book[category == 'fiction']")
		count = result.Count()
		if count != 2 {
			t.Errorf("Expected 2 fiction books, got %d", count)
		}

		// Test complex AND filter
		result = doc.Query("/store/book[price < 10 && inStock == true]")
		count = result.Count()
		if count != 2 {
			t.Errorf("Expected 2 books with price < 10 AND in stock, got %d", count)
		}

		// Test complex OR filter
		result = doc.Query("/store/book[price > 12 || category == 'reference']")
		count = result.Count()
		if count != 2 {
			t.Errorf("Expected 2 books with price > 12 OR reference category, got %d", count)
		}
	})

	t.Run("recursive_queries", func(t *testing.T) {
		// Test simple recursive search for name fields
		result := doc.Query("//name")
		count := result.Count()
		if count != 4 { // company.name + 3 employee names
			t.Errorf("Expected 4 name fields, got %d", count)
		}

		// Test recursive search with filter
		result = doc.Query("//employees[salary > 80000]")
		count = result.Count()
		if count != 1 {
			t.Errorf("Expected 1 employee with salary > 80000, got %d", count)
		}

		// Test recursive search for price fields
		result = doc.Query("//price")
		count = result.Count()
		if count != 3 {
			t.Errorf("Expected 3 price fields, got %d", count)
		}
	})

	t.Run("advanced_queries", func(t *testing.T) {
		// Test array slicing
		result := doc.Query("/store/book[1:3]")
		count := result.Count()
		if count != 2 {
			t.Errorf("Expected 2 books from slice [1:3], got %d", count)
		}

		// Test negative indexing
		result = doc.Query("/store/book[-1]")
		if !result.Exists() {
			t.Error("Expected last book to exist")
		}

		// Test recursive field access for prices
		result = doc.Query("//price")
		count = result.Count()
		if count != 3 {
			t.Errorf("Expected 3 prices from recursive query, got %d", count)
		}
	})
}
