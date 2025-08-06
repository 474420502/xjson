package xjson

import (
	"fmt"
	"testing"
)

func TestXPathEdgeCasesAndLimitations(t *testing.T) {
	// 专门测试边界情况和已知限制
	edgeCasesData := `{
		"edge_cases": {
			"numbers": {
				"zero": 0,
				"negative": -42,
				"float_zero": 0.0,
				"scientific": 1.23e-4,
				"large_int": 2147483647,
				"small_float": 0.000000001
			},
			"strings": {
				"empty": "",
				"whitespace": "   ",
				"newlines": "line1\nline2\nline3",
				"quotes": "He said \"Hello\"",
				"backslashes": "C:\\\\Users\\\\test",
				"unicode_escape": "\\u4e2d\\u6587"
			},
			"special_arrays": {
				"single_element": [42],
				"all_nulls": [null, null, null],
				"mixed_nulls": [1, null, "text", null, true],
				"nested_empties": [[], {}, "", null]
			},
			"complex_nesting": {
				"level1": {
					"level2": [
						{
							"level3": {
								"level4": [
									{
										"level5": {
											"final_value": "deep_nested_success"
										}
									}
								]
							}
						}
					]
				}
			}
		},
		"performance_test": {
			"large_array": [
				1, 2, 3, 4, 5, 6, 7, 8, 9, 10,
				11, 12, 13, 14, 15, 16, 17, 18, 19, 20,
				21, 22, 23, 24, 25, 26, 27, 28, 29, 30,
				31, 32, 33, 34, 35, 36, 37, 38, 39, 40,
				41, 42, 43, 44, 45, 46, 47, 48, 49, 50
			],
			"wide_object": {
				"field1": "value1", "field2": "value2", "field3": "value3", "field4": "value4", "field5": "value5",
				"field6": "value6", "field7": "value7", "field8": "value8", "field9": "value9", "field10": "value10",
				"field11": "value11", "field12": "value12", "field13": "value13", "field14": "value14", "field15": "value15",
				"field16": "value16", "field17": "value17", "field18": "value18", "field19": "value19", "field20": "value20"
			}
		}
	}`

	doc, err := ParseString(edgeCasesData)
	if err != nil {
		t.Fatalf("ParseString failed: %v", err)
	}

	t.Run("数字边界情况", func(t *testing.T) {
		// 零值
		zero, err := doc.Query("/edge_cases/numbers/zero").Int()
		if err != nil {
			t.Errorf("Query zero failed: %v", err)
		}
		if zero != 0 {
			t.Errorf("Expected 0, got %d", zero)
		}

		// 负数
		negative, err := doc.Query("/edge_cases/numbers/negative").Int()
		if err != nil {
			t.Errorf("Query negative failed: %v", err)
		}
		if negative != -42 {
			t.Errorf("Expected -42, got %d", negative)
		}

		// 浮点零
		floatZero, err := doc.Query("/edge_cases/numbers/float_zero").Float()
		if err != nil {
			t.Errorf("Query float zero failed: %v", err)
		}
		if floatZero != 0.0 {
			t.Errorf("Expected 0.0, got %f", floatZero)
		}

		// 科学记数法
		scientific, err := doc.Query("/edge_cases/numbers/scientific").Float()
		if err != nil {
			t.Errorf("Query scientific notation failed: %v", err)
		}
		if scientific != 1.23e-4 {
			t.Errorf("Expected 1.23e-4, got %e", scientific)
		}

		// 大整数
		largeInt, err := doc.Query("/edge_cases/numbers/large_int").Int()
		if err != nil {
			t.Errorf("Query large int failed: %v", err)
		}
		if largeInt != 2147483647 {
			t.Errorf("Expected 2147483647, got %d", largeInt)
		}

		// 小浮点数
		smallFloat, err := doc.Query("/edge_cases/numbers/small_float").Float()
		if err != nil {
			t.Errorf("Query small float failed: %v", err)
		}
		if smallFloat != 0.000000001 {
			t.Errorf("Expected 0.000000001, got %e", smallFloat)
		}
	})

	t.Run("字符串边界情况", func(t *testing.T) {
		// 空字符串
		empty, err := doc.Query("/edge_cases/strings/empty").String()
		if err != nil {
			t.Errorf("Query empty string failed: %v", err)
		}
		if empty != "" {
			t.Errorf("Expected empty string, got '%s'", empty)
		}

		// 空白字符
		whitespace, err := doc.Query("/edge_cases/strings/whitespace").String()
		if err != nil {
			t.Errorf("Query whitespace failed: %v", err)
		}
		if whitespace != "   " {
			t.Errorf("Expected '   ', got '%s'", whitespace)
		}

		// 换行符
		newlines, err := doc.Query("/edge_cases/strings/newlines").String()
		if err != nil {
			t.Errorf("Query newlines failed: %v", err)
		}
		if newlines != "line1\nline2\nline3" {
			t.Errorf("Expected 'line1\\nline2\\nline3', got '%s'", newlines)
		}

		// 引号
		quotes, err := doc.Query("/edge_cases/strings/quotes").String()
		if err != nil {
			t.Errorf("Query quotes failed: %v", err)
		}
		if quotes != "He said \"Hello\"" {
			t.Errorf("Expected 'He said \"Hello\"', got '%s'", quotes)
		}

		// 反斜杠
		backslashes, err := doc.Query("/edge_cases/strings/backslashes").String()
		if err != nil {
			t.Errorf("Query backslashes failed: %v", err)
		}
		if backslashes != "C:\\\\Users\\\\test" {
			t.Errorf("Expected 'C:\\\\\\\\Users\\\\\\\\test', got '%s'", backslashes)
		}
	})

	t.Run("特殊数组情况", func(t *testing.T) {
		// 单元素数组
		singleElement := doc.Query("/edge_cases/special_arrays/single_element")
		if !singleElement.IsArray() {
			t.Error("single_element should be array")
		}
		if singleElement.Count() != 1 {
			t.Errorf("Expected count 1, got %d", singleElement.Count())
		}
		value, err := doc.Query("/edge_cases/special_arrays/single_element[0]").Int()
		if err != nil {
			t.Errorf("Query single element failed: %v", err)
		}
		if value != 42 {
			t.Errorf("Expected 42, got %d", value)
		}

		// 全 null 数组
		allNulls := doc.Query("/edge_cases/special_arrays/all_nulls")
		if !allNulls.IsArray() {
			t.Error("all_nulls should be array")
		}
		if allNulls.Count() != 3 {
			t.Errorf("Expected count 3, got %d", allNulls.Count())
		}

		// 检查每个元素都是 null
		for i := 0; i < 3; i++ {
			nullElement := doc.Query(fmt.Sprintf("/edge_cases/special_arrays/all_nulls[%d]", i))
			if !nullElement.IsNull() {
				t.Errorf("Element at index %d should be null", i)
			}
		}

		// 混合 null 数组
		mixedNulls := doc.Query("/edge_cases/special_arrays/mixed_nulls")
		if mixedNulls.Count() != 5 {
			t.Errorf("Expected count 5, got %d", mixedNulls.Count())
		}

		// 检查特定元素
		firstInt, err := doc.Query("/edge_cases/special_arrays/mixed_nulls[0]").Int()
		if err != nil || firstInt != 1 {
			t.Errorf("Expected first element to be 1, got %d (err: %v)", firstInt, err)
		}

		secondNull := doc.Query("/edge_cases/special_arrays/mixed_nulls[1]")
		if !secondNull.IsNull() {
			t.Error("Second element should be null")
		}

		thirdString, err := doc.Query("/edge_cases/special_arrays/mixed_nulls[2]").String()
		if err != nil || thirdString != "text" {
			t.Errorf("Expected third element to be 'text', got '%s' (err: %v)", thirdString, err)
		}
	})

	t.Run("复杂嵌套边界", func(t *testing.T) {
		// 深层嵌套访问
		deepValue, err := doc.Query("/edge_cases/complex_nesting/level1/level2[0]/level3/level4[0]/level5/final_value").String()
		if err != nil {
			t.Errorf("Query deep nested value failed: %v", err)
		}
		if deepValue != "deep_nested_success" {
			t.Errorf("Expected 'deep_nested_success', got '%s'", deepValue)
		}

		// 验证中间层级
		level2 := doc.Query("/edge_cases/complex_nesting/level1/level2")
		if !level2.IsArray() {
			t.Error("level2 should be array")
		}

		level3 := doc.Query("/edge_cases/complex_nesting/level1/level2[0]/level3")
		if !level3.IsObject() {
			t.Error("level3 should be object")
		}

		level4 := doc.Query("/edge_cases/complex_nesting/level1/level2[0]/level3/level4")
		if !level4.IsArray() {
			t.Error("level4 should be array")
		}

		level5 := doc.Query("/edge_cases/complex_nesting/level1/level2[0]/level3/level4[0]/level5")
		if !level5.IsObject() {
			t.Error("level5 should be object")
		}
	})

	t.Run("性能相关测试", func(t *testing.T) {
		// 大数组访问
		largeArray := doc.Query("/performance_test/large_array")
		if !largeArray.IsArray() {
			t.Error("large_array should be array")
		}
		expectedCount := 50
		if largeArray.Count() != expectedCount {
			t.Errorf("Expected %d elements, got %d", expectedCount, largeArray.Count())
		}

		// 访问大数组的第一个和最后一个元素
		firstElement, err := doc.Query("/performance_test/large_array[0]").Int()
		if err != nil {
			t.Errorf("Query first element failed: %v", err)
		}
		if firstElement != 1 {
			t.Errorf("Expected 1, got %d", firstElement)
		}

		lastElement, err := doc.Query("/performance_test/large_array[-1]").Int()
		if err != nil {
			t.Errorf("Query last element failed: %v", err)
		}
		if lastElement != 50 {
			t.Errorf("Expected 50, got %d", lastElement)
		}

		// 中间位置访问
		middleElement, err := doc.Query("/performance_test/large_array[25]").Int()
		if err != nil {
			t.Errorf("Query middle element failed: %v", err)
		}
		if middleElement != 26 {
			t.Errorf("Expected 26, got %d", middleElement)
		}

		// 宽对象访问
		wideObject := doc.Query("/performance_test/wide_object")
		if !wideObject.IsObject() {
			t.Error("wide_object should be object")
		}

		// 访问宽对象的一些字段
		field1, err := doc.Query("/performance_test/wide_object/field1").String()
		if err != nil {
			t.Errorf("Query field1 failed: %v", err)
		}
		if field1 != "value1" {
			t.Errorf("Expected 'value1', got '%s'", field1)
		}

		field10, err := doc.Query("/performance_test/wide_object/field10").String()
		if err != nil {
			t.Errorf("Query field10 failed: %v", err)
		}
		if field10 != "value10" {
			t.Errorf("Expected 'value10', got '%s'", field10)
		}

		field20, err := doc.Query("/performance_test/wide_object/field20").String()
		if err != nil {
			t.Errorf("Query field20 failed: %v", err)
		}
		if field20 != "value20" {
			t.Errorf("Expected 'value20', got '%s'", field20)
		}
	})
}

func TestXPathErrorHandlingPatterns(t *testing.T) {
	// 测试各种错误处理模式
	errorTestData := `{
		"valid": {
			"string": "test",
			"number": 123,
			"boolean": true,
			"array": [1, 2, 3],
			"object": {"key": "value"}
		}
	}`

	doc, err := ParseString(errorTestData)
	if err != nil {
		t.Fatalf("ParseString failed: %v", err)
	}

	t.Run("不存在路径的错误处理", func(t *testing.T) {
		// 完全不存在的顶级路径
		nonExistent := doc.Query("/completely/non/existent/path")
		if nonExistent.Exists() {
			t.Error("Non-existent path should not exist")
		}

		// 尝试从不存在的路径获取值应该返回错误
		_, err := nonExistent.String()
		if err == nil {
			t.Error("Expected error when getting string from non-existent path")
		}

		_, err = nonExistent.Int()
		if err == nil {
			t.Error("Expected error when getting int from non-existent path")
		}

		_, err = nonExistent.Bool()
		if err == nil {
			t.Error("Expected error when getting bool from non-existent path")
		}

		_, err = nonExistent.Float()
		if err == nil {
			t.Error("Expected error when getting float from non-existent path")
		}
	})

	t.Run("类型不匹配的错误处理", func(t *testing.T) {
		// 尝试将字符串转换为数字
		stringValue := doc.Query("/valid/string")
		if !stringValue.Exists() {
			t.Fatal("String value should exist")
		}

		_, err := stringValue.Int()
		if err == nil {
			t.Error("Expected error when converting string to int")
		}

		_, err = stringValue.Float()
		if err == nil {
			t.Error("Expected error when converting string to float")
		}

		_, err = stringValue.Bool()
		if err == nil {
			t.Logf("String to bool conversion succeeded (implementation allows this)")
		} else {
			t.Logf("String to bool conversion failed as expected: %v", err)
		}

		// 尝试将数字转换为布尔值（这可能工作或不工作，取决于实现）
		numberValue := doc.Query("/valid/number")
		_, err = numberValue.Bool()
		if err != nil {
			t.Logf("Number to bool conversion failed as expected: %v", err)
		} else {
			t.Logf("Number to bool conversion succeeded (implementation allows this)")
		}
	})

	t.Run("数组越界错误处理", func(t *testing.T) {
		validArray := doc.Query("/valid/array")
		if !validArray.IsArray() {
			t.Fatal("Array should exist and be array")
		}

		// 正向越界
		outOfBounds := doc.Query("/valid/array[10]")
		if outOfBounds.Exists() {
			t.Error("Out of bounds positive index should not exist")
		}

		// 负向越界
		negativeOutOfBounds := doc.Query("/valid/array[-10]")
		if negativeOutOfBounds.Exists() {
			t.Error("Out of bounds negative index should not exist")
		}

		// 非数字索引（这个测试可能不适用，取决于解析器如何处理）
		// 我们跳过这个，因为路径解析器可能在解析阶段就拒绝非数字索引
	})

	t.Run("对象属性访问错误", func(t *testing.T) {
		// 在非对象上访问属性
		numberValue := doc.Query("/valid/number")
		propertyOnNumber := numberValue.Get("property")
		if propertyOnNumber.Exists() {
			t.Error("Property access on number should not exist")
		}

		arrayValue := doc.Query("/valid/array")
		propertyOnArray := arrayValue.Get("property")
		if propertyOnArray.Exists() {
			t.Error("Property access on array should not exist")
		}

		// 访问不存在的对象属性
		validObject := doc.Query("/valid/object")
		nonExistentProperty := validObject.Get("nonexistent")
		if nonExistentProperty.Exists() {
			t.Error("Non-existent property should not exist")
		}
	})

	t.Run("Must方法的行为", func(t *testing.T) {
		// Must 方法在错误时会 panic，所以我们测试有效值的 Must 行为
		validString := doc.Query("/valid/string")
		mustString := validString.MustString()
		if mustString != "test" {
			t.Errorf("MustString should return 'test', got '%s'", mustString)
		}

		validNumber := doc.Query("/valid/number")
		mustInt := validNumber.MustInt()
		if mustInt != 123 {
			t.Errorf("MustInt should return 123, got %d", mustInt)
		}

		mustFloat := validNumber.MustFloat()
		if mustFloat != 123.0 {
			t.Errorf("MustFloat should return 123.0, got %f", mustFloat)
		}

		validBool := doc.Query("/valid/boolean")
		mustBool := validBool.MustBool()
		if mustBool != true {
			t.Errorf("MustBool should return true, got %t", mustBool)
		}

		// 注意：Must 方法在遇到错误时会 panic，而不是返回零值
		// 这是设计上的选择，用于快速失败的场景
	})
}

func TestXPathIterationAndTraversalPatterns(t *testing.T) {
	// 测试迭代和遍历模式
	iterationData := `{
		"teams": [
			{
				"name": "Frontend",
				"members": [
					{"name": "Alice", "role": "Developer", "experience": 3},
					{"name": "Bob", "role": "Designer", "experience": 2}
				]
			},
			{
				"name": "Backend", 
				"members": [
					{"name": "Charlie", "role": "Developer", "experience": 5},
					{"name": "Diana", "role": "DevOps", "experience": 4}
				]
			}
		],
		"scores": [85, 90, 78, 92, 88],
		"categories": {
			"A": [1, 2, 3],
			"B": [4, 5, 6],
			"C": [7, 8, 9]
		}
	}`

	doc, err := ParseString(iterationData)
	if err != nil {
		t.Fatalf("ParseString failed: %v", err)
	}

	t.Run("ForEach遍历测试", func(t *testing.T) {
		// 遍历团队
		teams := doc.Query("/teams")
		teamCount := 0
		teams.ForEach(func(index int, value IResult) bool {
			teamCount++
			name, err := value.Get("name").String()
			if err != nil {
				t.Errorf("Failed to get team name at index %d: %v", index, err)
				return false
			}

			expectedNames := []string{"Frontend", "Backend"}
			if index >= len(expectedNames) {
				t.Errorf("Unexpected team at index %d", index)
				return false
			}

			if name != expectedNames[index] {
				t.Errorf("Expected team name '%s' at index %d, got '%s'", expectedNames[index], index, name)
			}

			return true // 继续迭代
		})

		if teamCount != 2 {
			t.Errorf("Expected to iterate 2 teams, got %d", teamCount)
		}

		// 遍历分数
		scores := doc.Query("/scores")
		scoreSum := 0
		scoreCount := 0
		scores.ForEach(func(index int, value IResult) bool {
			score := value.MustInt()
			scoreSum += score
			scoreCount++
			return true
		})

		expectedSum := 85 + 90 + 78 + 92 + 88
		if scoreSum != expectedSum {
			t.Errorf("Expected score sum %d, got %d", expectedSum, scoreSum)
		}
		if scoreCount != 5 {
			t.Errorf("Expected 5 scores, got %d", scoreCount)
		}
	})

	t.Run("嵌套迭代测试", func(t *testing.T) {
		// 遍历团队和成员
		teams := doc.Query("/teams")
		totalMembers := 0

		teams.ForEach(func(teamIndex int, team IResult) bool {
			members := team.Get("members")
			members.ForEach(func(memberIndex int, member IResult) bool {
				totalMembers++

				name := member.Get("name").MustString()
				role := member.Get("role").MustString()
				experience := member.Get("experience").MustInt()

				// 验证数据完整性
				if name == "" {
					t.Errorf("Empty name for member at team %d, member %d", teamIndex, memberIndex)
				}
				if role == "" {
					t.Errorf("Empty role for member at team %d, member %d", teamIndex, memberIndex)
				}
				if experience <= 0 {
					t.Errorf("Invalid experience %d for member at team %d, member %d", experience, teamIndex, memberIndex)
				}

				return true
			})
			return true
		})

		if totalMembers != 4 {
			t.Errorf("Expected 4 total members, got %d", totalMembers)
		}
	})

	t.Run("中断迭代测试", func(t *testing.T) {
		// 测试提前中断迭代
		scores := doc.Query("/scores")
		processedCount := 0

		scores.ForEach(func(index int, value IResult) bool {
			processedCount++
			score := value.MustInt()

			// 当找到分数90时停止
			if score == 90 {
				return false // 中断迭代
			}

			return true
		})

		// 应该处理了前两个元素：85, 90
		if processedCount != 2 {
			t.Errorf("Expected to process 2 elements before stopping, got %d", processedCount)
		}
	})

	t.Run("Count和索引访问对比", func(t *testing.T) {
		// 验证 Count() 和实际索引访问的一致性
		scores := doc.Query("/scores")
		count := scores.Count()

		// 通过索引访问验证计数
		accessibleCount := 0
		for i := 0; i < count+2; i++ { // 多试几个索引
			element := doc.Query(fmt.Sprintf("/scores[%d]", i))
			if element.Exists() {
				accessibleCount++
			}
		}

		if count != accessibleCount {
			t.Errorf("Count() returned %d but only %d elements are accessible by index", count, accessibleCount)
		}

		// 验证具体的值
		for i := 0; i < count; i++ {
			expectedValues := []int{85, 90, 78, 92, 88}
			actual := doc.Query(fmt.Sprintf("/scores[%d]", i)).MustInt()
			if actual != expectedValues[i] {
				t.Errorf("Element at index %d: expected %d, got %d", i, expectedValues[i], actual)
			}
		}
	})

	t.Run("对象键值遍历", func(t *testing.T) {
		// 测试对象的键值访问
		categories := doc.Query("/categories")
		if !categories.IsObject() {
			t.Fatal("categories should be an object")
		}

		// 获取所有键
		keys := categories.Keys()
		expectedKeys := []string{"A", "B", "C"}

		if len(keys) != len(expectedKeys) {
			t.Errorf("Expected %d keys, got %d", len(expectedKeys), len(keys))
		}

		// 验证键名（注意：键的顺序可能不保证）
		keyMap := make(map[string]bool)
		for _, key := range keys {
			keyMap[key] = true
		}

		for _, expectedKey := range expectedKeys {
			if !keyMap[expectedKey] {
				t.Errorf("Expected key '%s' not found", expectedKey)
			}
		}

		// 验证每个键对应的值
		for _, key := range keys {
			categoryArray := categories.Get(key)
			if !categoryArray.IsArray() {
				t.Errorf("Category '%s' should be an array", key)
				continue
			}

			if categoryArray.Count() != 3 {
				t.Errorf("Category '%s' should have 3 elements, got %d", key, categoryArray.Count())
			}
		}
	})
}
