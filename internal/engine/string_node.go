package engine

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

type stringNode struct {
	baseNode
	value string
}

func NewStringNode(value string, path string, funcs *map[string]func(Node) Node) Node {
	if funcs == nil {
		funcs = &map[string]func(Node) Node{} // Initialize if nil (for root)
	}
	return &stringNode{value: value, baseNode: baseNode{path: path, funcs: funcs}}
}

func (n *stringNode) Type() NodeType { return StringNode }
func (n *stringNode) Get(key string) Node {
	return NewInvalidNode(n.path+"."+key, ErrTypeAssertion)
}
func (n *stringNode) Index(i int) Node {
	return NewInvalidNode(n.path+strconv.FormatInt(int64(i), 10), ErrTypeAssertion)
}
func (n *stringNode) Query(path string) Node {
	return NewInvalidNode(n.path, errors.New("query not implemented"))
}
func (n *stringNode) ForEach(iterator func(interface{}, Node)) {
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
	t, err := time.Parse(time.RFC3339, n.value)
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
	t, err := time.Parse(time.RFC3339, n.value)
	if err != nil {
		panic(err)
	}
	return t
}
func (n *stringNode) Array() []Node     { return nil }
func (n *stringNode) MustArray() []Node { panic(ErrTypeAssertion) }
func (n *stringNode) Interface() interface{} {
	if n.err != nil {
		return nil
	}
	return n.value
}
func (n *stringNode) Func(name string, fn func(Node) Node) Node {
	if n.err != nil {
		return n
	}
	(*n.funcs)[name] = fn
	return n
}
func (n *stringNode) CallFunc(name string) Node {
	if n.err != nil {
		return n
	}
	if fn, ok := (*n.funcs)[name]; ok {
		return fn(n)
	}
	return NewInvalidNode(n.path, errors.New("function "+name+" not found"))
}
func (n *stringNode) RemoveFunc(name string) Node {
	if n.err != nil {
		return n
	}
	delete(*n.funcs, name)
	return n
}

// New methods for stringNode
func (n *stringNode) Filter(fn func(Node) bool) Node {
	// Return a new invalid node; do not mutate original so it can still be used afterwards
	return NewInvalidNode(n.path, ErrTypeAssertion)
}

func (n *stringNode) Map(fn func(Node) interface{}) Node {
	return NewInvalidNode(n.path, ErrTypeAssertion)
}

func (n *stringNode) Set(key string, value interface{}) Node {
	return NewInvalidNode(n.path, ErrTypeAssertion)
}

func (n *stringNode) Append(value interface{}) Node {
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
		return ""
	}
	return `"` + n.value + `"`
}

func (n *stringNode) RawFloat() (float64, bool) {
	return 0, false
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

func (n *stringNode) AsMap() map[string]Node     { return nil }
func (n *stringNode) MustAsMap() map[string]Node { panic(ErrTypeAssertion) }
