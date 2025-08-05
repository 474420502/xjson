package xjson

import (
	"testing"
)

func TestXPathAdvancedFeatures(t *testing.T) {
	jsonStr := `{
		"library": {
			"books": [
				{
					"id": 1,
					"title": "Go Programming",
					"author": "Alice",
					"price": 29.99,
					"category": "technical",
					"tags": ["programming", "golang"]
				},
				{
					"id": 2,
					"title": "Web Development",
					"author": "Bob",
					"price": 39.99,
					"category": "technical",
					"tags": ["web", "javascript"]
				},
				{
					"id": 3,
					"title": "Fiction Story",
					"author": "Carol",
					"price": 19.99,
					"category": "fiction",
					"tags": ["story", "adventure"]
				}
			],
			"magazines": [
				{
					"id": 101,
					"title": "Tech Monthly",
					"price": 9.99,
					"category": "technical"
				},
				{
					"id": 102,
					"title": "Art Weekly",
					"price": 7.99,
					"category": "art"
				}
			]
		},
		"customers": [
			{
				"name": "John",
				"age": 25,
				"orders": [
					{"bookId": 1, "quantity": 2},
					{"bookId": 2, "quantity": 1}
				]
			},
			{
				"name": "Jane",
				"age": 30,
				"orders": [
					{"bookId": 3, "quantity": 1}
				]
			}
		]
	}`

	x, err := ParseString(jsonStr)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	t.Run("XPath 绝对路径查询", func(t *testing.T) {
		// 测试根路径
		result := x.Query("/")
		if !result.Exists() {
			t.Errorf("Root path / should exist")
		}

		// 测试深层路径
		result = x.Query("/library/books[0]/title")
		if !result.Exists() {
			t.Errorf("Deep path /library/books[0]/title should exist")
		} else {
			title, _ := result.String()
			if title != "Go Programming" {
				t.Errorf("Expected 'Go Programming', got '%s'", title)
			}
		}

		// 测试不存在的路径
		result = x.Query("/nonexistent/path")
		if result.Exists() {
			t.Errorf("Nonexistent path should not exist")
		}
	})

	t.Run("XPath 数组操作", func(t *testing.T) {
		// 第一个元素
		result := x.Query("/library/books[0]/title")
		if !result.Exists() {
			t.Errorf("First book should exist")
		}

		// 最后一个元素
		result = x.Query("/library/books[-1]/title")
		if !result.Exists() {
			t.Errorf("Last book should exist")
		} else {
			title, _ := result.String()
			if title != "Fiction Story" {
				t.Errorf("Expected 'Fiction Story', got '%s'", title)
			}
		}

		// 数组切片
		result = x.Query("/library/books[0:2]")
		if !result.Exists() {
			t.Errorf("Book slice should exist")
		} else if result.Count() != 2 {
			t.Errorf("Expected 2 books in slice, got %d", result.Count())
		}

		// 负数切片
		result = x.Query("/library/books[-2:-1]")
		if !result.Exists() {
			t.Errorf("Negative slice should exist")
		} else if result.Count() != 1 {
			t.Errorf("Expected 1 book in negative slice, got %d", result.Count())
		}
	})

	t.Run("XPath 递归搜索", func(t *testing.T) {
		// 搜索所有 title 字段
		result := x.Query("//title")
		if !result.Exists() {
			t.Errorf("//title should find results")
		} else {
			count := result.Count()
			// 应该找到 3本书 + 2本杂志 = 5个title
			if count != 5 {
				t.Errorf("Expected 5 titles, got %d", count)
			}
		}

		// 搜索所有 price 字段
		result = x.Query("//price")
		if !result.Exists() {
			t.Errorf("//price should find results")
		} else {
			count := result.Count()
			// 应该找到 3本书 + 2本杂志 = 5个price
			if count != 5 {
				t.Errorf("Expected 5 prices, got %d", count)
			}
		}

		// 搜索所有 id 字段
		result = x.Query("//id")
		if !result.Exists() {
			t.Errorf("//id should find results")
		} else {
			count := result.Count()
			// 应该找到 3本书 + 2本杂志 = 5个id
			if count != 5 {
				t.Errorf("Expected 5 ids, got %d", count)
			}
		}

		// 搜索深层嵌套的字段
		result = x.Query("//bookId")
		if !result.Exists() {
			t.Errorf("//bookId should find results")
		} else {
			count := result.Count()
			// 应该找到订单中的 bookId
			if count != 3 {
				t.Errorf("Expected 3 bookIds, got %d", count)
			}
		}
	})

	t.Run("XPath 复合路径", func(t *testing.T) {
		// 嵌套数组访问
		result := x.Query("/customers[0]/orders[0]/bookId")
		if !result.Exists() {
			t.Errorf("Nested array access should work")
		} else {
			bookId, err := result.Int()
			if err != nil {
				t.Errorf("Failed to get int: %v", err)
			} else if bookId != 1 {
				t.Errorf("Expected bookId 1, got %d", bookId)
			}
		}

		// 多层路径
		result = x.Query("/library/books[1]/tags[0]")
		if !result.Exists() {
			t.Errorf("Multi-level path should work")
		} else {
			tag, _ := result.String()
			if tag != "web" {
				t.Errorf("Expected 'web', got '%s'", tag)
			}
		}
	})

	t.Run("XPath 边界情况", func(t *testing.T) {
		// 空路径
		result := x.Query("")
		if result.Exists() {
			t.Errorf("Empty path should not exist")
		}

		// 只有根斜杠
		result = x.Query("/")
		if !result.Exists() {
			t.Errorf("Root path should exist")
		}

		// 数组越界
		result = x.Query("/library/books[999]/title")
		if result.Exists() {
			t.Errorf("Out of bounds access should not exist")
		}

		// 负数越界
		result = x.Query("/library/books[-999]/title")
		if result.Exists() {
			t.Errorf("Negative out of bounds should not exist")
		}

		// 无效切片
		result = x.Query("/library/books[5:10]")
		if result.Exists() {
			t.Errorf("Invalid slice should not exist")
		}
	})
}
