package engine

import (
	"fmt"
	"reflect"

	"github.com/474420502/xjson/internal/core"
)

// Parse is the entry point for the engine package. It creates a new parser
// and starts parsing the raw data.
func Parse(data []byte) (core.Node, error) {
	p := newParser(data)
	return p.Parse()
}

// NewNodeFromInterface creates a new node from a Go interface.
// This is useful for creating nodes programmatically in tests or applications.
func NewNodeFromInterface(v interface{}) (core.Node, error) {
	switch val := v.(type) {
	case nil:
		return &nullNode{baseNode: baseNode{}}, nil
	case core.Node:
		return val, nil
	case string:
		return &stringNode{baseNode: baseNode{raw: []byte(val)}, value: val}, nil
	case []byte:
		s := string(val)
		return &stringNode{baseNode: baseNode{raw: []byte(s)}, value: s}, nil
	case bool:
		return &boolNode{baseNode: baseNode{}, value: val}, nil
	case int:
		return NewNodeFromInterface(int64(val))
	case int8:
		return NewNodeFromInterface(int64(val))
	case int16:
		return NewNodeFromInterface(int64(val))
	case int32:
		return NewNodeFromInterface(int64(val))
	case int64:
		return &numberNode{baseNode: baseNode{raw: []byte(fmt.Sprintf("%d", val))}}, nil
	case uint:
		return NewNodeFromInterface(uint64(val))
	case uint8:
		return NewNodeFromInterface(uint64(val))
	case uint16:
		return NewNodeFromInterface(uint64(val))
	case uint32:
		return NewNodeFromInterface(uint64(val))
	case uint64:
		return &numberNode{baseNode: baseNode{raw: []byte(fmt.Sprintf("%d", val))}}, nil
	case float32:
		return &numberNode{baseNode: baseNode{raw: []byte(fmt.Sprintf("%g", val))}}, nil
	case float64:
		return &numberNode{baseNode: baseNode{raw: []byte(fmt.Sprintf("%g", val))}}, nil
	case map[string]interface{}:
		children := make(map[string]core.Node)
		for k, v := range val {
			n, err := NewNodeFromInterface(v)
			if err != nil {
				return nil, err
			}
			children[k] = n
		}
		return &objectNode{baseNode: baseNode{}, children: children}, nil
	case []interface{}:
		arr := make([]core.Node, 0, len(val))
		for _, v := range val {
			n, err := NewNodeFromInterface(v)
			if err != nil {
				return nil, err
			}
			arr = append(arr, n)
		}
		return &arrayNode{baseNode: baseNode{}, children: arr}, nil
	default:
		rv := reflect.ValueOf(v)
		switch rv.Kind() {
		case reflect.Map:
			if rv.Type().Key().Kind() == reflect.String {
				iter := rv.MapRange()
				children := make(map[string]core.Node)
				for iter.Next() {
					k := iter.Key().String()
					child, err := NewNodeFromInterface(iter.Value().Interface())
					if err != nil {
						return nil, err
					}
					children[k] = child
				}
				return &objectNode{baseNode: baseNode{}, children: children}, nil
			}
		case reflect.Slice, reflect.Array:
			ln := rv.Len()
			arr := make([]core.Node, 0, ln)
			for i := 0; i < ln; i++ {
				child, err := NewNodeFromInterface(rv.Index(i).Interface())
				if err != nil {
					return nil, err
				}
				arr = append(arr, child)
			}
			return &arrayNode{baseNode: baseNode{}, children: arr}, nil
		}
		return nil, fmt.Errorf("unsupported type: %T", v)
	}
}
