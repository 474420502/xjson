package engine

import (
	"encoding/json"
	"fmt"
)

// ParseJSONToNode parses the JSON data and returns the root node.
func ParseJSONToNode(data string) (Node, error) {
	var v interface{}
	if err := json.Unmarshal([]byte(data), &v); err != nil {
		return nil, err
	}
	// Create a shared funcs map for the entire tree
	funcs := make(map[string]func(Node) Node)
	return buildNode(v, "", &funcs), nil
}

func buildNode(v interface{}, path string, funcs *map[string]func(Node) Node) Node {
	switch val := v.(type) {
	case map[string]interface{}:
		return buildObjectNode(val, path, funcs)
	case []interface{}:
		return buildArrayNode(val, path, funcs)
	case string:
		return NewStringNode(val, path, funcs)
	case float64:
		return NewNumberNode(val, path, funcs)
	case bool:
		return NewBoolNode(val, path, funcs)
	case nil:
		return NewNullNode(path, funcs)
	default:
		return NewInvalidNode(path, ErrInvalidNode)
	}
}

func buildObjectNode(m map[string]interface{}, path string, funcs *map[string]func(Node) Node) Node {
	nodes := make(map[string]Node, len(m))
	for k, v := range m {
		nodes[k] = buildNode(v, path+"."+k, funcs)
	}
	return NewObjectNode(nodes, path, funcs)
}

func buildArrayNode(s []interface{}, path string, funcs *map[string]func(Node) Node) Node {
	nodes := make([]Node, len(s))
	for i, v := range s {
		nodes[i] = buildNode(v, fmt.Sprintf("%s[%d]", path, i), funcs) // Corrected path for array elements
	}
	return NewArrayNode(nodes, path, funcs)
}
