package core

import (
	"time"
)

// NodeType defines the type of a JSON node.
type NodeType int

const (
	// ObjectNode represents a JSON object.
	ObjectNode NodeType = iota
	// ArrayNode represents a JSON array.
	ArrayNode
	// StringNode represents a JSON string.
	StringNode
	// NumberNode represents a JSON number.
	NumberNode
	// BoolNode represents a JSON boolean.
	BoolNode
	// NullNode represents a JSON null.
	NullNode
	// InvalidNode represents an invalid or non-existent node.
	InvalidNode
)

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
	// Type returns the data type of the node.
	Type() NodeType

	// IsValid checks if the node is valid. A node becomes invalid if a preceding
	// operation in a chain fails.
	IsValid() bool

	// Error returns the first error that occurred during a chain of operations.
	// It returns nil if all operations were successful.
	Error() error

	// Path returns the JSON path to the current node from the root.
	Path() string

	// Raw returns the raw JSON string representation of the node.
	Raw() string

	// Get retrieves a child node from an object by its key.
	// If the current node is not an object or the key does not exist, it returns
	// an InvalidNode.
	Get(key string) Node

	// Index retrieves a child node from an array by its index.
	// If the current node is not an array or the index is out of bounds, it returns
	// an InvalidNode.
	Index(i int) Node

	// Query executes a JSONPath-like query and returns the resulting node.
	// The query syntax supports key access, array indexing, and function calls.
	// Example: "users.0.name"
	Query(path string) Node

	// ForEach iterates over the elements of an array or the key-value pairs of an object.
	// For arrays, the callback receives the index and the element node.
	// For objects, the callback receives the key (as a string) and the value node.
	ForEach(iterator func(keyOrIndex interface{}, value Node))

	// Len returns the number of elements in an array or object.
	// It returns 0 for all other node types.
	Len() int

	// String returns the string value of the node.
	// For non-string types, it returns a string representation of the value.
	String() string

	// MustString is like String but panics if the node is not a string type.
	MustString() string

	// Float returns the float64 value of a number node.
	// Returns 0 if the node is not a number.
	Float() float64

	// MustFloat is like Float but panics if the node is not a number type.
	MustFloat() float64

	// Int returns the int64 value of a number node.
	// Returns 0 if the node is not a number.
	Int() int64

	// MustInt is like Int but panics if the node is not a number type.
	MustInt() int64

	// Bool returns the boolean value of a bool node.
	// Returns false if the node is not a boolean.
	Bool() bool

	// MustBool is like Bool but panics if the node is not a bool type.
	MustBool() bool

	// Time returns the time.Time value, assuming the string is in RFC3339 format.
	Time() time.Time

	// MustTime is like Time but panics if parsing fails or node is not a string.
	MustTime() time.Time

	// Array returns a slice of nodes if the node is an array.
	// Returns nil for other types.
	Array() []Node

	// MustArray is like Array but panics if the node is not an array.
	MustArray() []Node

	// Interface returns the underlying value of the node as a standard Go interface{}.
	// This is useful for comparisons and type assertions in tests or application code.
	// - For ObjectNode, it returns map[string]interface{}.
	// - For ArrayNode, it returns []interface{}.
	// - For StringNode, it returns string.
	// - For NumberNode, it returns float64.
	// - For BoolNode, it returns bool.
	// - For NullNode, it returns nil.
	Interface() interface{}

	// Filter applies a function to each element in a collection (array or object values)
	// and returns a new Node containing only the elements for which the function returns true.
	Filter(fn func(Node) bool) Node

	// Map applies a transformation function to each element in a collection (array or object values)
	// and returns a new Node (typically an array) containing the results of the transformations.
	Map(fn func(Node) interface{}) Node

	// Set sets the value for a given key in an object node.
	// If the node is not an object, it returns an InvalidNode.
	Set(key string, value interface{}) Node

	// Append adds a new value to an array node.
	// If the node is not an array, it returns an InvalidNode.
	Append(value interface{}) Node

	// RawFloat returns the underlying float64 value of a number node without creating a new Node object.
	// It returns the value and a boolean indicating success.
	RawFloat() (float64, bool)

	// RawString returns the underlying string value of a string node without creating a new Node object.
	// It returns the value and a boolean indicating success.
	RawString() (string, bool)

	// Strings returns a slice of strings if the node is an array of strings.
	// Returns nil if the node is not an array or contains non-string elements.
	Strings() []string

	// Contains checks if an array node contains a specific string value.
	// Returns false if the node is not an array or does not contain the value.
	Contains(value string) bool

	// Func registers a custom function that can be called during queries.
	Func(name string, fn func(Node) Node) Node

	// CallFunc calls a registered custom function.
	CallFunc(name string) Node

	// RemoveFunc removes a registered custom function.
	RemoveFunc(name string) Node
	GetFuncs() *map[string]func(Node) Node // This will need to be updated to core.Node
}
