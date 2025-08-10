package engine

import (
	"fmt"
	"time"

	"github.com/474420502/xjson/internal/core"
)

// invalidNode represents a node in an error state.
// Any operation on an invalidNode will return itself, propagating the error through the chain.
type invalidNode struct {
	baseNode
}

// newInvalidNode creates a new invalid node with a specific error.
func newInvalidNode(err error) *invalidNode {
	return &invalidNode{baseNode: baseNode{err: err}}
}

// newInvalidNodeWithMsg creates a new invalid node with a formatted error message.
func newInvalidNodeWithMsg(format string, a ...interface{}) *invalidNode {
	return newInvalidNode(fmt.Errorf(format, a...))
}

func (n *invalidNode) Type() core.NodeType                                      { return core.Invalid }
func (n *invalidNode) IsValid() bool                                            { return false }
func (n *invalidNode) Query(path string) core.Node                              { return n }
func (n *invalidNode) Get(key string) core.Node                                 { return n }
func (n *invalidNode) Index(i int) core.Node                                    { return n }
func (n *invalidNode) Filter(fn core.PredicateFunc) core.Node                   { return n }
func (n *invalidNode) Map(fn core.TransformFunc) core.Node                      { return n }
func (n *invalidNode) ForEach(fn func(keyOrIndex interface{}, value core.Node)) {}
func (n *invalidNode) Len() int                                                 { return 0 }
func (n *invalidNode) Set(key string, value interface{}) core.Node              { return n }
func (n *invalidNode) Append(value interface{}) core.Node                       { return n }
func (n *invalidNode) SetValue(value interface{}) core.Node                     { return n }
func (n *invalidNode) RegisterFunc(name string, fn core.UnaryPathFunc) core.Node {
	return n
}
func (n *invalidNode) CallFunc(name string) core.Node           { return n }
func (n *invalidNode) RemoveFunc(name string) core.Node         { return n }
func (n *invalidNode) Apply(fn core.PathFunc) core.Node         { return n }
func (n *invalidNode) String() string                           { return "" }
func (n *invalidNode) MustString() string                       { return "" }
func (n *invalidNode) Float() float64                           { return 0 }
func (n *invalidNode) MustFloat() float64                       { return 0 }
func (n *invalidNode) Int() int64                               { return 0 }
func (n *invalidNode) MustInt() int64                           { return 0 }
func (n *invalidNode) Bool() bool                               { return false }
func (n *invalidNode) MustBool() bool                           { return false }
func (n *invalidNode) Time() time.Time                          { return time.Time{} }
func (n *invalidNode) MustTime() time.Time                      { return time.Time{} }
func (n *invalidNode) Array() []core.Node                       { return nil }
func (n *invalidNode) MustArray() []core.Node                   { return nil }
func (n *invalidNode) Interface() interface{}                   { return nil }
func (n *invalidNode) RawFloat() (float64, bool)                { return 0, false }
func (n *invalidNode) RawString() (string, bool)                { return "", false }
func (n *invalidNode) Strings() []string                        { return nil }
func (n *invalidNode) Keys() []string                           { return nil }
func (n *invalidNode) Contains(value string) bool               { return false }
func (n *invalidNode) AsMap() map[string]core.Node              { return nil }
func (n *invalidNode) MustAsMap() map[string]core.Node          { return nil }
func (n *invalidNode) GetFuncs() *map[string]core.UnaryPathFunc { return nil }
func (n *invalidNode) Path() string                             { return "" }
