package xjson

import (
	"errors"
	"time"

	"github.com/474420502/xjson/internal/core"
	"github.com/474420502/xjson/internal/engine"
)

type Node = core.Node
type NodeType = core.NodeType

const (
	ObjectNode  = core.ObjectNode
	ArrayNode   = core.ArrayNode
	StringNode  = core.StringNode
	NumberNode  = core.NumberNode
	BoolNode    = core.BoolNode
	NullNode    = core.NullNode
	InvalidNode = core.InvalidNode
)

// nodeWrapper wraps an engine.Node to expose it as a core.Node.
// This acts as an adapter between the internal engine and the public API.
type nodeWrapper struct {
	engineNode engine.Node
}

// Ensuring nodeWrapper implements the core.Node interface at compile time.
var _ core.Node = (*nodeWrapper)(nil)

func (w *nodeWrapper) Type() core.NodeType {
	return core.NodeType(w.engineNode.Type())
}

func (w *nodeWrapper) IsValid() bool {
	return w.engineNode.IsValid()
}

func (w *nodeWrapper) Error() error {
	return w.engineNode.Error()
}

func (w *nodeWrapper) Path() string {
	return w.engineNode.Path()
}

func (w *nodeWrapper) Raw() string {
	return w.engineNode.Raw()
}

func (w *nodeWrapper) Get(key string) core.Node {
	return &nodeWrapper{engineNode: w.engineNode.Get(key)}
}

func (w *nodeWrapper) Index(i int) core.Node {
	return &nodeWrapper{engineNode: w.engineNode.Index(i)}
}

func (w *nodeWrapper) Query(path string) core.Node {
	return &nodeWrapper{engineNode: w.engineNode.Query(path)}
}

func (w *nodeWrapper) ForEach(iterator func(keyOrIndex interface{}, value core.Node)) {
	w.engineNode.ForEach(func(keyOrIndex interface{}, value engine.Node) {
		iterator(keyOrIndex, &nodeWrapper{engineNode: value})
	})
}

func (w *nodeWrapper) Len() int {
	return w.engineNode.Len()
}

func (w *nodeWrapper) String() string {
	return w.engineNode.String()
}

func (w *nodeWrapper) MustString() string {
	return w.engineNode.MustString()
}

func (w *nodeWrapper) Float() float64 {
	return w.engineNode.Float()
}

func (w *nodeWrapper) MustFloat() float64 {
	return w.engineNode.MustFloat()
}

func (w *nodeWrapper) Int() int64 {
	return w.engineNode.Int()
}

func (w *nodeWrapper) MustInt() int64 {
	return w.engineNode.MustInt()
}

func (w *nodeWrapper) Bool() bool {
	return w.engineNode.Bool()
}

func (w *nodeWrapper) MustBool() bool {
	return w.engineNode.MustBool()
}

func (w *nodeWrapper) Time() time.Time {
	return w.engineNode.Time()
}

func (w *nodeWrapper) MustTime() time.Time {
	return w.engineNode.MustTime()
}

func (w *nodeWrapper) Array() []core.Node {
	engineNodes := w.engineNode.Array()
	if engineNodes == nil {
		return nil
	}
	nodes := make([]core.Node, len(engineNodes))
	for i, n := range engineNodes {
		nodes[i] = &nodeWrapper{engineNode: n}
	}
	return nodes
}

func (w *nodeWrapper) MustArray() []core.Node {
	return w.Array()
}

func (w *nodeWrapper) Interface() interface{} {
	return w.engineNode.Interface()
}

func (w *nodeWrapper) Filter(fn func(core.Node) bool) core.Node {
	return &nodeWrapper{engineNode: w.engineNode.Filter(func(n engine.Node) bool {
		return fn(&nodeWrapper{engineNode: n})
	})}
}

func (w *nodeWrapper) Map(fn func(core.Node) interface{}) core.Node {
	return &nodeWrapper{engineNode: w.engineNode.Map(func(n engine.Node) interface{} {
		return fn(&nodeWrapper{engineNode: n})
	})}
}

func (w *nodeWrapper) Set(key string, value interface{}) core.Node {
	return &nodeWrapper{engineNode: w.engineNode.Set(key, value)}
}

func (w *nodeWrapper) Append(value interface{}) core.Node {
	return &nodeWrapper{engineNode: w.engineNode.Append(value)}
}

func (w *nodeWrapper) RawFloat() (float64, bool) {
	return w.engineNode.RawFloat()
}

func (w *nodeWrapper) RawString() (string, bool) {
	return w.engineNode.RawString()
}

func (w *nodeWrapper) Strings() []string {
	return w.engineNode.Strings()
}

func (w *nodeWrapper) Contains(value string) bool {
	return w.engineNode.Contains(value)
}

func (w *nodeWrapper) Func(name string, fn func(core.Node) core.Node) core.Node {
	engineFunc := func(n engine.Node) engine.Node {
		resultNode := fn(&nodeWrapper{engineNode: n})
		if wrapper, ok := resultNode.(*nodeWrapper); ok {
			return wrapper.engineNode
		}
		return nil
	}
	return &nodeWrapper{engineNode: w.engineNode.Func(name, engineFunc)}
}

func (w *nodeWrapper) CallFunc(name string) core.Node {
	return &nodeWrapper{engineNode: w.engineNode.CallFunc(name)}
}

func (w *nodeWrapper) RemoveFunc(name string) core.Node {
	return &nodeWrapper{engineNode: w.engineNode.RemoveFunc(name)}
}

func (w *nodeWrapper) GetFuncs() *map[string]func(core.Node) core.Node {
	engineFuncs := w.engineNode.GetFuncs()
	if engineFuncs == nil {
		return nil
	}
	coreFuncs := make(map[string]func(core.Node) core.Node)
	for name, engineFn := range *engineFuncs {
		// Create a closure to capture the engineFn
		fn := engineFn
		coreFuncs[name] = func(n core.Node) core.Node {
			if wrapper, ok := n.(*nodeWrapper); ok {
				return &nodeWrapper{engineNode: fn(wrapper.engineNode)}
			}
			return &nodeWrapper{engineNode: engine.NewInvalidNode(n.Path(), errors.New("node is not a wrapper"))}
		}
	}
	return &coreFuncs
}

// Parse takes a JSON string and returns the root node of the parsed structure.
func Parse(data string) (core.Node, error) {
	node, err := engine.ParseJSONToNode(data)
	if err != nil {
		return nil, err
	}
	return &nodeWrapper{engineNode: node}, nil
}

// ParseBytes is a convenience wrapper around Parse for byte slices.
func ParseBytes(data []byte) (core.Node, error) {
	return Parse(string(data))
}
