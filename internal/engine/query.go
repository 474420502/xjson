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
	OpSlice
	OpFunc
)

type Operation struct {
	Type  OpType
	Key   string
	Index int
	Slice [2]int // [start, end], -1 means not specified
	Func  string
}

func ParseQuery(path string) ([]Operation, error) {
	var ops []Operation
	// Normalize path: remove leading slash but don't replace dots with slashes
	path = strings.TrimPrefix(path, "/")

	// Handle paths that start with quoted keys
	if strings.HasPrefix(path, "['") || strings.HasPrefix(path, "[\"") {
		// Find the matching closing bracket for the quotes
		closeBracketIndex := -1
		braceCount := 0

		for i := 2; i < len(path); i++ {
			if path[i] == '[' {
				braceCount++
			} else if path[i] == ']' {
				if braceCount == 0 {
					closeBracketIndex = i
					break
				}
				braceCount--
			}
		}

		if closeBracketIndex == -1 {
			return nil, errors.New("unmatched '[' in path segment")
		}

		// Extract the quoted key
		quotedKey := path[2 : closeBracketIndex-1] // Remove [' and ]
		ops = append(ops, Operation{Type: OpGet, Key: quotedKey})

		// Process the rest of the path
		remainingPath := strings.TrimSpace(path[closeBracketIndex+1:])
		if remainingPath != "" {
			if strings.HasPrefix(remainingPath, "/") {
				remainingPath = remainingPath[1:]
			}
			if remainingPath != "" {
				remainingOps, err := ParseQuery(remainingPath)
				if err != nil {
					return nil, err
				}
				ops = append(ops, remainingOps...)
			}
		}
		return ops, nil
	}

	// Replace dots with slashes for normal paths, but handle quoted keys separately
	// Only replace dots that are NOT within brackets or quotes
	var processedPath string
	inBrackets := false
	inQuotes := false
	quoteChar := byte(0)

	for _, char := range path {
		if (char == '\'' || char == '"') && !inQuotes {
			inQuotes = true
			quoteChar = byte(char)
		} else if byte(char) == quoteChar && inQuotes {
			inQuotes = false
			quoteChar = 0
		} else if char == '[' && !inQuotes {
			inBrackets = true
		} else if char == ']' && !inQuotes {
			inBrackets = false
		} else if char == '.' && !inBrackets && !inQuotes {
			processedPath += "/"
			continue
		}
		processedPath += string(char)
	}
	path = processedPath
	parts := strings.Split(path, "/")

	for _, part := range parts {
		if part == "" {
			continue
		}

		// Handle bracket notation for indexes, functions, or quoted keys
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
				} else if strings.Contains(content, ":") {
					// Handle slice syntax [start:end]
					parts := strings.Split(content, ":")
					if len(parts) != 2 {
						return nil, fmt.Errorf("invalid slice syntax: %s", content)
					}

					var start, end int = -1, -1 // -1 means not specified

					if parts[0] != "" {
						var err error
						start, err = strconv.Atoi(parts[0])
						if err != nil {
							return nil, fmt.Errorf("invalid slice start: %s", parts[0])
						}
					}

					if parts[1] != "" {
						var err error
						end, err = strconv.Atoi(parts[1])
						if err != nil {
							return nil, fmt.Errorf("invalid slice end: %s", parts[1])
						}
					}

					ops = append(ops, Operation{Type: OpSlice, Slice: [2]int{start, end}})
				} else if (strings.HasPrefix(content, "'") && strings.HasSuffix(content, "'")) ||
					(strings.HasPrefix(content, "\"") && strings.HasSuffix(content, "\"")) {
					// Handle quoted key syntax ['key'] or ["key"]
					quotedKey := content[1 : len(content)-1]
					ops = append(ops, Operation{Type: OpGet, Key: quotedKey})
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
				// Handle cases like [0].name - split into separate operations
				if strings.HasPrefix(remaining, ".") {
					// Process the remaining path as separate operations
					remainingOps, err := ParseQuery(remaining[1:]) // Skip the dot
					if err != nil {
						return nil, err
					}
					ops = append(ops, remainingOps...)
				} else {
					// This case might indicate a malformed path, like "key[0]extra"
					return nil, fmt.Errorf("malformed path segment: unexpected characters after ']' in %s", part)
				}
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
			if currentNode.Type() == core.ArrayNode {
				// Handle negative index
				index := op.Index
				length := currentNode.Len()
				if index < 0 {
					return NewInvalidNode(currentNode.Path(), fmt.Errorf("index %d out of bounds for array of length %d", op.Index, length))
				}
				if index >= length {
					return NewInvalidNode(currentNode.Path(), fmt.Errorf("index %d out of bounds for array of length %d", op.Index, length))
				}
				currentNode = currentNode.Index(index)
			} else {
				return NewInvalidNode(currentNode.Path(), fmt.Errorf("cannot apply index operation to non-array node type %v", currentNode.Type()))
			}
		case OpSlice:
			currentNode = applySliceOperation(currentNode, op.Slice)
		case OpFunc:
			// ensure flatten before applying function if we have nested arrays
			currentNode = flattenIfNestedArrays(currentNode)
			currentNode = currentNode.CallFunc(op.Func)
		}
	}
	return currentNode
}

// applySliceOperation applies slice operation [start:end] to a node
func applySliceOperation(node core.Node, sliceRange [2]int) core.Node {
	if !node.IsValid() {
		return node
	}

	if node.Type() != core.ArrayNode {
		// For non-array nodes, return empty array
		return NewArrayNode([]core.Node{}, node.Path(), node.GetFuncs())
	}

	start, end := sliceRange[0], sliceRange[1]
	length := node.Len()

	// Handle unspecified bounds
	if start == -1 {
		start = 0
	}
	if end == -1 {
		end = length
	}

	// Handle negative indices
	if start < 0 {
		start = length + start
		if start < 0 {
			start = 0
		}
	}
	if end < 0 {
		end = length + end
		if end < 0 {
			end = 0
		}
	}

	// Validate bounds
	if start > end {
		return NewInvalidNode(node.Path(), fmt.Errorf("slice start (%d) cannot be greater than end (%d)", start, end))
	}
	if start < 0 || end > length {
		return NewInvalidNode(node.Path(), fmt.Errorf("slice bounds [%d:%d] out of range for array of length %d", start, end, length))
	}

	// Extract slice elements
	var result []core.Node
	for i := start; i < end; i++ {
		elem := node.Index(i)
		if elem.IsValid() {
			result = append(result, elem)
		}
	}

	return NewArrayNode(result, node.Path(), node.GetFuncs())
}
