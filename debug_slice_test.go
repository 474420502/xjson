package xjson

import (
	"fmt"
	"testing"
)

func TestXPathSliceDebug(t *testing.T) {
	jsonStr := `{
		"library": {
			"books": [
				{"title": "Go Programming"},
				{"title": "Web Development"},
				{"title": "Fiction Story"}
			]
		}
	}`

	x, err := ParseString(jsonStr)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// 测试负数切片
	result := x.Query("/library/books[-2:-1]")
	fmt.Printf("Result exists: %v\n", result.Exists())
	fmt.Printf("Result count: %d\n", result.Count())

	if result.Exists() {
		result.ForEach(func(i int, v IResult) bool {
			title, _ := v.String()
			fmt.Printf("Book %d: %s\n", i, title)
			return true
		})
	}

	// 测试正数切片对比
	result2 := x.Query("/library/books[1:2]")
	fmt.Printf("\nPositive slice [1:2] exists: %v\n", result2.Exists())
	fmt.Printf("Positive slice [1:2] count: %d\n", result2.Count())

	if result2.Exists() {
		result2.ForEach(func(i int, v IResult) bool {
			title, _ := v.String()
			fmt.Printf("Positive slice book %d: %s\n", i, title)
			return true
		})
	}
}
