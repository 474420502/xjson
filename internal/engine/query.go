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
	fmt.Printf("ParseQuery: path=%s, parts=%v\n", path, parts)

	for _, part := range parts {
		if part == "" {
			continue
		}
		fmt.Printf("ParseQuery: Processing part: %s\n", part)

		// Handle bracket notation for indexes or functions
		if strings.Contains(part, "[") {
			openBracketIndex := strings.Index(part, "[")
			keyPart := part[:openBracketIndex]
			if keyPart != "" {
				ops = append(ops, Operation{Type: OpGet, Key: keyPart})
				fmt.Printf("ParseQuery: Added OpGet: %s\n", keyPart)
			}

			remaining := part[openBracketIndex:]
			for strings.HasPrefix(remaining, "[") {
				closeBracketIndex := strings.Index(remaining, "]")
				if closeBracketIndex == -1 {
					return nil, errors.New("unmatched '[' in path segment")
				}
				content := remaining[1:closeBracketIndex]
				fmt.Printf("ParseQuery: squareBracketContent=%s\n", content)

				if strings.HasPrefix(content, "@") {
					ops = append(ops, Operation{Type: OpFunc, Func: content[1:]})
					fmt.Printf("ParseQuery: Added OpFunc: %s\n", content[1:])
				} else {
					index, err := strconv.Atoi(content)
					if err != nil {
						return nil, fmt.Errorf("invalid array index: %s", content)
					}
					ops = append(ops, Operation{Type: OpIndex, Index: index})
					fmt.Printf("ParseQuery: Added OpIndex: %d\n", index)
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
			fmt.Printf("ParseQuery: Added OpGet: %s\n", part)
		}
	}

	fmt.Printf("ParseQuery: Final ops: %v\n", ops)
	return ops, nil
}

func EvaluateQuery(node Node, ops []Operation) Node {
	currentNode := node
	fmt.Printf("EvaluateQuery: Starting with node type %v, path %s\n", currentNode.Type(), currentNode.Path())

	for i, op := range ops {
		fmt.Printf("EvaluateQuery: Step %d, Current node type %v, path %s, Operation: %+v\n", i, currentNode.Type(), currentNode.Path(), op)
		if !currentNode.IsValid() {
			fmt.Printf("EvaluateQuery: Current node is invalid, returning. Error: %v\n", currentNode.Error())
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
			fmt.Printf("EvaluateQuery:   OpGet('%s'). Result type %v, path %s, IsValid: %t\n", op.Key, currentNode.Type(), currentNode.Path(), currentNode.IsValid())
		case OpIndex:
			// OpIndex should only apply to the current node if it's an array.
			// It doesn't map over the array, it selects from it.
			currentNode = currentNode.Index(op.Index)
			fmt.Printf("EvaluateQuery:   OpIndex(%d). Result type %v, path %s, IsValid: %t\n", op.Index, currentNode.Type(), currentNode.Path(), currentNode.IsValid())
		case OpFunc:
			// The function is applied to the current node as a whole,
			// regardless of whether it's an array or an object.
			// The function itself contains the logic for how to process the node.
			currentNode = currentNode.CallFunc(op.Func)
			fmt.Printf("EvaluateQuery:   OpFunc('%s'). Result type %v, path %s, IsValid: %t\n", op.Func, currentNode.Type(), currentNode.Path(), currentNode.IsValid())
		}
	}
	fmt.Printf("EvaluateQuery: Final node type %v, path %s, IsValid: %t, Error: %v\n", currentNode.Type(), currentNode.Path(), currentNode.IsValid(), currentNode.Error())
	return currentNode
}
