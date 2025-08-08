package xjson

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestDebugAppendIssue(t *testing.T) {
	ecommerceJSON := `{
		"store": {
			"products": [
				{
					"name": "Product 1"
				}
			]
		}
	}`

	root, err := Parse(ecommerceJSON)
	assert.NoError(t, err)
	assert.True(t, root.IsValid())

	// Check initial length
	products := root.Get("store").Get("products")
	t.Logf("Initial products length: %d", products.Len())
	assert.Equal(t, 1, products.Len())

	// Add a new product
	newProduct := map[string]interface{}{
		"name": "Product 2",
	}
	
	// Append the new product
	products.Append(newProduct)
	
	// Check length after append - using the same reference
	t.Logf("Products length after append (same ref): %d", products.Len())
	
	// Get a fresh reference
	freshProducts := root.Get("store").Get("products")
	t.Logf("Products length after append (fresh ref): %d", freshProducts.Len())
	
	// They should both be 2
	assert.Equal(t, 2, products.Len())
	assert.Equal(t, 2, freshProducts.Len())
}