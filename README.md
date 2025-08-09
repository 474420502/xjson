# XJSON - Unified Node Model JSON Processor (v0.0.2)

**XJSON** is a powerful Go JSON processing library that features a completely unified **Node** model with support for path functions, streaming operations, and flexible query syntax.

## ✨ Core Features

* **🎯 Unified Node Type**: All operations are based on **xjson.Node**, no **Result** type
* **🧩 Path Functions**: Inject custom logic into queries using **/path[@func]/subpath** syntax
* **🔗 Chained Operations**: Support for fluent function registration, queries, and data manipulation
* **🌀 Robust Error Handling**: Unified error checking at the end of chained calls via **node.Error()**
* **⚡ Performance Oriented**: Zero-copy level performance through efficient chained operations and native value access
* **🌟 Wildcard Queries**: Support for **`*`** wildcards and complex path expressions

## 🚀 Quick Start

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

	// 1. Parse and check initial error
	root, err := xjson.Parse(data)
	if err != nil {
		panic(err)
	}

    // 2. Register functions
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

	// 3. Query using path functions
	cheapTitles := root.Query("/store/books[@cheap]/title").Strings()
	if err := root.Error(); err != nil {
		fmt.Println("Query failed:", err)
		return
	}
	fmt.Println("Cheap books:", cheapTitles) // ["Moby Dick"]

	// 4. Modify data
	root.Query("/store/books[@tagged]").Set("price", 9.99)
	if err := root.Error(); err != nil {
		fmt.Println("Modification failed:", err)
		return
	}

	// 5. Output result
	fmt.Println(root.String())
}
```

## 💡 Core Design

### 1. Unified Node Model

**All JSON elements (objects, arrays, strings, numbers, etc.), including query result sets, are represented by the Node interface.**

```go
type Node interface {
    // Basic access
    Type() NodeType
    IsValid() bool
    Error() error
    Path() string
    Raw() string
    
    // Query methods
    Query(path string) Node
    Get(key string) Node
    Index(i int) Node
  
    // Streaming operations
    Filter(fn PredicateFunc) Node
    Map(fn TransformFunc) Node
    ForEach(fn func(keyOrIndex interface{}, value Node)) 
    Len() int
  
    // Write operations
    Set(key string, value interface{}) Node
    Append(value interface{}) Node
  
    // Function support
    RegisterFunc(name string, fn UnaryPathFunc) Node
    CallFunc(name string) Node
    RemoveFunc(name string) Node
    Apply(fn PathFunc) Node
    
    // Type conversion
    String() string
    Float() float64
    Int() int64
    Bool() bool
    Array() []Node
    Interface() interface{}
    
    // Native value access (performance optimization)
    RawFloat() (float64, bool)
    RawString() (string, bool)
    // ... other native types
}
```

### 2. Function Type System

**XJSON provides multiple function types to support different operation scenarios:**

```go
// Path function - generic function container
type PathFunc interface{}

// Unary path function - node-to-node transformation
type UnaryPathFunc func(node Node) Node

// Predicate function - for filtering operations
type PredicateFunc func(node Node) bool

// Transform function - for mapping operations
type TransformFunc func(node Node) interface{}
```

### 3. Error Handling

**XJSON adopts a chain-call-friendly error handling pattern:**

```go
// No need to check err at every step
value := root.Query("/path/that/does/not/exist").Get("key").Int()

// Check uniformly at the end
if err := root.Error(); err != nil {
    fmt.Println("Operation chain failed:", err)
}
```

### 4. Path Query Syntax

**Supports rich query syntax:**

```go
// Basic path
"/store/books/0/title"

// Array indexing
"/store/books[0]/title"

// Function calls
"/store/books[@cheap]/title"

// Wildcards
"/store/*/title"  // Match title under all child nodes of store

// Mixed syntax
"/store/books[@filter][0]/name"
```

### 5. Function Registration and Calling

**The new version's function system is more powerful and flexible:**

```go
// Register function (recommended way)
root.RegisterFunc("filterFunc", func(n xjson.Node) xjson.Node {
    return n.Filter(func(child xjson.Node) bool {
        return child.Get("price").Float() > 10
    })
})

// Use function in path queries
result := root.Query("/items[@filterFunc]/name")

// Call function directly
result := root.CallFunc("filterFunc")

// Use Apply to apply function immediately
result := root.Apply(func(n xjson.Node) bool {
    return n.Get("active").Bool()
})

// Remove function
root.RemoveFunc("filterFunc")
```

## 🛠️ Complete API Reference

### Function Management

| Method | Description | Example |
|--------|-------------|---------|
| **RegisterFunc(name, fn)** | Register path function | `root.RegisterFunc("cheap", filterCheap)` |
| **CallFunc(name)** | Call function directly | `root.CallFunc("cheap")` |
| **RemoveFunc(name)** | Remove function | `root.RemoveFunc("cheap")` |
| **Apply(fn)** | Apply function immediately | `root.Apply(predicateFunc)` |
| **Error() error** | Return first error in chained calls | `if err := n.Error(); err != nil { ... }` |

### Streaming Operations

| Method | Description | Example |
|--------|-------------|---------|
| **Filter(fn)** | Filter node collection | `n.Filter(func(n Node) bool { return n.Get("active").Bool() })` |
| **Map(fn)** | Transform node collection | `n.Map(func(n Node) interface{} { return n.Get("name").String() })` |
| **ForEach(fn)** | Iterate over node collection | `n.ForEach(func(i interface{}, v Node) { fmt.Println(v.String()) })` |

### Native Value Access

| Method | Description | Example |
|--------|-------------|---------|
| **RawFloat()** | Directly get float64 value | `if price, ok := n.RawFloat(); ok { ... }` |
| **RawString()** | Directly get string value | `if name, ok := n.RawString(); ok { ... }` |
| **Strings()** | Get string array | `tags := n.Strings()` |
| **Contains(value)** | Check if contains string | `if n.Contains("target") { ... }` |

## ⚡ Performance Optimization

* **Function caching**: Compiled paths are cached to accelerate repeated queries
* **Native value access**: `Raw` series methods directly access data from underlying memory, avoiding creation of intermediate Node objects
* **Short-circuit optimization**: Supports early termination in certain filtering and query scenarios
* **Efficient chained operations**: Each operation is highly optimized to reduce data copying and memory allocation

**High-performance function example:**

```go
root.RegisterFunc("fastFilter", func(n xjson.Node) xjson.Node {
    return n.Filter(func(child xjson.Node) bool {
        // Directly get native float64 value, no Node overhead
        if price, ok := child.Get("price").RawFloat(); ok {
            return price < 20
        }
        return false
    })
})
```

## 📚 Use Cases

### Business Rule Encapsulation

```go
// Register inventory check function
root.RegisterFunc("inStock", func(n xjson.Node) xjson.Node {
    return n.Filter(func(p xjson.Node) bool {
        return p.Get("stock").Int() > 0 &&
               p.Get("status").String() == "active"
    })
})

// Use semantic queries
availableProducts := root.Query("/products[@inStock]")
```

### Data Transformation Pipeline

```go
import "strings"
import "math"

// Create data cleaning pipeline
root.RegisterFunc("sanitize", func(n xjson.Node) xjson.Node {
    return n.Map(func(item xjson.Node) interface{} {
        return map[string]interface{}{
            "id":    item.Get("id").String(),
            "name":  strings.TrimSpace(item.Get("name").String()),
            "price": math.Round(item.Get("price").Float()*100) / 100,
        }
    })
})

// Apply cleaning pipeline
cleanData := root.Query("/rawInput[@sanitize]")
```

### Complex Data Aggregation

```go
// Calculate average score
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

## 🌟 Design Advantages

* **Concept Simplification**: Only need to understand the single concept of **Node**, gentle learning curve
* **Flexible Composition**: Path functions and streaming operations combine seamlessly, strong expressive power
* **Robust and Reliable**: Chained error handling makes code cleaner and less error-prone
* **Excellent Performance**: Maintains high performance through efficient implementation and native access APIs
* **Type Safety**: Complete type system ensures compile-time type checking
* **Easy to Extend**: Modular design facilitates adding new features

## 🔄 Upgrade Guide

### Upgrading from v0.0.1 to v0.0.2

**Major Changes:**

1. **Function System Update**:
   ```go
   // Old version (deprecated)
   root.Func("name", fn)
   
   // New version (recommended)
   root.RegisterFunc("name", fn)
   ```

2. **New Apply Method**:
   ```go
   // Apply function immediately
   result := root.Apply(func(n xjson.Node) bool {
       return n.Get("active").Bool()
   })
   ```

3. **Enhanced Type System**:
   ```go
   // Use specific function types
   var filterFunc xjson.PredicateFunc = func(n xjson.Node) bool {
       return n.Get("price").Float() > 10
   }
   
   var transformFunc xjson.TransformFunc = func(n xjson.Node) interface{} {
       return n.Get("name").String()
   }
   ```

4. **Wildcard Support**:
   ```go
   // New wildcard queries
   result := root.Query("/store/*/title")
   ```

**Compatibility Notes:**
- The old `Func()` method is still available but marked as deprecated
- All existing query syntax remains valid
- New features are fully backward compatible

## 📄 License

MIT License
