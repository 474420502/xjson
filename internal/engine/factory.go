package engine

import (
	"fmt"

	"github.com/474420502/xjson/internal/core"
)

// NewNodeFromInterface converts a Go interface{} value into an xjson.Node.
// This is useful for creating new nodes from arbitrary Go types, especially
// when used with the Map function.
func NewNodeFromInterface(value interface{}, path string, funcs *map[string]func(core.Node) core.Node) (core.Node, error) {
	if funcs == nil {
		funcs = &map[string]func(core.Node) core.Node{} // Initialize if nil (for root)
	}
	switch v := value.(type) {
	case map[string]interface{}:
		obj := make(map[string]core.Node, len(v))
		for key, val := range v {
			node, err := NewNodeFromInterface(val, path+"."+key, funcs)
			if err != nil {
				return nil, err
			}
			obj[key] = node
		}
		return NewObjectNode(obj, path, funcs), nil
	case map[string]int:
		converted := make(map[string]interface{}, len(v))
		for k, val := range v {
			converted[k] = val
		}
		return NewNodeFromInterface(converted, path, funcs)
	case map[string]int64:
		converted := make(map[string]interface{}, len(v))
		for k, val := range v {
			converted[k] = val
		}
		return NewNodeFromInterface(converted, path, funcs)
	case map[string]float64:
		converted := make(map[string]interface{}, len(v))
		for k, val := range v {
			converted[k] = val
		}
		return NewNodeFromInterface(converted, path, funcs)
	case map[string]string:
		converted := make(map[string]interface{}, len(v))
		for k, val := range v {
			converted[k] = val
		}
		return NewNodeFromInterface(converted, path, funcs)
	case map[string]bool:
		converted := make(map[string]interface{}, len(v))
		for k, val := range v {
			converted[k] = val
		}
		return NewNodeFromInterface(converted, path, funcs)
	case []interface{}:
		arr := make([]core.Node, len(v))
		for i, val := range v {
			node, err := NewNodeFromInterface(val, path+fmt.Sprintf("[%d]", i), funcs)
			if err != nil {
				return nil, err
			}
			arr[i] = node
		}
		return NewArrayNode(arr, path, funcs), nil
	case string:
		return NewStringNode(v, path, funcs), nil
	case float64:
		return NewNumberNode(v, path, funcs), nil
	case int:
		return NewNumberNode(float64(v), path, funcs), nil
	case int64:
		return NewNumberNode(float64(v), path, funcs), nil
	case bool:
		return NewBoolNode(v, path, funcs), nil
	case nil:
		return NewNullNode(path, funcs), nil
	default:
		return nil, fmt.Errorf("unsupported type for node creation: %T", v)
	}
}
