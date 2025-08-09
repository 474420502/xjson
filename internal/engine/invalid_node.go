package engine

import (
	"time"

	"github.com/474420502/xjson/internal/core"
)

// invalidNode represents a node that is the result of a failed operation.
type invalidNode struct {
	baseNode
}

func NewInvalidNode(path string, err error) core.Node {
	node := &invalidNode{}
	node.path = path
	node.setError(err)
	// No funcs map for invalid nodes, as they don't participate in function calls
	return node
}

func (n *invalidNode) Type() core.NodeType         { return core.InvalidNode }
func (n *invalidNode) Get(key string) core.Node    { return n }
func (n *invalidNode) Index(i int) core.Node       { return n }
func (n *invalidNode) Query(path string) core.Node { return n }
func (n *invalidNode) ForEach(iterator func(interface{}, core.Node)) {
	_ = n.path // Placeholder for coverage
}
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
func (n *invalidNode) Array() []core.Node                             { return nil }
func (n *invalidNode) MustArray() []core.Node                         { panic(n.err) }
func (n *invalidNode) Interface() interface{}                    { return nil }

// Deprecated: Use RegisterFunc and CallFunc instead
func (n *invalidNode) Func(name string, fn func(core.Node) core.Node) core.Node { return n }

func (n *invalidNode) RegisterFunc(name string, fn core.UnaryPathFunc) core.Node { return n }
func (n *invalidNode) Apply(fn core.PathFunc) core.Node                          { return n }
func (n *invalidNode) CallFunc(name string) core.Node                 { return n }
func (n *invalidNode) RemoveFunc(name string) core.Node               { return n }

func (n *invalidNode) Filter(fn core.PredicateFunc) core.Node         { return n }
func (n *invalidNode) Map(fn core.TransformFunc) core.Node     { return n }
func (n *invalidNode) Set(key string, value interface{}) core.Node { return n }
func (n *invalidNode) Append(value interface{}) core.Node          { return n }
func (n *invalidNode) Raw() string {
	return "invalid"
}
func (n *invalidNode) RawFloat() (float64, bool)              { return 0, false }
func (n *invalidNode) RawString() (string, bool)              { return "", false }
func (n *invalidNode) Contains(value string) bool             { return false }
func (n *invalidNode) Strings() []string                      { return nil }
func (n *invalidNode) AsMap() map[string]core.Node                 { return nil }
func (n *invalidNode) MustAsMap() map[string]core.Node             { panic(n.err) }