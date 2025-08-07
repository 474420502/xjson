**好的，收到指示。这次我将提供完整的、未删节的最终版设计文档，包含您特别要求的“使用场景”部分。**

---

# XJSON - 统一节点模型JSON处理器 (v1.1 修订版)

**XJSON** **现在采用完全统一的** **Node** **模型，支持路径函数和流式操作。此版本采纳了社区建议，引入了更健壮的错误处理机制和更精确的性能模型描述。**

## ✨ 核心变更

* **🎯** **单一节点类型**：所有操作都基于 **xjson.Node**，无 **Result** **类型。**
* **🧩** **路径函数**：通过 **/path[@func]/subpath** **语法将自定义逻辑注入查询。**
* **🔗** **链式注册**：**node.Func("name", fn)** **支持流畅的函数定义。**
* **🌀** **健壮的错误处理**：通过 **node.Error()** **在链式调用末尾统一检查错误。**
* **⚡️** **性能导向**：通过高效的链式操作和原生值访问实现零拷贝级别的性能。

## 🚀 快速开始

**code**go

```
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

    // 2. 链式注册函数
	root.Func("cheap", func(n xjson.Node) xjson.Node {
		return n.Filter(func(child xjson.Node) bool {
			price, _ := child.Get("price").RawFloat()
			return price < 20
		})
	}).Func("tagged", func(n xjson.Node) xjson.Node {
		return n.Filter(func(child xjson.Node) bool {
			return child.Get("tags").Contains("adventure")
		})
	})

	// 3. 使用路径函数查询
	// 注意：在链式调用末尾检查错误
	cheapTitles := root.Query("/store/books[@cheap]/title").Strings()
	if err := root.Error(); err != nil {
		fmt.Println("查询失败:", err)
		return
	}
	fmt.Println("Cheap books:", cheapTitles) // ["Moby Dick"]

	// 4. 修改数据并检查错误
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

**所有 JSON 元素（对象、数组、字符串、数字等），包括查询结果集，都由** **Node** **接口表示。这极大地简化了 API。**

**code**Go

```
type Node interface {
    // 查询方法
    Query(path string) Node
    Get(key string) Node
    Index(i int) Node
  
    // 流式操作
    Filter(fn func(Node) bool) Node
    Map(fn func(Node) interface{}) Node
    ForEach(fn func(int, Node)) 
  
    // 写操作
    Set(key string, value interface{}) Node
    Append(value interface{}) Node
  
    // 函数支持
    Func(name string, fn func(Node) Node) Node
    CallFunc(name string) Node

    // 错误处理
    Error() error

    // 原生值访问 (用于性能优化)
    RawFloat() (float64, bool)
    RawString() (string, bool)
    // ... 其他原生类型
}
```

### 2. 错误处理

**XJSON 采用链式调用友好的错误处理模式。任何操作（如** **Query**, **Get**, **Set**）在执行失败时（例如路径不存在、类型不匹配），都会在内部记录第一个发生的错误。您可以在一系列操作后，通过调用 **Error()** **方法一次性检查整个链条中是否出现了问题。**

**code**Go

```
// 无需在每一步都检查 err
value := root.Query("/path/that/does/not/exist").Get("key").Int()

// 在最后统一检查
if err := root.Error(); err != nil {
    // 处理错误: "path /path/that/does/not/exist not found"
    fmt.Println("操作链失败:", err)
}
```

### 3. 路径函数语法

**code**Go

```
// 基本格式
/path/to/[@funcName]/remaining/path

// 实际用例
node.Query("/data/items[@filter]/name")

// 函数可以出现在任意位置
node.Query("/data[@process]/items[@filter]")
```

### 4. 函数签名设计

 **路径函数为** **type PathFunc func(Node) Node**，它接收一个 **Node** **并必须返回一个** **Node**。为了保证链式调用的健壮性，我们**强烈建议**：

> **路径函数应始终返回一个“节点集合”类型的** **Node**，即使该集合为空或只包含一个元素。

**一个代表集合的** **Node** **可以安全地接受后续的** **.Filter()**, **.Get()**（隐式映射）或 **/subpath**（路径查询）操作。如果函数返回一个终值（如字符串或数字节点），后续的链式调用可能会意外失败。

**示例：**

**code**Go

```
type PathFunc func(Node) Node

// ✅ 推荐: 总是返回一个可以被继续查询的节点 (即使是空的)
func filterExpensive(n xjson.Node) xjson.Node {
    // Filter 本身就返回一个节点集合，符合要求
    return n.Filter(func(child xjson.Node) bool {
        price, _ := child.Get("price").RawFloat()
        return price > 50
    })
}

// ❌ 不推荐: 返回一个终值节点，这会中断链式查询
func getFirstPrice(n xjson.Node) xjson.Node {
    // .Index(0).Get("price") 可能返回一个数字节点
    // 后续的 /name 查询会失败
    return n.Index(0).Get("price") 
}

// 使用时的区别
// root.Query("/books[@getFirstPrice]/author") // 可能失败！
```

## 🛠️ 完整API参考

### 函数管理

| **方法**             | **描述**                       | **示例**                                    |
| -------------------------- | ------------------------------------ | ------------------------------------------------- |
| **Func(name, fn)**   | **注册路径函数**               | **node.Func("cheap", filterCheap)**         |
| **CallFunc(name)**   | **直接调用函数**               | **node.CallFunc("cheap")**                  |
| **RemoveFunc(name)** | **移除函数**                   | **node.RemoveFunc("cheap")**                |
| **Error() error**    | **返回链式调用中的第一个错误** | **if err := n.Error(); err != nil { ... }** |

## ⚡ 性能优化

* **函数缓存**：编译后的路径（包含函数占位符）会被缓存，以加速重复查询。
* **高效链式操作**：虽然 XJSON 采用及早求值（Eager Evaluation），但每个操作都经过高度优化。**Node** **的内部表示被设计为尽可能减少数据拷贝和内存分配，从而实现高效的中间步骤。这提供了可预测的性能和相对简单的执行模型。**
* **短路优化**：在某些过滤和查询场景中支持提前终止，避免不必要的计算。
* **原生值访问**：对于性能敏感的代码路径，可以使用 **Raw** **系列方法（如** **RawFloat**, **RawString**）。这些方法直接从底层内存访问数据，避免创建中间 **Node** **对象，实现零拷贝读取。**

**code**Go

```
// 高性能函数示例，使用 RawFloat 避免分配
root.Func("fastFilter", func(n xjson.Node) xjson.Node {
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

**可以将复杂的业务逻辑封装成可复用的路径函数，使查询语义化。**

**code**Go

```
// 注册一个库存检查函数，用于筛选活跃且有库存的商品
root.Func("inStock", func(n xjson.Node) xjson.Node {
    return n.Filter(func(p xjson.Node) bool {
        // 注意：为简洁，此处省略了错误检查。在生产代码中，
        // 可在最后调用 root.Error() 来捕获整个链的错误。
        return p.Get("stock").Int() > 0 &&
               p.Get("status").String() == "active"
    })
})

// 使用语义化的路径函数进行查询
availableProducts := root.Query("/products[@inStock]")
```

### 数据转换管道

**使用路径函数和** **Map** **操作，可以构建强大的数据清洗和转换管道。**

**code**Go

```
import "strings"
import "math"

// 创建一个数据清洗管道函数
root.Func("sanitize", func(n xjson.Node) xjson.Node {
    return n.Map(func(item xjson.Node) interface{} {
        // 在 Map 的回调中，返回一个 map 或 struct 来定义新的结构
        return map[string]interface{}{
            "id":    item.Get("id").String(),
            "name":  strings.TrimSpace(item.Get("name").String()),
            "price": math.Round(item.Get("price").Float()*100) / 100, // 四舍五入到2位小数
        }
    })
})

// 对原始数据应用清洗管道，得到规整的数据
cleanData := root.Query("/rawInput[@sanitize]")
// cleanData 现在是一个包含清洗后对象的节点
```

## 🌟 设计优势

* **概念简化**：只需理解 **Node** **单一概念，学习曲线平缓。**
* **灵活组合**：路径函数与流式操作无缝结合，表达能力强。
* **健壮可靠**：链式错误处理机制让代码更简洁且不易出错。
* **性能优异**：通过高效实现和原生访问 API 保持高性能。
