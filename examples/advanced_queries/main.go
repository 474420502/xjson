package main

import (
	"fmt"
	"log"

	"github.com/474420502/xjson"
)

func main() {
	fmt.Println("=== XJSON Advanced Query Features Demo ===")
	fmt.Println()

	jsonData := `{
		"store": {
			"name": "Tech Store",
			"location": "San Francisco",
			"book": [
				{"category": "reference", "author": "Nigel Rees", "title": "Sayings of the Century", "isbn": "0-553-21311-3", "price": 8.95, "available": true},
				{"category": "fiction", "author": "Evelyn Waugh", "title": "Sword of Honour", "isbn": "0-679-43136-5", "price": 12.99, "available": false},
				{"category": "fiction", "author": "Herman Melville", "title": "Moby Dick", "isbn": "0-553-21311-3", "price": 8.99, "available": true},
				{"category": "fiction", "author": "J. R. R. Tolkien", "title": "The Lord of the Rings", "isbn": "0-395-19395-8", "price": 22.99, "available": true}
			],
			"bicycle": {"color": "red", "price": 19.95},
			"electronics": [
				{"name": "Laptop", "price": 999.99, "brand": "TechCorp"},
				{"name": "Mouse", "price": 25.50, "brand": "TechCorp"},
				{"name": "Keyboard", "price": 75.00, "brand": "TypeMaster"}
			]
		},
		"customer": {
			"name": "John Doe",
			"preferences": {
				"categories": ["fiction", "electronics"],
				"priceRange": {"min": 10, "max": 100}
			}
		}
	}`

	doc, err := xjson.ParseString(jsonData)
	if err != nil {
		log.Fatal(err)
	}

	// 1. Array Queries
	fmt.Println("1. Array Queries:")
	books := doc.Query("store.book")
	fmt.Printf("   Total books: %d\n", books.Count())
	firstBook := doc.Query("store.book[0]")
	fmt.Printf("   First book: %s\n", firstBook.Get("title").MustString())
	lastBook := doc.Query("store.book[-1]")
	fmt.Printf("   Last book: %s\n", lastBook.Get("title").MustString())

	// 2. Array Slicing
	fmt.Println("\n2. Array Slicing:")
	middleBooks := doc.Query("store.book[1:3]")
	fmt.Printf("   Books [1:3] count: %d\n", middleBooks.Count())
	middleBooks.ForEach(func(index int, book xjson.IResult) bool {
		title := book.Get("title").MustString()
		fmt.Printf("   - [%d] %s\n", index, title)
		return true
	})

	// 3. Recursive Queries (..)
	fmt.Println("\n3. Recursive Queries:")
	allPrices := doc.Query("store..price")
	fmt.Printf("   All prices found: %d\n", allPrices.Count())
	var totalValue float64
	allPrices.ForEach(func(index int, price xjson.IResult) bool {
		value := price.MustFloat()
		totalValue += value
		fmt.Printf("   - Price %d: $%.2f\n", index+1, value)
		return true
	})
	fmt.Printf("   Total value: $%.2f\n", totalValue)

	// 4. Recursive Name Search
	fmt.Println("\n4. Recursive Name Search:")
	allNames := doc.Query("..name")
	fmt.Printf("   All 'name' fields found: %d\n", allNames.Count())
	allNames.ForEach(func(index int, name xjson.IResult) bool {
		fmt.Printf("   - Name %d: %s\n", index+1, name.MustString())
		return true
	})

	// 5. Complex Nested Queries
	fmt.Println("\n5. Complex Nested Queries:")
	bookTitles, _ := doc.Query("store.book").Map(func(index int, book xjson.IResult) (interface{}, error) {
		return book.Get("title").String()
	})
	fmt.Printf("   Book titles: %v\n", bookTitles)
	fmt.Printf("   Available books:\n")
	doc.Query("store.book").ForEach(func(index int, book xjson.IResult) bool {
		if book.Get("available").MustBool() {
			title := book.Get("title").MustString()
			price := book.Get("price").MustFloat()
			fmt.Printf("   - %s ($%.2f)\n", title, price)
		}
		return true
	})

	// 6. Multiple Path Queries
	fmt.Println("\n6. Multiple Path Queries:")
	electronicsNames, _ := doc.Query("store.electronics").Map(func(index int, item xjson.IResult) (interface{}, error) {
		name, _ := item.Get("name").String()
		price, _ := item.Get("price").Float()
		return fmt.Sprintf("%s ($%.2f)", name, price), nil
	})
	fmt.Printf("   Electronics: %v\n", electronicsNames)

	// 7. Path Validation and Safety
	fmt.Println("\n7. Path Validation:")
	nonExistent := doc.Query("store.nonexistent.path")
	fmt.Printf("   Non-existent path exists: %t\n", nonExistent.Exists())
	customerAge := doc.Query("customer.age")
	if customerAge.Exists() {
		fmt.Printf("   Customer age: %d\n", customerAge.MustInt())
	} else {
		fmt.Printf("   Customer age: not specified\n")
	}

	// 8. Array Range Edge Cases
	fmt.Println("\n8. Array Range Edge Cases:")
	emptySlice := doc.Query("store.book[10:15]")
	fmt.Printf("   Empty slice count: %d\n", emptySlice.Count())
	fromMiddle := doc.Query("store.book[2:]")
	fmt.Printf("   From index 2 to end: %d books\n", fromMiddle.Count())
	toMiddle := doc.Query("store.book[:2]")
	fmt.Printf("   From start to index 2: %d books\n", toMiddle.Count())

	fmt.Println("\n=== Demo Complete ===")
}
