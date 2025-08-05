package xjson

import (
	"testing"
)

func TestXPathVsDoNotationComparison(t *testing.T) {
	jsonStr := `{
		"users": [
			{
				"id": 1,
				"profile": {
					"name": "Alice",
					"age": 30,
					"settings": {
						"theme": "dark",
						"notifications": true
					}
				},
				"orders": [
					{"id": 101, "total": 99.99},
					{"id": 102, "total": 49.99}
				]
			},
			{
				"id": 2,
				"profile": {
					"name": "Bob",
					"age": 25,
					"settings": {
						"theme": "light",
						"notifications": false
					}
				},
				"orders": [
					{"id": 201, "total": 199.99}
				]
			}
		],
		"config": {
			"version": "1.0",
			"features": ["auth", "payments", "notifications"]
		}
	}`

	x, err := ParseString(jsonStr)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	t.Run("简单字段访问对比", func(t *testing.T) {
		testCases := []struct {
			dotNotation   string
			xpathNotation string
			description   string
		}{
			{"config.version", "/config/version", "访问配置版本"},
			{"users[0].id", "/users[0]/id", "访问第一个用户ID"},
			{"users[0].profile.name", "/users[0]/profile/name", "访问用户名"},
			{"users[1].orders[0].total", "/users[1]/orders[0]/total", "访问订单总额"},
		}

		for _, tc := range testCases {
			dotResult := x.Query(tc.dotNotation)
			xpathResult := x.Query(tc.xpathNotation)

			if dotResult.Exists() != xpathResult.Exists() {
				t.Errorf("%s: 存在性不一致 - 点号:%v, XPath:%v",
					tc.description, dotResult.Exists(), xpathResult.Exists())
			}

			if dotResult.Exists() && xpathResult.Exists() {
				dotStr, _ := dotResult.String()
				xpathStr, _ := xpathResult.String()

				if dotStr != xpathStr {
					t.Errorf("%s: 结果不一致 - 点号:'%s', XPath:'%s'",
						tc.description, dotStr, xpathStr)
				}
			}
		}
	})

	t.Run("数组操作对比", func(t *testing.T) {
		testCases := []struct {
			dotNotation   string
			xpathNotation string
			description   string
		}{
			{"users[0]", "/users[0]", "第一个用户"},
			{"users[-1]", "/users[-1]", "最后一个用户"},
			{"config.features[0]", "/config/features[0]", "第一个功能"},
			{"config.features[-1]", "/config/features[-1]", "最后一个功能"},
		}

		for _, tc := range testCases {
			dotResult := x.Query(tc.dotNotation)
			xpathResult := x.Query(tc.xpathNotation)

			if dotResult.Exists() != xpathResult.Exists() {
				t.Errorf("%s: 存在性不一致", tc.description)
			}
		}
	})

	t.Run("数组切片对比", func(t *testing.T) {
		testCases := []struct {
			dotNotation   string
			xpathNotation string
			description   string
		}{
			{"users[0:2]", "/users[0:2]", "用户切片"},
			{"config.features[0:2]", "/config/features[0:2]", "功能切片"},
			{"users[0].orders[0:1]", "/users[0]/orders[0:1]", "订单切片"},
		}

		for _, tc := range testCases {
			dotResult := x.Query(tc.dotNotation)
			xpathResult := x.Query(tc.xpathNotation)

			if dotResult.Exists() != xpathResult.Exists() {
				t.Errorf("%s: 存在性不一致", tc.description)
			}

			if dotResult.Count() != xpathResult.Count() {
				t.Errorf("%s: 结果数量不一致 - 点号:%d, XPath:%d",
					tc.description, dotResult.Count(), xpathResult.Count())
			}
		}
	})

	t.Run("XPath 独有特性", func(t *testing.T) {
		// 根路径访问
		rootResult := x.Query("/")
		if !rootResult.Exists() {
			t.Errorf("根路径 / 应该存在")
		}

		// 递归搜索（点号语法不支持）
		allNames := x.Query("//name")
		if !allNames.Exists() {
			t.Errorf("递归搜索 //name 应该找到结果")
		} else if allNames.Count() != 2 {
			t.Errorf("应该找到 2 个 name 字段，实际找到 %d 个", allNames.Count())
		}

		// 搜索所有 id 字段
		allIds := x.Query("//id")
		if !allIds.Exists() {
			t.Errorf("递归搜索 //id 应该找到结果")
		} else {
			// 应该找到：2个用户ID + 3个订单ID = 5个
			expectedCount := 5
			if allIds.Count() != expectedCount {
				t.Errorf("应该找到 %d 个 id 字段，实际找到 %d 个", expectedCount, allIds.Count())
			}
		}

		// 搜索所有 total 字段
		allTotals := x.Query("//total")
		if !allTotals.Exists() {
			t.Errorf("递归搜索 //total 应该找到结果")
		} else if allTotals.Count() != 3 {
			t.Errorf("应该找到 3 个 total 字段，实际找到 %d 个", allTotals.Count())
		}
	})

	t.Run("错误处理对比", func(t *testing.T) {
		errorCases := []struct {
			dotNotation   string
			xpathNotation string
			description   string
		}{
			{"nonexistent", "/nonexistent", "不存在的字段"},
			{"users[999]", "/users[999]", "数组越界"},
			{"users[-999]", "/users[-999]", "负数越界"},
			{"users[0].nonexistent", "/users[0]/nonexistent", "嵌套不存在字段"},
		}

		for _, tc := range errorCases {
			dotResult := x.Query(tc.dotNotation)
			xpathResult := x.Query(tc.xpathNotation)

			if dotResult.Exists() || xpathResult.Exists() {
				t.Errorf("%s: 应该都不存在，但点号存在:%v, XPath存在:%v",
					tc.description, dotResult.Exists(), xpathResult.Exists())
			}
		}
	})
}
