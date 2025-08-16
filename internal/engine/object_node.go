package engine

import (
	"bytes"
	"fmt"
	"sort"
	"unsafe"

	"github.com/474420502/xjson/internal/core"
)

// compareStringBytes compares a string with []byte without allocating
func compareStringBytes(s string, b []byte) bool {
	if len(s) != len(b) {
		return false
	}
	if len(s) == 0 {
		return true
	}
	// Use unsafe to compare without allocation - simple and effective
	return s == unsafe.String(&b[0], len(b))
}

type objectNode struct {
	baseNode
	value      map[string]core.Node
	sortedKeys []string
	isDirty    bool
}

func (n *objectNode) Type() core.NodeType {
	return core.Object
}

func (n *objectNode) Len() int {
	n.lazyParse()
	return len(n.value)
}

func (n *objectNode) Get(key string) core.Node {
	if n.err != nil {
		return n
	}
	// 优先只解析目标 key
	n.lazyParsePath([]string{key})
	if child, ok := n.value[key]; ok {
		return child
	}
	return newInvalidNode(fmt.Errorf("key not found: %s", key))
}

// GetWithPath gets a child node with path information for lazy loading
func (n *objectNode) GetWithPath(key string, path []string) core.Node {
	if n.err != nil {
		return n
	}
	// Only parse what we need for this path
	n.lazyParsePath(path)

	if child, ok := n.value[key]; ok {
		return child
	}
	return newInvalidNode(fmt.Errorf("key not found: %s", key))
}

// LazyGet 实现真正的懒加载获取
func (n *objectNode) LazyGet(key string) core.Node {
	if n.err != nil {
		return n
	}
	n.lazyParsePath([]string{key})
	if child, ok := n.value[key]; ok {
		return child
	}
	return newInvalidNode(fmt.Errorf("key not found: %s", key))
}

// 辅助函数：查找匹配的大括号
func findMatchingBrace(data []byte, start int) int {
	if start >= len(data) || data[start] != '{' {
		return -1
	}

	level := 1
	for i := start + 1; i < len(data); i++ {
		switch data[i] {
		case '{':
			level++
		case '}':
			level--
			if level == 0 {
				return i
			}
		case '"':
			// 跳过字符串中的内容
			for j := i + 1; j < len(data); j++ {
				if data[j] == '"' && (j == 0 || data[j-1] != '\\') {
					i = j
					break
				}
			}
		}
	}

	return -1
}

// 辅助函数：查找匹配的方括号
func findMatchingBracket(data []byte, start int) int {
	if start >= len(data) || data[start] != '[' {
		return -1
	}

	level := 1
	for i := start + 1; i < len(data); i++ {
		switch data[i] {
		case '[':
			level++
		case ']':
			level--
			if level == 0 {
				return i
			}
		case '"':
			// 跳过字符串中的内容
			for j := i + 1; j < len(data); j++ {
				if data[j] == '"' && (j == 0 || data[j-1] != '\\') {
					i = j
					break
				}
			}
		}
	}

	return -1
}

// 辅助函数：查找匹配的引号
func findMatchingQuote(data []byte, start int) int {
	if start >= len(data) || data[start] != '"' {
		return -1
	}

	for i := start + 1; i < len(data); i++ {
		if data[i] == '"' && data[i-1] != '\\' {
			return i
		}
	}

	return -1
}

// 辅助函数：查找值的结束位置
func findValueEnd(data []byte, start int) int {
	for i := start; i < len(data); i++ {
		switch data[i] {
		case ' ', '\t', '\n', '\r', ',', '}', ']':
			return i - 1
		}
	}
	return len(data) - 1
}

func (n *objectNode) ForEach(fn func(keyOrIndex interface{}, value core.Node)) {
	if n.err != nil {
		return
	}
	n.lazyParse()
	if n.sortedKeys == nil {
		keys := make([]string, 0, len(n.value))
		for k := range n.value {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		n.sortedKeys = keys
	}

	for _, k := range n.sortedKeys {
		fn(k, n.value[k])
	}
}

func (n *objectNode) Set(key string, value interface{}) core.Node {
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

	// Update sorted keys
	found := false
	for _, k := range n.sortedKeys {
		if k == key {
			found = true
			break
		}
	}
	if !found {
		n.sortedKeys = append(n.sortedKeys, key)
		sort.Strings(n.sortedKeys)
	}

	child := NewNodeFromInterface(n, value, n.funcs)
	if !child.IsValid() {
		n.setError(child.Error())
		return n
	}
	n.value[key] = child

	return n
}

// SetByPath implements the SetByPath method for objectNode
func (n *objectNode) SetByPath(path string, value interface{}) core.Node {
	return n.baseNode.SetByPath(path, value)
}

// 新增辅助方法来避免重复代码
func (n *objectNode) containsKey(key string) bool {
	if n.sortedKeys == nil {
		return false
	}
	for _, k := range n.sortedKeys {
		if k == key {
			return true
		}
	}
	return false
}

func (n *objectNode) AsMap() map[string]core.Node {
	if n.err != nil {
		return nil
	}
	n.lazyParse()
	return n.value
}

func (n *objectNode) MustAsMap() map[string]core.Node {
	if n.err != nil {
		panic(n.err)
	}
	n.lazyParse()
	return n.value
}

func (n *objectNode) String() string {
	if n.err != nil {
		return ""
	}
	n.lazyParse()
	if !n.isDirty && n.Raw() != "" {
		// Check if any child node is dirty
		hasDirtyChild := false
		for _, child := range n.value {
			switch c := child.(type) {
			case *objectNode:
				if c.isDirty {
					hasDirtyChild = true
					break
				}
			case *arrayNode:
				if c.isDirty {
					hasDirtyChild = true
					break
				}
			}
		}

		if !hasDirtyChild {
			return n.Raw()
		}
	}

	var buf bytes.Buffer
	buf.WriteByte('{')
	if n.sortedKeys == nil {
		keys := make([]string, 0, len(n.value))
		for k := range n.value {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		n.sortedKeys = keys
	}
	for i, k := range n.sortedKeys {
		if i > 0 {
			buf.WriteByte(',')
		}
		buf.WriteString(fmt.Sprintf("%q:%s", k, n.value[k].String()))
	}
	buf.WriteByte('}')
	return buf.String()
}

func (n *objectNode) Keys() []string {
	if n.err != nil {
		return nil
	}
	n.lazyParse()
	if n.sortedKeys == nil {
		keys := make([]string, 0, len(n.value))
		for k := range n.value {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		n.sortedKeys = keys
	}
	return n.sortedKeys
}

func (n *objectNode) Interface() interface{} {
	if n.err != nil {
		return nil
	}
	n.lazyParse()
	m := make(map[string]interface{})
	for k, v := range n.value {
		m[k] = v.Interface()
	}
	return m
}

// lazyParse parses the entire object and sets up children with correct parents
func (n *objectNode) lazyParse() {
	if n.parsed.Load() || n.isDirty {
		return
	}
	if len(n.raw) == 0 { // constructed node
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
	p.pos = 0
	var parent core.Node
	if n.parent != nil {
		parent = n
	}
	parsedNode := p.parseObjectFull(parent)
	if err := parsedNode.Error(); err != nil {
		n.err = err
		return
	}

	if cast, ok := parsedNode.(*objectNode); ok {
		m := make(map[string]core.Node, len(cast.value))
		for k, child := range cast.value {
			if bn, ok := child.(*baseNode); ok {
				bn.parent = n
			} else if inode, ok := child.(interface{ setParent(core.Node) }); ok {
				inode.setParent(n)
			}
			m[k] = child
		}
		n.value = m
		n.sortedKeys = cast.sortedKeys
	}
}

// lazyParsePath parses only the nodes needed for a specific path
func (n *objectNode) lazyParsePath(path []string) {
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
	if n.parsed.Load() {
		n.mu.Unlock()
		return
	}

	// Check if we already have the key
	key := path[0]
	if n.value != nil {
		if _, ok := n.value[key]; ok {
			n.mu.Unlock()
			return
		}
	}

	// Try to find and parse only the requested key by scanning the raw object
	raw := n.raw
	// find opening brace
	pos := 0
	for pos < len(raw) && raw[pos] != '{' {
		pos++
	}
	if pos >= len(raw) {
		// malformed, fallback to full parse
		n.mu.Unlock()
		n.lazyParse()
		return
	}
	pos++ // skip '{'

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

	// Use direct string comparison without []byte conversion

	for pos < len(raw) {
		skipWS()
		if pos >= len(raw) {
			break
		}
		if raw[pos] == '}' {
			// end of object
			break
		}
		// expect quoted key
		if raw[pos] != '"' {
			// malformed or unexpected token, fallback to full parse
			n.mu.Unlock()
			n.lazyParse()
			return
		}
		// find key end
		keyEnd := findMatchingQuote(raw, pos)
		if keyEnd == -1 {
			n.mu.Unlock()
			n.lazyParse()
			return
		}
		keyRaw := raw[pos+1 : keyEnd]

		// Use direct string/byte comparison without allocation
		var keyMatches bool
		var keyStr string
		if bytes.IndexByte(keyRaw, '\\') == -1 {
			// No escapes, direct comparison and zero-copy string conversion
			keyMatches = compareStringBytes(key, keyRaw)
			if keyMatches && len(keyRaw) > 0 {
				keyStr = unsafe.String(&keyRaw[0], len(keyRaw))
			}
		} else {
			// Has escapes, need to unescape for comparison
			keyUnesc, err := unescape(keyRaw)
			if err != nil {
				n.mu.Unlock()
				n.lazyParse()
				return
			}
			keyMatches = compareStringBytes(key, keyUnesc)
			if keyMatches && len(keyUnesc) > 0 {
				keyStr = unsafe.String(&keyUnesc[0], len(keyUnesc))
			}
		}

		pos = keyEnd + 1
		skipWS()
		if pos >= len(raw) || raw[pos] != ':' {
			n.mu.Unlock()
			n.lazyParse()
			return
		}
		pos++ // skip ':'
		skipWS()
		if pos >= len(raw) {
			n.mu.Unlock()
			n.lazyParse()
			return
		}

		// determine value bounds
		var valEnd int
		switch raw[pos] {
		case '{':
			valEnd = findMatchingBrace(raw, pos)
		case '[':
			valEnd = findMatchingBracket(raw, pos)
		case '"':
			valEnd = findMatchingQuote(raw, pos)
		default:
			valEnd = findValueEnd(raw, pos)
		}
		if valEnd == -1 {
			n.mu.Unlock()
			n.lazyParse()
			return
		}

		// If this is the key we want, parse only the value segment and attach
		if keyMatches {
			segment := raw[pos : valEnd+1]

			// Fast path for string values to avoid parser creation overhead
			var child core.Node
			if len(segment) >= 2 && segment[0] == '"' && segment[len(segment)-1] == '"' {
				// String value - use optimized string node creation
				needsUnescape := bytes.IndexByte(segment[1:len(segment)-1], '\\') != -1
				child = NewRawStringNode(n, segment, 1, len(segment)-1, needsUnescape, n.funcs)
			} else {
				// Other values - use parser
				p := newParser(segment, n.funcs)
				child = p.doParse(n)
			}

			if child == nil || !child.IsValid() {
				if child != nil {
					n.err = child.Error()
					n.mu.Unlock()
					return
				}
				n.mu.Unlock()
				n.lazyParse()
				return
			}
			// ensure parent pointer is correct
			if bn, ok := child.(*baseNode); ok {
				bn.parent = n
			} else if inode, ok := child.(interface{ setParent(core.Node) }); ok {
				inode.setParent(n)
			}
			if n.value == nil {
				n.value = make(map[string]core.Node, 4) // Pre-allocate reasonable capacity
			}
			n.value[keyStr] = child
			n.mu.Unlock()
			return
		}

		// move position to after the value
		pos = valEnd + 1
		skipWS()
		if pos < len(raw) && raw[pos] == ',' {
			pos++
			continue
		} else if pos < len(raw) && raw[pos] == '}' {
			break
		} else {
			// malformed -> fallback
			n.mu.Unlock()
			n.lazyParse()
			return
		}
	}

	// key not found in object, fallback to full parse to establish state
	n.mu.Unlock()
	n.lazyParse()
}

func (n *objectNode) addChild(key string, child core.Node) {
	if n.value == nil {
		n.value = make(map[string]core.Node)
	}

	// Instead of type assertion, we use the Parent() and SetParent() pattern
	// All node types embed baseNode which has parent field
	if bn, ok := child.(*baseNode); ok {
		bn.parent = n
	} else if inode, ok := child.(interface{ setParent(core.Node) }); ok {
		inode.setParent(n)
	}

	n.value[key] = child
}
