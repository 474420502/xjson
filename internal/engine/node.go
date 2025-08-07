package engine

import (
	"bytes"
	"encoding/json" // Added for JSON marshaling
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// baseNode serves as the base for all concrete node types, implementing common
// functionalities like error handling and path management.
type baseNode struct {
	err   error
	path  string
	raw   *string
	funcs *map[string]func(Node) Node // Changed to pointer
}

func (n *baseNode) IsValid() bool {
	return n.err == nil
}

func (n *baseNode) Error() error {
	return n.err
}

func (n *baseNode) Path() string {
	return n.path
}

func (n *baseNode) Raw() string {
	if n.raw != nil {
		return *n.raw
	}
	return ""
}

func (n *baseNode) setError(err error) {
	if n.err == nil {
		n.err = err
	}
}

func (n *baseNode) GetFuncs() *map[string]func(Node) Node {
	return n.funcs
}

// invalidNode represents a node that is the result of a failed operation.
type invalidNode struct {
	baseNode
}

func NewInvalidNode(path string, err error) Node {
	node := &invalidNode{}
	node.path = path
	node.setError(err)
	// No funcs map for invalid nodes, as they don't participate in function calls
	return node
}

func (n *invalidNode) Type() NodeType                            { return InvalidNode }
func (n *invalidNode) Get(key string) Node                       { return n }
func (n *invalidNode) Index(i int) Node                          { return n }
func (n *invalidNode) Query(path string) Node                    { return n }
func (n *invalidNode) ForEach(iterator func(interface{}, Node))  {}
func (n *invalidNode) Len() int                                  { return 0 }
func (n *invalidNode) String() string                            { return "" }
func (n *invalidNode) MustString() string                        { panic(n.err) }
func (n *invalidNode) Float() float64                            { return 0 }
func (n *invalidNode) MustFloat() float64                        { panic(n.err) }
func (n *invalidNode) Int() int64                                { return 0 }
func (n *invalidNode) MustInt() int64                            { panic(n.err) }
func (n *invalidNode) Bool() bool                                { return false }
func (n *invalidNode) MustBool() bool                            { panic(n.err) }
func (n *invalidNode) Time() time.Time                           { return time.Time{} }
func (n *invalidNode) MustTime() time.Time                       { panic(n.err) }
func (n *invalidNode) Array() []Node                             { return nil }
func (n *invalidNode) MustArray() []Node                         { panic(n.err) }
func (n *invalidNode) AsMap() map[string]Node                    { return nil }
func (n *invalidNode) MustAsMap() map[string]Node                { panic(n.err) }
func (n *invalidNode) Interface() interface{}                    { return nil }
func (n *invalidNode) Func(name string, fn func(Node) Node) Node { return n }
func (n *invalidNode) CallFunc(name string) Node                 { return n }
func (n *invalidNode) RemoveFunc(name string) Node               { return n }

func (n *invalidNode) Filter(fn func(Node) bool) Node         { return n }
func (n *invalidNode) Map(fn func(Node) interface{}) Node     { return n }
func (n *invalidNode) Set(key string, value interface{}) Node { return n }
func (n *invalidNode) Append(value interface{}) Node          { return n }
func (n *invalidNode) RawFloat() (float64, bool)              { return 0, false }
func (n *invalidNode) RawString() (string, bool)              { return "", false }
func (n *invalidNode) Contains(value string) bool             { return false }
func (n *invalidNode) Strings() []string                      { return nil }

// Below are the concrete implementations for each JSON type.

type objectNode struct {
	baseNode
	value map[string]Node
}

func NewObjectNode(value map[string]Node, path string, funcs *map[string]func(Node) Node) Node {
	if funcs == nil {
		funcs = &map[string]func(Node) Node{} // Initialize if nil (for root)
	}
	return &objectNode{value: value, baseNode: baseNode{path: path, funcs: funcs}}
}

func (n *objectNode) Type() NodeType { return ObjectNode }
func (n *objectNode) Get(key string) Node {
	if n.err != nil {
		fmt.Printf("objectNode.Get: Node has error: %v\n", n.err)
		return n
	}
	fmt.Printf("objectNode.Get: Attempting to get key '%s' from path '%s'. Value map: %+v\n", key, n.path, n.value)
	if child, ok := n.value[key]; ok {
		fmt.Printf("objectNode.Get: Found child for key '%s'. Type: %v, Path: %s\n", key, child.Type(), child.Path())
		return child
	}
	fmt.Printf("objectNode.Get: Key '%s' not found in path '%s'.\n", key, n.path)
	return NewInvalidNode(n.path+"."+key, ErrNotFound)
}
func (n *objectNode) Index(i int) Node {
	return NewInvalidNode(n.path+fmt.Sprintf("[%d]", i), ErrTypeAssertion)
}
func (n *objectNode) Query(path string) Node {
	ops, err := ParseQuery(path)
	if err != nil {
		return NewInvalidNode(n.path, err)
	}
	return EvaluateQuery(n, ops)
}
func (n *objectNode) ForEach(iterator func(interface{}, Node)) {
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
func (n *objectNode) MustString() string  { panic(ErrTypeAssertion) }
func (n *objectNode) Float() float64      { return 0 }
func (n *objectNode) MustFloat() float64  { panic(ErrTypeAssertion) }
func (n *objectNode) Int() int64          { return 0 }
func (n *objectNode) MustInt() int64      { panic(ErrTypeAssertion) }
func (n *objectNode) Bool() bool          { return false }
func (n *objectNode) MustBool() bool      { panic(ErrTypeAssertion) }
func (n *objectNode) Time() time.Time     { return time.Time{} }
func (n *objectNode) MustTime() time.Time { panic(ErrTypeAssertion) }
func (n *objectNode) Array() []Node       { return nil }
func (n *objectNode) MustArray() []Node   { panic(ErrTypeAssertion) }
func (n *objectNode) AsMap() map[string]Node {
	if n.err != nil {
		return nil
	}
	return n.value
}
func (n *objectNode) MustAsMap() map[string]Node {
	if n.err != nil {
		panic(n.err)
	}
	return n.value
}
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
func (n *objectNode) Func(name string, fn func(Node) Node) Node {
	if n.err != nil {
		return n
	}
	(*n.funcs)[name] = fn
	return n
}
func (n *objectNode) CallFunc(name string) Node {
	if n.err != nil {
		return n
	}
	if fn, ok := (*n.funcs)[name]; ok {
		return fn(n)
	}
	return NewInvalidNode(n.path, fmt.Errorf("function %s not found", name))
}
func (n *objectNode) RemoveFunc(name string) Node {
	if n.err != nil {
		return n
	}
	delete(*n.funcs, name)
	return n
}

// New methods for objectNode
func (n *objectNode) Filter(fn func(Node) bool) Node {
	if n.err != nil {
		return n
	}
	// For an object, filter applies to its values.
	// Returns a new array node containing filtered values.
	filteredNodes := make([]Node, 0, len(n.value))
	for _, child := range n.value {
		if fn(child) {
			filteredNodes = append(filteredNodes, child)
		}
	}
	return NewArrayNode(filteredNodes, n.path, n.funcs)
}

func (n *objectNode) Map(fn func(Node) interface{}) Node {
	if n.err != nil {
		return n
	}
	// For an object, map applies to its values.
	// Returns a new array node containing mapped values.
	mappedValues := make([]Node, 0, len(n.value))
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

func (n *objectNode) Set(key string, value interface{}) Node {
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

func (n *objectNode) Append(value interface{}) Node {
	n.setError(ErrTypeAssertion) // Cannot append to an object
	return n
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

type arrayNode struct {
	baseNode
	value []Node
}

func NewArrayNode(value []Node, path string, funcs *map[string]func(Node) Node) Node {
	if funcs == nil {
		funcs = &map[string]func(Node) Node{} // Initialize if nil (for root)
	}
	return &arrayNode{value: value, baseNode: baseNode{path: path, funcs: funcs}}
}

func (n *arrayNode) Type() NodeType { return ArrayNode }
func (n *arrayNode) Get(key string) Node {
	return NewInvalidNode(n.path+"."+key, ErrTypeAssertion)
}
func (n *arrayNode) Index(i int) Node {
	if n.err != nil {
		return n
	}
	if i >= 0 && i < len(n.value) {
		return n.value[i]
	}
	return NewInvalidNode(n.path+fmt.Sprintf("[%d]", i), ErrIndexOutOfBounds)
}
func (n *arrayNode) Query(path string) Node {
	ops, err := ParseQuery(path)
	if err != nil {
		return NewInvalidNode(n.path, err)
	}
	return EvaluateQuery(n, ops)
}
func (n *arrayNode) ForEach(iterator func(interface{}, Node)) {
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
func (n *arrayNode) Array() []Node {
	if n.err != nil {
		return nil
	}
	return n.value
}
func (n *arrayNode) MustArray() []Node {
	if n.err != nil {
		panic(n.err)
	}
	return n.value
}
func (n *arrayNode) AsMap() map[string]Node     { return nil }
func (n *arrayNode) MustAsMap() map[string]Node { panic(ErrTypeAssertion) }
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
func (n *arrayNode) Func(name string, fn func(Node) Node) Node {
	if n.err != nil {
		return n
	}
	(*n.funcs)[name] = fn
	return n
}
func (n *arrayNode) CallFunc(name string) Node {
	if n.err != nil {
		return n
	}
	if fn, ok := (*n.funcs)[name]; ok {
		return fn(n)
	}
	return NewInvalidNode(n.path, fmt.Errorf("function %s not found", name))
}
func (n *arrayNode) RemoveFunc(name string) Node {
	if n.err != nil {
		return n
	}
	delete(*n.funcs, name)
	return n
}

// New methods for arrayNode
func (n *arrayNode) Filter(fn func(Node) bool) Node {
	if n.err != nil {
		return n
	}
	filteredNodes := make([]Node, 0, len(n.value))
	for _, child := range n.value {
		if fn(child) {
			filteredNodes = append(filteredNodes, child)
		}
	}
	return NewArrayNode(filteredNodes, n.path, n.funcs)
}

func (n *arrayNode) Map(fn func(Node) interface{}) Node {
	if n.err != nil {
		return n
	}
	mappedValues := make([]Node, 0, len(n.value))
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

func (n *arrayNode) Set(key string, value interface{}) Node {
	if n.err != nil {
		return n
	}
	// If the array contains objects, attempt to set the key on each object.
	// If it contains other types, it's an error.
	allObjects := true
	for _, child := range n.value {
		if child.Type() != ObjectNode {
			allObjects = false
			break
		}
	}

	if allObjects {
		for _, child := range n.value {
			child.Set(key, value) // Recursively call Set on child object
			if child.Error() != nil {
				n.setError(child.Error()) // Propagate error
				return n
			}
		}
		return n
	}

	n.setError(ErrTypeAssertion) // Cannot set key on an array if it contains non-object elements
	return n
}

func (n *arrayNode) Append(value interface{}) Node {
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
		if node.Type() == StringNode {
			s = append(s, node.String())
		} else {
			// If not all elements are strings, return nil or an error
			n.setError(errors.New("array contains non-string elements"))
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
		if child.Type() == StringNode && child.String() == value {
			return true
		}
	}
	return false
}

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
	return NewInvalidNode(n.path+fmt.Sprintf("[%d]", i), ErrTypeAssertion)
}
func (n *stringNode) Query(path string) Node {
	return NewInvalidNode(n.path, errors.New("query not implemented"))
}
func (n *stringNode) ForEach(iterator func(interface{}, Node)) {}
func (n *stringNode) Len() int                                 { return len(n.value) }
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
func (n *stringNode) Array() []Node              { return nil }
func (n *stringNode) MustArray() []Node          { panic(ErrTypeAssertion) }
func (n *stringNode) AsMap() map[string]Node     { return nil }
func (n *stringNode) MustAsMap() map[string]Node { panic(ErrTypeAssertion) }
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
	return NewInvalidNode(n.path, fmt.Errorf("function %s not found", name))
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
	n.setError(ErrTypeAssertion) // Cannot filter a string node
	return n
}

func (n *stringNode) Map(fn func(Node) interface{}) Node {
	n.setError(ErrTypeAssertion) // Cannot map a string node
	return n
}

func (n *stringNode) Set(key string, value interface{}) Node {
	n.setError(ErrTypeAssertion) // Cannot set key on a string
	return n
}

func (n *stringNode) Append(value interface{}) Node {
	n.setError(ErrTypeAssertion) // Cannot append to a string
	return n
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

type numberNode struct {
	baseNode
	value float64
}

func NewNumberNode(value float64, path string, funcs *map[string]func(Node) Node) Node {
	if funcs == nil {
		funcs = &map[string]func(Node) Node{} // Initialize if nil (for root)
	}
	return &numberNode{value: value, baseNode: baseNode{path: path, funcs: funcs}}
}

func (n *numberNode) Type() NodeType { return NumberNode }
func (n *numberNode) Get(key string) Node {
	return NewInvalidNode(n.path+"."+key, ErrTypeAssertion)
}
func (n *numberNode) Index(i int) Node {
	return NewInvalidNode(n.path+fmt.Sprintf("[%d]", i), ErrTypeAssertion)
}
func (n *numberNode) Query(path string) Node {
	return NewInvalidNode(n.path, errors.New("query not implemented"))
}
func (n *numberNode) ForEach(iterator func(interface{}, Node)) {}
func (n *numberNode) Len() int                                 { return 0 }
func (n *numberNode) String() string {
	if n.err != nil {
		return ""
	}
	return strconv.FormatFloat(n.value, 'f', -1, 64)
}
func (n *numberNode) MustString() string {
	if n.err != nil {
		panic(n.err)
	}
	return strconv.FormatFloat(n.value, 'f', -1, 64)
}
func (n *numberNode) Float() float64 {
	if n.err != nil {
		return 0
	}
	return n.value
}
func (n *numberNode) MustFloat() float64 {
	if n.err != nil {
		panic(n.err)
	}
	return n.value
}
func (n *numberNode) Int() int64 {
	if n.err != nil {
		return 0
	}
	return int64(n.value)
}
func (n *numberNode) MustInt() int64 {
	if n.err != nil {
		panic(n.err)
	}
	return int64(n.value)
}
func (n *numberNode) Bool() bool                 { return false }
func (n *numberNode) MustBool() bool             { panic(ErrTypeAssertion) }
func (n *numberNode) Time() time.Time            { return time.Time{} }
func (n *numberNode) MustTime() time.Time        { panic(ErrTypeAssertion) }
func (n *numberNode) Array() []Node              { return nil }
func (n *numberNode) MustArray() []Node          { panic(ErrTypeAssertion) }
func (n *numberNode) AsMap() map[string]Node     { return nil }
func (n *numberNode) MustAsMap() map[string]Node { panic(ErrTypeAssertion) }
func (n *numberNode) Interface() interface{} {
	if n.err != nil {
		return nil
	}
	return n.value
}
func (n *numberNode) Func(name string, fn func(Node) Node) Node {
	if n.err != nil {
		return n
	}
	(*n.funcs)[name] = fn
	return n
}
func (n *numberNode) CallFunc(name string) Node {
	if n.err != nil {
		return n
	}
	if fn, ok := (*n.funcs)[name]; ok {
		return fn(n)
	}
	return NewInvalidNode(n.path, fmt.Errorf("function %s not found", name))
}
func (n *numberNode) RemoveFunc(name string) Node {
	if n.err != nil {
		return n
	}
	delete(*n.funcs, name)
	return n
}

// New methods for numberNode
func (n *numberNode) Filter(fn func(Node) bool) Node {
	n.setError(ErrTypeAssertion) // Cannot filter a number node
	return n
}

func (n *numberNode) Map(fn func(Node) interface{}) Node {
	n.setError(ErrTypeAssertion) // Cannot map a number node
	return n
}

func (n *numberNode) Set(key string, value interface{}) Node {
	n.setError(ErrTypeAssertion) // Cannot set key on a number
	return n
}

func (n *numberNode) Append(value interface{}) Node {
	n.setError(ErrTypeAssertion) // Cannot append to a number
	return n
}

func (n *numberNode) RawFloat() (float64, bool) {
	if n.err != nil {
		return 0, false
	}
	return n.value, true
}

func (n *numberNode) RawString() (string, bool) {
	return "", false
}

func (n *numberNode) Strings() []string {
	return nil
}

func (n *numberNode) Contains(value string) bool {
	return false
}

type boolNode struct {
	baseNode
	value bool
}

func NewBoolNode(value bool, path string, funcs *map[string]func(Node) Node) Node {
	if funcs == nil {
		funcs = &map[string]func(Node) Node{} // Initialize if nil (for root)
	}
	return &boolNode{value: value, baseNode: baseNode{path: path, funcs: funcs}}
}

func (n *boolNode) Type() NodeType { return BoolNode }
func (n *boolNode) Get(key string) Node {
	return NewInvalidNode(n.path+"."+key, ErrTypeAssertion)
}
func (n *boolNode) Index(i int) Node {
	return NewInvalidNode(n.path+fmt.Sprintf("[%d]", i), ErrTypeAssertion)
}
func (n *boolNode) Query(path string) Node {
	return NewInvalidNode(n.path, errors.New("query not implemented"))
}
func (n *boolNode) ForEach(iterator func(interface{}, Node)) {}
func (n *boolNode) Len() int                                 { return 0 }
func (n *boolNode) String() string                           { return strconv.FormatBool(n.value) }
func (n *boolNode) MustString() string                       { return strconv.FormatBool(n.value) }
func (n *boolNode) Float() float64                           { return 0 }
func (n *boolNode) MustFloat() float64                       { panic(ErrTypeAssertion) }
func (n *boolNode) Int() int64                               { return 0 }
func (n *boolNode) MustInt() int64                           { panic(ErrTypeAssertion) }
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
func (n *boolNode) Time() time.Time            { return time.Time{} }
func (n *boolNode) MustTime() time.Time        { panic(ErrTypeAssertion) }
func (n *boolNode) Array() []Node              { return nil }
func (n *boolNode) MustArray() []Node          { panic(ErrTypeAssertion) }
func (n *boolNode) AsMap() map[string]Node     { return nil }
func (n *boolNode) MustAsMap() map[string]Node { panic(ErrTypeAssertion) }
func (n *boolNode) Interface() interface{} {
	if n.err != nil {
		return nil
	}
	return n.value
}
func (n *boolNode) Func(name string, fn func(Node) Node) Node {
	if n.err != nil {
		return n
	}
	(*n.funcs)[name] = fn
	return n
}
func (n *boolNode) CallFunc(name string) Node {
	if n.err != nil {
		return n
	}
	if fn, ok := (*n.funcs)[name]; ok {
		return fn(n)
	}
	return NewInvalidNode(n.path, fmt.Errorf("function %s not found", name))
}
func (n *boolNode) RemoveFunc(name string) Node {
	if n.err != nil {
		return n
	}
	delete(*n.funcs, name)
	return n
}

// New methods for boolNode
func (n *boolNode) Filter(fn func(Node) bool) Node {
	n.setError(ErrTypeAssertion) // Cannot filter a bool node
	return n
}

func (n *boolNode) Map(fn func(Node) interface{}) Node {
	n.setError(ErrTypeAssertion) // Cannot map a bool node
	return n
}

func (n *boolNode) Set(key string, value interface{}) Node {
	n.setError(ErrTypeAssertion) // Cannot set key on a bool
	return n
}

func (n *boolNode) Append(value interface{}) Node {
	n.setError(ErrTypeAssertion) // Cannot append to a bool
	return n
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
	return NewInvalidNode(n.path+fmt.Sprintf("[%d]", i), ErrTypeAssertion)
}
func (n *nullNode) Query(path string) Node {
	return NewInvalidNode(n.path, errors.New("query not implemented"))
}
func (n *nullNode) ForEach(iterator func(interface{}, Node)) {}
func (n *nullNode) Len() int                                 { return 0 }
func (n *nullNode) String() string                           { return "null" }
func (n *nullNode) MustString() string                       { panic(ErrTypeAssertion) }
func (n *nullNode) Float() float64                           { return 0 }
func (n *nullNode) MustFloat() float64                       { panic(ErrTypeAssertion) }
func (n *nullNode) Int() int64                               { return 0 }
func (n *nullNode) MustInt() int64                           { panic(ErrTypeAssertion) }
func (n *nullNode) Bool() bool                               { return false }
func (n *nullNode) MustBool() bool                           { panic(ErrTypeAssertion) }
func (n *nullNode) Time() time.Time                          { return time.Time{} }
func (n *nullNode) MustTime() time.Time                      { panic(ErrTypeAssertion) }
func (n *nullNode) Array() []Node                            { return nil }
func (n *nullNode) MustArray() []Node                        { panic(ErrTypeAssertion) }
func (n *nullNode) AsMap() map[string]Node                   { return nil }
func (n *nullNode) MustAsMap() map[string]Node               { panic(ErrTypeAssertion) }
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
	return NewInvalidNode(n.path, fmt.Errorf("function %s not found", name))
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
	n.setError(ErrTypeAssertion) // Cannot filter a null node
	return n
}

func (n *nullNode) Map(fn func(Node) interface{}) Node {
	n.setError(ErrTypeAssertion) // Cannot map a null node
	return n
}

func (n *nullNode) Set(key string, value interface{}) Node {
	n.setError(ErrTypeAssertion) // Cannot set key on a null
	return n
}

func (n *nullNode) Append(value interface{}) Node {
	n.setError(ErrTypeAssertion) // Cannot append to a null
	return n
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

// NewNodeFromInterface converts a Go interface{} value into an xjson.Node.
// This is useful for creating new nodes from arbitrary Go types, especially
// when used with the Map function.
func NewNodeFromInterface(value interface{}, path string, funcs *map[string]func(Node) Node) (Node, error) {
	if funcs == nil {
		funcs = &map[string]func(Node) Node{} // Initialize if nil (for root)
	}
	switch v := value.(type) {
	case map[string]interface{}:
		obj := make(map[string]Node, len(v))
		for key, val := range v {
			node, err := NewNodeFromInterface(val, path+"."+key, funcs)
			if err != nil {
				return nil, err
			}
			obj[key] = node
		}
		return NewObjectNode(obj, path, funcs), nil
	case []interface{}:
		arr := make([]Node, len(v))
		for i, val := range v {
			node, err := NewNodeFromInterface(val, path+fmt.Sprintf("[%d]", i), funcs)
			if err != nil {
				return nil, err
			}
			arr[i] = node
		}
		return NewArrayNode(arr, path, funcs), nil
	case string:
		return NewStringNode(v, path, funcs), nil
	case float64:
		return NewNumberNode(v, path, funcs), nil
	case int:
		return NewNumberNode(float64(v), path, funcs), nil
	case int64:
		return NewNumberNode(float64(v), path, funcs), nil
	case bool:
		return NewBoolNode(v, path, funcs), nil
	case nil:
		return NewNullNode(path, funcs), nil
	default:
		return nil, fmt.Errorf("unsupported type for node creation: %T", v)
	}
}
