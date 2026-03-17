# XJSON - Unified Node Model JSON Processor (v0.4.0)

**XJSON** **is a powerful Go JSON processing library that uses a fully unified** **Node** **model, supporting path functions, streaming operations, and flexible query syntax.**

## ✨ Core Features

* **🎯** **Single Node Type**: All operations are based on **xjson.Node**, with no **Result** **type.**
* **🧠** **Prepared Queries**: Compile hot query paths once with **CompileQuery** or **MustCompileQuery** and reuse them with minimal overhead.
* **🧩** **Path Functions**: Inject custom logic into queries using **/path[@func]/subpath** **syntax.**
* **🔗** **Chained Operations**: Support fluent function registration, querying, and data operations.
* **🌀** **Robust Error Handling**: Check for errors at the end of chained calls with **node.Error()**.
* **⚡️** **Performance-Oriented**: Achieve zero-copy level performance through efficient chained operations and native value access.
* **🌟** **Wildcard Queries**: Support **`*`** wildcards and complex path expressions.
* **🔍** **Recursive Descent**: Search for matching keys throughout the JSON tree using **//key** **syntax.**
* **⬆️** **Parent Path Navigation**: Access parent nodes flexibly with **../** **syntax for relative path navigation.**

## Current Behavior

- `Parse` keeps the tree lazy and parses child nodes on demand.
- `MustParse` eagerly expands the full tree and is useful when you want upfront validation or repeated full-tree access.
- `CompileQuery` and `MustCompileQuery` build reusable prepared-query handles for hot loops and repeated deep-path access.
- The path parser currently covers quoted special keys, empty keys such as `['']`, escaped quotes and backslashes, negative indexes, slices, recursive descent, and repeated parent navigation like `../../meta`.
- `Parse` and `MustParse` accept `string` or `[]byte` input.

## Benchmark Snapshot

Latest local benchmark run for the v0.4.0 release preparation:

| Scenario | XJSON | GJSON | JsonIter | encoding/json |
| :--- | :--- | :--- | :--- | :--- |
| Parse | `25930 ns/op` | N/A | `22173 ns/op` | `54504 ns/op` |
| Query on prepared data | `17.91 ns/op` | `441.9 ns/op` | `80.94 ns/op` | `80.37 ns/op` |
| Prepared query on prepared data | `16.40 ns/op` | N/A | N/A | N/A |
| Parse each time then query | `109626 ns/op` | N/A | `22145 ns/op` | `55180 ns/op` |
| Mutate only on prepared data | `50.30 ns/op` | N/A | `22.01 ns/op` | `22.03 ns/op` |
| Parse, mutate, then serialize | `63089 ns/op` | N/A | `51917 ns/op` | `74629 ns/op` |

Additional XJSON query split on the same machine:

- `BenchmarkXJSONQuery`: `17.91 ns/op`, `0 B/op`, `0 allocs/op` with repeated identical paths after the root query-result cache is warm.
- `BenchmarkXJSONPreparedQuery`: `16.40 ns/op`, `0 B/op`, `0 allocs/op` for the same path through a compiled prepared query.
- `BenchmarkXJSONQuery_OnceParse_FirstHit`: `124.2 ns/op`, `0 B/op`, `0 allocs/op` with the root query-result cache cleared before each iteration.
- `BenchmarkXJSONPreparedQuery_OnceParse_FirstHit`: `108.6 ns/op`, `0 B/op`, `0 allocs/op` for the prepared-query variant under the same root-cache-reset condition.

Unit test coverage snapshot from the same revision:

- Overall repository statement coverage: `87.5%`.
- Query hot path coverage highlights: `applySimpleQuery` `85.7%`, `fastScanObjectChildLocked` `87.2%`, `tryFastBracketQuery` `89.5%`.
- Query parser coverage highlights: `Parse` `81.0%`, `parseBracketExpression` `92.0%`, `parseQuotedKey` `100.0%`.

Memory snapshot from the same run:

| Benchmark | Memory |
| :--- | :--- |
| `BenchmarkXJSONParse` | `83040 B/op`, `424 allocs/op` |
| `BenchmarkXJSONQuery` | `0 B/op`, `0 allocs/op` |
| `BenchmarkXJSONPreparedQuery` | `0 B/op`, `0 allocs/op` |
| `BenchmarkXJSONSet_Prepared_MutateOnly` | `0 B/op`, `0 allocs/op` |
| `BenchmarkXJSONSet` | `473172 B/op`, `547 allocs/op` |
| `BenchmarkGJSONQuery` | `16 B/op`, `1 allocs/op` |
| `BenchmarkJsonIterParse` | `26602 B/op`, `567 allocs/op` |
| `BenchmarkStandardJSONParse` | `24960 B/op`, `446 allocs/op` |

Notes:

- Environment: `linux/amd64`, `AMD Ryzen 7 7700 8-Core Processor`.
- Command: `go test -run '^$' -bench 'Benchmark(XJSON|GJSON|JsonIter|StandardJSON)(Parse|Decode|Query|Set(_Prepared_MutateOnly)?|Query_OnceParse_(FirstHit|MultiQuery)|Query_LazyParse_EachQuery|PreparedQuery(_OnceParse_FirstHit)?)$' -benchmem ./...`
- Coverage command: `go test ./... -coverprofile=coverage.out && go tool cover -func=coverage.out`.
- All query benchmarks now target the same deep field: `...users[0].profile.personal.name`.
- `BenchmarkXJSONQuery` and `BenchmarkXJSONQuery_OnceParse_MultiQuery` reuse the same parsed root and identical query path, so the XJSON number reflects a root query-result cache hit after the first lookup.
- `BenchmarkXJSONPreparedQuery` removes per-call path-string dispatch and reuses a compiled query handle.
- `BenchmarkXJSONQuery_OnceParse_FirstHit` and `BenchmarkXJSONPreparedQuery_OnceParse_FirstHit` clear only the root query-result cache between iterations; descendant node state remains prepared.
- All mutation benchmarks now target the same deep object: `...users[0].profile.personal.age`, then serialize the whole document.
- `gjson` is query-only, so parse and mutation rows are marked `N/A`.
- `BenchmarkXJSONSet_Prepared_MutateOnly`, `BenchmarkJsonIterSet_Prepared_MutateOnly`, and `BenchmarkStandardJSONSet_Prepared_MutateOnly` isolate write-path cost on already prepared data.
- `BenchmarkXJSONSet`, `BenchmarkJsonIterSet`, and `BenchmarkStandardJSONSet` still include parse plus mutation plus serialization. They are useful as end-to-end write-path comparisons, but they are not isolated mutation-only costs.

## 🚀 Quick Start

XJSON provides both simple and advanced usage patterns. Here are examples for both levels:

### Basic Usage

XJSON's main purpose is to make JSON path querying easy and intuitive. Here are various examples of path syntax usage:

```
package main

import (
	"fmt"
	"github.com/474420502/xjson"
)

func main() {
	// Complex JSON data to demonstrate path querying
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

	// Parse JSON with lazy parsing (nodes parsed on demand)
	root, err := xjson.Parse(data)
	if err != nil {
		panic(err)
	}

	// 1. Basic key access
	store := root.Query("/store")
	fmt.Println("Store exists:", store.IsValid())

	// 2. Nested key access
	bookTitle := root.Query("/store/books[0]/title").String()
	fmt.Println("First book title:", bookTitle)

	// 3. Array indexing
	firstAuthor := root.Query("/store/books[0]/author/first_name").String()
	fmt.Println("First author's first name:", firstAuthor)

	// 4. Array slicing
	bookTitles := root.Query("/store/books[:]/title").Strings()
	fmt.Println("All book titles:", bookTitles)

	// 5. Accessing keys with special characters using quoted syntax
	userName := root.Query("/store/['special.keys']/['user.profile']/name").String()
	fmt.Println("User name with special keys:", userName)

	// 6. Accessing keys with dots in them
	userTheme := root.Query(`/store/['special.keys']/['user.profile']/settings/theme`).String()
	fmt.Println("User theme:", userTheme)

	// 7. Wildcard usage
	allFirstNames := root.Query("/store/books/*/author/first_name").Strings()
	fmt.Println("All author first names:", allFirstNames)

	// 8. Accessing array elements by condition (first element)
	firstRating := root.Query("/ratings[0]/score").Float()
	fmt.Printf("First rating score: %.1f\n", firstRating)

	fmt.Println("\n--- More Path Examples ---")

	// 9. Complex nested access
	cpuSpec := root.Query("/store/electronics/computers[0]/specifications/cpu").String()
	fmt.Println("CPU specification:", cpuSpec)

	// 10. Accessing boolean values
	inStock := root.Query("/store/electronics/computers[0]/in_stock").Bool()
	fmt.Println("Computer in stock:", inStock)

	// 11. Accessing array elements
	firstTag := root.Query("/store/books[0]/tags[0]").String()
	fmt.Println("First tag of first book:", firstTag)

	// 12. Accessing numeric values
	bookPrice := root.Query("/store/books[1]/price").Float()
	fmt.Printf("Second book price: $%.2f\n", bookPrice)
}
```

For different types of path operations:

```
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

	// Array index access
	firstUserId := root.Query("/users[0]/id").Int()
	fmt.Println("First user ID:", firstUserId)

	// Array slice access
	userNames := root.Query("/users[:]/name").Strings()
	fmt.Println("User names:", userNames)

	// Nested object access
	firstUserAge := root.Query("/users[0]/profile/age").Int()
	fmt.Println("First user age:", firstUserAge)

	// Array of objects property access
	allTags := root.Query("/users[*]/profile/tags").Strings()
	fmt.Println("All user tags:", allTags)

	// Nested array access
	firstUserFirstScore := root.Query("/users[0]/scores[0]").Int()
	fmt.Println("First user's first score:", firstUserFirstScore)

	// Boolean value access
	firstUserActive := root.Query("/users[0]/profile/active").Bool()
	fmt.Println("First user active:", firstUserActive)

	// Accessing metadata
	version := root.Query("/metadata/version").String()
	fmt.Println("Version:", version)
}
```

For working with special key names:

```
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

	// Accessing keys with dots
	firstName := root.Query(`/['user-data']/['user.profile']/['first.name']`).String()
	fmt.Println("First name:", firstName)

	// Accessing keys with slashes
	apiPath := root.Query(`/['api/v1/users']`).Len()
	fmt.Println("API users count:", apiPath)

	// Mixed regular and special keys
	userName := root.Query(`/['api/v1/users'][0]/['profile.data']/name`).String()
	fmt.Println("User name:", userName)

	// Deep access with special keys
	email := root.Query(`/['api/v1/users'][0]/['profile.data']/['contact-info']/['email.address']`).String()
	fmt.Println("Email:", email)

	// Accessing nested special keys
	theme := root.Query(`/['user-data']/['user.settings']/['ui.theme']`).String()
	fmt.Println("Theme:", theme)
}
```

For array operations:

```
func arrayExample() {
	data := `{
		"users": [
			{"name": "Alice", "age": 25},
			{"name": "Bob", "age": 30},
			{"name": "Charlie", "age": 35}
		]
	}`

	root, _ := xjson.Parse(data)

	// Get array length
	count := root.Get("users").Len()
	fmt.Printf("Total users: %d\n", count)

	// Access by index
	firstUser := root.Get("users").Index(0).Get("name").String()
	fmt.Printf("First user: %s\n", firstUser)

	// Iterate through array
	root.Get("users").ForEach(func(index interface{}, user xjson.Node) {
		name := user.Get("name").String()
		age := user.Get("age").Int()
		fmt.Printf("User %d: %s (age %d)\n", index, name, age)
	})
}
```

### Prepared Queries

When the same deep path is evaluated repeatedly, compile it once and reuse it:

```go
prepared := xjson.MustCompileQuery("/store/books[0]/title")

root, err := xjson.MustParse(data)
if err != nil {
	panic(err)
}

for i := 0; i < 1000; i++ {
	title := prepared.Query(root).String()
	_ = title
}
```

### Advanced Usage

For complex data processing with functions:

```
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

	// Parse JSON with full eager parsing
	root, err := xjson.MustParse(data)
	if err != nil {
		panic(err)
	}

	// Register custom functions
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

	// Complex queries using path functions
	cheapBooks := root.Query("/store/books[@cheap]/title").Strings()
	fmt.Println("Cheap books:", cheapBooks)

	// Find all in-stock electronics
	inStockItems := root.Query("/store/electronics[@inStock]/name").Strings()
	fmt.Println("In-stock items:", inStockItems)

	// Find programming books
	progBooks := root.Query("/store/books[@programming]/title").Strings()
	fmt.Println("Programming books:", progBooks)

	// Use recursive descent to find all prices
	allPrices := root.Query("//price").Map(func(n xjson.Node) interface{} {
		price, _ := n.RawFloat()
		return price
	})

	// Calculate average price
	var sum float64
	var count int
	allPrices.ForEach(func(_ interface{}, priceNode xjson.Node) {
		if price, ok := priceNode.Interface().(float64); ok {
			sum += price
			count++
		}
	})
	avgPrice := sum / float64(count)
	fmt.Printf("Average price: %.2f\n", avgPrice)

	// Use parent navigation
	firstBookTitle := root.Query("/store/books[0]/../books[0]/title").String()
	fmt.Println("First book (using parent navigation):", firstBookTitle)
}
```

For data modification:

```
func modificationExample() {
	data := `{
		"users": [
			{"id": 1, "name": "John", "active": true},
			{"id": 2, "name": "Jane", "active": false}
		]
	}`

	root, _ := xjson.Parse(data)

	// Modify existing data
	root.Query("/users[0]").Set("name", "John Doe")
	
	// Add new data
	newUser := map[string]interface{}{
		"id": 3,
		"name": "Bob",
		"active": true,
	}
	root.Query("/users").Append(newUser)

	// Use SetValue to replace entire node value
	root.Query("/users[1]/active").SetValue(true)

	fmt.Println("Modified data:", root.String())
}
```

## Lazy Iterators (ObjectIter / ArrayIter)

When working with very large JSON documents, iterating over keys or array elements without forcing full parsing of every child can save CPU and memory. XJSON's engine exposes lazy iterators (`ObjectIter` and `ArrayIter`) that scan the underlying bytes and only parse a value when you explicitly request it.

Important notes:

- Iterators operate in two modes:
  - raw-mode: scans the original `raw []byte` and returns key/value byte ranges without allocating child nodes.
  - parsed-mode: when the node has been modified (`isDirty==true`) or has no raw bytes, iterators traverse the in-memory `value` structures.
- Call `ParseValue()` to parse the current element's value into a `core.Node` only when needed. Parsed children may be cached back on the parent to speed subsequent access.
- Iterators are not safe for concurrent mutation. Avoid modifying the node while iterating.

Example (internal/engine usage):

```go
// Assume obj is an *objectNode or a Node known to be an object in the engine package
it := obj.Iter() // returns ObjectIter
for it.Next() {
	key := string(it.KeyRaw()) // cheap string conversion of raw key
	if shouldParse(key) {
		child := it.ParseValue() // parse on demand
		// use child (core.Node)
	} else {
		rawVal := it.ValueRaw() // raw bytes for the value
		// inspect rawVal without allocating a Node
	}
}
if err := it.Err(); err != nil {
	// handle iterator error
}
```


## 💡 Core Design

### 1. Unified Node Model

**All JSON elements (objects, arrays, strings, numbers, etc.), including query result sets, are represented by the** **Node** **interface.**

```go
type Node interface {
    // Basic Access
    Type() NodeType
    IsValid() bool
    Error() error
    Path() string
    Raw() string
    Parent() Node
  
    // Query Methods
    Query(path string) Node
    Get(key string) Node
    Index(i int) Node
  
    // Streaming Operations
    Filter(fn PredicateFunc) Node
    Map(fn TransformFunc) Node
    ForEach(fn func(keyOrIndex interface{}, value Node)) 
    Len() int
  
    // Write Operations
    Set(key string, value interface{}) Node
    Append(value interface{}) Node
    SetValue(value interface{}) Node
  
    // Function Support
    RegisterFunc(name string, fn UnaryPathFunc) Node
    CallFunc(name string) Node
    RemoveFunc(name string) Node
    Apply(fn PathFunc) Node
    GetFuncs() *map[string]UnaryPathFunc
  
    // Type Conversion
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
  
    // Native Value Access (Performance Optimization)
    RawFloat() (float64, bool)
    RawString() (string, bool)
  
    // Other Conversion Methods
    Strings() []string
    Keys() []string
    Contains(value string) bool
    AsMap() map[string]Node
    MustAsMap() map[string]Node
}
```

### 2. Function Type System

**XJSON provides multiple function types to support different operation scenarios:**

```go
// Path Function - Generic function container
type PathFunc interface{}

// Unary Path Function - Node to node transformation
type UnaryPathFunc func(node Node) Node

// Predicate Function - Used for filtering operations
type PredicateFunc func(node Node) bool

// Transform Function - Used for mapping operations
type TransformFunc func(node Node) interface{}
```

### 3. Error Handling

**XJSON uses chain-friendly error handling mode:**

```go
// No need to check err at every step
value := root.Query("/path/that/does/not/exist").Get("key").Int()

// Check at the end
if err := root.Error(); err != nil {
    fmt.Println("Operation chain failed:", err)
}
```

### 4. Parsing Methods

**XJSON provides two parsing methods with different behaviors:**

#### Lazy Parsing with Parse()

The [Parse()](file:///home/eson/workspace/xjson/xjson.go#L45-L65) function creates a lazy-parsed tree where nodes are parsed on-demand when accessed:

```go
// Nodes are not immediately parsed - they will be parsed when accessed
root, err := xjson.Parse(data)
if err != nil {
    panic(err)
}

// Only when accessing data, the relevant nodes are parsed
title := root.Query("/store/books[0]/title").String()
```

This approach is more memory-efficient for large JSON documents when you only need to access parts of the data.

#### Eager Parsing with MustParse()

The `MustParse()` function parses the entire JSON tree immediately:

```go
// All nodes are parsed immediately
root, err := xjson.MustParse(data)
if err != nil {
    panic(err)
}

// No additional parsing needed when accessing data
title := root.Query("/store/books[0]/title").String()
```

This approach is useful when you know you'll be accessing most of the data in the JSON document, or when you want to validate the entire document upfront.

### 5. Path Query Syntax

XJSON provides a powerful and flexible path query syntax that supports various data access patterns from simple to complex.

#### **Basic Syntax**

**4.1. Root Node**

Path queries always start with `/`, representing the root node.

* **Syntax**: `/`
* **Description**: Represents the root node of the JSON data.
* **Example**: `/store` gets the `store` key from the root node.

**Note**: `/store/books` and `store/books` are equivalent.

**4.2. Key Access**

Standard object field access is done directly by key name. Any string that conforms to Go language identifier conventions can be used directly as a path segment.

* **Syntax**: `/key1/key2`
* **Example**: `/store/books`, this path will sequentially get the `store` key and `books` key.

**4.3. Array Access**

Access array elements through square brackets `[...]`, supporting single index and range slicing.

* **Index Access**:

  * **Syntax**: `[<index>]`
  * **Description**: Get a single array element, index starts from 0.
  * **Example**: `/store/books[0]`, get the first element of the `books` array.
* **Slice Access**:

  * **Syntax**:
    * `[start:end]`: Get elements from `start` to `end-1`.
    * `[start:]`: Get elements from `start` to the end.
    * `[:end]`: Get elements from the beginning to `end-1`.
    * `[-N:]`: Get the last N elements.
  * **Description**: Get a subset of the array and return a new array node containing these elements.
  * **Example**: `/store/books[1:3]`, return a new array containing the second and third elements of the `books` array.

**4.4. Function Calls**

Call registered functions in the path using the `[@<funcName>]` syntax. Functions provide a powerful mechanism for data processing and filtering.

* **Syntax**: `[@<Function Name>]`
* **Identifier**: The `@` symbol clearly indicates this is a function call.
* **Requirement**: The function must be registered to the node via `RegisterFunc`.
* **Example**: `/store/books[@cheap]/title`, call the `cheap` function on the `books` array and extract `title` from the result.

**4.5. Wildcards**

The asterisk `*` acts as a wildcard to match all direct child elements of a node.

* **Syntax**: `*`
* **Behavior on Objects**: Match all values of the object and return a new array node containing these values.
* **Behavior on Arrays**: Match all elements of the array and return the array itself.
* **Example**: `/store/*/title`, get the `title` field of all direct child nodes under the `store` object (here it's the `books` array).

#### **Advanced Syntax**

**5.1. Chained and Mixed Syntax**

All core components can be freely combined to form powerful chained queries. The parser executes each operation from left to right.

* **Example**: `/store/books[@filter][0]/name`
  1. `/store/books`: Get the `books` array.
  2. `[@filter]`: Call the `filter` function on the array.
  3. `[0]`: Get the first element of the function return result (should be an array).
  4. `/name`: Get the `name` field of that element.

**5.2. Special Character Key Name Handling**

When object key names contain special characters such as `/`, `.`, `[`, `]` or non-alphanumeric characters, they must be delimited using square brackets and quotes `['<key>']` or `["<key>"]`.

* **Syntax**: `['<Key Name>']` or `["<Key Name>"]`
* **Key with Slash**: `/['/api/v1/users']`
* **Key with Dot**: `/data/['user.profile']/name`
* **Key with Quotes**:
  * If the key name is `a"key`, use `['a"key']`.
  * If the key name is `a'key`, use `["a'key"]`.
* **Empty Key**: `/['']/name`
* **Escaping**:
	* In single-quoted keys, escape `'` as `\'` and `\` as `\\`.
	* In double-quoted keys, escape `"` as `\"` and `\` as `\\`.
* **Mixed with Regular Paths**: `/data['user-settings']/theme`

**5.3. Recursive Descent**

Double slashes `//` are used to perform deep searches in the current node and all its descendants to find matching keys.

* **Syntax**: `//key`
* **Description**: Unlike `/` which only searches in direct children, `//` traverses the entire subtree and collects all nodes matching `key` into a new array node.
* **Example**: `//author` will search for all `author` fields at all levels starting from the root node.

**More Usage Examples:**

```go
// Find all price fields
allPrices := root.Query("//price").Strings()

// Find all books with tags
taggedBooks := root.Query("//books").Filter(func(n xjson.Node) bool {
    return n.Get("tags").Len() > 0
})

// Find all items in stock
inStockItems := root.Query("//in_stock").Filter(func(n xjson.Node) bool {
    return n.Bool() == true
})

// Combine with functions to find all cheap items
cheapItems := root.Query("//price[@cheap]")
```

**Best Practices:**

1. **Limit Search Scope**: First locate to the approximate area using precise paths, then use recursive descent

   ```go
   // Recommended: First locate to store, then search
   storePrices := root.Query("/store//price")

   // Avoid global search
   allPrices := root.Query("//price")
   ```
2. **Combine with Filter Functions**: Use the `Filter()` method to further filter results

   ```go
   // Find all prices and filter out the cheap ones
   cheapPrices := root.Query("//price").Filter(func(n xjson.Node) bool {
       price, _ := n.RawFloat()
       return price < 20
   })
   ```

3. **Use Caution**: Prioritize precise paths when the structure is known

> **Performance Warning**: Recursive descent `//` is a very powerful but costly operation. Because it needs to traverse the entire subtree of a node, it can become a performance bottleneck when processing large or deeply nested JSON data. It is recommended to use precise paths in performance-sensitive scenarios, and only use recursive descent when the data structure is uncertain or global search is truly needed.

**5.4. Parent Path Lookup**

The double dot `../` syntax is used to access the parent node of the current node, implementing relative path navigation.

* **Syntax**: `../key` or `../`
* **Description**: Allows navigation from the current node to the parent node, then continue querying downward. This is particularly useful when dealing with complex nested structures, allowing flexible data access without knowing the complete path.
* **Example**: `/store/books[0]/../electronics` navigates from the first book to the `store` node, then accesses `electronics`.

**Usage Examples:**

```go
// Navigate from book node to parent store, then get electronics
electronicsFromBook := root.Query("/store/books[0]/../electronics/laptops").Strings()

// Get all book parent category names
bookCategories := root.Query("/store/books[0]/../").Keys()

// Reference sibling fields in array elements
firstBookTitle := root.Query("/store/books[0]/title").String()
firstBookPrice := root.Query("/store/books[0]/../books[0]/price").Float()

// Multi-level parent navigation
rootFromDeep := root.Query("/store/electronics/laptops[0]/../../authors").Strings()
```

**Real-World Application Scenarios:**

1. **Related Data Query**: Find related data in nested structures

   ```go
   // Find categories of all in-stock items
   inStockCategories := root.Query("/store/*/laptops").Filter(func(n xjson.Node) bool {
       return n.Get("in_stock").Bool() == true
   }).Query("../..").Keys()
   ```
2. **Data Validation**: Check relationships between fields

   ```go
   // Validate if price is within reasonable range
   validatePrice := root.Query("/store/books").Filter(func(n xjson.Node) bool {
       price := n.Get("price").Float()
       category := n.Query("../").String() // Get parent information
       // Validate price based on category
       return isValidPriceForCategory(price, category)
   })
   ```

3. **Dynamic Path Construction**: Navigate when the specific structure is uncertain

   ```go
   // Find specific fields from any node
   findStoreInfo := root.Query("//price/../..") // Find the corresponding store from price
   ```

**Limitations and Considerations:**

1. **Root Node Limitation**: Using `../` on the root node will return an invalid node
2. **Performance Considerations**: Too much parent navigation may affect code readability, it's recommended to use precise paths when the structure is known
3. **Chained Usage**: Multiple `../` can be used consecutively for multi-level parent navigation

#### **Syntax Quick Reference**

| Category | Syntax | Description | Example |
| :--- | :--- | :--- | :--- |
| **Basic** | `/` | Separator between path segments. | `/store/books` |
| | `key` | Access object fields. | `/store` |
| **Array** | `[<index>]` | Access array elements by index. | `[0]`, `[-1]` |
| | `[start:end]` | Access array elements by range (slicing). | `[1:3]`, `[:-1]` |
| **Function** | `[@<name>]` | Call registered path functions. | `[@cheap]`, `[@inStock]` |
| **Advanced** | `*` | Match all direct child elements of object or array. | `/store/*` |
| | `//key` | Recursively search for `key` in all descendant nodes (high performance cost). | `//author` |
| | `../key` | Access parent node, then continue querying downward. | `/books[0]/../electronics` |
| **Special Characters** | `['<key>']` | Delimit key names containing special characters. | `['user.profile']` |
| | `["<key>"]` | Delimit key names containing single quotes. | `["a'key"]` |

Notes:

- Consecutive parent segments are supported, for example `/store/books[0]/../../meta`.
- Invalid path syntax returns an invalid node with an attached error; check `node.Error()` when you need to distinguish “not found” from “bad path”.

### 6. Function Registration and Calling

**The new function system is more powerful and flexible:**

```go
// Register function (recommended)
root.RegisterFunc("filterFunc", func(n xjson.Node) xjson.Node {
    return n.Filter(func(child xjson.Node) bool {
        return child.Get("price").Float() > 10
    })
})

// Use function in path query
result := root.Query("/items[@filterFunc]/name")

// Call function directly
result := root.CallFunc("filterFunc")

// Apply function immediately
result := root.Apply(func(n xjson.Node) bool {
    return n.Get("active").Bool()
})

// Remove function
root.RemoveFunc("filterFunc")

// Get registered functions
funcs := root.GetFuncs()
```

## 🛠️ Complete API Reference

### Top-Level Helpers

| Helper | Description | Example |
| --- | --- | --- |
| **Parse(data)** | Parse lazily from `string` or `[]byte` | `root, err := xjson.Parse(data)` |
| **MustParse(data)** | Parse eagerly from `string` or `[]byte` | `root, err := xjson.MustParse(data)` |
| **CompileQuery(path)** | Compile a reusable prepared query | `pq, err := xjson.CompileQuery("/users[0]/name")` |
| **MustCompileQuery(path)** | Compile a prepared query and panic on invalid syntax | `pq := xjson.MustCompileQuery("/users[0]/name")` |

### Prepared Queries

| Type / Method | Description | Example |
| --- | --- | --- |
| **PreparedQuery.Query(node)** | Execute the compiled query against a node | `pq.Query(root).String()` |
| **PreparedQuery.Path()** | Return the original compiled path | `pq.Path()` |

### Navigation and Mutation

| Method | Description | Example |
| --- | --- | --- |
| **Query(path)** | Evaluate an absolute or relative query path | `root.Query("/store/books[0]/title")` |
| **Get(key)** | Access an object field directly | `root.Get("store")` |
| **Index(i)** | Access an array element directly | `root.Get("books").Index(0)` |
| **Set(key, value)** | Set or replace an object field | `root.Query("/user").Set("name", "Alice")` |
| **Append(value)** | Append to an array | `root.Query("/users").Append(newUser)` |
| **SetValue(value)** | Replace the current node in-place | `root.Query("/users[1]/active").SetValue(true)` |
| **SetByPath(path, value)** | Set a value by path, creating intermediates when possible | `root.SetByPath("/config/theme", "dark")` |
| **Path()** | Return the canonical path of the current node | `root.Query("/users[0]/name").Path()` |
| **Parent()** | Return the parent node | `root.Query("/users[0]").Parent()` |

### Function Management

| Method | Description | Example |
| --- | --- | --- |
| **RegisterFunc(name, fn)** | Register path function | `root.RegisterFunc("cheap", filterCheap)` |
| **CallFunc(name)** | Call function directly | `root.CallFunc("cheap")` |
| **RemoveFunc(name)** | Remove function | `root.RemoveFunc("cheap")` |
| **Apply(fn)** | Apply a `UnaryPathFunc`, `PredicateFunc`, or `TransformFunc` immediately | `root.Apply(predicateFunc)` |
| **GetFuncs()** | Get registered functions | `funcs := root.GetFuncs()` |
| **Error() error** | Return the first error in chained calls | `if err := n.Error(); err != nil { ... }` |

### Streaming Operations

| Method | Description | Example |
| --- | --- | --- |
| **Filter(fn)** | Filter node collection | `n.Filter(func(n Node) bool { return n.Get("active").Bool() })` |
| **Map(fn)** | Transform node collection | `n.Map(func(n Node) interface{} { return n.Get("name").String() })` |
| **ForEach(fn)** | Iterate through node collection | `n.ForEach(func(i interface{}, v Node) { fmt.Println(v.String()) })` |

### Native Value Access

| Method | Description | Example |
| --- | --- | --- |
| **RawFloat()** | Directly get float64 value | `if price, ok := n.RawFloat(); ok { ... }` |
| **RawString()** | Directly get string value | `if name, ok := n.RawString(); ok { ... }` |
| **Strings()** | Get string array | `tags := n.Strings()` |
| **Contains(value)** | Check if string is contained | `if n.Contains("target") { ... }` |
| **AsMap()** | Get node as map | `obj := n.AsMap()` |
| **Keys()** | Get all keys of object | `keys := n.Keys()` |

### Forced Type Conversion

| Method | Description | Example |
| --- | --- | --- |
| **MustString()** | Get string value, panic on failure | `value := n.MustString()` |
| **MustFloat()** | Get float64 value, panic on failure | `value := n.MustFloat()` |
| **MustInt()** | Get int64 value, panic on failure | `value := n.MustInt()` |
| **MustBool()** | Get bool value, panic on failure | `value := n.MustBool()` |
| **MustTime()** | Get time.Time value, panic on failure | `value := n.MustTime()` |
| **MustArray()** | Get array value, panic on failure | `value := n.MustArray()` |
| **MustAsMap()** | Get map value, panic on failure | `value := n.MustAsMap()` |

## ⚡ Performance Optimization

* **Lazy Child Caching**: Parsed child nodes are cached back onto parents when safe, reducing repeated parsing work on hot paths.
* **Native Value Access**: `Raw` series methods directly access data from underlying memory, avoiding creation of intermediate **Node** objects.
* **Short-Circuit Optimization**: Support early termination in some filtering and query scenarios.
* **Efficient Chained Operations**: Each operation is highly optimized to reduce data copying and memory allocation.

Notes:

- Root query-result caching and compiled fast-query plans are enabled on hot paths.
- For repeated deep-path access in tight loops, prefer `CompileQuery` or `MustCompileQuery` over repeatedly reparsing the same path string.
- The internal lazy iterators described above are engine-level optimizations, not a stable public API.

**High-Performance Function Example:**

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

## 📚 Usage Scenarios

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

* **Concept Simplification**: Only need to understand the **Node** **concept, shallow learning curve.**
* **Flexible Combination**: Path functions seamlessly combine with streaming operations, strong expressive power.
* **Robust and Reliable**: Chained error handling mechanism makes code more concise and less error-prone.
* **Excellent Performance**: Maintain high performance through efficient implementation and native access APIs.
* **Type Safety**: Complete type system ensures compile-time type checking.
* **Easy to Extend**: Modular design facilitates adding new features.

## 🔄 Upgrade Guide

### Upgrading to v0.4.0

**Highlights:**

1. **Prepared Query API**:

	```go
	pq := xjson.MustCompileQuery("/users[0]/profile/name")
	name := pq.Query(root).String()
	```

2. **More Aggressive Query Fast Paths**:

	```go
	// Repeated deep-path reads now benefit from root query-result caching
	// and a compiled fast-plan executor.
	value := root.Query("/users[0]/profile/name").String()
	```

3. **Parse Semantics Are Explicit**:

	```go
	lazyRoot, _ := xjson.Parse(data)
	eagerRoot, _ := xjson.MustParse(data)
	```

4. **No Query Syntax Breakage**:

	Existing path syntax remains compatible, including quoted keys, negative indexes,
	slices, recursive descent, parent navigation, and path functions.

**Compatibility Notes:**

- Existing query code continues to work unchanged.
- `CompileQuery` and `MustCompileQuery` are additive APIs for hot paths; they do not replace `Node.Query`.
- Release benchmarks were refreshed to reflect the current prepared-query and root-cache behavior.

### Upgrading from v0.0.2 to v0.1.0

**Major Changes:**

1. **Enhanced Write Operations**:
   
   ```go
   // New SetValue method for direct value setting
   node.SetValue("new value")
   
   // Enhanced Set method with better error handling
   result := node.Set("key", "value")
   ```

2. **Additional Type Conversion Methods**:
   
   ```go
   // AsMap for object conversion
   objMap := node.AsMap()
   
   // MustAsMap for forced object conversion
   objMap := node.MustAsMap()
   
   // Keys for getting all object keys
   keys := node.Keys()
   ```

3. **Enhanced Error Handling**:
   
   ```go
   // More detailed error information
   if err := node.Error(); err != nil {
       fmt.Printf("Error at path %s: %v\n", node.Path(), err)
   }
   ```

4. **Performance Improvements**:
   
   ```go
   // Optimized RawString and RawFloat methods
   if str, ok := node.RawString(); ok {
       // Zero-copy string access
   }
   ```

5. **Parsing Method Changes**:
   
   ```go
   // New Parse method for lazy parsing (recommended for most use cases)
   root, err := xjson.Parse(data)
   
   // MustParse method for eager parsing (full immediate parsing)
   root, err := xjson.MustParse(data)
   ```

**Compatibility Notes:**

- All existing query syntax continues to work
- New features are fully backward compatible
- Performance improvements do not affect existing code
- The old `Parse` behavior is now provided by `MustParse`

## 📄 License

MIT License
