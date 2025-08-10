package engine

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/474420502/xjson/internal/core"
)

type arrayNode struct {
	baseNode
	value   []core.Node
	isDirty bool
}

func (n *arrayNode) Type() core.NodeType { return core.Array }

func (n *arrayNode) Len() int {
	if n.err != nil {
		return 0
	}
	n.lazyParse()
	return len(n.value)
}

func (n *arrayNode) Index(i int) core.Node {
	if n.err != nil {
		return n
	}
	n.lazyParse()
	if i < 0 {
		i = len(n.value) + i
	}
	if i >= 0 && i < len(n.value) {
		return n.value[i]
	}
	return newInvalidNode(fmt.Errorf("index out of bounds: %d", i))
}

func (n *arrayNode) ForEach(fn func(keyOrIndex interface{}, value core.Node)) {
	if n.err != nil {
		return
	}
	n.lazyParse()
	for i, v := range n.value {
		fn(i, v)
	}
}

func (n *arrayNode) Set(key string, value interface{}) core.Node {
	if n.err != nil {
		return n
	}
	n.lazyParse()
	idx, err := strconv.Atoi(key)
	if err != nil {
		return newInvalidNode(fmt.Errorf("invalid index for array set: %s", key))
	}

	if idx < 0 {
		idx = len(n.value) + idx
	}

	if idx >= 0 && idx < len(n.value) {
		n.isDirty = true
		child := NewNodeFromInterface(n, value, n.funcs)
		if !child.IsValid() {
			n.setError(child.Error())
			return n
		}
		n.value[idx] = child
	} else {
		return newInvalidNode(fmt.Errorf("index out of bounds for set: %d", idx))
	}

	return n
}

func (n *arrayNode) Append(value interface{}) core.Node {
	if n.err != nil {
		return n
	}
	n.lazyParse()
	n.isDirty = true
	child := NewNodeFromInterface(n, value, n.funcs)
	if !child.IsValid() {
		n.setError(child.Error())
		return n
	}
	n.value = append(n.value, child)
	return n
}

func (n *arrayNode) Array() []core.Node {
	if n.err != nil {
		return nil
	}
	n.lazyParse()
	if n.value == nil {
		return []core.Node{}
	}
	return n.value
}

func (n *arrayNode) MustArray() []core.Node {
	if n.err != nil {
		panic(n.err)
	}
	n.lazyParse()
	return n.value
}

func (n *arrayNode) String() string {
	if n.err != nil {
		return ""
	}
	n.lazyParse()
	if !n.isDirty && n.Raw() != "" {
		return n.Raw()
	}

	var buf bytes.Buffer
	buf.WriteByte('[')
	for i, v := range n.value {
		if i > 0 {
			buf.WriteByte(',')
		}
		buf.WriteString(v.String())
	}
	buf.WriteByte(']')
	return buf.String()
}

func (n *arrayNode) Interface() interface{} {
	if n.err != nil {
		return nil
	}
	n.lazyParse()
	s := make([]interface{}, len(n.value))
	for i, v := range n.value {
		s[i] = v.Interface()
	}
	return s
}

func (n *arrayNode) lazyParse() {
	if n.parsed.Load() || n.isDirty {
		return
	}
	if len(n.raw) == 0 { // nothing to parse (constructed node)
		n.parsed.Store(true)
		return
	}

	n.mu.Lock()
	defer n.mu.Unlock()

	if n.parsed.Load() || n.isDirty {
		return
	}
	defer n.parsed.Store(true)

	p := newParser(n.raw, n.funcs)
	// start from the beginning of raw to parse the array
	p.pos = 0
	parsedNode := p.parseArray(n)
	if err := parsedNode.Error(); err != nil {
		n.err = err
		return
	}

	// copy values
	if cast, ok := parsedNode.(*arrayNode); ok {
		n.value = cast.value
	}
}

func (n *arrayNode) addChild(child core.Node) {
	if n.value == nil {
		n.value = make([]core.Node, 0)
	}
	child.(*baseNode).parent = n
	n.value = append(n.value, child)
}

func (n *arrayNode) Strings() []string {
	if n.err != nil {
		return nil
	}
	n.lazyParse()
	res := make([]string, 0, len(n.value))
	for _, v := range n.value {
		res = append(res, v.String())
	}
	return res
}

func (n *arrayNode) Filter(fn core.PredicateFunc) core.Node {
	if n.err != nil {
		return n
	}
	n.lazyParse()
	out := NewArrayNode(n, nil, n.funcs)
	arr := out.(*arrayNode)
	for _, v := range n.value {
		if fn(v) {
			arr.value = append(arr.value, v)
		}
	}
	return out
}

func (n *arrayNode) Map(fn core.TransformFunc) core.Node {
	if n.err != nil {
		return n
	}
	n.lazyParse()
	out := NewArrayNode(n, nil, n.funcs)
	arr := out.(*arrayNode)
	for _, v := range n.value {
		arr.value = append(arr.value, NewNodeFromInterface(arr, fn(v), n.funcs))
	}
	return out
}
