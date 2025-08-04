package xjson

import (
	"testing"
)

func TestAdvancedXPathQueries(t *testing.T) {
	// 更复杂的测试数据
	jsonData := `{
		"store": {
			"book": [
				{
					"category": "reference",
					"author": "Nigel Rees",
					"title": "Sayings of the Century",
					"price": 8.95,
					"available": true
				},
				{
					"category": "fiction",
					"author": "Evelyn Waugh", 
					"title": "Sword of Honour",
					"price": 12.99,
					"available": false
				},
				{
					"category": "fiction",
					"author": "Herman Melville",
					"title": "Moby Dick",
					"price": 8.99,
					"available": true
				},
				{
					"category": "fiction",
					"author": "J. R. R. Tolkien",
					"title": "The Lord of the Rings",
					"price": 22.99,
					"available": true
				}
			],
			"bicycle": {
				"color": "red",
				"price": 19.95
			}
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
			query:    "store.book",
			expected: 4,
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
			query:    "store.book[0].title",
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
			query:    "store.book[-1].title",
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
			query:    "store.book[1:3]",
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
			query:    "store..price",
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

func TestFilterExpressions(t *testing.T) {
	// 为过滤器表达式准备的测试数据
	jsonData := `{
		"products": [
			{"id": 1, "name": "Laptop", "price": 999.99, "category": "electronics", "inStock": true},
			{"id": 2, "name": "Mouse", "price": 25.50, "category": "electronics", "inStock": true},
			{"id": 3, "name": "Desk", "price": 150.00, "category": "furniture", "inStock": false},
			{"id": 4, "name": "Chair", "price": 85.00, "category": "furniture", "inStock": true},
			{"id": 5, "name": "Book", "price": 12.99, "category": "education", "inStock": true}
		],
		"users": [
			{"name": "Alice", "age": 30, "active": true},
			{"name": "Bob", "age": 25, "active": false},
			{"name": "Charlie", "age": 35, "active": true}
		]
	}`

	doc, err := ParseString(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// 当前这些测试会失败，因为我们还没有实现过滤器
	// 但这些测试定义了我们想要实现的功能
	t.Run("price_filter_less_than", func(t *testing.T) {
		// 查找价格小于100的产品
		result := doc.Query("products[?(@.price < 100)]")
		if !result.Exists() {
			t.Error("Expected results for price < 100 filter")
			return
		}

		// 应该找到3个产品: Mouse (25.50), Chair (85.00), Book (12.99)
		count := result.Count()
		if count != 3 {
			t.Errorf("Expected 3 products with price < 100, got %d", count)
		}
	})

	t.Run("category_filter_equals", func(t *testing.T) {
		// 查找电子产品
		result := doc.Query("products[?(@.category == 'electronics')]")
		if !result.Exists() {
			t.Error("Expected results for category == 'electronics' filter")
			return
		}

		// 应该找到2个电子产品
		count := result.Count()
		if count != 2 {
			t.Errorf("Expected 2 electronics products, got %d", count)
		}
	})

	t.Run("boolean_filter", func(t *testing.T) {
		// 查找有库存的产品
		result := doc.Query("products[?(@.inStock == true)]")
		if !result.Exists() {
			t.Error("Expected results for inStock == true filter")
			return
		}

		count := result.Count()
		if count != 4 {
			t.Errorf("Expected 4 products in stock, got %d", count)
		}
	})

	t.Run("complex_filter_and", func(t *testing.T) {
		// 查找价格小于100且有库存的产品
		result := doc.Query("products[?(@.price < 100 && @.inStock == true)]")
		if !result.Exists() {
			t.Error("Expected results for complex AND filter")
			return
		}

		count := result.Count()
		if count != 3 {
			t.Errorf("Expected 3 products matching complex filter, got %d", count)
		}
	})

	t.Run("complex_filter_or", func(t *testing.T) {
		// 查找价格大于500或者类别是教育的产品
		result := doc.Query("products[?(@.price > 500 || @.category == 'education')]")
		count := result.Count()
		if count != 2 {
			t.Errorf("Expected 2 products matching OR filter, got %d", count)
		}
	})
}

func TestRecursiveQueries(t *testing.T) {
	jsonData := `{
		"company": {
			"name": "TechCorp",
			"departments": [
				{
					"name": "Engineering",
					"employees": [
						{"name": "Alice", "salary": 90000},
						{"name": "Bob", "salary": 85000}
					]
				},
				{
					"name": "Sales", 
					"employees": [
						{"name": "Charlie", "salary": 70000},
						{"name": "Diana", "salary": 75000}
					]
				}
			]
		}
	}`

	doc, err := ParseString(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	t.Run("recursive_name_search", func(t *testing.T) {
		t.Skip("Recursive queries not yet implemented")

		// 查找所有 name 字段（递归）
		result := doc.Query("//name")
		count := result.Count()
		// 应该找到: company.name, departments[0].name, departments[1].name,
		// 以及所有employee的name (4个)，总共7个
		if count != 7 {
			t.Errorf("Expected 7 name fields, got %d", count)
		}
	})

	t.Run("recursive_with_filter", func(t *testing.T) {
		t.Skip("Recursive queries and filters not yet implemented")

		// 递归查找所有薪水大于80000的员工
		result := doc.Query("//employees[?(@.salary > 80000)]")
		count := result.Count()
		if count != 2 {
			t.Errorf("Expected 2 high-salary employees, got %d", count)
		}
	})
}
