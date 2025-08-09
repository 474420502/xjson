package engine

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/474420502/xjson/internal/core"
)

type stringNode struct {
	baseNode
	value string
}

func NewStringNode(value string, path string, funcs *map[string]func(core.Node) core.Node) core.Node {
	if funcs == nil {
		funcs = &map[string]func(core.Node) core.Node{} // Initialize if nil (for root)
	}
	return &stringNode{value: value, baseNode: baseNode{path: path, funcs: funcs}}
}

func (n *stringNode) Type() core.NodeType { return core.StringNode }
func (n *stringNode) Get(key string) core.Node {
	return NewInvalidNode(n.path+"."+key, ErrTypeAssertion)
}
func (n *stringNode) Index(i int) core.Node {
	return NewInvalidNode(n.path+strconv.FormatInt(int64(i), 10), ErrTypeAssertion)
}
func (n *stringNode) Query(path string) core.Node {
	return NewInvalidNode(n.path, errors.New("query not implemented"))
}
func (n *stringNode) ForEach(iterator func(interface{}, core.Node)) {
	_ = n.path // Placeholder for coverage
}
func (n *stringNode) Len() int { return len(n.value) }
func (n *stringNode) String() string {
	if n.err != nil {
		return ""
	}
	return n.value
}
func (n *stringNode) MustString() string {
	if n.err != nil {
		panic(n.err)
	}
	return n.value
}
func (n *stringNode) Float() float64     { return 0 }
func (n *stringNode) MustFloat() float64 { panic(ErrTypeAssertion) }
func (n *stringNode) Int() int64         { return 0 }
func (n *stringNode) MustInt() int64     { panic(ErrTypeAssertion) }
func (n *stringNode) Bool() bool         { return false }
func (n *stringNode) MustBool() bool     { panic(ErrTypeAssertion) }
func (n *stringNode) Time() time.Time {
	if n.err != nil {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339Nano, n.value)
	if err != nil {
		n.setError(err)
		return time.Time{}
	}
	return t
}
func (n *stringNode) MustTime() time.Time {
	if n.err != nil {
		panic(n.err)
	}
	t, err := time.Parse(time.RFC3339Nano, n.value)
	if err != nil {
		panic(err)
	}
	return t
}
func (n *stringNode) Array() []core.Node     { return nil }
func (n *stringNode) MustArray() []core.Node { panic(ErrTypeAssertion) }
func (n *stringNode) Interface() interface{} {
	if n.err != nil {
		return nil
	}
	return n.value
}

// Deprecated: Use RegisterFunc and CallFunc instead
func (n *stringNode) Func(name string, fn func(core.Node) core.Node) core.Node {
	if n.err != nil {
		return n
	}
	(*n.funcs)[name] = fn
	return n
}

func (n *stringNode) RegisterFunc(name string, fn core.UnaryPathFunc) core.Node {
	if n.err != nil {
		return n
	}
	if n.funcs == nil {
		n.funcs = &map[string]func(core.Node) core.Node{}
	}
	(*n.funcs)[name] = fn
	return n
}

func (n *stringNode) Apply(fn core.PathFunc) core.Node {
	if n.err != nil {
		return n
	}

	switch f := fn.(type) {
	case core.PredicateFunc:
		// For a string node, if predicate returns true, return the node itself
		if f(n) {
			return n
		}
		// Otherwise return an invalid node
		return NewInvalidNode(n.path, fmt.Errorf("predicate returned false for string node"))
	case core.TransformFunc:
		// Apply transform function and create a new node from the result
		transformed := f(n)
		newNode, err := NewNodeFromInterface(transformed, n.path, n.funcs)
		if err != nil {
			return NewInvalidNode(n.path, err)
		}
		return newNode
	default:
		return NewInvalidNode(n.path, fmt.Errorf("unsupported function signature for Apply: %T", f))
	}
}

func (n *stringNode) CallFunc(name string) core.Node {
	if n.err != nil {
		return n
	}
	if fn, ok := (*n.funcs)[name]; ok {
		return fn(n)
	}
	return NewInvalidNode(n.path, errors.New("function "+name+" not found"))
}
func (n *stringNode) RemoveFunc(name string) core.Node {
	if n.err != nil {
		return n
	}
	delete(*n.funcs, name)
	return n
}

// New methods for stringNode
func (n *stringNode) Filter(fn core.PredicateFunc) core.Node {
	return NewInvalidNode(n.path, ErrTypeAssertion)
}

func (n *stringNode) Map(fn core.TransformFunc) core.Node {
	return NewInvalidNode(n.path, ErrTypeAssertion)
}

func (n *stringNode) Set(key string, value interface{}) core.Node {
	return NewInvalidNode(n.path, ErrTypeAssertion)
}

func (n *stringNode) Append(value interface{}) core.Node {
	if n.path == "" && n.raw == nil {
		return NewInvalidNode(n.path, ErrTypeAssertion)
	}
	n.setError(ErrTypeAssertion)
	return n
}

func (n *stringNode) Raw() string {
	if n.raw != nil {
		return *n.raw
	}
	if n.err != nil {
		return "null" // Match JSON representation of an error state
	}
	// For a non-root string node, marshal its value to get a valid JSON string literal.
	b, err := json.Marshal(n.value)
	if err != nil {
		// This should theoretically not happen for a simple string.
		n.setError(err)
		return "null"
	}
	return string(b)
}

func (n *stringNode) RawFloat() (float64, bool) {
	if n.err != nil {
		return 0, false
	}
	f, err := strconv.ParseFloat(n.value, 64)
	if err != nil {
		return 0, false
	}
	return f, true
}

func (n *stringNode) RawString() (string, bool) {
	if n.err != nil {
		return "", false
	}
	return n.value, true
}

func (n *stringNode) Strings() []string {
	if n.err != nil {
		return nil
	}
	return []string{n.value}
}

func (n *stringNode) Contains(value string) bool {
	if n.err != nil {
		return false
	}
	return strings.Contains(n.value, value)
}

func (n *stringNode) AsMap() map[string]core.Node     { return nil }
func (n *stringNode) MustAsMap() map[string]core.Node { panic(ErrTypeAssertion) }
