package xjson

import (
	"testing"
)

// TestXPathStyleMigrationExample 展示从点号风格到 XPath 风格的迁移示例
func TestXPathStyleMigrationExample(t *testing.T) {
	jsonData := `{
		"user": {
			"profile": {
				"name": "John",
				"settings": {
					"theme": "dark",
					"notifications": true
				}
			},
			"orders": [
				{"id": 1, "total": 99.99},
				{"id": 2, "total": 149.50}
			]
		},
		"config": {
			"version": "1.0",
			"features": ["search", "recommendations"]
		}
	}`

	doc, err := ParseString(jsonData)
	if err != nil {
		t.Fatalf("ParseString failed: %v", err)
	}

	t.Run("XPath风格示例", func(t *testing.T) {
		// 基本路径访问：/user/profile/name
		xpathName, err := doc.Query("/user/profile/name").String()
		if err != nil {
			t.Errorf("XPath style query failed: %v", err)
		} else if xpathName != "John" {
			t.Errorf("Expected 'John', got '%s'", xpathName)
		}

		// 数组访问：/user/orders[0]/total
		xpathOrderTotal, err := doc.Query("/user/orders[0]/total").Float()
		if err != nil {
			t.Errorf("XPath style array access failed: %v", err)
		} else if xpathOrderTotal != 99.99 {
			t.Errorf("Expected 99.99, got %f", xpathOrderTotal)
		}

		// 根级访问：/config/version
		xpathVersion, err := doc.Query("/config/version").String()
		if err != nil {
			t.Errorf("XPath style root access failed: %v", err)
		} else if xpathVersion != "1.0" {
			t.Errorf("Expected '1.0', got '%s'", xpathVersion)
		}

		// 递归查询仍然使用 //
		allNames := doc.Query("//name")
		if allNames.Count() != 1 {
			t.Errorf("Recursive query '//name' found %d results, expected 1", allNames.Count())
		}
	})
}
