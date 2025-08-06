package xjson

import (
	"fmt"
	"testing"
)

func TestXPathSliceDetailed(t *testing.T) {
	t.Run("Detailed slice tests", func(t *testing.T) {
		jsonStr := `{
		"books": [
			{"title": "Go Programming"},
			{"title": "Web Development"},
			{"title": "Fiction Story"}
		]
	}`

		x, err := ParseString(jsonStr)
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}

		// 测试不同的切片
		tests := []struct {
			path     string
			expected int
		}{
			{"/books[0:1]", 1},
			{"/books[1:2]", 1},
			{"/books[0:2]", 2},
			{"/books[-1:]", 2}, // [-1:] 应该返回最后2个元素
			{"/books[-2:-1]", 1},
			{"/books[-3:-2]", 1},
		}

		for _, test := range tests {
			result := x.Query(test.path)
			fmt.Printf("Path: %s, Expected: %d, Actual: %d\n",
				test.path, test.expected, result.Count())

			if result.Count() != test.expected {
				t.Errorf("Path %s: expected %d, got %d",
					test.path, test.expected, result.Count())
			}
		}
	})
}
