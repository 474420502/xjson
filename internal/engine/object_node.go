package engine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/474420502/xjson/internal/core"
)

type objectNode struct {
	baseNode
	value map[string]core.Node
}

func NewObjectNode(value map[string]core.Node, path string, funcs *map[string]func(core.Node) core.Node) core.Node {
	if funcs == nil {
		funcs = &map[string]func(core.Node) core.Node{} // Initialize if nil (for root)
	}
	return &objectNode{value: value, baseNode: baseNode{path: path, funcs: funcs}}
}

func (n *objectNode) Type() core.NodeType { return core.ObjectNode }
func (n *objectNode) Get(key string) core.Node {
	if n.err != nil {
		return n
	}
	if child, ok := n.value[key]; ok {
		return child
	}
	return NewInvalidNode(n.path+"."+key, ErrNotFound)
}
func (n *objectNode) Index(i int) core.Node {
	return NewInvalidNode(n.path+fmt.Sprintf("[%d]", i), ErrTypeAssertion)
}
func (n *objectNode) Query(path string) core.Node {
	ops, err := ParseQuery(path)
	if err != nil {
		return NewInvalidNode(n.path, err)
	}
	return EvaluateQuery(n, ops)
}
func (n *objectNode) ForEach(iterator func(interface{}, core.Node)) {
	if n.err != nil {
		return
	}
	for k, v := range n.value {
		iterator(k, v)
	}
}
func (n *objectNode) Len() int {
	if n.err != nil {
		return 0
	}
	return len(n.value)
}
func (n *objectNode) String() string {
	if n.err != nil {
		return ""
	}
	data, err := json.Marshal(n.Interface())
	if err != nil {
		n.setError(err)
		return ""
	}
	buf := new(bytes.Buffer)
	err = json.Compact(buf, data)
	if err != nil {
		n.setError(err)
		return ""
	}
	return buf.String()
}
func (n *objectNode) MustString() string {
	if n.err != nil {
		panic(n.err)
	}
	data, err := json.Marshal(n.Interface())
	if err != nil {
		panic(err) // Panic if marshaling fails
	}
	buf := new(bytes.Buffer)
	err = json.Compact(buf, data)
	if err != nil {
		panic(err) // Panic if compacting fails
	}
	return buf.String()
}
func (n *objectNode) Float() float64         { return 0 }
func (n *objectNode) MustFloat() float64     { panic(ErrTypeAssertion) }
func (n *objectNode) Int() int64             { return 0 }
func (n *objectNode) MustInt() int64         { panic(ErrTypeAssertion) }
func (n *objectNode) Bool() bool             { return false }
func (n *objectNode) MustBool() bool         { panic(ErrTypeAssertion) }
func (n *objectNode) Time() time.Time        { return time.Time{} }
func (n *objectNode) MustTime() time.Time    { panic(ErrTypeAssertion) }
func (n *objectNode) Array() []core.Node     { return nil }
func (n *objectNode) MustArray() []core.Node { panic(ErrTypeAssertion) }
func (n *objectNode) Interface() interface{} {
	if n.err != nil {
		return nil
	}
	m := make(map[string]interface{}, len(n.value))
	for k, v := range n.value {
		m[k] = v.Interface()
	}
	return m
}

// Deprecated: Use RegisterFunc and CallFunc instead
func (n *objectNode) Func(name string, fn func(core.Node) core.Node) core.Node {
	if n.err != nil {
		return n
	}
	(*n.funcs)[name] = fn
	return n
}

func (n *objectNode) RegisterFunc(name string, fn core.UnaryPathFunc) core.Node {
	if n.err != nil {
		return n
	}
	if n.funcs == nil {
		n.funcs = &map[string]func(core.Node) core.Node{}
	}
	(*n.funcs)[name] = fn
	return n
}

func (n *objectNode) Apply(fn core.PathFunc) core.Node {
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
		return n.baseNode.Apply(f)
	}
}

func (n *objectNode) CallFunc(name string) core.Node {
	if n.err != nil {
		return n
	}
	if fn, ok := (*n.funcs)[name]; ok {
		return fn(n)
	}
	return NewInvalidNode(n.path, fmt.Errorf("function %s not found", name))
}
func (n *objectNode) RemoveFunc(name string) core.Node {
	if n.err != nil {
		return n
	}
	delete(*n.funcs, name)
	return n
}

// New methods for objectNode
func (n *objectNode) Filter(fn core.PredicateFunc) core.Node {
	if n.err != nil {
		return n
	}

	// Handle nil function case
	if fn == nil {
		return NewInvalidNode(n.path, ErrTypeAssertion)
	}

	// For an object, filter applies to its values.
	// Returns a new array node containing filtered values.
	filteredNodes := make([]core.Node, 0, len(n.value))
	for _, child := range n.value {
		if fn(child) {
			filteredNodes = append(filteredNodes, child)
		}
	}
	return NewArrayNode(filteredNodes, n.path, n.funcs)
}

func (n *objectNode) Map(fn core.TransformFunc) core.Node {
	if n.err != nil {
		return n
	}

	// Handle nil function case
	if fn == nil {
		return NewInvalidNode(n.path, ErrTypeAssertion)
	}

	// For an object, map applies to its values.
	// Returns a new array node containing mapped values.
	mappedValues := make([]core.Node, 0, len(n.value))
	for _, child := range n.value {
		mappedValue := fn(child)
		// Convert mappedValue back to Node
		newNode, err := NewNodeFromInterface(mappedValue, n.path, n.funcs)
		if err != nil {
			return NewInvalidNode(n.path, err)
		}
		mappedValues = append(mappedValues, newNode)
	}
	return NewArrayNode(mappedValues, n.path, n.funcs)
}

func (n *objectNode) Set(key string, value interface{}) core.Node {
	if n.err != nil {
		return n
	}
	newNode, err := NewNodeFromInterface(value, n.path+"."+key, n.funcs)
	if err != nil {
		n.setError(err)
		return n
	}
	n.value[key] = newNode
	return n
}

func (n *objectNode) Append(value interface{}) core.Node {
	n.setError(ErrTypeAssertion) // Cannot append to an object
	return n
}

func (n *objectNode) Raw() string {
	if n.raw != nil {
		return *n.raw
	}
	if n.err != nil {
		return ""
	}
	data, err := json.Marshal(n.Interface())
	if err != nil {
		n.setError(err)
		return ""
	}
	return string(data)
}

func (n *objectNode) RawFloat() (float64, bool) {
	return 0, false
}

func (n *objectNode) RawString() (string, bool) {
	return "", false
}

func (n *objectNode) Strings() []string {
	return nil
}

func (n *objectNode) Contains(value string) bool {
	return false
}

func (n *objectNode) AsMap() map[string]core.Node {
	if n.err != nil {
		return nil
	}
	return n.value
}

func (n *objectNode) MustAsMap() map[string]core.Node {
	if n.err != nil {
		panic(n.err)
	}
	return n.value
}
