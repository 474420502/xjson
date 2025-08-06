package xjson

import (
	"testing"
)

func TestXPathPracticalExamples(t *testing.T) {
	// 基于实际使用场景的 JSON 数据，类似电商、API 响应等
	practicalData := `{
		"users": [
			{
				"id": 1,
				"name": "张三",
				"email": "zhangsan@example.com",
				"profile": {
					"age": 28,
					"city": "北京",
					"preferences": {
						"theme": "dark",
						"language": "zh-CN"
					}
				},
				"orders": [
					{
						"id": "order-001",
						"amount": 199.99,
						"status": "completed",
						"items": [
							{"product": "手机", "quantity": 1, "price": 199.99}
						]
					},
					{
						"id": "order-002",
						"amount": 89.50,
						"status": "pending",
						"items": [
							{"product": "充电器", "quantity": 2, "price": 44.75}
						]
					}
				]
			},
			{
				"id": 2,
				"name": "李四",
				"email": "lisi@example.com",
				"profile": {
					"age": 32,
					"city": "上海",
					"preferences": {
						"theme": "light",
						"language": "zh-CN"
					}
				},
				"orders": [
					{
						"id": "order-003",
						"amount": 299.00,
						"status": "completed",
						"items": [
							{"product": "平板", "quantity": 1, "price": 299.00}
						]
					}
				]
			}
		],
		"products": {
			"electronics": [
				{"name": "iPhone", "price": 999.99, "inStock": true, "category": "phone"},
				{"name": "iPad", "price": 599.99, "inStock": false, "category": "tablet"},
				{"name": "MacBook", "price": 1299.99, "inStock": true, "category": "laptop"}
			],
			"books": [
				{"title": "Go 语言实战", "author": "William Kennedy", "price": 59.99, "rating": 4.5},
				{"title": "设计模式", "author": "Gang of Four", "price": 79.99, "rating": 4.8},
				{"title": "算法导论", "author": "Thomas Cormen", "price": 129.99, "rating": 4.9}
			]
		},
		"config": {
			"apiVersion": "v1",
			"features": {
				"search": true,
				"recommendations": false,
				"analytics": true
			},
			"limits": {
				"maxUsers": 1000,
				"maxOrdersPerUser": 50
			}
		}
	}`

	doc, err := ParseString(practicalData)
	if err != nil {
		t.Fatalf("ParseString failed: %v", err)
	}

	t.Run("用户信息查询", func(t *testing.T) {
		// 获取第一个用户的姓名
		firstName, err := doc.Query("/users[0]/name").String()
		if err != nil {
			t.Errorf("Query first user name failed: %v", err)
		}
		if firstName != "张三" {
			t.Errorf("Expected '张三', got '%s'", firstName)
		}

		// 获取第二个用户的城市
		secondUserCity, err := doc.Query("/users[1]/profile/city").String()
		if err != nil {
			t.Errorf("Query second user city failed: %v", err)
		}
		if secondUserCity != "上海" {
			t.Errorf("Expected '上海', got '%s'", secondUserCity)
		}

		// 获取第一个用户的主题偏好
		theme, err := doc.Query("/users[0]/profile/preferences/theme").String()
		if err != nil {
			t.Errorf("Query user theme preference failed: %v", err)
		}
		if theme != "dark" {
			t.Errorf("Expected 'dark', got '%s'", theme)
		}
	})

	t.Run("订单信息查询", func(t *testing.T) {
		// 获取第一个用户的第一个订单金额
		firstOrderAmount, err := doc.Query("/users[0]/orders[0]/amount").Float()
		if err != nil {
			t.Errorf("Query first order amount failed: %v", err)
		}
		if firstOrderAmount != 199.99 {
			t.Errorf("Expected 199.99, got %f", firstOrderAmount)
		}

		// 获取第一个用户的第二个订单状态
		secondOrderStatus, err := doc.Query("/users[0]/orders[1]/status").String()
		if err != nil {
			t.Errorf("Query second order status failed: %v", err)
		}
		if secondOrderStatus != "pending" {
			t.Errorf("Expected 'pending', got '%s'", secondOrderStatus)
		}

		// 获取订单中的商品信息
		productName, err := doc.Query("/users[0]/orders[0]/items[0]/product").String()
		if err != nil {
			t.Errorf("Query order product name failed: %v", err)
		}
		if productName != "手机" {
			t.Errorf("Expected '手机', got '%s'", productName)
		}

		// 获取商品数量
		quantity, err := doc.Query("/users[0]/orders[0]/items[0]/quantity").Int()
		if err != nil {
			t.Errorf("Query product quantity failed: %v", err)
		}
		if quantity != 1 {
			t.Errorf("Expected 1, got %d", quantity)
		}
	})

	t.Run("产品信息查询", func(t *testing.T) {
		// 获取第一个电子产品的名称
		firstElectronics, err := doc.Query("/products/electronics[0]/name").String()
		if err != nil {
			t.Errorf("Query first electronics name failed: %v", err)
		}
		if firstElectronics != "iPhone" {
			t.Errorf("Expected 'iPhone', got '%s'", firstElectronics)
		}

		// 获取第一本书的标题
		firstBookTitle, err := doc.Query("/products/books[0]/title").String()
		if err != nil {
			t.Errorf("Query first book title failed: %v", err)
		}
		if firstBookTitle != "Go 语言实战" {
			t.Errorf("Expected 'Go 语言实战', got '%s'", firstBookTitle)
		}

		// 获取最后一本书的评分
		lastBookRating, err := doc.Query("/products/books[-1]/rating").Float()
		if err != nil {
			t.Errorf("Query last book rating failed: %v", err)
		}
		if lastBookRating != 4.9 {
			t.Errorf("Expected 4.9, got %f", lastBookRating)
		}
	})

	t.Run("配置信息查询", func(t *testing.T) {
		// 获取 API 版本
		apiVersion, err := doc.Query("/config/apiVersion").String()
		if err != nil {
			t.Errorf("Query API version failed: %v", err)
		}
		if apiVersion != "v1" {
			t.Errorf("Expected 'v1', got '%s'", apiVersion)
		}

		// 获取搜索功能状态
		searchEnabled, err := doc.Query("/config/features/search").Bool()
		if err != nil {
			t.Errorf("Query search feature failed: %v", err)
		}
		if !searchEnabled {
			t.Errorf("Expected search to be enabled")
		}

		// 获取最大用户数限制
		maxUsers, err := doc.Query("/config/limits/maxUsers").Int()
		if err != nil {
			t.Errorf("Query max users limit failed: %v", err)
		}
		if maxUsers != 1000 {
			t.Errorf("Expected 1000, got %d", maxUsers)
		}
	})

	t.Run("数组长度和类型检查", func(t *testing.T) {
		// 检查用户数组长度
		users := doc.Query("/users")
		if !users.IsArray() {
			t.Error("users should be an array")
		}
		userCount := users.Count()
		if userCount != 2 {
			t.Errorf("Expected 2 users, got %d", userCount)
		}

		// 检查电子产品数组长度
		electronics := doc.Query("/products/electronics")
		electronicsCount := electronics.Count()
		if electronicsCount != 3 {
			t.Errorf("Expected 3 electronics, got %d", electronicsCount)
		}

		// 检查第一个用户的订单数量
		userOrders := doc.Query("/users[0]/orders")
		orderCount := userOrders.Count()
		if orderCount != 2 {
			t.Errorf("Expected 2 orders for first user, got %d", orderCount)
		}
	})

	t.Run("类型验证和错误处理", func(t *testing.T) {
		// 测试不存在的路径
		nonExistent := doc.Query("/users[0]/nonExistentField")
		if nonExistent.Exists() {
			t.Error("Non-existent field should not exist")
		}

		// 测试数组越界
		outOfBounds := doc.Query("/users[10]/name")
		if outOfBounds.Exists() {
			t.Error("Out of bounds access should not exist")
		}

		// 测试类型不匹配
		config := doc.Query("/config")
		if !config.IsObject() {
			t.Error("config should be an object")
		}

		// 测试 null 值处理（虽然这个例子中没有 null）
		features := doc.Query("/config/features")
		if features.IsNull() {
			t.Error("features should not be null")
		}
	})
}

func TestXPathAdvancedNavigationExamples(t *testing.T) {
	// 更复杂的导航场景
	navigationData := `{
		"document": {
			"metadata": {
				"title": "技术文档",
				"version": "1.0",
				"authors": ["张三", "李四", "王五"]
			},
			"sections": [
				{
					"id": "intro",
					"title": "介绍",
					"content": "这是介绍部分",
					"subsections": [
						{
							"id": "overview",
							"title": "概述",
							"content": "系统概述",
							"examples": [
								{"type": "code", "content": "console.log('Hello');"},
								{"type": "text", "content": "这是一个示例"}
							]
						}
					]
				},
				{
					"id": "guide",
					"title": "使用指南",
					"content": "详细的使用指南",
					"subsections": [
						{
							"id": "installation",
							"title": "安装",
							"content": "安装步骤",
							"examples": [
								{"type": "command", "content": "npm install package"},
								{"type": "code", "content": "import package from 'package';"}
							]
						},
						{
							"id": "configuration",
							"title": "配置",
							"content": "配置说明",
							"examples": [
								{"type": "json", "content": "{\"key\": \"value\"}"}
							]
						}
					]
				}
			]
		}
	}`

	doc, err := ParseString(navigationData)
	if err != nil {
		t.Fatalf("ParseString failed: %v", err)
	}

	t.Run("多层嵌套导航", func(t *testing.T) {
		// 导航到深层嵌套的内容
		exampleContent, err := doc.Query("/document/sections[0]/subsections[0]/examples[0]/content").String()
		if err != nil {
			t.Errorf("Query deep nested content failed: %v", err)
		}
		if exampleContent != "console.log('Hello');" {
			t.Errorf("Expected 'console.log('Hello');', got '%s'", exampleContent)
		}

		// 获取第二个章节的第一个子章节标题
		subTitle, err := doc.Query("/document/sections[1]/subsections[0]/title").String()
		if err != nil {
			t.Errorf("Query subsection title failed: %v", err)
		}
		if subTitle != "安装" {
			t.Errorf("Expected '安装', got '%s'", subTitle)
		}
	})

	t.Run("数组中的对象导航", func(t *testing.T) {
		// 获取作者数组的第一个元素
		firstAuthor, err := doc.Query("/document/metadata/authors[0]").String()
		if err != nil {
			t.Errorf("Query first author failed: %v", err)
		}
		if firstAuthor != "张三" {
			t.Errorf("Expected '张三', got '%s'", firstAuthor)
		}

		// 获取最后一个作者
		lastAuthor, err := doc.Query("/document/metadata/authors[-1]").String()
		if err != nil {
			t.Errorf("Query last author failed: %v", err)
		}
		if lastAuthor != "王五" {
			t.Errorf("Expected '王五', got '%s'", lastAuthor)
		}

		// 获取作者数组的长度
		authors := doc.Query("/document/metadata/authors")
		authorCount := authors.Count()
		if authorCount != 3 {
			t.Errorf("Expected 3 authors, got %d", authorCount)
		}
	})

	t.Run("条件查找和类型检查", func(t *testing.T) {
		// 检查章节是否存在
		intro := doc.Query("/document/sections[0]")
		if !intro.Exists() {
			t.Error("Introduction section should exist")
		}

		// 检查子章节是否是对象
		subsection := doc.Query("/document/sections[0]/subsections[0]")
		if !subsection.IsObject() {
			t.Error("Subsection should be an object")
		}

		// 检查示例数组
		examples := doc.Query("/document/sections[0]/subsections[0]/examples")
		if !examples.IsArray() {
			t.Error("Examples should be an array")
		}
		exampleCount := examples.Count()
		if exampleCount != 2 {
			t.Errorf("Expected 2 examples, got %d", exampleCount)
		}
	})
}

func TestXPathRealWorldScenarios(t *testing.T) {
	// 模拟真实世界的 API 响应
	apiResponse := `{
		"status": "success",
		"timestamp": "2023-12-01T10:30:00Z",
		"data": {
			"pagination": {
				"page": 1,
				"limit": 10,
				"total": 25,
				"hasNext": true
			},
			"results": [
				{
					"id": "prod-001",
					"name": "智能手机",
					"description": "最新款智能手机",
					"price": {
						"amount": 2999.00,
						"currency": "CNY",
						"discount": 0.1
					},
					"availability": {
						"inStock": true,
						"quantity": 50,
						"location": "北京仓库"
					},
					"reviews": [
						{
							"id": "rev-001",
							"rating": 5,
							"comment": "非常好用",
							"author": "用户A",
							"verified": true
						},
						{
							"id": "rev-002",
							"rating": 4,
							"comment": "性价比不错",
							"author": "用户B",
							"verified": false
						}
					],
					"tags": ["电子产品", "手机", "智能设备"]
				},
				{
					"id": "prod-002",
					"name": "无线耳机",
					"description": "高音质无线耳机",
					"price": {
						"amount": 299.00,
						"currency": "CNY",
						"discount": 0.0
					},
					"availability": {
						"inStock": false,
						"quantity": 0,
						"location": "上海仓库"
					},
					"reviews": [],
					"tags": ["电子产品", "音频", "无线"]
				}
			]
		},
		"meta": {
			"requestId": "req-12345",
			"version": "1.2.3",
			"server": "api-server-01"
		}
	}`

	doc, err := ParseString(apiResponse)
	if err != nil {
		t.Fatalf("ParseString failed: %v", err)
	}

	t.Run("API 响应基本信息", func(t *testing.T) {
		// 检查响应状态
		status, err := doc.Query("/status").String()
		if err != nil {
			t.Errorf("Query status failed: %v", err)
		}
		if status != "success" {
			t.Errorf("Expected 'success', got '%s'", status)
		}

		// 获取分页信息
		currentPage, err := doc.Query("/data/pagination/page").Int()
		if err != nil {
			t.Errorf("Query current page failed: %v", err)
		}
		if currentPage != 1 {
			t.Errorf("Expected page 1, got %d", currentPage)
		}

		totalItems, err := doc.Query("/data/pagination/total").Int()
		if err != nil {
			t.Errorf("Query total items failed: %v", err)
		}
		if totalItems != 25 {
			t.Errorf("Expected 25 total items, got %d", totalItems)
		}

		hasNext, err := doc.Query("/data/pagination/hasNext").Bool()
		if err != nil {
			t.Errorf("Query hasNext failed: %v", err)
		}
		if !hasNext {
			t.Error("Expected hasNext to be true")
		}
	})

	t.Run("产品信息详细查询", func(t *testing.T) {
		// 第一个产品的基本信息
		productName, err := doc.Query("/data/results[0]/name").String()
		if err != nil {
			t.Errorf("Query product name failed: %v", err)
		}
		if productName != "智能手机" {
			t.Errorf("Expected '智能手机', got '%s'", productName)
		}

		// 价格信息
		price, err := doc.Query("/data/results[0]/price/amount").Float()
		if err != nil {
			t.Errorf("Query price failed: %v", err)
		}
		if price != 2999.00 {
			t.Errorf("Expected 2999.00, got %f", price)
		}

		currency, err := doc.Query("/data/results[0]/price/currency").String()
		if err != nil {
			t.Errorf("Query currency failed: %v", err)
		}
		if currency != "CNY" {
			t.Errorf("Expected 'CNY', got '%s'", currency)
		}

		// 库存信息
		inStock, err := doc.Query("/data/results[0]/availability/inStock").Bool()
		if err != nil {
			t.Errorf("Query inStock failed: %v", err)
		}
		if !inStock {
			t.Error("Expected product to be in stock")
		}

		quantity, err := doc.Query("/data/results[0]/availability/quantity").Int()
		if err != nil {
			t.Errorf("Query quantity failed: %v", err)
		}
		if quantity != 50 {
			t.Errorf("Expected quantity 50, got %d", quantity)
		}
	})

	t.Run("评论和标签信息", func(t *testing.T) {
		// 第一个产品的评论
		reviews := doc.Query("/data/results[0]/reviews")
		if !reviews.IsArray() {
			t.Error("Reviews should be an array")
		}
		reviewCount := reviews.Count()
		if reviewCount != 2 {
			t.Errorf("Expected 2 reviews, got %d", reviewCount)
		}

		// 第一条评论的评分
		firstRating, err := doc.Query("/data/results[0]/reviews[0]/rating").Int()
		if err != nil {
			t.Errorf("Query first review rating failed: %v", err)
		}
		if firstRating != 5 {
			t.Errorf("Expected rating 5, got %d", firstRating)
		}

		// 验证状态
		verified, err := doc.Query("/data/results[0]/reviews[0]/verified").Bool()
		if err != nil {
			t.Errorf("Query verified status failed: %v", err)
		}
		if !verified {
			t.Error("Expected first review to be verified")
		}

		// 标签数组
		tags := doc.Query("/data/results[0]/tags")
		if !tags.IsArray() {
			t.Error("Tags should be an array")
		}
		tagCount := tags.Count()
		if tagCount != 3 {
			t.Errorf("Expected 3 tags, got %d", tagCount)
		}

		// 第一个标签
		firstTag, err := doc.Query("/data/results[0]/tags[0]").String()
		if err != nil {
			t.Errorf("Query first tag failed: %v", err)
		}
		if firstTag != "电子产品" {
			t.Errorf("Expected '电子产品', got '%s'", firstTag)
		}
	})

	t.Run("第二个产品的对比", func(t *testing.T) {
		// 第二个产品的库存状态
		secondProductStock, err := doc.Query("/data/results[1]/availability/inStock").Bool()
		if err != nil {
			t.Errorf("Query second product stock failed: %v", err)
		}
		if secondProductStock {
			t.Error("Expected second product to be out of stock")
		}

		// 第二个产品的评论数量（应该为0）
		secondReviews := doc.Query("/data/results[1]/reviews")
		secondReviewCount := secondReviews.Count()
		if secondReviewCount != 0 {
			t.Errorf("Expected 0 reviews for second product, got %d", secondReviewCount)
		}

		// 第二个产品的折扣
		discount, err := doc.Query("/data/results[1]/price/discount").Float()
		if err != nil {
			t.Errorf("Query discount failed: %v", err)
		}
		if discount != 0.0 {
			t.Errorf("Expected discount 0.0, got %f", discount)
		}
	})

	t.Run("元数据信息", func(t *testing.T) {
		// 请求 ID
		requestId, err := doc.Query("/meta/requestId").String()
		if err != nil {
			t.Errorf("Query requestId failed: %v", err)
		}
		if requestId != "req-12345" {
			t.Errorf("Expected 'req-12345', got '%s'", requestId)
		}

		// API 版本
		version, err := doc.Query("/meta/version").String()
		if err != nil {
			t.Errorf("Query API version failed: %v", err)
		}
		if version != "1.2.3" {
			t.Errorf("Expected '1.2.3', got '%s'", version)
		}
	})
}
