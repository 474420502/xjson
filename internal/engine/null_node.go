package engine

import (
	"errors"
	"fmt"
	"time"

	"github.com/474420502/xjson/internal/core"
)

type nullNode struct {
	baseNode
}

func NewNullNode(path string, funcs *map[string]func(core.Node) core.Node) core.Node {
	if funcs == nil {
		funcs = &map[string]func(core.Node) core.Node{} // Initialize if nil (for root)
	}
	return &nullNode{baseNode: baseNode{path: path, funcs: funcs}}
}

func (n *nullNode) Type() core.NodeType { return core.NullNode }
func (n *nullNode) Get(key string) core.Node {
	return NewInvalidNode(n.path+"."+key, ErrTypeAssertion)
}
func (n *nullNode) Index(i int) core.Node {
	return NewInvalidNode(n.path+"["+string(rune(i))+"]", ErrTypeAssertion)
}
func (n *nullNode) Query(path string) core.Node {
	return NewInvalidNode(n.path, errors.New("query not implemented"))
}
func (n *nullNode) ForEach(iterator func(interface{}, core.Node)) {
	_ = n.path // Placeholder for coverage
}
func (n *nullNode) Len() int               { return 0 }
func (n *nullNode) String() string         { return "null" }
func (n *nullNode) MustString() string     { panic(ErrTypeAssertion) }
func (n *nullNode) Float() float64         { return 0 }
func (n *nullNode) MustFloat() float64     { panic(ErrTypeAssertion) }
func (n *nullNode) Int() int64             { return 0 }
func (n *nullNode) MustInt() int64         { panic(ErrTypeAssertion) }
func (n *nullNode) Bool() bool             { return false }
func (n *nullNode) MustBool() bool         { panic(ErrTypeAssertion) }
func (n *nullNode) Time() time.Time        { return time.Time{} }
func (n *nullNode) MustTime() time.Time    { panic(ErrTypeAssertion) }
func (n *nullNode) Array() []core.Node     { return nil }
func (n *nullNode) MustArray() []core.Node { panic(ErrTypeAssertion) }
func (n *nullNode) Interface() interface{} {
	return nil
}

// Deprecated: Use RegisterFunc and CallFunc instead
func (n *nullNode) Func(name string, fn func(core.Node) core.Node) core.Node {
	if n.err != nil {
		return n
	}
	(*n.funcs)[name] = fn
	return n
}

func (n *nullNode) RegisterFunc(name string, fn core.UnaryPathFunc) core.Node {
	if n.err != nil {
		return n
	}
	if n.funcs == nil {
		n.funcs = &map[string]func(core.Node) core.Node{}
	}
	(*n.funcs)[name] = fn
	return n
}

func (n *nullNode) Apply(fn core.PathFunc) core.Node {
	if fn == nil {
		panic("Apply function cannot be nil")
	}
	if n.err != nil {
		return n
	}

	switch f := fn.(type) {
	case core.PredicateFunc:
		return n.Filter(f)
	case core.TransformFunc:
		return n.Map(f)
	default:
		return NewInvalidNode(n.path, fmt.Errorf("unsupported function signature for Apply: %T", f))
	}
}

func (n *nullNode) CallFunc(name string) core.Node {
	if n.err != nil {
		return n
	}
	if fn, ok := (*n.funcs)[name]; ok {
		return fn(n)
	}
	return NewInvalidNode(n.path, errors.New("function "+name+" not found"))
}
func (n *nullNode) RemoveFunc(name string) core.Node {
	if n.err != nil {
		return n
	}
	delete(*n.funcs, name)
	return n
}

// New methods for nullNode
func (n *nullNode) Filter(fn core.PredicateFunc) core.Node {
	return NewInvalidNode(n.path, ErrTypeAssertion)
}

func (n *nullNode) Map(fn core.TransformFunc) core.Node {
	return NewInvalidNode(n.path, ErrTypeAssertion)
}

func (n *nullNode) Set(key string, value interface{}) core.Node {
	return NewInvalidNode(n.path, ErrTypeAssertion)
}

func (n *nullNode) Append(value interface{}) core.Node {
	if n.path == "" && n.raw == nil {
		return NewInvalidNode(n.path, ErrTypeAssertion)
	}
	n.setError(ErrTypeAssertion)
	return n
}

func (n *nullNode) Raw() string {
	if n.raw != nil {
		return *n.raw
	}
	return "null"
}

func (n *nullNode) RawFloat() (float64, bool) {
	return 0, false
}

func (n *nullNode) RawString() (string, bool) {
	return "", false
}

func (n *nullNode) Strings() []string {
	return nil
}

func (n *nullNode) Contains(value string) bool {
	return false
}

func (n *nullNode) AsMap() map[string]core.Node     { return nil }
func (n *nullNode) MustAsMap() map[string]core.Node { panic(ErrTypeAssertion) }
