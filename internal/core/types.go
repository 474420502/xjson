package core

import (
	"time"
)

// NodeType 定义了 JSON 节点的类型。
// NodeType defines the type of a JSON node.
type NodeType int

const (
	// ObjectNode 代表一个 JSON 对象。
	// ObjectNode represents a JSON object.
	ObjectNode NodeType = iota
	// ArrayNode 代表一个 JSON 数组。
	// ArrayNode represents a JSON array.
	ArrayNode
	// StringNode 代表一个 JSON 字符串。
	// StringNode represents a JSON string.
	StringNode
	// NumberNode 代表一个 JSON 数字。
	// NumberNode represents a JSON number.
	NumberNode
	// BoolNode 代表一个 JSON 布尔值。
	// BoolNode represents a JSON boolean.
	BoolNode
	// NullNode 代表一个 JSON null。
	// NullNode represents a JSON null.
	NullNode
	// InvalidNode 代表一个无效或不存在的节点。
	// InvalidNode represents an invalid or non-existent node.
	InvalidNode
)

// PathFunc 是一个通用的函数容器，用于将不同的函数签名传递给增强的 Func 方法。
// PathFunc is a generic function container for passing different function signatures
// to the enhanced Func method.
type PathFunc interface{}

// UnaryPathFunc 是一个节点到节点的转换函数，保持现有功能。
// 这是可以注册并稍后调用的函数签名。
// UnaryPathFunc is a node-to-node transformation function, maintaining existing functionality.
// This is the function signature that can be registered and called later.
type UnaryPathFunc func(node Node) Node

// PredicateFunc 是一个谓词函数，它接受一个节点并返回一个布尔值。
// 主要用于过滤操作。
// PredicateFunc is a predicate function that takes a node and returns a boolean.
// Mainly used for filtering operations.
type PredicateFunc func(node Node) bool

// TransformFunc 是一个转换函数，它接受一个节点并返回一个任意值。
// 主要用于映射操作。
// TransformFunc is a transformation function that takes a node and returns an arbitrary value.
// Mainly used for mapping operations.
type TransformFunc func(node Node) interface{}

// Accessor 提供了访问节点属性和值的方法。
// Accessor provides methods for accessing node properties and values.
type Accessor interface {
	// Type 返回节点的数据类型。
	// Type returns the data type of the node.
	Type() NodeType

	// IsValid 检查节点是否有效。如果链中的前一个操作失败，节点将变为无效。
	// IsValid checks if the node is valid. A node becomes invalid if a preceding
	// operation in a chain fails.
	IsValid() bool

	// Error 返回在操作链中发生的第一个错误。
	// 如果所有操作都成功，则返回 nil。
	// Error returns the first error that occurred during a chain of operations.
	// It returns nil if all operations were successful.
	Error() error

	// Path 返回从根到当前节点的 JSON 路径。
	// Path returns the JSON path to the current node from the root.
	Path() string

	// Raw 返回节点的原始 JSON 字符串表示形式。
	// Raw returns the raw JSON string representation of the node.
	Raw() string

	// Get 通过键从对象中检索子节点。
	// 如果当前节点不是对象或键不存在，它将返回一个 InvalidNode。
	// Get retrieves a child node from an object by its key.
	// If the current node is not an object or the key does not exist, it returns
	// an InvalidNode.
	Get(key string) Node

	// Index 通过索引从数组中检索子节点。
	// 如果当前节点不是数组或索引越界，它将返回一个 InvalidNode。
	// Index retrieves a child node from an array by its index.
	// If the current node is not an array or the index is out of bounds, it returns
	// an InvalidNode.
	Index(i int) Node

	// Query 执行一个类似 JSONPath 的查询并返回结果节点。
	// 查询语法支持键访问、数组索引和函数调用。
	// 示例: "users.0.name"
	// Query executes a JSONPath-like query and returns the resulting node.
	// The query syntax supports key access, array indexing, and function calls.
	// Example: "users.0.name"
	Query(path string) Node
}

// Iterable 提供了迭代节点集合的方法。
// Iterable provides methods for iterating over node collections.
type Iterable interface {
	// ForEach 迭代数组的元素或对象的键值对。
	// 对于数组，回调接收索引和元素节点。
	// 对于对象，回调接收键（作为字符串）和值节点。
	// ForEach iterates over the elements of an array or the key-value pairs of an object.
	// For arrays, the callback receives the index and the element node.
	// For objects, the callback receives the key (as a string) and the value node.
	ForEach(iterator func(keyOrIndex interface{}, value Node))

	// Len 返回数组或对象中的元素数量。
	// 对于所有其他节点类型，它返回 0。
	// Len returns the number of elements in an array or object.
	// It returns 0 for all other node types.
	Len() int
}

// Converter 提供了将节点值转换为不同类型的方法。
// Converter provides methods for converting node values to different types.
type Converter interface {
	// String 返回节点的字符串值。
	// 对于非字符串类型，它返回值的字符串表示形式。
	// String returns the string value of the node.
	// For non-string types, it returns a string representation of the value.
	String() string

	// MustString 类似于 String，但如果节点不是字符串类型则会 panic。
	// MustString is like String but panics if the node is not a string type.
	MustString() string

	// Float 返回数字节点的 float64 值。
	// 如果节点不是数字，则返回 0。
	// Float returns the float64 value of a number node.
	// Returns 0 if the node is not a number.
	Float() float64

	// MustFloat 类似于 Float，但如果节点不是数字类型则会 panic。
	// MustFloat is like Float but panics if the node is not a number type.
	MustFloat() float64

	// Int 返回数字节点的 int64 值。
	// 如果节点不是数字，则返回 0。
	// Int returns the int64 value of a number node.
	// Returns 0 if the node is not a number.
	Int() int64

	// MustInt 类似于 Int，但如果节点不是数字类型则会 panic。
	// MustInt is like Int but panics if the node is not a number type.
	MustInt() int64

	// Bool 返回布尔节点的布尔值。
	// 如果节点不是布尔值，则返回 false。
	// Bool returns the boolean value of a bool node.
	// Returns false if the node is not a boolean.
	Bool() bool

	// MustBool 类似于 Bool，但如果节点不是布尔类型则会 panic。
	// MustBool is like Bool but panics if the node is not a bool type.
	MustBool() bool

	// Time 返回 time.Time 值，假设字符串是 RFC3339 格式。
	// Time returns the time.Time value, assuming the string is in RFC3339 format.
	Time() time.Time

	// MustTime 类似于 Time，但如果解析失败或节点不是字符串，则会 panic。
	// MustTime is like Time but panics if parsing fails or node is not a string.
	MustTime() time.Time

	// Array 如果节点是数组，则返回节点切片。
	// 对于其他类型，返回 nil。
	// Array returns a slice of nodes if the node is an array.
	// Returns nil for other types.
	Array() []Node

	// MustArray 类似于 Array，但如果节点不是数组则会 panic。
	// MustArray is like Array but panics if the node is not an array.
	MustArray() []Node

	// Interface 返回节点的基础值作为标准的 Go interface{}。
	// 这对于测试或应用程序代码中的比较和类型断言很有用。
	// - 对于 ObjectNode，它返回 map[string]interface{}。
	// - 对于 ArrayNode，它返回 []interface{}。
	// - 对于 StringNode，它返回 string。
	// - 对于 NumberNode，它返回 float64。
	// - 对于 BoolNode，它返回 bool。
	// - 对于 NullNode，它返回 nil。
	// Interface returns the underlying value of the node as a standard Go interface{}.
	// This is useful for comparisons and type assertions in tests or application code.
	// - For ObjectNode, it returns map[string]interface{}.
	// - For ArrayNode, it returns []interface{}.
	// - For StringNode, it returns string.
	// - For NumberNode, it returns float64.
	// - For BoolNode, it returns bool.
	// - For NullNode, it returns nil.
	Interface() interface{}

	// RawFloat 返回数字节点的底层 float64 值，而无需创建新的 Node 对象。
	// 它返回值和表示成功的布尔值。
	// RawFloat returns the underlying float64 value of a number node without creating a new Node object.
	// It returns the value and a boolean indicating success.
	RawFloat() (float64, bool)

	// RawString 返回字符串节点的底层字符串值，而无需创建新的 Node 对象。
	// 它返回值和表示成功的布尔值。
	// RawString returns the underlying string value of a string node without creating a new Node object.
	// It returns the value and a boolean indicating success.
	RawString() (string, bool)

	// Strings 如果节点是字符串数组，则返回字符串切片。
	// 如果节点不是数组或包含非字符串元素，则返回 nil。
	// Strings returns a slice of strings if the node is an array of strings.
	// Returns nil if the node is not an array or contains non-string elements.
	Strings() []string

	// Contains 检查数组节点是否包含特定的字符串值。
	// 如果节点不是数组或不包含该值，则返回 false。
	// Contains checks if an array node contains a specific string value.
	// Returns false if the node is not an array or does not contain the value.
	Contains(value string) bool

	// AsMap 如果节点是对象，则返回 map[string]Node。
	// 对于其他类型，返回 nil。
	// AsMap returns a map[string]Node if the node is an object.
	// Returns nil for other types.
	AsMap() map[string]Node

	// MustAsMap 类似于 AsMap，但如果节点不是对象则会 panic。
	// MustAsMap is like AsMap but panics if the node is not an object.
	MustAsMap() map[string]Node
}

// Mutator 提供了修改节点值的方法。
// Mutator provides methods for modifying node values.
type Mutator interface {
	// Set 在对象节点中为给定键设置值。
	// 如果节点不是对象，它将返回一个 InvalidNode。
	// Set sets the value for a given key in an object node.
	// If the node is not an object, it returns an InvalidNode.
	Set(key string, value interface{}) Node

	// Append 将新值添加到数组节点。
	// 如果节点不是数组，它将返回一个 InvalidNode。
	// Append adds a new value to an array node.
	// If the node is not an array, it returns an InvalidNode.
	Append(value interface{}) Node
}

// Functional 提供了用于节点操作的高阶函数方法。
// Functional provides higher-order functional methods for node manipulation.
type Functional interface {
	// Deprecated: 使用 RegisterFunc 和 CallFunc 代替
	// Deprecated: Use RegisterFunc and CallFunc instead
	Func(name string, fn func(Node) Node) Node
	// Filter 对集合（数组或对象值）中的每个元素应用一个函数
	// 并返回一个仅包含函数返回 true 的元素的新 Node。
	// Filter applies a function to each element in a collection (array or object values)
	// and returns a new Node containing only the elements for which the function returns true.
	Filter(fn PredicateFunc) Node

	// Map 对集合（数组或对象值）中的每个元素应用转换函数
	// 并返回一个包含转换结果的新 Node（通常是数组）。
	// Map applies a transformation function to each element in a collection (array or object values)
	// and returns a new Node (typically an array) containing the results of the transformations.
	Map(fn TransformFunc) Node

	// RegisterFunc 注册一个可以在查询期间调用的命名函数。
	// 只有 UnaryPathFunc 函数 (Node -> Node) 可以注册以供以后使用。
	// RegisterFunc registers a named function that can be called during queries.
	// Only UnaryPathFunc functions (Node -> Node) can be registered for later use.
	RegisterFunc(name string, fn UnaryPathFunc) Node

	// Apply 立即对节点应用函数（谓词或转换）。
	// 此方法不注册任何内容，只是立即执行该函数。
	// Apply immediately applies a function (predicate or transformation) to the node.
	// This method does not register anything, it just executes the function immediately.
	Apply(fn PathFunc) Node

	// CallFunc 调用已注册的自定义函数。
	// CallFunc calls a registered custom function.
	CallFunc(name string) Node

	// RemoveFunc 删除已注册的自定义函数。
	// RemoveFunc removes a registered custom function.
	RemoveFunc(name string) Node

	// GetFuncs 返回已注册的自定义函数。
	// GetFuncs returns the registered custom functions.
	GetFuncs() *map[string]func(Node) Node
}

// Node 表示 JSON 结构中的单个元素。它提供了一个统一的
// 接口，用于访问和操作 JSON 数据，无论其底层类型如何
// （对象、数组、字符串等）。
//
// 所有遍历和操作方法都设计为可链接的。如果某个操作
// 失败（例如，访问不存在的键或在不适当的节点类型上调用方法），
// 节点将变为 InvalidNode，随后的链式调用将不执行任何操作。
// 错误在内部被记录，并可以在链的末尾使用 Error() 方法检索。
// Node represents a single element within a JSON structure. It provides a unified
// interface for accessing and manipulating JSON data, regardless of its underlying type
// (object, array, string, etc.).
//
// All traversal and manipulation methods are designed to be chainable. If an operation
// fails (e.g., accessing a non-existent key or calling a method on an
// inappropriate node type), the node becomes an InvalidNode, and subsequent
// chained calls will do nothing. The error is recorded internally and can be
// retrieved at the end of the chain using the Error() method.
type Node interface {
	Accessor
	Iterable
	Converter
	Mutator
	Functional
}
