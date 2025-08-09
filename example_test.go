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
	"github.com/stretchr/testify/assert"
)

func TestQuickStartExample(t *testing.T) {
	data := `
	{
		"store": {
			"books": [
				{
					"title": "Moby Dick",
					"price": 8.99,
					"tags": ["classic", "adventure"]
				},
				{
					"title": "Clean Code",
					"price": 29.99,
					"tags": ["programming"]
				}
			]
		}
	}`

	root, err := xjson.Parse(data)
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}

	// 1. 注册函数
	root.RegisterFunc("cheap", func(n xjson.Node) xjson.Node {
		return n.Filter(func(child xjson.Node) bool {
			price, ok := child.Get("price").RawFloat()
			return ok && price < 20
		})
	}).RegisterFunc("tagged", func(n xjson.Node) xjson.Node {
		return n.Filter(func(child xjson.Node) bool {
			return child.Get("tags").Contains("adventure")
		})
	})

	// 2. 查询
	cheapTitlesNode := root.Query("/store/books[@cheap]/title")
	if assert.NoError(t, cheapTitlesNode.Error()) {
		cheapTitles := cheapTitlesNode.Strings()
		if assert.NoError(t, cheapTitlesNode.Error()) {
			expectedCheapTitles := []string{"Moby Dick"}
			assert.Equal(t, expectedCheapTitles, cheapTitles, "期望的廉价书籍不匹配")
		}
	}

	// 3. 修改
	taggedBook := root.Query("/store/books[@tagged]")
	if assert.NoError(t, taggedBook.Error()) && taggedBook.Len() > 0 {
		bookToUpdate := taggedBook.Index(0)
		if assert.NoError(t, bookToUpdate.Error()) {
			assert.NoError(t, bookToUpdate.Set("price", 9.99).Error())
		}
	}

	// 4. 验证
	priceNode := root.Query("/store/books[0]/price")
	if assert.NoError(t, priceNode.Error()) {
		modifiedPrice, ok := priceNode.RawFloat()
		t.Logf("期望修改后的价格为 9.99, 实际为 %v (ok: %v)", modifiedPrice, ok)
		assert.True(t, ok)
		assert.Equal(t, 9.99, modifiedPrice)
	}

	// 5. 序列化
	expectedOutput := `{"store":{"books":[{"price":9.99,"tags":["classic","adventure"],"title":"Moby Dick"},{"price":29.99,"tags":["programming"],"title":"Clean Code"}]}}`
	actualJSONNode := root.String()
	assert.NoError(t, root.Error())
	if err := compareJSON(actualJSONNode, expectedOutput); err != nil {
		t.Errorf("期望的JSON输出不匹配: %v\n期望: %s\n实际: %s", err, expectedOutput, actualJSONNode)
	}

	// 6. 验证序列化结果
	var actualData, expectedData interface{}
	if err := json.Unmarshal([]byte(actualJSONNode), &actualData); err != nil {
		t.Fatalf("无法解析实际JSON: %v", err)
	}
	if err := json.Unmarshal([]byte(expectedOutput), &expectedData); err != nil {
		t.Fatalf("无法解析期望的JSON: %v", err)
	}
	if !reflect.DeepEqual(actualData, expectedData) {
		t.Error("反序列化后的数据结构不匹配")
	}
}

func TestBusinessRuleEncapsulation(t *testing.T) {
	data := `
	{
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
 
	root.RegisterFunc("inStock", func(n xjson.Node) xjson.Node {
		return n.Filter(func(p xjson.Node) bool {
			return p.Get("stock").Int() > 0 &&
				p.Get("status").String() == "active"
		})
	})

	availableProductsNode := root.Query("/products[@inStock]/id")
	if assert.NoError(t, availableProductsNode.Error()) {
		availableProducts := availableProductsNode.Strings()
		if assert.NoError(t, availableProductsNode.Error()) {
			expectedAvailableProducts := []string{"p1", "p4"}
			assert.Equal(t, expectedAvailableProducts, availableProducts, "期望的可用产品不匹配")
		}
	}
}

func TestDataTransformationPipeline(t *testing.T) {
	data := `
	{
		"rawInput": [
			{"id": "item1", "name": "  Product A  ", "price": 10.123},
			{"id": "item2", "name": "Product B", "price": 20.456}
		]
	}`

	root, err := xjson.Parse(data)
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}

	root.RegisterFunc("sanitize", func(n xjson.Node) xjson.Node {
		return n.Map(func(item xjson.Node) interface{} {
			return map[string]interface{}{
				"id":    item.Get("id").String(),
				"name":  strings.TrimSpace(item.Get("name").String()),
				"price": math.Round(item.Get("price").Float()*100) / 100,
			}
		})
	})

	cleanData := root.Query("/rawInput[@sanitize]")
	if assert.NoError(t, cleanData.Error()) {

		item1 := cleanData.Index(0)
		if assert.NoError(t, item1.Error()) {
			id := item1.Get("id")
			if assert.NoError(t, id.Error()) {
				assert.Equal(t, "item1", id.String())
			}
			name := item1.Get("name")
			if assert.NoError(t, name.Error()) {
				assert.Equal(t, "Product A", name.String())
			}
			price := item1.Get("price")
			if assert.NoError(t, price.Error()) {
				if p, ok := price.RawFloat(); ok {
					assert.Equal(t, 10.12, p)
				} else {
					t.Error("无法将价格转换为 float")
				}
			}
		}

		item2 := cleanData.Index(1)
		if assert.NoError(t, item2.Error()) {
			id := item2.Get("id")
			if assert.NoError(t, id.Error()) {
				assert.Equal(t, "item2", id.String())
			}
			name := item2.Get("name")
			if assert.NoError(t, name.Error()) {
				assert.Equal(t, "Product B", name.String())
			}
			price := item2.Get("price")
			if assert.NoError(t, price.Error()) {
				if p, ok := price.RawFloat(); ok {
					assert.Equal(t, 20.46, p)
				} else {
					t.Error("无法将价格转换为 float")
				}
			}
		}
	}
}

func TestComprehensiveErrorHandling(t *testing.T) {
	data := `{"a":{"b":[1,2,3]}}`
	root, err := xjson.Parse(data)
	assert.NoError(t, err)

	t.Run("PathNotFound", func(t *testing.T) {
		node := root.Query("/a/c")
		assert.Error(t, node.Error())
		assert.False(t, node.IsValid())
	})

	t.Run("IndexOutOfBounds", func(t *testing.T) {
		node := root.Query("/a/b[5]")
		assert.Error(t, node.Error())
		assert.False(t, node.IsValid())
	})

	t.Run("GetOnNonObject", func(t *testing.T) {
		node := root.Query("/a/b").Get("key")
		assert.Error(t, node.Error())
		assert.False(t, node.IsValid())
	})

	t.Run("UnregisteredFunction", func(t *testing.T) {
		node := root.Query("/a/b[@nonexistent]")
		assert.Error(t, node.Error())
		assert.False(t, node.IsValid())
	})
}

func TestAdvancedDataManipulation(t *testing.T) {
	data := `{"users":[{"id":1,"name":"Alice","scores":[80,90,95]},{"id":2,"name":"Bob","scores":[70,85,88]}],"metadata":{}}`
	root, err := xjson.Parse(data)
	if err != nil {
		t.Fatal(err)
	}

	root.RegisterFunc("withAvg", func(n xjson.Node) xjson.Node {
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
	if assert.NoError(t, processedUsers.Error()) {
		expectedJSON := `[{"avgScore":88.3,"name":"Alice"},{"avgScore":81,"name":"Bob"}]`
		actualJSON := processedUsers.String()
		assert.NoError(t, processedUsers.Error())
		if err := compareJSON(actualJSON, expectedJSON); err != nil {
			t.Errorf("Map 转换结果不匹配: %v", err)
		}
	}

	scoresNode := root.Query("/users[0]/scores")
	if assert.NoError(t, scoresNode.Error()) {
		assert.NoError(t, scoresNode.Append(100).Error())
	}
	refreshedScoresNode := root.Query("/users[0]/scores")
	if assert.NoError(t, refreshedScoresNode.Error()) {
		var newScores []int64
		refreshedScoresNode.ForEach(func(_ interface{}, val xjson.Node) {
			if assert.NoError(t, val.Error()) {
				newScores = append(newScores, val.Int())
			}
		})
		expectedScores := []int64{80, 90, 95, 100}
		assert.Equal(t, expectedScores, newScores, "分数不匹配")
	}

	metadataNode := root.Get("metadata")
	if assert.NoError(t, metadataNode.Error()) {
		assert.NoError(t, metadataNode.Set("lastUpdated", map[string]interface{}{"by": "test", "timestamp": 12345}).Error())
	}
	updatedByNode := root.Query("/metadata/lastUpdated/by")
	if assert.NoError(t, updatedByNode.Error()) {
		assert.Equal(t, "test", updatedByNode.String(), "updatedBy 不匹配")
	}
}

func TestFunctionManagement(t *testing.T) {
	data := `[1,5,10,15,20]`
	root, _ := xjson.Parse(data)

	root.RegisterFunc("greaterThan10", func(n xjson.Node) xjson.Node {
		return n.Filter(func(item xjson.Node) bool {
			return item.Int() > 10
		})
	})

	var result1 []int64
	greaterThan10Node := root.Query("[@greaterThan10]")
	if assert.NoError(t, greaterThan10Node.Error()) {
		greaterThan10Node.ForEach(func(_ interface{}, item xjson.Node) {
			if assert.NoError(t, item.Error()) {
				result1 = append(result1, item.Int())
			}
		})
	}
	expected := []int64{15, 20}
	assert.Equal(t, expected, result1, "路径函数调用结果不匹配")

	var result2 []int64
	callFuncNode := root.CallFunc("greaterThan10")
	if assert.NoError(t, callFuncNode.Error()) {
		callFuncNode.ForEach(func(_ interface{}, item xjson.Node) {
			if assert.NoError(t, item.Error()) {
				result2 = append(result2, item.Int())
			}
		})
	}
	assert.Equal(t, expected, result2, "CallFunc 调用结果不匹配")

	root.RemoveFunc("greaterThan10")
	node := root.Query("[@greaterThan10]")
	assert.Error(t, node.Error(), "期望在调用已移除函数后得到一个错误")
	assert.False(t, node.IsValid(), "期望节点无效")
}

func TestComplexChainedQuery(t *testing.T) {
	data := `{"departments":[{"name":"Engineering","teams":[{"name":"Backend","members":[{"name":"Alice","role":"Senior","active":true},{"name":"Bob","role":"Junior","active":false}]},{"name":"Frontend","members":[{"name":"Charlie","role":"Senior","active":true},{"name":"David","role":"Senior","active":true}]}]},{"name":"HR","teams":[{"name":"Recruiting","members":[{"name":"Eve","role":"Manager","active":true}]}]}]}`
	root, _ := xjson.Parse(data)

	root.RegisterFunc("seniors", func(n xjson.Node) xjson.Node {
		return n.Filter(func(member xjson.Node) bool {
			return member.Get("role").String() == "Senior"
		})
	}).RegisterFunc("active", func(n xjson.Node) xjson.Node {
		return n.Filter(func(member xjson.Node) bool {
			return member.Get("active").Bool()
		})
	})

	departments := root.Get("departments")
	if assert.NoError(t, departments.Error()) {
		engineeringDept := departments.Filter(func(dept xjson.Node) bool {
			name := dept.Get("name")
			return name.Error() == nil && name.String() == "Engineering"
		})

		if assert.NoError(t, engineeringDept.Error()) {
			namesNode := engineeringDept.Query("*/teams/*/members[@seniors][@active]/*/name")
			if assert.NoError(t, namesNode.Error()) {
				seniorNames := namesNode.Strings()
				assert.NoError(t, namesNode.Error())
				expectedNames := []string{"Alice", "Charlie", "David"}
				assert.Equal(t, expectedNames, seniorNames, "期望的工程师不匹配")
			}
		}
	}
}

func TestForEachIteration(t *testing.T) {
	data := `{"items":[{"value":10},{"value":20},{"value":30}]}`
	root, _ := xjson.Parse(data)

	var sum int64
	var count int
	items := root.Query("/items/value")
	if assert.NoError(t, items.Error()) {
		items.ForEach(func(_ interface{}, node xjson.Node) {
			if assert.NoError(t, node.Error()) {
				sum += node.Int()
				count++
			}
		})

		assert.Equal(t, 3, count, "迭代次数不匹配")
		assert.Equal(t, int64(60), sum, "总和不匹配")
	}
}

func TestNullAndEmptyValues(t *testing.T) {
	data := `{"a":null,"b":[],"c":{},"d":""}`
	root, _ := xjson.Parse(data)

	nodeA := root.Get("a")
	if assert.NoError(t, nodeA.Error()) {
		assert.Equal(t, xjson.NullNode, nodeA.Type())
	}
	nodeB := root.Get("b")
	if assert.NoError(t, nodeB.Error()) {
		assert.NotEqual(t, xjson.NullNode, nodeB.Type())
	}

	node := root.Get("a").Get("key")
	assert.Error(t, node.Error())
	assert.False(t, node.IsValid())

	root, _ = xjson.Parse(data)
	nodeB = root.Query("/b")
	if assert.NoError(t, nodeB.Error()) {
		assert.Equal(t, 0, nodeB.Len())
	}
	nodeC := root.Query("/c")
	if assert.NoError(t, nodeC.Error()) {
		assert.Equal(t, 0, nodeC.Len())
	}

	nodeD := root.Get("d")
	if assert.NoError(t, nodeD.Error()) {
		assert.Equal(t, "", nodeD.String())
	}
}

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

func mustPanic(t *testing.T, fn func(), msg string) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("期望 panic: %s", msg)
		}
	}()
	fn()
}

func TestRawAccessAndMustMethods(t *testing.T) {
	data := `{"s":"hello","n":123.45,"b":true,"arr":[1,2,3],"ts":"2024-12-31T23:59:59Z"}`
	root, err := xjson.Parse(data)
	if err != nil {
		t.Fatal(err)
	}

	sNode := root.Get("s")
	if assert.NoError(t, sNode.Error()) {
		if v, ok := sNode.RawString(); assert.True(t, ok) {
			assert.Equal(t, "hello", v)
		}
	}
	nNode := root.Get("n")
	if assert.NoError(t, nNode.Error()) {
		if f, ok := nNode.RawFloat(); assert.True(t, ok) {
			assert.Equal(t, 123.45, f)
		}
	}
	bNode := root.Get("b")
	if assert.NoError(t, bNode.Error()) {
		assert.True(t, bNode.Bool())
	}
	arrNode := root.Get("arr")
	if assert.NoError(t, arrNode.Error()) {
		assert.Equal(t, 3, arrNode.Len())
	}
	tsNode := root.Get("ts")
	if assert.NoError(t, tsNode.Error()) {
		assert.False(t, tsNode.Time().IsZero())
	}

	mustPanic(t, func() { root.Get("arr").MustString() }, "MustString 应 panic")
	mustPanic(t, func() { root.Get("s").MustInt() }, "MustInt 应 panic")
	mustPanic(t, func() { root.Get("n").MustBool() }, "MustBool 应 panic")
}

func TestArraySetOnMixedTypesError(t *testing.T) {
	data := `{"arr":[{"a":1},2,{"a":3}]}`
	root, _ := xjson.Parse(data)
	arr := root.Query("/arr")
	arr.Set("anyKey", "anyValue")
	t.Log("期望 Set 在混合类型数组上产生错误 (检查目标节点的 Error)")
	assert.Error(t, arr.Error())
}

func TestAppendOnObjectError(t *testing.T) {
	data := `{"obj":{"a":1}}`
	root, _ := xjson.Parse(data)
	obj := root.Query("/obj")
	obj.Append(2)
	t.Log("期望在对象上 Append 产生错误 (检查目标节点)")
	assert.Error(t, obj.Error())
}

func TestStringsAndContains(t *testing.T) {
	data := `{"tags":["go","json","query"],"mixed":["ok",1]}`
	root, _ := xjson.Parse(data)
	if !root.Query("/tags").Contains("json") {
		t.Error("Contains 失败")
	}
	ss := root.Query("/tags").Strings()
	if len(ss) != 3 || ss[1] != "json" {
		t.Error("Strings 返回不正确")
	}
	mixed := root.Query("/mixed")
	if assert.NoError(t, mixed.Error()) {
		_ = mixed.Strings()
		assert.Error(t, mixed.Error(), "期望 mixed.Strings() 在包含非字符串元素时产生错误")
	}
}

func TestNegativeIndexHandling(t *testing.T) {
	data := `{"a":[10,20]}`
	root, _ := xjson.Parse(data)
	node := root.Query("/a[-1]")
	assert.Error(t, node.Error())
	assert.False(t, node.IsValid())
}

func TestFunctionChainingAndRemove(t *testing.T) {
	data := `[{"x":1},{"x":2},{"x":3},{"x":4}]`
	root, _ := xjson.Parse(data)
	root.RegisterFunc("gt2", func(n xjson.Node) xjson.Node {
		return n.Filter(func(c xjson.Node) bool { return c.Get("x").Int() > 2 })
	})
	root.RegisterFunc("dbl", func(n xjson.Node) xjson.Node {
		return n.Map(func(c xjson.Node) interface{} { return map[string]int{"x": int(c.Get("x").Int() * 2)} })
	})
	chain := root.Query("[@gt2][@dbl]/*/x")
	if assert.NoError(t, chain.Error()) {
		mapped := chain.Map(func(n xjson.Node) interface{} {
			if n.Error() != nil {
				return n.Error()
			}
			return n.Int()
		})

		if assert.NoError(t, mapped.Error()) {
			assert.True(t, mapped.IsValid())
			vals := mapped.Interface()
			slice, ok := vals.([]interface{})
			if assert.True(t, ok, "期望返回切片") {
				assert.Equal(t, 2, len(slice))
				assert.InDelta(t, 6, slice[0], 1e-9)
				assert.InDelta(t, 8, slice[1], 1e-9)
			}
		}
	}
	root.RemoveFunc("gt2")
	if root.Query("[@gt2]").IsValid() {
		t.Error("移除函数后调用应无效")
	}
}

func TestMapAfterWildcardFlatten(t *testing.T) {
	data := `{"groups":[{"items":[1,2]},{"items":[3]}]}`
	root, _ := xjson.Parse(data)
	root.RegisterFunc("asIs", func(n xjson.Node) xjson.Node { return n })
	node := root.Query("/groups/*/items[@asIs]/*")
	if assert.NoError(t, node.Error()) {
		mapped := node.Map(func(n xjson.Node) interface{} {
			if n.Error() != nil {
				return n.Error()
			}
			return n.Int()
		})

		if assert.NoError(t, mapped.Error()) {
			vals := mapped.Interface()
			arr, ok := vals.([]interface{})
			if assert.True(t, ok) {
				assert.Len(t, arr, 3)
				assert.InDelta(t, 1, arr[0], 1e-9)
				assert.InDelta(t, 2, arr[1], 1e-9)
				assert.InDelta(t, 3, arr[2], 1e-9)
			}
		}
	}
}

func TestDebugBusinessScenario(t *testing.T) {
	ecommerceJSON := `{"store":{"products":[{"name":"Product 1"}]}}`

	root, err := xjson.Parse(ecommerceJSON)
	assert.NoError(t, err)
	assert.True(t, root.IsValid())

	products := root.Get("store").Get("products")
	if assert.NoError(t, products.Error()) {
		assert.Equal(t, 1, products.Len())

		newProduct := map[string]interface{}{"name": "Product 2"}
		assert.NoError(t, products.Append(newProduct).Error())

		t.Logf("Products length: %d", products.Len())
		assert.Equal(t, 2, products.Len())

		name1 := products.Index(0).Get("name")
		if assert.NoError(t, name1.Error()) {
			assert.Equal(t, "Product 1", name1.String())
		}
		name2 := products.Index(1).Get("name")
		if assert.NoError(t, name2.Error()) {
			assert.Equal(t, "Product 2", name2.String())
		}
	}
}

func TestDebugAppendIssue(t *testing.T) {
	ecommerceJSON := `{"store":{"products":[{"name":"Product 1"}]}}`

	root, err := xjson.Parse(ecommerceJSON)
	assert.NoError(t, err)
	assert.True(t, root.IsValid())

	products := root.Get("store").Get("products")
	if assert.NoError(t, products.Error()) {
		t.Logf("Initial products length: %d", products.Len())
		assert.Equal(t, 1, products.Len())

		newProduct := map[string]interface{}{"name": "Product 2"}
		assert.NoError(t, products.Append(newProduct).Error())
		t.Logf("Products length after append (same ref): %d", products.Len())

		freshProducts := root.Get("store").Get("products")
		if assert.NoError(t, freshProducts.Error()) {
			t.Logf("Products length after append (fresh ref): %d", freshProducts.Len())
			assert.Equal(t, 2, products.Len())
			assert.Equal(t, 2, freshProducts.Len())
		}
	}
}
