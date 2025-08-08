package xjson

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestDebugBusinessScenario(t *testing.T) {
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

	// Check initial products
	products := root.Get("store").Get("products")
	assert.Equal(t, 1, products.Len())

	// Add a new product
	newProduct := map[string]interface{}{
		"name": "Product 2",
	}
	
	products.Append(newProduct)
	
	// Check updated products
	products = root.Get("store").Get("products") // Get fresh reference
	t.Logf("Products length: %d", products.Len())
	assert.Equal(t, 2, products.Len())
	
	// Check product names
	assert.Equal(t, "Product 1", products.Index(0).Get("name").String())
	assert.Equal(t, "Product 2", products.Index(1).Get("name").String())
}