package engine

import (
	"strconv"

	"github.com/474420502/xjson/internal/core"
)

// Recreating constructors with signatures that match the test files' expectations.

func NewObjectNode(parent core.Node, children map[string]core.Node, path string, funcs *map[string]func(core.Node) core.Node) core.Node {
	// This is a temporary solution to match the test files.
	// The funcs parameter should ideally be *map[string]core.UnaryPathFunc
	var unaryFuncs *map[string]core.UnaryPathFunc
	if funcs != nil {
		tempMap := make(map[string]core.UnaryPathFunc)
		for k, v := range *funcs {
			tempMap[k] = v
		}
		unaryFuncs = &tempMap
	}

	return &objectNode{
		baseNode: baseNode{
			parent: parent,
			path:   path,
			funcs:  unaryFuncs,
		},
		children: children,
	}
}

func NewArrayNode(parent core.Node, children []core.Node, path string, funcs *map[string]func(core.Node) core.Node) core.Node {
	var unaryFuncs *map[string]core.UnaryPathFunc
	if funcs != nil {
		tempMap := make(map[string]core.UnaryPathFunc)
		for k, v := range *funcs {
			tempMap[k] = v
		}
		unaryFuncs = &tempMap
	}

	return &arrayNode{
		baseNode: baseNode{
			parent: parent,
			path:   path,
			funcs:  unaryFuncs,
		},
		children: children,
	}
}

func NewStringNode(parent core.Node, value string, path string, funcs *map[string]func(core.Node) core.Node) core.Node {
	var unaryFuncs *map[string]core.UnaryPathFunc
	if funcs != nil {
		tempMap := make(map[string]core.UnaryPathFunc)
		for k, v := range *funcs {
			tempMap[k] = v
		}
		unaryFuncs = &tempMap
	}

	return &stringNode{
		baseNode: baseNode{
			parent: parent,
			path:   path,
			funcs:  unaryFuncs,
			raw:    []byte(value),
		},
		value: value,
	}
}

func NewNumberNode(parent core.Node, value float64, path string, funcs *map[string]func(core.Node) core.Node) core.Node {
	raw := []byte(strconv.FormatFloat(value, 'f', -1, 64))
	var unaryFuncs *map[string]core.UnaryPathFunc
	if funcs != nil {
		tempMap := make(map[string]core.UnaryPathFunc)
		for k, v := range *funcs {
			tempMap[k] = v
		}
		unaryFuncs = &tempMap
	}

	return &numberNode{
		baseNode: baseNode{
			parent: parent,
			path:   path,
			funcs:  unaryFuncs,
			raw:    raw,
		},
	}
}

func NewBoolNode(parent core.Node, value bool, path string, funcs *map[string]func(core.Node) core.Node) core.Node {
	var unaryFuncs *map[string]core.UnaryPathFunc
	if funcs != nil {
		tempMap := make(map[string]core.UnaryPathFunc)
		for k, v := range *funcs {
			tempMap[k] = v
		}
		unaryFuncs = &tempMap
	}
	return &boolNode{
		baseNode: baseNode{
			parent: parent,
			path:   path,
			funcs:  unaryFuncs,
		},
		value: value,
	}
}

func NewNullNode(parent core.Node, path string, funcs *map[string]func(core.Node) core.Node) core.Node {
	var unaryFuncs *map[string]core.UnaryPathFunc
	if funcs != nil {
		tempMap := make(map[string]core.UnaryPathFunc)
		for k, v := range *funcs {
			tempMap[k] = v
		}
		unaryFuncs = &tempMap
	}

	return &nullNode{
		baseNode: baseNode{
			parent: parent,
			path:   path,
			funcs:  unaryFuncs,
		},
	}
}

func NewInvalidNode(parent core.Node, path string, err error) core.Node {
	return &invalidNode{
		baseNode: baseNode{
			parent: parent,
			path:   path,
			err:    err,
		},
	}
}
