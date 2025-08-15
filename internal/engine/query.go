package engine

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/474420502/xjson/internal/core"
)

// slice represents a slice operation, e.g., [start:end].
type slice struct {
	Start, End int
}

// Op represents a query operation type for the simple query parser here.
type Op int

const (
	OpKey Op = iota
	OpIndex
	OpSlice
	OpFunc
	OpWildcard
	OpRecursive
	OpParent
)

// QueryToken represents a token in the parsed path.
type QueryToken struct {
	Op    Op
	Value interface{}
}

// ParseQuery tokenizes a query path into a sequence of operations.
func ParseQuery(path string) ([]QueryToken, error) {
	if path == "" || path == "/" {
		return []QueryToken{}, nil
	}
	tokens := make([]QueryToken, 0)
	p := 0

	for p < len(path) {
		// Skip leading slashes but handle '//' for recursive descent
		if path[p] == '/' {
			if p+1 < len(path) && path[p+1] == '/' { // Recursive descent
				p += 2
				nextSep := findNextSeparator(path, p)
				key := path[p:nextSep]
				tokens = append(tokens, QueryToken{Op: OpRecursive, Value: key})
				p = nextSep
				continue
			}
			p++ // Skip single slash
			continue
		}

		if strings.HasPrefix(path[p:], "../") {
			tokens = append(tokens, QueryToken{Op: OpParent, Value: ".."})
			p += 3
			continue
		} else if strings.HasPrefix(path[p:], "..") {
			tokens = append(tokens, QueryToken{Op: OpParent, Value: ".."})
			p += 2
			continue
		}

		if path[p] == '[' {
			p++ // consume '['
			if p >= len(path) {
				return nil, fmt.Errorf("unclosed bracket in path")
			}
			if path[p] == '\'' || path[p] == '"' {
				quote := path[p]
				p++
				start := p
				for p < len(path) {
					if path[p] == '\\' { // escape
						p += 2
						continue
					}
					if path[p] == quote {
						break
					}
					p++
				}
				if p >= len(path) {
					return nil, fmt.Errorf("unclosed quote in key name")
				}
				// unescape simple escapes for quotes and backslashes
				raw := path[start:p]
				key := strings.ReplaceAll(strings.ReplaceAll(raw, `\\`, `\`), `\"`, `"`)
				tokens = append(tokens, QueryToken{Op: OpKey, Value: key})
				p++ // consume closing quote
				if p >= len(path) || path[p] != ']' {
					return nil, fmt.Errorf("missing closing bracket for quoted key")
				}
				p++ // consume ']'
				continue
			}

			end := strings.IndexByte(path[p:], ']')
			if end == -1 {
				return nil, fmt.Errorf("unclosed bracket in path: %s", path)
			}
			inner := path[p : p+end]
			p += end + 1

			if strings.HasPrefix(inner, "@") {
				tokens = append(tokens, QueryToken{Op: OpFunc, Value: strings.TrimPrefix(inner, "@")})
			} else if strings.Contains(inner, ":") {
				parts := strings.SplitN(inner, ":", 2)
				s, e := 0, -1
				var err error
				if parts[0] != "" {
					s, err = strconv.Atoi(parts[0])
					if err != nil {
						return nil, fmt.Errorf("invalid slice start: %s", parts[0])
					}
				}
				if parts[1] != "" {
					e, err = strconv.Atoi(parts[1])
					if err != nil {
						return nil, fmt.Errorf("invalid slice end: %s", parts[1])
					}
				}
				tokens = append(tokens, QueryToken{Op: OpSlice, Value: slice{Start: s, End: e}})
			} else {
				idx, err := strconv.Atoi(inner)
				if err != nil {
					return nil, fmt.Errorf("invalid index: %s", inner)
				}
				tokens = append(tokens, QueryToken{Op: OpIndex, Value: idx})
			}
			continue
		}

		nextSep := findNextSeparator(path, p)
		key := path[p:nextSep]
		if key == "*" {
			tokens = append(tokens, QueryToken{Op: OpWildcard, Value: "*"})
		} else if key != "" {
			if i, ok := tryParseInt(key); ok {
				tokens = append(tokens, QueryToken{Op: OpIndex, Value: i})
			} else {
				tokens = append(tokens, QueryToken{Op: OpKey, Value: key})
			}
		}
		p = nextSep
	}
	return tokens, nil
}

func findNextSeparator(path string, start int) int {
	for i := start; i < len(path); i++ {
		if path[i] == '/' || path[i] == '[' {
			return i
		}
	}
	return len(path)
}

func tryParseInt(s string) (int, bool) {
	if s == "" {
		return 0, false
	}
	// allow negative index
	if s[0] == '-' && len(s) == 1 {
		return 0, false
	}
	for i := 0; i < len(s); i++ {
		if s[i] == '-' && i == 0 {
			continue
		}
		if s[i] < '0' || s[i] > '9' {
			return 0, false
		}
	}
	n, err := strconv.Atoi(s)
	return n, err == nil
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
			parentRaw := data[pos : objEnd+1]
			parentNode := NewObjectNode(nil, parentRaw, funcs)
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
	// Try to get cached result first
	if bn, ok := start.(interface {
		getCachedQueryResult(string) (core.Node, bool)
	}); ok {
		if cachedResult, exists := bn.getCachedQueryResult(path); exists {
			return cachedResult
		}
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

	// Cache the result
	if bn, ok := start.(interface{ setCachedQueryResult(string, core.Node) }); ok {
		bn.setCachedQueryResult(path, cur)
	}

	return cur
}
