package xjson

import (
	"fmt"
	"testing"
)

func TestNegativeSliceLogic(t *testing.T) {
	jsonStr := `{
		"books": [
			{"title": "Book 0"},
			{"title": "Book 1"},
			{"title": "Book 2"}
		]
	}`

	x, err := ParseString(jsonStr)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// 测试负数切片
	result := x.Query("/books[-2:-1]")
	fmt.Printf("Negative slice [-2:-1] count: %d\n", result.Count())

	if result.Exists() {
		result.ForEach(func(i int, v IResult) bool {
			title, _ := v.String()
			fmt.Printf("Book %d: %s\n", i, title)
			return true
		})
	}

	// 对比正数切片
	result2 := x.Query("/books[1:2]")
	fmt.Printf("Positive slice [1:2] count: %d\n", result2.Count())

	if result2.Exists() {
		result2.ForEach(func(i int, v IResult) bool {
			title, _ := v.String()
			fmt.Printf("Positive slice book %d: %s\n", i, title)
			return true
		})
	}
}
