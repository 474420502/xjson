package xjson

import (
	"fmt"

	"github.com/474420502/xjson/internal/core"
	"github.com/474420502/xjson/internal/engine"
)

// NodeType is an alias for the core NodeType.
type NodeType = core.NodeType

const (
	Invalid = core.Invalid
	Object  = core.Object
	Array   = core.Array
	String  = core.String
	Number  = core.Number
	Bool    = core.Bool
	Null    = core.Null
)

// PathFunc is an alias for the core PathFunc.
type PathFunc = core.PathFunc

// UnaryPathFunc is an alias for the core UnaryPathFunc.
type UnaryPathFunc = core.UnaryPathFunc

// PredicateFunc is an alias for the core PredicateFunc.
type PredicateFunc = core.PredicateFunc

// TransformFunc is an alias for the core TransformFunc.
type TransformFunc = core.TransformFunc

// Node is an alias for the core Node.
type Node = core.Node

// Parse parses a raw JSON string or bytes and returns the root Node.
// This is the main entry point for using the XJSON library.
func Parse(data interface{}) (Node, error) {
	var raw []byte
	switch v := data.(type) {
	case string:
		raw = []byte(v)
	case []byte:
		raw = v
	default:
		return nil, fmt.Errorf("unsupported data type: %T", data)
	}

	if len(raw) == 0 {
		return nil, fmt.Errorf("empty data")
	}

	return engine.Parse(raw)
}
