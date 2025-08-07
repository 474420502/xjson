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

// ParseQuery parses a JSONPath-like query string into a sequence of operations.
// It supports slash-separated paths, array indexing, and function calls.
// Examples: "/store/books[0]/title", "/products[@inStock]/id"
func ParseQuery(path string) ([]Operation, error) {
	var ops []Operation
	if strings.HasPrefix(path, "/") {
		path = path[1:] // Remove leading slash if present
	}

	parts := strings.Split(path, "/")
	fmt.Printf("ParseQuery: path=%s, parts=%v\n", path, parts) // Keep this for debugging if needed
	for _, part := range parts {
		if part == "" {
			continue // Skip empty parts from leading/trailing slashes or double slashes
		}
		fmt.Printf("ParseQuery: Processing part: %s\n", part) // Keep this for debugging if needed

		// Check for array index or function call
		if strings.Contains(part, "[") {
			keyPart := part

			// Find the first '['
			openBracketIndex := strings.Index(part, "[")
			if openBracketIndex == -1 {
				// Should not happen if strings.Contains(part, "[") is true, but for safety
				return nil, errors.New("malformed path segment: missing '['")
			}

			keyPart = part[:openBracketIndex]
			remaining := part[openBracketIndex:]
			fmt.Printf("ParseQuery: keyPart=%s, remaining=%s\n", keyPart, remaining) // Keep this for debugging if needed

			// Extract content within square brackets
			for strings.HasPrefix(remaining, "[") {
				closeBracketIndex := strings.Index(remaining, "]")
				if closeBracketIndex == -1 {
					return nil, errors.New("unmatched '[' in path segment")
				}
				squareBracketContent := remaining[1:closeBracketIndex]
				fmt.Printf("ParseQuery: squareBracketContent=%s\n", squareBracketContent) // Keep this for debugging if needed

				if keyPart != "" {
					ops = append(ops, Operation{Type: OpGet, Key: keyPart})
					keyPart = "" // Key part is consumed
				}

				if strings.HasPrefix(squareBracketContent, "@") {
					ops = append(ops, Operation{Type: OpFunc, Func: squareBracketContent[1:]})
					fmt.Printf("ParseQuery: Added OpFunc: %s\n", squareBracketContent[1:]) // Keep this for debugging if needed
				} else {
					index, err := strconv.Atoi(squareBracketContent)
					if err != nil {
						return nil, fmt.Errorf("invalid array index or function name: %s (error: %v)", squareBracketContent, err)
					}
					ops = append(ops, Operation{Type: OpIndex, Index: index})
					fmt.Printf("ParseQuery: Added OpIndex: %d\n", index) // Keep this for debugging if needed
				}
				remaining = remaining[closeBracketIndex+1:]
				fmt.Printf("ParseQuery: remaining after bracket: %s\n", remaining) // Keep this for debugging if needed
			}

			// If there's anything left after square brackets, it's a key
			if remaining != "" {
				ops = append(ops, Operation{Type: OpGet, Key: remaining})
				fmt.Printf("ParseQuery: Added OpGet (remaining): %s\n", remaining) // Keep this for debugging if needed
			}

		} else {
			// Simple key access
			ops = append(ops, Operation{Type: OpGet, Key: part})
			fmt.Printf("ParseQuery: Added OpGet: %s\n", part) // Keep this for debugging if needed
		}
	}
	fmt.Printf("ParseQuery: Final ops: %v\n", ops) // Keep this for debugging if needed
	return ops, nil
}

// EvaluateQuery evaluates a sequence of operations on a given node.
func EvaluateQuery(node Node, ops []Operation) Node {
	currentNode := node
	fmt.Printf("EvaluateQuery: Starting with node type %v, path %s\n", currentNode.Type(), currentNode.Path())

	for i, op := range ops {
		fmt.Printf("EvaluateQuery: Step %d, Current node type %v, path %s, Operation: %+v\n", i, currentNode.Type(), currentNode.Path(), op)
		if !currentNode.IsValid() {
			fmt.Printf("EvaluateQuery: Current node is invalid, returning. Error: %v\n", currentNode.Error())
			return currentNode
		}

		if currentNode.Type() == ArrayNode {
			var nextNodes []Node
			fmt.Printf("EvaluateQuery: Current node is ArrayNode, iterating over elements.\n")
			currentNode.ForEach(func(idx interface{}, elementNode Node) {
				fmt.Printf("EvaluateQuery:   Processing array element %v, type %v, path %s\n", idx, elementNode.Type(), elementNode.Path())
				var resultNode Node
				switch op.Type {
				case OpGet:
					resultNode = elementNode.Get(op.Key)
					fmt.Printf("EvaluateQuery:     OpGet('%s') on element. Result type %v, path %s, IsValid: %t\n", op.Key, resultNode.Type(), resultNode.Path(), resultNode.IsValid())
				case OpIndex:
					resultNode = elementNode.Index(op.Index)
					fmt.Printf("EvaluateQuery:     OpIndex(%d) on element. Result type %v, path %s, IsValid: %t\n", op.Index, resultNode.Type(), resultNode.Path(), resultNode.IsValid())
				case OpFunc:
					// If a function returns an array, flatten it into nextNodes
					fmt.Printf("EvaluateQuery:     OpFunc('%s') on element.\n", op.Func)
					funcResult := elementNode.CallFunc(op.Func)
					fmt.Printf("EvaluateQuery:       Func result type %v, path %s, IsValid: %t\n", funcResult.Type(), funcResult.Path(), funcResult.IsValid())
					if funcResult.IsValid() && funcResult.Type() == ArrayNode {
						fmt.Printf("EvaluateQuery:         Func returned ArrayNode, flattening.\n")
						for _, n := range funcResult.Array() {
							fmt.Printf("EvaluateQuery:           Adding flattened node type %v, path %s\n", n.Type(), n.Path())
							nextNodes = append(nextNodes, n)
						}
					} else if funcResult.IsValid() {
						fmt.Printf("EvaluateQuery:         Func returned single node, adding.\n")
						nextNodes = append(nextNodes, funcResult)
					} else {
						fmt.Printf("EvaluateQuery:         Func returned invalid node. Error: %v\n", funcResult.Error())
					}
				default: // This case should ideally not be hit if all OpTypes are handled
					if resultNode.IsValid() {
						nextNodes = append(nextNodes, resultNode)
					}
				}
			})
			currentNode = NewArrayNode(nextNodes, currentNode.Path(), currentNode.GetFuncs())
			fmt.Printf("EvaluateQuery: After array processing, current node type %v, len %d, path %s\n", currentNode.Type(), currentNode.Len(), currentNode.Path())
		} else { // Handle non-array nodes
			fmt.Printf("EvaluateQuery: Current node is non-ArrayNode, type %v, path %s\n", currentNode.Type(), currentNode.Path())
			switch op.Type {
			case OpGet:
				currentNode = currentNode.Get(op.Key)
				fmt.Printf("EvaluateQuery:   OpGet('%s'). Result type %v, path %s, IsValid: %t\n", op.Key, currentNode.Type(), currentNode.Path(), currentNode.IsValid())
			case OpIndex:
				currentNode = currentNode.Index(op.Index)
				fmt.Printf("EvaluateQuery:   OpIndex(%d). Result type %v, path %s, IsValid: %t\n", op.Index, currentNode.Type(), currentNode.Path(), currentNode.IsValid())
			case OpFunc:
				// If a function returns an array, the current node becomes that array
				fmt.Printf("EvaluateQuery:   OpFunc('%s').\n", op.Func)
				funcResult := currentNode.CallFunc(op.Func)
				fmt.Printf("EvaluateQuery:     Func result type %v, path %s, IsValid: %t\n", funcResult.Type(), funcResult.Path(), funcResult.IsValid())
				if funcResult.IsValid() && funcResult.Type() == ArrayNode {
					fmt.Printf("EvaluateQuery:       Func returned ArrayNode, current node becomes this array.\n")
					currentNode = funcResult
				} else {
					fmt.Printf("EvaluateQuery:       Func returned single node or invalid, current node becomes this result.\n")
					currentNode = funcResult
				}
			}
		}
	}
	fmt.Printf("EvaluateQuery: Final node type %v, path %s, IsValid: %t, Error: %v\n", currentNode.Type(), currentNode.Path(), currentNode.IsValid(), currentNode.Error())
	return currentNode
}
