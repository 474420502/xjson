package engine

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/474420502/xjson/internal/core"
)

type OpType int

const (
	OpGet OpType = iota
	OpIndex
	OpFunc
)

type Operation struct {
	Type  OpType
	Key   string
	Index int
	Func  string
}

func ParseQuery(path string) ([]Operation, error) {
	var ops []Operation
	// Normalize path: remove leading slash and replace dots with slashes for uniform processing
	path = strings.TrimPrefix(path, "/")
	path = strings.ReplaceAll(path, ".", "/")

	parts := strings.Split(path, "/")

	for _, part := range parts {
		if part == "" {
			continue
		}

		// Handle bracket notation for indexes or functions
		if strings.Contains(part, "[") {
			openBracketIndex := strings.Index(part, "[")
			keyPart := part[:openBracketIndex]
			if keyPart != "" {
				ops = append(ops, Operation{Type: OpGet, Key: keyPart})
			}

			remaining := part[openBracketIndex:]
			for strings.HasPrefix(remaining, "[") {
				closeBracketIndex := strings.Index(remaining, "]")
				if closeBracketIndex == -1 {
					return nil, errors.New("unmatched '[' in path segment")
				}
				content := remaining[1:closeBracketIndex]

				if strings.HasPrefix(content, "@") {
					ops = append(ops, Operation{Type: OpFunc, Func: content[1:]})
				} else {
					index, err := strconv.Atoi(content)
					if err != nil {
						return nil, fmt.Errorf("invalid array index: %s", content)
					}
					ops = append(ops, Operation{Type: OpIndex, Index: index})
				}
				remaining = remaining[closeBracketIndex+1:]
			}
			if remaining != "" {
				// This case might indicate a malformed path, like "key[0]extra"
				return nil, fmt.Errorf("malformed path segment: unexpected characters after ']' in %s", part)
			}
		} else {
			// Simple key access
			ops = append(ops, Operation{Type: OpGet, Key: part})
		}
	}

	return ops, nil
}

func EvaluateQuery(node core.Node, ops []Operation) core.Node {
	currentNode := node

	// helper to flatten one level if children are arrays
	flattenIfNestedArrays := func(n core.Node) core.Node {
		if n.Type() != core.ArrayNode {
			return n
		}
		var hasArray bool
		n.ForEach(func(_ interface{}, v core.Node) {
			if v.Type() == core.ArrayNode {
				hasArray = true
			}
		})
		if !hasArray {
			return n
		}
		var flat []core.Node
		n.ForEach(func(_ interface{}, v core.Node) {
			if v.Type() == core.ArrayNode {
				v.ForEach(func(_ interface{}, inner core.Node) {
					if inner.IsValid() {
						flat = append(flat, inner)
					}
				})
			} else if v.IsValid() {
				flat = append(flat, v)
			}
		})
		return NewArrayNode(flat, n.Path(), n.GetFuncs())
	}

	for _, op := range ops {
		if !currentNode.IsValid() {
			return currentNode
		}

		switch op.Type {
		case OpGet:
			if op.Key == "*" {
				switch currentNode.Type() {
				case core.ObjectNode:
					var nextNodes []core.Node
					currentNode.ForEach(func(_ interface{}, v core.Node) {
						if v.IsValid() {
							nextNodes = append(nextNodes, v)
						}
					})
					currentNode = NewArrayNode(nextNodes, currentNode.Path(), currentNode.GetFuncs())
				case core.ArrayNode:
					// flatten nested arrays when wildcard applied
					currentNode = flattenIfNestedArrays(currentNode)
				default:
					return NewInvalidNode(currentNode.Path(), fmt.Errorf("wildcard '*' not applicable to node type"))
				}
				continue
			}
			if currentNode.Type() == core.ArrayNode {
				// map Get over elements
				var nextNodes []core.Node
				currentNode.ForEach(func(_ interface{}, elementNode core.Node) {
					child := elementNode.Get(op.Key)
					if child.IsValid() {
						nextNodes = append(nextNodes, child)
					}
				})
				currentNode = NewArrayNode(nextNodes, currentNode.Path(), currentNode.GetFuncs())
			} else {
				currentNode = currentNode.Get(op.Key)
			}
		case OpIndex:
			currentNode = currentNode.Index(op.Index)
		case OpFunc:
			// ensure flatten before applying function if we have nested arrays
			currentNode = flattenIfNestedArrays(currentNode)
			currentNode = currentNode.CallFunc(op.Func)
		}
	}
	return currentNode
}
