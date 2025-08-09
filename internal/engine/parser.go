package engine

import (
	"encoding/json"
	"fmt"

	"github.com/474420502/xjson/internal/core"
)

// ParseJSONToNode parses the JSON data and returns the root node.
func ParseJSONToNode(data string) (core.Node, error) {
	var v interface{}
	if err := json.Unmarshal([]byte(data), &v); err != nil {
		return nil, err
	}
	// Create a shared funcs map for the entire tree
	funcs := make(map[string]func(core.Node) core.Node)
	// Create the root node
	rootNode := buildNode(v, "", &funcs)

	// Set the raw string only on the root node
	if bn, ok := rootNode.(interface{ setRaw(*string) }); ok {
		bn.setRaw(&data)
	}

	return rootNode, nil
}

func buildNode(v interface{}, path string, funcs *map[string]func(core.Node) core.Node) core.Node {
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

func buildObjectNode(m map[string]interface{}, path string, funcs *map[string]func(core.Node) core.Node) core.Node {
	nodes := make(map[string]core.Node, len(m))
	for k, v := range m {
		nodes[k] = buildNode(v, path+"."+k, funcs)
	}
	return NewObjectNode(nodes, path, funcs)
}

func buildArrayNode(s []interface{}, path string, funcs *map[string]func(core.Node) core.Node) core.Node {
	nodes := make([]core.Node, len(s))
	for i, v := range s {
		nodes[i] = buildNode(v, fmt.Sprintf("%s[%d]", path, i), funcs)
	}
	return NewArrayNode(nodes, path, funcs)
}
