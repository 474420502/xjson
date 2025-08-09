package engine

import (
	"strings"
	"testing"
	"time"

	"github.com/474420502/xjson/internal/core"
	"github.com/stretchr/testify/assert"
)

func TestObjectNodeCoverage(t *testing.T) {
	// 创建一个带原始数据的根对象节点
	// Create a root object node with raw data
	jsonData := `{"str":"test","num":42,"bool":true}`
	funcs := make(map[string]func(core.Node) core.Node)
	obj := NewObjectNode(
		map[string]core.Node{
			"str":  NewStringNode("test", "", &funcs),
			"num":  NewNumberNode(42, "", &funcs),
			"bool": NewBoolNode(true, "", &funcs),
		},
		"",
		&funcs,
	)
	// 为根节点设置原始数据
	// Set raw data for the root node
	obj.(*objectNode).raw = &jsonData

	t.Run("ForEach", func(t *testing.T) {
		count := 0
		var keys []string
		obj.ForEach(func(key interface{}, value core.Node) {
			count++
			keys = append(keys, key.(string))
		})
		assert.Equal(t, 3, count)
		assert.Contains(t, keys, "str")
		assert.Contains(t, keys, "num")
		assert.Contains(t, keys, "bool")
	})

	t.Run("String", func(t *testing.T) {
		s := obj.String()
		assert.Contains(t, s, `"str":"test"`)
		assert.Contains(t, s, `"num":42`)
		assert.Contains(t, s, `"bool":true`)
	})

	t.Run("MustString", func(t *testing.T) {
		s := obj.MustString()
		assert.Contains(t, s, `"str":"test"`)
		assert.Contains(t, s, `"num":42`)
		assert.Contains(t, s, `"bool":true`)
	})

	t.Run("Interface", func(t *testing.T) {
		i := obj.Interface()
		m, ok := i.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, 3, len(m))
		assert.Equal(t, "test", m["str"])
	})

	t.Run("Append", func(t *testing.T) {
		result := obj.Append("value")
		assert.Error(t, result.Error())
		assert.False(t, result.IsValid())
		assert.Equal(t, ErrTypeAssertion, result.Error())
	})

	t.Run("Raw", func(t *testing.T) {
		s := obj.Raw()
		assert.Equal(t, `{"str":"test","num":42,"bool":true}`, s)
	})

	t.Run("RawFloat", func(t *testing.T) {
		_, ok := obj.RawFloat()
		assert.False(t, ok)
	})

	t.Run("RawString", func(t *testing.T) {
		_, ok := obj.RawString()
		assert.False(t, ok)
	})

	t.Run("Strings", func(t *testing.T) {
		s := obj.Strings()
		assert.Nil(t, s)
	})

	t.Run("Contains", func(t *testing.T) {
		c := obj.Contains("test")
		assert.False(t, c)
	})
}

func TestArrayNodeCoverageInCoverageTest(t *testing.T) {
	// 创建一个带原始数据的根数组节点
	// Create a root array node with raw data
	jsonData := `["first","second","third"]`
	funcs := make(map[string]func(core.Node) core.Node)
	arr := NewArrayNode(
		[]core.Node{
			NewStringNode("first", "", &funcs),
			NewStringNode("second", "", &funcs),
			NewStringNode("third", "", &funcs),
		},
		"",
		&funcs,
	)
	// 为根节点设置原始数据
	// Set raw data for the root node
	arr.(*arrayNode).raw = &jsonData

	t.Run("ForEach", func(t *testing.T) {
		count := 0
		var values []string
		arr.ForEach(func(key interface{}, value core.Node) {
			count++
			values = append(values, value.String())
		})
		assert.Equal(t, 3, count)
		assert.Equal(t, "first", values[0])
		assert.Equal(t, "second", values[1])
		assert.Equal(t, "third", values[2])
	})

	t.Run("String", func(t *testing.T) {
		s := arr.String()
		assert.Equal(t, `["first","second","third"]`, s)
	})

	t.Run("Len", func(t *testing.T) {
		l := arr.Len()
		assert.Equal(t, 3, l)
	})

	t.Run("Interface", func(t *testing.T) {
		i := arr.Interface()
		s, ok := i.([]interface{})
		assert.True(t, ok)
		assert.Equal(t, 3, len(s))
		assert.Equal(t, "first", s[0])
	})

	t.Run("Raw", func(t *testing.T) {
		s := arr.Raw()
		assert.Equal(t, `["first","second","third"]`, s)
	})

	t.Run("RawFloat", func(t *testing.T) {
		_, ok := arr.RawFloat()
		assert.False(t, ok)
	})

	t.Run("RawString", func(t *testing.T) {
		_, ok := arr.RawString()
		assert.False(t, ok)
	})
}

func TestStringNodeCoverage(t *testing.T) {
	// 创建一个字符串节点
	// Create a string node
	funcs := make(map[string]func(core.Node) core.Node)
	str := NewStringNode("test string", "", &funcs)

	t.Run("ForEach", func(t *testing.T) {
		count := 0
		str.ForEach(func(key interface{}, value core.Node) {
			count++
		})
		assert.Equal(t, 0, count)
	})

	t.Run("Len", func(t *testing.T) {
		l := str.Len()
		assert.Equal(t, 11, l) // "test string" 的长度
	})

	t.Run("String", func(t *testing.T) {
		s := str.String()
		assert.Equal(t, "test string", s)
	})

	t.Run("MustString", func(t *testing.T) {
		s := str.MustString()
		assert.Equal(t, "test string", s)
	})

	t.Run("Time", func(t *testing.T) {
		// 测试有效的时间字符串
		// Test with valid time string
		timeStr := NewStringNode("2023-01-01T00:00:00Z", "", &funcs)
		tm := timeStr.Time()
		assert.Equal(t, time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), tm)

		// 测试无效的时间字符串
		// Test with invalid time string
		invalidTimeStr := NewStringNode("invalid", "", &funcs)
		tm = invalidTimeStr.Time()
		assert.Equal(t, time.Time{}, tm)
		assert.NotNil(t, invalidTimeStr.Error())
	})

	t.Run("MustTime", func(t *testing.T) {
		// 测试有效的时间字符串
		// Test with valid time string
		timeStr := NewStringNode("2023-01-01T00:00:00Z", "", &funcs)
		tm := timeStr.MustTime()
		assert.Equal(t, time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), tm)

		// 测试无效的时间字符串 - 应该 panic
		// Test with invalid time string - should panic
		invalidTimeStr := NewStringNode("invalid", "", &funcs)
		defer func() {
			r := recover()
			assert.NotNil(t, r)
		}()
		invalidTimeStr.MustTime()
	})

	t.Run("Array", func(t *testing.T) {
		a := str.Array()
		assert.Nil(t, a)
	})

	t.Run("MustArray", func(t *testing.T) {
		defer func() {
			r := recover()
			assert.NotNil(t, r)
		}()
		str.MustArray()
	})

	t.Run("Interface", func(t *testing.T) {
		i := str.Interface()
		s, ok := i.(string)
		assert.True(t, ok)
		assert.Equal(t, "test string", s)
	})

	t.Run("Filter", func(t *testing.T) {
		result := str.Filter(func(n core.Node) bool { return true })
		assert.Error(t, result.Error())
		assert.False(t, result.IsValid())
		assert.Equal(t, ErrTypeAssertion, result.Error())
	})

	t.Run("Map", func(t *testing.T) {
		result := str.Map(func(n core.Node) interface{} { return nil })
		assert.Error(t, result.Error())
		assert.False(t, result.IsValid())
		assert.Equal(t, ErrTypeAssertion, result.Error())
	})

	t.Run("Set", func(t *testing.T) {
		result := str.Set("key", "value")
		assert.Error(t, result.Error())
		assert.False(t, result.IsValid())
		assert.Equal(t, ErrTypeAssertion, result.Error())
	})

	t.Run("Append", func(t *testing.T) {
		result := str.Append("value")
		assert.Error(t, result.Error())
		assert.False(t, result.IsValid())
		assert.Equal(t, ErrTypeAssertion, result.Error())
	})

	t.Run("Raw", func(t *testing.T) {
		s := str.Raw()
		assert.Equal(t, `"test string"`, s)
	})

	t.Run("RawFloat", func(t *testing.T) {
		_, ok := str.RawFloat()
		assert.False(t, ok)
	})

	t.Run("RawString", func(t *testing.T) {
		s, ok := str.RawString()
		assert.True(t, ok)
		assert.Equal(t, "test string", s)
	})

	t.Run("Strings", func(t *testing.T) {
		s := str.Strings()
		assert.Equal(t, []string{"test string"}, s)
	})

	t.Run("Contains", func(t *testing.T) {
		c := str.Contains("test")
		assert.True(t, c)

		c = str.Contains("xyz")
		assert.False(t, c)
	})
}

func TestNumberNodeCoverage(t *testing.T) {
	// 创建一个数字节点
	// Create a number node
	funcs := make(map[string]func(core.Node) core.Node)
	num := NewNumberNode(42.5, "", &funcs)

	t.Run("ForEach", func(t *testing.T) {
		count := 0
		num.ForEach(func(key interface{}, value core.Node) {
			count++
		})
		assert.Equal(t, 0, count)
	})

	t.Run("Len", func(t *testing.T) {
		l := num.Len()
		assert.Equal(t, 0, l)
	})

	t.Run("String", func(t *testing.T) {
		s := num.String()
		assert.Equal(t, "42.5", s)
	})

	t.Run("MustString", func(t *testing.T) {
		defer func() {
			r := recover()
			assert.NotNil(t, r)
		}()
		num.MustString()
	})

	t.Run("Float", func(t *testing.T) {
		f := num.Float()
		assert.Equal(t, 42.5, f)
	})

	t.Run("MustFloat", func(t *testing.T) {
		f := num.MustFloat()
		assert.Equal(t, 42.5, f)
	})

	t.Run("Int", func(t *testing.T) {
		i := num.Int()
		assert.Equal(t, int64(42), i)
	})

	t.Run("MustInt", func(t *testing.T) {
		i := num.MustInt()
		assert.Equal(t, int64(42), i)
	})

	t.Run("Bool", func(t *testing.T) {
		b := num.Bool()
		assert.False(t, b)
	})

	t.Run("MustBool", func(t *testing.T) {
		defer func() {
			r := recover()
			assert.NotNil(t, r)
		}()
		num.MustBool()
	})

	t.Run("Time", func(t *testing.T) {
		tm := num.Time()
		assert.Equal(t, time.Time{}, tm)
	})

	t.Run("MustTime", func(t *testing.T) {
		defer func() {
			r := recover()
			assert.NotNil(t, r)
		}()
		num.MustTime()
	})

	t.Run("Array", func(t *testing.T) {
		a := num.Array()
		assert.Nil(t, a)
	})

	t.Run("MustArray", func(t *testing.T) {
		defer func() {
			r := recover()
			assert.NotNil(t, r)
		}()
		num.MustArray()
	})

	t.Run("Interface", func(t *testing.T) {
		i := num.Interface()
		f, ok := i.(float64)
		assert.True(t, ok)
		assert.Equal(t, 42.5, f)
	})

	t.Run("Filter", func(t *testing.T) {
		result := num.Filter(func(n core.Node) bool { return true })
		assert.Error(t, result.Error())
		assert.False(t, result.IsValid())
		assert.Equal(t, ErrTypeAssertion, result.Error())
	})

	t.Run("Map", func(t *testing.T) {
		result := num.Map(func(n core.Node) interface{} { return nil })
		assert.Error(t, result.Error())
		assert.False(t, result.IsValid())
		assert.Equal(t, ErrTypeAssertion, result.Error())
	})

	t.Run("Set", func(t *testing.T) {
		result := num.Set("key", "value")
		assert.Error(t, result.Error())
		assert.False(t, result.IsValid())
		assert.Equal(t, ErrTypeAssertion, result.Error())
	})

	t.Run("Append", func(t *testing.T) {
		result := num.Append("value")
		assert.Error(t, result.Error())
		assert.False(t, result.IsValid())
		assert.Equal(t, ErrTypeAssertion, result.Error())
	})

	t.Run("Raw", func(t *testing.T) {
		s := num.Raw()
		assert.Equal(t, "42.5", s)
	})

	t.Run("RawFloat", func(t *testing.T) {
		f, ok := num.RawFloat()
		assert.True(t, ok)
		assert.Equal(t, 42.5, f)
	})

	t.Run("RawString", func(t *testing.T) {
		_, ok := num.RawString()
		assert.False(t, ok)
	})

	t.Run("Strings", func(t *testing.T) {
		s := num.Strings()
		assert.Nil(t, s)
	})

	t.Run("Contains", func(t *testing.T) {
		c := num.Contains("42")
		assert.False(t, c)
	})
}

func TestBoolNodeCoverage(t *testing.T) {
	// 创建一个布尔节点
	// Create a boolean node
	funcs := make(map[string]func(core.Node) core.Node)
	b := NewBoolNode(true, "", &funcs)

	t.Run("ForEach", func(t *testing.T) {
		count := 0
		b.ForEach(func(key interface{}, value core.Node) {
			count++
		})
		assert.Equal(t, 0, count)
	})

	t.Run("Len", func(t *testing.T) {
		l := b.Len()
		assert.Equal(t, 0, l)
	})

	t.Run("String", func(t *testing.T) {
		s := b.String()
		assert.Equal(t, "true", s)
	})

	t.Run("MustString", func(t *testing.T) {
		defer func() {
			r := recover()
			assert.NotNil(t, r)
		}()
		b.MustString()
	})

	t.Run("Float", func(t *testing.T) {
		f := b.Float()
		assert.Equal(t, 0.0, f)
	})

	t.Run("MustFloat", func(t *testing.T) {
		defer func() {
			r := recover()
			assert.NotNil(t, r)
		}()
		b.MustFloat()
	})

	t.Run("Int", func(t *testing.T) {
		i := b.Int()
		assert.Equal(t, int64(0), i)
	})

	t.Run("MustInt", func(t *testing.T) {
		defer func() {
			r := recover()
			assert.NotNil(t, r)
		}()
		b.MustInt()
	})

	t.Run("Bool", func(t *testing.T) {
		boolVal := b.Bool()
		assert.True(t, boolVal)
	})

	t.Run("MustBool", func(t *testing.T) {
		boolVal := b.MustBool()
		assert.True(t, boolVal)
	})

	t.Run("Time", func(t *testing.T) {
		tm := b.Time()
		assert.Equal(t, time.Time{}, tm)
	})

	t.Run("MustTime", func(t *testing.T) {
		defer func() {
			r := recover()
			assert.NotNil(t, r)
		}()
		b.MustTime()
	})

	t.Run("Array", func(t *testing.T) {
		a := b.Array()
		assert.Nil(t, a)
	})

	t.Run("MustArray", func(t *testing.T) {
		defer func() {
			r := recover()
			assert.NotNil(t, r)
		}()
		b.MustArray()
	})

	t.Run("Interface", func(t *testing.T) {
		i := b.Interface()
		boolVal, ok := i.(bool)
		assert.True(t, ok)
		assert.True(t, boolVal)
	})

	t.Run("Filter", func(t *testing.T) {
		result := b.Filter(func(n core.Node) bool { return true })
		assert.Error(t, result.Error())
		assert.False(t, result.IsValid())
		assert.Equal(t, ErrTypeAssertion, result.Error())
	})

	t.Run("Map", func(t *testing.T) {
		result := b.Map(func(n core.Node) interface{} { return nil })
		assert.Error(t, result.Error())
		assert.False(t, result.IsValid())
		assert.Equal(t, ErrTypeAssertion, result.Error())
	})

	t.Run("Set", func(t *testing.T) {
		result := b.Set("key", "value")
		assert.Error(t, result.Error())
		assert.False(t, result.IsValid())
		assert.Equal(t, ErrTypeAssertion, result.Error())
	})

	t.Run("Append", func(t *testing.T) {
		result := b.Append("value")
		assert.Error(t, result.Error())
		assert.False(t, result.IsValid())
		assert.Equal(t, ErrTypeAssertion, result.Error())
	})

	t.Run("Raw", func(t *testing.T) {
		s := b.Raw()
		assert.Equal(t, "true", s)
	})

	t.Run("RawFloat", func(t *testing.T) {
		_, ok := b.RawFloat()
		assert.False(t, ok)
	})

	t.Run("RawString", func(t *testing.T) {
		_, ok := b.RawString()
		assert.False(t, ok)
	})

	t.Run("Strings", func(t *testing.T) {
		s := b.Strings()
		assert.Nil(t, s)
	})

	t.Run("Contains", func(t *testing.T) {
		c := b.Contains("true")
		assert.False(t, c)
	})
}

func TestNullNodeCoverage(t *testing.T) {
	// 创建一个空节点
	// Create a null node
	funcs := make(map[string]func(core.Node) core.Node)
	n := NewNullNode("", &funcs)

	t.Run("ForEach", func(t *testing.T) {
		count := 0
		n.ForEach(func(key interface{}, value core.Node) {
			count++
		})
		assert.Equal(t, 0, count)
	})

	t.Run("Len", func(t *testing.T) {
		l := n.Len()
		assert.Equal(t, 0, l)
	})

	t.Run("String", func(t *testing.T) {
		s := n.String()
		assert.Equal(t, "null", s)
	})

	t.Run("MustString", func(t *testing.T) {
		defer func() {
			r := recover()
			assert.NotNil(t, r)
		}()
		n.MustString()
	})

	t.Run("Float", func(t *testing.T) {
		f := n.Float()
		assert.Equal(t, 0.0, f)
	})

	t.Run("MustFloat", func(t *testing.T) {
		defer func() {
			r := recover()
			assert.NotNil(t, r)
		}()
		n.MustFloat()
	})

	t.Run("Int", func(t *testing.T) {
		i := n.Int()
		assert.Equal(t, int64(0), i)
	})

	t.Run("MustInt", func(t *testing.T) {
		defer func() {
			r := recover()
			assert.NotNil(t, r)
		}()
		n.MustInt()
	})

	t.Run("Bool", func(t *testing.T) {
		b := n.Bool()
		assert.False(t, b)
	})

	t.Run("MustBool", func(t *testing.T) {
		defer func() {
			r := recover()
			assert.NotNil(t, r)
		}()
		n.MustBool()
	})

	t.Run("Time", func(t *testing.T) {
		tm := n.Time()
		assert.Equal(t, time.Time{}, tm)
	})

	t.Run("MustTime", func(t *testing.T) {
		defer func() {
			r := recover()
			assert.NotNil(t, r)
		}()
		n.MustTime()
	})

	t.Run("Array", func(t *testing.T) {
		a := n.Array()
		assert.Nil(t, a)
	})

	t.Run("MustArray", func(t *testing.T) {
		defer func() {
			r := recover()
			assert.NotNil(t, r)
		}()
		n.MustArray()
	})

	t.Run("Interface", func(t *testing.T) {
		i := n.Interface()
		assert.Nil(t, i)
	})

	t.Run("Filter", func(t *testing.T) {
		result := n.Filter(func(n core.Node) bool { return true })
		assert.Error(t, result.Error())
		assert.False(t, result.IsValid())
		assert.Equal(t, ErrTypeAssertion, result.Error())
	})

	t.Run("Map", func(t *testing.T) {
		result := n.Map(func(n core.Node) interface{} { return nil })
		assert.Error(t, result.Error())
		assert.False(t, result.IsValid())
		assert.Equal(t, ErrTypeAssertion, result.Error())
	})

	t.Run("Set", func(t *testing.T) {
		result := n.Set("key", "value")
		assert.Error(t, result.Error())
		assert.False(t, result.IsValid())
		assert.Equal(t, ErrTypeAssertion, result.Error())
	})

	t.Run("Append", func(t *testing.T) {
		result := n.Append("value")
		assert.Error(t, result.Error())
		assert.False(t, result.IsValid())
		assert.Equal(t, ErrTypeAssertion, result.Error())
	})

	t.Run("Raw", func(t *testing.T) {
		s := n.Raw()
		assert.Equal(t, "null", s)
	})

	t.Run("RawFloat", func(t *testing.T) {
		_, ok := n.RawFloat()
		assert.False(t, ok)
	})

	t.Run("RawString", func(t *testing.T) {
		_, ok := n.RawString()
		assert.False(t, ok)
	})

	t.Run("Strings", func(t *testing.T) {
		s := n.Strings()
		assert.Nil(t, s)
	})

	t.Run("Contains", func(t *testing.T) {
		c := n.Contains("null")
		assert.False(t, c)
	})
}

func TestInvalidNodeCoverage(t *testing.T) {
	// 创建一个无效节点
	// Create an invalid node
	invalid := NewInvalidNode("test", ErrNotFound)

	t.Run("String", func(t *testing.T) {
		s := invalid.String()
		assert.Equal(t, "", s)
	})

	t.Run("MustString", func(t *testing.T) {
		defer func() {
			r := recover()
			assert.NotNil(t, r)
			assert.Equal(t, ErrNotFound, r)
		}()
		invalid.MustString()
	})

	t.Run("Float", func(t *testing.T) {
		f := invalid.Float()
		assert.Equal(t, 0.0, f)
	})

	t.Run("MustFloat", func(t *testing.T) {
		defer func() {
			r := recover()
			assert.NotNil(t, r)
			assert.Equal(t, ErrNotFound, r)
		}()
		invalid.MustFloat()
	})

	t.Run("Int", func(t *testing.T) {
		i := invalid.Int()
		assert.Equal(t, int64(0), i)
	})

	t.Run("MustInt", func(t *testing.T) {
		defer func() {
			r := recover()
			assert.NotNil(t, r)
			assert.Equal(t, ErrNotFound, r)
		}()
		invalid.MustInt()
	})

	t.Run("Bool", func(t *testing.T) {
		b := invalid.Bool()
		assert.False(t, b)
	})

	t.Run("MustBool", func(t *testing.T) {
		defer func() {
			r := recover()
			assert.NotNil(t, r)
			assert.Equal(t, ErrNotFound, r)
		}()
		invalid.MustBool()
	})

	t.Run("Time", func(t *testing.T) {
		tm := invalid.Time()
		assert.Equal(t, time.Time{}, tm)
	})

	t.Run("MustTime", func(t *testing.T) {
		defer func() {
			r := recover()
			assert.NotNil(t, r)
			assert.Equal(t, ErrNotFound, r)
		}()
		invalid.MustTime()
	})

	t.Run("Array", func(t *testing.T) {
		a := invalid.Array()
		assert.Nil(t, a)
	})

	t.Run("MustArray", func(t *testing.T) {
		defer func() {
			r := recover()
			assert.NotNil(t, r)
			assert.Equal(t, ErrNotFound, r)
		}()
		invalid.MustArray()
	})

	t.Run("Interface", func(t *testing.T) {
		i := invalid.Interface()
		assert.Nil(t, i)
	})

	t.Run("ForEach", func(t *testing.T) {
		count := 0
		invalid.ForEach(func(key interface{}, value core.Node) {
			count++
		})
		assert.Equal(t, 0, count)
	})

	t.Run("Len", func(t *testing.T) {
		l := invalid.Len()
		assert.Equal(t, 0, l)
	})

	t.Run("Filter", func(t *testing.T) {
		result := invalid.Filter(func(n core.Node) bool { return true })
		assert.Equal(t, invalid, result)
	})

	t.Run("Map", func(t *testing.T) {
		result := invalid.Map(func(n core.Node) interface{} { return nil })
		assert.Equal(t, invalid, result)
	})

	t.Run("Set", func(t *testing.T) {
		result := invalid.Set("key", "value")
		assert.Equal(t, invalid, result)
	})

	t.Run("Append", func(t *testing.T) {
		result := invalid.Append("value")
		assert.Equal(t, invalid, result)
	})

	t.Run("RawFloat", func(t *testing.T) {
		_, ok := invalid.RawFloat()
		assert.False(t, ok)
	})

	t.Run("RawString", func(t *testing.T) {
		_, ok := invalid.RawString()
		assert.False(t, ok)
	})

	t.Run("Strings", func(t *testing.T) {
		s := invalid.Strings()
		assert.Nil(t, s)
	})

	t.Run("Contains", func(t *testing.T) {
		c := invalid.Contains("test")
		assert.False(t, c)
	})

	t.Run("Func", func(t *testing.T) {
		result := invalid.RegisterFunc("test", func(n core.Node) core.Node { return n })
		assert.Equal(t, invalid, result)
	})

	t.Run("CallFunc", func(t *testing.T) {
		result := invalid.CallFunc("test")
		assert.Equal(t, invalid, result)
	})

	t.Run("RemoveFunc", func(t *testing.T) {
		result := invalid.RemoveFunc("test")
		assert.Equal(t, invalid, result)
	})
}

func TestQueryAndParseCoverage(t *testing.T) {
	// 测试 ParseQuery 函数覆盖率
	// Test ParseQuery function coverage
	t.Run("ParseQuery", func(t *testing.T) {
		// 测试普通路径
		// Test normal path
		ops, err := ParseQuery("a.b.c")
		assert.NoError(t, err)
		assert.Equal(t, 3, len(ops))
		assert.Equal(t, OpGet, ops[0].Type)
		assert.Equal(t, "a", ops[0].Key)
		assert.Equal(t, OpGet, ops[1].Type)
		assert.Equal(t, "b", ops[1].Key)
		assert.Equal(t, OpGet, ops[2].Type)
		assert.Equal(t, "c", ops[2].Key)

		// 测试带数组索引的路径
		// Test path with array index
		ops, err = ParseQuery("a[0].b")
		assert.NoError(t, err)
		assert.Equal(t, 3, len(ops))
		assert.Equal(t, OpGet, ops[0].Type)
		assert.Equal(t, "a", ops[0].Key)
		assert.Equal(t, OpIndex, ops[1].Type)
		assert.Equal(t, 0, ops[1].Index)
		assert.Equal(t, OpGet, ops[2].Type)
		assert.Equal(t, "b", ops[2].Key)

		// 测试带函数的路径
		// Test path with function
		ops, err = ParseQuery("a[@func].b")
		assert.NoError(t, err)
		assert.Equal(t, 3, len(ops))
		assert.Equal(t, OpGet, ops[0].Type)
		assert.Equal(t, "a", ops[0].Key)
		assert.Equal(t, OpFunc, ops[1].Type)
		assert.Equal(t, "func", ops[1].Func)
		assert.Equal(t, OpGet, ops[2].Type)
		assert.Equal(t, "b", ops[2].Key)

		// 测试带多个索引的路径
		// Test path with multiple indexes
		ops, err = ParseQuery("a[0][1]")
		assert.NoError(t, err)
		assert.Equal(t, 3, len(ops))
		assert.Equal(t, OpGet, ops[0].Type)
		assert.Equal(t, "a", ops[0].Key)
		assert.Equal(t, OpIndex, ops[1].Type)
		assert.Equal(t, 0, ops[1].Index)
		assert.Equal(t, OpIndex, ops[2].Type)
		assert.Equal(t, 1, ops[2].Index)

		// 测试错误情况
		// Test error cases
		_, err = ParseQuery("a[")
		assert.Error(t, err)

		_, err = ParseQuery("a[b]")
		assert.Error(t, err)

		_, err = ParseQuery("a[0]extra")
		assert.Error(t, err)
	})

	// 测试 ParseJSONToNode 函数覆盖率
	// Test ParseJSONToNode function coverage
	t.Run("ParseJSONToNode", func(t *testing.T) {
		// 测试有效的 JSON
		// Test valid JSON
		node, err := ParseJSONToNode(`{"a": 1, "b": [2, 3]}`)
		assert.NoError(t, err)
		assert.True(t, node.IsValid())
		assert.Equal(t, core.ObjectNode, node.Type())
		assert.Equal(t, 1.0, node.Get("a").Float())

		// 测试无效的 JSON
		// Test invalid JSON
		node, err = ParseJSONToNode(`{"a": }`)
		assert.Error(t, err)
		assert.Nil(t, node)
	})

	// 测试 buildNode 函数覆盖率的边界情况
	// Test buildNode function coverage with edge cases
	t.Run("BuildNodeEdgeCases", func(t *testing.T) {
		funcs := make(map[string]func(core.Node) core.Node)

		// 测试未知类型（应创建无效节点）
		// Test with unknown type (should create invalid node)
		node := buildNode(struct{}{}, "", &funcs)
		assert.Equal(t, core.InvalidNode, node.Type())
		assert.Equal(t, ErrInvalidNode, node.Error())
	})
}

func TestNodeMethodCoverage(t *testing.T) {
	// 创建一个函数映射
	// Create a function map
	funcs := make(map[string]func(core.Node) core.Node)

	// 测试 ObjectNode 中未完全覆盖的方法
	// Test ObjectNode methods not fully covered
	t.Run("ObjectNodeMethods", func(t *testing.T) {
		obj := NewObjectNode(map[string]core.Node{
			"key1": NewStringNode("value1", "", &funcs),
			"key2": NewStringNode("value2", "", &funcs),
		}, "", &funcs)

		// 测试使用不存在的键 Get
		// Test Get with non-existing key
		result := obj.Get("nonexistent")
		assert.Error(t, result.Error())
		assert.False(t, result.IsValid())
		assert.Equal(t, ErrNotFound, result.Error())

		// 测试在对象上使用 Index (应失败)
		// Test Index on object (should fail)
		result = obj.Index(0)
		assert.Error(t, result.Error())
		assert.False(t, result.IsValid())
		assert.Equal(t, ErrTypeAssertion, result.Error())

		// 测试 Func, CallFunc, RemoveFunc
		// Test Func, CallFunc, RemoveFunc
		obj.RegisterFunc("testFunc", func(n core.Node) core.Node {
			return NewStringNode("function_result", "", &funcs)
		})

		result = obj.CallFunc("testFunc")
		if assert.NoError(t, result.Error()) {
			assert.True(t, result.IsValid())
			assert.Equal(t, "function_result", result.String())
		}

		obj.RemoveFunc("testFunc")
		result = obj.CallFunc("testFunc")
		assert.Error(t, result.Error())
		assert.False(t, result.IsValid())
	})

	// 测试 ArrayNode 中未完全覆盖的方法
	// Test ArrayNode methods not fully covered
	t.Run("ArrayNodeMethods", func(t *testing.T) {
		arr := NewArrayNode([]core.Node{
			NewStringNode("item1", "", &funcs),
			NewStringNode("item2", "", &funcs),
		}, "", &funcs)

		// 测试在数组上使用 Get (应失败)
		// Test Get on array (should fail)
		result := arr.Get("key")
		assert.Error(t, result.Error())
		assert.False(t, result.IsValid())
		assert.Equal(t, ErrTypeAssertion, result.Error())

		// 测试越界索引
		// Test Index with out of bounds
		result = arr.Index(10)
		assert.Error(t, result.Error())
		assert.False(t, result.IsValid())
		assert.Equal(t, ErrIndexOutOfBounds, result.Error())

		// 测试 Func, CallFunc, RemoveFunc
		// Test Func, CallFunc, RemoveFunc
		arr.RegisterFunc("doubleFunc", func(n core.Node) core.Node {
			return NewNumberNode(n.Float()*2, "", &funcs)
		})

		// 在数组上调用函数 - 应应用于每个元素
		// Call function on array - should apply to each element
		numArr := NewArrayNode([]core.Node{
			NewNumberNode(1, "", &funcs),
			NewNumberNode(2, "", &funcs),
		}, "", &funcs)

		numArr.RegisterFunc("doubleFunc", func(n core.Node) core.Node {
			return NewNumberNode(n.Float()*2, "", &funcs)
		})

		result = numArr.CallFunc("doubleFunc")
		if assert.NoError(t, result.Error()) {
			assert.True(t, result.IsValid())
			assert.Equal(t, 2, result.Len())
			if assert.NoError(t, result.Index(0).Error()) {
				assert.Equal(t, 2.0, result.Index(0).Float())
			}
			if assert.NoError(t, result.Index(1).Error()) {
				assert.Equal(t, 4.0, result.Index(1).Float())
			}
		}

		numArr.RemoveFunc("doubleFunc")
		result = numArr.CallFunc("doubleFunc")
		assert.Error(t, result.Error())
		assert.False(t, result.IsValid())
	})

	// 测试 StringNode 中未完全覆盖的方法
	// Test StringNode methods not fully covered
	t.Run("StringNodeMethods", func(t *testing.T) {
		str := NewStringNode("test", "", &funcs)

		// 测试 Func, CallFunc, RemoveFunc
		// Test Func, CallFunc, RemoveFunc
		str.RegisterFunc("upperFunc", func(n core.Node) core.Node {
			return NewStringNode(strings.ToUpper(n.String()), "", &funcs)
		})

		result := str.CallFunc("upperFunc")
		if assert.NoError(t, result.Error()) {
			assert.True(t, result.IsValid())
			assert.Equal(t, "TEST", result.String())
		}

		str.RemoveFunc("upperFunc")
		result = str.CallFunc("upperFunc")
		assert.Error(t, result.Error())
		assert.False(t, result.IsValid())
	})

	// 测试 NumberNode 中未完全覆盖的方法
	// Test NumberNode methods not fully covered
	t.Run("NumberNodeMethods", func(t *testing.T) {
		num := NewNumberNode(42, "", &funcs)

		// 测试 Func, CallFunc, RemoveFunc
		// Test Func, CallFunc, RemoveFunc
		num.RegisterFunc("doubleFunc", func(n core.Node) core.Node {
			return NewNumberNode(n.Float()*2, "", &funcs)
		})

		result := num.CallFunc("doubleFunc")
		if assert.NoError(t, result.Error()) {
			assert.True(t, result.IsValid())
			assert.Equal(t, 84.0, result.Float())
		}

		num.RemoveFunc("doubleFunc")
		result = num.CallFunc("doubleFunc")
		assert.Error(t, result.Error())
		assert.False(t, result.IsValid())
	})

	// 测试 BoolNode 中未完全覆盖的方法
	// Test BoolNode methods not fully covered
	t.Run("BoolNodeMethods", func(t *testing.T) {
		b := NewBoolNode(true, "", &funcs)

		// 测试 Func, CallFunc, RemoveFunc
		// Test Func, CallFunc, RemoveFunc
		b.RegisterFunc("invertFunc", func(n core.Node) core.Node {
			return NewBoolNode(!n.Bool(), "", &funcs)
		})

		result := b.CallFunc("invertFunc")
		if assert.NoError(t, result.Error()) {
			assert.True(t, result.IsValid())
			assert.False(t, result.Bool())
		}

		b.RemoveFunc("invertFunc")
		result = b.CallFunc("invertFunc")
		assert.Error(t, result.Error())
		assert.False(t, result.IsValid())
	})

	// 测试 NullNode 中未完全覆盖的方法
	// Test NullNode methods not fully covered
	t.Run("NullNodeMethods", func(t *testing.T) {
		n := NewNullNode("", &funcs)

		// 测试 Func, CallFunc, RemoveFunc
		// Test Func, CallFunc, RemoveFunc
		n.RegisterFunc("nullFunc", func(n core.Node) core.Node {
			return NewStringNode("null_result", "", &funcs)
		})

		result := n.CallFunc("nullFunc")
		if assert.NoError(t, result.Error()) {
			assert.True(t, result.IsValid())
			assert.Equal(t, "null_result", result.String())
		}

		n.RemoveFunc("nullFunc")
		result = n.CallFunc("nullFunc")
		assert.Error(t, result.Error())
		assert.False(t, result.IsValid())
	})
}

func TestArrayNodeSetMethod(t *testing.T) {
	// 创建一个函数映射
	// Create a function map
	funcs := make(map[string]func(core.Node) core.Node)

	t.Run("SetOnArrayOfObjects", func(t *testing.T) {
		// 创建对象数组
		// Create array of objects
		arr := NewArrayNode([]core.Node{
			NewObjectNode(map[string]core.Node{"a": NewNumberNode(1, "", &funcs)}, "", &funcs),
			NewObjectNode(map[string]core.Node{"b": NewNumberNode(2, "", &funcs)}, "", &funcs),
		}, "", &funcs)

		// 在所有对象上设置一个新字段
		// Set a new field on all objects
		result := arr.Set("newField", "newValue")
		assert.NoError(t, result.Error())
		assert.True(t, result.IsValid())

		// 检查字段是否已添加到所有对象
		// Check that the field was added to all objects
		if assert.NoError(t, arr.Index(0).Error()) {
			field1 := arr.Index(0).Get("newField")
			if assert.NoError(t, field1.Error()) {
				assert.Equal(t, "newValue", field1.String())
			}
		}
		if assert.NoError(t, arr.Index(1).Error()) {
			field2 := arr.Index(1).Get("newField")
			if assert.NoError(t, field2.Error()) {
				assert.Equal(t, "newValue", field2.String())
			}
		}
	})

	t.Run("SetOnArrayOfNonObjects", func(t *testing.T) {
		// 创建非对象数组
		// Create array of non-objects
		arr := NewArrayNode([]core.Node{
			NewNumberNode(1, "", &funcs),
			NewNumberNode(2, "", &funcs),
		}, "", &funcs)

		// 尝试设置字段 - 应该失败
		// Try to set a field - should fail
		result := arr.Set("newField", "newValue")
		assert.Error(t, result.Error())
		assert.False(t, result.IsValid())
		assert.Equal(t, ErrTypeAssertion, result.Error())
	})

	t.Run("SetWithErrorPropagation", func(t *testing.T) {
		// 创建一个带无效节点的数组
		// Create array with an invalid node
		invalid := NewInvalidNode("", ErrNotFound)
		arr := NewArrayNode([]core.Node{
			NewObjectNode(map[string]core.Node{"a": NewNumberNode(1, "", &funcs)}, "", &funcs),
			invalid,
		}, "", &funcs)

		// 尝试设置字段 - 应该传播无效节点，而不是产生类型错误
		// Try to set a field - should propagate the invalid node, not produce a type error
		result := arr.Set("newField", "newValue")
		assert.Error(t, result.Error())
		assert.False(t, result.IsValid())
		// 错误应该是来自无效节点的错误
		// The error should be the one from the invalid node
		assert.Equal(t, ErrNotFound, result.Error())
	})
}

func TestArrayNodeStringsMethod(t *testing.T) {
	// 创建一个函数映射
	// Create a function map
	funcs := make(map[string]func(core.Node) core.Node)

	t.Run("StringsOnStringArray", func(t *testing.T) {
		// 创建字符串数组
		// Create a string array
		arr := NewArrayNode([]core.Node{
			NewStringNode("first", "", &funcs),
			NewStringNode("second", "", &funcs),
			NewStringNode("third", "", &funcs),
		}, "", &funcs)

		strings := arr.Strings()
		assert.Equal(t, []string{"first", "second", "third"}, strings)
	})

	t.Run("StringsOnMixedArray", func(t *testing.T) {
		// 创建混合类型数组
		// Create a mixed-type array
		arr := NewArrayNode([]core.Node{
			NewStringNode("first", "", &funcs),
			NewNumberNode(2, "", &funcs),
			NewStringNode("third", "", &funcs),
		}, "", &funcs)

		strings := arr.Strings()
		assert.Nil(t, strings)
		assert.NotNil(t, arr.Error())
	})
}

func TestArrayNodeContainsMethod(t *testing.T) {
	// 创建一个函数映射
	// Create a function map
	funcs := make(map[string]func(core.Node) core.Node)

	t.Run("ContainsInStringArray", func(t *testing.T) {
		// 创建字符串数组
		// Create a string array
		arr := NewArrayNode([]core.Node{
			NewStringNode("first", "", &funcs),
			NewStringNode("second", "", &funcs),
			NewStringNode("third", "", &funcs),
		}, "", &funcs)

		assert.True(t, arr.Contains("second"))
		assert.False(t, arr.Contains("fourth"))
	})

	t.Run("ContainsInNonStringArray", func(t *testing.T) {
		// 创建非字符串数组
		// Create a non-string array
		arr := NewArrayNode([]core.Node{
			NewNumberNode(1, "", &funcs),
			NewNumberNode(2, "", &funcs),
		}, "", &funcs)

		assert.False(t, arr.Contains("1"))
	})
}
