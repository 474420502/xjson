package engine

import (
	"encoding/json"
	"fmt"

	"github.com/474420502/xjson/internal/core"
)

// ParseJSONToNode 解析 JSON 数据并返回根节点。
// ParseJSONToNode parses the JSON data and returns the root node.
func ParseJSONToNode(data string) (core.Node, error) {
	var v interface{}
	if err := json.Unmarshal([]byte(data), &v); err != nil {
		return nil, err
	}
	// 为整个树创建一个共享的函数映射
	// Create a shared funcs map for the entire tree
	funcs := make(map[string]func(core.Node) core.Node)
	// 创建根节点
	// Create the root node
	rootNode := buildNode(v, "", &funcs)

	// 仅在根节点上设置原始字符串
	// Set the raw string only on the root node
	if bn, ok := rootNode.(interface{ setRaw(*string) }); ok {
		bn.setRaw(&data)
	}

	return rootNode, nil
}

// buildNode 根据值的类型构建不同类型的节点
// buildNode builds different types of nodes based on the value type
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

// buildObjectNode 构建对象节点
// buildObjectNode builds an object node
func buildObjectNode(m map[string]interface{}, path string, funcs *map[string]func(core.Node) core.Node) core.Node {
	nodes := make(map[string]core.Node, len(m))
	for k, v := range m {
		nodes[k] = buildNode(v, path+"."+k, funcs)
	}
	return NewObjectNode(nodes, path, funcs)
}

// buildArrayNode 构建数组节点
// buildArrayNode builds an array node
func buildArrayNode(s []interface{}, path string, funcs *map[string]func(core.Node) core.Node) core.Node {
	nodes := make([]core.Node, len(s))
	for i, v := range s {
		nodes[i] = buildNode(v, fmt.Sprintf("%s[%d]", path, i), funcs)
	}
	return NewArrayNode(nodes, path, funcs)
}
