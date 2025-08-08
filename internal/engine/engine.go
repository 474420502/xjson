package engine

import (
	"errors"
	"time"
)

// NodeType defines the type of a JSON node.
type NodeType int

const (
	ObjectNode NodeType = iota
	ArrayNode
	StringNode
	NumberNode
	BoolNode
	NullNode
	InvalidNode
)

func (t NodeType) String() string {
	switch t {
	case ObjectNode:
		return "ObjectNode"
	case ArrayNode:
		return "ArrayNode"
	case StringNode:
		return "StringNode"
	case NumberNode:
		return "NumberNode"
	case BoolNode:
		return "BoolNode"
	case NullNode:
		return "NullNode"
	case InvalidNode:
		return "InvalidNode"
	default:
		return "UnknownNode"
	}
}

// Node represents a single element within a JSON structure.
type Node interface {
	Type() NodeType
	IsValid() bool
	Error() error
	Path() string
	Raw() string
	Get(key string) Node
	Index(i int) Node
	Query(path string) Node
	ForEach(iterator func(keyOrIndex interface{}, value Node))
	Len() int
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
	Filter(fn func(Node) bool) Node
	Map(fn func(Node) interface{}) Node
	Set(key string, value interface{}) Node
	Append(value interface{}) Node
	RawFloat() (float64, bool)
	RawString() (string, bool)
	Strings() []string
	Contains(value string) bool
	Func(name string, fn func(Node) Node) Node
	CallFunc(name string) Node
	RemoveFunc(name string) Node
	GetFuncs() *map[string]func(Node) Node
}

var (
	ErrInvalidNode      = errors.New("invalid node")
	ErrTypeAssertion    = errors.New("type assertion failed")
	ErrIndexOutOfBounds = errors.New("index out of bounds")
	ErrNotFound         = errors.New("not found")
)
