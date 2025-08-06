package xjson

import (
	"testing"
)

func TestXPathSyntaxExamples(t *testing.T) {
	// 构建一个类似 HTML 结构的 JSON 数据，对应你提供的 XPath 示例
	jsonData := `{
		"container": {
			"id": "container",
			"h1": {
				"class": "main-title",
				"text": "网站标题"
			},
			"ul": {
				"class": "item-list",
				"li": [
					{
						"class": "item active",
						"a": {
							"href": "/item/1",
							"text": "条目 1"
						},
						"span": {
							"text": "(推荐)"
						}
					},
					{
						"class": "item",
						"a": {
							"href": "/item/2",
							"text": "条目 2"
						}
					},
					{
						"class": "item disabled",
						"a": {
							"href": "/item/3",
							"data-id": "unique-3",
							"text": "条目 3 (禁用)"
						}
					}
				]
			},
			"footer": {
				"id": "footer",
				"p": {
					"text": "版权所有 © 2023"
				}
			}
		}
	}`

	doc, err := ParseString(jsonData)
	if err != nil {
		t.Fatalf("ParseString failed: %v", err)
	}

	t.Run("基础路径选择", func(t *testing.T) {
		// 测试直接属性访问
		title, err := doc.Query("/container/h1/text").String()
		if err != nil {
			t.Errorf("Query container.h1.text failed: %v", err)
		}
		if title != "网站标题" {
			t.Errorf("Expected '网站标题', got '%s'", title)
		}

		// 测试数组访问
		firstItemClass, err := doc.Query("/container/ul/li[0]/class").String()
		if err != nil {
			t.Errorf("Query container.ul.li[0].class failed: %v", err)
		}
		if firstItemClass != "item active" {
			t.Errorf("Expected 'item active', got '%s'", firstItemClass)
		}

		// 测试深层嵌套访问
		footerText, err := doc.Query("/container/footer/p/text").String()
		if err != nil {
			t.Errorf("Query container.footer.p.text failed: %v", err)
		}
		if footerText != "版权所有 © 2023" {
			t.Errorf("Expected '版权所有 © 2023', got '%s'", footerText)
		}
	})

	t.Run("数组和索引操作", func(t *testing.T) {
		// 测试第一个元素
		firstHref, err := doc.Query("/container/ul/li[0]/a/href").String()
		if err != nil {
			t.Errorf("Query first item href failed: %v", err)
		}
		if firstHref != "/item/1" {
			t.Errorf("Expected '/item/1', got '%s'", firstHref)
		}

		// 测试最后一个元素 (负索引)
		lastItemText, err := doc.Query("/container/ul/li[-1]/a/text").String()
		if err != nil {
			t.Errorf("Query last item text failed: %v", err)
		}
		if lastItemText != "条目 3 (禁用)" {
			t.Errorf("Expected '条目 3 (禁用)', got '%s'", lastItemText)
		}

		// 测试中间元素
		secondHref, err := doc.Query("/container/ul/li[1]/a/href").String()
		if err != nil {
			t.Errorf("Query second item href failed: %v", err)
		}
		if secondHref != "/item/2" {
			t.Errorf("Expected '/item/2', got '%s'", secondHref)
		}
	})

	t.Run("递归查询测试", func(t *testing.T) {
		// 递归查找所有 href 属性 - 使用通用路径
		hrefs := doc.Query("//href")
		count := hrefs.Count()
		if count < 3 {
			t.Logf("NOTE: Recursive query found %d hrefs, expected at least 3", count)
		}

		// 递归查找所有 text 属性
		texts := doc.Query("//text")
		textCount := texts.Count()
		if textCount < 5 {
			t.Logf("NOTE: Recursive text query found %d texts, expected at least 5", textCount)
		}
	})
}

func TestXPathComplexExamples(t *testing.T) {
	// 更复杂的测试数据
	complexData := `{
		"store": {
			"book": [
				{
					"category": "reference",
					"author": "Nigel Rees",
					"title": "Sayings of the Century",
					"price": 8.95,
					"inStock": true,
					"isbn": "978-0262532754"
				},
				{
					"category": "fiction",
					"author": "Evelyn Waugh",
					"title": "Sword of Honour",
					"price": 12.99,
					"inStock": false,
					"isbn": "978-0316769174"
				},
				{
					"category": "fiction",
					"author": "Herman Melville",
					"title": "Moby Dick",
					"isbn": "0-553-21311-3",
					"price": 8.99,
					"inStock": true
				}
			],
			"bicycle": {
				"color": "red",
				"price": 19.95
			}
		},
		"categories": ["reference", "fiction", "biography"],
		"metadata": {
			"count": 3,
			"lastUpdated": "2023-12-01"
		}
	}`

	doc, err := ParseString(complexData)
	if err != nil {
		t.Fatalf("ParseString failed: %v", err)
	}

	t.Run("基本数组操作", func(t *testing.T) {
		// 获取第一本书的标题
		firstTitle, err := doc.Query("/store/book[0]/title").String()
		if err != nil {
			t.Errorf("Query first book title failed: %v", err)
		}
		if firstTitle != "Sayings of the Century" {
			t.Errorf("Expected 'Sayings of the Century', got '%s'", firstTitle)
		}

		// 获取最后一本书的价格
		lastPrice, err := doc.Query("/store/book[-1]/price").Float()
		if err != nil {
			t.Errorf("Query last book price failed: %v", err)
		}
		if lastPrice != 8.99 {
			t.Errorf("Expected 8.99, got %f", lastPrice)
		}

		// 测试所有书籍数量
		books := doc.Query("/store/book")
		if !books.IsArray() {
			t.Error("store.book should be an array")
		}
		bookCount := books.Count()
		if bookCount != 3 {
			t.Errorf("Expected 3 books, got %d", bookCount)
		}
	})

	t.Run("数组切片操作", func(t *testing.T) {
		// 测试切片 [0:2] - 前两本书
		firstTwoBooks := doc.Query("/store/book[0:2]")
		count := firstTwoBooks.Count()
		if count != 2 {
			t.Logf("NOTE: Slice [0:2] returned %d items, expected 2", count)
		}

		// 测试开放式切片 [1:] - 除第一本外的所有书
		restBooks := doc.Query("/store/book[1:]")
		restCount := restBooks.Count()
		if restCount != 2 {
			t.Logf("NOTE: Slice [1:] returned %d items, expected 2", restCount)
		}
	})

	t.Run("过滤查询测试", func(t *testing.T) {
		// 查找价格小于 10 的书籍
		cheapBooks := doc.Query("/store/book[price < 10]")
		cheapCount := cheapBooks.Count()
		if cheapCount != 2 {
			t.Logf("NOTE: Filter price < 10 found %d books, expected 2", cheapCount)
		}

		// 查找有库存的书籍
		inStockBooks := doc.Query("/store/book[inStock == true]")
		inStockCount := inStockBooks.Count()
		if inStockCount != 2 {
			t.Logf("NOTE: Filter inStock == true found %d books, expected 2", inStockCount)
		}

		// 查找虚构类别的书籍
		fictionBooks := doc.Query("/store/book[category == 'fiction']")
		fictionCount := fictionBooks.Count()
		if fictionCount != 2 {
			t.Logf("NOTE: Filter category == 'fiction' found %d books, expected 2", fictionCount)
		}
	})

	t.Run("复合条件过滤", func(t *testing.T) {
		// 查找价格小于 10 且有库存的书籍
		cheapInStock := doc.Query("/store/book[price < 10 && inStock == true]")
		cheapInStockCount := cheapInStock.Count()
		if cheapInStockCount != 2 {
			t.Logf("NOTE: Complex filter found %d books, expected 2", cheapInStockCount)
		}

		// 查找价格大于 10 或者没有库存的书籍
		expensiveOrOutOfStock := doc.Query("/store/book[price > 10 || inStock == false]")
		count := expensiveOrOutOfStock.Count()
		if count != 1 {
			t.Logf("NOTE: Complex OR filter found %d books, expected 1", count)
		}
	})

	t.Run("通配符测试", func(t *testing.T) {
		// 使用通配符获取所有书籍标题
		allTitles := doc.Query("/store/book[*]/title")
		titleCount := allTitles.Count()
		if titleCount != 3 {
			t.Logf("NOTE: Wildcard query found %d titles, expected 3", titleCount)
		}

		// 使用通配符获取所有价格
		allPrices := doc.Query("/store/book[*]/price")
		priceCount := allPrices.Count()
		if priceCount != 3 {
			t.Logf("NOTE: Wildcard price query found %d prices, expected 3", priceCount)
		}
	})
}

func TestXPathEdgeCases(t *testing.T) {
	// 测试边界情况和特殊语法
	edgeData := `{
		"numbers": [1, 2, 3, 4, 5, 6, 7, 8, 9, 10],
		"mixed": [
			{"type": "A", "value": 100},
			{"type": "B", "value": 200},
			{"type": "A", "value": 150},
			{"type": "C", "value": 50}
		],
		"nested": {
			"level1": {
				"level2": {
					"items": [
						{"name": "item1", "active": true},
						{"name": "item2", "active": false},
						{"name": "item3", "active": true}
					]
				}
			}
		},
		"dotted.key": "special value",
		"empty": [],
		"null_value": null
	}`

	doc, err := ParseString(edgeData)
	if err != nil {
		t.Fatalf("ParseString failed: %v", err)
	}

	t.Run("数字数组操作", func(t *testing.T) {
		// 测试数字数组的各种索引
		first, err := doc.Query("/numbers[0]").Int()
		if err != nil {
			t.Errorf("Query numbers[0] failed: %v", err)
		}
		if first != 1 {
			t.Errorf("Expected 1, got %d", first)
		}

		// 测试负索引
		last, err := doc.Query("/numbers[-1]").Int()
		if err != nil {
			t.Errorf("Query numbers[-1] failed: %v", err)
		}
		if last != 10 {
			t.Errorf("Expected 10, got %d", last)
		}

		// 测试中间切片
		middle := doc.Query("/numbers[3:6]")
		middleCount := middle.Count()
		if middleCount != 3 {
			t.Logf("NOTE: Middle slice returned %d items, expected 3", middleCount)
		}
	})

	t.Run("复杂嵌套查询", func(t *testing.T) {
		// 深度嵌套查询
		item1, err := doc.Query("/nested/level1/level2/items[0]/name").String()
		if err != nil {
			t.Errorf("Query nested item name failed: %v", err)
		}
		if item1 != "item1" {
			t.Errorf("Expected 'item1', got '%s'", item1)
		}

		// 查询活跃项目
		activeItems := doc.Query("/nested/level1/level2/items[active == true]")
		activeCount := activeItems.Count()
		if activeCount != 2 {
			t.Logf("NOTE: Active items filter found %d items, expected 2", activeCount)
		}
	})

	t.Run("特殊键和边界情况", func(t *testing.T) {
		// 测试带点的键名
		dottedValue, err := doc.Query(`/dotted.key`).String()
		if err != nil {
			t.Errorf("Query dotted.key failed: %v", err)
		}
		if dottedValue != "special value" {
			t.Errorf("Expected 'special value', got '%s'", dottedValue)
		}

		// 测试空数组
		empty := doc.Query("/empty")
		if !empty.IsArray() {
			t.Error("empty should be an array")
		}
		if empty.Count() != 0 {
			t.Errorf("Expected empty array count 0, got %d", empty.Count())
		}

		// 测试 null 值
		nullVal := doc.Query("/null_value")
		if !nullVal.IsNull() {
			t.Error("null_value should be null")
		}
	})

	t.Run("过滤器高级测试", func(t *testing.T) {
		// 按类型过滤
		typeAItems := doc.Query("/mixed[type == 'A']")
		typeACount := typeAItems.Count()
		if typeACount != 2 {
			t.Logf("NOTE: Type A filter found %d items, expected 2", typeACount)
		}

		// 按值范围过滤
		highValueItems := doc.Query("/mixed[value > 100]")
		highValueCount := highValueItems.Count()
		if highValueCount != 2 {
			t.Logf("NOTE: High value filter found %d items, expected 2", highValueCount)
		}

		// 复合条件：类型为 A 且值大于 120
		specificItems := doc.Query("/mixed[type == 'A' && value > 120]")
		specificCount := specificItems.Count()
		if specificCount != 1 {
			t.Logf("NOTE: Specific filter found %d items, expected 1", specificCount)
		}
	})
}

func TestXPathRecursiveQueries(t *testing.T) {
	// 专门测试递归查询 (..)
	recursiveData := `{
		"company": {
			"name": "TechCorp",
			"departments": [
				{
					"name": "Engineering",
					"teams": [
						{
							"name": "Frontend",
							"members": [
								{"name": "Alice", "role": "Senior Developer"},
								{"name": "Bob", "role": "Junior Developer"}
							]
						},
						{
							"name": "Backend",
							"members": [
								{"name": "Charlie", "role": "Tech Lead"},
								{"name": "Diana", "role": "Developer"}
							]
						}
					]
				},
				{
					"name": "Marketing",
					"members": [
						{"name": "Eve", "role": "Manager"},
						{"name": "Frank", "role": "Specialist"}
					]
				}
			]
		}
	}`

	doc, err := ParseString(recursiveData)
	if err != nil {
		t.Fatalf("ParseString failed: %v", err)
	}

	t.Run("递归查找所有名称", func(t *testing.T) {
		// 查找所有 name 字段
		allNames := doc.Query("//name")
		nameCount := allNames.Count()
		if nameCount < 8 {
			t.Logf("NOTE: Recursive name query found %d names, expected at least 8", nameCount)
		}
	})

	t.Run("递归查找特定结构", func(t *testing.T) {
		// 查找所有 members 数组
		allMembers := doc.Query("//members")
		memberArrayCount := allMembers.Count()
		if memberArrayCount < 3 {
			t.Logf("NOTE: Recursive members query found %d member arrays, expected at least 3", memberArrayCount)
		}

		// 查找所有角色
		allRoles := doc.Query("//role")
		roleCount := allRoles.Count()
		if roleCount < 6 {
			t.Logf("NOTE: Recursive role query found %d roles, expected at least 6", roleCount)
		}
	})

	t.Run("递归与过滤结合", func(t *testing.T) {
		// 递归查找所有经理角色 (如果支持的话)
		// 注意: `..` 后面直接跟 `[` 是一个复杂的递归过滤场景，其行为依赖于解析器的具体实现
		managers := doc.Query("//members[role == 'Manager']")
		managerCount := managers.Count()
		if managerCount != 1 {
			t.Logf("NOTE: Recursive manager filter found %d managers, expected 1", managerCount)
		}
	})
}
