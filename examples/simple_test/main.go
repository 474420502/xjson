package main

import (
	"fmt"
	"log"

	"github.com/474420502/xjson"
)

func main() {
	// 简化的测试数据
	jsonData := `{
		"products": [
			{"name": "Mouse", "price": 25.50},
			{"name": "Chair", "price": 85.00}
		]
	}`

	doc, err := xjson.ParseString(jsonData)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Testing basic filter:")
	result := doc.Query("products[?(@.price < 100)]")
	fmt.Printf("Result exists: %v\n", result.Exists())
	fmt.Printf("Result count: %v\n", result.Count())

	if result.Exists() {
		for i := 0; i < result.Count(); i++ {
			item := result.Index(i)
			fmt.Printf("Item %d exists: %v\n", i, item.Exists())

			// 试试直接打印Raw数据
			fmt.Printf("Item %d raw: %v\n", i, item)
		}
	}
}
