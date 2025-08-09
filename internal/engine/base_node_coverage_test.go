package engine

import (
	"testing"
	"time"

	"github.com/474420502/xjson/internal/core"
	"github.com/stretchr/testify/assert"
)

func TestBaseNode(t *testing.T) {
	// 创建一个 baseNode 实例
	// Create a baseNode instance
	bn := &baseNode{}

	// 测试设置了原始值的 Raw 方法
	// Test Raw with raw value set
	rawStr := "raw string"
	bn.setRaw(&rawStr)
	assert.Equal(t, "raw string", bn.Raw())

	// 测试当 raw 为 nil 时的 Raw 方法（应回退到 String()）
	// Test Raw when raw is nil (should fall back to String())
	bn.raw = nil
	assert.Equal(t, "", bn.Raw()) // String() 返回 ""

	// 测试所有的存根方法
	// Test all the stub methods
	assert.Equal(t, core.InvalidNode, bn.Type())
	assert.Nil(t, bn.Get("key"))
	assert.Nil(t, bn.Index(0))
	assert.Nil(t, bn.Query("path"))

	// ForEach 不应引起任何问题
	// ForEach should not cause any issues
	bn.ForEach(func(key interface{}, value core.Node) {})

	assert.Equal(t, 0, bn.Len())
	assert.Equal(t, "", bn.String())
	assert.Panics(t, func() { bn.MustString() })
	assert.Equal(t, float64(0), bn.Float())
	assert.Panics(t, func() { bn.MustFloat() })
	assert.Equal(t, int64(0), bn.Int())
	assert.Panics(t, func() { bn.MustInt() })
	assert.False(t, bn.Bool())
	assert.Panics(t, func() { bn.MustBool() })
	assert.Equal(t, time.Time{}, bn.Time())
	assert.Panics(t, func() { bn.MustTime() })
	assert.Nil(t, bn.Array())
	assert.Panics(t, func() { bn.MustArray() })
	assert.Nil(t, bn.Interface())
	assert.Nil(t, bn.Set("key", "value"))
	assert.Nil(t, bn.Append("value"))

	f, ok := bn.RawFloat()
	assert.Equal(t, float64(0), f)
	assert.False(t, ok)

	s, ok := bn.RawString()
	assert.Equal(t, "", s)
	assert.False(t, ok)

	assert.False(t, bn.Contains("value"))
	assert.Nil(t, bn.Strings())
	assert.Nil(t, bn.AsMap())
	assert.Panics(t, func() { bn.MustAsMap() })

	assert.Nil(t, bn.CallFunc("name"))
	assert.Nil(t, bn.RemoveFunc("name"))

	// Filter 应该返回一个 InvalidNode
	// Filter should return an InvalidNode
	filterNode := bn.Filter(func(n core.Node) bool { return true })
	assert.NotNil(t, filterNode)
	assert.Equal(t, core.InvalidNode, filterNode.Type())
	assert.Error(t, filterNode.Error())

	// Map 应该返回一个 InvalidNode
	// Map should return an InvalidNode
	mapNode := bn.Map(func(n core.Node) interface{} { return n })
	assert.NotNil(t, mapNode)
	assert.Equal(t, core.InvalidNode, mapNode.Type())
	assert.Error(t, mapNode.Error())

	assert.Nil(t, bn.RegisterFunc("name", nil))

	// 测试 Apply 为 nil 的情况, 必须 panic
	// Test the case where Apply is nil, it must panic
	assert.Panics(t, func() { bn.Apply(nil) })

	// 覆盖在 baseNode 上的 ForEach
	// Cover ForEach on a baseNode
	var forEachCalled bool
	bn.ForEach(func(key interface{}, value core.Node) {
		forEachCalled = true
	})
	assert.False(t, forEachCalled, "ForEach on baseNode should not call the iterator")
}

// 单独测试 GetFuncs 以覆盖它
// Test GetFuncs separately to cover it
func TestBaseNodeGetFuncs(t *testing.T) {
	// 创建一个 baseNode 实例
	// Create a baseNode instance
	bn := &baseNode{}
	assert.Nil(t, bn.GetFuncs())

	// 创建一个函数映射并设置给 bn.funcs
	// Create a function map and set it to bn.funcs
	funcs := make(map[string]func(core.Node) core.Node)
	bn.funcs = &funcs
	assert.Same(t, &funcs, bn.GetFuncs())
}
