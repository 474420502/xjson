package main

import (
	"fmt"
	"log"

	"github.com/474420502/xjson"
)

func main() {
	// 示例JSON数据
	jsonData := `{
		"store": {
			"products": [
				{"id": 1, "name": "Laptop", "price": 999.99, "category": "electronics", "inStock": true},
				{"id": 2, "name": "Mouse", "price": 25.50, "category": "electronics", "inStock": true},
				{"id": 3, "name": "Desk", "price": 150.00, "category": "furniture", "inStock": false},
				{"id": 4, "name": "Chair", "price": 85.00, "category": "furniture", "inStock": true},
				{"id": 5, "name": "Book", "price": 12.99, "category": "education", "inStock": true}
			],
			"customers": [
				{"name": "Alice", "age": 30, "premium": true},
				{"name": "Bob", "age": 25, "premium": false},
				{"name": "Charlie", "age": 35, "premium": true}
			]
		}
	}`

	doc, err := xjson.ParseString(jsonData)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("=== Filter Expression Examples ===")
	fmt.Println()

	// 示例1: 基本价格过滤器
	fmt.Println("1. Products under $100:")
	result := doc.Query("store.products[?(@.price < 100)]")
	if result.Exists() {
		for i := 0; i < result.Count(); i++ {
			name, _ := result.Index(i).Get("name").String()
			price, _ := result.Index(i).Get("price").Float()
			fmt.Printf("   - %s: $%.2f\n", name, price)
		}
	}
	fmt.Println()

	// 示例2: 类别过滤器
	fmt.Println("2. Electronics products:")
	result = doc.Query("store.products[?(@.category == 'electronics')]")
	if result.Exists() {
		for i := 0; i < result.Count(); i++ {
			name, _ := result.Index(i).Get("name").String()
			fmt.Printf("   - %s\n", name)
		}
	}
	fmt.Println()

	// 示例3: 布尔过滤器
	fmt.Println("3. Products in stock:")
	result = doc.Query("store.products[?(@.inStock == true)]")
	if result.Exists() {
		for i := 0; i < result.Count(); i++ {
			name, _ := result.Index(i).Get("name").String()
			fmt.Printf("   - %s\n", name)
		}
	}
	fmt.Println()

	// 示例4: 复杂AND过滤器
	fmt.Println("4. Affordable products in stock (price < $100 AND inStock):")
	result = doc.Query("store.products[?(@.price < 100 && @.inStock == true)]")
	if result.Exists() {
		for i := 0; i < result.Count(); i++ {
			name, _ := result.Index(i).Get("name").String()
			price, _ := result.Index(i).Get("price").Float()
			fmt.Printf("   - %s: $%.2f\n", name, price)
		}
	}
	fmt.Println()

	// 示例5: 复杂OR过滤器
	fmt.Println("5. Expensive products OR education category:")
	result = doc.Query("store.products[?(@.price > 500 || @.category == 'education')]")
	if result.Exists() {
		for i := 0; i < result.Count(); i++ {
			name, _ := result.Index(i).Get("name").String()
			price, _ := result.Index(i).Get("price").Float()
			category, _ := result.Index(i).Get("category").String()
			fmt.Printf("   - %s: $%.2f (%s)\n", name, price, category)
		}
	}
	fmt.Println()

	// 示例6: 客户过滤器
	fmt.Println("6. Premium customers over 30:")
	result = doc.Query("store.customers[?(@.age > 30 && @.premium == true)]")
	if result.Exists() {
		for i := 0; i < result.Count(); i++ {
			name, _ := result.Index(i).Get("name").String()
			age, _ := result.Index(i).Get("age").Int()
			fmt.Printf("   - %s (age %d)\n", name, age)
		}
	}
}
