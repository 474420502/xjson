# XJSON

[![Go Version](https://img.shields.io/badge/go-%3E%3D1.20-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/474420502/xjson)](https://goreportcard.com/report/github.com/474420502/xjson)

**XJSON** 是一个支持类 XPath 查询、懒解析和高性能读写操作的现代化 Go JSON 库。

## ✨ 核心特性

- 🚀 **高性能懒解析**: 只读操作零内存分配，直接在原始字节上执行
- 🔍 **类 XPath 查询**: 支持强大的路径表达式和过滤器，如 `//books[?(@.price < 20)]/title`
- ✏️ **便捷写操作**: 写时物化机制，无缝集成读写功能
- 🌐 **跨语言友好**: 接口驱动设计，易于移植到其他编程语言
- 📊 **内存高效**: 针对大型 JSON 文档优化，支持部分解析

## 🚀 快速开始

### 安装

```bash
go get github.com/474420502/xjson
```

### 基础用法

```go
package main

import (
    "fmt"
    "github.com/474420502/xjson"
)

func main() {
    // JSON 数据
    jsonStr := `{
        "store": {
            "book": [
                {"category": "fiction", "author": "Herman Melville", "title": "Moby Dick", "price": 8.99},
                {"category": "fiction", "author": "J.R.R. Tolkien", "title": "The Lord of the Rings", "price": 22.99},
                {"category": "science", "author": "Carl Sagan", "title": "Cosmos", "price": 15.99}
            ]
        }
    }`

    // 解析文档（懒解析，几乎零开销）
    doc, err := xjson.Parse([]byte(jsonStr))
    if err != nil {
        panic(err)
    }

    // 1. 基础路径查询
    title, _ := doc.Query("store.book[0].title").String()
    fmt.Println("第一本书:", title) // 输出: 第一本书: Moby Dick

    // 2. 使用过滤器查询（核心功能！）
    fictionBooks := doc.Query("//book[?(@.category == 'fiction')]/title")
    fmt.Printf("小说类图书数量: %d\n", fictionBooks.Count()) // 输出: 2
    
    fictionBooks.ForEach(func(index int, result xjson.IResult) bool {
        title, _ := result.String()
        fmt.Printf("  %d. %s\n", index+1, title)
        return true // 继续遍历
    })

    // 3. 复杂查询
    cheapBooks := doc.Query("//book[?(@.price < 20 && @.category == 'fiction')]/author")
    author, _ := cheapBooks.First().String()
    fmt.Println("便宜小说作者:", author) // 输出: Herman Melville

    // 4. 修改数据（触发写时物化）
    err = doc.Set("store.book[0].price", 12.99)
    if err != nil {
        panic(err)
    }

    // 5. 添加新书
    newBook := map[string]interface{}{
        "category": "programming",
        "author":   "Robert C. Martin", 
        "title":    "Clean Code",
        "price":    29.99,
    }
    err = doc.Set("store.book[3]", newBook)
    if err != nil {
        panic(err)
    }

    // 6. 输出修改后的 JSON
    result, _ := doc.String()
    fmt.Println("修改后的 JSON:")
    fmt.Println(result)
}
```

## 写操作：Materialize-on-Write

XJSON的核心特性是**懒解析**（Lazy Parsing）和**写时物化**（Materialize-on-Write）：

- **读操作**: 保持原始JSON字节，零分配，超高性能
- **写操作**: 首次Set/Delete时自动物化为Go结构，支持灵活修改

```go
package main

import (
    "fmt"
    "github.com/474420502/xjson"
)

func main() {
    jsonStr := `{"user": {"name": "John", "age": 30}, "active": true}`
    
    doc, _ := xjson.ParseString(jsonStr)
    
    // 读操作：不触发物化，保持高性能
    fmt.Printf("物化状态: %t\n", doc.IsMaterialized()) // false
    name := doc.Query("user.name").MustString()
    fmt.Printf("名字: %s\n", name) // John
    fmt.Printf("读取后物化状态: %t\n", doc.IsMaterialized()) // 仍然是 false
    
    // 写操作：自动触发物化
    doc.Set("user.age", 31)
    fmt.Printf("写入后物化状态: %t\n", doc.IsMaterialized()) // true
    
    // 后续写操作在已物化的结构上进行
    doc.Set("user.email", "john@example.com")
    doc.Set("user.preferences.theme", "dark")  // 自动创建嵌套结构
    doc.Delete("active")  // 删除字段
    
    // 获取最终结果
    result, _ := doc.String()
    fmt.Println("最终JSON:", result)
    // 输出: {"user":{"age":31,"email":"john@example.com","name":"John","preferences":{"theme":"dark"}}}
}
```

### 写操作 API

```go
// 设置值（支持路径创建）
err := doc.Set("user.profile.avatar", "avatar.jpg")

// 删除字段
err := doc.Delete("user.temp_data")

// 检查物化状态
if doc.IsMaterialized() {
    fmt.Println("文档已物化，后续操作在内存结构上进行")
}
```

### 更多示例

```go
// 数组操作
books := doc.Query("store.book")        // 获取所有书籍
firstBook := doc.Query("store.book[0]")  // 第一本书
lastBook := doc.Query("store.book[-1]")  // 最后一本书（负索引）
middleBooks := doc.Query("store.book[1:3]") // 切片 [1:3]

// 递归查询所有价格
allPrices := doc.Query("store..price")   // 递归查找所有price字段
allPrices.ForEach(func(i int, result xjson.IResult) bool {
    price, _ := result.Float()
    fmt.Printf("Price %d: %.2f\n", i, price)
    return true
})

// 数组切片和范围
firstTwo := doc.Query("store.book[:2]")    // 前两本书
fromSecond := doc.Query("store.book[1:]")  // 从第二本到最后
middle := doc.Query("store.book[1:3]")     // 中间范围

// 递归搜索所有名称字段
allNames := doc.Query("..name")
fmt.Printf("Found %d name fields\n", allNames.Count())
```

// 通配符查询
allCategories := doc.Query("store.book[*].category")

// 存在性检查
hasISBN := doc.Query("//book[?(@.isbn.exists)]/title")

// 数组包含查询（假设有 tags 字段）
goBooks := doc.Query("//book[?(@.tags.includes('go'))]/title")
```

## 📖 查询语法

XJSON 支持类 XPath 的强大查询语法：

### 基础语法
- `.` 或 `/` - 路径分隔符
- `*` - 通配符，匹配任意键或数组元素
- `//` - 递归下降，在任意层级搜索

### 数组操作
- `[0]` - 数组索引（支持负数 `[-1]`）
- `[*]` - 所有数组元素
- `[1:3]` - 数组切片（不包含结束索引）
- `[:2]` - 从开头到索引 2
- `[1:]` - 从索引 1 到结尾

### 过滤器表达式（强大功能）
- `[?(@.price < 20)]` - 数值比较
- `[?(@.category == "fiction")]` - 字符串相等
- `[?(@.isbn.exists)]` - 字段存在性检查
- `[?(@.tags.includes("go"))]` - 数组包含检查
- `[?(@.price > 10 && @.category == "fiction")]` - 逻辑与
- `[?(@.price < 5 || @.rating > 4)]` - 逻辑或

### 操作符
- 比较: `==`, `!=`, `<`, `<=`, `>`, `>=`
- 逻辑: `&&`, `||`, `!`
- 特殊: `exists`, `includes`

## 🎯 设计优势

### 性能优化
- **懒解析**: 只读操作直接在原始字节上执行，零内存分配
- **写时物化**: 仅在首次写操作时解析完整 JSON 树
- **智能缓存**: 查询结果复用，避免重复计算

### 使用灵活
- **渐进式**: 从简单的路径查询到复杂的过滤表达式
- **类型安全**: 丰富的类型转换方法和错误处理
- **链式调用**: 自然的 API 设计

### 跨语言支持
- **接口驱动**: 核心概念可轻松映射到其他语言
- **一致性**: 语法和行为在不同语言实现中保持一致

## 📚 完整文档

- [设计文档](DESIGN.md) - 详细的架构设计和原理说明
- [查询语法指南](docs/query-syntax.md) - 完整的查询语法参考
- [性能指南](docs/performance.md) - 性能优化和最佳实践
- [API 参考](https://pkg.go.dev/github.com/474420502/xjson) - 完整的 API 文档

## 🔄 对比其他库

| 特性 | XJSON | gjson | 标准库 encoding/json |
|------|-------|-------|---------------------|
| 只读性能 | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐ |
| 查询语法 | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐ |
| 写入支持 | ⭐⭐⭐⭐⭐ | ❌ | ⭐⭐⭐⭐⭐ |
| 内存效率 | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ |
| 易用性 | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐ |

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！请参阅 [贡献指南](CONTRIBUTING.md)。

## 📄 许可证

MIT License - 详见 [LICENSE](LICENSE) 文件。

## 🚧 项目状态

当前版本: **v0.3.0-rc** (生产就绪候选版本)

- [x] ✅ 核心架构设计
- [x] ✅ 基础解析器实现  
- [x] ✅ XPath 语法解析器
- [x] ✅ 懒解析引擎
- [x] ✅ 写时物化机制
- [x] ✅ 过滤器表达式支持
- [x] ✅ 函数调用支持 (exists, includes)
- [x] ✅ 递归查询支持 (..)
- [x] ✅ 完整测试覆盖
- [x] ✅ 性能基准测试
- [x] ✅ 文档和示例完善
- [ ] 🔄 最终性能优化
- [ ] 📋 v1.0.0 生产版本

### 性能表现

| 操作类型 | 当前性能 | 内存分配 |
|---------|---------|---------|
| 简单查询 | 1292 ns/op | 1048 B/op |
| 过滤查询 | 8474 ns/op | 4734 B/op |
| 递归查询 | 3622 ns/op | 2448 B/op |
| 数组切片 | 2778 ns/op | 2760 B/op |
| 写操作 | 1430 ns/op | 1097 B/op |

### 功能支持矩阵

| 功能 | 支持状态 | 示例 |
|-----|---------|------|
| 简单路径 | ✅ | `user.name` |
| 数组访问 | ✅ | `products[0]`, `products[-1]` |
| 数组切片 | ✅ | `products[1:3]`, `items[:5]` |
| 递归查询 | ✅ | `..price`, `//user` |
| 过滤器 | ✅ | `products[?(@.price < 100)]` |
| 布尔过滤 | ✅ | `users[?(@.active == true)]` |
| 逻辑操作 | ✅ | `items[?(@.price > 10 && @.available == true)]` |
| 函数调用 | ✅ | `data[?exists(@.field)]` |
| 写操作 | ✅ | `Set()`, `Delete()` |
| 类型转换 | ✅ | `String()`, `Int()`, `Float()`, `Bool()` |

---

⭐ 如果这个项目对您有帮助，请给我们一个 Star！
