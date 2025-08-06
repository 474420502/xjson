package xjson

import (
	"encoding/json"
	"reflect"
	"testing"
)

// TestMigratedMegaFinalCoverage contains migrated tests from mega_final_coverage_test.go
func TestMigratedMegaFinalCoverage(t *testing.T) {
	jsonStr := `{
		"object": {
			"field1": "simple dotted path",
			"field2": {
				"nested": "deeper dotted path",
				"array": [1, 2]
			}
		},
		"with.dots": "dotted_key_value",
		"number": 123,
		"direct": "direct field access",
		"boolean": true,
		"null": null
	}`

	doc, err := ParseString(jsonStr)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	testCases := []struct {
		name     string
		path     string
		expected string
	}{
		{"direct field access", "/direct", "direct field access"},
		{"direct dotted key access", `/"with.dots"`, "dotted_key_value"},
		{"direct number access", "/number", "123"},
		{"simple path", "/object/field1", "simple dotted path"},
		{"path to object", "/object/field2", `{"array":[1,2],"nested":"deeper dotted path"}`},
		{"deeper path", "/object/field2/nested", "deeper dotted path"},
		{"path to array", "/object/field2/array", "[1,2]"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := doc.Query(tc.path)
			if !result.Exists() {
				t.Errorf("Expected path '%s' to exist, but it didn't", tc.path)
				return
			}
			val, _ := result.String()
			if val != tc.expected {
				t.Errorf("Expected value '%s' for path '%s', but got '%s'", tc.expected, tc.path, val)
			}
		})
	}
}

func TestMigratedModifierSet(t *testing.T) {
	tests := []struct {
		name         string
		initialJSON  string
		path         string
		value        interface{}
		expectedJSON string
		expectError  bool
	}{
		{
			name:         "set simple property",
			initialJSON:  `{"name": "John"}`,
			path:         "/name",
			value:        "Jane",
			expectedJSON: `{"name":"Jane"}`,
		},
		{
			name:         "set nested property",
			initialJSON:  `{"user": {"name": "John"}}`,
			path:         "/user/name",
			value:        "Jane",
			expectedJSON: `{"user":{"name":"Jane"}}`,
		},
		{
			name:         "create new property",
			initialJSON:  `{"name": "John"}`,
			path:         "/age",
			value:        30,
			expectedJSON: `{"age":30,"name":"John"}`,
		},
		{
			name:         "create nested property",
			initialJSON:  `{}`,
			path:         "/user/name",
			value:        "John",
			expectedJSON: `{"user":{"name":"John"}}`,
		},
		{
			name:         "set array element",
			initialJSON:  `{"items": ["a", "b", "c"]}`,
			path:         "/items[1]",
			value:        "new_b",
			expectedJSON: `{"items":["a","new_b","c"]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := ParseString(tt.initialJSON)
			if err != nil {
				t.Fatalf("Failed to parse initial JSON: %v", err)
			}

			err = doc.Set(tt.path, tt.value)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Compare JSON strings semantically
			var expected, actual interface{}
			if err := json.Unmarshal([]byte(tt.expectedJSON), &expected); err != nil {
				t.Fatalf("Failed to unmarshal expected JSON: %v", err)
			}
			actualStr, err := doc.String()
			if err != nil {
				t.Fatalf("Failed to get string from doc: %v", err)
			}
			if err := json.Unmarshal([]byte(actualStr), &actual); err != nil {
				t.Fatalf("Failed to unmarshal actual JSON: %v", err)
			}

			if !reflect.DeepEqual(expected, actual) {
				t.Errorf("Expected JSON %s, got %s", tt.expectedJSON, actualStr)
			}
		})
	}
}

func TestMigratedAdvancedQueries(t *testing.T) {
	jsonData := `{
		"store": {
			"book": [
				{ "category": "reference", "author": "Nigel Rees", "title": "Sayings of the Century", "price": 8.95, "available": true },
				{ "category": "fiction", "author": "Evelyn Waugh", "title": "Sword of Honour", "price": 12.99, "available": false },
				{ "category": "fiction", "author": "Herman Melville", "title": "Moby Dick", "price": 8.99, "available": true },
				{ "category": "fiction", "author": "J. R. R. Tolkien", "title": "The Lord of the Rings", "price": 22.99, "available": true }
			],
			"bicycle": { "color": "red", "price": 19.95 }
		}
	}`

	doc, err := ParseString(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	tests := []struct {
		name     string
		query    string
		expected int // Expected number of results
		validate func(t *testing.T, result IResult)
	}{
		{
			name:     "all books",
			query:    "/store/book",
			expected: 1, // The result is a single array containing 4 books
			validate: func(t *testing.T, result IResult) {
				if !result.IsArray() {
					t.Error("Expected array result")
				}
				if result.Count() != 4 {
					t.Errorf("Expected 4 books, got %d", result.Count())
				}
			},
		},
		{
			name:     "first book title",
			query:    "/store/book[0]/title",
			expected: 1,
			validate: func(t *testing.T, result IResult) {
				title := result.MustString()
				if title != "Sayings of the Century" {
					t.Errorf("Expected 'Sayings of the Century', got '%s'", title)
				}
			},
		},
		{
			name:     "last book by negative index",
			query:    "/store/book[-1]/title",
			expected: 1,
			validate: func(t *testing.T, result IResult) {
				title := result.MustString()
				if title != "The Lord of the Rings" {
					t.Errorf("Expected 'The Lord of the Rings', got '%s'", title)
				}
			},
		},
		{
			name:     "books slice [1:3]",
			query:    "/store/book[1:3]",
			expected: 2,
			validate: func(t *testing.T, result IResult) {
				if !result.IsArray() {
					t.Error("Expected array result")
				}
				if result.Count() != 2 {
					t.Errorf("Expected 2 books, got %d", result.Count())
				}
			},
		},
		{
			name:     "all prices",
			query:    "/store//price",
			expected: 5, // 4 books + 1 bicycle
			validate: func(t *testing.T, result IResult) {
				if !result.IsArray() {
					t.Error("Expected array result")
				}
				count := 0
				result.ForEach(func(index int, value IResult) bool {
					price := value.MustFloat()
					if price <= 0 {
						t.Errorf("Invalid price at index %d: %f", index, price)
					}
					count++
					return true
				})
				if count != 5 {
					t.Errorf("Expected 5 prices, got %d", count)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := doc.Query(tt.query)
			if !result.Exists() && tt.expected > 0 {
				t.Errorf("Query '%s' returned no results", tt.query)
				return
			}

			if tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestMigratedFilterExpressions(t *testing.T) {
	jsonData := `{
		"products": [
			{"id": 1, "name": "Laptop", "price": 999.99, "category": "electronics", "inStock": true},
			{"id": 2, "name": "Mouse", "price": 25.50, "category": "electronics", "inStock": true},
			{"id": 3, "name": "Desk", "price": 150.00, "category": "furniture", "inStock": false},
			{"id": 4, "name": "Chair", "price": 85.00, "category": "furniture", "inStock": true},
			{"id": 5, "name": "Book", "price": 12.99, "category": "education", "inStock": true}
		]
	}`

	doc, err := ParseString(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	t.Run("price_filter_less_than", func(t *testing.T) {
		result := doc.Query("/products[?(@.price < 100)]")
		if !result.Exists() {
			t.Error("Expected results for price < 100 filter")
			return
		}
		if count := result.Count(); count != 3 {
			t.Errorf("Expected 3 products with price < 100, got %d", count)
		}
	})

	t.Run("category_filter_equals", func(t *testing.T) {
		result := doc.Query("/products[?(@.category == 'electronics')]")
		if !result.Exists() {
			t.Error("Expected results for category == 'electronics' filter")
			return
		}
		if count := result.Count(); count != 2 {
			t.Errorf("Expected 2 electronics products, got %d", count)
		}
	})

	t.Run("boolean_filter", func(t *testing.T) {
		result := doc.Query("/products[?(@.inStock == true)]")
		if !result.Exists() {
			t.Error("Expected results for inStock == true filter")
			return
		}
		if count := result.Count(); count != 4 {
			t.Errorf("Expected 4 products in stock, got %d", count)
		}
	})

	t.Run("complex_filter_and", func(t *testing.T) {
		result := doc.Query("/products[?(@.price < 100 && @.inStock == true)]")
		if !result.Exists() {
			t.Error("Expected results for complex AND filter")
			return
		}
		if count := result.Count(); count != 3 {
			t.Errorf("Expected 3 products matching complex filter, got %d", count)
		}
	})

	t.Run("complex_filter_or", func(t *testing.T) {
		result := doc.Query("/products[?(@.price > 500 || @.category == 'education')]")
		if count := result.Count(); count != 2 {
			t.Errorf("Expected 2 products matching OR filter, got %d", count)
		}
	})
}
