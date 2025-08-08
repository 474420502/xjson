package engine

import (
	"time"
)

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

func (n *invalidNode) Type() NodeType         { return InvalidNode }
func (n *invalidNode) Get(key string) Node    { return n }
func (n *invalidNode) Index(i int) Node       { return n }
func (n *invalidNode) Query(path string) Node { return n }
func (n *invalidNode) ForEach(iterator func(interface{}, Node)) {
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
func (n *invalidNode) Array() []Node                             { return nil }
func (n *invalidNode) MustArray() []Node                         { panic(n.err) }
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
func (n *invalidNode) AsMap() map[string]Node                 { return nil }
func (n *invalidNode) MustAsMap() map[string]Node             { panic(n.err) }
