package engine

import (
	"fmt"
	"strconv"

	"github.com/474420502/xjson/internal/core"
)

func newInvalidNode(err error) core.Node {
	n := &invalidNode{baseNode: baseNode{err: err}}
	n.baseNode.self = n
	return n
}

func NewObjectNode(parent core.Node, raw []byte, funcs *map[string]core.UnaryPathFunc) core.Node {
	n := &objectNode{
		baseNode: baseNode{
			raw:    raw,
			parent: parent,
			funcs:  funcs,
		},
		// Don't pre-allocate map - allocate only when needed to reduce memory pressure
	}
	n.baseNode.self = n
	return n
}

func NewArrayNode(parent core.Node, raw []byte, funcs *map[string]core.UnaryPathFunc) core.Node {
	n := &arrayNode{
		baseNode: baseNode{
			raw:    raw,
			parent: parent,
			funcs:  funcs,
		},
		value: make([]core.Node, 0),
	}
	n.baseNode.self = n
	return n
}

func NewStringNode(parent core.Node, val string, funcs *map[string]core.UnaryPathFunc) core.Node {
	n := &stringNode{
		baseNode: baseNode{
			raw:    []byte(val),
			parent: parent,
			funcs:  funcs,
		},
		value:         val,
		decoded:       true,
		needsUnescape: false,
	}
	n.baseNode.self = n
	return n
}

// NewRawStringNode creates a string node from raw quoted bytes (including quotes).
// start/end are indexes into raw for the unquoted value (start inclusive, end exclusive).
// needsUnescape indicates whether the value contains escape sequences and must be unescaped when requested.
func NewRawStringNode(parent core.Node, raw []byte, start int, end int, needsUnescape bool, funcs *map[string]core.UnaryPathFunc) core.Node {
	n := &stringNode{
		baseNode: baseNode{
			raw:    raw,
			parent: parent,
			funcs:  funcs,
			start:  start,
			end:    end,
		},
		value:         "",
		decoded:       false,
		needsUnescape: needsUnescape,
	}
	n.baseNode.self = n
	return n
}

func NewNumberNode(parent core.Node, raw []byte, funcs *map[string]core.UnaryPathFunc) core.Node {
	n := &numberNode{
		baseNode: baseNode{
			raw:    raw,
			parent: parent,
			funcs:  funcs,
		},
	}
	n.baseNode.self = n
	return n
}

func NewBoolNode(parent core.Node, val bool, funcs *map[string]core.UnaryPathFunc) core.Node {
	n := &boolNode{
		baseNode: baseNode{
			raw:    []byte(strconv.FormatBool(val)),
			parent: parent,
			funcs:  funcs,
		},
		value: val,
	}
	n.baseNode.self = n
	return n
}

func NewNullNode(parent core.Node, funcs *map[string]core.UnaryPathFunc) core.Node {
	n := &nullNode{
		baseNode: baseNode{
			raw:    []byte("null"),
			parent: parent,
			funcs:  funcs,
		},
	}
	n.baseNode.self = n
	return n
}

func NewNodeFromInterface(parent core.Node, v interface{}, funcs *map[string]core.UnaryPathFunc) core.Node {
	switch val := v.(type) {
	case map[string]interface{}:
		node := NewObjectNode(parent, nil, funcs).(*objectNode)
		node.isDirty = true
		for key, value := range val {
			child := NewNodeFromInterface(node, value, funcs)
			if !child.IsValid() {
				node.err = child.Error()
				return node
			}
			node.value[key] = child
		}
		return node
	case []interface{}:
		node := NewArrayNode(parent, nil, funcs).(*arrayNode)
		node.isDirty = true
		for _, value := range val {
			child := NewNodeFromInterface(node, value, funcs)
			if !child.IsValid() {
				node.err = child.Error()
				return node
			}
			node.value = append(node.value, child)
		}
		return node
	case string:
		return NewStringNode(parent, val, funcs)
	case float64:
		return NewNumberNode(parent, []byte(strconv.FormatFloat(val, 'f', -1, 64)), funcs)
	case int:
		return NewNumberNode(parent, []byte(strconv.Itoa(val)), funcs)
	case int64:
		return NewNumberNode(parent, []byte(strconv.FormatInt(val, 10)), funcs)
	case bool:
		return NewBoolNode(parent, val, funcs)
	case nil:
		return NewNullNode(parent, funcs)
	default:
		return newInvalidNode(fmt.Errorf("unsupported type %T for NewNodeFromInterface", v))
	}
}
