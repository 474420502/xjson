package engine

import (
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

// SetParent sets the parent of the current node.
func (n *baseNode) SetParent(parent core.Node) {
	n.parent = parent
}

// Stub implementations for methods to be implemented by concrete node types.
// These will panic if called on a baseNode directly.

func (n *baseNode) Type() core.NodeType {
	panic("Type() must be implemented by concrete node type")
}

func (n *baseNode) Query(path string) core.Node {
	panic("Query() must be implemented by concrete node type")
}

func (n *baseNode) Get(key string) core.Node {
	panic("Get() must be implemented by concrete node type")
}

func (n *baseNode) Index(i int) core.Node {
	panic("Index() must be implemented by concrete node type")
}

func (n *baseNode) Filter(fn core.PredicateFunc) core.Node {
	panic("Filter() must be implemented by concrete node type")
}

func (n *baseNode) Map(fn core.TransformFunc) core.Node {
	panic("Map() must be implemented by concrete node type")
}

func (n *baseNode) ForEach(fn func(keyOrIndex interface{}, value core.Node)) {
	panic("ForEach() must be implemented by concrete node type")
}

func (n *baseNode) Len() int {
	panic("Len() must be implemented by concrete node type")
}

func (n *baseNode) RegisterFunc(name string, fn core.UnaryPathFunc) core.Node {
	panic("RegisterFunc() must be implemented by concrete node type")
}

func (n *baseNode) RemoveFunc(name string) core.Node {
	panic("RemoveFunc() must be implemented by concrete node type")
}

func (n *baseNode) Set(key string, value interface{}) core.Node {
	panic("Set() must be implemented by concrete node type")
}

func (n *baseNode) Append(value interface{}) core.Node {
	panic("Append() must be implemented by concrete node type")
}

func (n_ *baseNode) CallFunc(name string) core.Node {
	panic("CallFunc() must be implemented by concrete node type")
}

func (n *baseNode) Apply(fn core.PathFunc) core.Node {
	panic("Apply() must be implemented by concrete node type")
}

func (n *baseNode) String() string {
	return n.Raw()
}

func (n *baseNode) MustString() string {
	panic("MustString() must be implemented by concrete node type")
}

func (n *baseNode) Float() float64 {
	panic("Float() must be implemented by concrete node type")
}

func (n *baseNode) MustFloat() float64 {
	panic("MustFloat() must be implemented by concrete node type")
}

func (n *baseNode) Int() int64 {
	panic("Int() must be implemented by concrete node type")
}

func (n *baseNode) MustInt() int64 {
	panic("MustInt() must be implemented by concrete node type")
}

func (n *baseNode) Bool() bool {
	panic("Bool() must be implemented by concrete node type")
}

func (n *baseNode) MustBool() bool {
	panic("MustBool() must be implemented by concrete node type")
}

func (n *baseNode) Time() time.Time {
	panic("Time() must be implemented by concrete node type")
}

func (n *baseNode) MustTime() time.Time {
	panic("MustTime() must be implemented by concrete node type")
}

func (n *baseNode) Array() []core.Node {
	panic("Array() must be implemented by concrete node type")
}

func (n *baseNode) MustArray() []core.Node {
	panic("MustArray() must be implemented by concrete node type")
}

func (n *baseNode) Interface() interface{} {
	panic("Interface() must be implemented by concrete node type")
}

func (n *baseNode) RawFloat() (float64, bool) {
	panic("RawFloat() must be implemented by concrete node type")
}

func (n *baseNode) RawString() (string, bool) {
	panic("RawString() must be implemented by concrete node type")
}

func (n *baseNode) Strings() []string {
	panic("Strings() must be implemented by concrete node type")
}

func (n *baseNode) Contains(value string) bool {
	panic("Contains() must be implemented by concrete node type")
}

func (n *baseNode) AsMap() map[string]core.Node {
	panic("AsMap() must be implemented by concrete node type")
}

func (n *baseNode) MustAsMap() map[string]core.Node {
	panic("MustAsMap() must be implemented by concrete node type")
}
