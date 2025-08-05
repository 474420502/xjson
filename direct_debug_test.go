package xjson

import (
	"fmt"
	"strings"
	"testing"
)

func TestDirectFilterExecution(t *testing.T) {
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

	// Force materialization first
	_ = doc.Query("products")

	fmt.Println("=== Testing path parsing ===")

	path := "products[?(@.price < 100)]"
	fmt.Printf("Testing path: %s\n", path)

	// Check the path characteristics
	fmt.Printf("Contains '[': %t\n", strings.Contains(path, "["))
	fmt.Printf("Ends with ']': %t\n", strings.HasSuffix(path, "]"))
	fmt.Printf("Contains '.': %t\n", strings.Contains(path, "."))
	fmt.Printf("Starts with '/': %t\n", strings.HasPrefix(path, "/"))

	// Test splitPath directly (simulate what the real code does)
	fmt.Println("\n=== Testing splitPath ===")

	// Since splitPath is not directly accessible, let's check what should happen
	if strings.Contains(path, ".") {
		// This path would go into the dotted paths logic
		fmt.Println("Path contains dots, will use splitPath logic")

		// Let's manually test what splitPath should return
		// According to the logic in splitPath, it should handle brackets correctly
		// and return ["products[?(@", "price < 100)]"] - this is wrong!

		// The real issue might be in splitPath - let's think about it
		// splitPath should recognize that the dots inside [?(...)] should not split the path
	}

	// Now test the actual query
	result := doc.Query(path)
	fmt.Printf("\nResult exists: %t, count: %d\n", result.Exists(), result.Count())
}
