package engine

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/474420502/xjson/internal/core"
)

type boolNode struct {
	baseNode
	value bool
}

func NewBoolNode(value bool, path string, funcs *map[string]func(core.Node) core.Node) core.Node {
	if funcs == nil {
		funcs = &map[string]func(core.Node) core.Node{} // Initialize if nil (for root)
	}
	return &boolNode{value: value, baseNode: baseNode{path: path, funcs: funcs}}
}

func (n *boolNode) Type() core.NodeType { return core.BoolNode }
func (n *boolNode) Get(key string) core.Node {
	return NewInvalidNode(n.path+"."+key, ErrTypeAssertion)
}
func (n *boolNode) Index(i int) core.Node {
	return NewInvalidNode(n.path+strconv.FormatInt(int64(i), 10), ErrTypeAssertion)
}
func (n *boolNode) Query(path string) core.Node {
	return NewInvalidNode(n.path, errors.New("query not implemented"))
}
func (n *boolNode) ForEach(iterator func(interface{}, core.Node)) {
	_ = n.path // Placeholder for coverage
}
func (n *boolNode) Len() int           { return 0 }
func (n *boolNode) String() string     { return strconv.FormatBool(n.value) }
func (n *boolNode) MustString() string { panic(ErrTypeAssertion) }
func (n *boolNode) Float() float64     { return 0 }
func (n *boolNode) MustFloat() float64 { panic(ErrTypeAssertion) }
func (n *boolNode) Int() int64         { return 0 }
func (n *boolNode) MustInt() int64     { panic(ErrTypeAssertion) }
func (n *boolNode) Bool() bool {
	if n.err != nil {
		return false
	}
	return n.value
}
func (n *boolNode) MustBool() bool {
	if n.err != nil {
		panic(n.err)
	}
	return n.value
}
func (n *boolNode) Time() time.Time        { return time.Time{} }
func (n *boolNode) MustTime() time.Time    { panic(ErrTypeAssertion) }
func (n *boolNode) Array() []core.Node     { return nil }
func (n *boolNode) MustArray() []core.Node { panic(ErrTypeAssertion) }
func (n *boolNode) Interface() interface{} {
	if n.err != nil {
		return nil
	}
	return n.value
}

// Deprecated: Use RegisterFunc and CallFunc instead
func (n *boolNode) Func(name string, fn func(core.Node) core.Node) core.Node {
	if n.err != nil {
		return n
	}
	(*n.funcs)[name] = fn
	return n
}

func (n *boolNode) RegisterFunc(name string, fn core.UnaryPathFunc) core.Node {
	if n.err != nil {
		return n
	}
	if n.funcs == nil {
		n.funcs = &map[string]func(core.Node) core.Node{}
	}
	(*n.funcs)[name] = fn
	return n
}

func (n *boolNode) Apply(fn core.PathFunc) core.Node {
	if n.err != nil {
		return n
	}

	switch f := fn.(type) {
	case core.PredicateFunc:
		// For a boolean node, if predicate returns true, return the node itself
		if f(n) {
			return n
		}
		// Otherwise return an invalid node
		return NewInvalidNode(n.path, fmt.Errorf("predicate returned false for boolean node"))
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

func (n *boolNode) CallFunc(name string) core.Node {
	if n.err != nil {
		return n
	}
	if fn, ok := (*n.funcs)[name]; ok {
		return fn(n)
	}
	return NewInvalidNode(n.path, errors.New("function "+name+" not found"))
}
func (n *boolNode) RemoveFunc(name string) core.Node {
	if n.err != nil {
		return n
	}
	delete(*n.funcs, name)
	return n
}

// New methods for boolNode
func (n *boolNode) Filter(fn core.PredicateFunc) core.Node {
	return NewInvalidNode(n.path, ErrTypeAssertion)
}

func (n *boolNode) Map(fn core.TransformFunc) core.Node {
	return NewInvalidNode(n.path, ErrTypeAssertion)
}

func (n *boolNode) Set(key string, value interface{}) core.Node {
	return NewInvalidNode(n.path, ErrTypeAssertion)
}

func (n *boolNode) Append(value interface{}) core.Node {
	if n.path == "" && n.raw == nil {
		return NewInvalidNode(n.path, ErrTypeAssertion)
	}
	n.setError(ErrTypeAssertion)
	return n
}

func (n *boolNode) Raw() string {
	if n.raw != nil {
		return *n.raw
	}
	if n.err != nil {
		return ""
	}
	return strconv.FormatBool(n.value)
}

func (n *boolNode) RawFloat() (float64, bool) {
	return 0, false
}

func (n *boolNode) RawString() (string, bool) {
	return "", false
}

func (n *boolNode) Strings() []string {
	return nil
}

func (n *boolNode) Contains(value string) bool {
	return false
}

func (n *boolNode) AsMap() map[string]core.Node     { return nil }
func (n *boolNode) MustAsMap() map[string]core.Node { panic(ErrTypeAssertion) }
