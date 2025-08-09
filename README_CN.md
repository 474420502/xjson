# XJSON - 统一节点模型JSON处理器 (v0.0.2 修订版)

**XJSON** **是一个强大的 Go JSON 处理库，采用完全统一的** **Node** **模型，支持路径函数、流式操作和灵活的查询语法。**

## ✨ 核心特性

* **🎯** **单一节点类型**：所有操作都基于 **xjson.Node**，无 **Result** **类型。**
* **🧩** **路径函数**：通过 **/path[@func]/subpath** **语法将自定义逻辑注入查询。**
* **🔗** **链式操作**：支持流畅的函数注册、查询和数据操作。
* **🌀** **健壮的错误处理**：通过 **node.Error()** **在链式调用末尾统一检查错误。**
* **⚡️** **性能导向**：通过高效的链式操作和原生值访问实现零拷贝级别的性能。
* **🌟** **通配符查询**：支持 **`*`** 通配符和复杂的路径表达式。

## 🚀 快速开始

```go
package main

import (
	"fmt"
	"github.com/474420502/xjson"
)

func main() {
	data := `{
		"store": {
			"books": [
				{"title": "Moby Dick", "price": 8.99, "tags": ["classic", "adventure"]},
				{"title": "Clean Code", "price": 29.99, "tags": ["programming"]},
				{"title": "Go Programming", "price": 35.50, "tags": ["programming", "go"]},
				{"title": "Design Patterns", "price": 45.00, "tags": ["programming", "design"]}
			],
			"electronics": {
				"laptops": [
					{"name": "MacBook Pro", "price": 1999.99, "in_stock": true},
					{"name": "ThinkPad X1", "price": 1599.99, "in_stock": false}
				]
			}
		},
		"authors": [
			{"name": "Herman Melville", "nationality": "American"},
			{"name": "Robert C. Martin", "nationality": "American"}
		]
	}`

	// 1. 解析并检查初始错误
	root, err := xjson.Parse(data)
	if err != nil {
		panic(err)
	}

    // 2. 注册函数
	root.RegisterFunc("cheap", func(n xjson.Node) xjson.Node {
		return n.Filter(func(child xjson.Node) bool {
			price, _ := child.Get("price").RawFloat()
			return price < 20
		})
	}).RegisterFunc("tagged", func(n xjson.Node) xjson.Node {
		return n.Filter(func(child xjson.Node) bool {
			return child.Get("tags").Contains("adventure")
		})
	}).RegisterFunc("expensive", func(n xjson.Node) xjson.Node {
		return n.Filter(func(child xjson.Node) bool {
			price, _ := child.Get("price").RawFloat()
			return price > 30
		})
	})

	// 3. 基础路径查询示例
	fmt.Println("=== 基础路径查询 ===")
	
	// 键访问
	booksNode := root.Query("/store/books")
	fmt.Printf("Books array type: %v\n", booksNode.Type())
	
	// 数组索引
	firstBook := root.Query("/store/books[0]/title").String()
	fmt.Printf("First book title: %s\n", firstBook)
	
	// 数组切片
	middleBooks := root.Query("/store/books[1:3]/title").Strings()
	fmt.Printf("Middle books (1:3): %v\n", middleBooks)
	
	// 获取最后两本书
	lastTwoBooks := root.Query("/store/books[-2:]/title").Strings()
	fmt.Printf("Last two books: %v\n", lastTwoBooks)

	// 4. 函数调用示例
	fmt.Println("\n=== 函数调用示例 ===")
	
	// 使用路径函数查询
	cheapTitles := root.Query("/store/books[@cheap]/title").Strings()
	if err := root.Error(); err != nil {
		fmt.Println("查询失败:", err)
		return
	}
	fmt.Println("Cheap books:", cheapTitles) // ["Moby Dick"]

	// 链式函数调用
	expensiveProgrammingBooks := root.Query("/store/books[@expensive][@tagged='programming']/title").Strings()
	fmt.Println("Expensive programming books:", expensiveProgrammingBooks)

	// 5. 通配符和高级语法
	fmt.Println("\n=== 通配符和高级语法 ===")
	
	// 通配符查询
	allStoreItems := root.Query("/store/*")
	fmt.Printf("All store items count: %d\n", allStoreItems.Len())
	
	// 递归下降查询
	allAuthors := root.Query("//name").Strings()
	fmt.Println("All authors in document:", allAuthors)
	
	// 特殊字符键名处理（假设有这样的数据）
	specialKeyData := `{"data/user-profile": {"name": "John", "age": 30}}`
	specialRoot, _ := xjson.Parse(specialKeyData)
	userName := specialRoot.Query("/['data/user-profile']/name").String()
	fmt.Printf("User name from special key: %s\n", userName)

	// 6. 修改数据
	fmt.Println("\n=== 数据修改 ===")
	root.Query("/store/books[@tagged]").Set("price", 9.99)
	if err := root.Error(); err != nil {
		fmt.Println("修改失败:", err)
		return
	}

	// 7. 输出最终结果
	fmt.Println("\n=== 最终JSON ===")
	fmt.Println(root.String())
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
  
    // 函数支持
    RegisterFunc(name string, fn UnaryPathFunc) Node
    CallFunc(name string) Node
    RemoveFunc(name string) Node
    Apply(fn PathFunc) Node
    GetFuncs() *map[string]func(Node) Node
    
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

*   **语法**: `/`
*   **描述**: 代表 JSON 数据的根节点。
*   **示例**: `/store` 从根节点获取 `store` 键的值。

**注意**: `/store/books` 和 `store/books` 这两种写法是等效的。

**4.2. 键访问**

标准的对象字段访问通过键名直接完成。任何符合 Go 语言标识符习惯的字符串都可以直接作为路径段。

*   **语法**: `/key1/key2`
*   **示例**: `/store/books`，这段路径会依次获取 `store` 键和 `books` 键。

**4.3. 数组访问**

通过方括号 `[...]` 访问数组元素，支持单个索引和范围切片。

*   **索引访问**:
    *   **语法**: `[<index>]`
    *   **描述**: 获取单个数组元素，索引从 0 开始。
    *   **示例**: `/store/books[0]`，获取 `books` 数组的第一个元素。

*   **切片访问**:
    *   **语法**:
        *   `[start:end]`: 获取从 `start` 到 `end-1` 的元素。
        *   `[start:]`: 获取从 `start` 到末尾的元素。
        *   `[:end]`: 获取从开头到 `end-1` 的元素。
        *   `[-N:]`: 获取最后 N 个元素。
    *   **描述**: 获取数组的一个子集，并返回一个包含这些元素的新数组节点。
    *   **示例**: `/store/books[1:3]`，返回一个包含 `books` 数组中第二个和第三个元素的新数组。

**4.4. 函数调用**

在路径中通过 `[@<funcName>]` 语法调用已注册的函数。函数提供了一种强大的数据处理和过滤机制。

*   **语法**: `[@<函数名>]`
*   **标志符**: `@` 符号明确表示这是一个函数调用。
*   **要求**: 函数必须已通过 `RegisterFunc` 注册到节点上。
*   **示例**: `/store/books[@cheap]/title`，在 `books` 数组上调用 `cheap` 函数，并从结果中提取 `title`。

**4.5. 通配符**

星号 `*` 作为通配符，用于匹配一个节点下的所有直接子元素。

*   **语法**: `*`
*   **对象上的行为**: 匹配对象的所有值，并返回一个包含这些值的新数组节点。
*   **数组上的行为**: 匹配数组的所有元素，并返回该数组自身。
*   **示例**: `/store/*/title`，获取 `store` 对象下所有直接子节点（在这里是 `books` 数组）的 `title` 字段。

#### **高级语法**

**5.1. 链式与混合语法**

所有核心组件都可以自由组合，形成强大的链式查询。解析器会从左到右依次执行每个操作。

*   **示例**: `/store/books[@filter][0]/name`
    1.  `/store/books`: 获取 `books` 数组。
    2.  `[@filter]`: 在该数组上调用 `filter` 函数。
    3.  `[0]`: 获取函数返回结果（应为一个数组）的第一个元素。
    4.  `/name`: 获取该元素的 `name` 字段。

**5.2. 特殊字符键名处理**

当对象键名包含 `/`, `.`, `[`, `]` 等特殊字符或非字母数字时，必须使用方括号和引号 `['<key>']` 或 `["<key>"]` 的形式来界定。

*   **语法**: `['<键名>']` 或 `["<键名>"]`
*   **键名包含斜杠**: `/['/api/v1/users']`
*   **键名包含点号**: `/data/['user.profile']/name`
*   **键名包含引号**:
    *   如果键名为 `a"key`，使用 `['a"key']`。
    *   如果键名为 `a'key`，使用 `["a'key"]`。
*   **与普通路径混合**: `/data['user-settings']/theme`

**5.3. 递归下降**

双斜杠 `//` 用于在当前节点及其所有后代中进行深度搜索，查找匹配的键。

*   **语法**: `//key`
*   **描述**: 与 `/` 只在直接子节点中查找不同，`//` 会遍历整个子树，将所有匹配 `key` 的节点收集到一个新的数组节点中。
*   **示例**: `//author` 将从根节点开始，查找所有层级下的 `author` 字段。

> **性能警告**：递归下降 `//` 是一个非常强大但开销极大的操作。因为它需要遍历一个节点下的整个子树，当处理大型或深层嵌套的 JSON 数据时，可能会成为性能瓶颈。建议仅在数据结构不确定或确实需要全局搜索时使用，在性能敏感的场景下应优先使用精确路径。

#### **语法速查表**

| 分类 | 语法 | 描述 | 示例 |
| :--- | :--- | :--- | :--- |
| **基础** | `/` | 路径段之间的分隔符。 | `/store/books` |
| | `key` | 访问对象的字段。 | `/store` |
| **数组** | `[<index>]` | 按索引访问数组元素。 | `[0]`, `[-1]` |
| | `[start:end]` | 按范围访问数组元素（切片）。 | `[1:3]`, `[:-1]` |
| **函数** | `[@<name>]` | 调用已注册的路径函数。 | `[@cheap]`, `[@inStock]` |
| **高级** | `*` | 匹配对象或数组的所有直接子元素。 | `/store/*` |
| | `//key` | 递归搜索所有后代节点中的 `key` (性能开销大)。 | `//author` |
| **特殊字符** | `['<key>']` | 界定包含特殊字符的键名。 | `['user.profile']`|
| | `["<key>"]`| 界定包含单引号的键名。 | `["a'key"]` |

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

| 方法 | 描述 | 示例 |
|------|------|------|
| **RegisterFunc(name, fn)** | 注册路径函数 | `root.RegisterFunc("cheap", filterCheap)` |
| **CallFunc(name)** | 直接调用函数 | `root.CallFunc("cheap")` |
| **RemoveFunc(name)** | 移除函数 | `root.RemoveFunc("cheap")` |
| **Apply(fn)** | 立即应用函数 | `root.Apply(predicateFunc)` |
| **GetFuncs()** | 获取已注册函数 | `funcs := root.GetFuncs()` |
| **Error() error** | 返回链式调用中的第一个错误 | `if err := n.Error(); err != nil { ... }` |

### 流式操作

| 方法 | 描述 | 示例 |
|------|------|------|
| **Filter(fn)** | 过滤节点集合 | `n.Filter(func(n Node) bool { return n.Get("active").Bool() })` |
| **Map(fn)** | 转换节点集合 | `n.Map(func(n Node) interface{} { return n.Get("name").String() })` |
| **ForEach(fn)** | 遍历节点集合 | `n.ForEach(func(i interface{}, v Node) { fmt.Println(v.String()) })` |

### 原生值访问

| 方法 | 描述 | 示例 |
|------|------|------|
| **RawFloat()** | 直接获取 float64 值 | `if price, ok := n.RawFloat(); ok { ... }` |
| **RawString()** | 直接获取 string 值 | `if name, ok := n.RawString(); ok { ... }` |
| **Strings()** | 获取字符串数组 | `tags := n.Strings()` |
| **Contains(value)** | 检查是否包含字符串 | `if n.Contains("target") { ... }` |
| **AsMap()** | 获取节点为 map | `obj := n.AsMap()` |

### 强制类型转换

| 方法 | 描述 | 示例 |
|------|------|------|
| **MustString()** | 获取字符串值，失败时 panic | `value := n.MustString()` |
| **MustFloat()** | 获取 float64 值，失败时 panic | `value := n.MustFloat()` |
| **MustInt()** | 获取 int64 值，失败时 panic | `value := n.MustInt()` |
| **MustBool()** | 获取 bool 值，失败时 panic | `value := n.MustBool()` |
| **MustTime()** | 获取 time.Time 值，失败时 panic | `value := n.MustTime()` |
| **MustArray()** | 获取数组值，失败时 panic | `value := n.MustArray()` |
| **MustAsMap()** | 获取 map 值，失败时 panic | `value := n.MustAsMap()` |

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

### 从 v0.0.1 升级到 v0.0.2

**主要变化：**

1. **函数系统更新**：
   ```go
   // 旧版本 (已弃用)
   root.Func("name", fn)
   
   // 新版本 (推荐)
   root.RegisterFunc("name", fn)
   ```

2. **新增 Apply 方法**：
   ```go
   // 立即应用函数
   result := root.Apply(func(n xjson.Node) bool {
       return n.Get("active").Bool()
   })
   ```

3. **类型系统增强**：
   ```go
   // 使用具体的函数类型
   var filterFunc xjson.PredicateFunc = func(n xjson.Node) bool {
       return n.Get("price").Float() > 10
   }
   
   var transformFunc xjson.TransformFunc = func(n xjson.Node) interface{} {
       return n.Get("name").String()
   }
   ```

4. **通配符支持**：
   ```go
   // 新增通配符查询
   result := root.Query("/store/*/title")
   ```

5. **新增方法**：
   ```go
   // Must* 方法在类型不匹配时 panic
   value := root.MustString()
   
   // AsMap 用于对象转换
   obj := root.AsMap()
   
   // GetFuncs 用于获取已注册函数
   funcs := root.GetFuncs()
   ```

**兼容性说明：**
- 旧的 `Func()` 方法仍然可用，但已被标记为弃用
- 所有现有的查询语法继续有效
- 新功能完全向后兼容

## 📄 许可证

MIT License
