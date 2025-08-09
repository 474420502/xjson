package engine

import (
	"testing"

	"github.com/474420502/xjson/internal/core"
)

// TestInvalidNodeForEach 测试 InvalidNode 的 ForEach 方法
func TestInvalidNodeForEach(t *testing.T) {
	invalidNode := NewInvalidNode("", nil)

	// 测试 InvalidNode 的 ForEach 方法
	invalidNode.ForEach(func(keyOrIndex interface{}, value core.Node) {
		t.Error("InvalidNode ForEach should not call the iterator function")
	})
}

// TestStringNodeApply 测试 StringNode 的 Apply 方法的新分支
func TestStringNodeApply(t *testing.T) {
	strNode := NewStringNode("test", "", nil)

	// 测试 Apply 方法使用不支持的函数类型 - 这应该返回 InvalidNode
	result := strNode.Apply(func() string {
		return "unsupported"
	})

	if result.Type() != core.InvalidNode {
		t.Error("Apply with unsupported function type should return InvalidNode")
	}
}

// TestStringNodeRawMethods 测试 StringNode 的 Raw 相关方法
func TestStringNodeRawMethods(t *testing.T) {
	strNode := NewStringNode("123.45", "", nil)

	// 测试 RawFloat 方法
	val, ok := strNode.RawFloat()
	if !ok || val != 123.45 {
		t.Error("RawFloat should return correct float value")
	}

	// 测试 RawString 方法
	valStr, ok := strNode.RawString()
	if !ok || valStr != "123.45" {
		t.Error("RawString should return correct string value")
	}

	// 测试非数字字符串的 RawFloat
	strNode = NewStringNode("not_a_number", "", nil)
	val, ok = strNode.RawFloat()
	if ok {
		t.Error("RawFloat should return false for non-numeric string")
	}
}

// TestStringNodeStrings 测试 StringNode 的 Strings 方法
func TestStringNodeStrings(t *testing.T) {
	// StringNode 的 Strings 方法返回包含自身字符串的切片
	strNode := NewStringNode("test", "", nil)
	result := strNode.Strings()
	if len(result) != 1 || result[0] != "test" {
		t.Error("StringNode Strings should return slice with its own string")
	}
}

// TestStringNodeContains 测试 StringNode 的 Contains 方法
func TestStringNodeContains(t *testing.T) {
	// StringNode 的 Contains 方法检查字符串是否包含子字符串
	strNode := NewStringNode("hello world", "", nil)

	// 测试包含的情况
	result := strNode.Contains("world")
	if !result {
		t.Error("StringNode Contains should return true for substring")
	}

	// 测试不包含的情况
	result = strNode.Contains("xyz")
	if result {
		t.Error("StringNode Contains should return false for non-substring")
	}
}

// TestStringNodeTime 测试 StringNode 的 Time 方法
func TestStringNodeTime(t *testing.T) {
	// 测试有效的时间格式
	timeStr := "2023-01-01T00:00:00Z"
	strNode := NewStringNode(timeStr, "", nil)

	result := strNode.Time()
	if result.IsZero() {
		t.Error("Time should parse valid RFC3339 time string")
	}

	// 测试无效的时间格式
	strNode = NewStringNode("invalid_time", "", nil)
	result = strNode.Time()
	if !result.IsZero() {
		t.Error("Time should return zero time for invalid time string")
	}
}

// TestNumberNodeRawFloat 测试 NumberNode 的 RawFloat 方法
func TestNumberNodeRawFloat(t *testing.T) {
	numNode := NewNumberNode(123.45, "", nil)

	val, ok := numNode.RawFloat()
	if !ok || val != 123.45 {
		t.Error("RawFloat should return correct float value")
	}
}

// TestNumberNodeApply 测试 NumberNode 的 Apply 方法
func TestNumberNodeApply(t *testing.T) {
	numNode := NewNumberNode(42, "", nil)

	// 测试 Apply 方法使用 TransformFunc - 需要查看number_node.go的实现
	// 让我们先测试不支持的函数类型
	result := numNode.Apply(func() int {
		return 0
	})

	if result.Type() != core.InvalidNode {
		t.Error("Apply with unsupported function type should handle gracefully")
	}
}

// TestObjectNodeSet 测试 ObjectNode 的 Set 方法
func TestObjectNodeSet(t *testing.T) {
	obj := map[string]core.Node{
		"existing": NewStringNode("value", "", nil),
	}
	objNode := NewObjectNode(obj, "", nil)

	// 测试设置新值
	result := objNode.Set("new_key", "new_value")
	if result.Type() != core.ObjectNode {
		t.Error("Set should return ObjectNode")
	}

	// 测试在无效节点上调用 Set
	invalidNode := NewInvalidNode("", nil)
	result = invalidNode.Set("key", "value")
	if result.Type() != core.InvalidNode {
		t.Error("Set on InvalidNode should return InvalidNode")
	}
}

// TestObjectNodeAsMap 测试 ObjectNode 的 AsMap 方法
func TestObjectNodeAsMap(t *testing.T) {
	obj := map[string]core.Node{
		"key1": NewStringNode("value1", "", nil),
		"key2": NewNumberNode(42, "", nil),
	}
	objNode := NewObjectNode(obj, "", nil)

	result := objNode.AsMap()
	if len(result) != 2 {
		t.Error("AsMap should return correct map length")
	}

	if result["key1"].String() != "value1" {
		t.Error("AsMap should preserve values")
	}
}

// TestArrayNodeRaw 测试 ArrayNode 的 Raw 方法
func TestArrayNodeRaw(t *testing.T) {
	arr := []core.Node{
		NewStringNode("value1", "", nil),
		NewNumberNode(42, "", nil),
	}
	arrNode := NewArrayNode(arr, "", nil)

	// 测试 Raw 方法生成 JSON 字符串
	result := arrNode.Raw()
	expected := `["value1",42]`
	if result != expected {
		t.Error("Raw should generate correct JSON string")
	}
}

// TestFactoryUnsupportedType 测试工厂方法处理不支持的类型
func TestFactoryUnsupportedType(t *testing.T) {
	// 测试不支持的类型
	_, err := NewNodeFromInterface(struct{}{}, "", nil)
	if err == nil {
		t.Error("NewNodeFromInterface should return error for unsupported type")
	}
}

// TestQueryParseEdgeCases 测试查询解析的边界情况
func TestQueryParseEdgeCases(t *testing.T) {
	// 测试空路径
	ops, err := ParseQuery("")
	if err != nil || len(ops) != 0 {
		t.Error("Empty path should return empty operations")
	}

	// 测试只有斜杠的路径
	ops, err = ParseQuery("/")
	if err != nil || len(ops) != 0 {
		t.Error("Path with only slash should return empty operations")
	}

	// 测试不匹配的括号
	_, err = ParseQuery("key[0")
	if err == nil {
		t.Error("Unmatched bracket should return error")
	}

	// 测试无效的索引
	_, err = ParseQuery("key[invalid]")
	if err == nil {
		t.Error("Invalid index should return error")
	}
}
