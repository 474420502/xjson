package xjson

import (
	"fmt"
	"testing"
)

func TestDebugFilter(t *testing.T) {
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

	// Test basic path access first
	products := doc.Query("/products")
	fmt.Printf("Products exists: %t, count: %d\n", products.Exists(), products.Count())

	// Test array access
	firstProduct := doc.Query("/products[0]")
	fmt.Printf("First product exists: %t\n", firstProduct.Exists())
	if firstProduct.Exists() {
		price := doc.Query("/products[0]/price")
		priceVal, _ := price.Float()
		fmt.Printf("First product price: %v\n", priceVal)
	}

	// Test simple array filter for debugging
	secondProduct := doc.Query("/products[1]")
	fmt.Printf("Second product exists: %t\n", secondProduct.Exists())
	if secondProduct.Exists() {
		price2 := doc.Query("/products[1]/price")
		priceVal2, _ := price2.Float()
		fmt.Printf("Second product price: %v\n", priceVal2)
	}

	// Force materialization and try again
	fmt.Println("\n--- Forcing materialization ---")
	_ = doc.Query("dummy") // This will force materialization

	// Now test filter - step by step
	fmt.Println("\n--- Testing filter step by step ---")

	// Test the filter expression
	filterQuery := "/products[?(@.price < 100)]"
	fmt.Printf("Filter query: %s\n", filterQuery)

	result := doc.Query(filterQuery)
	fmt.Printf("Filter result exists: %t\n", result.Exists())
	fmt.Printf("Filter result count: %d\n", result.Count())

	// Test specific cases to debug the filter logic
	fmt.Println("\n--- Debugging specific filter cases ---")

	// Test a simple equality filter that should match
	idFilter := "/products[?(@.id == 1)]"
	fmt.Printf("ID filter query: %s\n", idFilter)
	idResult := doc.Query(idFilter)
	fmt.Printf("ID filter result exists: %t, count: %d\n", idResult.Exists(), idResult.Count())
}
