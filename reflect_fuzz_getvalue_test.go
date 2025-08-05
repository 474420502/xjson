package xjson

import (
	"reflect"
	"testing"
)

// This test uses reflection to directly call getValue with types and values
// that normal JSON and Query API cannot produce, to hit unreachable branches.
func TestReflectFuzzGetValueBranches(t *testing.T) {
	doc := &Document{}

	// 1. Pass a struct as data (not map, not array)
	t.Run("struct_as_data", func(t *testing.T) {
		type dummy struct{ A int }
		v := dummy{A: 1}
		res := doc.getValue(v, "A")
		if res != nil {
			t.Errorf("Expected nil for struct input, got %v", res)
		}
	})

	// 2. Pass a channel as data
	t.Run("chan_as_data", func(t *testing.T) {
		ch := make(chan int)
		res := doc.getValue(ch, "0")
		if res != nil {
			t.Errorf("Expected nil for chan input, got %v", res)
		}
	})

	// 3. Pass a function as data
	t.Run("func_as_data", func(t *testing.T) {
		f := func() {}
		res := doc.getValue(f, "0")
		if res != nil {
			t.Errorf("Expected nil for func input, got %v", res)
		}
	})

	// 4. Pass a pointer to a map
	t.Run("ptr_to_map", func(t *testing.T) {
		m := map[string]interface{}{"x": 1}
		res := doc.getValue(&m, "x")
		if res != nil {
			t.Errorf("Expected nil for pointer to map, got %v", res)
		}
	})

	// 5. Pass a pointer to an array
	t.Run("ptr_to_array", func(t *testing.T) {
		a := []interface{}{1, 2, 3}
		res := doc.getValue(&a, "[0]")
		if res != nil {
			t.Errorf("Expected nil for pointer to array, got %v", res)
		}
	})

	// 6. Pass nil as data
	t.Run("nil_data", func(t *testing.T) {
		res := doc.getValue(nil, "any")
		if res != nil {
			t.Errorf("Expected nil for nil input, got %v", res)
		}
	})

	// 7. Pass a slice of int (not []interface{})
	t.Run("slice_of_int", func(t *testing.T) {
		s := []int{1, 2, 3}
		res := doc.getValue(s, "[0]")
		if res != nil {
			t.Errorf("Expected nil for []int input, got %v", res)
		}
	})

	// 8. Pass a map with non-string keys
	t.Run("map_with_nonstring_keys", func(t *testing.T) {
		m := map[int]interface{}{1: "a"}
		res := doc.getValue(m, "1")
		if res != nil {
			t.Errorf("Expected nil for map[int] input, got %v", res)
		}
	})

	// 9. Pass a deeply nested invalid type
	t.Run("deeply_nested_invalid", func(t *testing.T) {
		m := map[string]interface{}{"x": []int{1, 2, 3}}
		res := doc.getValue(m, "x[0]")
		if res != nil {
			t.Errorf("Expected nil for nested []int, got %v", res)
		}
	})

	// 10. Pass a reflect.Value
	t.Run("reflect_value", func(t *testing.T) {
		rv := reflect.ValueOf(map[string]interface{}{"a": 1})
		res := doc.getValue(rv, "a")
		if res != nil {
			t.Errorf("Expected nil for reflect.Value input, got %v", res)
		}
	})
}
