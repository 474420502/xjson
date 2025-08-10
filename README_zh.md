# XJSON - 统一节点模型JSON处理器 (v0.1.0)

**XJSON** **是一个强大的 Go JSON 处理库，采用完全统一的** **Node** **模型，支持路径函数、流式操作和灵活的查询语法。**

## ✨ 核心特性

* **🎯** **单一节点类型**：所有操作都基于 **xjson.Node**，无 **Result** **类型。**
* **🧩** **路径函数**：通过 **/path[@func]/subpath** **语法将自定义逻辑注入查询。**
* **🔗** **链式操作**：支持流畅的函数注册、查询和数据操作。
* **🌀** **健壮的错误处理**：通过 **node.Error()** **在链式调用末尾统一检查错误。**
* **⚡️** **性能导向**：通过高效的链式操作和原生值访问实现零拷贝级别的性能。
* **🌟** **通配符查询**：支持 **`*`** 通配符和复杂的路径表达式。
* **🔍** **递归下降**：通过 **//key** **语法在整个JSON树中深度搜索匹配的键。**
* **⬆️** **上级路径**：通过 **../** **语法访问父级节点，实现灵活的相对路径导航。**

## 🚀 快速开始

XJSON 提供了基础和高级两种使用模式。以下是两种级别的示例：

### 基础用法

XJSON 的主要目的是让 JSON 路径查询变得简单直观。以下是各种路径语法的使用示例：

```go
package main

import (
	"fmt"
	"github.com/474420502/xjson"
)

func main() {
	// 复杂的 JSON 数据用于演示路径查询
	data := `{
		"store": {
			"books": [
				{
					"title": "Moby Dick",
					"price": 8.99,
					"author": {
						"first_name": "Herman",
						"last_name": "Melville"
					},
					"tags": ["classic", "adventure"],
					"isbn-10": "0123456789"
				},
				{
					"title": "Clean Code",
					"price": 29.99,
					"author": {
						"first_name": "Robert",
						"last_name": "Martin"
					},
					"tags": ["programming", "coding"]
				}
			],
			"electronics": {
				"computers": [
					{
						"name": "Laptop",
						"price": 999.99,
						"specifications": {
							"cpu": "Intel i7",
							"memory": "16GB"
						},
						"in_stock": true
					}
				]
			},
			"special.keys": {
				"user.profile": {
					"name": "John Doe",
					"settings": {
						"theme": "dark",
						"notifications": true
					}
				}
			}
		},
		"ratings": [
			{"book": "Moby Dick", "score": 4.5},
			{"book": "Clean Code", "score": 4.8}
		]
	}`

	// 解析 JSON
	root, err := xjson.Parse(data)
	if err != nil {
		panic(err)
	}

	// 1. 基础键访问
	store := root.Query("/store")
	fmt.Println("商店存在:", store.IsValid())

	// 2. 嵌套键访问
	bookTitle := root.Query("/store/books[0]/title").String()
	fmt.Println("第一本书标题:", bookTitle)

	// 3. 数组索引
	firstAuthor := root.Query("/store/books[0]/author/first_name").String()
	fmt.Println("第一位作者的名字:", firstAuthor)

	// 4. 数组切片
	bookTitles := root.Query("/store/books[:]/title").Strings()
	fmt.Println("所有书名:", bookTitles)

	// 5. 使用引号语法访问包含特殊字符的键
	userName := root.Query("/store/['special.keys']/['user.profile']/name").String()
	fmt.Println("特殊键的用户名:", userName)

	// 6. 访问包含点号的键
	userTheme := root.Query(`/store/['special.keys']/['user.profile']/settings/theme`).String()
	fmt.Println("用户主题:", userTheme)

	// 7. 通配符使用
	allFirstNames := root.Query("/store/books/*/author/first_name").Strings()
	fmt.Println("所有作者的名字:", allFirstNames)

	// 8. 按条件访问数组元素（第一个元素）
	firstRating := root.Query("/ratings[0]/score").Float()
	fmt.Printf("第一个评分: %.1f\n", firstRating)

	fmt.Println("\n--- 更多路径示例 ---")

	// 9. 复杂嵌套访问
	cpuSpec := root.Query("/store/electronics/computers[0]/specifications/cpu").String()
	fmt.Println("CPU规格:", cpuSpec)

	// 10. 访问布尔值
	inStock := root.Query("/store/electronics/computers[0]/in_stock").Bool()
	fmt.Println("电脑有库存:", inStock)

	// 11. 访问数组元素
	firstTag := root.Query("/store/books[0]/tags[0]").String()
	fmt.Println("第一本书的第一个标签:", firstTag)

	// 12. 访问数值
	bookPrice := root.Query("/store/books[1]/price").Float()
	fmt.Printf("第二本书价格: $%.2f\n", bookPrice)
}
```

不同类型的路径操作：

```go
func pathExamples() {
	data := `{
		"users": [
			{
				"id": 1,
				"name": "Alice",
				"profile": {
					"age": 25,
					"active": true,
					"tags": ["developer", "go", "json"]
				},
				"scores": [95, 87, 92]
			},
			{
				"id": 2,
				"name": "Bob",
				"profile": {
					"age": 30,
					"active": false,
					"tags": ["manager", "planning"]
				},
				"scores": [88, 91, 79]
			}
		],
		"metadata": {
			"version": "1.0",
			"created": "2023-01-01"
		}
	}`

	root, _ := xjson.Parse(data)

	// 数组索引访问
	firstUserId := root.Query("/users[0]/id").Int()
	fmt.Println("第一个用户ID:", firstUserId)

	// 数组切片访问
	userNames := root.Query("/users[:]/name").Strings()
	fmt.Println("用户名:", userNames)

	// 嵌套对象访问
	firstUserAge := root.Query("/users[0]/profile/age").Int()
	fmt.Println("第一个用户年龄:", firstUserAge)

	// 对象数组属性访问
	allTags := root.Query("/users[*]/profile/tags").Strings()
	fmt.Println("所有用户标签:", allTags)

	// 嵌套数组访问
	firstUserFirstScore := root.Query("/users[0]/scores[0]").Int()
	fmt.Println("第一个用户的第一分数:", firstUserFirstScore)

	// 布尔值访问
	firstUserActive := root.Query("/users[0]/profile/active").Bool()
	fmt.Println("第一个用户是否活跃:", firstUserActive)

	// 访问元数据
	version := root.Query("/metadata/version").String()
	fmt.Println("版本:", version)
}
```

处理特殊键名：

``go
func specialKeysExample() {
	data := `{
		"user-data": {
			"user.profile": {
				"first.name": "John",
				"last.name": "Doe"
			},
			"user.settings": {
				"ui.theme": "dark",
				"email.notifications": true
			}
		},
		"api/v1/users": [
			{
				"id": 1,
				"profile.data": {
					"name": "Alice",
					"contact-info": {
						"email.address": "alice@example.com"
					}
				}
			}
		]
	}`

	root, _ := xjson.Parse(data)

	// 访问包含点号的键
	firstName := root.Query(`/['user-data']/['user.profile']/['first.name']`).String()
	fmt.Println("名字:", firstName)

	// 访问包含斜杠的键
	apiPath := root.Query(`/['api/v1/users']`).Len()
	fmt.Println("API用户数量:", apiPath)

	// 混合常规键和特殊键
	userName := root.Query(`/['api/v1/users'][0]/['profile.data']/name`).String()
	fmt.Println("用户名:", userName)

	// 使用特殊键的深层访问
	email := root.Query(`/['api/v1/users'][0]/['profile.data']/['contact-info']/['email.address']`).String()
	fmt.Println("邮箱:", email)

	// 访问嵌套的特殊键
	theme := root.Query(`/['user-data']/['user.settings']/['ui.theme']`).String()
	fmt.Println("主题:", theme)
}
```

数组操作：

``go
func arrayExample() {
	data := `{
		"users": [
			{"name": "Alice", "age": 25},
			{"name": "Bob", "age": 30},
			{"name": "Charlie", "age": 35}
		]
	}`

	root, _ := xjson.Parse(data)

	// 获取数组长度
	count := root.Get("users").Len()
	fmt.Printf("用户总数: %d\n", count)

	// 通过索引访问
	firstUser := root.Get("users").Index(0).Get("name").String()
	fmt.Printf("第一个用户: %s\n", firstUser)

	// 遍历数组
	root.Get("users").ForEach(func(index interface{}, user xjson.Node) {
		name := user.Get("name").String()
		age := user.Get("age").Int()
		fmt.Printf("用户 %d: %s (年龄 %d)\n", index, name, age)
	})
}
```

### 高级用法

使用函数进行复杂数据处理：

``go
func advancedExample() {
	data := `{
		"store": {
			"books": [
				{"title": "Moby Dick", "price": 8.99, "tags": ["classic", "adventure"]},
				{"title": "Clean Code", "price": 29.99, "tags": ["programming"]},
				{"title": "Go in Action", "price": 19.99, "tags": ["programming", "golang"]}
			],
			"electronics": [
				{"name": "Laptop", "price": 999.99, "in_stock": true},
				{"name": "Mouse", "price": 29.99, "in_stock": false}
			]
		}
	}`

	root, err := xjson.Parse(data)
	if err != nil {
		panic(err)
	}

	// 注册自定义函数
	root.RegisterFunc("cheap", func(n xjson.Node) xjson.Node {
		return n.Filter(func(child xjson.Node) bool {
			price, _ := child.Get("price").RawFloat()
			return price < 20
		})
	}).RegisterFunc("inStock", func(n xjson.Node) xjson.Node {
		return n.Filter(func(child xjson.Node) bool {
			return child.Get("in_stock").Bool()
		})
	}).RegisterFunc("programming", func(n xjson.Node) xjson.Node {
		return n.Filter(func(child xjson.Node) bool {
			return child.Get("tags").Contains("programming")
		})
	})

	// 使用路径函数进行复杂查询
	cheapBooks := root.Query("/store/books[@cheap]/title").Strings()
	fmt.Println("便宜的书籍:", cheapBooks)

	// 查找所有有库存的电子产品
	inStockItems := root.Query("/store/electronics[@inStock]/name").Strings()
	fmt.Println("有库存的商品:", inStockItems)

	// 查找编程类书籍
	progBooks := root.Query("/store/books[@programming]/title").Strings()
	fmt.Println("编程类书籍:", progBooks)

	// 使用递归下降查找所有价格
	allPrices := root.Query("//price").Map(func(n xjson.Node) interface{} {
		price, _ := n.RawFloat()
		return price
	})

	// 计算平均价格
	var sum float64
	var count int
	allPrices.ForEach(func(_ interface{}, priceNode xjson.Node) {
		if price, ok := priceNode.Interface().(float64); ok {
			sum += price
			count++
		}
	})
	avgPrice := sum / float64(count)
	fmt.Printf("平均价格: %.2f\n", avgPrice)

	// 使用上级路径导航
	firstBookTitle := root.Query("/store/books[0]/../books[0]/title").String()
	fmt.Println("第一本书 (使用上级路径导航):", firstBookTitle)
}
```

数据修改：

``go
func modificationExample() {
	data := `{
		"users": [
			{"id": 1, "name": "John", "active": true},
			{"id": 2, "name": "Jane", "active": false}
		]
	}`

	root, _ := xjson.Parse(data)

	// 修改现有数据
	root.Query("/users[0]").Set("name", "John Doe")
	
	// 添加新数据
	newUser := map[string]interface{}{
		"id": 3,
		"name": "Bob",
		"active": true,
	}
	root.Query("/users").Append(newUser)

	// 使用 SetValue 替换整个节点值
	root.Query("/users[1]/active").SetValue(true)

	fmt.Println("修改后的数据:", root.String())
}
```

## 💡 核心设计

### 1. 统一节点模型

**所有 JSON 元素（对象、数组、字符串、数字等），包括查询结果集，都由** **Node** **接口表示。**

```go
type Node interface {
    // 基础访问
    Type() NodeType
    IsValid() bool
    Error() error
    Path() string
    Raw() string
    Parent() Node
  
    // 查询方法
    Query(path string) Node
    Get(key string) Node
    Index(i int) Node
  
    // 流式操作
    Filter(fn PredicateFunc) Node
    Map(fn TransformFunc) Node
    ForEach(fn func(keyOrIndex interface{}, value Node)) 
    Len() int
  
    // 写操作
    Set(key string, value interface{}) Node
    Append(value interface{}) Node
    SetValue(value interface{}) Node
  
    // 函数支持
    RegisterFunc(name string, fn UnaryPathFunc) Node
    CallFunc(name string) Node
    RemoveFunc(name string) Node
    Apply(fn PathFunc) Node
    GetFuncs() *map[string]UnaryPathFunc
  
    // 类型转换
    String() string
    MustString() string
    Float() float64
    MustFloat() float64
    Int() int64
    MustInt() int64
    Bool() bool
    MustBool() bool
    Time() time.Time
    MustTime() time.Time
    Array() []Node
    MustArray() []Node
    Interface() interface{}
  
    // 原生值访问 (性能优化)
    RawFloat() (float64, bool)
    RawString() (string, bool)
  
    // 其他转换方法
    Strings() []string
    Keys() []string
    Contains(value string) bool
    AsMap() map[string]Node
    MustAsMap() map[string]Node
}
```

### 2. 函数类型系统

**XJSON 提供了多种函数类型以支持不同的操作场景：**

```go
// 路径函数 - 通用函数容器
type PathFunc interface{}

// 一元路径函数 - 节点到节点的转换
type UnaryPathFunc func(node Node) Node

// 谓词函数 - 用于过滤操作
type PredicateFunc func(node Node) bool

// 转换函数 - 用于映射操作
type TransformFunc func(node Node) interface{}
```

### 3. 错误处理

**XJSON 采用链式调用友好的错误处理模式：**

```go
// 无需在每一步都检查 err
value := root.Query("/path/that/does/not/exist").Get("key").Int()

// 在最后统一检查
if err := root.Error(); err != nil {
    fmt.Println("操作链失败:", err)
}
```

### 4. 路径查询语法

XJSON 提供了强大而灵活的路径查询语法，支持从简单到复杂的各种数据访问模式。

#### **基础语法**

**4.1. 根节点**

路径查询总是以 `/` 开头，表示从根节点开始。

* **语法**: `/`
* **描述**: 代表 JSON 数据的根节点。
* **示例**: `/store` 从根节点获取 `store` 键的值。

**注意**: `/store/books` 和 `store/books` 这两种写法是等效的。

**4.2. 键访问**

标准的对象字段访问通过键名直接完成。任何符合 Go 语言标识符习惯的字符串都可以直接作为路径段。

* **语法**: `/key1/key2`
* **示例**: `/store/books`，这段路径会依次获取 `store` 键和 `books` 键。

**4.3. 数组访问**

通过方括号 `[...]` 访问数组元素，支持单个索引和范围切片。

* **索引访问**:

  * **语法**: `[<index>]`
  * **描述**: 获取单个数组元素，索引从 0 开始。
  * **示例**: `/store/books[0]`，获取 `books` 数组的第一个元素。
* **切片访问**:

  * **语法**:
    * `[start:end]`: 获取从 `start` 到 `end-1` 的元素。
    * `[start:]`: 获取从 `start` 到末尾的元素。
    * `[:end]`: 获取从开头到 `end-1` 的元素。
    * `[-N:]`: 获取最后 N 个元素。
  * **描述**: 获取数组的一个子集，并返回一个包含这些元素的新数组节点。
  * **示例**: `/store/books[1:3]`，返回一个包含 `books` 数组中第二个和第三个元素的新数组。

**4.4. 函数调用**

在路径中通过 `[@<funcName>]` 语法调用已注册的函数。函数提供了一种强大的数据处理和过滤机制。

* **语法**: `[@<函数名>]`
* **标志符**: `@` 符号明确表示这是一个函数调用。
* **要求**: 函数必须已通过 `RegisterFunc` 注册到节点上。
* **示例**: `/store/books[@cheap]/title`，在 `books` 数组上调用 `cheap` 函数，并从结果中提取 `title`。

**4.5. 通配符**

星号 `*` 作为通配符，用于匹配一个节点下的所有直接子元素。

* **语法**: `*`
* **对象上的行为**: 匹配对象的所有值，并返回一个包含这些值的新数组节点。
* **数组上的行为**: 匹配数组的所有元素，并返回该数组自身。
* **示例**: `/store/*/title`，获取 `store` 对象下所有直接子节点（在这里是 `books` 数组）的 `title` 字段。

#### **高级语法**

**5.1. 链式与混合语法**

所有核心组件都可以自由组合，形成强大的链式查询。解析器会从左到右依次执行每个操作。

* **示例**: `/store/books[@filter][0]/name`
  1. `/store/books`: 获取 `books` 数组。
  2. `[@filter]`: 在该数组上调用 `filter` 函数。
  3. `[0]`: 获取函数返回结果（应为一个数组）的第一个元素。
  4. `/name`: 获取该元素的 `name` 字段。

**5.2. 特殊字符键名处理**

当对象键名包含 `/`, `.`, `[`, `]` 等特殊字符或非字母数字时，必须使用方括号和引号 `['<key>']` 或 `["<key>"]` 的形式来界定。

* **语法**: `['<键名>']` 或 `["<键名>"]`
* **键名包含斜杠**: `/['/api/v1/users']`
* **键名包含点号**: `/data/['user.profile']/name`
* **键名包含引号**:
  * 如果键名为 `a"key`，使用 `['a"key']`。
  * 如果键名为 `a'key`，使用 `["a'key"]`。
* **与普通路径混合**: `/data['user-settings']/theme`

**5.3. 递归下降**

双斜杠 `//` 用于在当前节点及其所有后代中进行深度搜索，查找匹配的键。

* **语法**: `//key`
* **描述**: 与 `/` 只在直接子节点中查找不同，`//` 会遍历整个子树，将所有匹配 `key` 的节点收集到一个新的数组节点中。
* **示例**: `//author` 将从根节点开始，查找所有层级下的 `author` 字段。

**更多使用示例：**

```go
// 查找所有价格字段
allPrices := root.Query("//price").Strings()

// 查找所有包含标签的书籍
taggedBooks := root.Query("//books").Filter(func(n xjson.Node) bool {
    return n.Get("tags").Len() > 0
})

// 查找所有库存为 true 的商品
inStockItems := root.Query("//in_stock").Filter(func(n xjson.Node) bool {
    return n.Bool() == true
})

// 结合函数使用，查找所有低价商品
cheapItems := root.Query("//price[@cheap]")
```

**最佳实践：**

1. **限制搜索范围**：先使用精确路径定位到大致区域，再使用递归下降

   ```go
   // 推荐：先定位到 store，再搜索
   storePrices := root.Query("/store//price")

   // 避免全局搜索
   allPrices := root.Query("//price")
   ```
2. **结合过滤函数**：使用 `Filter()` 方法进一步筛选结果

   ```go
   // 找到所有价格并筛选出低价的
   cheapPrices := root.Query("//price").Filter(func(n xjson.Node) bool {
       price, _ := n.RawFloat()
       return price < 20
   })
   ```
3. **谨慎使用**：在已知结构的情况下优先使用精确路径

> **性能警告**：递归下降 `//` 是一个非常强大但开销极大的操作。因为它需要遍历一个节点下的整个子树，当处理大型或深层嵌套的 JSON 数据时，可能会成为性能瓶颈。建议仅在数据结构不确定或确实需要全局搜索时使用，在性能敏感的场景下应优先使用精确路径。

**5.4. 上级路径查找**

双点 `../` 语法用于访问当前节点的父级节点，实现相对路径导航。

* **语法**: `../key` 或 `../`
* **描述**: 允许从当前节点向上导航到父级节点，然后继续向下查询。这在处理复杂嵌套结构时特别有用，可以在不知道完整路径的情况下进行灵活的数据访问。
* **示例**: `/store/books[0]/../electronics` 从第一本书向上导航到 `store` 节点，然后访问 `electronics`。

**使用示例：**

```go
// 从书籍节点导航到父级 store，然后获取 electronics
electronicsFromBook := root.Query("/store/books[0]/../electronics/laptops").Strings()

// 获取所有书籍的父级分类名称
bookCategories := root.Query("/store/books[0]/../").Keys()

// 在数组元素中引用兄弟字段
firstBookTitle := root.Query("/store/books[0]/title").String()
firstBookPrice := root.Query("/store/books[0]/../books[0]/price").Float()

// 多级上级导航
rootFromDeep := root.Query("/store/electronics/laptops[0]/../../authors").Strings()
```

**实际应用场景：**

1. **关联数据查询**：在嵌套结构中查找相关数据

   ```go
   // 找到所有有库存商品的分类
   inStockCategories := root.Query("/store/*/laptops").Filter(func(n xjson.Node) bool {
       return n.Get("in_stock").Bool() == true
   }).Query("../..").Keys()
   ```
2. **数据验证**：检查字段间的关系

   ```go
   // 验证价格是否在合理范围内
   validatePrice := root.Query("/store/books").Filter(func(n xjson.Node) bool {
       price := n.Get("price").Float()
       category := n.Query("../").String() // 获取父级信息
       // 根据分类验证价格
       return isValidPriceForCategory(price, category)
   })
   ```
3. **动态路径构建**：在不确定具体结构时进行导航

   ```go
   // 从任意节点向上查找特定字段
   findStoreInfo := root.Query("//price/../..") // 从价格找到对应的 store
   ```

**限制和注意事项：**

1. **根节点限制**：在根节点使用 `../` 会返回无效节点
2. **性能考虑**：过多的上级导航可能影响代码可读性，建议在已知结构时使用精确路径
3. **链式使用**：可以连续使用多个 `../` 进行多级向上导航

#### **语法速查表**

| 分类               | 语法            | 描述                                            | 示例                         |
| :----------------- | :-------------- | :---------------------------------------------- | :--------------------------- |
| **基础**     | `/`           | 路径段之间的分隔符。                            | `/store/books`             |
|                    | `key`         | 访问对象的字段。                                | `/store`                   |
| **数组**     | `[<index>]`   | 按索引访问数组元素。                            | `[0]`, `[-1]`            |
|                    | `[start:end]` | 按范围访问数组元素（切片）。                    | `[1:3]`, `[:-1]`         |
| **函数**     | `[@<name>]`   | 调用已注册的路径函数。                          | `[@cheap]`, `[@inStock]` |
| **高级**     | `*`           | 匹配对象或数组的所有直接子元素。                | `/store/*`                 |
|                    | `//key`       | 递归搜索所有后代节点中的 `key` (性能开销大)。 | `//author`                 |
|                    | `../key`      | 访问父级节点，然后继续向下查询。                | `/books[0]/../electronics` |
| **特殊字符** | `['<key>']`   | 界定包含特殊字符的键名。                        | `['user.profile']`         |
|                    | `["<key>"]`   | 界定包含单引号的键名。                          | `["a'key"]`                |

### 6. 函数注册和调用

**新版本的函数系统更加强大和灵活：**

```go
// 注册函数（推荐方式）
root.RegisterFunc("filterFunc", func(n xjson.Node) xjson.Node {
    return n.Filter(func(child xjson.Node) bool {
        return child.Get("price").Float() > 10
    })
})

// 路径查询中使用函数
result := root.Query("/items[@filterFunc]/name")

// 直接调用函数
result := root.CallFunc("filterFunc")

// 使用 Apply 立即应用函数
result := root.Apply(func(n xjson.Node) bool {
    return n.Get("active").Bool()
})

// 移除函数
root.RemoveFunc("filterFunc")

// 获取已注册函数
funcs := root.GetFuncs()
```

## 🛠️ 完整 API 参考

### 函数管理

| 方法                             | 描述                       | 示例                                        |
| -------------------------------- | -------------------------- | ------------------------------------------- |
| **RegisterFunc(name, fn)** | 注册路径函数               | `root.RegisterFunc("cheap", filterCheap)` |
| **CallFunc(name)**         | 直接调用函数               | `root.CallFunc("cheap")`                  |
| **RemoveFunc(name)**       | 移除函数                   | `root.RemoveFunc("cheap")`                |
| **Apply(fn)**              | 立即应用函数               | `root.Apply(predicateFunc)`               |
| **GetFuncs()**             | 获取已注册函数             | `funcs := root.GetFuncs()`                |
| **Error() error**          | 返回链式调用中的第一个错误 | `if err := n.Error(); err != nil { ... }` |

### 流式操作

| 方法                  | 描述         | 示例                                                                   |
| --------------------- | ------------ | ---------------------------------------------------------------------- |
| **Filter(fn)**  | 过滤节点集合 | `n.Filter(func(n Node) bool { return n.Get("active").Bool() })`      |
| **Map(fn)**     | 转换节点集合 | `n.Map(func(n Node) interface{} { return n.Get("name").String() })`  |
| **ForEach(fn)** | 遍历节点集合 | `n.ForEach(func(i interface{}, v Node) { fmt.Println(v.String()) })` |

### 原生值访问

| 方法                      | 描述                | 示例                                         |
| ------------------------- | ------------------- | -------------------------------------------- |
| **RawFloat()**      | 直接获取 float64 值 | `if price, ok := n.RawFloat(); ok { ... }` |
| **RawString()**     | 直接获取 string 值  | `if name, ok := n.RawString(); ok { ... }` |
| **Strings()**       | 获取字符串数组      | `tags := n.Strings()`                      |
| **Contains(value)** | 检查是否包含字符串  | `if n.Contains("target") { ... }`          |
| **AsMap()**         | 获取节点为 map      | `obj := n.AsMap()`                         |
| **Keys()**          | 获取对象的所有键    | `keys := n.Keys()`                         |

### 强制类型转换

| 方法                   | 描述                            | 示例                        |
| ---------------------- | ------------------------------- | --------------------------- |
| **MustString()** | 获取字符串值，失败时 panic      | `value := n.MustString()` |
| **MustFloat()**  | 获取 float64 值，失败时 panic   | `value := n.MustFloat()`  |
| **MustInt()**    | 获取 int64 值，失败时 panic     | `value := n.MustInt()`    |
| **MustBool()**   | 获取 bool 值，失败时 panic      | `value := n.MustBool()`   |
| **MustTime()**   | 获取 time.Time 值，失败时 panic | `value := n.MustTime()`   |
| **MustArray()**  | 获取数组值，失败时 panic        | `value := n.MustArray()`  |
| **MustAsMap()**  | 获取 map 值，失败时 panic       | `value := n.MustAsMap()`  |

## ⚡ 性能优化

* **函数缓存**：编译后的路径会被缓存，以加速重复查询。
* **原生值访问**：`Raw` 系列方法直接从底层内存访问数据，避免创建中间 **Node** 对象。
* **短路优化**：在某些过滤和查询场景中支持提前终止。
* **高效链式操作**：每个操作都经过高度优化，减少数据拷贝和内存分配。

**高性能函数示例：**

```go
root.RegisterFunc("fastFilter", func(n xjson.Node) xjson.Node {
    return n.Filter(func(child xjson.Node) bool {
        // 直接获取原生 float64 值，无 Node 开销
        if price, ok := child.Get("price").RawFloat(); ok {
            return price < 20
        }
        return false
    })
})
```

## 📚 使用场景

### 业务规则封装

```go
// 注册库存检查函数
root.RegisterFunc("inStock", func(n xjson.Node) xjson.Node {
    return n.Filter(func(p xjson.Node) bool {
        return p.Get("stock").Int() > 0 &&
               p.Get("status").String() == "active"
    })
})

// 使用语义化查询
availableProducts := root.Query("/products[@inStock]")
```

### 数据转换管道

```go
import "strings"
import "math"

// 创建数据清洗管道
root.RegisterFunc("sanitize", func(n xjson.Node) xjson.Node {
    return n.Map(func(item xjson.Node) interface{} {
        return map[string]interface{}{
            "id":    item.Get("id").String(),
            "name":  strings.TrimSpace(item.Get("name").String()),
            "price": math.Round(item.Get("price").Float()*100) / 100,
        }
    })
})

// 应用清洗管道
cleanData := root.Query("/rawInput[@sanitize]")
```

### 复杂数据聚合

```go
// 计算平均分
root.RegisterFunc("withAvg", func(n xjson.Node) xjson.Node {
    return n.Map(func(user xjson.Node) interface{} {
        scoresNode := user.Get("scores")
        var sum int64 = 0
        scoresNode.ForEach(func(_ interface{}, score xjson.Node) {
            sum += score.Int()
        })
        avg := float64(sum) / float64(scoresNode.Len())
        return map[string]interface{}{
            "name":     user.Get("name").String(),
            "avgScore": math.Round(avg*10) / 10,
        }
    })
})

processedUsers := root.Query("/users[@withAvg]")
```

## 🌟 设计优势

* **概念简化**：只需理解 **Node** **单一概念，学习曲线平缓。**
* **灵活组合**：路径函数与流式操作无缝结合，表达能力强。
* **健壮可靠**：链式错误处理机制让代码更简洁且不易出错。
* **性能优异**：通过高效实现和原生访问 API 保持高性能。
* **类型安全**：完善的类型系统确保编译时的类型检查。
* **易于扩展**：模块化设计便于添加新功能。

## 🔄 升级指南

### 从 v0.0.2 升级到 v0.1.0

**主要变化：**

1. **增强写操作**：

   ```go
   // 新增 SetValue 方法用于直接设置值
   node.SetValue("new value")

   // 增强的 Set 方法具有更好的错误处理
   result := node.Set("key", "value")
   ```
2. **新增类型转换方法**：

   ```go
   // AsMap 用于对象转换
   objMap := node.AsMap()

   // MustAsMap 用于强制对象转换
   objMap := node.MustAsMap()

   // Keys 用于获取所有对象键
   keys := node.Keys()
   ```
3. **增强错误处理**：

   ```go
   // 更详细的错误信息
   if err := node.Error(); err != nil {
       fmt.Printf("路径 %s 处发生错误: %v\n", node.Path(), err)
   }
   ```
4. **性能改进**：

   ```go
   // 优化的 RawString 和 RawFloat 方法
   if str, ok := node.RawString(); ok {
       // 零拷贝字符串访问
   }
   ```

**兼容性说明：**

- 所有现有的查询语法继续有效
- 新功能完全向后兼容
- 性能改进不影响现有代码

## 📄 许可证

MIT License
