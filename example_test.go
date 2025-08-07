package xjson_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/474420502/xjson"
)

func TestQuickStartExample(t *testing.T) {
	data := `{
		"store": {
			"books": [
				{"title": "Moby Dick", "price": 8.99, "tags": ["classic", "adventure"]},
				{"title": "Clean Code", "price": 29.99, "tags": ["programming"]}
			]
		}
	}`

	root, err := xjson.Parse(data)
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}

	root.Func("cheap", func(n xjson.Node) xjson.Node {
		return n.Filter(func(child xjson.Node) bool {
			price, ok := child.Get("price").RawFloat()
			return ok && price < 20
		})
	}).Func("tagged", func(n xjson.Node) xjson.Node {
		return n.Filter(func(child xjson.Node) bool {
			return child.Get("tags").Contains("adventure")
		})
	})

	cheapTitles := root.Query("/store/books[@cheap]/title").Strings()
	if err := root.Error(); err != nil {
		t.Errorf("查询失败: %v", err)
	}
	expectedCheapTitles := []string{"Moby Dick"}
	if !compareStringSlices(cheapTitles, expectedCheapTitles) {
		t.Errorf("期望的廉价书籍: %v, 实际: %v", expectedCheapTitles, cheapTitles)
	}

	root.Query("/store/books[@tagged]").Set("price", 9.99)
	if err := root.Error(); err != nil {
		t.Errorf("修改失败: %v", err)
	}

	// 验证修改后的价格
	modifiedPrice, ok := root.Query("/store/books[0]/price").RawFloat()
	if !ok || modifiedPrice != 9.99 {
		t.Errorf("期望修改后的价格为 9.99, 实际为 %v (ok: %v)", modifiedPrice, ok)
	}

	// 验证整个JSON字符串输出
	expectedOutput := `{"store":{"books":[{"title":"Moby Dick","price":9.99,"tags":["classic","adventure"]},{"title":"Clean Code","price":29.99,"tags":["programming"]}]}}`
	actualJSON := root.String()
	if err := compareJSON(actualJSON, expectedOutput); err != nil {
		t.Errorf("期望的JSON输出不匹配: %v\n期望: %s\n实际: %s", err, expectedOutput, actualJSON)
	}
}

func TestBusinessRuleEncapsulation(t *testing.T) {
	data := `{
		"products": [
			{"id": "p1", "name": "Laptop", "stock": 10, "status": "active"},
			{"id": "p2", "name": "Mouse", "stock": 0, "status": "active"},
			{"id": "p3", "name": "Keyboard", "stock": 5, "status": "inactive"},
			{"id": "p4", "name": "Monitor", "stock": 20, "status": "active"}
		]
	}`

	root, err := xjson.Parse(data)
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}

	root.Func("inStock", func(n xjson.Node) xjson.Node {
		return n.Filter(func(p xjson.Node) bool {
			return p.Get("stock").Int() > 0 &&
				p.Get("status").String() == "active"
		})
	})

	availableProducts := root.Query("/products[@inStock]/id").Strings()
	if err := root.Error(); err != nil {
		t.Errorf("查询失败: %v", err)
	}

	expectedAvailableProducts := []string{"p1", "p4"}
	if !compareStringSlices(availableProducts, expectedAvailableProducts) {
		t.Errorf("期望的可用产品: %v, 实际: %v", expectedAvailableProducts, availableProducts)
	}
}

func TestDataTransformationPipeline(t *testing.T) {
	data := `{
		"rawInput": [
			{"id": "item1", "name": "  Product A  ", "price": 10.123},
			{"id": "item2", "name": "Product B", "price": 20.456}
		]
	}`

	root, err := xjson.Parse(data)
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}

	root.Func("sanitize", func(n xjson.Node) xjson.Node {
		return n.Map(func(item xjson.Node) interface{} {
			return map[string]interface{}{
				"id":    item.Get("id").String(),
				"name":  strings.TrimSpace(item.Get("name").String()),
				"price": math.Round(item.Get("price").Float()*100) / 100,
			}
		})
	})

	cleanData := root.Query("/rawInput[@sanitize]")
	if err := root.Error(); err != nil {
		t.Errorf("数据清洗失败: %v", err)
	}

	// 验证清洗后的数据
	if cleanData.Index(0).Get("id").String() != "item1" {
		t.Errorf("期望 id 为 item1, 实际为 %s", cleanData.Index(0).Get("id").String())
	}
	if cleanData.Index(0).Get("name").String() != "Product A" {
		t.Errorf("期望 name 为 'Product A', 实际为 '%s'", cleanData.Index(0).Get("name").String())
	}
	if price, ok := cleanData.Index(0).Get("price").RawFloat(); !ok || price != 10.12 {
		t.Errorf("期望 price 为 10.12, 实际为 %v", price)
	}

	if cleanData.Index(1).Get("id").String() != "item2" {
		t.Errorf("期望 id 为 item2, 实际为 %s", cleanData.Index(1).Get("id").String())
	}
	if cleanData.Index(1).Get("name").String() != "Product B" {
		t.Errorf("期望 name 为 'Product B', 实际为 '%s'", cleanData.Index(1).Get("name").String())
	}
	if price, ok := cleanData.Index(1).Get("price").RawFloat(); !ok || price != 20.46 {
		t.Errorf("期望 price 为 20.46, 实际为 %v", price)
	}
}

func TestComprehensiveErrorHandling(t *testing.T) {
	data := `{"a":{"b":[1,2,3]}}`

	// 1. 路径不存在
	t.Run("PathNotFound", func(t *testing.T) {
		root, _ := xjson.Parse(data)
		node := root.Query("/a/c")
		if node.IsValid() {
			t.Error("期望节点无效，但它有效")
		}
	})

	// 2. 索引越界
	t.Run("IndexOutOfBounds", func(t *testing.T) {
		root, _ := xjson.Parse(data)
		node := root.Query("/a/b[5]")
		if node.IsValid() {
			t.Error("期望节点无效，但它有效")
		}
	})

	// 3. 在非对象上执行 Get
	t.Run("GetOnNonObject", func(t *testing.T) {
		root, _ := xjson.Parse(data)
		node := root.Query("/a/b").Get("key")
		if node.IsValid() {
			t.Error("期望节点无效，但它有效")
		}
	})

	// 4. 调用未注册的函数
	t.Run("UnregisteredFunction", func(t *testing.T) {
		root, _ := xjson.Parse(data)
		node := root.Query("/a/b[@nonexistent]")
		if node.IsValid() {
			t.Error("期望节点无效，但它有效")
		}
	})

}

func TestAdvancedDataManipulation(t *testing.T) {
	data := `{
		"users": [
			{"id": 1, "name": "Alice", "scores": [80, 90, 95]},
			{"id": 2, "name": "Bob", "scores": [70, 85, 88]}
		],
		"metadata": {}
	}`
	root, err := xjson.Parse(data)
	if err != nil {
		t.Fatal(err)
	}

	// 1. 使用 Map 计算平均分并转换结构
	root.Func("withAvg", func(n xjson.Node) xjson.Node {
		return n.Map(func(user xjson.Node) interface{} {
			scoresNode := user.Get("scores")
			var sum int64 = 0
			scoresNode.ForEach(func(_ interface{}, score xjson.Node) {
				sum += score.Int()
			})
			avg := float64(sum) / float64(scoresNode.Len())
			return map[string]interface{}{
				"name":     user.Get("name").String(),
				"avgScore": math.Round(avg*10) / 10,
			}
		})
	})

	processedUsers := root.Query("/users[@withAvg]")
	if err := root.Error(); err != nil {
		t.Errorf("数据处理失败: %v", err)
	}
	expectedJSON := `[{"name":"Alice","avgScore":88.3},{"name":"Bob","avgScore":81}]`
	if err := compareJSON(processedUsers.String(), expectedJSON); err != nil {
		t.Errorf("Map 转换结果不匹配: %v", err)
	}

	// 2. 对特定用户追加一个新分数
	root.Query("/users[0]/scores").Append(100)
	if err := root.Error(); err != nil {
		t.Errorf("Append 操作失败: %v", err)
	}

	var newScores []interface{}
	root.Query("/users[0]/scores").ForEach(func(_ interface{}, val xjson.Node) {
		newScores = append(newScores, val.Interface())
	})
	expectedScores := []interface{}{float64(80), float64(90), float64(95), float64(100)}
	if !reflect.DeepEqual(newScores, expectedScores) {
		t.Errorf("期望的分数: %v, 实际: %v", expectedScores, newScores)
	}

	// 3. 设置一个全新的嵌套对象
	root.Get("metadata").Set("lastUpdated", map[string]interface{}{"by": "test", "timestamp": 12345})
	if err := root.Error(); err != nil {
		t.Errorf("Set 操作失败: %v", err)
	}
	updatedBy := root.Query("/metadata/lastUpdated/by").String()
	if updatedBy != "test" {
		t.Errorf("期望 updatedBy 为 'test', 实际为 '%s'", updatedBy)
	}
}

func TestFunctionManagement(t *testing.T) {
	data := `[1, 5, 10, 15, 20]`
	root, _ := xjson.Parse(data)

	// 1. 注册函数
	root.Func("greaterThan10", func(n xjson.Node) xjson.Node {
		return n.Filter(func(item xjson.Node) bool {
			return item.Int() > 10
		})
	})

	// 2. 通过路径调用
	var result1 []int64
	root.Query("[@greaterThan10]").ForEach(func(_ interface{}, item xjson.Node) {
		result1 = append(result1, item.Int())
	})
	expected := []int64{15, 20}
	if !reflect.DeepEqual(result1, expected) {
		t.Errorf("路径函数调用结果不匹配. 期望 %v, 实际 %v", expected, result1)
	}

	// 3. 直接调用函数
	var result2 []int64
	root.CallFunc("greaterThan10").ForEach(func(_ interface{}, item xjson.Node) {
		result2 = append(result2, item.Int())
	})
	if !reflect.DeepEqual(result2, expected) {
		t.Errorf("CallFunc 调用结果不匹配. 期望 %v, 实际 %v", expected, result2)
	}

	// 4. 移除函数
	root.RemoveFunc("greaterThan10")
	node := root.Query("[@greaterThan10]")
	if node.IsValid() {
		t.Errorf("期望在调用已移除函数后得到一个无效节点")
	}
}

func TestComplexChainedQuery(t *testing.T) {
	// t.Skip("跳过此测试，因为 Query 方法在处理复杂的通配符和路径函数组合时行为不符合预期，这可能反映了核心库的一个问题。")

	data := `{
		"departments": [
			{
				"name": "Engineering",
				"teams": [
					{"name": "Backend", "members": [
						{"name": "Alice", "role": "Senior", "active": true},
						{"name": "Bob", "role": "Junior", "active": false}
					]},
					{"name": "Frontend", "members": [
						{"name": "Charlie", "role": "Senior", "active": true},
						{"name": "David", "role": "Senior", "active": true}
					]}
				]
			},
			{
				"name": "HR",
				"teams": [
					{"name": "Recruiting", "members": [
						{"name": "Eve", "role": "Manager", "active": true}
					]}
				]
			}
		]
	}`
	root, _ := xjson.Parse(data)

	// 查找所有活跃的 "Senior" 工程师
	root.Func("seniors", func(n xjson.Node) xjson.Node {
		return n.Filter(func(member xjson.Node) bool {
			return member.Get("role").String() == "Senior"
		})
	}).Func("active", func(n xjson.Node) xjson.Node {
		return n.Filter(func(member xjson.Node) bool {
			return member.Get("active").Bool()
		})
	})

	// 最终策略：保持在 root 上下文中进行完整的链式调用
	seniorNames := root.Get("departments").Filter(func(dept xjson.Node) bool {
		// 筛选出工程部门
		return dept.Get("name").String() == "Engineering"
	}).Query(
		// 在筛选出的部门上执行查询:
		// 1. `*/teams` -> 获取 teams 数组
		// 2. `/*/members` -> 获取所有 team 的 members, 结果是一个扁平的成员数组
		// 3. `[@seniors][@active]` -> 在成员数组上应用函数
		// 4. `/*/name` -> 从最终的成员对象中提取 name
		"*/teams/*/members[@seniors][@active]/*/name",
	).Strings()

	if err := root.Error(); err != nil {
		t.Errorf("复杂查询失败: %v", err)
	}

	expectedNames := []string{"Alice", "Charlie", "David"}
	if !compareStringSlices(seniorNames, expectedNames) {
		t.Errorf("期望的工程师: %v, 实际: %v", expectedNames, seniorNames)
	}
}

func TestForEachIteration(t *testing.T) {
	data := `{"items": [{"value": 10}, {"value": 20}, {"value": 30}]}`
	root, _ := xjson.Parse(data)

	var sum int64
	var count int
	root.Query("/items/value").ForEach(func(_ interface{}, node xjson.Node) {
		sum += node.Int()
		count++
	})

	if err := root.Error(); err != nil {
		t.Errorf("ForEach 失败: %v", err)
	}
	if count != 3 {
		t.Errorf("期望迭代 3 次, 实际 %d 次", count)
	}
	if sum != 60 {
		t.Errorf("期望总和为 60, 实际为 %d", sum)
	}
}

func TestNullAndEmptyValues(t *testing.T) {
	data := `{
		"a": null,
		"b": [],
		"c": {},
		"d": ""
	}`
	root, _ := xjson.Parse(data)

	// 1. 检查节点类型
	if root.Get("a").Type() != xjson.NullNode {
		t.Error("期望 /a 的类型是 NullNode")
	}
	if root.Get("b").Type() == xjson.NullNode {
		t.Error("期望 /b 的类型不是 NullNode")
	}

	// 2. 在 null 上操作
	node := root.Get("a").Get("key")
	if node.IsValid() {
		t.Error("期望在 null 上 Get 返回一个无效节点")
	}

	// 3. 查询空数组和空对象
	root, _ = xjson.Parse(data) // 重置节点以清除错误
	if root.Query("/b").Len() != 0 {
		t.Error("期望空数组长度为 0")
	}
	if root.Query("/c").Len() != 0 {
		t.Error("期望空对象长度为 0")
	}

	// 4. 获取空字符串
	if root.Get("d").String() != "" {
		t.Error("期望获取到空字符串")
	}
}

// 辅助函数，用于比较字符串切片
func compareStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// compareJSON compares two JSON strings by unmarshaling them into interfaces
// and then using reflect.DeepEqual.
func compareJSON(actual, expected string) error {
	var actualI, expectedI interface{}

	err := json.Unmarshal([]byte(actual), &actualI)
	if err != nil {
		return fmt.Errorf("failed to unmarshal actual JSON: %w", err)
	}
	err = json.Unmarshal([]byte(expected), &expectedI)
	if err != nil {
		return fmt.Errorf("failed to unmarshal expected JSON: %w", err)
	}

	if !reflect.DeepEqual(actualI, expectedI) {
		return errors.New("JSON content mismatch")
	}
	return nil
}
