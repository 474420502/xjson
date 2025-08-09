package engine

import (
	"fmt"
	"time"

	"github.com/474420502/xjson/internal/core"
)

// baseNode serves as the base for all concrete node types, implementing common
// functionalities like error handling and path management.
type baseNode struct {
	err   error
	path  string
	raw   *string
	funcs *map[string]func(core.Node) core.Node // Changed to pointer
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
	// Fallback for non-root nodes: marshal self.
	// This will be a problem for types that don't override String().
	// Specifically, stringNode needs to marshal to a quoted string.
	return n.String()
}

func (n *baseNode) setRaw(raw *string) {
	n.raw = raw
}

func (n *baseNode) setError(err error) {
	if n.err == nil {
		n.err = err
	}
}

func (n *baseNode) GetFuncs() *map[string]func(core.Node) core.Node {
	return n.funcs
}

// Common methods that will be overridden by specific node types but need to be defined
// to satisfy the Node interface

func (n *baseNode) Type() core.NodeType                           { return core.InvalidNode }
func (n *baseNode) Get(key string) core.Node                      { return nil }
func (n *baseNode) Index(i int) core.Node                         { return nil }
func (n *baseNode) Query(path string) core.Node                   { return nil }
func (n *baseNode) ForEach(iterator func(interface{}, core.Node)) {}
func (n *baseNode) Len() int                                      { return 0 }
func (n *baseNode) String() string                                { return "" }
func (n *baseNode) MustString() string                            { panic("not implemented") }
func (n *baseNode) Float() float64                                { return 0 }
func (n *baseNode) MustFloat() float64                            { panic("not implemented") }
func (n *baseNode) Int() int64                                    { return 0 }
func (n *baseNode) MustInt() int64                                { panic("not implemented") }
func (n *baseNode) Bool() bool                                    { return false }
func (n *baseNode) MustBool() bool                                { panic("not implemented") }
func (n *baseNode) Time() time.Time                               { return time.Time{} }
func (n *baseNode) MustTime() time.Time                           { panic("not implemented") }
func (n *baseNode) Array() []core.Node                            { return nil }
func (n *baseNode) MustArray() []core.Node                        { panic("not implemented") }
func (n *baseNode) Interface() interface{}                        { return nil }
func (n *baseNode) Set(key string, value interface{}) core.Node   { return nil }
func (n *baseNode) Append(value interface{}) core.Node            { return nil }
func (n *baseNode) RawFloat() (float64, bool)                     { return 0, false }
func (n *baseNode) RawString() (string, bool)                     { return "", false }
func (n *baseNode) Contains(value string) bool                    { return false }
func (n *baseNode) Strings() []string                             { return nil }
func (n *baseNode) AsMap() map[string]core.Node                   { return nil }
func (n *baseNode) MustAsMap() map[string]core.Node               { panic("not implemented") }

// Deprecated: Use RegisterFunc and CallFunc instead
func (n *baseNode) Func(name string, fn func(core.Node) core.Node) core.Node { return nil }

func (n *baseNode) CallFunc(name string) core.Node   { return nil }
func (n *baseNode) RemoveFunc(name string) core.Node { return nil }
func (n *baseNode) Filter(fn core.PredicateFunc) core.Node {
	return NewInvalidNode(n.path, ErrTypeAssertion)
}
func (n *baseNode) Map(fn core.TransformFunc) core.Node {
	return NewInvalidNode(n.path, ErrTypeAssertion)
}
func (n *baseNode) RegisterFunc(name string, fn core.UnaryPathFunc) core.Node { return nil }
func (n *baseNode) Apply(fn core.PathFunc) core.Node {
	// 这个不能改, 必须panic
	if fn == nil {
		panic("Apply function cannot be nil")
	}
	return NewInvalidNode(n.path, fmt.Errorf("unsupported function signature for Apply: %T", fn))
}
