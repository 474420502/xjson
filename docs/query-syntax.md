# XJSON 查询语法指南

XJSON 支持类 XPath 的强大查询语法，专为 JSON 数据结构设计。本指南详细介绍了所有支持的语法特性和使用示例。

## 目录
- [基础语法](#基础语法)
- [路径选择器](#路径选择器)
- [数组操作](#数组操作)
- [过滤器表达式](#过滤器表达式)
- [高级功能](#高级功能)
- [完整示例](#完整示例)

## 基础语法

### 路径分隔符

XJSON 支持两种路径分隔符：

```go
// 使用点号（推荐）
result := doc.Query("store.book.title")

// 使用斜杠
result := doc.Query("store/book/title")
```

### 根节点

虽然支持 `$` 表示根节点，但在 `doc.Query()` 中通常可以省略：

```go
// 这两种写法等价
result1 := doc.Query("$.store.book")
result2 := doc.Query("store.book")
```

## 路径选择器

### 1. 直接子节点选择 (`.` 或 `/`)

选择直接子节点：

```go
// JSON: {"user": {"name": "Alice", "profile": {"age": 25}}}

name := doc.Query("user.name")           // "Alice"
age := doc.Query("user.profile.age")    // 25
```

### 2. 通配符选择 (`*`)

匹配任意键名或数组元素：

```go
// JSON: {"users": {"alice": {...}, "bob": {...}}}

allUsers := doc.Query("users.*")         // 所有用户对象
allFields := doc.Query("user.*")         // 用户对象的所有字段
```

### 3. 递归下降选择 (`//` 或 `..`)

在任意层级搜索指定键名：

```go
// JSON: {"data": {"items": [{"name": "item1"}, {"meta": {"name": "item2"}}]}}

allNames := doc.Query("//name")          // ["item1", "item2"]
allItems := doc.Query("//items")         // 数组内容
```

## 数组操作

### 1. 数组索引

使用方括号访问数组元素：

```go
// JSON: {"books": ["book1", "book2", "book3"]}

first := doc.Query("books[0]")           // "book1"
last := doc.Query("books[-1]")          // "book3" (负数索引)
second := doc.Query("books[1]")         // "book2"
```

### 2. 数组切片

支持 Python 风格的数组切片：

```go
// JSON: {"numbers": [1, 2, 3, 4, 5]}

first3 := doc.Query("numbers[0:3]")     // [1, 2, 3] (不包含索引3)
first2 := doc.Query("numbers[:2]")      // [1, 2]
last2 := doc.Query("numbers[3:]")      // [4, 5]
middle := doc.Query("numbers[1:4]")     // [2, 3, 4]
```

### 3. 全部元素

使用 `[*]` 选择数组的所有元素：

```go
// JSON: {"items": [{"id": 1}, {"id": 2}, {"id": 3}]}

allItems := doc.Query("items[*]")       // 所有元素
allIds := doc.Query("items[*].id")      // [1, 2, 3]
```

## 过滤器表达式

过滤器是 XJSON 最强大的功能，使用 `[?(<expression>)]` 语法。

### 1. 基础比较

支持所有标准比较操作符：

```go
// JSON: {"products": [
//   {"name": "laptop", "price": 999.99, "category": "electronics"},
//   {"name": "book", "price": 15.99, "category": "books"},
//   {"name": "phone", "price": 599.99, "category": "electronics"}
// ]}

// 价格比较
cheap := doc.Query("products[?(@.price < 50)]")
expensive := doc.Query("products[?(@.price >= 500)]")

// 字符串比较
electronics := doc.Query("products[?(@.category == 'electronics')]")
notBooks := doc.Query("products[?(@.category != 'books')]")
```

### 2. 逻辑操作

支持 `&&`（与）、`||`（或）、`!`（非）逻辑操作：

```go
// 复合条件
cheapElectronics := doc.Query("products[?(@.price < 700 && @.category == 'electronics')]")

// 多重条件
mobileDev := doc.Query("products[?(@.category == 'electronics' || @.category == 'mobile')]")

// 否定条件
notExpensive := doc.Query("products[?(!(@.price > 800))]")
```

### 3. 存在性检查 (`exists`)

检查字段是否存在：

```go
// JSON: {"users": [
//   {"name": "Alice", "email": "alice@example.com"},
//   {"name": "Bob"},
//   {"name": "Charlie", "email": "charlie@example.com", "phone": "123-456-7890"}
// ]}

hasEmail := doc.Query("users[?(@.email.exists)]")     // Alice, Charlie
hasPhone := doc.Query("users[?(@.phone.exists)]")     // Charlie
noEmail := doc.Query("users[?(!@.email.exists)]")     // Bob
```

### 4. 数组包含检查 (`includes`)

检查数组是否包含特定值：

```go
// JSON: {"articles": [
//   {"title": "Go Tutorial", "tags": ["go", "programming", "tutorial"]},
//   {"title": "Python Guide", "tags": ["python", "programming"]},
//   {"title": "Frontend Tips", "tags": ["javascript", "frontend", "web"]}
// ]}

goArticles := doc.Query("articles[?(@.tags.includes('go'))]")
programmingArticles := doc.Query("articles[?(@.tags.includes('programming'))]")
```

### 5. 嵌套路径过滤

过滤器可以访问嵌套对象：

```go
// JSON: {"orders": [
//   {"id": 1, "customer": {"name": "Alice", "vip": true}, "total": 299.99},
//   {"id": 2, "customer": {"name": "Bob", "vip": false}, "total": 59.99},
//   {"id": 3, "customer": {"name": "Charlie", "vip": true}, "total": 199.99}
// ]}

vipOrders := doc.Query("orders[?(@.customer.vip == true)]")
bigOrders := doc.Query("orders[?(@.total > 200 && @.customer.vip)]")
```

## 高级功能

### 1. 组合查询

可以组合多个查询步骤：

```go
// JSON: {"company": {"departments": [
//   {"name": "engineering", "employees": [
//     {"name": "Alice", "skills": ["go", "python"]},
//     {"name": "Bob", "skills": ["javascript", "react"]}
//   ]},
//   {"name": "design", "employees": [
//     {"name": "Charlie", "skills": ["photoshop", "illustrator"]}
//   ]}
// ]}}

// 找到所有会 Go 的员工
goDevs := doc.Query("company.departments[*].employees[?(@.skills.includes('go'))]")

// 工程部的第一个员工
firstEngineer := doc.Query("company.departments[?(@.name == 'engineering')].employees[0]")
```

### 2. 递归搜索与过滤

结合递归搜索和过滤器：

```go
// 在整个文档中搜索所有价格大于100的项目
expensiveItems := doc.Query("//[?(@.price > 100)]")

// 搜索所有包含特定标签的对象
tagged := doc.Query("//[?(@.tags.includes('important'))]")
```

### 3. 多级过滤

可以在查询路径的不同级别应用过滤器：

```go
// JSON: {"stores": [
//   {"name": "Store A", "products": [
//     {"name": "Product 1", "price": 10, "inStock": true},
//     {"name": "Product 2", "price": 20, "inStock": false}
//   ]},
//   {"name": "Store B", "products": [
//     {"name": "Product 3", "price": 15, "inStock": true}
//   ]}
// ]}

// 找到有库存产品的商店的库存产品
inStockProducts := doc.Query("stores[?(@.products[?(@.inStock == true)].length > 0)].products[?(@.inStock == true)]")
```

## 操作符参考

### 比较操作符
- `==` - 等于
- `!=` - 不等于
- `<` - 小于
- `<=` - 小于等于
- `>` - 大于
- `>=` - 大于等于

### 逻辑操作符
- `&&` - 逻辑与
- `||` - 逻辑或
- `!` - 逻辑非

### 特殊操作符
- `exists` - 字段存在性检查
- `includes` - 数组包含检查

## 完整示例

### 电商数据查询

```go
jsonData := `{
  "store": {
    "name": "Tech Store",
    "categories": [
      {
        "name": "laptops",
        "products": [
          {
            "id": 1,
            "name": "MacBook Pro",
            "price": 1299.99,
            "specs": {"ram": "16GB", "storage": "512GB"},
            "tags": ["apple", "premium", "laptop"],
            "inStock": true,
            "reviews": [
              {"rating": 5, "comment": "Excellent!"},
              {"rating": 4, "comment": "Good but expensive"}
            ]
          },
          {
            "id": 2,
            "name": "ThinkPad X1",
            "price": 999.99,
            "specs": {"ram": "8GB", "storage": "256GB"},
            "tags": ["lenovo", "business", "laptop"],
            "inStock": false
          }
        ]
      },
      {
        "name": "phones",
        "products": [
          {
            "id": 3,
            "name": "iPhone 13",
            "price": 799.99,
            "specs": {"storage": "128GB"},
            "tags": ["apple", "phone"],
            "inStock": true
          }
        ]
      }
    ]
  }
}`

doc, _ := xjson.ParseString(jsonData)

// 1. 基础查询
storeName := doc.Query("store.name").MustString()                    // "Tech Store"

// 2. 数组索引
firstCategory := doc.Query("store.categories[0].name").MustString()  // "laptops"

// 3. 通配符查询
allProducts := doc.Query("store.categories[*].products[*]")

// 4. 递归搜索
allPrices := doc.Query("//price")                                    // [1299.99, 999.99, 799.99]

// 5. 过滤器 - 价格范围
midRange := doc.Query("//products[*][?(@.price >= 800 && @.price < 1200)]")

// 6. 过滤器 - 标签包含
appleProducts := doc.Query("//products[*][?(@.tags.includes('apple'))]")

// 7. 过滤器 - 库存检查
inStock := doc.Query("//products[*][?(@.inStock == true)]")

// 8. 过滤器 - 存在性检查
hasReviews := doc.Query("//products[*][?(@.reviews.exists)]")

// 9. 复杂嵌套查询
premiumInStock := doc.Query("store.categories[*].products[?(@.price > 1000 && @.inStock == true && @.tags.includes('premium'))]")

// 10. 数组切片
firstTwoCategories := doc.Query("store.categories[:2]")

// 11. 组合查询
highRatedProducts := doc.Query("//products[*][?(@.reviews[*].rating >= 4)]")
```

### 用户数据管理

```go
userData := `{
  "users": [
    {
      "id": 1,
      "profile": {
        "name": "Alice Johnson",
        "email": "alice@example.com",
        "age": 28,
        "preferences": {
          "theme": "dark",
          "notifications": true
        }
      },
      "roles": ["user", "moderator"],
      "posts": [
        {"id": 101, "title": "Hello World", "likes": 15},
        {"id": 102, "title": "Go Tutorial", "likes": 42}
      ]
    },
    {
      "id": 2,
      "profile": {
        "name": "Bob Smith",
        "email": "bob@example.com",
        "age": 34
      },
      "roles": ["user"],
      "posts": [
        {"id": 201, "title": "JavaScript Tips", "likes": 23}
      ]
    }
  ]
}`

doc, _ := xjson.ParseString(userData)

// 查找所有管理员
moderators := doc.Query("users[?(@.roles.includes('moderator'))]")

// 查找年龄大于30的用户
adults := doc.Query("users[?(@.profile.age > 30)]")

// 查找有热门帖子的用户（超过20个赞）
popularUsers := doc.Query("users[?(@.posts[?(@.likes > 20)].length > 0)]")

// 查找所有帖子标题
allTitles := doc.Query("users[*].posts[*].title")

// 查找Alice的设置
alicePrefs := doc.Query("users[?(@.profile.name == 'Alice Johnson')].profile.preferences")
```

## 性能建议

1. **具体化路径**: 尽量使用具体的路径而不是通配符
2. **早期过滤**: 在路径早期应用过滤器以减少搜索范围
3. **避免深度递归**: 在大型文档中谨慎使用 `//` 操作符
4. **合理使用切片**: 使用数组切片限制结果集大小

## 错误处理

查询语法错误会在解析时返回详细的错误信息：

```go
result := doc.Query("invalid[syntax")
if !result.Exists() {
    // 处理查询失败的情况
}

// 或者检查具体错误
value, err := result.String()
if err != nil {
    log.Printf("Query error: %v", err)
}
```
