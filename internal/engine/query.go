package engine

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
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

func EvaluateQuery(node Node, ops []Operation) Node {
	currentNode := node

	for _, op := range ops {
		if !currentNode.IsValid() {
			return currentNode
		}

		switch op.Type {
		case OpGet:
			if currentNode.Type() == ArrayNode {
				var nextNodes []Node
				currentNode.ForEach(func(_ interface{}, elementNode Node) {
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
			// OpIndex should only apply to the current node if it's an array.
			// It doesn't map over the array, it selects from it.
			currentNode = currentNode.Index(op.Index)
		case OpFunc:
			// The function is applied to the current node as a whole,
			// regardless of whether it's an array or an object.
			// The function itself contains the logic for how to process the node.
			currentNode = currentNode.CallFunc(op.Func)
		}
	}
	return currentNode
}
