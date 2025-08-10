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

// ParseQuery tokenizes a query path into a sequence of operations.
func ParseQuery(path string) ([]core.QueryToken, error) {
	// Minimal parser supporting: /key segments, [index], [start:end], [@func], //recursive, and ../ parent
	if path == "" || path == "/" {
		return []core.QueryToken{}, nil
	}
	tokens := make([]core.QueryToken, 0)

	for len(path) > 0 {
		// Skip a single leading '/' but preserve '//' for recursive descent
		if strings.HasPrefix(path, "/") && !strings.HasPrefix(path, "//") {
			path = path[1:]
			continue
		}
		// Handle parent navigation
		if strings.HasPrefix(path, "../") {
			tokens = append(tokens, core.QueryToken{Op: core.OpParent, Value: ".."})
			path = path[3:]
			continue
		} else if path == "../" || path == ".." {
			tokens = append(tokens, core.QueryToken{Op: core.OpParent, Value: ".."})
			break
		}

		// Handle recursive descent
		if strings.HasPrefix(path, "//") {
			// Find the key after //
			path = path[2:] // Skip //
			nextSlash := strings.IndexByte(path, '/')
			nextBracket := strings.IndexByte(path, '[')

			cut := len(path)
			if nextSlash >= 0 && nextSlash < cut {
				cut = nextSlash
			}
			if nextBracket >= 0 && nextBracket < cut {
				cut = nextBracket
			}

			if cut > 0 {
				key := path[:cut]
				tokens = append(tokens, core.QueryToken{Op: core.OpRecursive, Value: key})
				path = path[cut:]
				if strings.HasPrefix(path, "/") {
					path = path[1:]
				}
				continue
			} else {
				// Special case: "//" at the end, which means find all nodes
				tokens = append(tokens, core.QueryToken{Op: core.OpRecursive, Value: ""})
				break
			}
		}

		if path[0] == '[' { // index or func
			end := strings.IndexByte(path, ']')
			if end <= 0 {
				return nil, fmt.Errorf("unclosed bracket in path: %s", path)
			}
			inner := path[1:end]
			if strings.HasPrefix(inner, "@") { // func
				name := strings.TrimPrefix(inner, "@")
				tokens = append(tokens, core.QueryToken{Op: core.OpFunc, Value: name})
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
				tokens = append(tokens, core.QueryToken{Op: core.OpSlice, Value: slice{Start: s, End: e}})
			} else { // index
				idx, err := strconv.Atoi(inner)
				if err != nil {
					return nil, fmt.Errorf("invalid index: %s", inner)
				}
				tokens = append(tokens, core.QueryToken{Op: core.OpIndex, Value: idx})
			}
			if end+1 < len(path) && path[end+1] == '/' {
				path = path[end+2:]
			} else {
				path = path[end+1:]
				if strings.HasPrefix(path, "/") {
					path = path[1:]
				}
			}
			continue
		}
		// key until next '/' or '['
		nextSlash := strings.IndexByte(path, '/')
		nextBracket := strings.IndexByte(path, '[')
		cut := len(path)
		if nextSlash >= 0 && nextSlash < cut {
			cut = nextSlash
		}
		if nextBracket >= 0 && nextBracket < cut {
			cut = nextBracket
		}
		key := path[:cut]
		if key != "" {
			// If the segment is a pure integer (possibly with leading -), treat it as an index op
			if idx, err := strconv.Atoi(key); err == nil && (len(key) == 1 || (key[0] != '0' && key[0] != '+')) {
				tokens = append(tokens, core.QueryToken{Op: core.OpIndex, Value: idx})
			} else {
				tokens = append(tokens, core.QueryToken{Op: core.OpKey, Value: key})
			}
		}
		if cut < len(path) && path[cut] == '/' {
			path = path[cut+1:]
		} else {
			path = path[cut:]
		}
	}
	return tokens, nil
}

// recursiveSearch performs a recursive search for a key in the node tree
func recursiveSearch(node core.Node, key string) core.Node {
	results := make([]core.Node, 0)

	var walk func(core.Node)
	walk = func(n core.Node) {
		if !n.IsValid() {
			return
		}
		switch n.Type() {
		case core.Object:
			// If searching for a specific key inside this object
			if key != "" {
				c := n.Get(key)
				if c != nil && c.IsValid() {
					results = append(results, c)
				}
			}
			// Recurse into all values
			n.ForEach(func(_ interface{}, v core.Node) {
				walk(v)
			})
		case core.Array:
			// Recurse into all elements
			n.ForEach(func(_ interface{}, v core.Node) {
				walk(v)
			})
		}
	}

	walk(node)

	return &arrayNode{
		baseNode: baseNode{funcs: node.GetFuncs()},
		children: results,
	}
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
		case core.OpKey:
			if o, ok := cur.(*objectNode); ok {
				cur = o.Get(t.Value.(string))
			} else {
				return newInvalidNode(fmt.Errorf("not an object for key access"))
			}
		case core.OpIndex:
			if a, ok := cur.(*arrayNode); ok {
				cur = a.Index(t.Value.(int))
			} else {
				return newInvalidNode(fmt.Errorf("not an array for index access"))
			}
		case core.OpSlice:
			if a, ok := cur.(*arrayNode); ok {
				a.ensureParsed()
				s := t.Value.(slice)
				start := s.Start
				end := s.End
				if start < 0 {
					start = len(a.children) + start
				}
				if end < 0 {
					end = len(a.children) + end
				}
				if end == -1 || end > len(a.children) {
					end = len(a.children)
				}
				if start < 0 {
					start = 0
				}
				if start > end {
					start = end
				}
				cur = &arrayNode{baseNode: a.baseNode, children: a.children[start:end]}
			} else {
				return newInvalidNode(fmt.Errorf("not an array for slice access"))
			}
		case core.OpFunc:
			name := t.Value.(string)
			cur = cur.CallFunc(name)
		case core.OpRecursive:
			// Handle recursive descent
			key := t.Value.(string)
			cur = recursiveSearch(cur, key)
		case core.OpParent:
			// Handle parent navigation
			// Try to get parent from different node types
			var parent core.Node
			found := false

			// Try object node
			if on, ok := cur.(*objectNode); ok && on.parent != nil {
				parent = on.parent
				found = true
			} else if an, ok := cur.(*arrayNode); ok && an.parent != nil {
				// Try array node
				parent = an.parent
				found = true
			} else if bn, ok := cur.(*baseNode); ok && bn.parent != nil {
				// Try base node
				parent = bn.parent
				found = true
			} else if pn, ok := cur.(interface{ GetParent() core.Node }); ok {
				// Try interface with GetParent method
				parent = pn.GetParent()
				if parent != nil {
					found = true
				}
			}

			if found {
				cur = parent
			} else {
				return newInvalidNode(fmt.Errorf("no parent node available"))
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
