package engine

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/474420502/xjson/internal/core"
)

// newRawBoolNode builds a bool node using provided raw slice and value without extra parsing
func newRawBoolNode(parent core.Node, raw []byte, val bool, funcs *map[string]core.UnaryPathFunc) core.Node {
	n := NewBoolNode(parent, val, funcs).(*boolNode)
	n.raw = raw
	n.start = 0
	n.end = len(raw)
	return n
}

// newRawNullNode builds a null node using provided raw slice without extra parsing
func newRawNullNode(parent core.Node, raw []byte, funcs *map[string]core.UnaryPathFunc) core.Node {
	n := NewNullNode(parent, funcs).(*nullNode)
	n.raw = raw
	n.start = 0
	n.end = len(raw)
	return n
}

func recursiveSearch(node core.Node, key string) core.Node {
	// Try optimized raw-byte recursive scan when possible to avoid parsing full subtrees.
	results := make([]core.Node, 0)

	// helper to append result
	appendResult := func(n core.Node) {
		if n != nil && n.IsValid() {
			results = append(results, n)
		}
	}

	// recursive byte-scan for raw nodes
	var recursiveScanBytes func(data []byte, funcs *map[string]core.UnaryPathFunc)
	recursiveScanBytes = func(data []byte, funcs *map[string]core.UnaryPathFunc) {
		if len(data) == 0 {
			return
		}
		// skip whitespace
		i := 0
		for i < len(data) && (data[i] == ' ' || data[i] == '\n' || data[i] == '\r' || data[i] == '\t') {
			i++
		}
		if i >= len(data) {
			return
		}
		switch data[i] {
		case '{':
			// scan object fields
			pos := i
			objEnd := findMatchingBrace(data, pos)
			if objEnd == -1 {
				return
			}
			// lazily allocate a temporary parent node only when we need it.
			// We keep the raw parent slice so the parser can set correct
			// Parent() pointers on children when a match is found. This
			// avoids allocating for every scanned object while preserving
			// parent linkage when required by callers/tests.
			parentRaw := data[pos : objEnd+1]
			var parentNode core.Node = nil
			pos++ // skip '{'
			skipWS := func() {
				for pos < len(data) {
					c := data[pos]
					if c == ' ' || c == '\n' || c == '\r' || c == '\t' {
						pos++
						continue
					}
					break
				}
			}
			for pos < len(data) {
				skipWS()
				if pos >= len(data) || data[pos] == '}' {
					break
				}
				if data[pos] != '"' {
					// malformed, abort
					return
				}
				keyEnd := findMatchingQuote(data, pos)
				if keyEnd == -1 {
					return
				}
				keyRaw := data[pos+1 : keyEnd]
				keyUnesc, err := unescape(keyRaw)
				if err != nil {
					return
				}
				keyStr := string(keyUnesc)
				pos = keyEnd + 1
				skipWS()
				if pos >= len(data) || data[pos] != ':' {
					return
				}
				pos++
				skipWS()
				if pos >= len(data) {
					return
				}
				var valEnd int
				switch data[pos] {
				case '{':
					valEnd = findMatchingBrace(data, pos)
				case '[':
					valEnd = findMatchingBracket(data, pos)
				case '"':
					valEnd = findMatchingQuote(data, pos)
				default:
					valEnd = findValueEnd(data, pos)
				}
				if valEnd == -1 {
					return
				}
				// if key matches, parse value and append
				if key == "" || key == keyStr {
					segment := data[pos : valEnd+1]
					// allocate parentNode lazily so Parent() can be set on child
					if parentNode == nil {
						parentNode = NewObjectNode(nil, parentRaw, funcs)
					}
					p := newParser(segment, funcs)
					// parse with parentNode so that Parent() works for the child
					child := p.doParse(parentNode)
					appendResult(child)
				}
				// recurse into value if it's a composite
				first := getFirstNonWhitespaceChar(data[pos : valEnd+1])
				if first == '{' || first == '[' {
					recursiveScanBytes(data[pos:valEnd+1], funcs)
				}
				pos = valEnd + 1
				skipWS()
				if pos < len(data) && data[pos] == ',' {
					pos++
					continue
				}
				if pos < len(data) && data[pos] == '}' {
					break
				}
				// malformed -> abort
				return
			}
		case '[':
			// scan array elements
			pos := i
			pos++ // skip '['
			skipWS := func() {
				for pos < len(data) {
					c := data[pos]
					if c == ' ' || c == '\n' || c == '\r' || c == '\t' {
						pos++
						continue
					}
					break
				}
			}
			for pos < len(data) {
				skipWS()
				if pos >= len(data) || data[pos] == ']' {
					break
				}
				var elemEnd int
				switch data[pos] {
				case '{':
					elemEnd = findMatchingBrace(data, pos)
				case '[':
					elemEnd = findMatchingBracket(data, pos)
				case '"':
					elemEnd = findMatchingQuote(data, pos)
				default:
					elemEnd = findValueEnd(data, pos)
				}
				if elemEnd == -1 {
					return
				}
				// recurse into element
				first := getFirstNonWhitespaceChar(data[pos : elemEnd+1])
				if first == '{' || first == '[' {
					recursiveScanBytes(data[pos:elemEnd+1], funcs)
				}
				pos = elemEnd + 1
				skipWS()
				if pos < len(data) && data[pos] == ',' {
					pos++
					continue
				}
				if pos < len(data) && data[pos] == ']' {
					break
				}
				return
			}
		default:
			return
		}
	}

	// If start node can be scanned as raw, prefer that.
	if on, ok := node.(*objectNode); ok && !on.parsed.Load() && !on.isDirty && len(on.raw) > 0 {
		recursiveScanBytes(on.raw, on.GetFuncs())
		arr := NewArrayNode(nil, nil, node.GetFuncs())
		arr.(*arrayNode).value = results
		return arr
	}
	if an, ok := node.(*arrayNode); ok && !an.parsed.Load() && !an.isDirty && len(an.raw) > 0 {
		recursiveScanBytes(an.raw, an.GetFuncs())
		arr := NewArrayNode(nil, nil, node.GetFuncs())
		arr.(*arrayNode).value = results
		return arr
	}

	// fallback to original behavior for parsed/dirty nodes
	var walk func(core.Node)
	walk = func(n core.Node) {
		if !n.IsValid() {
			return
		}
		switch n.Type() {
		case core.Object:
			if key != "" {
				c := n.Get(key)
				if c != nil && c.IsValid() {
					results = append(results, c)
				}
			}
			n.ForEach(func(_ interface{}, v core.Node) {
				walk(v)
			})
		case core.Array:
			n.ForEach(func(_ interface{}, v core.Node) {
				walk(v)
			})
		}
	}
	walk(node)
	arr := NewArrayNode(nil, nil, node.GetFuncs())
	arr.(*arrayNode).value = results
	return arr
}

// newInvalidNode creates a new invalid node with the given error
func applySimpleQuery(start core.Node, path string) core.Node {
	// Try to get cached result first so repeated identical queries can bypass
	// both path scanning and per-segment object lookups.
	if enableQueryCache {
		if bn, ok := start.(interface {
			getCachedQueryResult(string) (core.Node, bool)
		}); ok {
			if cachedResult, exists := bn.getCachedQueryResult(path); exists {
				return cachedResult
			}
		}
	}
	// First, try a composite fast-path that supports keys and [index] without allocations
	if res := tryFastBracketQuery(start, path); res != nil {
		if enableQueryCache {
			if bn, ok := start.(interface{ setCachedQueryResult(string, core.Node) }); ok {
				bn.setCachedQueryResult(path, res)
			}
		}
		return res
	}
	// DEBUG: observe query flow
	// Try a conservative fast-path for simple slash-separated key paths
	if res := tryFastSlashQuery(start, path); res != nil {
		if enableQueryCache {
			if bn, ok := start.(interface{ setCachedQueryResult(string, core.Node) }); ok {
				bn.setCachedQueryResult(path, res)
			}
		}
		return res
	}

	tokens, err := ParseQuery(path)

	if err != nil {
		return newInvalidNode(err)
	}

	cur := start
	for _, t := range tokens {

		if !cur.IsValid() {
			return cur
		}

		switch t.Op {
		case OpKey:
			key := t.Value.(string)
			if a, ok := cur.(*arrayNode); ok {
				// Try to use iterator to avoid fully parsing the array
				it := a.Iter()
				results := make([]core.Node, 0)
				for it.Next() {
					// prefer ParseValue() which works for parsed and raw modes
					if elem := it.ParseValue(); elem.IsValid() {
						if elem.Type() == core.Object {
							res := elem.Get(key)
							if res.IsValid() {
								results = append(results, res)
							}
						}
					}
				}
				if len(results) == 0 {
					return newInvalidNode(fmt.Errorf("key '%s' not found in any array element", key))
				}
				newArr := NewArrayNode(a, nil, a.GetFuncs())
				newArr.(*arrayNode).value = results
				newArr.(*arrayNode).isDirty = true
				cur = newArr
			} else if o, ok := cur.(*objectNode); ok {
				cur = o.Get(key)
			} else {
				return newInvalidNode(fmt.Errorf("not an object for key access '%s' on node type %v", key, cur.Type()))
			}
		case OpIndex:
			if a, ok := cur.(*arrayNode); ok {
				cur = a.Index(t.Value.(int))
			} else {
				return newInvalidNode(fmt.Errorf("not an array for index access: %v", cur.Raw()))
			}
		case OpSlice:
			if a, ok := cur.(*arrayNode); ok {
				a.lazyParse() // 确保在访问前解析
				s := t.Value.(slice)
				arrLen := len(a.value)

				start := s.Start
				if start < 0 {
					start = arrLen + start
				}
				if start < 0 {
					start = 0
				} else if start > arrLen {
					start = arrLen
				}

				end := s.End
				if s.End == -1 {
					end = arrLen
				} else if end < 0 {
					end = arrLen + end
				}

				if end > arrLen {
					end = arrLen
				}

				if start > end {
					start = end
				}

				newArr := NewArrayNode(a, nil, a.GetFuncs())
				newArr.(*arrayNode).value = a.value[start:end]
				newArr.(*arrayNode).isDirty = true
				cur = newArr
			} else {
				return newInvalidNode(fmt.Errorf("not an array for slice access"))
			}
		case OpWildcard:
			results := make([]core.Node, 0)
			if o, ok := cur.(*objectNode); ok {
				// attempt raw-mode iteration to avoid full parse
				it := o.Iter()
				for it.Next() {
					if child := it.ParseValue(); child.IsValid() {
						results = append(results, child)
					}
				}
				if err := it.Err(); err != nil {
					// fallback to full parse if iterator failed
					o.lazyParse()
					for _, v := range o.value {
						results = append(results, v)
					}
				}
			} else if a, ok := cur.(*arrayNode); ok {
				it := a.Iter()
				for it.Next() {
					if child := it.ParseValue(); child.IsValid() {
						results = append(results, child)
					}
				}
				if err := it.Err(); err != nil {
					a.lazyParse()
					results = a.value
				}
			}
			newArr := NewArrayNode(cur, nil, cur.GetFuncs())
			newArr.(*arrayNode).value = results
			newArr.(*arrayNode).isDirty = true
			cur = newArr
		case OpFunc:
			name := t.Value.(string)
			cur = cur.CallFunc(name)
		case OpRecursive:
			key := t.Value.(string)
			cur = recursiveSearch(cur, key)
		case OpParent:
			if p := cur.Parent(); p != nil && p != cur {
				cur = p
			} else {
				// No parent available or already at root, return invalid node for navigation above root
				return newInvalidNode(fmt.Errorf("no parent node available for node %v", cur.Raw()))
			}
		default:
			return newInvalidNode(fmt.Errorf("unsupported op"))
		}
		if cur == nil {
			return newInvalidNode(fmt.Errorf("nil during query"))
		}
	}

	// Cache the result (optional)
	if enableQueryCache {
		if bn, ok := start.(interface{ setCachedQueryResult(string, core.Node) }); ok {
			bn.setCachedQueryResult(path, cur)
		}
	}

	return cur
}

func directObjectChild(o *objectNode, key string) core.Node {
	if child, ok := o.value[key]; ok {
		return child
	}
	return sharedInvalidNode()
}

func directArrayChild(a *arrayNode, idx int) core.Node {
	if idx < 0 {
		idx = len(a.value) + idx
	}
	if idx >= 0 && idx < len(a.value) {
		return a.value[idx]
	}
	return newInvalidNode(fmt.Errorf("index out of bounds: %d", idx))
}

func fastConstructObjectChild(o *objectNode, segment []byte) core.Node {
	var child core.Node
	if len(segment) >= 2 && segment[0] == '"' && segment[len(segment)-1] == '"' {
		needsUnescape := bytes.IndexByte(segment[1:len(segment)-1], '\\') != -1
		child = NewRawStringNode(o, segment, 1, len(segment)-1, needsUnescape, o.funcs)
	} else if len(segment) > 0 && (segment[0] == '{' || segment[0] == '[') {
		if segment[0] == '{' {
			child = NewObjectNode(o, segment, o.funcs)
		} else {
			child = NewArrayNode(o, segment, o.funcs)
		}
	} else {
		if len(segment) > 0 {
			switch segment[0] {
			case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				child = NewNumberNode(o, segment, o.funcs)
			case 't', 'f':
				child = newRawBoolNode(o, segment, segment[0] == 't', o.funcs)
			case 'n':
				child = newRawNullNode(o, segment, o.funcs)
			default:
				p := newParser(segment, o.funcs)
				child = p.doParse(o)
			}
		} else {
			p := newParser(segment, o.funcs)
			child = p.doParse(o)
		}
	}
	if child == nil || !child.IsValid() {
		return nil
	}
	if bn, ok := child.(*baseNode); ok {
		bn.parent = o
	} else if inode, ok := child.(interface{ setParent(core.Node) }); ok {
		inode.setParent(o)
	}
	return child
}

func fastScanObjectChildLocked(o *objectNode, key string) (core.Node, bool, bool) {
	if child, ok := o.value[key]; ok {
		return child, true, true
	}

	raw := o.raw
	pos := 0
	for pos < len(raw) && raw[pos] != '{' {
		pos++
	}
	if pos >= len(raw) {
		return nil, false, false
	}
	pos++

	skipWS := func() {
		for pos < len(raw) {
			c := raw[pos]
			if c == ' ' || c == '\n' || c == '\r' || c == '\t' {
				pos++
			} else {
				break
			}
		}
	}

	for pos < len(raw) {
		skipWS()
		if pos >= len(raw) || raw[pos] == '}' {
			break
		}
		if raw[pos] != '"' {
			return nil, false, false
		}
		keyEnd := findMatchingQuote(raw, pos)
		if keyEnd == -1 {
			return nil, false, false
		}
		keyRaw := raw[pos+1 : keyEnd]
		match := bytes.IndexByte(keyRaw, '\\') == -1 && compareStringBytes(key, keyRaw)
		pos = keyEnd + 1
		skipWS()
		if pos >= len(raw) || raw[pos] != ':' {
			return nil, false, false
		}
		pos++
		skipWS()
		if pos >= len(raw) {
			return nil, false, false
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
			return nil, false, false
		}

		if match {
			child := fastConstructObjectChild(o, raw[pos:valEnd+1])
			if child == nil {
				return nil, false, false
			}
			if o.value == nil {
				o.value = make(map[string]core.Node, 4)
			}
			o.value[key] = child
			return child, true, true
		}

		pos = valEnd + 1
		skipWS()
		if pos < len(raw) && raw[pos] == ',' {
			pos++
			continue
		}
		if pos < len(raw) && raw[pos] == '}' {
			break
		}
		return nil, false, false
	}

	return nil, false, true
}

// tryFastSlashQuery attempts a zero-allocation fast path for very simple
// slash-separated key lookups like "a/b/c" when nodes are still in raw form.
// Returns nil if the path is not eligible or fast path couldn't resolve.
func tryFastSlashQuery(start core.Node, path string) core.Node {
	// DEBUG LOGGING - remove after diagnosis
	// fmt.Printf("tryFastSlashQuery path=%q startType=%v\n", path, start.Type())
	if path == "" || strings.ContainsAny(path, "[]*@.") || strings.Contains(path, "//") || strings.HasPrefix(path, "../") {
		return nil
	}

	// Manual scan of path components to avoid allocations from TrimLeft and Split.
	// This preserves behavior for leading/trailing slashes.
	p := 0
	for p < len(path) && path[p] == '/' {
		p++
	}
	if p >= len(path) {
		return start
	}
	cur := start
	partIndex := 0
	for p < len(path) {
		// find next separator
		q := p
		for q < len(path) && path[q] != '/' {
			q++
		}
		part := path[p:q]
		if !cur.IsValid() {
			return newInvalidNode(fmt.Errorf("invalid during fast-path"))
		}
		if part == "" {
			// skip empty segments (shouldn't happen after trimming, but be defensive)
			continue
		}
		// Only handle object nodes in raw/unparsed state or already parsed
		if o, ok := cur.(*objectNode); ok && !o.isDirty {

			// If parsed already, use standard Get
			if o.parsed.Load() {
				cur = o.Get(part)
				// advance to next part
				if q >= len(path) {
					return cur
				}
				partIndex++
				p = q + 1
				for p < len(path) && path[p] == '/' {
					p++
				}
				continue
			}
			// raw scan for key
			raw := o.raw
			// acquire lock similar to lazyParsePath to safely inspect/modify value
			o.mu.Lock()
			// re-check parsed and existing map under lock
			if o.parsed.Load() {
				o.mu.Unlock()
				cur = o.Get(part)
				continue
			}
			if o.value != nil {
				if child, ok := o.value[part]; ok {
					o.mu.Unlock()
					cur = child
					// advance to next part
					if q >= len(path) {
						return cur
					}
					partIndex++
					p = q + 1
					for p < len(path) && path[p] == '/' {
						p++
					}
					continue
				}
			}
			if len(raw) == 0 {
				return nil
			}
			// find key in raw using similar logic to lazyParsePath
			pos := 0
			for pos < len(raw) && raw[pos] != '{' {
				pos++
			}
			if pos >= len(raw) {
				return nil
			}
			pos++
			found := false
			for pos < len(raw) {
				// skip ws
				for pos < len(raw) && (raw[pos] == ' ' || raw[pos] == '\n' || raw[pos] == '\r' || raw[pos] == '\t') {
					pos++
				}
				if pos >= len(raw) || raw[pos] == '}' {
					break
				}
				if raw[pos] != '"' {
					return nil
				}
				keyEnd := findMatchingQuote(raw, pos)
				if keyEnd == -1 {
					return nil
				}
				keyRaw := raw[pos+1 : keyEnd]
				// compare quickly without unescaping
				if bytes.IndexByte(keyRaw, '\\') == -1 && compareStringBytes(part, keyRaw) {
					// found — parse value segment and set cur
					pos = keyEnd + 1
					// skip ws and colon
					for pos < len(raw) && (raw[pos] == ' ' || raw[pos] == '\n' || raw[pos] == '\r' || raw[pos] == '\t') {
						pos++
					}
					if pos >= len(raw) || raw[pos] != ':' {
						return nil
					}
					pos++
					for pos < len(raw) && (raw[pos] == ' ' || raw[pos] == '\n' || raw[pos] == '\r' || raw[pos] == '\t') {
						pos++
					}
					if pos >= len(raw) {
						return nil
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
						return nil
					}
					segment := raw[pos : valEnd+1]
					var child core.Node
					if len(segment) >= 2 && segment[0] == '"' && segment[len(segment)-1] == '"' {
						needsUnescape := bytes.IndexByte(segment[1:len(segment)-1], '\\') != -1
						child = NewRawStringNode(o, segment, 1, len(segment)-1, needsUnescape, o.funcs)
					} else {
						p := newParser(segment, o.funcs)
						child = p.doParse(o)
					}
					if child == nil || !child.IsValid() {
						o.mu.Unlock()
						return nil
					}
					// ensure parent pointer is correct
					if bn, ok := child.(*baseNode); ok {
						bn.parent = o
					} else if inode, ok := child.(interface{ setParent(core.Node) }); ok {
						inode.setParent(o)
					}
					if o.value == nil {
						o.value = make(map[string]core.Node, 4)
					}
					// store into parent so subsequent operations and String() see it
					o.value[part] = child
					o.mu.Unlock()
					if q >= len(path) {
						return child
					}
					cur = child
					found = true
					break
				}
				// skip to after value
				pos = keyEnd + 1
				// skip ws and colon
				for pos < len(raw) && (raw[pos] == ' ' || raw[pos] == '\n' || raw[pos] == '\r' || raw[pos] == '\t') {
					pos++
				}
				if pos >= len(raw) || raw[pos] != ':' {
					return nil
				}
				pos++
				for pos < len(raw) && (raw[pos] == ' ' || raw[pos] == '\n' || raw[pos] == '\r' || raw[pos] == '\t') {
					pos++
				}
				if pos >= len(raw) {
					return nil
				}
				var valEnd2 int
				switch raw[pos] {
				case '{':
					valEnd2 = findMatchingBrace(raw, pos)
				case '[':
					valEnd2 = findMatchingBracket(raw, pos)
				case '"':
					valEnd2 = findMatchingQuote(raw, pos)
				default:
					valEnd2 = findValueEnd(raw, pos)
				}
				if valEnd2 == -1 {
					o.mu.Unlock()
					return nil
				}
				pos = valEnd2 + 1
				// skip comma or end
				for pos < len(raw) && (raw[pos] == ' ' || raw[pos] == '\n' || raw[pos] == '\r' || raw[pos] == '\t') {
					pos++
				}
				if pos < len(raw) && raw[pos] == ',' {
					pos++
					continue
				}
				if pos < len(raw) && raw[pos] == '}' {
					break
				}
				o.mu.Unlock()
				return nil
			}
			if !found {
				o.mu.Unlock()
				return sharedInvalidNode()
			}
			partIndex++
			if q >= len(path) {
				break
			}
			p = q + 1
			// skip consecutive slashes
			for p < len(path) && path[p] == '/' {
				p++
			}
		} else {
			// not object node or not supported fast path, bail out
			return nil
		}
	}
	return nil
}

// tryRawDirectPath walks the path directly over raw JSON bytes, avoiding
// building intermediate nodes. It supports segments like a/b/c and [idx]
// after a key, e.g., a/b[0]/c. It returns nil if the path isn't eligible.
func tryRawDirectPath(start core.Node, path string) core.Node {
	if path == "" || strings.ContainsAny(path, "*@.") || strings.Contains(path, "//") || strings.HasPrefix(path, "../") {
		return nil
	}
	// root must be object or array with raw data and not parsed/dirty
	var curRaw []byte
	var isObj bool
	switch n := start.(type) {
	case *objectNode:
		if n.isDirty || n.parsed.Load() || len(n.raw) == 0 {
			return nil
		}
		curRaw = n.raw
		isObj = true
	case *arrayNode:
		if n.isDirty || n.parsed.Load() || len(n.raw) == 0 {
			return nil
		}
		curRaw = n.raw
		isObj = false
	default:
		return nil
	}

	// path iterator
	i := 0
	// skip leading '/'
	for i < len(path) && path[i] == '/' {
		i++
	}
	if i >= len(path) {
		// return root as-is
		if isObj {
			return start
		}
		return start
	}

	// helper: find value by key in object raw at top level
	scanObj := func(raw []byte, key string) ([]byte, bool) {
		// find '{'
		pos := 0
		for pos < len(raw) && raw[pos] != '{' {
			pos++
		}
		if pos >= len(raw) {
			return nil, false
		}
		pos++
		skipWS := func() {
			for pos < len(raw) {
				c := raw[pos]
				if c == ' ' || c == '\n' || c == '\r' || c == '\t' {
					pos++
				} else {
					break
				}
			}
		}
		for pos < len(raw) {
			skipWS()
			if pos >= len(raw) || raw[pos] == '}' {
				break
			}
			if raw[pos] != '"' {
				return nil, false
			}
			keyEnd := findMatchingQuote(raw, pos)
			if keyEnd == -1 {
				return nil, false
			}
			keyRaw := raw[pos+1 : keyEnd]
			// compare quickly if no escapes
			match := bytes.IndexByte(keyRaw, '\\') == -1 && compareStringBytes(key, keyRaw)
			if !match && bytes.IndexByte(keyRaw, '\\') != -1 {
				un, err := unescape(keyRaw)
				if err == nil {
					match = compareStringBytes(key, un)
				}
			}
			pos = keyEnd + 1
			skipWS()
			if pos >= len(raw) || raw[pos] != ':' {
				return nil, false
			}
			pos++
			skipWS()
			if pos >= len(raw) {
				return nil, false
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
				return nil, false
			}
			if match {
				return raw[pos : valEnd+1], true
			}
			pos = valEnd + 1
			skipWS()
			if pos < len(raw) && raw[pos] == ',' {
				pos++
				continue
			}
			if pos < len(raw) && raw[pos] == '}' {
				break
			}
			return nil, false
		}
		return nil, false
	}

	// helper: find element by index in array raw at top level
	scanArr := func(raw []byte, idx int) ([]byte, bool) {
		// Only non-negative indices supported fast
		if idx < 0 {
			return nil, false
		}
		pos := 0
		for pos < len(raw) && raw[pos] != '[' {
			pos++
		}
		if pos >= len(raw) {
			return nil, false
		}
		pos++
		skipWS := func() {
			for pos < len(raw) {
				c := raw[pos]
				if c == ' ' || c == '\n' || c == '\r' || c == '\t' {
					pos++
				} else {
					break
				}
			}
		}
		cur := 0
		for pos < len(raw) {
			skipWS()
			if pos >= len(raw) {
				break
			}
			if raw[pos] == ']' {
				break
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
				return nil, false
			}
			if cur == idx {
				return raw[elemStart : elemEnd+1], true
			}
			cur++
			pos = elemEnd + 1
			skipWS()
			if pos < len(raw) && raw[pos] == ',' {
				pos++
				continue
			}
			if pos < len(raw) && raw[pos] == ']' {
				break
			}
			// malformed
			return nil, false
		}
		return nil, false
	}

	// Iterate path components
	for i < len(path) {
		if isObj {
			// extract key
			kStart := i
			for i < len(path) && path[i] != '/' && path[i] != '[' {
				i++
			}
			if kStart == i {
				return nil
			}
			key := path[kStart:i]
			seg, ok := scanObj(curRaw, key)
			if !ok {
				return sharedInvalidNode()
			}
			curRaw = seg
			// update container type for next phase
			if len(seg) > 0 && seg[0] == '{' {
				isObj = true
			} else if len(seg) > 0 && seg[0] == '[' {
				isObj = false
			} else {
				isObj = false
			}
		} else {
			// current is array; expect [index]
			if i >= len(path) || path[i] != '[' {
				return nil
			}
			i++
			neg := false
			if i < len(path) && path[i] == '-' {
				neg = true
				i++
			}
			val := 0
			has := false
			for i < len(path) && path[i] >= '0' && path[i] <= '9' {
				has = true
				val = val*10 + int(path[i]-'0')
				i++
			}
			if !has || i >= len(path) || path[i] != ']' {
				return nil
			}
			if neg {
				val = -val
			}
			i++
			seg, ok := scanArr(curRaw, val)
			if !ok {
				return sharedInvalidNode()
			}
			curRaw = seg
			if len(seg) > 0 && seg[0] == '{' {
				isObj = true
			} else if len(seg) > 0 && seg[0] == '[' {
				isObj = false
			} else {
				isObj = false
			}
		}

		// Consume zero or more [index] following
		for i < len(path) && path[i] == '[' {
			i++
			neg := false
			if i < len(path) && path[i] == '-' {
				neg = true
				i++
			}
			val := 0
			has := false
			for i < len(path) && path[i] >= '0' && path[i] <= '9' {
				has = true
				val = val*10 + int(path[i]-'0')
				i++
			}
			if !has || i >= len(path) || path[i] != ']' {
				return nil
			}
			if neg {
				val = -val
			}
			i++
			// we must be indexing into an array
			seg, ok := scanArr(curRaw, val)
			if !ok {
				return sharedInvalidNode()
			}
			curRaw = seg
			if len(seg) > 0 && seg[0] == '{' {
				isObj = true
			} else if len(seg) > 0 && seg[0] == '[' {
				isObj = false
			} else {
				isObj = false
			}
		}

		// If next char is '/', move to next segment
		if i < len(path) {
			if path[i] == '/' {
				for i < len(path) && path[i] == '/' {
					i++
				}
				continue
			}
			// unexpected token
			return nil
		}
		// End of path -> build node from curRaw
		if len(curRaw) == 0 {
			return sharedInvalidNode()
		}
		switch curRaw[0] {
		case '{':
			return NewObjectNode(start, curRaw, start.GetFuncs())
		case '[':
			return NewArrayNode(start, curRaw, start.GetFuncs())
		case '"':
			needsUnescape := bytes.IndexByte(curRaw[1:len(curRaw)-1], '\\') != -1
			return NewRawStringNode(start, curRaw, 1, len(curRaw)-1, needsUnescape, start.GetFuncs())
		default:
			// Primitive: number/bool/null
			c := curRaw[0]
			if (c >= '0' && c <= '9') || c == '-' {
				return NewNumberNode(start, curRaw, start.GetFuncs())
			}
			if len(curRaw) >= 4 && (curRaw[0] == 't' || curRaw[0] == 'f') {
				// true/false
				val := curRaw[0] == 't'
				return newRawBoolNode(start, curRaw, val, start.GetFuncs())
			}
			if len(curRaw) >= 4 && curRaw[0] == 'n' { // null
				return newRawNullNode(start, curRaw, start.GetFuncs())
			}
			// Fallback safety
			p := newParser(curRaw, start.GetFuncs())
			return p.doParse(start)
		}
	}
	return nil
}

// tryFastBracketQuery accelerates queries of the form:
//
//	a/b/c, a/b[0]/c, a/b[10]/c/d, ...
//
// It avoids building token slices and, crucially, constructs composite child
// nodes (objects/arrays) in lazy form from raw segments instead of fully parsing
// them. Returns nil if the path is not eligible or resolution fails.
func tryFastBracketQuery(start core.Node, path string) core.Node {
	if path == "" || strings.HasPrefix(path, "../") || strings.HasPrefix(path, "//") {
		return nil
	}
	// Skip leading slashes
	i := 0
	for i < len(path) && path[i] == '/' {
		i++
	}
	if i >= len(path) {
		return start
	}

	cur := start
	for i < len(path) {
		if !cur.IsValid() {
			return sharedInvalidNode()
		}
		// Expect an object key segment before optional [index] chain
		// Extract key until '/' or '['
		kStart := i
		for i < len(path) && path[i] != '/' && path[i] != '[' {
			switch path[i] {
			case '*', '@', '.':
				return nil
			}
			i++
		}
		if kStart < i { // have a key
			key := path[kStart:i]
			// current must be object
			o, ok := cur.(*objectNode)
			if !ok || o.isDirty {
				return nil
			}
			// If already parsed, use normal Get (fast enough)
			if o.parsed.Load() {
				cur = directObjectChild(o, key)
			} else {
				o.mu.Lock()
				// re-check under lock
				if o.parsed.Load() {
					o.mu.Unlock()
					cur = o.Get(key)
				} else {
					child, found, ok := fastScanObjectChildLocked(o, key)
					o.mu.Unlock()
					if !ok {
						return nil
					}
					if !found {
						return sharedInvalidNode()
					}
					cur = child
				}
			}
		}

		// Handle zero or more [index] after the key
		for i < len(path) && path[i] == '[' {
			// parse integer index
			i++
			if i >= len(path) {
				return nil
			}
			neg := false
			if path[i] == '-' {
				neg = true
				i++
			}
			val := 0
			hasDigit := false
			for i < len(path) && path[i] >= '0' && path[i] <= '9' {
				hasDigit = true
				val = val*10 + int(path[i]-'0')
				i++
			}
			if i >= len(path) || path[i] != ']' || !hasDigit {
				return nil
			}
			if neg {
				val = -val
			}
			i++ // consume ']'

			// current must be array to index
			a, ok := cur.(*arrayNode)
			if !ok {
				return nil
			}
			if a.parsed.Load() || len(a.raw) == 0 {
				cur = directArrayChild(a, val)
			} else {
				cur = a.Index(val)
			}
			if !cur.IsValid() {
				return cur
			}
		}

		// If next is '/', consume and continue to next key
		if i < len(path) {
			if path[i] == '/' {
				if i+1 < len(path) && path[i+1] == '/' {
					return nil
				}
				i++
				// skip consecutive slashes
				for i < len(path) && path[i] == '/' {
					i++
				}
				continue
			}
			// Unexpected token -> bail to generic path
			return nil
		}
		// End of path
		return cur
	}
	return cur
}
