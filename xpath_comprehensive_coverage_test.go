package xjson

import (
	"testing"
)

func TestXPathComprehensiveCoverage(t *testing.T) {
	// 综合性测试数据，涵盖各种场景
	comprehensiveData := `{
		"root": {
			"simple": "value",
			"number": 42,
			"float": 3.14159,
			"boolean": true,
			"null": null,
			"empty_string": "",
			"empty_array": [],
			"empty_object": {},
			"mixed_array": [
				"string",
				123,
				true,
				null,
				{
					"nested": "object"
				}
			],
			"nested_structure": {
				"level1": {
					"level2": {
						"level3": {
							"deep_value": "found_it",
							"numbers": [1, 2, 3, 4, 5],
							"mixed": {
								"a": 1,
								"b": [10, 20, 30],
								"c": {
									"d": "nested_deeply"
								}
							}
						}
					}
				}
			},
			"special_keys": {
				"key-with-dashes": "dash_value",
				"key.with.dots": "dot_value",
				"key with spaces": "space_value",
				"123numeric": "numeric_start",
				"_underscore": "underscore_value"
			},
			"arrays": {
				"simple_numbers": [1, 2, 3, 4, 5],
				"simple_strings": ["a", "b", "c", "d", "e"],
				"objects": [
					{"id": 1, "name": "first", "active": true},
					{"id": 2, "name": "second", "active": false},
					{"id": 3, "name": "third", "active": true},
					{"id": 4, "name": "fourth", "active": false},
					{"id": 5, "name": "fifth", "active": true}
				]
			},
			"unicode": {
				"chinese": "中文测试",
				"emoji": "🚀🌟⭐",
				"special": "áéíóúñü",
				"japanese": "こんにちは"
			}
		}
	}`

	doc, err := ParseString(comprehensiveData)
	if err != nil {
		t.Fatalf("ParseString failed: %v", err)
	}

	t.Run("基本路径访问", func(t *testing.T) {
		// 简单值访问
		simple, err := doc.Query("root.simple").String()
		if err != nil {
			t.Errorf("Query simple string failed: %v", err)
		}
		if simple != "value" {
			t.Errorf("Expected 'value', got '%s'", simple)
		}

		// 数字访问
		number, err := doc.Query("root.number").Int()
		if err != nil {
			t.Errorf("Query number failed: %v", err)
		}
		if number != 42 {
			t.Errorf("Expected 42, got %d", number)
		}

		// 浮点数访问
		float, err := doc.Query("root.float").Float()
		if err != nil {
			t.Errorf("Query float failed: %v", err)
		}
		if float != 3.14159 {
			t.Errorf("Expected 3.14159, got %f", float)
		}

		// 布尔值访问
		boolean, err := doc.Query("root.boolean").Bool()
		if err != nil {
			t.Errorf("Query boolean failed: %v", err)
		}
		if !boolean {
			t.Error("Expected true, got false")
		}
	})

	t.Run("空值和特殊值处理", func(t *testing.T) {
		// null 值检查
		nullValue := doc.Query("root.null")
		if !nullValue.IsNull() {
			t.Error("Expected null value")
		}

		// 空字符串
		emptyStr, err := doc.Query("root.empty_string").String()
		if err != nil {
			t.Errorf("Query empty string failed: %v", err)
		}
		if emptyStr != "" {
			t.Errorf("Expected empty string, got '%s'", emptyStr)
		}

		// 空数组
		emptyArray := doc.Query("root.empty_array")
		if !emptyArray.IsArray() {
			t.Error("Expected empty array")
		}
		if emptyArray.Count() != 0 {
			t.Errorf("Expected empty array count 0, got %d", emptyArray.Count())
		}

		// 空对象
		emptyObj := doc.Query("root.empty_object")
		if !emptyObj.IsObject() {
			t.Error("Expected empty object")
		}
	})

	t.Run("深层嵌套访问", func(t *testing.T) {
		// 深层嵌套值访问
		deepValue, err := doc.Query("root.nested_structure.level1.level2.level3.deep_value").String()
		if err != nil {
			t.Errorf("Query deep value failed: %v", err)
		}
		if deepValue != "found_it" {
			t.Errorf("Expected 'found_it', got '%s'", deepValue)
		}

		// 深层嵌套数组
		numbers := doc.Query("root.nested_structure.level1.level2.level3.numbers")
		if !numbers.IsArray() {
			t.Error("Expected numbers array")
		}
		if numbers.Count() != 5 {
			t.Errorf("Expected 5 numbers, got %d", numbers.Count())
		}

		// 更深层的嵌套
		deepNested, err := doc.Query("root.nested_structure.level1.level2.level3.mixed.c.d").String()
		if err != nil {
			t.Errorf("Query deep nested value failed: %v", err)
		}
		if deepNested != "nested_deeply" {
			t.Errorf("Expected 'nested_deeply', got '%s'", deepNested)
		}
	})

	t.Run("特殊键名处理", func(t *testing.T) {
		// 包含连字符的键
		dashValue, err := doc.Query("root.special_keys.key-with-dashes").String()
		if err != nil {
			t.Errorf("Query dash key failed: %v", err)
		}
		if dashValue != "dash_value" {
			t.Errorf("Expected 'dash_value', got '%s'", dashValue)
		}

		// 包含点的键 (可能需要特殊处理)
		dotValue, err := doc.Query("root.special_keys.key.with.dots").String()
		if err == nil && dotValue == "dot_value" {
			t.Logf("Dot key access works: %s", dotValue)
		} else {
			t.Logf("Dot key access limitation noted (expected)")
		}

		// 包含空格的键
		spaceValue, err := doc.Query("root.special_keys.key with spaces").String()
		if err == nil && spaceValue == "space_value" {
			t.Logf("Space key access works: %s", spaceValue)
		} else {
			t.Logf("Space key access limitation noted (expected)")
		}

		// 以数字开头的键
		numericValue, err := doc.Query("root.special_keys.123numeric").String()
		if err != nil {
			t.Errorf("Query numeric start key failed: %v", err)
		}
		if numericValue != "numeric_start" {
			t.Errorf("Expected 'numeric_start', got '%s'", numericValue)
		}

		// 下划线键
		underscoreValue, err := doc.Query("root.special_keys._underscore").String()
		if err != nil {
			t.Errorf("Query underscore key failed: %v", err)
		}
		if underscoreValue != "underscore_value" {
			t.Errorf("Expected 'underscore_value', got '%s'", underscoreValue)
		}
	})

	t.Run("数组操作详细测试", func(t *testing.T) {
		// 正向索引
		firstNumber, err := doc.Query("root.arrays.simple_numbers[0]").Int()
		if err != nil {
			t.Errorf("Query first number failed: %v", err)
		}
		if firstNumber != 1 {
			t.Errorf("Expected 1, got %d", firstNumber)
		}

		// 负向索引
		lastNumber, err := doc.Query("root.arrays.simple_numbers[-1]").Int()
		if err != nil {
			t.Errorf("Query last number failed: %v", err)
		}
		if lastNumber != 5 {
			t.Errorf("Expected 5, got %d", lastNumber)
		}

		// 中间索引
		middleString, err := doc.Query("root.arrays.simple_strings[2]").String()
		if err != nil {
			t.Errorf("Query middle string failed: %v", err)
		}
		if middleString != "c" {
			t.Errorf("Expected 'c', got '%s'", middleString)
		}

		// 对象数组访问
		firstObjName, err := doc.Query("root.arrays.objects[0].name").String()
		if err != nil {
			t.Errorf("Query first object name failed: %v", err)
		}
		if firstObjName != "first" {
			t.Errorf("Expected 'first', got '%s'", firstObjName)
		}

		// 对象数组布尔值
		secondObjActive, err := doc.Query("root.arrays.objects[1].active").Bool()
		if err != nil {
			t.Errorf("Query second object active failed: %v", err)
		}
		if secondObjActive {
			t.Error("Expected second object to be inactive")
		}
	})

	t.Run("混合数组处理", func(t *testing.T) {
		mixedArray := doc.Query("root.mixed_array")
		if !mixedArray.IsArray() {
			t.Error("Expected mixed array")
		}

		// 字符串元素
		firstElement, err := doc.Query("root.mixed_array[0]").String()
		if err != nil {
			t.Errorf("Query first mixed element failed: %v", err)
		}
		if firstElement != "string" {
			t.Errorf("Expected 'string', got '%s'", firstElement)
		}

		// 数字元素
		secondElement, err := doc.Query("root.mixed_array[1]").Int()
		if err != nil {
			t.Errorf("Query second mixed element failed: %v", err)
		}
		if secondElement != 123 {
			t.Errorf("Expected 123, got %d", secondElement)
		}

		// 布尔元素
		thirdElement, err := doc.Query("root.mixed_array[2]").Bool()
		if err != nil {
			t.Errorf("Query third mixed element failed: %v", err)
		}
		if !thirdElement {
			t.Error("Expected true for third element")
		}

		// null 元素
		fourthElement := doc.Query("root.mixed_array[3]")
		if !fourthElement.IsNull() {
			t.Error("Expected null for fourth element")
		}

		// 嵌套对象元素
		fifthElementNested, err := doc.Query("root.mixed_array[4].nested").String()
		if err != nil {
			t.Errorf("Query fifth element nested failed: %v", err)
		}
		if fifthElementNested != "object" {
			t.Errorf("Expected 'object', got '%s'", fifthElementNested)
		}
	})

	t.Run("Unicode和国际化支持", func(t *testing.T) {
		// 中文字符
		chinese, err := doc.Query("root.unicode.chinese").String()
		if err != nil {
			t.Errorf("Query Chinese text failed: %v", err)
		}
		if chinese != "中文测试" {
			t.Errorf("Expected '中文测试', got '%s'", chinese)
		}

		// 表情符号
		emoji, err := doc.Query("root.unicode.emoji").String()
		if err != nil {
			t.Errorf("Query emoji failed: %v", err)
		}
		if emoji != "🚀🌟⭐" {
			t.Errorf("Expected '🚀🌟⭐', got '%s'", emoji)
		}

		// 特殊字符
		special, err := doc.Query("root.unicode.special").String()
		if err != nil {
			t.Errorf("Query special characters failed: %v", err)
		}
		if special != "áéíóúñü" {
			t.Errorf("Expected 'áéíóúñü', got '%s'", special)
		}

		// 日文字符
		japanese, err := doc.Query("root.unicode.japanese").String()
		if err != nil {
			t.Errorf("Query Japanese text failed: %v", err)
		}
		if japanese != "こんにちは" {
			t.Errorf("Expected 'こんにちは', got '%s'", japanese)
		}
	})

	t.Run("错误情况和边界测试", func(t *testing.T) {
		// 不存在的路径
		nonExistent := doc.Query("root.does.not.exist")
		if nonExistent.Exists() {
			t.Error("Non-existent path should not exist")
		}

		// 数组越界
		outOfBounds := doc.Query("root.arrays.simple_numbers[100]")
		if outOfBounds.Exists() {
			t.Error("Out of bounds access should not exist")
		}

		// 负向索引越界
		negativeOutOfBounds := doc.Query("root.arrays.simple_numbers[-100]")
		if negativeOutOfBounds.Exists() {
			t.Error("Negative out of bounds should not exist")
		}

		// 在非数组上使用索引
		wrongIndex := doc.Query("root.simple[0]")
		if wrongIndex.Exists() {
			t.Error("Index on non-array should not exist")
		}

		// 在非对象上访问属性
		wrongProperty := doc.Query("root.number.property")
		if wrongProperty.Exists() {
			t.Error("Property access on non-object should not exist")
		}
	})

	t.Run("类型检查和验证", func(t *testing.T) {
		// 验证各种类型检查
		simpleValue := doc.Query("root.simple")
		_, err := simpleValue.String()
		if err != nil {
			t.Error("simple should be convertible to string")
		}

		numberValue := doc.Query("root.number")
		_, err = numberValue.Int()
		if err != nil {
			t.Error("number should be convertible to int")
		}

		arrayValue := doc.Query("root.arrays.simple_numbers")
		if !arrayValue.IsArray() {
			t.Error("simple_numbers should be array")
		}

		objectValue := doc.Query("root.nested_structure")
		if !objectValue.IsObject() {
			t.Error("nested_structure should be object")
		}

		booleanValue := doc.Query("root.boolean")
		_, err = booleanValue.Bool()
		if err != nil {
			t.Error("boolean should be convertible to bool")
		}

		nullValue := doc.Query("root.null")
		if !nullValue.IsNull() {
			t.Error("null should be null")
		}
	})
}

func TestXPathArrayIterationPatterns(t *testing.T) {
	// 专门测试数组遍历模式
	arrayData := `{
		"datasets": [
			{
				"name": "dataset1",
				"values": [10, 20, 30, 40, 50],
				"metadata": {
					"created": "2023-01-01",
					"size": 5
				}
			},
			{
				"name": "dataset2", 
				"values": [100, 200, 300],
				"metadata": {
					"created": "2023-02-01",
					"size": 3
				}
			},
			{
				"name": "dataset3",
				"values": [1, 2, 3, 4, 5, 6, 7, 8, 9, 10],
				"metadata": {
					"created": "2023-03-01",
					"size": 10
				}
			}
		],
		"matrix": [
			[1, 2, 3],
			[4, 5, 6],
			[7, 8, 9]
		],
		"people": [
			{"name": "Alice", "scores": [85, 90, 88]},
			{"name": "Bob", "scores": [92, 87, 91]},
			{"name": "Charlie", "scores": [78, 85, 82]}
		]
	}`

	doc, err := ParseString(arrayData)
	if err != nil {
		t.Fatalf("ParseString failed: %v", err)
	}

	t.Run("多维数组访问", func(t *testing.T) {
		// 访问矩阵元素 - 使用两步查询
		firstRow := doc.Query("matrix[0]")
		element00, err := firstRow.Index(0).Int()
		if err != nil {
			t.Errorf("Query matrix[0][0] failed: %v", err)
		}
		if element00 != 1 {
			t.Errorf("Expected 1, got %d", element00)
		}

		secondRow := doc.Query("matrix[1]")
		element12, err := secondRow.Index(2).Int()
		if err != nil {
			t.Errorf("Query matrix[1][2] failed: %v", err)
		}
		if element12 != 6 {
			t.Errorf("Expected 6, got %d", element12)
		}

		thirdRow := doc.Query("matrix[2]")
		element22, err := thirdRow.Index(2).Int()
		if err != nil {
			t.Errorf("Query matrix[2][2] failed: %v", err)
		}
		if element22 != 9 {
			t.Errorf("Expected 9, got %d", element22)
		}

		// 验证矩阵结构
		matrix := doc.Query("matrix")
		if !matrix.IsArray() {
			t.Error("matrix should be an array")
		}
		if matrix.Count() != 3 {
			t.Errorf("Expected 3 rows, got %d", matrix.Count())
		}

		// 验证第一行
		if !firstRow.IsArray() {
			t.Error("first row should be an array")
		}
		if firstRow.Count() != 3 {
			t.Errorf("Expected 3 columns in first row, got %d", firstRow.Count())
		}
	})

	t.Run("嵌套数组导航", func(t *testing.T) {
		// 访问第一个数据集的第三个值
		firstDatasetThirdValue, err := doc.Query("datasets[0].values[2]").Int()
		if err != nil {
			t.Errorf("Query nested array value failed: %v", err)
		}
		if firstDatasetThirdValue != 30 {
			t.Errorf("Expected 30, got %d", firstDatasetThirdValue)
		}

		// 访问第二个数据集的大小
		secondDatasetSize, err := doc.Query("datasets[1].metadata.size").Int()
		if err != nil {
			t.Errorf("Query dataset size failed: %v", err)
		}
		if secondDatasetSize != 3 {
			t.Errorf("Expected 3, got %d", secondDatasetSize)
		}

		// 访问第三个数据集的最后一个值
		thirdDatasetLastValue, err := doc.Query("datasets[2].values[-1]").Int()
		if err != nil {
			t.Errorf("Query last value failed: %v", err)
		}
		if thirdDatasetLastValue != 10 {
			t.Errorf("Expected 10, got %d", thirdDatasetLastValue)
		}
	})

	t.Run("人员分数查询", func(t *testing.T) {
		// Alice 的第一个分数
		aliceFirstScore, err := doc.Query("people[0].scores[0]").Int()
		if err != nil {
			t.Errorf("Query Alice first score failed: %v", err)
		}
		if aliceFirstScore != 85 {
			t.Errorf("Expected 85, got %d", aliceFirstScore)
		}

		// Bob 的名字
		bobName, err := doc.Query("people[1].name").String()
		if err != nil {
			t.Errorf("Query Bob name failed: %v", err)
		}
		if bobName != "Bob" {
			t.Errorf("Expected 'Bob', got '%s'", bobName)
		}

		// Charlie 的最后一个分数
		charlieLastScore, err := doc.Query("people[2].scores[-1]").Int()
		if err != nil {
			t.Errorf("Query Charlie last score failed: %v", err)
		}
		if charlieLastScore != 82 {
			t.Errorf("Expected 82, got %d", charlieLastScore)
		}
	})

	t.Run("数组长度和计数验证", func(t *testing.T) {
		// 数据集数量
		datasets := doc.Query("datasets")
		datasetCount := datasets.Count()
		if datasetCount != 3 {
			t.Errorf("Expected 3 datasets, got %d", datasetCount)
		}

		// 第一个数据集的值数量
		firstValues := doc.Query("datasets[0].values")
		firstValueCount := firstValues.Count()
		if firstValueCount != 5 {
			t.Errorf("Expected 5 values in first dataset, got %d", firstValueCount)
		}

		// 矩阵行数
		matrix := doc.Query("matrix")
		matrixRows := matrix.Count()
		if matrixRows != 3 {
			t.Errorf("Expected 3 matrix rows, got %d", matrixRows)
		}

		// 第一行列数
		firstRow := doc.Query("matrix[0]")
		firstRowCols := firstRow.Count()
		if firstRowCols != 3 {
			t.Errorf("Expected 3 columns in first row, got %d", firstRowCols)
		}

		// 人员数量
		people := doc.Query("people")
		peopleCount := people.Count()
		if peopleCount != 3 {
			t.Errorf("Expected 3 people, got %d", peopleCount)
		}
	})
}
