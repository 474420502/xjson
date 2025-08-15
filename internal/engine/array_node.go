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
	// 如果是负索引，先完整解析以确保长度
	if i < 0 {
		n.lazyParse()
	} else {
		n.lazyParsePath([]string{fmt.Sprintf("%d", i)})
	}
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

		// Clear query cache since we're modifying the node
		n.baseNode.clearQueryCache()
	} else {
		return newInvalidNode(fmt.Errorf("index out of bounds for set: %d", idx))
	}

	return n
}

// SetByPath implements the SetByPath method for arrayNode
func (n *arrayNode) SetByPath(path string, value interface{}) core.Node {
	return n.baseNode.SetByPath(path, value)
}

func (n *arrayNode) Append(value interface{}) core.Node {
	if n.err != nil {
		return n
	}
	n.lazyParse()
	n.isDirty = true // Mark as dirty so String() will regenerate

	// Also mark all ancestors as dirty to ensure String() regeneration
	current := n.parent
	for current != nil {
		if obj, ok := current.(*objectNode); ok {
			obj.isDirty = true
			current = obj.parent
		} else if arr, ok := current.(*arrayNode); ok {
			arr.isDirty = true
			current = arr.parent
		} else {
			break
		}
	}

	// Clear query cache since we're modifying the node
	n.baseNode.clearQueryCache()

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
	// 如果未修改并且存在原始数据，则返回原始数据
	if !n.isDirty && n.Raw() != "" {
		return n.Raw()
	}

	var buf bytes.Buffer
	buf.WriteByte('[')
	for i, v := range n.value {
		if i > 0 {
			// 在每个元素之间插入逗号
			buf.WriteByte(',')
		}
		// 将每个元素转换为字符串并添加到缓冲区
		buf.WriteString(v.String())
	}
	buf.WriteByte(']')
	// 返回构建好的字符串表示
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
	// For root node, pass nil as parent to avoid setting root as its own parent
	var parent core.Node
	if n.parent != nil {
		parent = n
	}
	parsedNode := p.parseArray(parent)
	if err := parsedNode.Error(); err != nil {
		n.err = err
		return
	}

	// copy values and reparent children to this node
	if cast, ok := parsedNode.(*arrayNode); ok {
		vals := make([]core.Node, 0, len(cast.value))
		for _, child := range cast.value {
			if bn, ok := child.(*baseNode); ok {
				bn.parent = n
			} else if inode, ok := child.(interface{ setParent(core.Node) }); ok {
				inode.setParent(n)
			}
			vals = append(vals, child)
		}
		n.value = vals
	}
}

// lazyParsePath parses only the nodes needed for a specific path
func (n *arrayNode) lazyParsePath(path []string) {
	if n.parsed.Load() {
		return
	}
	if len(n.raw) == 0 {
		n.parsed.Store(true)
		return
	}
	if len(path) == 0 {
		n.lazyParse()
		return
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.parsed.Load() || n.isDirty {
		return
	}

	// Check if we already have enough elements
	idx, err := strconv.Atoi(path[0])
	if err != nil {
		n.lazyParse()
		return
	}

	// For negative indices, we need to parse the entire array to know the length
	if idx < 0 {
		n.lazyParse()
		return
	}

	// If we already have enough elements, no need to parse more
	if idx < len(n.value) {
		return
	}
	// We'll scan raw to parse elements one-by-one until we reach idx or the end.
	raw := n.raw
	// find opening bracket
	pos := 0
	for pos < len(raw) && raw[pos] != '[' {
		pos++
	}
	if pos >= len(raw) {
		n.lazyParse()
		return
	}
	pos++ // skip '['

	// helper to skip whitespace
	skipWS := func() {
		for pos < len(raw) {
			c := raw[pos]
			if c == ' ' || c == '\n' || c == '\r' || c == '\t' {
				pos++
				continue
			}
			break
		}
	}

	// initialize value slice if nil
	if n.value == nil {
		n.value = make([]core.Node, 0)
	}

	curIndex := len(n.value)
	for pos < len(raw) {
		skipWS()
		if pos >= len(raw) {
			break
		}
		if raw[pos] == ']' {
			// end of array
			n.parsed.Store(true)
			return
		}

		// determine element end
		elemStart := pos
		var elemEnd int
		switch raw[pos] {
		case '{':
			elemEnd = findMatchingBrace(raw, pos)
		case '[':
			elemEnd = findMatchingBracket(raw, pos)
		case '"':
			elemEnd = findMatchingQuote(raw, pos)
		default:
			elemEnd = findValueEnd(raw, pos)
		}
		if elemEnd == -1 {
			n.lazyParse()
			return
		}

		// parse this element
		segment := raw[elemStart : elemEnd+1]
		p := newParser(segment, n.funcs)
		child := p.doParse(n)
		if child == nil || !child.IsValid() {
			if child != nil {
				n.err = child.Error()
			} else {
				n.lazyParse()
			}
			return
		}
		// set proper parent
		if bn, ok := child.(*baseNode); ok {
			bn.parent = n
		} else if inode, ok := child.(interface{ setParent(core.Node) }); ok {
			inode.setParent(n)
		}
		n.value = append(n.value, child)

		// if we've reached the requested index, stop parsing further
		if curIndex == idx {
			return
		}
		curIndex++

		// advance position
		pos = elemEnd + 1
		skipWS()
		if pos < len(raw) && raw[pos] == ',' {
			pos++
			continue
		}
		if pos < len(raw) && raw[pos] == ']' {
			n.parsed.Store(true)
			return
		}
		// unexpected token -> fallback to full parse
		n.lazyParse()
		return
	}

	// if we reach here, we didn't find enough elements; parse full array to establish state
	n.lazyParse()
}

func (n *arrayNode) addChild(child core.Node) {
	if n.value == nil {
		n.value = make([]core.Node, 0)
	}

	// Instead of type assertion, we use the Parent() and SetParent() pattern
	// All node types embed baseNode which has parent field
	if bn, ok := child.(*baseNode); ok {
		bn.parent = n
	} else if inode, ok := child.(interface{ setParent(core.Node) }); ok {
		inode.setParent(n)
	}

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
