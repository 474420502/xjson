package xjson

import (
	"testing"
)

func TestXPathStandardSyntax(t *testing.T) {
	jsonStr := `{
		"store": {
			"book": [
				{"title": "Book 0", "price": 10.5},
				{"title": "Book 1", "price": 15.0},
				{"title": "Book 2", "price": 20.0},
				{"title": "Book 3", "price": 25.0}
			],
			"bicycle": {
				"color": "red",
				"price": 19.95
			}
		},
		"expensive": 10
	}`

	x, err := ParseString(jsonStr)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	t.Run("基本 XPath 路径访问", func(t *testing.T) {
		// 测试基本的绝对路径：/store/book[0]/title
		result := x.Query("/store/book[0]/title")
		if !result.Exists() {
			t.Errorf("XPath /store/book[0]/title should exist")
		} else {
			title, err := result.String()
			if err != nil {
				t.Errorf("Failed to get string: %v", err)
			} else if title != "Book 0" {
				t.Errorf("Expected 'Book 0', got '%s'", title)
			}
		}
	})

	t.Run("XPath 数组访问", func(t *testing.T) {
		// 负数索引：/store/book[-1]/title
		result := x.Query("/store/book[-1]/title")
		if !result.Exists() {
			t.Errorf("XPath /store/book[-1]/title should exist")
		} else {
			title, err := result.String()
			if err != nil {
				t.Errorf("Failed to get string: %v", err)
			} else if title != "Book 3" {
				t.Errorf("Expected 'Book 3', got '%s'", title)
			}
		}

		// 数组切片：/store/book[1:3]
		result = x.Query("/store/book[1:3]")
		if !result.Exists() {
			t.Errorf("XPath /store/book[1:3] should exist")
		} else if result.Count() != 2 {
			t.Errorf("Expected 2 results, got %d", result.Count())
		}
	})

	t.Run("递归查询 //", func(t *testing.T) {
		// 查找所有 price 字段
		result := x.Query("//price")
		if !result.Exists() {
			t.Errorf("Recursive query //price should find results")
		} else {
			count := result.Count()
			// 应该找到 5 个 price：4本书 + 1个自行车
			if count != 5 {
				t.Errorf("Expected 5 price fields, got %d", count)
			}
		}

		// 查找所有 title 字段
		result = x.Query("//title")
		if !result.Exists() {
			t.Errorf("Recursive query //title should find results")
		} else {
			count := result.Count()
			// 应该找到 4 个 title
			if count != 4 {
				t.Errorf("Expected 4 title fields, got %d", count)
			}
		}
	})

	t.Run("对比点号语法与 XPath 语法", func(t *testing.T) {
		// 点号语法
		dotResult := x.Query("store.book[0].title")
		// XPath 语法
		xpathResult := x.Query("/store/book[0]/title")

		if dotResult.Exists() != xpathResult.Exists() {
			t.Errorf("Dot notation and XPath should have same existence result")
		}

		if dotResult.Exists() {
			dotStr, _ := dotResult.String()
			xpathStr, _ := xpathResult.String()

			if dotStr != xpathStr {
				t.Errorf("Dot notation result '%s' != XPath result '%s'", dotStr, xpathStr)
			}
		}
	})
}
