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
				{"title": "Clean Code", "price": 29.99, "tags": ["programming"]}
			]
		}
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
	})

	// 3. 使用路径函数查询
	cheapTitles := root.Query("/store/books[@cheap]/title").Strings()
	if err := root.Error(); err != nil {
		fmt.Println("查询失败:", err)
		return
	}
	fmt.Println("Cheap books:", cheapTitles) // ["Moby Dick"]

	// 4. 修改数据
	root.Query("/store/books[@tagged]").Set("price", 9.99)
	if err := root.Error(); err != nil {
		fmt.Println("修改失败:", err)
		return
	}

	// 5. 输出结果
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
    
    // 类型转换
    String() string
    Float() float64
    Int() int64
    Bool() bool
    Array() []Node
    Interface() interface{}
    
    // 原生值访问 (性能优化)
    RawFloat() (float64, bool)
    RawString() (string, bool)
    // ... 其他原生类型
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

**支持丰富的查询语法：**

```go
// 基本路径
"/store/books/0/title"

// 数组索引
"/store/books[0]/title"

// 函数调用
"/store/books[@cheap]/title"

// 通配符
"/store/*/title"  // 匹配 store 下所有子节点的 title

// 混合语法
"/store/books[@filter][0]/name"
```

### 5. 函数注册和调用

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
```

## 🛠️ 完整 API 参考

### 函数管理

| 方法 | 描述 | 示例 |
|------|------|------|
| **RegisterFunc(name, fn)** | 注册路径函数 | `root.RegisterFunc("cheap", filterCheap)` |
| **CallFunc(name)** | 直接调用函数 | `root.CallFunc("cheap")` |
| **RemoveFunc(name)** | 移除函数 | `root.RemoveFunc("cheap")` |
| **Apply(fn)** | 立即应用函数 | `root.Apply(predicateFunc)` |
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

**兼容性说明：**
- 旧的 `Func()` 方法仍然可用，但已被标记为弃用
- 所有现有的查询语法继续有效
- 新功能完全向后兼容

## 📄 许可证

MIT License
