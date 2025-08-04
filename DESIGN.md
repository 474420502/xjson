# XJSON 设计文档

## 项目概述

XJSON 是一个支持类 XPath 语法、懒解析、高性能的 Go JSON 库，专为解决 gjson（快但只读）和标准库（灵活但慢）之间的性能与灵活性平衡问题而设计。

## 核心设计理念

### 1. 类 XPath 语法子集 (XPath-like Subset)
- 实现对 JSON 更有意义的强大查询子集
- 平衡查询能力与实现复杂性
- 支持递归下降、过滤器表达式等高级功能

### 2. 性能优先的懒解析 (Performance-First Lazy Parsing)
- 所有只读操作直接在原始 `[]byte` 上进行
- 零内存分配的路径扫描器
- 借鉴 gjson 的高性能原理

### 3. 写时物化 (Materialize-on-Write)
- 首次写操作时自动将 JSON 解析为可变 Go 结构
- 后续读写操作在物化树上进行
- 一次性完整解析，避免重复开销

### 4. 接口驱动与可移植性 (Interface-Driven & Portable)
- 清晰的接口设计 (IDocument, IResult)
- 核心概念可轻松映射到其他面向对象语言
- 为跨语言迁移奠定基础

## 核心组件

### Document 对象
- 表示整个 JSON 文档
- 用户交互的唯一入口
- 管理内部状态（原始字节 vs 物化树）

### Query 解析器
- 解析类 XPath 查询字符串的微型编译器
- 将查询字符串转换为可执行的"查询计划"

### Engine 执行引擎
- 根据查询计划执行遍历和筛选
- 支持在 `[]byte` 和物化树上执行

### Result 对象
- 封装查询结果集
- 支持零个、一个或多个节点匹配
- 提供便捷的数据提取方法

## 类 XPath 语法规范

### 基础路径选择
- `.` 或 `/` 作为路径分隔符：`data.users` 或 `data/users`
- `$` 表示根节点（在 `doc.Query()` 中可省略）
- `@` 表示当前节点（在过滤器中使用）

### 高级选择器
- `*` 通配符：匹配任意 key 或数组成员
- `..` 或 `//` 递归下降：`//name` 在任意层级搜索 key

### 数组操作
- 索引访问：`[0]`, `[-1]` (从后往前)
- 全部元素：`[*]`
- 切片操作：`[1:3]` (不含3), `[:2]`, `[1:]`

### 过滤器表达式 (核心功能)
- 语法：`[?(<expression>)]`
- 比较操作符：`==`, `!=`, `<`, `<=`, `>`, `>=`
- 特殊操作符：`exists`, `includes` (用于数组)
- 逻辑操作符：`&&`, `||`, `!`

### 示例
```
//books[?(@.price < 20 && @.category == "fiction")]/author
$.store.book[?(@.price > 10)].title
//user[@.id == 123]/profile/name
data.items[?(@.tags.includes("go"))].description
```

## 性能特性

### 懒解析优势
- 只读操作无内存分配
- 直接在原始字节上扫描
- 适合大型 JSON 文档的部分读取

### 写时物化机制
- 首次写操作触发完整解析
- 后续操作在内存树上执行
- 适合"先读后写"的使用模式

### 最佳实践
- 优先执行所有读取操作
- 批量执行写入操作
- 避免频繁的读写交替

## 接口设计

### IDocument 接口
```go
type IDocument interface {
    Query(xpath string) IResult
    Set(path string, value interface{}) error
    Delete(path string) error
    Bytes() ([]byte, error)
    String() (string, error)
}
```

### IResult 接口
```go
type IResult interface {
    String() (string, error)
    Int() (int, error)
    Float() (float64, error)
    Bool() (bool, error)
    Count() int
    First() IResult
    ForEach(func(index int, value IResult) bool)
    Exists() bool
}
```

## 跨语言移植考虑

### Python 映射
```python
doc.query('//user').first.string
doc['path.to.set'] = value  # 通过 __setitem__
```

### Java 映射
```java
doc.query("//user").first().asString()
doc.set("path.to.set", value)
```

## 实现计划

### 阶段 1：核心解析器
- [ ] 基础 JSON 扫描器
- [ ] XPath 语法解析器
- [ ] 查询执行引擎

### 阶段 2：懒解析实现
- [ ] 零分配路径遍历
- [ ] 过滤器表达式求值
- [ ] 结果集封装

### 阶段 3：写时物化
- [ ] 物化触发机制
- [ ] Set/Delete 操作
- [ ] 状态管理

### 阶段 4：优化与测试
- [ ] 性能基准测试
- [ ] 内存使用优化
- [ ] 完整测试覆盖

## 设计权衡

### 优势
- 读取性能接近 gjson
- 写入能力媲美标准库
- 强大的查询语法
- 跨语言移植友好

### 限制
- 写时物化会产生一次性开销
- 非线程安全（初始版本）
- XPath 子集不支持完整 XPath 1.0
- 对极大 JSON 文档的内存使用需要考虑

## 后续扩展

### 可能的增强功能
- 流式写入支持
- 线程安全版本
- 更丰富的过滤器函数
- JSON Schema 验证集成
- 增量更新机制
