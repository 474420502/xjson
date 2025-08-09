package xjson

import (
	"github.com/474420502/xjson/internal/core"
	"github.com/474420502/xjson/internal/engine"
)

// 公共类型别名
// Public type aliases
type Node = core.Node
type NodeType = core.NodeType

// 面向公共 API 的函数类型别名
// Functional type aliases for public API
type PathFunc = core.PathFunc
type UnaryPathFunc = core.UnaryPathFunc
type PredicateFunc = core.PredicateFunc
type TransformFunc = core.TransformFunc

// 定义节点类型的常量
// Define constants for node types
const (
	ObjectNode  = core.ObjectNode
	ArrayNode   = core.ArrayNode
	StringNode  = core.StringNode
	NumberNode  = core.NumberNode
	BoolNode    = core.BoolNode
	NullNode    = core.NullNode
	InvalidNode = core.InvalidNode
)

// Parse 接受一个 JSON 字符串并返回解析结构的根节点。
// 返回的 Node 可用于导航和操作 JSON 数据。
// Parse takes a JSON string and returns the root node of the parsed structure.
// The returned Node can be used to navigate and manipulate the JSON data.
func Parse(data string) (Node, error) {
	return engine.ParseJSONToNode(data)
}

// ParseBytes 是围绕 Parse 为字节切片提供的便捷包装器。
// ParseBytes is a convenience wrapper around Parse for byte slices.
func ParseBytes(data []byte) (Node, error) {
	return Parse(string(data))
}

// NewNodeFromInterface 从 Go 的 interface{} 创建一个新 Node。
// 这对于以编程方式构建节点非常有用。
// NewNodeFromInterface creates a new Node from a Go interface{}.
// This is useful for building nodes programmatically.
func NewNodeFromInterface(value interface{}) (Node, error) {
	return engine.NewNodeFromInterface(value, "", nil)
}

// NewParser 是一个用于创建解析器的辅助函数，主要用于测试目的。
// NewParser is a helper for creating a parser, used for testing purposes.
func NewParser(data string) (Node, error) {
	return engine.ParseJSONToNode(data)
}
