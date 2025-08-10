package core

import (
	"fmt"
	"strconv"
	"time"
)

// NodeType defines the type of a JSON node.
type NodeType int

const (
	Invalid NodeType = iota
	Object
	Array
	String
	Number
	Bool
	Null
)

// String returns the string representation of the NodeType.
func (t NodeType) String() string {
	switch t {
	case Object:
		return "object"
	case Array:
		return "array"
	case String:
		return "string"
	case Number:
		return "number"
	case Bool:
		return "bool"
	case Null:
		return "null"
	default:
		return "invalid"
	}
}

// PathFunc is a generic function container for path operations.
type PathFunc interface{}

// UnaryPathFunc is a function that transforms a node to another node.
type UnaryPathFunc func(node Node) Node

// PredicateFunc is a function that returns true or false for a node.
type PredicateFunc func(node Node) bool

// TransformFunc is a function that transforms a node into any value.
type TransformFunc func(node Node) interface{}

// Node represents any element in a JSON structure.
type Node interface {
	Type() NodeType
	IsValid() bool
	Error() error
	Path() string
	Raw() string
	Query(path string) Node
	Get(key string) Node
	Index(i int) Node
	Filter(fn PredicateFunc) Node
	Map(fn TransformFunc) Node
	ForEach(fn func(keyOrIndex interface{}, value Node))
	Len() int
	Set(key string, value interface{}) Node
	Append(value interface{}) Node
	SetValue(value interface{}) Node
	RegisterFunc(name string, fn UnaryPathFunc) Node
	CallFunc(name string) Node
	RemoveFunc(name string) Node
	Apply(fn PathFunc) Node
	GetFuncs() *map[string]UnaryPathFunc
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
	RawFloat() (float64, bool)
	RawString() (string, bool)
	Strings() []string
	Keys() []string
	Contains(value string) bool
	AsMap() map[string]Node
	MustAsMap() map[string]Node
}

// OpCode defines the type of a query operation.
type OpCode int

const (
	OpKey OpCode = iota
	OpIndex
	OpSlice
	OpFunc
	OpWildcard
	OpRecursive
	OpParent
)

// QueryToken represents a single parsed operation from a query path.
type QueryToken struct {
	Op    OpCode
	Type  OpCode
	Value interface{}
}

// Pre-defined errors for common issues.
var (
	ErrNotFound         = fmt.Errorf("path or key not found")
	ErrTypeAssertion    = fmt.Errorf("type assertion failed")
	ErrIndexOutOfBounds = fmt.Errorf("index out of bounds")
	ErrInvalidNode      = fmt.Errorf("operation on an invalid node")
)

// --- Node Implementations ---

// BaseNode contains the common fields and methods for all node types.
type BaseNode struct {
	RawBytes   []byte
	StartPos   int
	EndPos     int
	PathStr    string
	ParentNode Node
	Err        error
	Funcs      *map[string]UnaryPathFunc
}

func (n *BaseNode) Error() error {
	return n.Err
}

func (n *BaseNode) SetError(err error) {
	if n.Err == nil {
		n.Err = err
	}
}

func (n *BaseNode) Raw() string {
	if n.StartPos >= n.EndPos {
		return ""
	}
	return string(n.RawBytes[n.StartPos:n.EndPos])
}

func (n *BaseNode) IsValid() bool {
	return n.Err == nil
}

func (n *BaseNode) GetFuncs() *map[string]UnaryPathFunc {
	return n.Funcs
}

func (n *BaseNode) Path() string {
	return n.PathStr
}

// InvalidNode represents a node in an error state.
type InvalidNode struct {
	BaseNode
}

func (n *InvalidNode) Type() NodeType { return Invalid }
func (n *InvalidNode) IsValid() bool  { return false }

// ... (All other Node methods for InvalidNode return self or zero value)
func (n *InvalidNode) Query(path string) Node                              { return n }
func (n *InvalidNode) Get(key string) Node                                 { return n }
func (n *InvalidNode) Index(i int) Node                                    { return n }
func (n *InvalidNode) Filter(fn PredicateFunc) Node                        { return n }
func (n *InvalidNode) Map(fn TransformFunc) Node                           { return n }
func (n *InvalidNode) ForEach(fn func(keyOrIndex interface{}, value Node)) {}
func (n *InvalidNode) Len() int                                            { return 0 }
func (n *InvalidNode) Set(key string, value interface{}) Node              { return n }
func (n *InvalidNode) Append(value interface{}) Node                       { return n }
func (n *InvalidNode) SetValue(value interface{}) Node                     { return n }
func (n *InvalidNode) RegisterFunc(name string, fn UnaryPathFunc) Node     { return n }
func (n *InvalidNode) CallFunc(name string) Node                           { return n }
func (n *InvalidNode) RemoveFunc(name string) Node                         { return n }
func (n *InvalidNode) Apply(fn PathFunc) Node                              { return n }
func (n *InvalidNode) String() string                                      { return "" }
func (n *InvalidNode) MustString() string                                  { panic(n.Err) }
func (n *InvalidNode) Float() float64                                      { return 0 }
func (n *InvalidNode) MustFloat() float64                                  { panic(n.Err) }
func (n *InvalidNode) Int() int64                                          { return 0 }
func (n *InvalidNode) MustInt() int64                                      { panic(n.Err) }
func (n *InvalidNode) Bool() bool                                          { return false }
func (n *InvalidNode) MustBool() bool                                      { panic(n.Err) }
func (n *InvalidNode) Time() time.Time                                     { return time.Time{} }
func (n *InvalidNode) MustTime() time.Time                                 { panic(n.Err) }
func (n *InvalidNode) Array() []Node                                       { return nil }
func (n *InvalidNode) MustArray() []Node                                   { panic(n.Err) }
func (n *InvalidNode) Interface() interface{}                              { return nil }
func (n *InvalidNode) RawFloat() (float64, bool)                           { return 0, false }
func (n *InvalidNode) RawString() (string, bool)                           { return "", false }
func (n *InvalidNode) Strings() []string                                   { return nil }
func (n *InvalidNode) Keys() []string                                      { return nil }
func (n *InvalidNode) Contains(value string) bool                          { return false }
func (n *InvalidNode) AsMap() map[string]Node                              { return nil }
func (n *InvalidNode) MustAsMap() map[string]Node                          { panic(n.Err) }

// ObjectNode represents a JSON object.
type ObjectNode struct {
	BaseNode
	Children map[string]Node
}

func (n *ObjectNode) Type() NodeType { return Object }
func (n *ObjectNode) Get(key string) Node {
	if n.Err != nil {
		return &InvalidNode{BaseNode{Err: n.Err}}
	}
	// Lazy parsing logic is now in the engine. Assume children are populated.
	if child, ok := n.Children[key]; ok {
		return child
	}
	return &InvalidNode{BaseNode{Err: ErrNotFound}}
}
func (n *ObjectNode) ForEach(fn func(keyOrIndex interface{}, value Node)) {
	if n.Err != nil {
		return
	}
	for k, v := range n.Children {
		fn(k, v)
	}
}
func (n *ObjectNode) Len() int {
	if n.Err != nil {
		return 0
	}
	return len(n.Children)
}
func (n *ObjectNode) AsMap() map[string]Node {
	if n.Err != nil {
		return nil
	}
	return n.Children
}
func (n *ObjectNode) MustAsMap() map[string]Node {
	if n.Err != nil {
		panic(n.Err)
	}
	return n.Children
}
func (n *ObjectNode) Keys() []string {
	if n.Err != nil {
		return nil
	}
	keys := make([]string, 0, len(n.Children))
	for k := range n.Children {
		keys = append(keys, k)
	}
	return keys
}
func (n *ObjectNode) Interface() interface{} {
	if n.Err != nil {
		return nil
	}
	m := make(map[string]interface{})
	for k, v := range n.Children {
		m[k] = v.Interface()
	}
	return m
}

//... other ObjectNode methods

// ArrayNode represents a JSON array.
type ArrayNode struct {
	BaseNode
	Children []Node
}

func (n *ArrayNode) Type() NodeType { return Array }
func (n *ArrayNode) Index(i int) Node {
	if n.Err != nil {
		return &InvalidNode{BaseNode{Err: n.Err}}
	}
	// Lazy parsing logic is now in the engine.
	if i < 0 {
		i = len(n.Children) + i
	}
	if i < 0 || i >= len(n.Children) {
		return &InvalidNode{BaseNode{Err: ErrIndexOutOfBounds}}
	}
	return n.Children[i]
}
func (n *ArrayNode) ForEach(fn func(keyOrIndex interface{}, value Node)) {
	if n.Err != nil {
		return
	}
	for i, v := range n.Children {
		fn(i, v)
	}
}
func (n *ArrayNode) Len() int {
	if n.Err != nil {
		return 0
	}
	return len(n.Children)
}
func (n *ArrayNode) Array() []Node {
	if n.Err != nil {
		return nil
	}
	return n.Children
}
func (n *ArrayNode) MustArray() []Node {
	if n.Err != nil {
		panic(n.Err)
	}
	return n.Children
}
func (n *ArrayNode) Keys() []string { return nil }
func (n *ArrayNode) Interface() interface{} {
	if n.Err != nil {
		return nil
	}
	s := make([]interface{}, len(n.Children))
	for i, v := range n.Children {
		s[i] = v.Interface()
	}
	return s
}

//... other ArrayNode methods

// StringNode represents a JSON string.
type StringNode struct {
	BaseNode
	Value string
}

func (n *StringNode) Type() NodeType            { return String }
func (n *StringNode) RawString() (string, bool) { return n.Value, true }
func (n *StringNode) String() string            { return n.Value }
func (n *StringNode) MustString() string        { return n.Value }
func (n *StringNode) Float() float64 {
	f, err := strconv.ParseFloat(n.Value, 64)
	if err != nil {
		n.SetError(err)
		return 0
	}
	return f
}
func (n *StringNode) MustFloat() float64 {
	f, err := strconv.ParseFloat(n.Value, 64)
	if err != nil {
		panic(err)
	}
	return f
}
func (n *StringNode) Int() int64 {
	i, err := strconv.ParseInt(n.Value, 10, 64)
	if err != nil {
		n.SetError(err)
		return 0
	}
	return i
}
func (n *StringNode) MustInt() int64 {
	i, err := strconv.ParseInt(n.Value, 10, 64)
	if err != nil {
		panic(err)
	}
	return i
}
func (n *StringNode) Bool() bool {
	b, err := strconv.ParseBool(n.Value)
	if err != nil {
		n.SetError(err)
		return false
	}
	return b
}
func (n *StringNode) MustBool() bool {
	b, err := strconv.ParseBool(n.Value)
	if err != nil {
		panic(err)
	}
	return b
}
func (n *StringNode) Time() time.Time {
	t, err := time.Parse(time.RFC3339, n.Value)
	if err != nil {
		n.SetError(err)
		return time.Time{}
	}
	return t
}
func (n *StringNode) MustTime() time.Time {
	t, err := time.Parse(time.RFC3339, n.Value)
	if err != nil {
		panic(err)
	}
	return t
}
func (n *StringNode) Interface() interface{} { return n.Value }
func (n *StringNode) Keys() []string         { return nil }

// NumberNode represents a JSON number.
type NumberNode struct {
	BaseNode
}

func (n *NumberNode) Type() NodeType { return Number }
func (n *NumberNode) RawFloat() (float64, bool) {
	f, err := strconv.ParseFloat(n.Raw(), 64)
	return f, err == nil
}
func (n *NumberNode) Float() float64 {
	f, _ := n.RawFloat()
	return f
}
func (n *NumberNode) MustFloat() float64 {
	f, err := strconv.ParseFloat(n.Raw(), 64)
	if err != nil {
		panic(err)
	}
	return f
}
func (n *NumberNode) Int() int64 {
	i, err := strconv.ParseInt(n.Raw(), 10, 64)
	if err != nil {
		if f, ok := n.RawFloat(); ok {
			return int64(f)
		}
		n.SetError(err)
		return 0
	}
	return i
}
func (n *NumberNode) MustInt() int64 {
	i, err := strconv.ParseInt(n.Raw(), 10, 64)
	if err != nil {
		if f, ok := n.RawFloat(); ok {
			return int64(f)
		}
		panic(err)
	}
	return i
}
func (n *NumberNode) Interface() interface{} {
	if i, err := strconv.ParseInt(n.Raw(), 10, 64); err == nil {
		return i
	}
	f, _ := n.RawFloat()
	return f
}
func (n *NumberNode) Keys() []string { return nil }

// BoolNode represents a JSON boolean.
type BoolNode struct {
	BaseNode
	Value bool
}

func (n *BoolNode) Type() NodeType         { return Bool }
func (n *BoolNode) Bool() bool             { return n.Value }
func (n *BoolNode) MustBool() bool         { return n.Value }
func (n *BoolNode) Interface() interface{} { return n.Value }
func (n *BoolNode) Keys() []string         { return nil }

// NullNode represents a JSON null.
type NullNode struct {
	BaseNode
}

func (n *NullNode) Type() NodeType         { return Null }
func (n *NullNode) Interface() interface{} { return nil }
func (n *NullNode) Keys() []string         { return nil }

// Placeholder methods to satisfy the Node interface for all concrete types.
// These should be implemented with actual logic.

func (n *BaseNode) Query(path string) Node {
	n.SetError(fmt.Errorf("not implemented"))
	return &InvalidNode{BaseNode{Err: n.Err}}
}
func (n *BaseNode) Get(key string) Node {
	n.SetError(fmt.Errorf("not a key-value node"))
	return &InvalidNode{BaseNode{Err: n.Err}}
}
func (n *BaseNode) Index(i int) Node {
	n.SetError(fmt.Errorf("not an array node"))
	return &InvalidNode{BaseNode{Err: n.Err}}
}
func (n *BaseNode) Filter(fn PredicateFunc) Node {
	n.SetError(fmt.Errorf("not implemented"))
	return &InvalidNode{BaseNode{Err: n.Err}}
}
func (n *BaseNode) Map(fn TransformFunc) Node {
	n.SetError(fmt.Errorf("not implemented"))
	return &InvalidNode{BaseNode{Err: n.Err}}
}
func (n *BaseNode) ForEach(fn func(keyOrIndex interface{}, value Node)) {
	n.SetError(fmt.Errorf("not a collection node"))
}
func (n *BaseNode) Len() int { n.SetError(fmt.Errorf("not a collection node")); return 0 }
func (n *BaseNode) Set(key string, value interface{}) Node {
	n.SetError(fmt.Errorf("not implemented"))
	return &InvalidNode{BaseNode{Err: n.Err}}
}
func (n *BaseNode) Append(value interface{}) Node {
	n.SetError(fmt.Errorf("not implemented"))
	return &InvalidNode{BaseNode{Err: n.Err}}
}
func (n *BaseNode) SetValue(value interface{}) Node {
	n.SetError(fmt.Errorf("not implemented"))
	return &InvalidNode{BaseNode{Err: n.Err}}
}
func (n *BaseNode) RegisterFunc(name string, fn UnaryPathFunc) Node {
	n.SetError(fmt.Errorf("not implemented"))
	return &InvalidNode{BaseNode{Err: n.Err}}
}
func (n *BaseNode) CallFunc(name string) Node {
	n.SetError(fmt.Errorf("not implemented"))
	return &InvalidNode{BaseNode{Err: n.Err}}
}
func (n *BaseNode) RemoveFunc(name string) Node {
	n.SetError(fmt.Errorf("not implemented"))
	return &InvalidNode{BaseNode{Err: n.Err}}
}
func (n *BaseNode) Apply(fn PathFunc) Node {
	n.SetError(fmt.Errorf("not implemented"))
	return &InvalidNode{BaseNode{Err: n.Err}}
}
func (n *BaseNode) MustString() string         { n.SetError(ErrTypeAssertion); panic(n.Err) }
func (n *BaseNode) Float() float64             { n.SetError(ErrTypeAssertion); return 0 }
func (n *BaseNode) MustFloat() float64         { n.SetError(ErrTypeAssertion); panic(n.Err) }
func (n *BaseNode) Int() int64                 { n.SetError(ErrTypeAssertion); return 0 }
func (n *BaseNode) MustInt() int64             { n.SetError(ErrTypeAssertion); panic(n.Err) }
func (n *BaseNode) Bool() bool                 { n.SetError(ErrTypeAssertion); return false }
func (n *BaseNode) MustBool() bool             { n.SetError(ErrTypeAssertion); panic(n.Err) }
func (n *BaseNode) Time() time.Time            { n.SetError(ErrTypeAssertion); return time.Time{} }
func (n *BaseNode) MustTime() time.Time        { n.SetError(ErrTypeAssertion); panic(n.Err) }
func (n *BaseNode) Array() []Node              { n.SetError(ErrTypeAssertion); return nil }
func (n *BaseNode) MustArray() []Node          { n.SetError(ErrTypeAssertion); panic(n.Err) }
func (n *BaseNode) RawFloat() (float64, bool)  { n.SetError(ErrTypeAssertion); return 0, false }
func (n *BaseNode) RawString() (string, bool)  { n.SetError(ErrTypeAssertion); return "", false }
func (n *BaseNode) Strings() []string          { n.SetError(ErrTypeAssertion); return nil }
func (n *BaseNode) Keys() []string             { n.SetError(ErrTypeAssertion); return nil }
func (n *BaseNode) Contains(value string) bool { n.SetError(ErrTypeAssertion); return false }
func (n *BaseNode) AsMap() map[string]Node     { n.SetError(ErrTypeAssertion); return nil }
func (n *BaseNode) MustAsMap() map[string]Node { n.SetError(ErrTypeAssertion); panic(n.Err) }
