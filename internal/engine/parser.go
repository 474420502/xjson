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
	// The raw string is passed to the root node
	return buildNode(v, "", &funcs, &data), nil
}

func buildNode(v interface{}, path string, funcs *map[string]func(Node) Node, raw *string) Node {
	switch val := v.(type) {
	case map[string]interface{}:
		return buildObjectNode(val, path, funcs, raw)
	case []interface{}:
		return buildArrayNode(val, path, funcs, raw)
	case string:
		return NewStringNode(val, path, funcs) // Primitives don't need the full raw string
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

func buildObjectNode(m map[string]interface{}, path string, funcs *map[string]func(Node) Node, raw *string) Node {
	nodes := make(map[string]Node, len(m))
	for k, v := range m {
		// Children nodes don't get the raw string, only the root does for the .Raw() method.
		nodes[k] = buildNode(v, path+"."+k, funcs, nil)
	}
	node := NewObjectNode(nodes, path, funcs).(*objectNode)
	node.raw = raw // Set raw string on the root object node
	return node
}

func buildArrayNode(s []interface{}, path string, funcs *map[string]func(Node) Node, raw *string) Node {
	nodes := make([]Node, len(s))
	for i, v := range s {
		nodes[i] = buildNode(v, fmt.Sprintf("%s[%d]", path, i), funcs, nil)
	}
	node := NewArrayNode(nodes, path, funcs).(*arrayNode)
	node.raw = raw // Set raw string on the root array node
	return node
}
