package engine

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestObjectNodeCoverage(t *testing.T) {
	// Create a root object node with raw data
	jsonData := `{"str":"test","num":42,"bool":true}`
	funcs := make(map[string]func(Node) Node)
	obj := NewObjectNode(
		map[string]Node{
			"str":  NewStringNode("test", "", &funcs),
			"num":  NewNumberNode(42, "", &funcs),
			"bool": NewBoolNode(true, "", &funcs),
		},
		"",
		&funcs,
	)
	// Set raw data for the root node
	obj.(*objectNode).raw = &jsonData

	t.Run("ForEach", func(t *testing.T) {
		count := 0
		var keys []string
		obj.ForEach(func(key interface{}, value Node) {
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

func TestArrayNodeCoverage(t *testing.T) {
	jsonData := `["first","second","third"]`
	funcs := make(map[string]func(Node) Node)
	arr := NewArrayNode(
		[]Node{
			NewStringNode("first", "", &funcs),
			NewStringNode("second", "", &funcs),
			NewStringNode("third", "", &funcs),
		},
		"",
		&funcs,
	)
	// Set raw data for the root node
	arr.(*arrayNode).raw = &jsonData

	t.Run("ForEach", func(t *testing.T) {
		count := 0
		var values []string
		arr.ForEach(func(key interface{}, value Node) {
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
	funcs := make(map[string]func(Node) Node)
	str := NewStringNode("test string", "", &funcs)

	t.Run("ForEach", func(t *testing.T) {
		count := 0
		str.ForEach(func(key interface{}, value Node) {
			count++
		})
		assert.Equal(t, 0, count)
	})

	t.Run("Len", func(t *testing.T) {
		l := str.Len()
		assert.Equal(t, 11, l) // length of "test string"
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
		// Test with valid time string
		timeStr := NewStringNode("2023-01-01T00:00:00Z", "", &funcs)
		tm := timeStr.Time()
		assert.Equal(t, time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), tm)

		// Test with invalid time string
		invalidTimeStr := NewStringNode("invalid", "", &funcs)
		tm = invalidTimeStr.Time()
		assert.Equal(t, time.Time{}, tm)
		assert.NotNil(t, invalidTimeStr.Error())
	})

	t.Run("MustTime", func(t *testing.T) {
		// Test with valid time string
		timeStr := NewStringNode("2023-01-01T00:00:00Z", "", &funcs)
		tm := timeStr.MustTime()
		assert.Equal(t, time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), tm)

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
		result := str.Filter(func(n Node) bool { return true })
		assert.False(t, result.IsValid())
		assert.Equal(t, ErrTypeAssertion, result.Error())
	})

	t.Run("Map", func(t *testing.T) {
		result := str.Map(func(n Node) interface{} { return nil })
		assert.False(t, result.IsValid())
		assert.Equal(t, ErrTypeAssertion, result.Error())
	})

	t.Run("Set", func(t *testing.T) {
		result := str.Set("key", "value")
		assert.False(t, result.IsValid())
		assert.Equal(t, ErrTypeAssertion, result.Error())
	})

	t.Run("Append", func(t *testing.T) {
		result := str.Append("value")
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
	funcs := make(map[string]func(Node) Node)
	num := NewNumberNode(42.5, "", &funcs)

	t.Run("ForEach", func(t *testing.T) {
		count := 0
		num.ForEach(func(key interface{}, value Node) {
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
		result := num.Filter(func(n Node) bool { return true })
		assert.False(t, result.IsValid())
		assert.Equal(t, ErrTypeAssertion, result.Error())
	})

	t.Run("Map", func(t *testing.T) {
		result := num.Map(func(n Node) interface{} { return nil })
		assert.False(t, result.IsValid())
		assert.Equal(t, ErrTypeAssertion, result.Error())
	})

	t.Run("Set", func(t *testing.T) {
		result := num.Set("key", "value")
		assert.False(t, result.IsValid())
		assert.Equal(t, ErrTypeAssertion, result.Error())
	})

	t.Run("Append", func(t *testing.T) {
		result := num.Append("value")
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
	funcs := make(map[string]func(Node) Node)
	b := NewBoolNode(true, "", &funcs)

	t.Run("ForEach", func(t *testing.T) {
		count := 0
		b.ForEach(func(key interface{}, value Node) {
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
		result := b.Filter(func(n Node) bool { return true })
		assert.False(t, result.IsValid())
		assert.Equal(t, ErrTypeAssertion, result.Error())
	})

	t.Run("Map", func(t *testing.T) {
		result := b.Map(func(n Node) interface{} { return nil })
		assert.False(t, result.IsValid())
		assert.Equal(t, ErrTypeAssertion, result.Error())
	})

	t.Run("Set", func(t *testing.T) {
		result := b.Set("key", "value")
		assert.False(t, result.IsValid())
		assert.Equal(t, ErrTypeAssertion, result.Error())
	})

	t.Run("Append", func(t *testing.T) {
		result := b.Append("value")
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
	funcs := make(map[string]func(Node) Node)
	n := NewNullNode("", &funcs)

	t.Run("ForEach", func(t *testing.T) {
		count := 0
		n.ForEach(func(key interface{}, value Node) {
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
		result := n.Filter(func(n Node) bool { return true })
		assert.False(t, result.IsValid())
		assert.Equal(t, ErrTypeAssertion, result.Error())
	})

	t.Run("Map", func(t *testing.T) {
		result := n.Map(func(n Node) interface{} { return nil })
		assert.False(t, result.IsValid())
		assert.Equal(t, ErrTypeAssertion, result.Error())
	})

	t.Run("Set", func(t *testing.T) {
		result := n.Set("key", "value")
		assert.False(t, result.IsValid())
		assert.Equal(t, ErrTypeAssertion, result.Error())
	})

	t.Run("Append", func(t *testing.T) {
		result := n.Append("value")
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
		invalid.ForEach(func(key interface{}, value Node) {
			count++
		})
		assert.Equal(t, 0, count)
	})

	t.Run("Len", func(t *testing.T) {
		l := invalid.Len()
		assert.Equal(t, 0, l)
	})

	t.Run("Filter", func(t *testing.T) {
		result := invalid.Filter(func(n Node) bool { return true })
		assert.Equal(t, invalid, result)
	})

	t.Run("Map", func(t *testing.T) {
		result := invalid.Map(func(n Node) interface{} { return nil })
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
		result := invalid.Func("test", func(n Node) Node { return n })
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
	// Test ParseQuery function coverage
	t.Run("ParseQuery", func(t *testing.T) {
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

		// Test error cases
		_, err = ParseQuery("a[")
		assert.Error(t, err)

		_, err = ParseQuery("a[b]")
		assert.Error(t, err)

		_, err = ParseQuery("a[0]extra")
		assert.Error(t, err)
	})

	// Test ParseJSONToNode function coverage
	t.Run("ParseJSONToNode", func(t *testing.T) {
		// Test valid JSON
		node, err := ParseJSONToNode(`{"a": 1, "b": [2, 3]}`)
		assert.NoError(t, err)
		assert.True(t, node.IsValid())
		assert.Equal(t, ObjectNode, node.Type())
		assert.Equal(t, 1.0, node.Get("a").Float())

		// Test invalid JSON
		node, err = ParseJSONToNode(`{"a": }`)
		assert.Error(t, err)
		assert.Nil(t, node)
	})

	// Test buildNode function coverage with edge cases
	t.Run("BuildNodeEdgeCases", func(t *testing.T) {
		funcs := make(map[string]func(Node) Node)

		// Test with unknown type (should create invalid node)
		node := buildNode(struct{}{}, "", &funcs, nil)
		assert.Equal(t, InvalidNode, node.Type())
		assert.Equal(t, ErrInvalidNode, node.Error())
	})
}

func TestNodeMethodCoverage(t *testing.T) {
	funcs := make(map[string]func(Node) Node)

	// Test ObjectNode methods not fully covered
	t.Run("ObjectNodeMethods", func(t *testing.T) {
		obj := NewObjectNode(map[string]Node{
			"key1": NewStringNode("value1", "", &funcs),
			"key2": NewStringNode("value2", "", &funcs),
		}, "", &funcs)

		// Test Get with non-existing key
		result := obj.Get("nonexistent")
		assert.False(t, result.IsValid())
		assert.Equal(t, ErrNotFound, result.Error())

		// Test Index on object (should fail)
		result = obj.Index(0)
		assert.False(t, result.IsValid())
		assert.Equal(t, ErrTypeAssertion, result.Error())

		// Test Func, CallFunc, RemoveFunc
		obj.Func("testFunc", func(n Node) Node {
			return NewStringNode("function_result", "", &funcs)
		})

		result = obj.CallFunc("testFunc")
		assert.True(t, result.IsValid())
		assert.Equal(t, "function_result", result.String())

		obj.RemoveFunc("testFunc")
		result = obj.CallFunc("testFunc")
		assert.False(t, result.IsValid())
	})

	// Test ArrayNode methods not fully covered
	t.Run("ArrayNodeMethods", func(t *testing.T) {
		arr := NewArrayNode([]Node{
			NewStringNode("item1", "", &funcs),
			NewStringNode("item2", "", &funcs),
		}, "", &funcs)

		// Test Get on array (should fail)
		result := arr.Get("key")
		assert.False(t, result.IsValid())
		assert.Equal(t, ErrTypeAssertion, result.Error())

		// Test Index with out of bounds
		result = arr.Index(10)
		assert.False(t, result.IsValid())
		assert.Equal(t, ErrIndexOutOfBounds, result.Error())

		// Test Func, CallFunc, RemoveFunc
		arr.Func("doubleFunc", func(n Node) Node {
			return NewNumberNode(n.Float()*2, "", &funcs)
		})

		// Call function on array - should apply to each element
		numArr := NewArrayNode([]Node{
			NewNumberNode(1, "", &funcs),
			NewNumberNode(2, "", &funcs),
		}, "", &funcs)

		numArr.Func("doubleFunc", func(n Node) Node {
			return NewNumberNode(n.Float()*2, "", &funcs)
		})

		result = numArr.CallFunc("doubleFunc")
		assert.True(t, result.IsValid())
		assert.Equal(t, 2, result.Len())
		assert.Equal(t, 2.0, result.Index(0).Float())
		assert.Equal(t, 4.0, result.Index(1).Float())

		numArr.RemoveFunc("doubleFunc")
		result = numArr.CallFunc("doubleFunc")
		assert.False(t, result.IsValid())
	})

	// Test StringNode methods not fully covered
	t.Run("StringNodeMethods", func(t *testing.T) {
		str := NewStringNode("test", "", &funcs)

		// Test Func, CallFunc, RemoveFunc
		str.Func("upperFunc", func(n Node) Node {
			return NewStringNode(strings.ToUpper(n.String()), "", &funcs)
		})

		result := str.CallFunc("upperFunc")
		assert.True(t, result.IsValid())
		assert.Equal(t, "TEST", result.String())

		str.RemoveFunc("upperFunc")
		result = str.CallFunc("upperFunc")
		assert.False(t, result.IsValid())
	})

	// Test NumberNode methods not fully covered
	t.Run("NumberNodeMethods", func(t *testing.T) {
		num := NewNumberNode(42, "", &funcs)

		// Test Func, CallFunc, RemoveFunc
		num.Func("doubleFunc", func(n Node) Node {
			return NewNumberNode(n.Float()*2, "", &funcs)
		})

		result := num.CallFunc("doubleFunc")
		assert.True(t, result.IsValid())
		assert.Equal(t, 84.0, result.Float())

		num.RemoveFunc("doubleFunc")
		result = num.CallFunc("doubleFunc")
		assert.False(t, result.IsValid())
	})

	// Test BoolNode methods not fully covered
	t.Run("BoolNodeMethods", func(t *testing.T) {
		b := NewBoolNode(true, "", &funcs)

		// Test Func, CallFunc, RemoveFunc
		b.Func("invertFunc", func(n Node) Node {
			return NewBoolNode(!n.Bool(), "", &funcs)
		})

		result := b.CallFunc("invertFunc")
		assert.True(t, result.IsValid())
		assert.False(t, result.Bool())

		b.RemoveFunc("invertFunc")
		result = b.CallFunc("invertFunc")
		assert.False(t, result.IsValid())
	})

	// Test NullNode methods not fully covered
	t.Run("NullNodeMethods", func(t *testing.T) {
		n := NewNullNode("", &funcs)

		// Test Func, CallFunc, RemoveFunc
		n.Func("nullFunc", func(n Node) Node {
			return NewStringNode("null_result", "", &funcs)
		})

		result := n.CallFunc("nullFunc")
		assert.True(t, result.IsValid())
		assert.Equal(t, "null_result", result.String())

		n.RemoveFunc("nullFunc")
		result = n.CallFunc("nullFunc")
		assert.False(t, result.IsValid())
	})
}

func TestArrayNodeSetMethod(t *testing.T) {
	funcs := make(map[string]func(Node) Node)

	t.Run("SetOnArrayOfObjects", func(t *testing.T) {
		// Create array of objects
		arr := NewArrayNode([]Node{
			NewObjectNode(map[string]Node{"a": NewNumberNode(1, "", &funcs)}, "", &funcs),
			NewObjectNode(map[string]Node{"b": NewNumberNode(2, "", &funcs)}, "", &funcs),
		}, "", &funcs)

		// Set a new field on all objects
		result := arr.Set("newField", "newValue")
		assert.True(t, result.IsValid())

		// Check that the field was added to all objects
		assert.Equal(t, "newValue", arr.Index(0).Get("newField").String())
		assert.Equal(t, "newValue", arr.Index(1).Get("newField").String())
	})

	t.Run("SetOnArrayOfNonObjects", func(t *testing.T) {
		// Create array of non-objects
		arr := NewArrayNode([]Node{
			NewNumberNode(1, "", &funcs),
			NewNumberNode(2, "", &funcs),
		}, "", &funcs)

		// Try to set a field - should fail
		result := arr.Set("newField", "newValue")
		assert.False(t, result.IsValid())
		assert.Equal(t, ErrTypeAssertion, result.Error())
	})

	t.Run("SetWithErrorPropagation", func(t *testing.T) {
		// Create array with an invalid node
		invalid := NewInvalidNode("", ErrNotFound)
		arr := NewArrayNode([]Node{
			NewObjectNode(map[string]Node{"a": NewNumberNode(1, "", &funcs)}, "", &funcs),
			invalid,
		}, "", &funcs)

		// Try to set a field - should propagate the invalid node, not produce a type error
		result := arr.Set("newField", "newValue")
		assert.False(t, result.IsValid())
		// The error should be the one from the invalid node
		assert.Equal(t, ErrNotFound, result.Error())
	})
}

func TestArrayNodeStringsMethod(t *testing.T) {
	funcs := make(map[string]func(Node) Node)

	t.Run("StringsOnStringArray", func(t *testing.T) {
		arr := NewArrayNode([]Node{
			NewStringNode("first", "", &funcs),
			NewStringNode("second", "", &funcs),
			NewStringNode("third", "", &funcs),
		}, "", &funcs)

		strings := arr.Strings()
		assert.Equal(t, []string{"first", "second", "third"}, strings)
	})

	t.Run("StringsOnMixedArray", func(t *testing.T) {
		arr := NewArrayNode([]Node{
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
	funcs := make(map[string]func(Node) Node)

	t.Run("ContainsInStringArray", func(t *testing.T) {
		arr := NewArrayNode([]Node{
			NewStringNode("first", "", &funcs),
			NewStringNode("second", "", &funcs),
			NewStringNode("third", "", &funcs),
		}, "", &funcs)

		assert.True(t, arr.Contains("second"))
		assert.False(t, arr.Contains("fourth"))
	})

	t.Run("ContainsInNonStringArray", func(t *testing.T) {
		arr := NewArrayNode([]Node{
			NewNumberNode(1, "", &funcs),
			NewNumberNode(2, "", &funcs),
		}, "", &funcs)

		assert.False(t, arr.Contains("1"))
	})
}
