package main

import (
	"fmt"
	"log"

	"github.com/474420502/xjson"
)

func main() {
	jsonData := `{
		"store": {
			"book": [
				{"title": "Book 1", "price": 10.0},
				{"title": "Book 2", "price": 20.0},
				{"title": "Book 3", "price": 30.0},
				{"title": "Book 4", "price": 40.0}
			]
		}
	}`

	doc, err := xjson.ParseString(jsonData)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("=== Debug Query Results ===")

	// Test store.book query
	result := doc.Query("store.book")
	fmt.Printf("store.book exists: %t\n", result.Exists())
	fmt.Printf("store.book is array: %t\n", result.IsArray())
	fmt.Printf("store.book count: %d\n", result.Count())

	if result.Exists() {
		raw := result.Raw()
		fmt.Printf("Raw type: %T\n", raw)
		fmt.Printf("Raw value: %v\n", raw)
	}

	// Test book[1:3] query
	fmt.Println("\n=== Test Array Slice ===")
	sliceResult := doc.Query("store.book[1:3]")
	fmt.Printf("store.book[1:3] exists: %t\n", sliceResult.Exists())
	fmt.Printf("store.book[1:3] is array: %t\n", sliceResult.IsArray())
	fmt.Printf("store.book[1:3] count: %d\n", sliceResult.Count())

	// Test recursive query
	fmt.Println("\n=== Test Recursive Query ===")
	pricesResult := doc.Query("store..price")
	fmt.Printf("store..price exists: %t\n", pricesResult.Exists())
	fmt.Printf("store..price is array: %t\n", pricesResult.IsArray())
	fmt.Printf("store..price count: %d\n", pricesResult.Count())

	// Test filter expression: price < 30
	fmt.Println("\n=== Test Filter Expression ===")
	filterResult := doc.Query("store.book[?(@.price < 30)]")
	fmt.Printf("store.book[?(@.price < 30)] exists: %t\n", filterResult.Exists())
	fmt.Printf("store.book[?(@.price < 30)] count: %d\n", filterResult.Count())
	for i := 0; i < filterResult.Count(); i++ {
		title, _ := filterResult.Index(i).Get("title").String()
		price, _ := filterResult.Index(i).Get("price").Float()
		fmt.Printf("  - %s: %.2f\n", title, price)
	}

	// Test filter with AND
	fmt.Println("\n=== Test Filter AND Expression ===")
	andResult := doc.Query("store.book[?(@.price >= 20 && @.price <= 30)]")
	fmt.Printf("store.book[?(@.price >= 20 && @.price <= 30)] count: %d\n", andResult.Count())
	for i := 0; i < andResult.Count(); i++ {
		title, _ := andResult.Index(i).Get("title").String()
		price, _ := andResult.Index(i).Get("price").Float()
		fmt.Printf("  - %s: %.2f\n", title, price)
	}

	// Test recursive query for all prices
	fmt.Println("\n=== Test Recursive Query for All Prices ===")
	allPrices := doc.Query("..price")
	fmt.Printf("..price exists: %t\n", allPrices.Exists())
	fmt.Printf("..price count: %d\n", allPrices.Count())
	for i := 0; i < allPrices.Count(); i++ {
		price, _ := allPrices.Index(i).Float()
		fmt.Printf("  - price: %.2f\n", price)
	}
}
