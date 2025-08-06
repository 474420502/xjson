package xjson

import (
	"fmt"
	"testing"
)

func TestFilterEvaluation(t *testing.T) {
	jsonData := `{
		"products": [
			{"id": 1, "name": "Laptop", "price": 999.99, "category": "electronics", "inStock": true},
			{"id": 2, "name": "Mouse", "price": 25.50, "category": "electronics", "inStock": true}
		]
	}`

	doc, err := ParseString(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Force materialization
	doc.Query("/products")

	// Get the products array directly to test filtering
	products := doc.Query("/products")
	fmt.Printf("Products count: %d\n", products.Count())

	// Access the internal array for testing
	// Since we can't directly access internal methods, let's use a direct approach

	// Test if we can access individual products
	prod1 := doc.Query("/products[0]")
	fmt.Printf("Product 1 exists: %t\n", prod1.Exists())

	if prod1.Exists() {
		price1 := doc.Query("/products[0]/price")
		priceVal1, _ := price1.Float()
		fmt.Printf("Product 1 price: %v\n", priceVal1)

		// Test the comparison logic manually
		fmt.Printf("Price < 100: %t\n", priceVal1 < 100)
	}

	prod2 := doc.Query("/products[1]")
	if prod2.Exists() {
		price2 := doc.Query("/products[1]/price")
		priceVal2, _ := price2.Float()
		fmt.Printf("Product 2 price: %v\n", priceVal2)
		fmt.Printf("Price < 100: %t\n", priceVal2 < 100)
	}

	// Now try the actual filter
	fmt.Println("\n=== Testing actual filter ===")
	filterResult := doc.Query("/products[?(@.price < 100)]")
	fmt.Printf("Filter result exists: %t, count: %d\n", filterResult.Exists(), filterResult.Count())

	// Try a different filter format that might work
	fmt.Println("\n=== Testing alternative filters ===")

	// Try with string comparison
	filterResult2 := doc.Query("/products[?(@.category == 'electronics')]")
	fmt.Printf("Category filter result exists: %t, count: %d\n", filterResult2.Exists(), filterResult2.Count())

	// Try with integer comparison
	filterResult3 := doc.Query("/products[?(@.id == 2)]")
	fmt.Printf("ID filter result exists: %t, count: %d\n", filterResult3.Exists(), filterResult3.Count())
}
