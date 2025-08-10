package engine

import (
	"time"

	"github.com/474420502/xjson/internal/core"
)

// invalidNode represents a node in an error state.
type invalidNode struct {
	baseNode
}

func (n *invalidNode) Parent() core.Node {
	return n.parent
}

func (n *invalidNode) Type() core.NodeType                                      { return core.Invalid }
func (n *invalidNode) IsValid() bool                                            { return false }
func (n *invalidNode) Query(path string) core.Node                              { return n }
func (n *invalidNode) Get(key string) core.Node                                 { return n }
func (n *invalidNode) Index(i int) core.Node                                    { return n }
func (n *invalidNode) ForEach(fn func(keyOrIndex interface{}, value core.Node)) {}
func (n *invalidNode) Len() int                                                 { return 0 }
func (n *invalidNode) Set(key string, value interface{}) core.Node              { return n }
func (n *invalidNode) Append(value interface{}) core.Node                       { return n }

func (n *invalidNode) String() string                  { return "invalid" }
func (n *invalidNode) MustString() string              { panic(n.err) }
func (n *invalidNode) Float() float64                  { return 0 }
func (n *invalidNode) MustFloat() float64              { panic(n.err) }
func (n *invalidNode) Int() int64                      { return 0 }
func (n *invalidNode) MustInt() int64                  { panic(n.err) }
func (n *invalidNode) Bool() bool                      { return false }
func (n *invalidNode) MustBool() bool                  { panic(n.err) }
func (n *invalidNode) Time() time.Time                 { return time.Time{} }
func (n *invalidNode) MustTime() time.Time             { panic(n.err) }
func (n *invalidNode) Array() []core.Node              { return nil }
func (n *invalidNode) MustArray() []core.Node          { panic(n.err) }
func (n *invalidNode) Interface() interface{}          { return nil }
func (n *invalidNode) RawString() (string, bool)       { return "", false }
func (n *invalidNode) Strings() []string               { return nil }
func (n *invalidNode) Keys() []string                  { return nil }
func (n *invalidNode) Contains(value string) bool      { return false }
func (n *invalidNode) AsMap() map[string]core.Node     { return nil }
func (n *invalidNode) MustAsMap() map[string]core.Node { panic(n.err) }
