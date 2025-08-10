package engine

import (
	"fmt"
	"time"

	"github.com/474420502/xjson/internal/core"
)

// baseNode contains the common fields and methods for all node types.
// It implements the parts of the core.Node interface that are common to all types.
type baseNode struct {
	raw    []byte                         // The raw byte slice of the entire JSON data
	start  int                            // The start index of the node's data in the raw slice
	end    int                            // The end index of the node's data in the raw slice
	path   string                         // The JSON path to this node
	parent core.Node                      // Pointer to the parent node
	err    error                          // Stores the first error that occurred in a chain of operations
	funcs  *map[string]core.UnaryPathFunc // Registered functions
}

// newBaseNode creates a new baseNode.
func newBaseNode(raw []byte, start, end int, parent core.Node, funcs *map[string]core.UnaryPathFunc) baseNode {
	return baseNode{
		raw:    raw,
		start:  start,
		end:    end,
		parent: parent,
		err:    nil, // Initially no error
		funcs:  funcs,
	}
}

// Error returns the first error that occurred in a chain of operations.
func (n *baseNode) Error() error {
	return n.err
}

// setError sets an error if no error has been set before.
func (n *baseNode) setError(err error) {
	if n.err == nil {
		n.err = err
	}
}

// Raw returns the raw string representation of the node.
func (n *baseNode) Raw() string {
	if n.start >= n.end {
		return ""
	}
	return string(n.raw[n.start:n.end])
}

// IsValid checks if the node is valid. An invalid node usually results from a failed query.
func (n *baseNode) IsValid() bool {
	return n.err == nil
}

// GetFuncs returns the map of registered functions.
func (n *baseNode) GetFuncs() *map[string]core.UnaryPathFunc {
	return n.funcs
}

// Path returns the JSON path of the current node.
func (n *baseNode) Path() string {
	return n.path
}

// GetParent returns the parent node.
func (n *baseNode) GetParent() core.Node {
	return n.parent
}

// SetParent sets the parent of the current node.
func (n *baseNode) SetParent(parent core.Node) {
	n.parent = parent
}

// SetValue is a placeholder for the SetValue method.
func (n *baseNode) SetValue(value interface{}) core.Node {
	n.setError(core.ErrTypeAssertion)
	return newInvalidNode(n.err)
}

// Stub implementations for methods to be implemented by concrete node types.
// These will panic if called on a baseNode directly.

func (n *baseNode) Type() core.NodeType {
	return core.Invalid
}

func (n *baseNode) Query(path string) core.Node {
	if n.err != nil {
		return newInvalidNode(n.err)
	}
	// Allow querying from any node (including leaf nodes) to enable parent navigation and more.
	return applySimpleQuery(n, path)
}

func (n *baseNode) Get(key string) core.Node {
	n.setError(fmt.Errorf("not an object node"))
	return newInvalidNode(n.err)
}

func (n *baseNode) Index(i int) core.Node {
	n.setError(fmt.Errorf("not an array node"))
	return newInvalidNode(n.err)
}

func (n *baseNode) Filter(fn core.PredicateFunc) core.Node {
	n.setError(fmt.Errorf("filter requires array node"))
	return newInvalidNode(n.err)
}

func (n *baseNode) Map(fn core.TransformFunc) core.Node {
	n.setError(fmt.Errorf("map requires array node"))
	return newInvalidNode(n.err)
}

func (n *baseNode) ForEach(fn func(keyOrIndex interface{}, value core.Node)) {
	n.setError(fmt.Errorf("not a collection node"))
}

func (n *baseNode) Len() int {
	n.setError(fmt.Errorf("not a collection node"))
	return 0
}

func (n *baseNode) RegisterFunc(name string, fn core.UnaryPathFunc) core.Node {
	if n.err != nil {
		return newInvalidNode(n.err)
	}
	if n.funcs == nil || *n.funcs == nil {
		m := make(map[string]core.UnaryPathFunc)
		n.funcs = &m
	}
	// copy-on-write to avoid side effects on shared maps
	newFuncs := make(map[string]core.UnaryPathFunc, len(*n.funcs)+1)
	for k, v := range *n.funcs {
		newFuncs[k] = v
	}
	newFuncs[name] = fn
	n.funcs = &newFuncs
	return newInvalidNode(fmt.Errorf("register on base node not chainable"))
}

func (n *baseNode) RemoveFunc(name string) core.Node {
	if n.err != nil {
		return newInvalidNode(n.err)
	}
	if n.funcs == nil || *n.funcs == nil {
		return newInvalidNode(fmt.Errorf("no funcs to remove"))
	}
	newFuncs := make(map[string]core.UnaryPathFunc, len(*n.funcs))
	for k, v := range *n.funcs {
		if k != name {
			newFuncs[k] = v
		}
	}
	n.funcs = &newFuncs
	return newInvalidNode(fmt.Errorf("remove on base node not chainable"))
}

func (n *baseNode) Set(key string, value interface{}) core.Node {
	n.setError(fmt.Errorf("set requires object node"))
	return newInvalidNode(n.err)
}

func (n *baseNode) Append(value interface{}) core.Node {
	n.setError(fmt.Errorf("append requires array node"))
	return newInvalidNode(n.err)
}

func (n_ *baseNode) CallFunc(name string) core.Node {
	if n_.err != nil {
		return newInvalidNode(n_.err)
	}
	if n_.funcs == nil || *n_.funcs == nil {
		return newInvalidNode(fmt.Errorf("func not found: %s", name))
	}
	// Not safe to call with base node (doesn't satisfy core.Node). Concrete nodes override this.
	if _, ok := (*n_.funcs)[name]; ok {
		return newInvalidNode(fmt.Errorf("call on unsupported node type"))
	}
	return newInvalidNode(fmt.Errorf("func not found: %s", name))
}

func (n *baseNode) Apply(fn core.PathFunc) core.Node {
	if n.err != nil {
		return newInvalidNode(n.err)
	}
	// Concrete nodes override. Base returns unsupported.
	return newInvalidNode(fmt.Errorf("apply not supported on this node type"))
}

func (n *baseNode) String() string {
	return n.Raw()
}

func (n *baseNode) MustString() string {
	panic(core.ErrTypeAssertion)
}

func (n *baseNode) Float() float64 {
	n.setError(core.ErrTypeAssertion)
	return 0
}

func (n *baseNode) MustFloat() float64 {
	panic(core.ErrTypeAssertion)
}

func (n *baseNode) Int() int64 {
	n.setError(core.ErrTypeAssertion)
	return 0
}

func (n *baseNode) MustInt() int64 {
	panic(core.ErrTypeAssertion)
}

func (n *baseNode) Bool() bool {
	n.setError(core.ErrTypeAssertion)
	return false
}

func (n *baseNode) MustBool() bool {
	panic(core.ErrTypeAssertion)
}

func (n *baseNode) Time() time.Time {
	n.setError(core.ErrTypeAssertion)
	return time.Time{}
}

func (n *baseNode) MustTime() time.Time {
	panic(core.ErrTypeAssertion)
}

func (n *baseNode) Array() []core.Node {
	n.setError(core.ErrTypeAssertion)
	return nil
}

func (n *baseNode) MustArray() []core.Node {
	panic(core.ErrTypeAssertion)
}

func (n *baseNode) Interface() interface{} {
	return nil
}

func (n *baseNode) RawFloat() (float64, bool) {
	return 0, false
}

func (n *baseNode) RawString() (string, bool) {
	return "", false
}

func (n *baseNode) Strings() []string {
	n.setError(core.ErrTypeAssertion)
	return nil
}

func (n *baseNode) Contains(value string) bool {
	n.setError(core.ErrTypeAssertion)
	return false
}

func (n *baseNode) AsMap() map[string]core.Node {
	n.setError(core.ErrTypeAssertion)
	return nil
}

func (n *baseNode) MustAsMap() map[string]core.Node {
	panic(core.ErrTypeAssertion)
}

// Keys provides object keys; base node marks error and returns nil.
func (n *baseNode) Keys() []string {
	n.setError(fmt.Errorf("not an object node"))
	return nil
}
