package engine

import (
	"errors"
	"time"
)

type nullNode struct {
	baseNode
}

func NewNullNode(path string, funcs *map[string]func(Node) Node) Node {
	if funcs == nil {
		funcs = &map[string]func(Node) Node{} // Initialize if nil (for root)
	}
	return &nullNode{baseNode: baseNode{path: path, funcs: funcs}}
}

func (n *nullNode) Type() NodeType { return NullNode }
func (n *nullNode) Get(key string) Node {
	return NewInvalidNode(n.path+"."+key, ErrTypeAssertion)
}
func (n *nullNode) Index(i int) Node {
	return NewInvalidNode(n.path+"["+string(rune(i))+"]", ErrTypeAssertion)
}
func (n *nullNode) Query(path string) Node {
	return NewInvalidNode(n.path, errors.New("query not implemented"))
}
func (n *nullNode) ForEach(iterator func(interface{}, Node)) {
	_ = n.path // Placeholder for coverage
}
func (n *nullNode) Len() int            { return 0 }
func (n *nullNode) String() string      { return "null" }
func (n *nullNode) MustString() string  { panic(ErrTypeAssertion) }
func (n *nullNode) Float() float64      { return 0 }
func (n *nullNode) MustFloat() float64  { panic(ErrTypeAssertion) }
func (n *nullNode) Int() int64          { return 0 }
func (n *nullNode) MustInt() int64      { panic(ErrTypeAssertion) }
func (n *nullNode) Bool() bool          { return false }
func (n *nullNode) MustBool() bool      { panic(ErrTypeAssertion) }
func (n *nullNode) Time() time.Time     { return time.Time{} }
func (n *nullNode) MustTime() time.Time { panic(ErrTypeAssertion) }
func (n *nullNode) Array() []Node       { return nil }
func (n *nullNode) MustArray() []Node   { panic(ErrTypeAssertion) }
func (n *nullNode) Interface() interface{} {
	return nil
}
func (n *nullNode) Func(name string, fn func(Node) Node) Node {
	if n.err != nil {
		return n
	}
	(*n.funcs)[name] = fn
	return n
}
func (n *nullNode) CallFunc(name string) Node {
	if n.err != nil {
		return n
	}
	if fn, ok := (*n.funcs)[name]; ok {
		return fn(n)
	}
	return NewInvalidNode(n.path, errors.New("function "+name+" not found"))
}
func (n *nullNode) RemoveFunc(name string) Node {
	if n.err != nil {
		return n
	}
	delete(*n.funcs, name)
	return n
}

// New methods for nullNode
func (n *nullNode) Filter(fn func(Node) bool) Node {
	return NewInvalidNode(n.path, ErrTypeAssertion)
}

func (n *nullNode) Map(fn func(Node) interface{}) Node {
	return NewInvalidNode(n.path, ErrTypeAssertion)
}

func (n *nullNode) Set(key string, value interface{}) Node {
	return NewInvalidNode(n.path, ErrTypeAssertion)
}

func (n *nullNode) Append(value interface{}) Node {
	if n.path == "" && n.raw == nil {
		return NewInvalidNode(n.path, ErrTypeAssertion)
	}
	n.setError(ErrTypeAssertion)
	return n
}

func (n *nullNode) Raw() string {
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

func (n *nullNode) AsMap() map[string]Node     { return nil }
func (n *nullNode) MustAsMap() map[string]Node { panic(ErrTypeAssertion) }
