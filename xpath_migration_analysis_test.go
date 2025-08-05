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

	t.Run("当前点号风格示例", func(t *testing.T) {
		// 当前支持的点号风格
		name, err := doc.Query("user.profile.name").String()
		if err != nil {
			t.Errorf("Current dot style query failed: %v", err)
		} else if name != "John" {
			t.Errorf("Expected 'John', got '%s'", name)
		}

		// 当前的数组访问
		firstOrderTotal, err := doc.Query("user.orders[0].total").Float()
		if err != nil {
			t.Errorf("Current array access failed: %v", err)
		} else if firstOrderTotal != 99.99 {
			t.Errorf("Expected 99.99, got %f", firstOrderTotal)
		}
	})

	t.Run("期望的XPath风格示例", func(t *testing.T) {
		// 期望支持的 XPath 风格（目前不工作）

		// 基本路径访问：/user/profile/name
		xpathName := doc.Query("/user/profile/name")
		t.Logf("XPath style '/user/profile/name' exists: %t", xpathName.Exists())

		// 数组访问：/user/orders[0]/total
		xpathOrderTotal := doc.Query("/user/orders[0]/total")
		t.Logf("XPath style '/user/orders[0]/total' exists: %t", xpathOrderTotal.Exists())

		// 根级访问：/config/version
		xpathVersion := doc.Query("/config/version")
		t.Logf("XPath style '/config/version' exists: %t", xpathVersion.Exists())

		// 递归查询仍然使用 //（这个应该已经工作）
		allNames := doc.Query("//name")
		t.Logf("Recursive query '//name' found %d results", allNames.Count())
	})

	t.Run("兼容性检查", func(t *testing.T) {
		// 检查是否需要保持向后兼容
		t.Log("当前代码使用点号风格，迁移到 XPath 风格需要考虑：")
		t.Log("1. 是否保持向后兼容（同时支持两种风格）")
		t.Log("2. 或者完全迁移到 XPath 风格")
		t.Log("3. 测试用例的迁移策略")
	})
}

// TestProposedXPathChanges 展示需要进行的具体改动
func TestProposedXPathChanges(t *testing.T) {
	t.Run("路径分割改动", func(t *testing.T) {
		// 当前：strings.Split(path, ".")
		// 改为：strings.Split(path, "/")

		currentPaths := []string{
			"user.profile.name",
			"user.orders[0].total",
			"config.features[1]",
		}

		expectedXPathPaths := []string{
			"/user/profile/name",
			"/user/orders[0]/total",
			"/config/features[1]",
		}

		t.Logf("当前路径格式: %v", currentPaths)
		t.Logf("期望XPath格式: %v", expectedXPathPaths)
	})

	t.Run("特殊情况处理", func(t *testing.T) {
		specialCases := map[string]string{
			"根路径访问": "/ -> 返回整个文档",
			"递归查询":  "//name -> 查找所有name字段",
			"空路径":   "'' -> 返回整个文档",
			"相对路径":  "name -> 直接字段访问",
			"绝对路径":  "/user/name -> 从根开始的路径",
		}

		for scenario, description := range specialCases {
			t.Logf("%s: %s", scenario, description)
		}
	})
}

// TestMigrationStrategy 展示迁移策略选项
func TestMigrationStrategy(t *testing.T) {
	t.Run("选项1_完全迁移", func(t *testing.T) {
		t.Log("完全迁移到 XPath 风格：")
		t.Log("优点：API一致性，符合XPath标准")
		t.Log("缺点：破坏性改动，需要更新所有现有代码")
		t.Log("需要改动：")
		t.Log("- 将所有 strings.Split(path, '.') 改为 strings.Split(path, '/')")
		t.Log("- 处理根路径 '/' 的特殊情况")
		t.Log("- 更新所有测试用例")
		t.Log("- 更新文档和示例")
	})

	t.Run("选项2_向后兼容", func(t *testing.T) {
		t.Log("同时支持两种风格：")
		t.Log("优点：不破坏现有代码，渐进式迁移")
		t.Log("缺点：代码复杂度增加，维护负担")
		t.Log("实现方式：")
		t.Log("- 检查路径是否以 '/' 开头来决定使用哪种解析方式")
		t.Log("- 保留现有的点号解析逻辑")
		t.Log("- 添加新的斜杠解析逻辑")
	})

	t.Run("选项3_配置选项", func(t *testing.T) {
		t.Log("通过配置选择路径风格：")
		t.Log("优点：用户可以选择偏好的风格")
		t.Log("缺点：API复杂度增加")
		t.Log("实现方式：")
		t.Log("- 添加 ParseOptions 结构")
		t.Log("- 提供 PathStyle 枚举（DotStyle, XPathStyle）")
		t.Log("- 在 Document 中存储选择的风格")
	})

	t.Run("推荐方案", func(t *testing.T) {
		t.Log("推荐：选项2（向后兼容）")
		t.Log("理由：")
		t.Log("1. 不会破坏现有用户的代码")
		t.Log("2. 可以渐进式迁移")
		t.Log("3. 实现相对简单")
		t.Log("4. 未来可以标记点号风格为deprecated")
	})
}
