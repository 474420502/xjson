package engine

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/474420502/xjson/internal/core"
)

type baseNode struct {
	raw    []byte
	start  int
	end    int
	parent core.Node
	funcs  *map[string]core.UnaryPathFunc
	err    error

	// lazy parse helpers for composite nodes
	parsed atomic.Bool
	mu     sync.Mutex

	// self holds the concrete node implementing core.Node to avoid losing
	// the dynamic type when methods are promoted from the embedded baseNode.
	self core.Node
}

func (n *baseNode) Raw() string {
	if n.err != nil {
		return ""
	}
	if len(n.raw) == 0 {
		return ""
	}
	s, e := n.start, n.end
	if e == 0 {
		e = len(n.raw)
	}
	if s < 0 {
		s = 0
	}
	if e > len(n.raw) {
		e = len(n.raw)
	}
	if s > e {
		return ""
	}
	return string(n.raw[s:e])
}

func (n *baseNode) RawBytes() []byte {
	if n.err != nil {
		return nil
	}
	if len(n.raw) == 0 {
		return nil
	}
	s, e := n.start, n.end
	if e == 0 {
		e = len(n.raw)
	}
	if s < 0 {
		s = 0
	}
	if e > len(n.raw) {
		e = len(n.raw)
	}
	if s > e {
		return nil
	}
	return n.raw[s:e]
}

func (n *baseNode) IsValid() bool {
	return n.err == nil
}

func (n *baseNode) Error() error {
	return n.err
}

func (n *baseNode) Parent() core.Node {
	return n.parent
}

func (n *baseNode) GetFuncs() *map[string]core.UnaryPathFunc {
	return n.funcs
}

func (n *baseNode) setError(err error) {
	if n.err == nil {
		n.err = err
	}
}

func (n *baseNode) Path() string {
	// Path building logic can be complex, so this is a simplified stub.
	// A full implementation would require knowing the key/index.
	if n.parent != nil {
		return n.parent.Path() + "/?"
	}
	return ""
}

func (n *baseNode) Query(path string) core.Node {
	if n.err != nil {
		return newInvalidNode(n.err)
	}
	start := core.Node(n)
	if n.self != nil {
		start = n.self
	}
	return applySimpleQuery(start, path)
}

func (n *baseNode) RegisterFunc(name string, fn core.UnaryPathFunc) core.Node {
	if n.err != nil {
		return n.selfOrMe()
	}
	if n.funcs == nil {
		newFuncs := make(map[string]core.UnaryPathFunc)
		n.funcs = &newFuncs
	}
	(*n.funcs)[name] = fn
	return n.selfOrMe()
}

func (n *baseNode) RemoveFunc(name string) core.Node {
	if n.err != nil {
		return n.selfOrMe()
	}
	if n.funcs != nil {
		delete(*n.funcs, name)
	}
	return n.selfOrMe()
}

func (n *baseNode) CallFunc(name string) core.Node {
	if n.err != nil {
		return n.selfOrMe()
	}
	if n.funcs != nil {
		if fn, ok := (*n.funcs)[name]; ok {
			// Always call with the concrete node
			if n.self != nil {
				return fn(n.self)
			}
			return fn(n)
		}
	}
	return newInvalidNode(fmt.Errorf("func '%s' not found", name))
}

func (n *baseNode) selfOrMe() core.Node {
	if n.self != nil {
		return n.self
	}
	return n
}

// Default/placeholder implementations for methods that must be overridden
// by concrete types (like objectNode, arrayNode, etc.).

func (n *baseNode) Type() core.NodeType { return core.Invalid }
func (n *baseNode) Len() int            { return 1 }
func (n *baseNode) Get(key string) core.Node {
	return newInvalidNode(fmt.Errorf("get not supported on type %s", n.Type()))
}
func (n *baseNode) Index(i int) core.Node {
	return newInvalidNode(fmt.Errorf("index not supported on type %s", n.Type()))
}
func (n *baseNode) Set(key string, value interface{}) core.Node {
	return newInvalidNode(fmt.Errorf("set not supported on type %s", n.Type()))
}
func (n *baseNode) Append(value interface{}) core.Node {
	return newInvalidNode(fmt.Errorf("append not supported on type %s", n.Type()))
}

func (n *baseNode) Filter(fn core.PredicateFunc) core.Node {
	return newInvalidNode(fmt.Errorf("filter not supported on type %s", n.Type()))
}

func (n *baseNode) Map(fn core.TransformFunc) core.Node {
	return newInvalidNode(fmt.Errorf("map not supported on type %s", n.Type()))
}

func (n *baseNode) ForEach(fn func(keyOrIndex interface{}, value core.Node)) {
	fn(nil, n)
}
func (n *baseNode) SetValue(v interface{}) core.Node {
	return newInvalidNode(fmt.Errorf("setValue not supported on type %s", n.Type()))
}
func (n *baseNode) Apply(fn core.PathFunc) core.Node {
	return newInvalidNode(fmt.Errorf("apply not supported on type %s", n.Type()))
}
func (n *baseNode) String() string                  { return n.Raw() }
func (n *baseNode) MustString() string              { panic(core.ErrTypeAssertion) }
func (n *baseNode) Float() float64                  { return 0 }
func (n *baseNode) MustFloat() float64              { panic(core.ErrTypeAssertion) }
func (n *baseNode) Int() int64                      { return 0 }
func (n *baseNode) MustInt() int64                  { panic(core.ErrTypeAssertion) }
func (n *baseNode) Bool() bool                      { return false }
func (n *baseNode) MustBool() bool                  { panic(core.ErrTypeAssertion) }
func (n *baseNode) Time() time.Time                 { return time.Time{} }
func (n *baseNode) MustTime() time.Time             { panic(core.ErrTypeAssertion) }
func (n *baseNode) Array() []core.Node              { return nil }
func (n *baseNode) MustArray() []core.Node          { panic(core.ErrTypeAssertion) }
func (n *baseNode) Interface() interface{}          { return nil }
func (n *baseNode) RawFloat() (float64, bool)       { return 0, false }
func (n *baseNode) RawString() (string, bool)       { s := n.Raw(); return s, true }
func (n *baseNode) Strings() []string               { return []string{n.String()} }
func (n *baseNode) Keys() []string                  { return nil }
func (n *baseNode) Contains(value string) bool      { return n.String() == value }
func (n *baseNode) AsMap() map[string]core.Node     { return nil }
func (n *baseNode) MustAsMap() map[string]core.Node { panic(core.ErrTypeAssertion) }
