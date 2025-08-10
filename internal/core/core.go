package core

import (
	"errors"
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
	Parent() Node
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

// ErrTypeAssertion is returned when a Must* conversion fails.
var ErrTypeAssertion = errors.New("type assertion failed")
