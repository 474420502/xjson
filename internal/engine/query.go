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
				for p < len(path) && path[p] != quote {
					p++
				}
				if p >= len(path) {
					return nil, fmt.Errorf("unclosed quote in key name")
				}
				key := path[start:p]
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
	if s[0] == '-' {
		for i := 1; i < len(s); i++ {
			if s[i] < '0' || s[i] > '9' {
				return 0, false
			}
		}
	} else {
		for i := 0; i < len(s); i++ {
			if s[i] < '0' || s[i] > '9' {
				return 0, false
			}
		}
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0, false
	}
	return v, true
}

func recursiveSearch(node core.Node, key string) core.Node {
	results := make([]core.Node, 0)

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

func applySimpleQuery(start core.Node, path string) core.Node {
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
				results := make([]core.Node, 0)
				a.lazyParse()
				for _, item := range a.value {
					if item.Type() == core.Object {
						res := item.Get(key)
						if res.IsValid() {
							results = append(results, res)
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
				a.lazyParse()
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
				o.lazyParse()
				for _, v := range o.value {
					results = append(results, v)
				}
			} else if a, ok := cur.(*arrayNode); ok {
				a.lazyParse()
				results = a.value
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
			if p := cur.Parent(); p != nil {
				cur = p
			} else {
				return newInvalidNode(fmt.Errorf("no parent node available for node %v", cur.Raw()))
			}
		default:
			return newInvalidNode(fmt.Errorf("unsupported op"))
		}
		if cur == nil {
			return newInvalidNode(fmt.Errorf("nil during query"))
		}
	}
	return cur
}
