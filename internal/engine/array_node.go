package engine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/474420502/xjson/internal/core"
)

type arrayNode struct {
	baseNode
	value []core.Node
}

func NewArrayNode(value []core.Node, path string, funcs *map[string]func(core.Node) core.Node) core.Node {
	if funcs == nil {
		funcs = &map[string]func(core.Node) core.Node{} // Initialize if nil (for root)
	}
	return &arrayNode{value: value, baseNode: baseNode{path: path, funcs: funcs}}
}

func (n *arrayNode) Type() core.NodeType { return core.ArrayNode }
func (n *arrayNode) Get(key string) core.Node {
	return NewInvalidNode(n.path+"."+key, ErrTypeAssertion)
}
func (n *arrayNode) Index(i int) core.Node {
	if n.err != nil {
		return n
	}
	if i >= 0 && i < len(n.value) {
		return n.value[i]
	}
	return NewInvalidNode(n.path+fmt.Sprintf("[%d]", i), ErrIndexOutOfBounds)
}
func (n *arrayNode) Query(path string) core.Node {
	ops, err := ParseQuery(path)
	if err != nil {
		return NewInvalidNode(n.path, err)
	}
	return EvaluateQuery(n, ops)
}
func (n *arrayNode) ForEach(iterator func(interface{}, core.Node)) {
	if n.err != nil {
		return
	}
	for i, v := range n.value {
		iterator(i, v)
	}
}
func (n *arrayNode) Len() int {
	if n.err != nil {
		return 0
	}
	return len(n.value)
}
func (n *arrayNode) String() string {
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
func (n *arrayNode) MustString() string  { panic(ErrTypeAssertion) }
func (n *arrayNode) Float() float64      { return 0 }
func (n *arrayNode) MustFloat() float64  { panic(ErrTypeAssertion) }
func (n *arrayNode) Int() int64          { return 0 }
func (n *arrayNode) MustInt() int64      { panic(ErrTypeAssertion) }
func (n *arrayNode) Bool() bool          { return false }
func (n *arrayNode) MustBool() bool      { panic(ErrTypeAssertion) }
func (n *arrayNode) Time() time.Time     { return time.Time{} }
func (n *arrayNode) MustTime() time.Time { panic(ErrTypeAssertion) }
func (n *arrayNode) Array() []core.Node {
	if n.err != nil {
		return nil
	}
	return n.value
}
func (n *arrayNode) MustArray() []core.Node {
	if n.err != nil {
		panic(n.err)
	}
	return n.value
}
func (n *arrayNode) Interface() interface{} {
	if n.err != nil {
		return nil
	}
	s := make([]interface{}, len(n.value))
	for i, v := range n.value {
		s[i] = v.Interface()
	}
	return s
}

// Deprecated: Use RegisterFunc and CallFunc instead
func (n *arrayNode) Func(name string, fn func(core.Node) core.Node) core.Node {
	if n.err != nil {
		return n
	}
	if n.funcs == nil {
		n.funcs = &map[string]func(core.Node) core.Node{}
	}
	(*n.funcs)[name] = fn
	return n
}

func (n *arrayNode) RegisterFunc(name string, fn core.UnaryPathFunc) core.Node {
	if n.err != nil {
		return n
	}
	if n.funcs == nil {
		n.funcs = &map[string]func(core.Node) core.Node{}
	}
	(*n.funcs)[name] = fn
	return n
}

func (n *arrayNode) Apply(fn core.PathFunc) core.Node {
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

func (n *arrayNode) CallFunc(name string) core.Node {
	if n.err != nil {
		return n
	}
	if fn, ok := (*n.funcs)[name]; ok {
		// First attempt: apply function to whole array
		res := fn(n)
		if res != nil {
			// If the function returns an ArrayNode (e.g., Filter/Map semantics) or InvalidNode, use it directly.
			if res.Type() == core.ArrayNode || res.Type() == core.InvalidNode {
				return res
			}
		}
		// Fallback: treat the function as element-wise transformation (legacy behavior expected by engine tests)
		var results []core.Node
		for _, child := range n.value {
			results = append(results, fn(child))
		}
		return NewArrayNode(results, n.path, n.funcs)
	}
	return NewInvalidNode(n.path, fmt.Errorf("function %s not found", name))
}
func (n *arrayNode) RemoveFunc(name string) core.Node {
	if n.err != nil {
		return n
	}
	delete(*n.funcs, name)
	return n
}

// New methods for arrayNode
func (n *arrayNode) Filter(fn core.PredicateFunc) core.Node {
	if n.err != nil {
		return n
	}
	
	// Handle nil function case
	if fn == nil {
		return NewInvalidNode(n.path, ErrTypeAssertion)
	}
	
	filteredNodes := make([]core.Node, 0, len(n.value))
	for _, child := range n.value {
		if fn(child) {
			filteredNodes = append(filteredNodes, child)
		}
	}
	return NewArrayNode(filteredNodes, n.path, n.funcs)
}

func (n *arrayNode) Map(fn core.TransformFunc) core.Node {
	if n.err != nil {
		return n
	}
	
	// Handle nil function case
	if fn == nil {
		return NewInvalidNode(n.path, ErrTypeAssertion)
	}
	
	mappedValues := make([]core.Node, 0, len(n.value))
	for _, child := range n.value {
		mappedValue := fn(child)
		newNode, err := NewNodeFromInterface(mappedValue, n.path, n.funcs)
		if err != nil {
			return NewInvalidNode(n.path, err)
		}
		mappedValues = append(mappedValues, newNode)
	}
	return NewArrayNode(mappedValues, n.path, n.funcs)
}

func (n *arrayNode) Set(key string, value interface{}) core.Node {
	if n.err != nil {
		return n
	}

	for _, child := range n.value {
		// If a child node itself is invalid, we should not proceed.
		if !child.IsValid() {
			n.setError(child.Error())
			return n
		}

		if child.Type() != core.ObjectNode {
			n.setError(ErrTypeAssertion) // Set error if any element is not an object.
			return n
		}
	}

	// If all elements are valid object nodes, proceed to set values.
	for _, child := range n.value {
		child.Set(key, value)
		// After setting, if a child has an error, propagate it.
		if !child.IsValid() {
			n.setError(child.Error())
			return n
		}
	}

	return n
}

func (n *arrayNode) Append(value interface{}) core.Node {
	if n.err != nil {
		return n
	}
	newNode, err := NewNodeFromInterface(value, n.path+fmt.Sprintf("[%d]", len(n.value)), n.funcs)
	if err != nil {
		n.setError(err)
		return n
	}
	n.value = append(n.value, newNode)
	return n
}

func (n *arrayNode) Raw() string {
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

func (n *arrayNode) RawFloat() (float64, bool) {
	return 0, false
}

func (n *arrayNode) RawString() (string, bool) {
	return "", false
}

func (n *arrayNode) Strings() []string {
	if n.err != nil {
		return nil
	}
	var s []string
	for _, node := range n.value {
		if node.Type() == core.StringNode {
			s = append(s, node.String())
		} else {
			// If not all elements are strings, return nil or an error
			n.setError(fmt.Errorf("array contains non-string elements"))
			return nil
		}
	}
	return s
}

func (n *arrayNode) Contains(value string) bool {
	if n.err != nil {
		return false
	}
	for _, child := range n.value {
		if child.Type() == core.StringNode && child.String() == value {
			return true
		}
	}
	return false
}

func (n *arrayNode) AsMap() map[string]core.Node     { return nil }
func (n *arrayNode) MustAsMap() map[string]core.Node { panic(ErrTypeAssertion) }