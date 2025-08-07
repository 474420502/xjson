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

// 辅助函数，用于比较字符串切片
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
