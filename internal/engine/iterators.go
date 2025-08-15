package engine

import (
	"fmt"
	"sort"

	"github.com/474420502/xjson/internal/core"
)

// ObjectIter provides a lazy iterator over object key/value pairs.
type ObjectIter interface {
	Next() bool
	KeyRaw() []byte
	ValueRaw() []byte
	ParseValue() core.Node
	Err() error
}

// ArrayIter provides a lazy iterator over array elements.
type ArrayIter interface {
	Next() bool
	Index() int
	ValueRaw() []byte
	ParseValue() core.Node
	Err() error
}

// objectIterator scans an object's raw bytes without creating child nodes.
type objectIterator struct {
	node *objectNode
	// raw scanning state
	raw     []byte
	pos     int
	rawMode bool
	// parsed-mode state
	keys []string
	idx  int
	// current item
	curKey   string
	valStart int
	valEnd   int
	err      error
}

// arrayIterator scans an array's raw bytes without creating child nodes.
type arrayIterator struct {
	node    *arrayNode
	raw     []byte
	pos     int
	rawMode bool
	// parsed-mode
	idx      int
	curIndex int
	// current element bounds
	valStart int
	valEnd   int
	err      error
}

// Iter returns an ObjectIter for the objectNode.
func (n *objectNode) Iter() ObjectIter {
	if n == nil {
		return &objectIterator{err: fmt.Errorf("nil node")}
	}
	if n.err != nil {
		return &objectIterator{err: n.err}
	}
	// If node is dirty or has no raw, fall back to parsed mode
	if n.isDirty || len(n.raw) == 0 {
		// prepare keys sorted
		if n.sortedKeys == nil {
			keys := make([]string, 0, len(n.value))
			for k := range n.value {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			n.sortedKeys = keys
		}
		return &objectIterator{node: n, rawMode: false, keys: n.sortedKeys, idx: -1}
	}
	return &objectIterator{node: n, rawMode: true, raw: n.raw, pos: 0, idx: -1}
}

// Next advances the object iterator to the next key/value pair.
func (it *objectIterator) Next() bool {
	if it.err != nil {
		return false
	}
	if !it.rawMode {
		it.idx++
		if it.idx >= len(it.keys) {
			return false
		}
		it.curKey = it.keys[it.idx]
		it.valStart, it.valEnd = 0, 0
		return true
	}
	raw := it.raw
	var pos int
	// only locate opening brace on first iteration
	if it.pos == 0 {
		pos = 0
		for pos < len(raw) && raw[pos] != '{' {
			pos++
		}
		if pos >= len(raw) {
			it.err = fmt.Errorf("malformed object")
			return false
		}
		pos++ // skip '{'
	} else {
		pos = it.pos
	}
	// helper
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
	// if pos cached from previous iteration, use it
	for pos < len(raw) {
		skipWS()
		if pos >= len(raw) {
			it.pos = pos
			return false
		}
		if raw[pos] == '}' {
			it.pos = pos
			return false
		}
		if raw[pos] != '"' {
			it.err = fmt.Errorf("unexpected token in object at pos %d", pos)
			return false
		}
		keyEnd := findMatchingQuote(raw, pos)
		if keyEnd == -1 {
			it.err = fmt.Errorf("unterminated key string")
			return false
		}
		keyRaw := raw[pos+1 : keyEnd]
		keyUnesc, err := unescape(keyRaw)
		if err != nil {
			it.err = err
			return false
		}
		keyStr := string(keyUnesc)
		pos = keyEnd + 1
		skipWS()
		if pos >= len(raw) || raw[pos] != ':' {
			it.err = fmt.Errorf("missing ':' after key")
			return false
		}
		pos++
		skipWS()
		if pos >= len(raw) {
			it.err = fmt.Errorf("unexpected end after ':'")
			return false
		}
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
			it.err = fmt.Errorf("unterminated value for key %s", keyStr)
			return false
		}
		it.curKey = keyStr
		it.valStart = pos
		it.valEnd = valEnd
		// advance pos to after value and optional comma
		pos = valEnd + 1
		skipWS()
		if pos < len(raw) && raw[pos] == ',' {
			pos++
		}
		it.pos = pos
		return true
	}
	it.pos = pos
	return false
}

func (it *objectIterator) KeyRaw() []byte {
	if it == nil || it.err != nil {
		return nil
	}
	if it.rawMode {
		return []byte(it.curKey)
	}
	return []byte(it.curKey)
}

func (it *objectIterator) ValueRaw() []byte {
	if it == nil || it.err != nil {
		return nil
	}
	if it.rawMode {
		if it.valStart < 0 || it.valEnd < it.valStart {
			return nil
		}
		return it.raw[it.valStart : it.valEnd+1]
	}
	// parsed mode
	if it.node == nil {
		return nil
	}
	if v, ok := it.node.value[it.curKey]; ok {
		return []byte(v.String())
	}
	return nil
}

// ParseValue parses the current value and returns a core.Node. It will cache the child
// on the parent object when in raw mode.
func (it *objectIterator) ParseValue() core.Node {
	if it.err != nil {
		return newInvalidNode(it.err)
	}
	if !it.rawMode {
		if it.node == nil {
			return newInvalidNode(fmt.Errorf("nil node"))
		}
		if c, ok := it.node.value[it.curKey]; ok {
			return c
		}
		return newInvalidNode(fmt.Errorf("key not found: %s", it.curKey))
	}
	segment := it.raw[it.valStart : it.valEnd+1]
	p := newParser(segment, it.node.GetFuncs())
	child := p.doParse(it.node)
	if child == nil || !child.IsValid() {
		if child != nil {
			return child
		}
		return newInvalidNode(fmt.Errorf("failed to parse segment"))
	}
	// set parent and optionally cache on parent
	if bn, ok := child.(*baseNode); ok {
		bn.parent = it.node
	} else if inode, ok := child.(interface{ setParent(core.Node) }); ok {
		inode.setParent(it.node)
	}
	// cache into parent map if available and parent not parsed or dirty
	it.node.mu.Lock()
	if !it.node.parsed.Load() && !it.node.isDirty {
		if it.node.value == nil {
			it.node.value = make(map[string]core.Node)
		}
		it.node.value[it.curKey] = child
	}
	it.node.mu.Unlock()
	return child
}

func (it *objectIterator) Err() error { return it.err }

// Array iterator implementation
func (n *arrayNode) Iter() ArrayIter {
	if n == nil {
		return &arrayIterator{err: fmt.Errorf("nil node")}
	}
	if n.err != nil {
		return &arrayIterator{err: n.err}
	}
	if n.isDirty || len(n.raw) == 0 {
		return &arrayIterator{node: n, rawMode: false, idx: -1, curIndex: -1}
	}
	return &arrayIterator{node: n, raw: n.raw, pos: 0, rawMode: true, idx: -1, curIndex: -1}
}

func (it *arrayIterator) Next() bool {
	if it.err != nil {
		return false
	}
	if !it.rawMode {
		it.curIndex++
		if it.curIndex >= len(it.node.value) {
			return false
		}
		return true
	}
	raw := it.raw
	var pos int
	if it.pos == 0 {
		pos = 0
		for pos < len(raw) && raw[pos] != '[' {
			pos++
		}
		if pos >= len(raw) {
			it.err = fmt.Errorf("malformed array")
			return false
		}
		pos++ // skip '['
	} else {
		pos = it.pos
	}
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
	if it.pos != 0 {
		pos = it.pos
	}
	curIndex := 0
	// If we have already appended some elements in previous iterations, count them
	// We approximate by keeping curIndex from previous
	curIndex = it.curIndex + 1
	for pos < len(raw) {
		skipWS()
		if pos >= len(raw) {
			it.pos = pos
			return false
		}
		if raw[pos] == ']' {
			it.pos = pos
			return false
		}
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
			it.err = fmt.Errorf("unterminated array element")
			return false
		}
		it.valStart = elemStart
		it.valEnd = elemEnd
		it.curIndex = curIndex
		// advance
		pos = elemEnd + 1
		skipWS()
		if pos < len(raw) && raw[pos] == ',' {
			pos++
		}
		it.pos = pos
		return true
	}
	it.pos = pos
	return false
}

func (it *arrayIterator) Index() int { return it.curIndex }

func (it *arrayIterator) ValueRaw() []byte {
	if it.err != nil {
		return nil
	}
	if !it.rawMode {
		if it.node == nil || it.curIndex < 0 || it.curIndex >= len(it.node.value) {
			return nil
		}
		return []byte(it.node.value[it.curIndex].String())
	}
	if it.valStart < 0 || it.valEnd < it.valStart {
		return nil
	}
	return it.raw[it.valStart : it.valEnd+1]
}

func (it *arrayIterator) ParseValue() core.Node {
	if it.err != nil {
		return newInvalidNode(it.err)
	}
	if !it.rawMode {
		if it.node == nil {
			return newInvalidNode(fmt.Errorf("nil node"))
		}
		if it.curIndex < 0 || it.curIndex >= len(it.node.value) {
			return newInvalidNode(fmt.Errorf("index out of range: %d", it.curIndex))
		}
		return it.node.value[it.curIndex]
	}
	segment := it.raw[it.valStart : it.valEnd+1]
	p := newParser(segment, it.node.GetFuncs())
	child := p.doParse(it.node)
	if child == nil || !child.IsValid() {
		if child != nil {
			return child
		}
		return newInvalidNode(fmt.Errorf("failed to parse array element"))
	}
	if bn, ok := child.(*baseNode); ok {
		bn.parent = it.node
	} else if inode, ok := child.(interface{ setParent(core.Node) }); ok {
		inode.setParent(it.node)
	}
	// cache newly parsed child for future parsed-mode iterations
	it.node.mu.Lock()
	if !it.node.parsed.Load() && !it.node.isDirty {
		// ensure value slice capacity
		if it.curIndex >= len(it.node.value) {
			// extend with nils up to index
			for len(it.node.value) <= it.curIndex {
				it.node.value = append(it.node.value, nil)
			}
		}
		it.node.value[it.curIndex] = child
	}
	it.node.mu.Unlock()
	return child
}

func (it *arrayIterator) Err() error { return it.err }
