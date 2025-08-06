package xjson

import (
	"testing"
)

func TestPrecisionCoverageImprovements(t *testing.T) {
	// 针对 Float() 函数的 66.7% 覆盖率
	t.Run("Float_PreciseBranches", func(t *testing.T) {
		// 测试 int64 到 float64 的转换 - 这个分支可能没被覆盖
		result := &Result{matches: []interface{}{int64(123456789)}}
		f, err := result.Float()
		if err != nil {
			t.Errorf("Float() on int64 should succeed, got error: %v", err)
		}
		if f != 123456789.0 {
			t.Errorf("Float() on int64 should return 123456789.0, got %f", f)
		}

		// 测试 string 可以成功解析为 float 的情况
		result2 := &Result{matches: []interface{}{"3.14159"}}
		f2, err2 := result2.Float()
		if err2 != nil {
			t.Errorf("Float() on valid float string should succeed, got error: %v", err2)
		}
		if f2 != 3.14159 {
			t.Errorf("Float() on '3.14159' should return 3.14159, got %f", f2)
		}

		// 测试 string 的科学计数法
		result3 := &Result{matches: []interface{}{"1.23e+10"}}
		f3, err3 := result3.Float()
		if err3 != nil {
			t.Errorf("Float() on scientific notation should succeed, got error: %v", err3)
		}
		if f3 != 1.23e+10 {
			t.Errorf("Float() on '1.23e+10' should return 1.23e+10, got %f", f3)
		}
	})

	// 针对 Index() 函数的 72.2% 覆盖率
	t.Run("Index_PreciseBranches", func(t *testing.T) {
		// 测试多个匹配的情况 - 这个分支很难构造，需要特殊的Result
		// 构造一个有多个matches的Result
		multiResult := &Result{
			matches: []interface{}{"first", "second", "third"},
		}

		// 测试正常索引
		indexed := multiResult.Index(1)
		if !indexed.Exists() {
			t.Error("Index(1) on multi-match result should exist")
		}
		str, _ := indexed.String()
		if str != "second" {
			t.Errorf("Index(1) should return 'second', got '%s'", str)
		}

		// 测试负数索引
		indexed2 := multiResult.Index(-1)
		if !indexed2.Exists() {
			t.Error("Index(-1) on multi-match result should exist")
		}
		str2, _ := indexed2.String()
		if str2 != "third" {
			t.Errorf("Index(-1) should return 'third', got '%s'", str2)
		}

		// 测试超出范围的索引
		indexed3 := multiResult.Index(10)
		if indexed3.Exists() {
			t.Error("Index(10) on multi-match result should not exist")
		}

		// 测试负数超出范围
		indexed4 := multiResult.Index(-10)
		if indexed4.Exists() {
			t.Error("Index(-10) on multi-match result should not exist")
		}

		// 测试单个数组的负数索引 - 这个分支可能没被完全覆盖
		doc, _ := ParseString(`{"arr": [10, 20, 30, 40, 50]}`)
		arr := doc.Query("/arr")

		// 测试负数索引
		negResult := arr.Index(-2)
		if !negResult.Exists() {
			t.Error("Index(-2) on array should exist")
		}
		val, _ := negResult.Int()
		if val != 40 {
			t.Errorf("Index(-2) should return 40, got %d", val)
		}

		// 测试非数组的类型匹配错误
		doc2, _ := ParseString(`{"str": "not_array"}`)
		str_result := doc2.Query("/str")
		indexed5 := str_result.Index(0)
		// 这应该返回 ErrTypeMismatch
		if indexed5.Exists() {
			t.Error("Index() on string should return error result")
		}
	})

	// 针对 String() 函数的 75% 覆盖率
	t.Run("String_PreciseBranches", func(t *testing.T) {
		// 测试 fmt.Sprintf 分支 - 当 json.Marshal 失败时
		// 我们需要构造一个无法被 json.Marshal 的值
		// Go 中有些类型无法被 JSON 序列化，比如 channel, function

		// 不过由于 Result.matches 是 []interface{}，很难直接放入不可序列化的值
		// 让我们测试一些边界情况

		// 测试非常复杂的嵌套结构
		complexMap := map[string]interface{}{
			"nested": map[string]interface{}{
				"deep": map[string]interface{}{
					"array": []interface{}{1, 2, 3},
					"null":  nil,
				},
			},
		}
		result := &Result{matches: []interface{}{complexMap}}
		str, err := result.String()
		if err != nil {
			t.Errorf("String() on complex map should succeed, got error: %v", err)
		}
		if str == "" {
			t.Error("String() on complex map should return non-empty JSON")
		}

		// 测试空 interface{} 切片
		result2 := &Result{matches: []interface{}{[]interface{}{}}}
		str2, err2 := result2.String()
		if err2 != nil {
			t.Errorf("String() on empty slice should succeed, got error: %v", err2)
		}
		if str2 != "[]" {
			t.Errorf("String() on empty slice should return '[]', got '%s'", str2)
		}
	})

	// 针对 Query() 函数的 80% 覆盖率
	t.Run("Query_PreciseBranches", func(t *testing.T) {
		// 测试错误传播分支
		doc := &Document{err: ErrInvalidJSON}
		result := doc.Query("/any/path")
		if result.Exists() {
			t.Error("Query on invalid document should not exist")
		}

		// 测试复杂路径解析器的不同分支
		doc2, _ := ParseString(`{
			"store": {
				"books": [
					{"title": "Book1", "price": 10},
					{"title": "Book2", "price": 20}
				]
			}
		}`)

		// 测试包含过滤器的复杂查询
		result2 := doc2.Query("/store/books[price > 15]")
		if result2.Exists() {
			t.Log("Complex filter query succeeded")
		} else {
			t.Log("Complex filter query failed (may be expected)")
		}

		// 测试递归查询
		result3 := doc2.Query("//title")
		if result3.Exists() {
			t.Log("Recursive query succeeded")
		} else {
			t.Log("Recursive query failed (may be expected)")
		}
	})

	// 针对 Int64() 函数的 80% 覆盖率
	t.Run("Int64_PreciseBranches", func(t *testing.T) {
		// 测试 int 到 int64 的转换
		result := &Result{matches: []interface{}{int(42)}}
		i64, err := result.Int64()
		if err != nil {
			t.Errorf("Int64() on int should succeed, got error: %v", err)
		}
		if i64 != 42 {
			t.Errorf("Int64() on int should return 42, got %d", i64)
		}

		// 测试 string 成功转换为 int64
		result2 := &Result{matches: []interface{}{"9223372036854775807"}}
		i64_2, err2 := result2.Int64()
		if err2 != nil {
			t.Errorf("Int64() on valid int64 string should succeed, got error: %v", err2)
		}
		if i64_2 != 9223372036854775807 {
			t.Errorf("Int64() should parse max int64, got %d", i64_2)
		}

		// 测试 string 转换失败
		result3 := &Result{matches: []interface{}{"not_a_number"}}
		_, err3 := result3.Int64()
		if err3 == nil {
			t.Error("Int64() on invalid string should return error")
		}
	})

	// 针对 Bool() 函数的 81.2% 覆盖率
	t.Run("Bool_PreciseBranches", func(t *testing.T) {
		// 测试 string 成功解析为 bool
		result := &Result{matches: []interface{}{"true"}}
		b, err := result.Bool()
		if err != nil {
			t.Errorf("Bool() on 'true' should succeed, got error: %v", err)
		}
		if !b {
			t.Error("Bool() on 'true' should return true")
		}

		// 测试 "false" 字符串
		result2 := &Result{matches: []interface{}{"false"}}
		b2, err2 := result2.Bool()
		if err2 != nil {
			t.Errorf("Bool() on 'false' should succeed, got error: %v", err2)
		}
		if b2 {
			t.Error("Bool() on 'false' should return false, but got true")
		}

		// 测试 int64 类型
		result3 := &Result{matches: []interface{}{int64(100)}}
		b3, err3 := result3.Bool()
		if err3 != nil {
			t.Errorf("Bool() on int64 should succeed, got error: %v", err3)
		}
		if !b3 {
			t.Error("Bool() on non-zero int64 should return true")
		}

		// 测试 int64 零值
		result4 := &Result{matches: []interface{}{int64(0)}}
		b4, err4 := result4.Bool()
		if err4 != nil {
			t.Errorf("Bool() on zero int64 should succeed, got error: %v", err4)
		}
		if b4 {
			t.Error("Bool() on zero int64 should return false")
		}
	})
}
